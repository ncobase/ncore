package discovery

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/uuid"
)

// ServiceCache service cache implementation
type ServiceCache struct {
	services   map[string]*api.AgentService
	mu         sync.RWMutex
	ttl        time.Duration
	lastUpdate time.Time
	metrics    struct {
		hits      atomic.Int64
		misses    atomic.Int64
		updates   atomic.Int64
		evictions atomic.Int64
	}
}

// NewServiceCache creates a new service cache
func NewServiceCache(ttl time.Duration) *ServiceCache {
	return &ServiceCache{
		services: make(map[string]*api.AgentService),
		ttl:      ttl,
	}
}

// Get gets a service from cache
func (sc *ServiceCache) Get(name string) (*api.AgentService, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if time.Since(sc.lastUpdate) > sc.ttl {
		sc.metrics.misses.Add(1)
		return nil, false
	}

	svc, ok := sc.services[name]
	if ok {
		sc.metrics.hits.Add(1)
	} else {
		sc.metrics.misses.Add(1)
	}

	return svc, ok
}

// Update updates the cache
func (sc *ServiceCache) Update(services map[string]*api.AgentService) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	oldSize := len(sc.services)
	sc.services = services
	sc.lastUpdate = time.Now()
	sc.metrics.updates.Add(1)

	// Track evictions (services removed from cache)
	newSize := len(sc.services)
	if newSize < oldSize {
		sc.metrics.evictions.Add(int64(oldSize - newSize))
	}
}

// Clear clears the cache
func (sc *ServiceCache) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	evicted := int64(len(sc.services))
	sc.services = make(map[string]*api.AgentService)
	sc.lastUpdate = time.Time{}
	sc.metrics.evictions.Add(evicted)
}

// GetStats returns comprehensive cache statistics
func (sc *ServiceCache) GetStats() map[string]any {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	hits := sc.metrics.hits.Load()
	misses := sc.metrics.misses.Load()
	total := hits + misses

	var hitRate float64
	if total > 0 {
		hitRate = (float64(hits) / float64(total)) * 100.0
	}

	return map[string]any{
		"size":         len(sc.services),
		"ttl_seconds":  sc.ttl.Seconds(),
		"last_update":  sc.lastUpdate,
		"cache_hits":   hits,
		"cache_misses": misses,
		"hit_rate":     hitRate,
		"updates":      sc.metrics.updates.Load(),
		"evictions":    sc.metrics.evictions.Load(),
		"age_seconds":  time.Since(sc.lastUpdate).Seconds(),
		"is_expired":   time.Since(sc.lastUpdate) > sc.ttl,
	}
}

// ConsulConfig holds consul configuration
type ConsulConfig struct {
	Address   string
	Scheme    string
	Discovery struct {
		HealthCheck   bool
		CheckInterval string
		Timeout       string
	}
}

// ServiceDiscovery handles service registration and discovery
type ServiceDiscovery struct {
	consul       *api.Client
	serviceCache *ServiceCache
	config       *ConsulConfig
	metrics      struct {
		registrations   atomic.Int64
		deregistrations atomic.Int64
		lookups         atomic.Int64
		healthChecks    atomic.Int64
		errors          atomic.Int64
	}
}

// NewServiceDiscovery creates a new service discovery instance
func NewServiceDiscovery(config *ConsulConfig) (*ServiceDiscovery, error) {
	if config == nil {
		return nil, nil
	}

	consulConfig := api.DefaultConfig()
	consulConfig.Address = config.Address
	consulConfig.Scheme = config.Scheme

	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Consul client: %v", err)
	}

	return &ServiceDiscovery{
		consul:       consulClient,
		serviceCache: NewServiceCache(30 * time.Second),
		config:       config,
	}, nil
}

// RegisterService registers a service with Consul
func (sd *ServiceDiscovery) RegisterService(name string, info *types.ServiceInfo) error {
	if sd.consul == nil {
		return nil
	}

	if info == nil || info.Address == "" {
		sd.metrics.errors.Add(1)
		return fmt.Errorf("invalid service info")
	}

	registration := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%s", name, uuid.New().String()[:8]),
		Name:    name,
		Address: info.Address,
		Tags:    info.Tags,
		Meta:    info.Meta,
	}

	if sd.config.Discovery.HealthCheck {
		healthCheckURL := fmt.Sprintf("%s://%s/health",
			sd.config.Scheme,
			info.Address)

		registration.Check = &api.AgentServiceCheck{
			HTTP:                           healthCheckURL,
			Interval:                       sd.config.Discovery.CheckInterval,
			Timeout:                        sd.config.Discovery.Timeout,
			DeregisterCriticalServiceAfter: "5m",
			Status:                         "passing",
			TLSSkipVerify:                  true,
		}
	}

	if err := sd.consul.Agent().ServiceRegister(registration); err != nil {
		sd.metrics.errors.Add(1)
		return fmt.Errorf("failed to register service: %w", err)
	}

	sd.serviceCache.Clear()
	sd.metrics.registrations.Add(1)

	logger.Infof(context.Background(),
		"service registered successfully: %s, address: %s",
		name,
		info.Address)

	return nil
}

