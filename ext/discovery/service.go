package discovery

import (
	"context"
	"fmt"
	"ncore/ext/core"
	"ncore/pkg/logger"
	"ncore/pkg/uuid"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
)

// ServiceCache service cache implementation
type ServiceCache struct {
	services   map[string]*api.AgentService
	mu         sync.RWMutex
	ttl        time.Duration
	lastUpdate time.Time
}

// NewServiceCache creates a new service cache
func NewServiceCache(ttl time.Duration) *ServiceCache {
	return &ServiceCache{
		services: make(map[string]*api.AgentService),
		ttl:      ttl,
	}
}

// Get gets a service
func (sc *ServiceCache) Get(name string) (*api.AgentService, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if time.Since(sc.lastUpdate) > sc.ttl {
		return nil, false
	}

	svc, ok := sc.services[name]
	return svc, ok
}

// Update updates the cache
func (sc *ServiceCache) Update(services map[string]*api.AgentService) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.services = services
	sc.lastUpdate = time.Now()
}

// Clear clears the cache
func (sc *ServiceCache) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.services = make(map[string]*api.AgentService)
	sc.lastUpdate = time.Time{}
}

// GetStats returns the cache stats
func (sc *ServiceCache) GetStats() map[string]any {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return map[string]any{
		"size":        len(sc.services),
		"last_update": sc.lastUpdate,
		"ttl":         sc.ttl,
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
func (sd *ServiceDiscovery) RegisterService(name string, info *core.ServiceInfo) error {
	if sd.consul == nil {
		// Consul is not configured
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
		return fmt.Errorf("failed to register service: %w", err)
	}

	// clear cache
	sd.serviceCache.Clear()

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
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	// clear cache
	sd.serviceCache.Clear()

	logger.Infof(context.Background(),
		"service deregistered successfully: %s",
		name)

	return nil
}

// GetService gets a service from Consul
func (sd *ServiceDiscovery) GetService(name string) (*api.AgentService, error) {
	if sd.consul == nil {
		return nil, fmt.Errorf("consul client not initialized")
	}

	// cache hit
	if svc, ok := sd.serviceCache.Get(name); ok {
		return svc, nil
	}

	// cache miss, get from Consul
	services, err := sd.consul.Agent().Services()
	if err != nil {
		return nil, fmt.Errorf("failed to get services from consul: %w", err)
	}

	// update cache
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
		return core.ServiceStatusUnknown
	}

	checks, _, err := sd.consul.Health().Checks(name, &api.QueryOptions{})
	if err != nil {
		logger.Errorf(context.Background(),
			"failed to get health checks for service %s: %v",
			name,
			err)
		return core.ServiceStatusUnknown
	}

	for _, check := range checks {
		if check.Status != "passing" {
			logger.Warnf(context.Background(),
				"service %s health check failed: %s",
				name,
				check.Output)
			return core.ServiceStatusUnhealthy
		}
	}

	return core.ServiceStatusHealthy
}

// GetHealthyServices gets healthy services
func (sd *ServiceDiscovery) GetHealthyServices(name string) ([]*api.ServiceEntry, error) {
	if sd.consul == nil {
		return nil, fmt.Errorf("consul client not initialized")
	}

	services, _, err := sd.consul.Health().Service(name, "", true, &api.QueryOptions{})
	if err != nil {
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

// GetCacheStats returns the service cache stats
func (sd *ServiceDiscovery) GetCacheStats() map[string]any {
	return sd.serviceCache.GetStats()
}
