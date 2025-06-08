package metrics

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

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

	// Batch storage
	batchBuffer []*Snapshot
	batchSize   int
	lastFlush   time.Time
	flushTicker *time.Ticker
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

// NewCollector creates a new metrics collector
func NewCollector(storage Storage, enabled bool) *Collector {
	c := &Collector{
		extensions: make(map[string]*ExtensionMetrics),
		storage:    storage,
		enabled:    enabled,
		startTime:  time.Now(),
		batchSize:  100,
		lastFlush:  time.Now(),
		stopChan:   make(chan struct{}),
		system: SystemMetrics{
			StartTime: time.Now(),
		},
	}

	// Start background flush routine if enabled and storage available
	if enabled && storage != nil {
		c.flushTicker = time.NewTicker(30 * time.Second) // Flush every 30 seconds
		c.wg.Add(1)
		go c.flushRoutine()
	}

	return c
}

// NewCollectorWithMemoryStorage creates collector with memory storage
func NewCollectorWithMemoryStorage(enabled bool) *Collector {
	var storage Storage
	if enabled {
		storage = NewMemoryStorage()
	}
	return NewCollector(storage, enabled)
}

// UpgradeToRedisStorage upgrades from memory to Redis storage with proper error handling
func (c *Collector) UpgradeToRedisStorage(client interface{}, keyPrefix string, retention time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Type assertion to get Redis client from data layer
	redisClient, ok := client.(*redis.Client)
	if !ok {
		return fmt.Errorf("client is not a Redis client, got type %T", client)
	}

	// Test Redis connection before proceeding
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

// migrateFromMemoryToRedis migrates data from memory to Redis storage with error handling
func (c *Collector) migrateFromMemoryToRedis(memStorage *MemoryStorage, redisStorage *RedisStorage) error {
	// Get all data from memory storage
	memStorage.mu.RLock()
	defer memStorage.mu.RUnlock()

	var allSnapshots []*Snapshot
	for _, snapshots := range memStorage.data {
		allSnapshots = append(allSnapshots, snapshots...)
	}

	// Store in Redis if we have data
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

	if c.stopChan != nil {
		close(c.stopChan)
		c.stopChan = nil
	}

	if c.flushTicker != nil {
		c.flushTicker.Stop()
	}

	// Final flush before stopping
	c.flush()

	// Wait for background routines to finish
	c.wg.Wait()
}

// IsEnabled returns whether metrics collection is enabled
func (c *Collector) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.enabled
}

// SetEnabled enables or disables metrics collection
func (c *Collector) SetEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = enabled
}

// Extension lifecycle metrics with improved error handling

func (c *Collector) ExtensionLoaded(name string, duration time.Duration) {
	if !c.enabled || name == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	metrics := c.getOrCreateExtensionMetrics(name)
	metrics.LoadTime = duration.Milliseconds()
	metrics.LoadedAt = time.Now()

	// Store snapshot
	c.storeSnapshot(&Snapshot{
		ExtensionName: name,
		MetricType:    "load_time",
		Value:         duration.Milliseconds(),
		Timestamp:     time.Now(),
	})
}

func (c *Collector) ExtensionInitialized(name string, duration time.Duration, err error) {
	if !c.enabled || name == "" {
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

	// Store snapshot
	c.storeSnapshot(&Snapshot{
		ExtensionName: name,
		MetricType:    "init_time",
		Value:         duration.Milliseconds(),
		Timestamp:     time.Now(),
	})
}

func (c *Collector) ExtensionUnloaded(name string) {
	if !c.enabled || name == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.extensions, name)

	// Store unload event
	c.storeSnapshot(&Snapshot{
		ExtensionName: name,
		MetricType:    "unload_event",
		Value:         1,
		Timestamp:     time.Now(),
	})
}

// Service and event metrics with proper validation

