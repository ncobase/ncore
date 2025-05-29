package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ncobase/ncore/logging/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// ClientPool manages gRPC client connections
type ClientPool struct {
	connections map[string]*grpc.ClientConn
	mu          sync.RWMutex
}

// NewClientPool creates a new client pool
func NewClientPool() *ClientPool {
	return &ClientPool{
		connections: make(map[string]*grpc.ClientConn),
	}
}

// GetConnection gets or creates a connection to the service
func (p *ClientPool) GetConnection(ctx context.Context, serviceName, address string) (*grpc.ClientConn, error) {
	p.mu.RLock()
	if conn, exists := p.connections[serviceName]; exists {
		p.mu.RUnlock()
		return conn, nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if conn, exists := p.connections[serviceName]; exists {
		return conn, nil
	}

	// Create new connection
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(clientUnaryInterceptor),
		grpc.WithStreamInterceptor(clientStreamInterceptor),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s at %s: %v", serviceName, address, err)
	}

	p.connections[serviceName] = conn
	logger.Infof(ctx, "gRPC client connected to %s at %s", serviceName, address)

	return conn, nil
}

// CheckHealth checks service health
func (p *ClientPool) CheckHealth(ctx context.Context, serviceName string) error {
	conn, exists := p.connections[serviceName]
	if !exists {
		return fmt.Errorf("no connection to service %s", serviceName)
	}

	client := grpc_health_v1.NewHealthClient(conn)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{
		Service: serviceName,
	})
	if err != nil {
		return err
	}

	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("service %s not serving: %v", serviceName, resp.Status)
	}

	return nil
}

// Close closes all connections
func (p *ClientPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for name, conn := range p.connections {
		if err := conn.Close(); err != nil {
			logger.Errorf(context.Background(), "failed to close connection to %s: %v", name, err)
		}
	}
	p.connections = make(map[string]*grpc.ClientConn)
}

// clientUnaryInterceptor provides logging for client unary calls
func clientUnaryInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	duration := time.Since(start)

	if err != nil {
		logger.Errorf(ctx, "gRPC client call failed: %s, duration: %v, error: %v", method, duration, err)
	} else {
		logger.Debugf(ctx, "gRPC client call: %s, duration: %v", method, duration)
	}

	return err
}

// clientStreamInterceptor provides logging for client stream calls
func clientStreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	start := time.Now()
	stream, err := streamer(ctx, desc, cc, method, opts...)
	duration := time.Since(start)

	if err != nil {
		logger.Errorf(ctx, "gRPC client stream failed: %s, duration: %v, error: %v", method, duration, err)
	} else {
		logger.Debugf(ctx, "gRPC client stream: %s, duration: %v", method, duration)
	}

	return stream, err
}
