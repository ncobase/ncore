# Extension System

> A flexible and robust extension system that provides dynamic loading, lifecycle management, and inter-module communication capabilities.

## Overview

The Extension System is designed to provide a plugin architecture that allows for:

- Dynamic loading/unloading of extensions
- Service discovery and registration
- Event-driven communication between extensions
- Lifecycle management
- Health monitoring and circuit breaking

## Architecture

```plaintext
├── event_bus.go          # Event system implementation
├── interface.go          # Core interfaces
├── manager.go            # Extension manager
├── manager_plugins.go    # Plugin management
├── manager_routes.go     # HTTP routing
├── manager_utils.go      # Utilities
├── plugin.go            # Plugin system
├── README.md             # This file
└── service_discovery.go  # Service discovery
```

## Core Components

### 1. Event System

The event system provides asynchronous communication between extensions.

Features:

- Type-safe event data structure
- Retry mechanism
- Panic recovery
- Metrics collection
- Queue size monitoring

Example:

```go
// Subscribe to events
manager.SubscribeEvent("user.created", func(data any) {
    eventData := data.(EventData)
    // Handle event
})

// Publish events with retry
manager.eventBus.PublishWithRetry("user.created", userData, 3)

// Get event metrics
metrics := manager.eventBus.GetMetrics()
```

### 2. Service Discovery

Built-in service discovery mechanism using Consul.

Features:

- Service registration/deregistration
- Health checking
- Service caching
- Cache TTL management
- Status monitoring

Example:

```go
// Register a service
info := &ServiceInfo{
    Address: "localhost:8080",
    Tags:    []string{"api", "v1"},
    Meta:    map[string]string{"version": "1.0"},
}
err := manager.RegisterConsulService("user-service", info)

// Get service info with caching
service, err := manager.GetConsulService("user-service")

// Check service health
status := manager.CheckServiceHealth("user-service")
```

### 3. Plugin Management

Supports dynamic loading and unloading of plugins across platforms.

Features:

- Cross-platform support (.so, .dylib, .dll)
- Hot-reloading
- Dependency management
- Initialization ordering
- Resource cleanup

Example:

```go
// Load all plugins
err := manager.LoadPlugins()

// Load specific plugin
err := manager.loadPlugin("./plugins/my-plugin.so")

// Reload plugin
err := manager.ReloadPlugin("my-plugin")
```

## Extension Lifecycle

An extension goes through the following phases:

1. **Registration**

 ```go
 err := manager.Register(myExtension)
 ```

2. **Pre-initialization**

 ```go
 func (e *Extension) PreInit() error {
     // Setup resources
 }
 ```

3. **Initialization**

 ```go
 func (e *Extension) Init(conf *config.Config, m *Manager) error {
     // Initialize extension
 }
 ```

4. **Post-initialization**

 ```go
 func (e *Extension) PostInit() error {
     // Post-setup tasks
 }
   ```

5. **Cleanup**

 ```go
 func (e *Extension) Cleanup() error {
     // Cleanup resources
 }
   ```

## HTTP API Endpoints

The system provides RESTful APIs for extension management:

```
GET  /exts              # List all extensions
POST /exts/load        # Load an extension
POST /exts/unload      # Unload an extension
POST /exts/reload      # Reload an extension
```

## Circuit Breaking

Built-in circuit breaker for fault tolerance:

```go
result, err := manager.ExecuteWithCircuitBreaker("service-name", func() (any, error) {
    // Your code here
    return result, nil
})
```

## Best Practices

### 1. Extension Development

- Implement all interface methods
- Handle cleanup properly
- Use dependency injection
- Follow error handling patterns
- Include proper logging
- Add metrics where appropriate

### 2. Service Discovery

- Always validate service info
- Handle registration failures gracefully
- Implement proper health checks
- Use appropriate TTL values for caching
- Monitor service health status

### 3. Event Handling

- Use strongly typed event data
- Implement retry for important events
- Handle panics in event handlers
- Monitor event metrics
- Clean up event handlers

### 4. Resource Management

- Implement PreCleanup and Cleanup
- Close connections properly
- Clear caches when needed
- Handle timeouts appropriately
- Release resources in error cases

## Configuration

Example configuration:

```go
type Config struct {
    Extension struct {
        Path     string   // Plugin directory path
        Mode     string   // Plugin mode
        Includes []string // Included plugins
        Excludes []string // Excluded plugins
    }
    Consul *struct {
        Address    string // Consul address
        Scheme     string // HTTP/HTTPS
        Discovery struct {
            HealthCheck   bool
            CheckInterval string
            Timeout      string
        }
    }
}
```

## Monitoring

### Metrics Available:

- Event processing stats
- Service health status
- Cache hit rates
- Circuit breaker states
- Plugin loading status

Example:

```go
// Get event metrics
eventMetrics := manager.eventBus.GetMetrics()

// Get cache stats
cacheStats := manager.GetServiceCacheStats()
```

## Testing

Recommended testing approaches:

1. Unit tests for individual components
2. Integration tests for plugin loading
3. Event system testing
4. Service discovery testing
5. Cache behavior testing

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Create tests for new functionality
5. Create pull request

