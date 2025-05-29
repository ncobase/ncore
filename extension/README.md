# NCore Extension System

A flexible and robust extension system that provides dynamic loading, lifecycle management, dependency handling, and inter-service communication capabilities.

## Features

- **Dynamic Loading**: Load extensions from files or built-in registration
- **Dependency Management**: Strong and weak dependency support with automatic resolution
- **Service Discovery**: Consul-based service registration and discovery
- **Event System**: Unified event handling with memory and message queue support
- **gRPC Integration**: Optional gRPC service support for distributed communication
- **Circuit Breaker**: Built-in fault tolerance for service calls
- **Hot Reload**: Runtime plugin loading and unloading
- **Cross-Service Calls**: Unified local and remote service calling interface

## Basic Usage

### Create Extension

```go
package myext

import (
    "github.com/ncobase/ncore/extension/registry"
    "github.com/ncobase/ncore/extension/types"
)

type MyExtension struct {
    types.OptionalImpl
}

func init() {
    registry.RegisterToGroupWithWeakDeps(New(), "core", []string{"user"})
}

func (m *MyExtension) Name() string { return "my-extension" }
func (m *MyExtension) Version() string { return "1.0.0" }
func (m *MyExtension) Dependencies() []string { return []string{} }

func (m *MyExtension) Init(conf *config.Config, manager types.ManagerInterface) error {
    // Initialize extension
    return nil
}

func (m *MyExtension) GetMetadata() types.Metadata {
    return types.Metadata{
        Name: m.Name(), Version: m.Version(),
        Description: "My extension", Type: "module", Group: "core",
    }
}
```

### Use Manager

```go
func main() {
    mgr, err := manager.NewManager(config)
    if err != nil { panic(err) }
    defer mgr.Cleanup()

    if err := mgr.InitExtensions(); err != nil { panic(err) }
    
    ext, err := mgr.GetExtensionByName("my-extension")
    // Use extension...
}
```

## Extension Lifecycle

Extensions follow a structured initialization process:

1. **Registration** - Auto-registered via `init()` or manually with `manager.RegisterExtension()`
2. **Dependency Resolution** - System calculates initialization order based on dependencies
3. **Pre-initialization** - `PreInit()` method for early setup
4. **Initialization** - `Init(config, manager)` method with full context
5. **Post-initialization** - `PostInit()` method for cross-extension communication
6. **Runtime** - Extension is active and serving requests
7. **Cleanup** - `PreCleanup()` and `Cleanup()` methods for resource cleanup

## Dependency Management

### Dependency Types

**Strong Dependencies** - Required and must be initialized first:

```go
func (m *MyExtension) Dependencies() []string {
    return []string{"required-module"}
}
```

**Weak Dependencies** - Optional, graceful degradation when missing:

```go
func (m *MyExtension) GetAllDependencies() []types.DependencyEntry {
    return []types.DependencyEntry{
        {Name: "required-module", Type: types.StrongDependency},
        {Name: "optional-module", Type: types.WeakDependency},
    }
}
```

### Dependency Resolution

The system automatically:

- Detects and prevents circular dependencies
- Calculates optimal initialization order
- Handles missing weak dependencies gracefully
- Provides detailed error messages for resolution failures

Extensions with weak dependencies should handle missing services:

```go
func (m *MyExtension) PostInit() error {
    if userService, err := m.manager.GetServiceByName("user"); err == nil {
        m.userService = userService
    } else {
        log.Warn("User service unavailable, limited functionality")
    }
    return nil
}
```

## Service Communication

### Service Calling Strategies

```go
// Default local-first strategy
result, err := manager.CallService(ctx, "user-service", "GetUser", userID)

// Explicit strategy
result, err := manager.CallServiceWithOptions(ctx, "user-service", "GetUser", userID, 
    &types.CallOptions{
        Strategy: types.LocalFirst,  // LocalFirst, RemoteFirst, LocalOnly, RemoteOnly
        Timeout:  30 * time.Second,
    })
```

**Strategy Behavior**:

- `LocalFirst`: Try local service, fallback to gRPC
- `RemoteFirst`: Try gRPC service, fallback to local
- `LocalOnly`: Local service only, fail if unavailable
- `RemoteOnly`: gRPC service only, fail if unavailable

