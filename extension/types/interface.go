package types

import (
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/ncobase/ncore/config"
)

// Handler represents the handler for an extension
type Handler any

// HandlerImpl represents the default implementation of the Handler interface
type HandlerImpl struct{}

// Service represents the service for an extension
type Service any

// ServiceImpl represents the default implementation of the Service interface
type ServiceImpl struct{}

// Interface defines the structure for an extension (Plugin / Module)
type Interface interface {
	// Name returns the name of the extension
	Name() string
	// Init initializes the extension with the given config
	Init(conf *config.Config, m ManagerInterface) error
	// GetHandlers returns the handlers for the extension
	GetHandlers() Handler
	// GetServices returns the services for the extension
	GetServices() Service
	// GetMetadata returns the metadata of the extension
	GetMetadata() Metadata
	// Version returns the version of the extension
	Version() string
	// Dependencies returns the strong dependencies of the extension
	Dependencies() []string
	// OptionalMethods returns the optional methods of the extension
	OptionalMethods
}

// OptionalMethods represents the optional methods for an extension
type OptionalMethods interface {
	// GetAllDependencies returns all dependencies with their types (Optional)
	GetAllDependencies() []DependencyEntry
	// PreInit performs any necessary setup before initialization
	PreInit() error
	// PostInit performs any necessary setup after initialization
	PostInit() error
	// RegisterRoutes registers routes for the extension (optional)
	RegisterRoutes(router *gin.RouterGroup)
	// GetPublisher returns an event publisher interface if the module supports publishing events
	GetPublisher() any
	// GetSubscriber returns an event subscriber interface if the module supports subscribing to events
	GetSubscriber() any
	// PreCleanup performs any necessary cleanup before the main cleanup
	PreCleanup() error
	// Cleanup cleans up the extension
	Cleanup() error
	// Status returns the status of the extension
	Status() string
	// NeedServiceDiscovery returns if the extension needs to be registered as a service
	NeedServiceDiscovery() bool
	// GetServiceInfo returns service registration info if NeedServiceDiscovery returns true
	GetServiceInfo() *ServiceInfo
}

// Wrapper wraps an Interface instance with its metadata
type Wrapper struct {
	// Metadata is the metadata of the extension
	Metadata Metadata `json:"metadata"`
	// Instance is the instance of the extension
	Instance Interface `json:"instance,omitempty"`
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

// ManagerInterface defines the interface for extension manager operations
type ManagerInterface interface {
	GetConfig() *config.Config
	Register(ext Interface) error
	InitExtensions() error
	GetExtension(name string) (Interface, error)
	GetExtensions() map[string]*Wrapper
	Cleanup()

	// Handler & Service access

	GetHandler(name string) (Handler, error)
	GetHandlers() map[string]Handler
	GetService(name string) (Service, error)
	GetServices() map[string]Service
	GetMetadata() map[string]Metadata
	GetStatus() map[string]string

	// Cross service methods

	GetCrossService(extensionName, servicePath string) (any, error)
	RegisterCrossService(key string, service any)
	RefreshCrossServices()

	// Plugin management

	LoadPlugins() error
	LoadPlugin(path string) error
	ReloadPlugin(name string) error
	UnloadPlugin(name string) error
	ReloadPlugins() error

	// Events

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

	// HTTP

	RegisterRoutes(router *gin.Engine)
	ManageRoutes(router *gin.RouterGroup)

	// Circuit breaker

	ExecuteWithCircuitBreaker(extensionName string, fn func() (any, error)) (any, error)

	// Message queue

	PublishMessage(exchange, routingKey string, body []byte) error
	SubscribeToMessages(queue string, handler func([]byte) error) error

	// Metrics

	GetEventBusMetrics() map[string]any
}
