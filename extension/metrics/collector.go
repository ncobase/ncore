package metrics

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/ncobase/ncore/extension/config"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/redis/go-redis/v9"
)

// Collector manages extension metrics collection with storage
type Collector struct {
	mu         sync.RWMutex
	extensions map[string]*ExtensionMetrics
	system     SystemMetrics
	storage    Storage
	enabled    bool
	startTime  time.Time

	// Background processing
	batchBuffer []*Snapshot
	batchSize   int
	lastFlush   time.Time
	flushTicker *time.Ticker
	stopChan    chan struct{}
	stopped     bool
	wg          sync.WaitGroup
}

// NewCollector creates a new metrics collector from extension config
func NewCollector(cfg *config.MetricsConfig) *Collector {
	if cfg == nil || !cfg.Enabled {
		// Metrics disabled, create a disabled collector
		return &Collector{
			extensions: make(map[string]*ExtensionMetrics),
			enabled:    false,
			startTime:  time.Now(),
			system: SystemMetrics{
				StartTime: time.Now(),
			},
		}
	}

	// Parse configuration values
	flushInterval := 30 * time.Second
	if cfg.FlushInterval != "" {
		if interval, err := time.ParseDuration(cfg.FlushInterval); err == nil {
			flushInterval = interval
		}
	}

	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	c := &Collector{
		extensions: make(map[string]*ExtensionMetrics),
		storage:    NewMemoryStorage(),
		enabled:    true,
		startTime:  time.Now(),
		batchSize:  batchSize,
		lastFlush:  time.Now(),
		stopChan:   make(chan struct{}),
		system: SystemMetrics{
			StartTime: time.Now(),
		},
	}

	// Start background flush routine
	c.flushTicker = time.NewTicker(flushInterval)
	c.wg.Add(1)
	go c.flushRoutine()

	return c
}

// UpgradeToRedisStorage upgrades from memory to Redis storage
func (c *Collector) UpgradeToRedisStorage(client any, keyPrefix string, retention time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stopped {
		return fmt.Errorf("collector is stopped")
	}

	// Type assertion to get Redis client
	redisClient, ok := client.(*redis.Client)
	if !ok {
		return fmt.Errorf("client is not a Redis client, got type %T", client)
	}

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis connection test failed: %w", err)
	}

	// Create Redis storage
	redisStorage := NewRedisStorage(redisClient, keyPrefix, retention)

	// Migrate existing data if we have memory storage
	if memStorage, isMemory := c.storage.(*MemoryStorage); isMemory {
		if err := c.migrateFromMemoryToRedis(memStorage, redisStorage); err != nil {
			return fmt.Errorf("failed to migrate data to Redis: %w", err)
		}
	}

	// Switch to Redis storage
	c.storage = redisStorage
	return nil
}

// migrateFromMemoryToRedis migrates data from memory to Redis storage
func (c *Collector) migrateFromMemoryToRedis(memStorage *MemoryStorage, redisStorage *RedisStorage) error {
	memStorage.mu.RLock()
	defer memStorage.mu.RUnlock()

	var allSnapshots []*Snapshot
	for _, snapshots := range memStorage.data {
		allSnapshots = append(allSnapshots, snapshots...)
	}

	if len(allSnapshots) > 0 {
		if err := redisStorage.StoreBatch(allSnapshots); err != nil {
			return fmt.Errorf("failed to store migrated data: %w", err)
		}
		logger.Infof(nil, "Migrated %d metric snapshots to Redis storage", len(allSnapshots))
	}

	return nil
}

// Stop gracefully stops the collector
func (c *Collector) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stopped {
		return
	}

	c.stopped = true

	// Stop background routines first
	if c.flushTicker != nil {
		c.flushTicker.Stop()
		c.flushTicker = nil
	}

	if c.stopChan != nil {
		close(c.stopChan)
		c.stopChan = nil
	}

	// Wait for background routines to finish
	c.wg.Wait()

	// Final flush before stopping (only if storage is still available)
	if c.storage != nil {
		c.flushUnsafe()
	}
}

// IsEnabled returns whether metrics collection is enabled
func (c *Collector) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.enabled && !c.stopped
}

// SetEnabled enables or disables metrics collection
func (c *Collector) SetEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = enabled
}

// Extension lifecycle metrics

func (c *Collector) ExtensionLoaded(name string, duration time.Duration) {
	if !c.IsEnabled() || name == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	metrics := c.getOrCreateExtensionMetrics(name)
	metrics.LoadTime = duration.Milliseconds()
	metrics.LoadedAt = time.Now()

	c.storeSnapshotUnsafe(&Snapshot{
		ExtensionName: name,
		MetricType:    "load_time",
		Value:         duration.Milliseconds(),
		Timestamp:     time.Now(),
	})
}

func (c *Collector) ExtensionInitialized(name string, duration time.Duration, err error) {
	if !c.IsEnabled() || name == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	metrics := c.getOrCreateExtensionMetrics(name)
	metrics.InitTime = duration.Milliseconds()
	metrics.InitializedAt = time.Now()

	if err != nil {
		metrics.Status = "failed"
	} else {
		metrics.Status = "active"
	}

	c.storeSnapshotUnsafe(&Snapshot{
		ExtensionName: name,
		MetricType:    "init_time",
		Value:         duration.Milliseconds(),
		Timestamp:     time.Now(),
	})
}

