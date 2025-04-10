package manager

import (
	"fmt"

	"github.com/ncobase/ncore/ext/types"

	"github.com/hashicorp/consul/api"
)

// RegisterConsulService registers a service with Consul
func (m *Manager) RegisterConsulService(name string, info *types.ServiceInfo) error {
	if m.serviceDiscovery == nil {
		return nil
	}
	return m.serviceDiscovery.RegisterService(name, info)
}

// DeregisterConsulService deregisters a service from Consul
func (m *Manager) DeregisterConsulService(name string) error {
	if m.serviceDiscovery == nil {
		return nil
	}
	return m.serviceDiscovery.DeregisterService(name)
}

// GetConsulService gets a service from Consul
func (m *Manager) GetConsulService(name string) (*api.AgentService, error) {
	if m.serviceDiscovery == nil {
		return nil, fmt.Errorf("service discovery not initialized")
	}
	return m.serviceDiscovery.GetService(name)
}

// CheckServiceHealth checks service health
func (m *Manager) CheckServiceHealth(name string) string {
	if m.serviceDiscovery == nil {
		return types.ServiceStatusUnknown
	}
	return m.serviceDiscovery.CheckServiceHealth(name)
}

// GetHealthyServices gets healthy services
func (m *Manager) GetHealthyServices(name string) ([]*api.ServiceEntry, error) {
	if m.serviceDiscovery == nil {
		return nil, fmt.Errorf("service discovery not initialized")
	}
	return m.serviceDiscovery.GetHealthyServices(name)
}

// GetServiceCacheStats returns the service cache stats
func (m *Manager) GetServiceCacheStats() map[string]any {
	if m.serviceDiscovery == nil {
		return map[string]any{}
	}
	return m.serviceDiscovery.GetCacheStats()
}
