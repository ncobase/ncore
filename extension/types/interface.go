package types

import (
	"context"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/ncobase/ncore/config"
)

// Handler represents the handler for an extension
type Handler any

// Service represents the service for an extension
type Service any

// Interface defines the core structure for an extension
type Interface interface {
	// Core methods

	Name() string
	Version() string
	Init(conf *config.Config, m ManagerInterface) error
	GetMetadata() Metadata

	// Resource methods

	GetHandlers() Handler
	GetServices() Service

	// Dependency methods

	Dependencies() []string

	// Optional methods interface

	OptionalMethods
}

// OptionalMethods represents optional methods for an extension
type OptionalMethods interface {
	// Lifecycle methods

	PreInit() error
	PostInit() error
	PreCleanup() error
	Cleanup() error

	// Status and health

	Status() string

	// Dependency management

	GetAllDependencies() []DependencyEntry

	// Service discovery

	NeedServiceDiscovery() bool
	GetServiceInfo() *ServiceInfo

	// Event handling

	GetPublisher() any
	GetSubscriber() any

	// HTTP routing

	RegisterRoutes(router *gin.RouterGroup)
}

// Wrapper wraps an Interface instance with its metadata
type Wrapper struct {
	Metadata Metadata  `json:"metadata"`
	Instance Interface `json:"instance,omitempty"`
}

// CallStrategy defines service calling strategy
type CallStrategy int

const (
	LocalFirst  CallStrategy = iota // Try local first, fallback to remote
	RemoteFirst                     // Try remote first, fallback to local
	LocalOnly                       // Local only
	RemoteOnly                      // Remote only
)

// CallOptions defines options for service calls
type CallOptions struct {
	Strategy CallStrategy
	Timeout  time.Duration
}

// CallResult represents service call result
type CallResult struct {
	Response any
	Error    error
	IsLocal  bool
	IsRemote bool
	Duration time.Duration
}

// ManagerInterface defines the interface for extension manager operations
type ManagerInterface interface {
	// Configuration

	GetConfig() *config.Config

	// Extension management

	InitExtensions() error
	RegisterExtension(ext Interface) error
	GetExtensionByName(name string) (Interface, error)
	ListExtensions() map[string]*Wrapper

	// Handler and Service access

	GetHandlerByName(name string) (Handler, error)
	ListHandlers() map[string]Handler
	GetServiceByName(name string) (Service, error)
	ListServices() map[string]Service

	// Cross-service methods

	GetCrossService(extensionName, servicePath string) (any, error)
	RegisterCrossService(key string, service any)

	// Service calling methods

	CallService(ctx context.Context, serviceName, methodName string, req any) (*CallResult, error)
	CallServiceWithOptions(ctx context.Context, serviceName, methodName string, req any, opts *CallOptions) (*CallResult, error)

	// Plugin management

	LoadPlugins() error
	LoadPlugin(path string) error
	ReloadPlugin(name string) error
	UnloadPlugin(name string) error

	// Event handling

	GetExtensionPublisher(name string, publisherType reflect.Type) (any, error)
	GetExtensionSubscriber(name string, subscriberType reflect.Type) (any, error)
	PublishEvent(eventName string, data any, target ...EventTarget)
	PublishEventWithRetry(eventName string, data any, maxRetries int, target ...EventTarget)
	SubscribeEvent(eventName string, handler func(any), source ...EventTarget)

	// Service discovery

	RegisterConsulService(name string, info *ServiceInfo) error
	DeregisterConsulService(name string) error
	GetConsulService(name string) (*api.AgentService, error)
	CheckServiceHealth(name string) string
	GetHealthyServices(name string) ([]*api.ServiceEntry, error)
	GetServiceCacheStats() map[string]any

	// HTTP routing

	RegisterRoutes(router *gin.Engine)
	ManageRoutes(router *gin.RouterGroup)

	// Circuit breaker

	ExecuteWithCircuitBreaker(extensionName string, fn func() (any, error)) (any, error)

	// Message queue

	PublishMessage(exchange, routingKey string, body []byte) error
	SubscribeToMessages(queue string, handler func([]byte) error) error

	// Metrics and status

	GetMetadata() map[string]Metadata
	GetStatus() map[string]string
	GetEventsMetrics() map[string]any

	// Cleanup

	Cleanup()

	// Backward compatibility - deprecated methods

	Register(ext Interface) error                     // deprecated: use RegisterExtension
	GetExtension(name string) (Interface, error)      // deprecated: use GetExtensionByName
	GetExtensions() map[string]*Wrapper               // deprecated: use ListExtensions
	GetHandler(name string) (Handler, error)          // deprecated: use GetHandlerByName
	GetHandlers() map[string]Handler                  // deprecated: use ListHandlers
	GetService(extensionName string) (Service, error) // deprecated: use GetServiceByName
	GetServices() map[string]Service                  // deprecated: use ListServices
	RefreshCrossServices()                            // deprecated: automatic refresh
}

// PluginLoaderInterface defines the interface for plugin loading/unloading
type PluginLoaderInterface interface {
	LoadPlugin(path string, manager ManagerInterface) error
	UnloadPlugin(name string) error
	GetPlugin(name string) *Wrapper
	GetPlugins() map[string]*Wrapper
	RegisterPlugin(ext Interface, metadata Metadata)
	GetRegisteredPlugins() []*Wrapper
}

// ServiceDiscoveryInterface defines the interface for service discovery operations
type ServiceDiscoveryInterface interface {
	RegisterService(name string, info *ServiceInfo) error
	DeregisterService(name string) error
	GetService(name string) (*api.AgentService, error)
	CheckServiceHealth(name string) string
	GetHealthyServices(name string) ([]*api.ServiceEntry, error)
	SetCacheTTL(ttl time.Duration)
	ClearCache()
	GetCacheStats() map[string]any
}

// EventBusInterface defines the interface for event bus operations
type EventBusInterface interface {
	Subscribe(eventName string, handler func(any))
	Publish(eventName string, data any)
	PublishWithRetry(eventName string, data any, maxRetries int)
	GetMetrics() map[string]any
}

// EventTarget defines where an event should be published
type EventTarget int

const (
	EventTargetMemory EventTarget                            = 1 << iota // In-memory event bus
	EventTargetQueue                                                     // Message queue (RabbitMQ/Kafka)
	EventTargetAll    = EventTargetMemory | EventTargetQueue             // All available targets
)
