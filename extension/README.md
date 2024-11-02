# Extension System

> This is the plug-in and module extension system that provides dynamic loading, lifecycle management, and inter-module communication capabilities.

## Structure

```plaintext
├── event_bus.go          # Event bus implementation
├── interface.go          # Extension interface definitions
├── manager.go            # Core manager implementation
├── manager_discovery.go  # Service discovery and registration
├── manager_plugins.go    # Plugin loading and management
├── manager_routes.go     # Route management and HTTP handlers
├── manager_utils.go      # Utility functions
├── plugin.go            # Plugin system implementation
└── README.md            # This file
```

## Component Details

### Core Components

#### event_bus.go

Implements a pub/sub event system that enables:

- Asynchronous communication between extensions
- Event-driven architecture support
- Message broadcasting capabilities

#### interface.go

Defines the core interfaces for the extension system:

- Extension lifecycle methods (Init, PreInit, PostInit, etc.)
- Service and handler interfaces
- Metadata structures

#### manager.go

Contains the core extension manager implementation:

- Extension registry management
- Circuit breaker integration
- Basic extension operations
- Resource cleanup handling

### Manager Components

#### manager_discovery.go

Handles service discovery and registration:

- Service registration and deregistration
- Service discovery interface
- Multiple discovery backend support
- Service health monitoring
- Service metadata management

Current implementations:

- Consul service discovery
- (Expandable for other discovery mechanisms)

##### Usage example

```go
// Register a service
err := manager.RegisterService("myservice", "localhost", 8080)

// Get service details
service, err := manager.GetService("myservice")

// Get all services
services, err := manager.GetServices()
```

#### manager_plugins.go

Manages plugin operations:

- Cross-platform plugin loading
- Plugin lifecycle management
- Hot-reloading capabilities
- Platform-specific extension handling

#### manager_routes.go

- Handles HTTP routing and API endpoints:

- Route registration
- Handler management
- API endpoint exposure
- Circuit breaker integration for routes

#### manager_utils.go

Provides utility functions:

- Dependency resolution
- Startup sequence management
- Configuration helpers
- Plugin filtering
- Status management
- Metadata operations

### Plugin System

#### plugin.go

Implements core plugin functionality:

- Plugin loading mechanisms
- Plugin registration
- Plugin interface implementations
- Plugin state management

## Usage

### Basic Extension Registration

```go
// Create a new extension manager
manager, err := extension.NewManager(config)
if err != nil {
log.Fatal(err)
}

// Register an extension
err = manager.Register(myExtension)
if err != nil {
log.Fatal(err)
}
```

### Plugin Loading

```go
// Load all plugins
if err := manager.LoadPlugins(); err != nil {
log.Fatal(err)
}

// Load specific plugin
if err := manager.loadPlugin("path/to/plugin.so"); err != nil {
log.Fatal(err)
}
```

### Event Communication

```go
// Subscribe to events
manager.SubscribeEvent("eventName", func(data any) {
// Handle event
})

// Publish events
manager.PublishEvent("eventName", data)
```

## Platform Support

The extension system supports multiple platforms:

- Linux (.so)
- macOS (.dylib)
- Windows (.dll)

## Best Practices

### Dependency Management

- Clearly define dependencies in extension metadata
- Avoid circular dependencies
- Use dependency injection where possible

### Error Handling

- Always check for errors when loading plugins
- Implement proper cleanup in error cases
- Use circuit breakers for external service calls

### Resource Management

- Clean up resources in PreCleanup and Cleanup methods
- Implement proper shutdown sequences
- Handle connection pooling appropriately

### Testing

- Write tests for each extension
- Mock external services in tests
- Test cross-platform compatibility

## Contributing

When contributing to the extension system:

- Follow the existing code structure
- Add appropriate documentation
- Include tests for new functionality
- Ensure cross-platform compatibility
