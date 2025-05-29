package manager

import (
	"context"
	"fmt"

	exgrpc "github.com/ncobase/ncore/extension/grpc"
	"github.com/ncobase/ncore/logging/logger"
)

// GRPCExtension defines interface for extensions that provide gRPC services
type GRPCExtension interface {
	RegisterGRPCServices(server *exgrpc.Server)
}

// initGRPCSupport initializes gRPC support if enabled
func (m *Manager) initGRPCSupport() error {
	if m.conf.GRPC == nil || !m.conf.GRPC.Enabled {
		return nil
	}

	// Create gRPC server
	address := fmt.Sprintf("%s:%d", m.conf.GRPC.Host, m.conf.GRPC.Port)
	server, err := exgrpc.NewServer(address)
	if err != nil {
		return fmt.Errorf("failed to create gRPC server: %v", err)
	}

	// Create service registry
	consulAddr := ""
	if m.conf.Consul != nil {
		consulAddr = m.conf.Consul.Address
	}

	registry, err := exgrpc.NewServiceRegistry(consulAddr)
	if err != nil {
		return fmt.Errorf("failed to create gRPC registry: %v", err)
	}

	m.grpcServer = server
	m.grpcRegistry = registry

	// Register services from extensions
	m.registerGRPCServices()

	// Start server in background
	go func() {
		if err := server.Start(); err != nil {
			logger.Errorf(context.Background(), "gRPC server error: %v", err)
		}
	}()

	logger.Infof(context.Background(), "gRPC server started on %s", server.GetAddress())
	return nil
}

// registerGRPCServices registers gRPC services from extensions
func (m *Manager) registerGRPCServices() {
	for name, wrapper := range m.extensions {
		if grpcExt, ok := wrapper.Instance.(GRPCExtension); ok {
			grpcExt.RegisterGRPCServices(m.grpcServer)
			logger.Infof(context.Background(), "registered gRPC services from extension %s", name)
		}
	}
}
