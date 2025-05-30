package manager

import (
	"context"
	"fmt"
	"time"

	extMetrics "github.com/ncobase/ncore/extension/metrics"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/redis/go-redis/v9"
)

// MetricsManager handles all metrics-related functionality for the extension manager
type MetricsManager struct {
	collector     *extMetrics.Collector
	dataCollector *extMetrics.DataCollector
	enabled       bool
	retention     time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewMetricsManager creates a new metrics manager
func NewMetricsManager(ctx context.Context, enabled bool, retention time.Duration) *MetricsManager {
	if !enabled {
		return &MetricsManager{enabled: false}
	}

	metricsCtx, cancel := context.WithCancel(ctx)

	// Start with memory storage
	storage := extMetrics.NewMemoryStorage()
	collector := extMetrics.NewCollector(storage, retention)
	dataCollector := extMetrics.NewDataCollector(collector)

	mm := &MetricsManager{
		collector:     collector,
		dataCollector: dataCollector,
		enabled:       true,
		retention:     retention,
		ctx:           metricsCtx,
		cancel:        cancel,
	}

	// Start cleanup routine
	go mm.cleanupRoutine()

	return mm
}

// UpgradeToRedisStorage upgrades from memory to Redis storage
func (mm *MetricsManager) UpgradeToRedisStorage(redis *redis.Client, keyPrefix string) {
	if !mm.enabled || redis == nil {
		return
	}

	redisStorage := extMetrics.NewRedisStorage(redis, keyPrefix, mm.retention)
	mm.collector = extMetrics.NewCollector(redisStorage, mm.retention)
	mm.dataCollector = extMetrics.NewDataCollector(mm.collector)
}

// cleanupRoutine runs periodic cleanup of old metrics
func (mm *MetricsManager) cleanupRoutine() {
	if !mm.enabled {
		return
	}

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := mm.collector.Cleanup(); err != nil {
				logger.Errorf(mm.ctx, "metrics cleanup failed: %v", err)
			}
		case <-mm.ctx.Done():
			return
		}
	}
}

// Extension metrics tracking methods

// ExtensionLoaded tracks when an extension is loaded
func (mm *MetricsManager) ExtensionLoaded(name string, duration time.Duration) {
	if !mm.enabled {
		return
	}

	mm.collector.Inc("extensions", "loaded_total", map[string]string{"name": name})
	mm.collector.Observe("extensions", "load_duration_seconds", duration.Seconds(), map[string]string{"name": name})
}

// ExtensionInitialized tracks extension initialization
func (mm *MetricsManager) ExtensionInitialized(name string, duration time.Duration, err error) {
	if !mm.enabled {
		return
	}

	labels := map[string]string{"name": name}
	if err != nil {
		labels["status"] = "error"
		mm.collector.Inc("extensions", "init_errors_total", labels)
	} else {
		labels["status"] = "success"
		mm.collector.Inc("extensions", "initialized_total", labels)
	}
	mm.collector.Observe("extensions", "init_duration_seconds", duration.Seconds(), map[string]string{"name": name})
}

// ExtensionUnloaded tracks when an extension is unloaded
func (mm *MetricsManager) ExtensionUnloaded(name string) {
	if !mm.enabled {
		return
	}

	mm.collector.Inc("extensions", "unloaded_total", map[string]string{"name": name})
}

// ExtensionPhase tracks initialization phases
func (mm *MetricsManager) ExtensionPhase(name, phase string, duration time.Duration, err error) {
	if !mm.enabled {
		return
	}

	labels := map[string]string{"name": name, "phase": phase}
	if err != nil {
		labels["status"] = "error"
		mm.collector.Inc("extensions", "phase_errors_total", labels)
	} else {
		labels["status"] = "success"
	}
	mm.collector.Observe("extensions", "phase_duration_seconds", duration.Seconds(), labels)
}

// ServiceCall tracks service call metrics
func (mm *MetricsManager) ServiceCall(serviceName, methodName string, duration time.Duration, err error) {
	if !mm.enabled {
		return
	}

	labels := map[string]string{
		"service": serviceName,
		"method":  methodName,
	}
	if err != nil {
		labels["status"] = "error"
		mm.collector.Inc("services", "call_errors_total", labels)
	} else {
		labels["status"] = "success"
		mm.collector.Inc("services", "calls_total", labels)
	}
	mm.collector.Observe("services", "call_duration_seconds", duration.Seconds(), labels)
}

// EventPublished tracks event publication
func (mm *MetricsManager) EventPublished(eventName string, target string) {
	if !mm.enabled {
		return
	}

	labels := map[string]string{
		"event":  eventName,
		"target": target,
	}
	mm.collector.Inc("events", "published_total", labels)
}

// EventDelivered tracks event delivery
func (mm *MetricsManager) EventDelivered(eventName string, err error) {
	if !mm.enabled {
		return
	}

	labels := map[string]string{"event": eventName}
	if err != nil {
		labels["status"] = "error"
		mm.collector.Inc("events", "delivery_errors_total", labels)
	} else {
		labels["status"] = "success"
		mm.collector.Inc("events", "delivered_total", labels)
	}
}

// PluginLoadAttempt tracks plugin load attempts
func (mm *MetricsManager) PluginLoadAttempt(name string, success bool) {
	if !mm.enabled {
		return
	}

	labels := map[string]string{"name": name}
	if success {
		labels["status"] = "success"
	} else {
		labels["status"] = "error"
	}
	mm.collector.Inc("extensions", "load_attempts_total", labels)
}

// CircuitBreakerTripped tracks circuit breaker events
func (mm *MetricsManager) CircuitBreakerTripped(extensionName string) {
	if !mm.enabled {
		return
	}

	mm.collector.Inc("circuit_breaker", "tripped_total", map[string]string{"extension": extensionName})
}

