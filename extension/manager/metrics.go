package manager

import (
	"fmt"
	"time"

	"github.com/ncobase/ncore/extension/metrics"
)

// GetMetrics returns comprehensive real-time metrics
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

// GetExtensionMetrics returns metrics for specific extension
func (m *Manager) GetExtensionMetrics(name string) *metrics.ExtensionMetrics {
	if m.metricsCollector == nil {
		return nil
	}

	return m.metricsCollector.GetExtensionMetrics(name)
}

// GetSystemMetrics returns system metrics
func (m *Manager) GetSystemMetrics() metrics.SystemMetrics {
	if m.metricsCollector == nil {
		return metrics.SystemMetrics{}
	}

	m.updateSystemMetrics()
	return m.metricsCollector.GetSystemMetrics()
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

// updateSystemMetrics updates system-level metrics with improved implementation
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

// trackEventPublished tracks event published with event type
func (m *Manager) trackEventPublished(extensionName string, eventType string) {
	if m.metricsCollector != nil {
		m.metricsCollector.EventPublished(extensionName, eventType)
	}
}

// trackEventReceived tracks event received with event type
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

// getSecurityStatus returns security status
func (m *Manager) getSecurityStatus() map[string]any {
	status := map[string]any{
		"sandbox_enabled": m.sandbox != nil,
	}

	if m.conf.Extension.Security != nil {
		status["signature_required"] = m.conf.Extension.Security.RequireSignature
		status["trusted_sources"] = len(m.conf.Extension.Security.TrustedSources)
		status["allowed_paths"] = len(m.conf.Extension.Security.AllowedPaths)
		status["blocked_extensions"] = len(m.conf.Extension.Security.BlockedExtensions)
		status["allow_unsafe"] = m.conf.Extension.Security.AllowUnsafe
	}

	return status
}
