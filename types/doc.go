// Package types provides common type definitions, constants, and interfaces
// used throughout the ncore framework and extensions.
//
// This package includes:
//   - Generic type aliases (JSON, JSONArray, StringArray)
//   - Extension lifecycle interfaces
//   - Service discovery types
//   - Event system types
//   - Platform-specific constants
//   - Common data structures
//
// # Type Aliases
//
// Convenient type aliases for common patterns:
//
//	type JSON = map[string]any           // Generic JSON object
//	type JSONArray = []JSON              // Array of JSON objects
//	type StringArray = []string          // String slice
//	type AnyArray = []any                // Generic slice
//
// Usage:
//
//	data := types.JSON{
//	    "name": "John",
//	    "age":  30,
//	}
//
//	items := types.JSONArray{
//	    {"id": 1, "name": "Item 1"},
//	    {"id": 2, "name": "Item 2"},
//	}
//
// # Extension Lifecycle
//
// Extensions implement these interfaces for lifecycle management:
//
//	type Extension interface {
//	    Name() string
//	    Version() string
//	    Init() error
//	    Start() error
//	    Stop() error
//	}
//
//	type OptionalLifecycle interface {
//	    PreInit() error   // Before Init
//	    PostInit() error  // After Init
//	    PreStart() error  // Before Start
//	    PostStart() error // After Start
//	    PreStop() error   // Before Stop
//	    PostStop() error  // After Stop
//	}
//
// # Service Discovery
//
// Types for service registration and discovery:
//
//	type ServiceInfo struct {
//	    ID      string
//	    Name    string
//	    Address string
//	    Port    int
//	    Tags    []string
//	    Meta    map[string]string
//	    Health  ServiceHealth
//	}
//
//	type ServiceHealth struct {
//	    Status      string
//	    LastCheck   time.Time
//	    FailureRate float64
//	}
//
// # Event System
//
// Event routing and handling types:
//
//	type EventTarget int
//	const (
//	    EventTargetLocal  EventTarget = 1  // Local only
//	    EventTargetRemote EventTarget = 2  // Remote only
//	    EventTargetAll    EventTarget = 3  // Both
//	)
//
//	type Event struct {
//	    Type      string
//	    Payload   any
//	    Target    EventTarget
//	    Timestamp time.Time
//	}
//
// # Platform Constants
//
// Platform-specific file extensions and identifiers:
//
//	const (
//	    ExtDarwin  = ".dylib"  // macOS
//	    ExtLinux   = ".so"     // Linux
//	    ExtWindows = ".dll"    // Windows
//	)
//
// Usage:
//
//	ext := types.ExtDarwin
//	if runtime.GOOS == "linux" {
//	    ext = types.ExtLinux
//	}
//	pluginPath := "plugin" + ext
//
// # Dependency Management
//
// Define extension dependencies:
//
//	type DependencyType string
//	const (
//	    StrongDependency DependencyType = "strong"  // Required
//	    WeakDependency   DependencyType = "weak"    // Optional
//	)
//
//	type Dependency struct {
//	    Name    string
//	    Version string
//	    Type    DependencyType
//	}
//
// # Select Options
//
// Standard structure for dropdown/select options:
//
//	type SelectOption struct {
//	    Label string
//	    Value string
//	    Icon  string
//	}
//
//	options := []types.SelectOption{
//	    {Label: "Active", Value: "active", Icon: "check"},
//	    {Label: "Inactive", Value: "inactive", Icon: "x"},
//	}
//
// # Best Practices
//
//   - Use type aliases for consistency across codebase
//   - Implement all required lifecycle methods in extensions
//   - Use strong dependencies for critical components
//   - Define clear event types for pub/sub patterns
//   - Leverage SelectOption for UI consistency
package types