// SystemMetrics tracks system-level metrics
func (mm *MetricsManager) SystemMetrics(memoryMB int64, goroutines int, gcCycles uint32) {
	if !mm.enabled {
		return
	}

	mm.collector.Set("system", "memory_usage_mb", float64(memoryMB), nil)
	mm.collector.Set("system", "goroutines", float64(goroutines), nil)
	mm.collector.Set("system", "gc_cycles", float64(gcCycles), nil)
}

// Data layer metrics implementation (implements metrics.ExtensionCollector)

func (mm *MetricsManager) DBQuery(duration time.Duration, err error) {
	if mm.dataCollector != nil {
		mm.dataCollector.DBQuery(duration, err)
	}
}

func (mm *MetricsManager) DBTransaction(err error) {
	if mm.dataCollector != nil {
		mm.dataCollector.DBTransaction(err)
	}
}

func (mm *MetricsManager) DBConnections(count int) {
	if mm.dataCollector != nil {
		mm.dataCollector.DBConnections(count)
	}
}

func (mm *MetricsManager) RedisCommand(command string, err error) {
	if mm.dataCollector != nil {
		mm.dataCollector.RedisCommand(command, err)
	}
}

func (mm *MetricsManager) RedisConnections(count int) {
	if mm.dataCollector != nil {
		mm.dataCollector.RedisConnections(count)
	}
}

func (mm *MetricsManager) MongoOperation(operation string, err error) {
	if mm.dataCollector != nil {
		mm.dataCollector.MongoOperation(operation, err)
	}
}

func (mm *MetricsManager) SearchQuery(engine string, err error) {
	if mm.dataCollector != nil {
		mm.dataCollector.SearchQuery(engine, err)
	}
}

func (mm *MetricsManager) SearchIndex(engine, operation string) {
	if mm.dataCollector != nil {
		mm.dataCollector.SearchIndex(engine, operation)
	}
}

func (mm *MetricsManager) MQPublish(system string, err error) {
	if mm.dataCollector != nil {
		mm.dataCollector.MQPublish(system, err)
	}
}

func (mm *MetricsManager) MQConsume(system string, err error) {
	if mm.dataCollector != nil {
		mm.dataCollector.MQConsume(system, err)
	}
}

func (mm *MetricsManager) HealthCheck(component string, healthy bool) {
	if mm.dataCollector != nil {
		mm.dataCollector.HealthCheck(component, healthy)
	}
}

// Query and snapshot methods

// GetAllCollections returns all metric collections
func (mm *MetricsManager) GetAllCollections() map[string]*extMetrics.MetricCollection {
	if !mm.enabled {
		return make(map[string]*extMetrics.MetricCollection)
	}
	return mm.collector.GetAllCollections()
}

// Snapshot creates snapshots of all metrics
func (mm *MetricsManager) Snapshot() map[string][]*extMetrics.MetricSnapshot {
	if !mm.enabled {
		return make(map[string][]*extMetrics.MetricSnapshot)
	}
	return mm.collector.Snapshot()
}

// Query retrieves historical metrics
func (mm *MetricsManager) Query(collection string, start, end time.Time) ([]*extMetrics.MetricSnapshot, error) {
	if !mm.enabled {
		return nil, fmt.Errorf("metrics not enabled")
	}
	return mm.collector.Query(collection, start, end)
}

// QueryLatest retrieves latest metrics
func (mm *MetricsManager) QueryLatest(collection string, limit int) ([]*extMetrics.MetricSnapshot, error) {
	if !mm.enabled {
		return nil, fmt.Errorf("metrics not enabled")
	}
	return mm.collector.QueryLatest(collection, limit)
}

// GetStats returns storage statistics
func (mm *MetricsManager) GetStats() map[string]any {
	if !mm.enabled {
		return map[string]any{"status": "disabled"}
	}

	collections := mm.collector.GetAllCollections()
	stats := map[string]any{
		"status":            "enabled",
		"total_collections": len(collections),
		"collections":       make(map[string]any),
		"timestamp":         time.Now(),
	}

	collectionStats := stats["collections"].(map[string]any)
	totalMetrics := 0

	for name, collection := range collections {
		metricCount := len(collection.Metrics)
		totalMetrics += metricCount

		collectionStats[name] = map[string]any{
			"metric_count": metricCount,
			"last_updated": collection.LastUpdated,
		}
	}

	stats["total_metrics"] = totalMetrics
	return stats
}

// IsEnabled returns whether metrics are enabled
func (mm *MetricsManager) IsEnabled() bool {
	return mm.enabled
}

// Cleanup stops the metrics manager
func (mm *MetricsManager) Cleanup() {
	if mm.cancel != nil {
		mm.cancel()
	}

	if mm.enabled && mm.collector != nil {
		if err := mm.collector.Cleanup(); err != nil {
			logger.Errorf(context.Background(), "final metrics cleanup failed: %v", err)
		}
	}
}

// Helper methods for converting metric types
func (mm *MetricsManager) getMetricTypeString(t extMetrics.MetricType) string {
	switch t {
	case extMetrics.Counter:
		return "counter"
	case extMetrics.Gauge:
		return "gauge"
	case extMetrics.Histogram:
		return "histogram"
	case extMetrics.Summary:
		return "summary"
	default:
		return "unknown"
	}
}

// getTargetString converts target flag to string for metrics
func (mm *MetricsManager) getTargetString(target int) string {
	switch target {
	case 1: // EventTargetMemory
		return "memory"
	case 2: // EventTargetQueue
		return "queue"
	case 3: // EventTargetAll
		return "all"
	default:
		return "unknown"
	}
}
