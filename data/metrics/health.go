package metrics

import (
	"context"
	"time"
)

type HealthMonitor struct {
	collector  Collector
	components map[string]HealthChecker
}

type HealthChecker interface {
	Check(ctx context.Context) error
	Name() string
}

func NewHealthMonitor(collector Collector) *HealthMonitor {
	return &HealthMonitor{
		collector:  collector,
		components: make(map[string]HealthChecker),
	}
}

func (h *HealthMonitor) RegisterComponent(checker HealthChecker) {
	h.components[checker.Name()] = checker
}

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

func (h *HealthMonitor) checkComponent(ctx context.Context, checker HealthChecker) bool {
	checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := checker.Check(checkCtx)
	return err == nil
}
