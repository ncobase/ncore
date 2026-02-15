// Package grpc provides gRPC server and client infrastructure with automatic
// service discovery, health checks, and connection pooling.
//
// This package offers:
//   - Production-ready gRPC server with graceful shutdown
//   - Client connection management with load balancing
//   - Service registration and discovery (Consul integration)
//   - Health check support (gRPC Health Checking Protocol)
//   - Reflection API for development tools
//   - TLS/mTLS support for secure communication
//
// # Server Usage
//
// Create and start a gRPC server:
//
//	server, err := grpc.NewServer("localhost:50051")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Register your services
//	pb.RegisterYourServiceServer(server.GetServer(), &yourService{})
//
//	// Start server
//	if err := server.Start(); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Graceful shutdown
//	defer server.Stop()
//
// # Client Usage
//
// Create a client connection:
//
//	client, err := grpc.NewClient("localhost:50051")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Use the connection
//	conn := client.GetConnection()
//	yourClient := pb.NewYourServiceClient(conn)
//
// # Service Discovery
//
// Register service with Consul for automatic discovery:
//
//	registry := grpc.NewRegistry(consulClient)
//	err := registry.Register(ctx, &grpc.ServiceInfo{
//	    ID:      "service-1",
//	    Name:    "user-service",
//	    Address: "localhost",
//	    Port:    50051,
//	    Tags:    []string{"v1", "production"},
//	})
//
// Discover and connect to services:
//
//	services, err := registry.Discover(ctx, "user-service")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Connect to first available instance
//	if len(services) > 0 {
//	    addr := fmt.Sprintf("%s:%d", services[0].Address, services[0].Port)
//	    client, _ := grpc.NewClient(addr)
//	}
//
// # Health Checks
//
// The server automatically registers health check service:
//
//	// Set service health status
//	server.SetServingStatus("user-service", grpc.HealthCheckResponse_SERVING)
//
//	// Check from client
//	healthClient := grpc_health_v1.NewHealthClient(conn)
//	resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{
//	    Service: "user-service",
//	})
//
// # TLS Configuration
//
// Enable TLS for secure communication:
//
//	server, err := grpc.NewServerWithTLS(
//	    "localhost:50051",
//	    "server.crt",
//	    "server.key",
//	    "ca.crt", // For mTLS
//	)
//
// # Best Practices
//
//   - Enable health checks for all production services
//   - Use service discovery for multi-instance deployments
//   - Implement proper error handling and retries
//   - Configure keepalive for long-lived connections
//   - Use TLS in production environments
//   - Enable reflection only in development
//   - Monitor connection pool metrics
//   - Implement circuit breakers for resilience
package grpc