### Cross-Service Access

```go
// Direct service access
userService, err := manager.GetServiceByName("user")

// Cross-service field access
authService, err := manager.GetCrossService("auth", "TokenManager")
```

## Event System

### Event Targets

The event system supports multiple transport mechanisms:

- `EventTargetMemory`: In-memory, single instance, high performance
- `EventTargetQueue`: Message queue (RabbitMQ/Kafka), distributed, persistent
- `EventTargetAll`: Both targets simultaneously

### Event Operations

```go
// Subscribe (auto-selects best transport)
manager.SubscribeEvent("user.created", func(data any) {
    eventData := data.(types.EventData)
    // Handle event
})

// Subscribe to specific transport
manager.SubscribeEvent("user.created", handler, types.EventTargetMemory)

// Publish (auto-selects best transport)
manager.PublishEvent("user.created", userData)

// Publish with retry for critical events
manager.PublishEventWithRetry("payment.failed", paymentData, 3)
```

### Event Data Structure

```go
type EventData struct {
    Time      time.Time `json:"time"`
    Source    string    `json:"source"`
    EventType string    `json:"event_type"`
    Data      any       `json:"data"`
}
```

## Configuration

```yaml
extension:
  path: "./plugins"          # Plugin directory
  mode: "file"              # "file" or "c2hlbgo" (built-in)
  includes: ["auth", "user"] # Include specific plugins
  excludes: ["debug"]       # Exclude plugins

consul:
  address: "localhost:8500"  # Consul server
  scheme: "http"
  discovery:
    health_check: true       # Enable health checks
    check_interval: "10s"    # Health check interval
    timeout: "3s"           # Health check timeout

grpc:
  enabled: true             # Enable gRPC support
  host: "localhost"
  port: 9090
```

## Advanced Features

### gRPC Integration

Extensions can provide gRPC services:

```go
func (m *MyExtension) RegisterGRPCServices(server *grpc.Server) {
    pb.RegisterMyServiceServer(server, m.grpcService)
}
```

### Service Discovery

Extensions can register with service discovery:

```go
func (m *MyExtension) NeedServiceDiscovery() bool { return true }

func (m *MyExtension) GetServiceInfo() *types.ServiceInfo {
    return &types.ServiceInfo{
        Address: "localhost:8080",
        Tags:    []string{"api", "v1"},
        Meta:    map[string]string{"version": "1.0"},
    }
}
```

### Circuit Breaker

Protect against service failures:

```go
result, err := manager.ExecuteWithCircuitBreaker("external-service", func() (any, error) {
    return callExternalAPI()
})
```

### Plugin Loading Modes

**File Mode**: Load plugins from filesystem

- Supports `.so` files on Linux, `.dll` on Windows
- Hot reload capability
- Include/exclude filtering

**Built-in Mode**: Use statically compiled extensions

- Better performance and security
- No filesystem dependencies
- Compile-time dependency resolution

## Management API

REST endpoints for runtime management:

- `GET /exts` - List all extensions with metadata
- `GET /exts/status` - Get extension status and health
- `POST /exts/load?name=plugin` - Load specific plugin
- `POST /exts/unload?name=plugin` - Unload plugin
- `POST /exts/reload?name=plugin` - Reload plugin
- `GET /exts/metrics` - System metrics and performance data

## Performance Considerations

- **Service Discovery**: Use appropriate cache TTL (default 30s)
- **Event Transport**: Choose memory for high-frequency, queue for reliability
- **Circuit Breaker**: Monitor failure rates and adjust thresholds
- **Plugin Loading**: Prefer built-in mode for production environments

## Troubleshooting

### Common Issues

**Circular Dependencies**

```
Error: cyclic dependency detected in extensions: [module-a, module-b]
```

*Solution*: Convert one dependency to weak dependency type

**Service Not Found**

```
Error: extension 'user-service' not found
```  

*Solution*: Check extension registration and initialization order

**gRPC Connection Failed**

```
Error: failed to get gRPC connection for service-name  
```

*Solution*: Verify service discovery configuration and network connectivity
