package extension

import (
	"context"
	"fmt"
	"ncore/pkg/logger"
	"ncore/pkg/uuid"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
)

// Service Status constants
const (
	ServiceStatusHealthy   = "healthy"
	ServiceStatusUnhealthy = "unhealthy"
	ServiceStatusUnknown   = "unknown"
)

// serviceCache service cache implementation
type serviceCache struct {
	services   map[string]*api.AgentService
	mu         sync.RWMutex
	ttl        time.Duration
	lastUpdate time.Time
}

// newServiceCache creates a new service cache
func newServiceCache(ttl time.Duration) *serviceCache {
	return &serviceCache{
		services: make(map[string]*api.AgentService),
		ttl:      ttl,
	}
}

// get gets a service
func (sc *serviceCache) get(name string) (*api.AgentService, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if time.Since(sc.lastUpdate) > sc.ttl {
		return nil, false
	}

	svc, ok := sc.services[name]
	return svc, ok
}

// update updates the cache
func (sc *serviceCache) update(services map[string]*api.AgentService) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.services = services
	sc.lastUpdate = time.Now()
}

// clear clears the cache
func (sc *serviceCache) clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.services = make(map[string]*api.AgentService)
	sc.lastUpdate = time.Time{}
}

// RegisterConsulService registers a service with Consul
func (m *Manager) RegisterConsulService(name string, info *ServiceInfo) error {
	if m.conf.Consul == nil || m.consul == nil {
		// Consul is not configured
		// log.Info(context.Background(), "Consul is not configured")
		return nil
	}

	if info == nil || info.Address == "" {
		return fmt.Errorf("invalid service info")
	}

	registration := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%s", name, uuid.New().String()[:8]),
		Name:    name,
		Address: info.Address,
		Tags:    info.Tags,
		Meta:    info.Meta,
	}

	if m.conf.Consul.Discovery.HealthCheck {
		healthCheckURL := fmt.Sprintf("%s://%s/health",
			m.conf.Consul.Scheme,
			info.Address)

		registration.Check = &api.AgentServiceCheck{
			HTTP:                           healthCheckURL,
			Interval:                       m.conf.Consul.Discovery.CheckInterval,
			Timeout:                        m.conf.Consul.Discovery.Timeout,
			DeregisterCriticalServiceAfter: "5m",
			Status:                         "passing",
			TLSSkipVerify:                  true,
		}
	}

	if err := m.consul.Agent().ServiceRegister(registration); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	// clear cache
	m.serviceCache.clear()

	logger.Infof(context.Background(),
		"service registered successfully: %s, address: %s",
		name,
		info.Address)

	return nil
}

// DeregisterConsulService deregisters a service from Consul
func (m *Manager) DeregisterConsulService(name string) error {
	if m.conf.Consul == nil || m.consul == nil {
		return nil
	}

	if err := m.consul.Agent().ServiceDeregister(name); err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	// clear cache
	m.serviceCache.clear()

	logger.Infof(context.Background(),
		"service deregistered successfully: %s",
		name)

	return nil
}

// GetConsulService gets a service from Consul
func (m *Manager) GetConsulService(name string) (*api.AgentService, error) {
	// cache hit
	if svc, ok := m.serviceCache.get(name); ok {
		return svc, nil
	}

	// cache miss, get from Consul
	services, err := m.consul.Agent().Services()
	if err != nil {
		return nil, fmt.Errorf("failed to get services from consul: %w", err)
	}

	// update cache
	m.serviceCache.update(services)

	service, ok := services[name]
	if !ok {
		return nil, fmt.Errorf("service %s not found in Consul", name)
	}

	return service, nil
}

// CheckServiceHealth checks service health
func (m *Manager) CheckServiceHealth(name string) string {
	if m.consul == nil {
		return ServiceStatusUnknown
	}

	checks, _, err := m.consul.Health().Checks(name, &api.QueryOptions{})
	if err != nil {
		logger.Errorf(context.Background(),
			"failed to get health checks for service %s: %v",
			name,
			err)
		return ServiceStatusUnknown
	}

	for _, check := range checks {
		if check.Status != "passing" {
			logger.Warnf(context.Background(),
				"service %s health check failed: %s",
				name,
				check.Output)
			return ServiceStatusUnhealthy
		}
	}

	return ServiceStatusHealthy
}

// GetHealthyServices gets healthy services
func (m *Manager) GetHealthyServices(name string) ([]*api.ServiceEntry, error) {
	if m.consul == nil {
		return nil, fmt.Errorf("consul client not initialized")
	}

	services, _, err := m.consul.Health().Service(name, "", true, &api.QueryOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get healthy services: %w", err)
	}

	return services, nil
}

// SetServiceCacheTTL sets the service cache TTL
func (m *Manager) SetServiceCacheTTL(ttl time.Duration) {
	m.serviceCache.mu.Lock()
	defer m.serviceCache.mu.Unlock()
	m.serviceCache.ttl = ttl
}

// ClearServiceCache clears the service cache
func (m *Manager) ClearServiceCache() {
	m.serviceCache.clear()
}

// GetServiceCacheStats returns the service cache stats
func (m *Manager) GetServiceCacheStats() map[string]any {
	m.serviceCache.mu.RLock()
	defer m.serviceCache.mu.RUnlock()

	return map[string]any{
		"size":        len(m.serviceCache.services),
		"last_update": m.serviceCache.lastUpdate,
		"ttl":         m.serviceCache.ttl,
	}
}
