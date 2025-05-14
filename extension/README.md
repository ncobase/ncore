# NCore Extension System

> A flexible and robust extension system that provides dynamic loading, lifecycle management, dependency handling, and inter-module communication capabilities.

## Overview

The Extension System is designed to provide a plugin architecture that allows for:

- Dynamic loading/unloading of extensions
- Automatic extension self-registration
- Smart dependency management with support for weak dependencies
- Service discovery and registration
- Event-driven communication between extensions
- Lifecycle management
- Health monitoring and circuit breaking

## Architecture

```plaintext
├── discovery/          # Service discovery
│   └── service.go      # Service discovery implementation
├── event/              # Event
│   └── event_bus.go    # Event bus implementation
├── manager/            # Extension manager
│   ├── discovery.go    # Service discovery methods
│   ├── http.go         # HTTP routing
│   ├── manager.go      # Main manager implementation
│   ├── message.go      # Messaging functionality
│   ├── plugin.go       # Plugin management
│   └── utils.go        # Dependency resolution
├── plugin/             # Plugin
│   └── plugin.go       # Plugin loading/management
├── registry/           # Extension registry
│   └── registry.go     # Self-registration system
├── types/               # Core interfaces and types
│   ├── dependency.go   # Dependency type definitions
│   ├── event.go        # Event data structures
│   ├── extension.go    # Extension metadata
│   ├── interface.go    # Interface definitions
│   ├── optional_impl.go# Default implementation of optional methods
│   ├── service.go      # Service information
│   └── status.go       # Extension status constants
└── README.md           # This file
```

## Core Components

### 1. Core Interfaces and Types

The `types` package defines the essential interfaces and types for the extension system:

- `Interface`: The main extension interface
- `OptionalMethods`: Optional methods extensions can implement
- `ManagerInterface`: Manager functionality for extensions to use
- `DependencyType`/`DependencyEntry`: Represents dependencies with their types (strong or weak)
- `EventData`: Structure for passing event data
- `ServiceInfo`: Information for service registration
- `Metadata`: Extension metadata information

### 2. Registry System

The `registry` package provides a self-registration mechanism for extensions:

```go
// In your extension's init() function
func init() {
  // Simple registration
  registry.Register(New())
  
  // Or with group
  registry.RegisterToGroup(New(), "core")
  
  // With weak dependencies
  registry.RegisterWithWeakDeps(New(), []string{"optional-module"})
  
  // Or with both group and weak dependencies
  registry.RegisterToGroupWithWeakDeps(New(), "core", []string{"optional-module"})
}
```

### 3. Dependency Management

The system supports strong and weak dependencies:

- **Strong dependencies**: Must be present and initialized before an extension
- **Weak dependencies**: Optional, extension will function without them

```go
// Define strong dependencies
func (e *MyExtension) Dependencies() []string {
  return []string{"required-module"}
}

// Define all dependencies including weak ones
func (e *MyExtension) GetAllDependencies() []ext.DependencyEntry {
  return []ext.DependencyEntry{
    {Name: "required-module", Type: ext.StrongDependency},
    {Name: "optional-module", Type: ext.WeakDependency},
  }
}
```

### 4. Event System

The `event` package provides unified event handling with automatic transport selection:

```go
// Event target options
const (
    EventTargetMemory // In-memory event bus
    EventTargetQueue // Message queue (RabbitMQ/Kafka)
    EventTargetAll   // All available targets
)

// Subscribe to events (default: message queue if available, otherwise in-memory)
manager.SubscribeEvent("user.created", func(data any) {
  eventData := data.(types.EventData)
  // Handle event
})

// Subscribe to specific sources
manager.SubscribeEvent("user.created", handler, EventTargetMemory) // In-memory only
manager.SubscribeEvent("user.created", handler, EventTargetQueue) // Message queue only

// Publish events (default: message queue if available, otherwise in-memory)
manager.PublishEvent("user.created", userData)

// Publish to specific targets
manager.PublishEvent("user.created", userData, EventTargetMemory) // In-memory only
manager.PublishEvent("user.created", userData, EventTargetQueue) // Message queue only

// Publish with retry
manager.PublishEventWithRetry("important.event", eventData, 3)
```

### 5. Service Discovery

The `discovery` package provides service discovery mechanisms using Consul.

```go
// Register a service
info := &types.ServiceInfo{
    Address: "localhost:8080",
    Tags:    []string{"api", "v1"},
    Meta:    map[string]string{"version": "1.0"},
}
err := manager.RegisterConsulService("user-service", info)

// Check service health
status := manager.CheckServiceHealth("user-service")
```

### 6. Plugin Management

The `plugin` package supports dynamic loading and unloading of plugins across platforms.

```go
// Load all plugins
err := manager.LoadPlugins()

// Load specific plugin
err := manager.LoadPlugin("./plugins/my-plugin.so")

// Reload plugin
err := manager.ReloadPlugin("my-plugin")
```

### 7. Manager

The `manager` package coordinates all extension functionality through a unified API.

```go
// Create a new manager
manager, err := manager.NewManager(config)

// Initialize extensions
err := manager.InitExtensions()

// Get an extension
ext, err := manager.GetExtension("my-extension")

// Get a service
svc, err := manager.GetService("my-service")
```

## Extension Lifecycle

An extension goes through the following phases:

1. **Registration**: Auto-registered via `init()` or manually with `manager.Register()`
2. **Pre-initialization**: `PreInit()` method
3. **Initialization**: `Init(conf, manager)` method
4. **Post-initialization**: `PostInit()` method
5. **Cleanup**: `PreCleanup()` and `Cleanup()` methods

## HTTP API Endpoints

The manager provides RESTful APIs for extension management:

```plaintext
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

- Use self-registration through the registry system
- Properly classify dependencies as strong or weak
- Handle missing optional dependencies gracefully
- Implement proper cleanup and resource management
- Follow error handling patterns
- Include proper logging and metrics

### 2. Dependency Management

- Keep strong dependencies minimal
- Use weak dependencies for optional features
- Design interfaces to break circular dependencies
- Handle graceful degradation when dependencies are missing
- Check for optional dependencies in your PostInit method

```go
// Example of graceful degradation with optional dependencies
func (m *Module) PostInit() error {
  // Required dependency
  userService, err := m.em.GetService("user")
  if err != nil {
    return fmt.Errorf("failed to get user service: %v", err)
  }
  
  // Optional dependency
  var analyticsService interface{}
  as, err := m.em.GetService("analytics")
  if err == nil && as != nil {
    analyticsService = as
    logger.Info("Analytics service available")
  } else {
    logger.Info("Analytics service not available, some features will be limited")
  }
  
  // Initialize with available services
  m.initializeService(userService, analyticsService)
  return nil
}
```

### 3. Service Discovery

- Always validate service info
- Handle registration failures gracefully
- Implement proper health checks
- Use appropriate TTL values for caching
- Monitor service health status

### 4. Event Handling

- Use appropriate event targets based on requirements:
  - Use default behavior for most cases (message queue if available, otherwise in-memory)
  - Use `EventTargetQueue` for events that need to reach all system instances
  - Use `EventTargetMemory` for high-performance, instance-specific events
- Implement retry for important events
- Keep event payloads serializable and concise
- Use clear, descriptive event names with namespacing

### 5. Resource Management

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
    HotReload bool    // Enable hot reloading
  }
  Consul *struct {
    Address    string // Consul address
    Scheme     string // HTTP/HTTPS
    Discovery struct {
      HealthCheck   bool
      CheckInterval string
      Timeout       string
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