func (c *Collector) ServiceCall(extensionName string, success bool) {
	if !c.enabled || extensionName == "" {
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

	// Store snapshot
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
	if !c.enabled || extensionName == "" {
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

	// Store snapshot with event type label
	c.storeSnapshot(&Snapshot{
		ExtensionName: extensionName,
		MetricType:    "event_published",
		Value:         1,
		Labels:        map[string]string{"event_type": eventType},
		Timestamp:     time.Now(),
	})
}

func (c *Collector) EventReceived(extensionName string, eventType string) {
	if !c.enabled || extensionName == "" {
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

	// Store snapshot with event type label
	c.storeSnapshot(&Snapshot{
		ExtensionName: extensionName,
		MetricType:    "event_received",
		Value:         1,
		Labels:        map[string]string{"event_type": eventType},
		Timestamp:     time.Now(),
	})
}

func (c *Collector) CircuitBreakerTripped(extensionName string) {
	if !c.enabled || extensionName == "" {
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

	// Store snapshot
	c.storeSnapshot(&Snapshot{
		ExtensionName: extensionName,
		MetricType:    "circuit_breaker_trip",
		Value:         1,
		Timestamp:     time.Now(),
	})
}

// System metrics collection with better resource tracking

func (c *Collector) UpdateSystemMetrics() {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Collect memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	c.system.MemoryUsageMB = int64(m.Alloc / 1024 / 1024)
	c.system.GoroutineCount = runtime.NumGoroutine()
	c.system.GCCycles = m.NumGC

	// Store system snapshots periodically
	now := time.Now()
	if now.Sub(c.lastFlush) > time.Minute {
		c.storeSnapshot(&Snapshot{
			ExtensionName: "system",
			MetricType:    "memory_usage",
			Value:         c.system.MemoryUsageMB,
			Timestamp:     now,
		})

		c.storeSnapshot(&Snapshot{
			ExtensionName: "system",
			MetricType:    "goroutine_count",
			Value:         int64(c.system.GoroutineCount),
			Timestamp:     now,
		})

		c.storeSnapshot(&Snapshot{
			ExtensionName: "system",
			MetricType:    "gc_cycles",
			Value:         int64(c.system.GCCycles),
			Timestamp:     now,
		})
	}
}

func (c *Collector) UpdateServiceDiscoveryMetrics(registered int, hits, misses int64) {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.system.ServicesRegistered = registered
	c.system.ServiceCacheHits = hits
	c.system.ServiceCacheMisses = misses

	// Store service discovery snapshots
	now := time.Now()
	c.storeSnapshot(&Snapshot{
		ExtensionName: "system",
		MetricType:    "services_registered",
		Value:         int64(registered),
		Timestamp:     now,
	})

	c.storeSnapshot(&Snapshot{
		ExtensionName: "system",
		MetricType:    "service_cache_hits",
		Value:         hits,
		Timestamp:     now,
	})

	c.storeSnapshot(&Snapshot{
		ExtensionName: "system",
		MetricType:    "service_cache_misses",
		Value:         misses,
		Timestamp:     now,
	})
}

// Query methods with improved error handling

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
		limit = 10 // Default limit
	}
	return c.storage.GetLatest(extensionName, limit)
}

func (c *Collector) GetStorageStats() map[string]any {
	if c.storage == nil {
		return map[string]any{"status": "not_configured"}
	}
	return c.storage.GetStats()
}

// Real-time access methods with proper synchronization

func (c *Collector) GetExtensionMetrics(name string) *ExtensionMetrics {
	if !c.enabled || name == "" {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	metrics, exists := c.extensions[name]
	if !exists {
		return nil
	}

	// Return a deep copy with atomic values converted to regular int64 for JSON serialization
	copied := &ExtensionMetrics{
		Name:          metrics.Name,
		LoadTime:      metrics.LoadTime,
		InitTime:      metrics.InitTime,
		LoadedAt:      metrics.LoadedAt,
		InitializedAt: metrics.InitializedAt,
		Status:        metrics.Status,
		// Convert atomic values to regular int64 for JSON serialization
		ServiceCalls:        metrics.serviceCalls.Load(),
		ServiceErrors:       metrics.serviceErrors.Load(),
		EventsPublished:     metrics.eventsPublished.Load(),
		EventsReceived:      metrics.eventsReceived.Load(),
		CircuitBreakerTrips: metrics.circuitBreakerTrips.Load(),
	}

	return copied
}

func (c *Collector) GetSystemMetrics() SystemMetrics {
	if !c.enabled {
		return SystemMetrics{}
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to prevent race conditions
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
	if !c.enabled {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]*ExtensionMetrics)
	for name, metrics := range c.extensions {
		// Create a deep copy with atomic values converted for JSON
		copied := &ExtensionMetrics{
			Name:          metrics.Name,
			LoadTime:      metrics.LoadTime,
			InitTime:      metrics.InitTime,
			LoadedAt:      metrics.LoadedAt,
			InitializedAt: metrics.InitializedAt,
			Status:        metrics.Status,
			// Convert atomic values to regular int64 for JSON serialization
			ServiceCalls:        metrics.serviceCalls.Load(),
			ServiceErrors:       metrics.serviceErrors.Load(),
			EventsPublished:     metrics.eventsPublished.Load(),
			EventsReceived:      metrics.eventsReceived.Load(),
			CircuitBreakerTrips: metrics.circuitBreakerTrips.Load(),
		}

		result[name] = copied
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
	if c.storage == nil || snapshot == nil {
		return
	}

	// Add to batch buffer
	c.batchBuffer = append(c.batchBuffer, snapshot)

	// Flush if buffer is full
	if len(c.batchBuffer) >= c.batchSize {
		c.flush()
	}
}

func (c *Collector) flush() {
	if c.storage == nil || len(c.batchBuffer) == 0 {
		return
	}

	// Store batch
	if err := c.storage.StoreBatch(c.batchBuffer); err != nil {
		logger.Errorf(nil, "Failed to flush metrics batch: %v", err)
	}

	// Clear buffer
	c.batchBuffer = c.batchBuffer[:0]
	c.lastFlush = time.Now()
}

func (c *Collector) flushRoutine() {
	defer c.wg.Done()

	for {
		select {
		case <-c.flushTicker.C:
			c.mu.Lock()
			c.flush()
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