func (c *Collector) ExtensionUnloaded(name string) {
	if !c.IsEnabled() || name == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.extensions, name)

	c.storeSnapshotUnsafe(&Snapshot{
		ExtensionName: name,
		MetricType:    "unload_event",
		Value:         1,
		Timestamp:     time.Now(),
	})
}

// Service and event metrics

func (c *Collector) ServiceCall(extensionName string, success bool) {
	if !c.IsEnabled() || extensionName == "" {
		return
	}

	c.mu.RLock()
	metrics, exists := c.extensions[extensionName]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		metrics = c.getOrCreateExtensionMetrics(extensionName)
		c.mu.Unlock()
	}

	metrics.serviceCalls.Add(1)
	if !success {
		metrics.serviceErrors.Add(1)
	}

	value := int64(1)
	if !success {
		value = 0
	}

	c.storeSnapshot(&Snapshot{
		ExtensionName: extensionName,
		MetricType:    "service_call",
		Value:         value,
		Labels:        map[string]string{"success": fmt.Sprintf("%t", success)},
		Timestamp:     time.Now(),
	})
}

func (c *Collector) EventPublished(extensionName string, eventType string) {
	if !c.IsEnabled() || extensionName == "" {
		return
	}

	c.mu.RLock()
	metrics, exists := c.extensions[extensionName]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		metrics = c.getOrCreateExtensionMetrics(extensionName)
		c.mu.Unlock()
	}

	metrics.eventsPublished.Add(1)

	c.storeSnapshot(&Snapshot{
		ExtensionName: extensionName,
		MetricType:    "event_published",
		Value:         1,
		Labels:        map[string]string{"event_type": eventType},
		Timestamp:     time.Now(),
	})
}

func (c *Collector) EventReceived(extensionName string, eventType string) {
	if !c.IsEnabled() || extensionName == "" {
		return
	}

	c.mu.RLock()
	metrics, exists := c.extensions[extensionName]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		metrics = c.getOrCreateExtensionMetrics(extensionName)
		c.mu.Unlock()
	}

	metrics.eventsReceived.Add(1)

	c.storeSnapshot(&Snapshot{
		ExtensionName: extensionName,
		MetricType:    "event_received",
		Value:         1,
		Labels:        map[string]string{"event_type": eventType},
		Timestamp:     time.Now(),
	})
}

func (c *Collector) CircuitBreakerTripped(extensionName string) {
	if !c.IsEnabled() || extensionName == "" {
		return
	}

	c.mu.RLock()
	metrics, exists := c.extensions[extensionName]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		metrics = c.getOrCreateExtensionMetrics(extensionName)
		c.mu.Unlock()
	}

	metrics.circuitBreakerTrips.Add(1)

	c.storeSnapshot(&Snapshot{
		ExtensionName: extensionName,
		MetricType:    "circuit_breaker_trip",
		Value:         1,
		Timestamp:     time.Now(),
	})
}

// System metrics collection

