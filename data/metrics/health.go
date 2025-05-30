package metrics

import (
	"context"
	"time"
)

// HealthMonitor monitors data layer component health
type HealthMonitor struct {
	collector  Collector
	components map[string]HealthChecker
}

// HealthChecker interface for health checking
type HealthChecker interface {
	Check(ctx context.Context) error
	Name() string
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(collector Collector) *HealthMonitor {
	return &HealthMonitor{
		collector:  collector,
		components: make(map[string]HealthChecker),
	}
}

// RegisterComponent registers a component for health monitoring
func (h *HealthMonitor) RegisterComponent(checker HealthChecker) {
	h.components[checker.Name()] = checker
}

// CheckAll performs health check on all registered components
func (h *HealthMonitor) CheckAll(ctx context.Context) map[string]bool {
	results := make(map[string]bool)

	for name, checker := range h.components {
		healthy := h.checkComponent(ctx, checker)
		results[name] = healthy

		if h.collector != nil {
			h.collector.HealthCheck(name, healthy)
		}
	}

	return results
}

// CheckComponent checks a specific component
func (h *HealthMonitor) CheckComponent(ctx context.Context, name string) bool {
	if checker, exists := h.components[name]; exists {
		healthy := h.checkComponent(ctx, checker)
		if h.collector != nil {
			h.collector.HealthCheck(name, healthy)
		}
		return healthy
	}
	return false
}

// checkComponent performs the actual health check
func (h *HealthMonitor) checkComponent(ctx context.Context, checker HealthChecker) bool {
	// Set a reasonable timeout for health checks
	checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := checker.Check(checkCtx)
	return err == nil
}
