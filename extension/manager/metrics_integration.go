package manager

import (
	"fmt"
	"runtime"
	"time"

	extMetrics "github.com/ncobase/ncore/extension/metrics"
)

// Metrics query and management methods

// QueryMetrics queries historical metrics
func (m *Manager) QueryMetrics(collection string, start, end time.Time) ([]*extMetrics.MetricSnapshot, error) {
	if m.metricsManager == nil {
		return nil, fmt.Errorf("metrics manager not initialized")
	}
	return m.metricsManager.Query(collection, start, end)
}

// GetLatestMetrics gets latest metrics for a collection
func (m *Manager) GetLatestMetrics(collection string, limit int) ([]*extMetrics.MetricSnapshot, error) {
	if m.metricsManager == nil {
		return nil, fmt.Errorf("metrics manager not initialized")
	}
	return m.metricsManager.QueryLatest(collection, limit)
}

// GetMetricsStats returns storage statistics
func (m *Manager) GetMetricsStats() map[string]any {
	if m.metricsManager == nil {
		return map[string]any{"status": "disabled"}
	}
	return m.metricsManager.GetStats()
}

// GetMetrics returns comprehensive system metrics
func (m *Manager) GetMetrics() map[string]any {
	metrics := map[string]any{
		"timestamp": time.Now(),
	}

	// Core metrics from collections
	if m.metricsManager != nil && m.metricsManager.IsEnabled() {
		collections := m.metricsManager.GetAllCollections()
		metricsData := make(map[string]any)

		for name, collection := range collections {
			snapshots := make([]map[string]any, 0, len(collection.Metrics))
			for _, metric := range collection.Metrics {
				snapshots = append(snapshots, map[string]any{
					"name":      metric.Name,
					"type":      m.getMetricTypeString(metric.Type),
					"value":     metric.Value.Load(),
					"labels":    metric.Labels,
					"timestamp": metric.Timestamp,
					"help":      metric.Help,
					"unit":      metric.Unit,
				})
			}

			metricsData[name] = map[string]any{
				"name":         collection.Name,
				"metrics":      snapshots,
				"last_updated": collection.LastUpdated,
			}
		}

		metrics["collections"] = metricsData
		metrics["storage"] = m.GetMetricsStats()
	}

	// Service discovery cache stats
	metrics["service_cache"] = m.GetServiceCacheStats()

	// Data layer health and stats
	if m.data != nil {
		metrics["data_health"] = m.data.Health(m.ctx)
		if dataStats := m.data.GetStats(); dataStats != nil {
			metrics["data_stats"] = dataStats
		}
	}

	// Security status if sandbox enabled
	if m.sandbox != nil {
		metrics["security"] = m.getSecurityStatus()
	}

	// Resource usage if monitoring enabled
	if m.resourceMonitor != nil {
		metrics["resource_usage"] = m.resourceMonitor.GetAllMetrics()
	}

	// System metrics
	if m.conf.Extension.Performance != nil && m.conf.Extension.Performance.EnableMetrics {
		systemMetrics := map[string]any{
			"memory_usage_mb": getMemStats() / 1024 / 1024,
			"goroutines":      runtime.NumGoroutine(),
			"gc_cycles":       getGCStats(),
		}
		metrics["system"] = systemMetrics
		m.trackSystemMetrics()
	}

	return metrics
}

// GetSpecificMetrics returns specific metric types
func (m *Manager) GetSpecificMetrics(metricType string) map[string]any {
	switch metricType {
	case "collections":
		if m.metricsManager != nil && m.metricsManager.IsEnabled() {
			return map[string]any{"collections": m.metricsManager.GetAllCollections()}
		}
		return map[string]any{"collections": map[string]any{}}

	case "storage":
		return m.GetMetricsStats()

	case "service_cache", "cache":
		return m.GetServiceCacheStats()

	case "data":
		if m.data != nil {
			return m.data.GetStats()
		}
		return map[string]any{"status": "not_initialized"}

	case "system":
		if m.conf.Extension.Performance != nil && m.conf.Extension.Performance.EnableMetrics {
			return map[string]any{
				"memory_usage_mb": getMemStats() / 1024 / 1024,
				"goroutines":      runtime.NumGoroutine(),
				"gc_cycles":       getGCStats(),
			}
		}
		return map[string]any{"status": "disabled"}

	case "security":
		if m.sandbox != nil {
			return m.getSecurityStatus()
		}
		return map[string]any{"status": "not_enabled"}

	case "resource":
		if m.resourceMonitor != nil {
			return map[string]any{"resource_usage": m.resourceMonitor.GetAllMetrics()}
		}
		return map[string]any{"status": "not_enabled"}

	default:
		return map[string]any{"error": "invalid metric type"}
	}
}

// Data layer metrics interface implementation (delegates to MetricsManager)

func (m *Manager) DBQuery(duration time.Duration, err error) {
	if m.metricsManager != nil {
		m.metricsManager.DBQuery(duration, err)
	}
}

func (m *Manager) DBTransaction(err error) {
	if m.metricsManager != nil {
		m.metricsManager.DBTransaction(err)
	}
}

func (m *Manager) DBConnections(count int) {
	if m.metricsManager != nil {
		m.metricsManager.DBConnections(count)
	}
}

func (m *Manager) RedisCommand(command string, err error) {
	if m.metricsManager != nil {
		m.metricsManager.RedisCommand(command, err)
	}
}

func (m *Manager) RedisConnections(count int) {
	if m.metricsManager != nil {
		m.metricsManager.RedisConnections(count)
	}
}

func (m *Manager) MongoOperation(operation string, err error) {
	if m.metricsManager != nil {
		m.metricsManager.MongoOperation(operation, err)
	}
}

func (m *Manager) SearchQuery(engine string, err error) {
	if m.metricsManager != nil {
		m.metricsManager.SearchQuery(engine, err)
	}
}

func (m *Manager) SearchIndex(engine, operation string) {
	if m.metricsManager != nil {
		m.metricsManager.SearchIndex(engine, operation)
	}
}

func (m *Manager) MQPublish(system string, err error) {
	if m.metricsManager != nil {
		m.metricsManager.MQPublish(system, err)
	}
}

func (m *Manager) MQConsume(system string, err error) {
	if m.metricsManager != nil {
		m.metricsManager.MQConsume(system, err)
	}
}

func (m *Manager) HealthCheck(component string, healthy bool) {
	if m.metricsManager != nil {
		m.metricsManager.HealthCheck(component, healthy)
	}
}

// Helper methods for metrics tracking

func (m *Manager) trackSystemMetrics() {
	if m.metricsManager != nil && m.conf.Extension.Performance != nil && m.conf.Extension.Performance.EnableMetrics {
		memoryMB := getMemStats() / 1024 / 1024
		goroutines := runtime.NumGoroutine()
		gcCycles := getGCStats()

		m.metricsManager.SystemMetrics(int64(memoryMB), goroutines, gcCycles)
	}
}

func (m *Manager) getMetricTypeString(t extMetrics.MetricType) string {
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

// Helper functions for system metrics
func getMemStats() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

func getGCStats() uint32 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.NumGC
}
