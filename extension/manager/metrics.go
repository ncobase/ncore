package manager

import (
	"fmt"
	"time"

	"github.com/ncobase/ncore/extension/metrics"
)

// GetMetrics returns comprehensive real-time metrics (extension layer only)
func (m *Manager) GetMetrics() map[string]any {
	if m.metricsCollector == nil {
		return map[string]any{
			"enabled":   false,
			"timestamp": time.Now(),
		}
	}

	// Update system metrics before getting snapshot
	m.updateSystemMetrics()

	result := map[string]any{
		"enabled":    true,
		"timestamp":  time.Now(),
		"system":     m.metricsCollector.GetSystemMetrics(),
		"extensions": m.metricsCollector.GetAllExtensionMetrics(),
		"storage":    m.metricsCollector.GetStorageStats(),
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

// GetSystemMetrics returns comprehensive system metrics from all layers
func (m *Manager) GetSystemMetrics() map[string]any {
	result := map[string]any{
		"timestamp": time.Now(),
		"layers":    make(map[string]any),
	}

	layers := result["layers"].(map[string]any)

	// Extension layer metrics
	layers["extension"] = m.GetMetrics()

	// Data layer metrics
	layers["data"] = m.GetDataMetrics()

	// Service discovery metrics
	if m.serviceDiscovery != nil {
		layers["service_discovery"] = m.GetServiceCacheStats()
	}

	// Events metrics
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

	// Detailed metrics
	details["extensions"] = m.GetMetrics()
	details["data"] = m.GetDataMetrics()
	details["events"] = m.GetEventsMetrics()
	details["service_discovery"] = m.GetServiceCacheStats()

	return result
}

// GetExtensionMetrics returns metrics for specific extension
func (m *Manager) GetExtensionMetrics(name string) *metrics.ExtensionMetrics {
	if m.metricsCollector == nil {
		return nil
	}

	return m.metricsCollector.GetExtensionMetrics(name)
}

// QueryHistoricalMetrics queries historical metrics with aggregation
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

// Helper methods for metrics aggregation

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

	// Simple health check without full health details
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

// updateSystemMetrics updates system-level metrics
func (m *Manager) updateSystemMetrics() {
	if m.metricsCollector == nil || !m.metricsCollector.IsEnabled() {
		return
	}

	// Update basic system metrics (memory, goroutines, etc.)
	m.metricsCollector.UpdateSystemMetrics()

	// Update service discovery metrics if available
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

// Extension metrics tracking methods (keep existing ones)

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

// CleanupOldMetrics removes metrics older than the specified duration
func (m *Manager) CleanupOldMetrics(maxAge time.Duration) error {
	if m.metricsCollector == nil {
		return fmt.Errorf("metrics collector not initialized")
	}

	return m.metricsCollector.CleanupOldMetrics(maxAge)
}
