package manager

import (
	"fmt"
	"time"

	"github.com/ncobase/ncore/extension/metrics"
)

// IsMetricsEnabled returns whether metrics collection is enabled
func (m *Manager) IsMetricsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.metricsCollector != nil && m.metricsCollector.IsEnabled()
}

// GetMetrics returns comprehensive metrics
func (m *Manager) GetMetrics() map[string]any {
	if m.metricsCollector == nil {
		return map[string]any{
			"enabled":   false,
			"timestamp": time.Now(),
		}
	}

	if !m.metricsCollector.IsEnabled() {
		return map[string]any{
			"enabled":   false,
			"timestamp": time.Now(),
		}
	}

	m.updateSystemMetrics()

	result := map[string]any{
		"enabled":    true,
		"timestamp":  time.Now(),
		"system":     m.metricsCollector.GetSystemMetrics(),
		"extensions": m.metricsCollector.GetAllExtensionMetrics(),
		"storage":    m.metricsCollector.GetStorageStats(),
	}

	if m.conf.Extension.Metrics != nil {
		result["config"] = map[string]any{
			"flush_interval": m.conf.Extension.Metrics.FlushInterval,
			"batch_size":     m.conf.Extension.Metrics.BatchSize,
			"retention":      m.conf.Extension.Metrics.Retention,
			"storage_type":   m.conf.Extension.Metrics.Storage.Type,
		}
	}

	return result
}

// GetDataMetrics returns data layer metrics
func (m *Manager) GetDataMetrics() map[string]any {
	if m.data == nil {
		return map[string]any{
			"status":    "unavailable",
			"timestamp": time.Now(),
		}
	}

	return m.data.GetStats()
}

// GetSystemMetrics returns system metrics from all layers
func (m *Manager) GetSystemMetrics() map[string]any {
	result := map[string]any{
		"timestamp": time.Now(),
		"layers":    make(map[string]any),
	}

	layers := result["layers"].(map[string]any)

	layers["extension"] = m.GetMetrics()
	layers["data"] = m.GetDataMetrics()

	if m.serviceDiscovery != nil {
		layers["service_discovery"] = m.GetServiceCacheStats()
	}

	layers["events"] = m.GetEventsMetrics()

	return result
}

// GetComprehensiveMetrics returns all metrics in a structured format
func (m *Manager) GetComprehensiveMetrics() map[string]any {
	result := map[string]any{
		"timestamp": time.Now(),
		"summary":   make(map[string]any),
		"details":   make(map[string]any),
	}

	summary := result["summary"].(map[string]any)
	details := result["details"].(map[string]any)

	// Summary information
	summary["total_extensions"] = len(m.extensions)
	summary["active_extensions"] = m.countActiveExtensions()
	summary["data_layer_status"] = m.getDataLayerStatus()
	summary["messaging_status"] = m.getMessagingStatus()
	summary["extension_metrics_enabled"] = m.IsMetricsEnabled()
	summary["data_metrics_enabled"] = m.isDataMetricsEnabled()

	// Detailed metrics
	details["extensions"] = m.GetMetrics()
	details["data"] = m.GetDataMetrics()
	details["events"] = m.GetEventsMetrics()
	details["service_discovery"] = m.GetServiceCacheStats()

	return result
}

// GetEventsMetrics returns event metrics
func (m *Manager) GetEventsMetrics() map[string]any {
	if m.eventDispatcher == nil {
		return map[string]any{"status": "disabled"}
	}

	// Get metrics from event dispatcher
	dispatcherMetrics := m.eventDispatcher.GetMetrics()

	// Get extension-specific event metrics from collector
	extensionEventMetrics := make(map[string]any)
	if m.metricsCollector != nil && m.metricsCollector.IsEnabled() {
		extensions := m.metricsCollector.GetAllExtensionMetrics()
		for name, ext := range extensions {
			extensionEventMetrics[name] = map[string]any{
				"published": ext.EventsPublished,
				"received":  ext.EventsReceived,
			}
		}
	}

	return map[string]any{
		"dispatcher": dispatcherMetrics,
		"extensions": extensionEventMetrics,
		"timestamp":  time.Now(),
		"status":     "active",
	}
}

// GetServiceCacheStats returns service cache statistics
func (m *Manager) GetServiceCacheStats() map[string]any {
	if m.serviceDiscovery == nil {
		return map[string]any{
			"status": "not_initialized",
		}
	}

	stats := m.serviceDiscovery.GetCacheStats()
	stats["status"] = "active"
	return stats
}

// GetExtensionMetrics returns metrics for specific extension
func (m *Manager) GetExtensionMetrics(name string) *metrics.ExtensionMetrics {
	if m.metricsCollector == nil {
		return nil
	}

	return m.metricsCollector.GetExtensionMetrics(name)
}

// QueryHistoricalMetrics queries historical metrics
func (m *Manager) QueryHistoricalMetrics(opts *metrics.QueryOptions) ([]*metrics.AggregatedMetrics, error) {
	if m.metricsCollector == nil {
		return nil, fmt.Errorf("metrics collector not initialized")
	}

	return m.metricsCollector.Query(opts)
}

