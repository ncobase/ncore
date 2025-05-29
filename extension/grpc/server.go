package grpc

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ncobase/ncore/logging/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Server wraps gRPC server with extension support
type Server struct {
	server   *grpc.Server
	listener net.Listener
	address  string
	services map[string]any
	mu       sync.RWMutex
	health   *health.Server
}

// NewServer creates a new gRPC server
func NewServer(address string, opts ...grpc.ServerOption) (*Server, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %v", address, err)
	}

	defaultOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(unaryInterceptor),
		grpc.StreamInterceptor(streamInterceptor),
	}
	opts = append(defaultOpts, opts...)

	server := grpc.NewServer(opts...)
	healthServer := health.NewServer()

	grpc_health_v1.RegisterHealthServer(server, healthServer)
	reflection.Register(server)

	return &Server{
		server:   server,
		listener: listener,
		address:  listener.Addr().String(),
		services: make(map[string]any),
		health:   healthServer,
	}, nil
}

// RegisterService registers a gRPC service
func (s *Server) RegisterService(name string, service any, registrar func(*grpc.Server, any)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	registrar(s.server, service)
	s.services[name] = service
	s.health.SetServingStatus(name, grpc_health_v1.HealthCheckResponse_SERVING)

	logger.Infof(context.Background(), "gRPC service registered: %s", name)
}

// Start starts the gRPC server
func (s *Server) Start() error {
	logger.Infof(context.Background(), "starting gRPC server on %s", s.address)
	return s.server.Serve(s.listener)
}

// Stop gracefully stops the server
func (s *Server) Stop(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.server.Stop()
		return ctx.Err()
	}
}

// GetAddress returns server address
func (s *Server) GetAddress() string {
	return s.address
}

// unaryInterceptor provides logging for unary calls
func unaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	if err != nil {
		logger.Errorf(ctx, "gRPC unary call failed: %s, duration: %v, error: %v", info.FullMethod, duration, err)
	} else {
		logger.Debugf(ctx, "gRPC unary call: %s, duration: %v", info.FullMethod, duration)
	}

	return resp, err
}

// streamInterceptor provides logging for stream calls
func streamInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	start := time.Now()
	err := handler(srv, ss)
	duration := time.Since(start)

	if err != nil {
		logger.Errorf(ss.Context(), "gRPC stream call failed: %s, duration: %v, error: %v", info.FullMethod, duration, err)
	} else {
		logger.Debugf(ss.Context(), "gRPC stream call: %s, duration: %v", info.FullMethod, duration)
	}

	return err
}
