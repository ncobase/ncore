# NCore Extension System

> A flexible and robust extension that provides dynamic loading, lifecycle management, and inter-module
> communication capabilities.

## Overview

The Extension System is designed to provide a plugin architecture that allows for:

- Dynamic loading/unloading of extensions
- Service discovery and registration
- Event-driven communication between extensions
- Lifecycle management
- Health monitoring and circuit breaking

## Architecture

```plaintext
├── core/               # Core interfaces and types
│   ├── interface.go    # Interface definitions
│   ├── optional_impl.go# Default implementation of optional methods
│   └── types.go        # Common type definitions
├── discovery/          # Service discovery
│   └── service.go      # Service discovery implementation
├── event/              # Event
│   └── event_bus.go    # Event bus implementation
├── manager/            # Extension manager
│   ├── discovery.go    # Service discovery methods
│   ├── http.go         # HTTP routing
│   ├── manager.go      # Main manager implementation
│   ├── message.go      # Messaging functionality
│   ├── methods.go      # Additional manager methods
│   ├── plugin.go       # Plugin management
│   └── utils.go        # Utility functions
├── plugin/             # Plugin
│   └── plugin.go       # Plugin loading/management
└── README.md           # This file
```

## Core Components

### 1. Core Interfaces

The core package defines the essential interfaces and types for the extension:

- `Interface`: The main extension interface
- `OptionalMethods`: Optional methods extensions can implement
- `ManagerInterface`: Manager functionality for extensions to use
- Common types like `ServiceInfo`, `Metadata`, `EventData`

### 2. Event System

The event package provides asynchronous communication between extensions.

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
    eventData := data.(core.EventData)
    // Handle event
})

// Publish events with retry
manager.PublishEvent("user.created", userData)

// Get event metrics
metrics := manager.GetEventBusMetrics()
```

### 3. Service Discovery

The discovery package provides service discovery mechanisms using Consul.

Features:

- Service registration/deregistration
- Health checking
- Service caching
- Cache TTL management
- Status monitoring

Example:

```go
// Register a service
info := &core.ServiceInfo{
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

### 4. Plugin Management

The plugin package supports dynamic loading and unloading of plugins across platforms.

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
err := manager.LoadPlugin("./plugins/my-plugin.so")

// Reload plugin
err := manager.ReloadPlugin("my-plugin")
```

### 5. Manager

The manager package coordinates all extension functionality through a unified API.

Features:

- Extension lifecycle management
- Plugin registration and management
- Event coordination
- Service discovery integration
- Circuit breaking
- HTTP API endpoints

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
 func (e *Extension) Init(conf *config.Config, m nec.ManagerInterface) error {
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

The manager provides RESTful APIs for extension management:

```
GET  /exts              # List all extensions
POST /exts/load         # Load an extension
POST /exts/unload       # Unload an extension
POST /exts/reload       # Reload an extension
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

### Metrics Available

- Event processing stats
- Service health status
- Cache hit rates
- Circuit breaker states
- Plugin loading status

Example:

```go
// Get event metrics
eventMetrics := manager.GetEventBusMetrics()

// Get cache stats
cacheStats := manager.GetServiceCacheStats()
```

## Testing

Recommended testing approaches:

1. Unit tests for individual components
2. Integration tests for plugin loading
3. Event testing
4. Service discovery testing
5. Cache behavior testing

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Create tests for new functionality
5. Create pull request