// GetLatestMetrics gets latest metrics for an extension
func (m *Manager) GetLatestMetrics(extensionName string, limit int) ([]*metrics.Snapshot, error) {
	if m.metricsCollector == nil {
		return nil, fmt.Errorf("metrics collector not initialized")
	}

	return m.metricsCollector.GetLatest(extensionName, limit)
}

// GetMetricsStorageStats returns storage statistics
func (m *Manager) GetMetricsStorageStats() map[string]any {
	if m.metricsCollector == nil {
		return map[string]any{"status": "disabled"}
	}

	return m.metricsCollector.GetStorageStats()
}

// countActiveExtensions counts active extensions
func (m *Manager) countActiveExtensions() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, ext := range m.extensions {
		if ext.Instance.Status() == "active" {
			count++
		}
	}
	return count
}

// getDataLayerStatus returns data layer status
func (m *Manager) getDataLayerStatus() string {
	if m.data == nil {
		return "unavailable"
	}

	stats := m.data.GetStats()
	if status, ok := stats["status"].(string); ok {
		return status
	}

	return "unknown"
}

// getMessagingStatus returns messaging status
func (m *Manager) getMessagingStatus() map[string]any {
	if m.data == nil {
		return map[string]any{
			"available": false,
			"services":  map[string]bool{},
		}
	}

	return map[string]any{
		"available": m.data.IsMessagingAvailable(),
		"services": map[string]bool{
			"rabbitmq": m.data.RabbitMQ != nil && m.data.RabbitMQ.IsConnected(),
			"kafka":    m.data.Kafka != nil && m.data.Kafka.IsConnected(),
		},
	}
}

// isDataMetricsEnabled checks if data layer metrics are enabled
func (m *Manager) isDataMetricsEnabled() bool {
	if m.data == nil {
		return false
	}

	stats := m.data.GetStats()

	if status, ok := stats["status"].(string); ok {
		return status != "metrics_unavailable"
	}

	if _, hasDB := stats["database"].(map[string]any); hasDB {
		return true
	}

	return false
}

// updateSystemMetrics updates system-level metrics
func (m *Manager) updateSystemMetrics() {
	if m.metricsCollector == nil || !m.metricsCollector.IsEnabled() {
		return
	}

	m.metricsCollector.UpdateSystemMetrics()

	if m.serviceDiscovery != nil {
		cacheStats := m.serviceDiscovery.GetCacheStats()

		registered := 0
		if registeredVal, ok := cacheStats["total"].(int); ok {
			registered = registeredVal
		}

		hits := int64(0)
		if hitsVal, ok := cacheStats["cache_hits"].(int64); ok {
			hits = hitsVal
		}

		misses := int64(0)
		if missesVal, ok := cacheStats["cache_misses"].(int64); ok {
			misses = missesVal
		}

		m.metricsCollector.UpdateServiceDiscoveryMetrics(registered, hits, misses)
	}
}

// Extension metrics tracking methods

// trackExtensionLoaded tracks extension loading
func (m *Manager) trackExtensionLoaded(name string, duration time.Duration) {
	if m.metricsCollector != nil {
		m.metricsCollector.ExtensionLoaded(name, duration)
	}
}

// trackExtensionInitialized tracks extension initialization
func (m *Manager) trackExtensionInitialized(name string, duration time.Duration, err error) {
	if m.metricsCollector != nil {
		m.metricsCollector.ExtensionInitialized(name, duration, err)
	}
}

// trackExtensionUnloaded tracks extension unloading
func (m *Manager) trackExtensionUnloaded(name string) {
	if m.metricsCollector != nil {
		m.metricsCollector.ExtensionUnloaded(name)
	}
}

// trackServiceCall tracks service call
func (m *Manager) trackServiceCall(extensionName string, success bool) {
	if m.metricsCollector != nil {
		m.metricsCollector.ServiceCall(extensionName, success)
	}
}

// trackEventPublished tracks event published
func (m *Manager) trackEventPublished(extensionName string, eventType string) {
	if m.metricsCollector != nil {
		m.metricsCollector.EventPublished(extensionName, eventType)
	}
}

// trackEventReceived tracks event received
func (m *Manager) trackEventReceived(extensionName string, eventType string) {
	if m.metricsCollector != nil {
		m.metricsCollector.EventReceived(extensionName, eventType)
	}
}

// trackCircuitBreakerTripped tracks circuit breaker tripped
func (m *Manager) trackCircuitBreakerTripped(extensionName string) {
	if m.metricsCollector != nil {
		m.metricsCollector.CircuitBreakerTripped(extensionName)
	}
}

// CleanupOldMetrics removes old metrics
func (m *Manager) CleanupOldMetrics(maxAge time.Duration) error {
	if m.metricsCollector == nil {
		return fmt.Errorf("metrics collector not initialized")
	}

	return m.metricsCollector.CleanupOldMetrics(maxAge)
}