// DeregisterService deregisters a service from Consul
func (sd *ServiceDiscovery) DeregisterService(name string) error {
	if sd.consul == nil {
		return nil
	}

	if err := sd.consul.Agent().ServiceDeregister(name); err != nil {
		sd.metrics.errors.Add(1)
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	sd.serviceCache.Clear()
	sd.metrics.deregistrations.Add(1)

	logger.Infof(context.Background(),
		"service deregistered successfully: %s",
		name)

	return nil
}

// GetService gets a service from Consul
func (sd *ServiceDiscovery) GetService(name string) (*api.AgentService, error) {
	if sd.consul == nil {
		sd.metrics.errors.Add(1)
		return nil, fmt.Errorf("consul client not initialized")
	}

	sd.metrics.lookups.Add(1)

	// Try cache first
	if svc, ok := sd.serviceCache.Get(name); ok {
		return svc, nil
	}

	// Cache miss, fetch from Consul
	services, err := sd.consul.Agent().Services()
	if err != nil {
		sd.metrics.errors.Add(1)
		return nil, fmt.Errorf("failed to get services from consul: %w", err)
	}

	sd.serviceCache.Update(services)

	service, ok := services[name]
	if !ok {
		return nil, fmt.Errorf("service %s not found in Consul", name)
	}

	return service, nil
}

// CheckServiceHealth checks service health
func (sd *ServiceDiscovery) CheckServiceHealth(name string) string {
	if sd.consul == nil {
		return types.ServiceStatusUnknown
	}

	sd.metrics.healthChecks.Add(1)

	checks, _, err := sd.consul.Health().Checks(name, &api.QueryOptions{})
	if err != nil {
		sd.metrics.errors.Add(1)
		logger.Errorf(context.Background(),
			"failed to get health checks for service %s: %v",
			name,
			err)
		return types.ServiceStatusUnknown
	}

	for _, check := range checks {
		if check.Status != "passing" {
			logger.Warnf(context.Background(),
				"service %s health check failed: %s",
				name,
				check.Output)
			return types.ServiceStatusUnhealthy
		}
	}

	return types.ServiceStatusHealthy
}

// GetHealthyServices gets healthy services
func (sd *ServiceDiscovery) GetHealthyServices(name string) ([]*api.ServiceEntry, error) {
	if sd.consul == nil {
		sd.metrics.errors.Add(1)
		return nil, fmt.Errorf("consul client not initialized")
	}

	sd.metrics.lookups.Add(1)

	services, _, err := sd.consul.Health().Service(name, "", true, &api.QueryOptions{})
	if err != nil {
		sd.metrics.errors.Add(1)
		return nil, fmt.Errorf("failed to get healthy services: %w", err)
	}

	return services, nil
}

// SetCacheTTL sets the service cache TTL
func (sd *ServiceDiscovery) SetCacheTTL(ttl time.Duration) {
	sd.serviceCache.mu.Lock()
	defer sd.serviceCache.mu.Unlock()
	sd.serviceCache.ttl = ttl
}

// ClearCache clears the service cache
func (sd *ServiceDiscovery) ClearCache() {
	sd.serviceCache.Clear()
}

// GetCacheStats returns comprehensive cache and discovery metrics
func (sd *ServiceDiscovery) GetCacheStats() map[string]any {
	cacheStats := sd.serviceCache.GetStats()

	// Add service discovery specific metrics
	cacheStats["registrations"] = sd.metrics.registrations.Load()
	cacheStats["deregistrations"] = sd.metrics.deregistrations.Load()
	cacheStats["lookups"] = sd.metrics.lookups.Load()
	cacheStats["health_checks"] = sd.metrics.healthChecks.Load()
	cacheStats["errors"] = sd.metrics.errors.Load()

	return cacheStats
}
