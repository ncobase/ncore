package grpc

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/consul/api"
	"github.com/ncobase/ncore/logging/logger"
	"google.golang.org/grpc"
)

// ServiceInfo contains gRPC service information
type ServiceInfo struct {
	Name    string
	Address string
	Port    int
	Tags    []string
	Meta    map[string]string
}

// ServiceRegistry handles gRPC service registration and discovery
type ServiceRegistry struct {
	consul     *api.Client
	clientPool *ClientPool
	services   map[string]*ServiceInfo
	mu         sync.RWMutex
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry(consulAddr string) (*ServiceRegistry, error) {
	var consulClient *api.Client

	if consulAddr != "" {
		config := api.DefaultConfig()
		config.Address = consulAddr

		var err error
		consulClient, err = api.NewClient(config)
		if err != nil {
			logger.Warnf(context.Background(), "failed to create consul client: %v", err)
		}
	}

	return &ServiceRegistry{
		consul:     consulClient,
		clientPool: NewClientPool(),
		services:   make(map[string]*ServiceInfo),
	}, nil
}

// RegisterService registers a gRPC service
func (r *ServiceRegistry) RegisterService(ctx context.Context, info *ServiceInfo) error {
	r.mu.Lock()
	r.services[info.Name] = info
	r.mu.Unlock()

	if r.consul != nil {
		serviceReg := &api.AgentServiceRegistration{
			ID:      fmt.Sprintf("%s-%s", info.Name, info.Address),
			Name:    info.Name,
			Address: info.Address,
			Port:    info.Port,
			Tags:    append(info.Tags, "grpc"),
			Meta:    info.Meta,
			Check: &api.AgentServiceCheck{
				GRPC:     fmt.Sprintf("%s:%d", info.Address, info.Port),
				Interval: "10s",
				Timeout:  "3s",
			},
		}

		if err := r.consul.Agent().ServiceRegister(serviceReg); err != nil {
			return fmt.Errorf("failed to register service %s: %v", info.Name, err)
		}

		logger.Infof(ctx, "gRPC service %s registered with consul", info.Name)
	}

	return nil
}

// DiscoverService discovers a service and returns its address
func (r *ServiceRegistry) DiscoverService(ctx context.Context, serviceName string) (string, error) {
	// Check local registry first
	r.mu.RLock()
	if info, exists := r.services[serviceName]; exists {
		r.mu.RUnlock()
		return fmt.Sprintf("%s:%d", info.Address, info.Port), nil
	}
	r.mu.RUnlock()

	// Try consul if available
	if r.consul != nil {
		services, _, err := r.consul.Health().Service(serviceName, "grpc", true, nil)
		if err != nil {
			return "", fmt.Errorf("failed to discover service %s: %v", serviceName, err)
		}

		if len(services) == 0 {
			return "", fmt.Errorf("service %s not found", serviceName)
		}

		service := services[0].Service
		return fmt.Sprintf("%s:%d", service.Address, service.Port), nil
	}

	return "", fmt.Errorf("service %s not found", serviceName)
}

// GetConnection gets a gRPC connection to a service
func (r *ServiceRegistry) GetConnection(ctx context.Context, serviceName string) (*grpc.ClientConn, error) {
	address, err := r.DiscoverService(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	return r.clientPool.GetConnection(ctx, serviceName, address)
}

// Close closes the registry and all connections
func (r *ServiceRegistry) Close() {
	r.clientPool.Close()
}
