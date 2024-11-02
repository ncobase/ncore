package extension

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

// RegisterConsulService registers a service with Consul
func (m *Manager) RegisterConsulService(name string, address string, port int) error {
	return m.consul.Agent().ServiceRegister(&api.AgentServiceRegistration{
		Name:    name,
		Address: address,
		Port:    port,
	})
}

// DeregisterConsulService deregisters a service from Consul
func (m *Manager) DeregisterConsulService(name string) error {
	if m.conf.Consul == nil || m.consul == nil {
		// Consul is not configured
		// log.Info(context.Background(), "Consul is not configured")
		return nil
	}
	return m.consul.Agent().ServiceDeregister(name)
}

// GetConsulService returns a service from Consul
func (m *Manager) GetConsulService(name string) (*api.AgentService, error) {
	services, err := m.consul.Agent().Services()
	if err != nil {
		return nil, err
	}
	service, ok := services[name]
	if !ok {
		return nil, fmt.Errorf("service %s not found in Consul", name)
	}
	return service, nil
}