func (c *Collector) UpdateSystemMetrics() {
	if !c.IsEnabled() {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	c.system.MemoryUsageMB = int64(m.Alloc / 1024 / 1024)
	c.system.GoroutineCount = runtime.NumGoroutine()
	c.system.GCCycles = m.NumGC

	now := time.Now()
	if now.Sub(c.lastFlush) > time.Minute {
		c.storeSnapshotUnsafe(&Snapshot{
			ExtensionName: "system",
			MetricType:    "memory_usage",
			Value:         c.system.MemoryUsageMB,
			Timestamp:     now,
		})

		c.storeSnapshotUnsafe(&Snapshot{
			ExtensionName: "system",
			MetricType:    "goroutine_count",
			Value:         int64(c.system.GoroutineCount),
			Timestamp:     now,
		})

		c.storeSnapshotUnsafe(&Snapshot{
			ExtensionName: "system",
			MetricType:    "gc_cycles",
			Value:         int64(c.system.GCCycles),
			Timestamp:     now,
		})
	}
}

func (c *Collector) UpdateServiceDiscoveryMetrics(registered int, hits, misses int64) {
	if !c.IsEnabled() {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.system.ServicesRegistered = registered
	c.system.ServiceCacheHits = hits
	c.system.ServiceCacheMisses = misses

	now := time.Now()
	c.storeSnapshotUnsafe(&Snapshot{
		ExtensionName: "system",
		MetricType:    "services_registered",
		Value:         int64(registered),
		Timestamp:     now,
	})

	c.storeSnapshotUnsafe(&Snapshot{
		ExtensionName: "system",
		MetricType:    "service_cache_hits",
		Value:         hits,
		Timestamp:     now,
	})

	c.storeSnapshotUnsafe(&Snapshot{
		ExtensionName: "system",
		MetricType:    "service_cache_misses",
		Value:         misses,
		Timestamp:     now,
	})
}

// Query methods

func (c *Collector) Query(opts *QueryOptions) ([]*AggregatedMetrics, error) {
	if c.storage == nil {
		return nil, fmt.Errorf("storage not configured")
	}
	if opts == nil {
		return nil, fmt.Errorf("query options cannot be nil")
	}
	return c.storage.Query(opts)
}

func (c *Collector) GetLatest(extensionName string, limit int) ([]*Snapshot, error) {
	if c.storage == nil {
		return nil, fmt.Errorf("storage not configured")
	}
	if extensionName == "" {
		return nil, fmt.Errorf("extension name cannot be empty")
	}
	if limit <= 0 {
		limit = 10
	}
	return c.storage.GetLatest(extensionName, limit)
}

func (c *Collector) GetStorageStats() map[string]any {
	if c.storage == nil {
		return map[string]any{"status": "not_configured"}
	}
	return c.storage.GetStats()
}

// Real-time access methods

func (c *Collector) GetExtensionMetrics(name string) *ExtensionMetrics {
	if !c.IsEnabled() || name == "" {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	metrics, exists := c.extensions[name]
	if !exists {
		return nil
	}

	return &ExtensionMetrics{
		Name:                metrics.Name,
		LoadTime:            metrics.LoadTime,
		InitTime:            metrics.InitTime,
		LoadedAt:            metrics.LoadedAt,
		InitializedAt:       metrics.InitializedAt,
		Status:              metrics.Status,
		ServiceCalls:        metrics.serviceCalls.Load(),
		ServiceErrors:       metrics.serviceErrors.Load(),
		EventsPublished:     metrics.eventsPublished.Load(),
		EventsReceived:      metrics.eventsReceived.Load(),
		CircuitBreakerTrips: metrics.circuitBreakerTrips.Load(),
	}
}

func (c *Collector) GetSystemMetrics() SystemMetrics {
	if !c.IsEnabled() {
		return SystemMetrics{}
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return SystemMetrics{
		StartTime:          c.system.StartTime,
		MemoryUsageMB:      c.system.MemoryUsageMB,
		GoroutineCount:     c.system.GoroutineCount,
		GCCycles:           c.system.GCCycles,
		ServicesRegistered: c.system.ServicesRegistered,
		ServiceCacheHits:   c.system.ServiceCacheHits,
		ServiceCacheMisses: c.system.ServiceCacheMisses,
	}
}

func (c *Collector) GetAllExtensionMetrics() map[string]*ExtensionMetrics {
	if !c.IsEnabled() {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]*ExtensionMetrics)
	for name, metrics := range c.extensions {
		result[name] = &ExtensionMetrics{
			Name:                metrics.Name,
			LoadTime:            metrics.LoadTime,
			InitTime:            metrics.InitTime,
			LoadedAt:            metrics.LoadedAt,
			InitializedAt:       metrics.InitializedAt,
			Status:              metrics.Status,
			ServiceCalls:        metrics.serviceCalls.Load(),
			ServiceErrors:       metrics.serviceErrors.Load(),
			EventsPublished:     metrics.eventsPublished.Load(),
			EventsReceived:      metrics.eventsReceived.Load(),
			CircuitBreakerTrips: metrics.circuitBreakerTrips.Load(),
		}
	}

	return result
}

// Internal helper methods

func (c *Collector) getOrCreateExtensionMetrics(name string) *ExtensionMetrics {
	// Caller must hold write lock
	metrics, exists := c.extensions[name]
	if !exists {
		metrics = &ExtensionMetrics{
			Name:   name,
			Status: "loading",
		}
		c.extensions[name] = metrics
	}
	return metrics
}

func (c *Collector) storeSnapshot(snapshot *Snapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.storeSnapshotUnsafe(snapshot)
}

func (c *Collector) storeSnapshotUnsafe(snapshot *Snapshot) {
	if c.storage == nil || snapshot == nil || c.stopped {
		return
	}

	c.batchBuffer = append(c.batchBuffer, snapshot)

	if len(c.batchBuffer) >= c.batchSize {
		c.flushUnsafe()
	}
}

func (c *Collector) flushUnsafe() {
	if c.storage == nil || len(c.batchBuffer) == 0 || c.stopped {
		return
	}

	if err := c.storage.StoreBatch(c.batchBuffer); err != nil {
		logger.Errorf(nil, "Failed to flush metrics batch: %v", err)
	}

	c.batchBuffer = c.batchBuffer[:0]
	c.lastFlush = time.Now()
}

func (c *Collector) flushRoutine() {
	defer c.wg.Done()

	for {
		select {
		case <-c.flushTicker.C:
			c.mu.Lock()
			if !c.stopped {
				c.flushUnsafe()
			}
			c.mu.Unlock()
		case <-c.stopChan:
			return
		}
	}
}

// CleanupOldMetrics removes metrics older than the specified duration
func (c *Collector) CleanupOldMetrics(maxAge time.Duration) error {
	if c.storage == nil {
		return fmt.Errorf("storage not configured")
	}

	before := time.Now().Add(-maxAge)
	return c.storage.Cleanup(before)
}
