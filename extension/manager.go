package extension

import (
	"context"
	"fmt"
	"ncobase/common/config"
	"ncobase/common/data"
	"ncobase/common/log"
	"sync"

	"github.com/hashicorp/consul/api"
	"github.com/sony/gobreaker"
)

// Manager represents a extension / plugin manager
type Manager struct {
	extensions      map[string]*Wrapper
	conf            *config.Config
	mu              sync.RWMutex
	initialized     bool
	eventBus        *EventBus
	consul          *api.Client
	circuitBreakers map[string]*gobreaker.CircuitBreaker
	data            *data.Data
}

// NewManager creates a new extension / plugin manager
func NewManager(conf *config.Config) (*Manager, error) {
	d, cleanup, err := data.New(conf.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to create data connections: %v", err)
	}

	var consulClient *api.Client
	if conf.Consul == nil {
		consulConfig := api.DefaultConfig()
		consulConfig.Address = conf.Consul.Address
		consulConfig.Scheme = conf.Consul.Scheme
		consulClient, err = api.NewClient(consulConfig)
		if err != nil {
			cleanup()
			return nil, fmt.Errorf("failed to create Consul client: %v", err)
		}
	}

	return &Manager{
		extensions:      make(map[string]*Wrapper),
		conf:            conf,
		eventBus:        NewEventBus(),
		consul:          consulClient,
		circuitBreakers: make(map[string]*gobreaker.CircuitBreaker),
		data:            d,
	}, nil
}

// Register registers a extension
func (m *Manager) Register(f Interface) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("cannot register extension after initialization")
	}

	name := f.Name()
	if _, exists := m.extensions[name]; exists {
		return fmt.Errorf("extension %s already registered", name)
	}

	m.extensions[name] = &Wrapper{
		Metadata: f.GetMetadata(),
		Instance: f,
	}

	return nil
}

// InitExtensions initializes all registered extensions
func (m *Manager) InitExtensions() error {
	m.mu.Lock()
	if m.initialized {
		m.mu.Unlock()
		return fmt.Errorf("extensions already initialized")
	}
	// Check dependencies before determining initialization order
	if err := m.checkDependencies(); err != nil {
		m.mu.Unlock()
		return err
	}
	initOrder, err := getInitOrder(m.extensions)
	if err != nil {
		log.Errorf(context.Background(), "failed to determine initialization order: %v", err)
		m.mu.Unlock()
		return err
	}
	m.mu.Unlock() // Unlock after dependencies check and order determination

	// Pre-initialization
	for _, name := range initOrder {
		extension := m.extensions[name]
		if err := extension.Instance.PreInit(); err != nil {
			log.Errorf(context.Background(), "failed pre-initialization of extension %s: %v", name, err)
			continue // Skip current extension and move to the next one
		}
	}

	// Initialization
	for _, name := range initOrder {
		extension := m.extensions[name]
		if err := extension.Instance.Init(m.conf, m); err != nil {
			log.Errorf(context.Background(), "failed to initialize extension %s: %v", name, err)
			continue // Skip current extension and move to the next one
		}
	}

	// Post-initialization
	for _, name := range initOrder {
		extension := m.extensions[name]
		if err := extension.Instance.PostInit(); err != nil {
			log.Errorf(context.Background(), "failed post-initialization of extension %s: %v", name, err)
			continue // Skip current extension and move to the next one
		}
	}

	// Ensure all services are initialized
	for _, extension := range m.extensions {
		_ = extension.Instance.GetServices()
	}

	// Lock again to safely update the initialized flag
	m.mu.Lock()
	m.initialized = true
	m.mu.Unlock()

	// log.Infof(context.Background(), " All extensions initialized successfully")
	return nil
}

// GetExtension returns a specific extension
func (m *Manager) GetExtension(name string) (Interface, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	extension, exists := m.extensions[name]
	if !exists {
		return nil, fmt.Errorf("extension %s not found", name)
	}

	return extension.Instance, nil
}

// GetExtensions returns the loaded extensions
func (m *Manager) GetExtensions() map[string]*Wrapper {
	m.mu.RLock()
	defer m.mu.RUnlock()

	extensions := make(map[string]*Wrapper)
	for name, extension := range m.extensions {
		extensions[name] = extension
	}
	return extensions
}

// Cleanup cleans up all loaded extensions
func (m *Manager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, extension := range m.extensions {
		if err := extension.Instance.PreCleanup(); err != nil {
			log.Errorf(context.Background(), "failed pre-cleanup of extension %s: %v", extension.Metadata.Name, err)
		}
		if err := extension.Instance.Cleanup(); err != nil {
			log.Errorf(context.Background(), "failed to cleanup extension %s: %v", extension.Metadata.Name, err)
		}
		if err := m.DeregisterConsulService(extension.Metadata.Name); err != nil {
			log.Errorf(context.Background(), "failed to deregister service %s from Consul: %v", extension.Metadata.Name, err)
		}
	}

	m.extensions = make(map[string]*Wrapper)
	m.circuitBreakers = make(map[string]*gobreaker.CircuitBreaker)
	m.initialized = false

	if errs := m.data.Close(); len(errs) > 0 {
		log.Errorf(context.Background(), "errors closing data connections: %v", errs)
	}
}

// GetHandler returns a specific handler from a extension
func (m *Manager) GetHandler(f string) (Handler, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	extension, exists := m.extensions[f]
	if !exists {
		return nil, fmt.Errorf("extension %s not found", f)
	}

	handler := extension.Instance.GetHandlers()
	if handler == nil {
		return nil, fmt.Errorf("no handler found in extension %s", f)
	}

	return handler, nil
}

// GetHandlers returns all registered extension handlers
func (m *Manager) GetHandlers() map[string]Handler {
	m.mu.RLock()
	defer m.mu.RUnlock()
	handlers := make(map[string]Handler)
	for name, extension := range m.extensions {
		handlers[name] = extension.Instance.GetHandlers()
	}
	return handlers
}

// GetService returns a specific service from a extension
func (m *Manager) GetService(extensionName string) (Service, error) {
	m.mu.RLock()
	extension, exists := m.extensions[extensionName]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("extension %s not found", extensionName)
	}

	service := extension.Instance.GetServices()
	if service == nil {
		return nil, fmt.Errorf("no service found in extension %s", extensionName)
	}

	return service, nil
}

// GetServices returns all registered extension services
func (m *Manager) GetServices() map[string]Service {
	m.mu.RLock()
	defer m.mu.RUnlock()

	services := make(map[string]Service)
	for name, extension := range m.extensions {
		services[name] = extension.Instance.GetServices()
	}
	return services
}

// GetMetadata returns the metadata of all registered extensions
func (m *Manager) GetMetadata() map[string]Metadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metadata := make(map[string]Metadata)
	for name, extension := range m.extensions {
		metadata[name] = extension.Metadata
	}
	return metadata
}

// GetStatus returns the status of all registered extensions
func (m *Manager) GetStatus() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]string)
	for name, extension := range m.extensions {
		status[name] = extension.Instance.Status()
	}
	return status
}

// ExecuteWithCircuitBreaker executes a function with circuit breaker protection
func (m *Manager) ExecuteWithCircuitBreaker(extensionName string, fn func() (any, error)) (any, error) {
	cb, ok := m.circuitBreakers[extensionName]
	if !ok {
		return nil, fmt.Errorf("circuit breaker not found for extension %s", extensionName)
	}

	return cb.Execute(fn)
}

// PublishMessage publishes a message using RabbitMQ or Kafka
func (m *Manager) PublishMessage(exchange, routingKey string, body []byte) error {
	if m.data.Svc.RabbitMQ != nil {
		return m.data.PublishToRabbitMQ(exchange, routingKey, body)
	} else if m.data.Svc.Kafka != nil {
		return m.data.PublishToKafka(context.Background(), routingKey, nil, body)
	}
	return fmt.Errorf("no message queue service available")
}

// SubscribeToMessages subscribes to messages from RabbitMQ or Kafka
func (m *Manager) SubscribeToMessages(queue string, handler func([]byte) error) error {
	if m.data.Svc.RabbitMQ != nil {
		return m.data.ConsumeFromRabbitMQ(queue, handler)
	} else if m.data.Svc.Kafka != nil {
		return m.data.ConsumeFromKafka(context.Background(), queue, "group", handler)
	}
	return fmt.Errorf("no message queue service available")
}

// PublishEvent publishes an event to all extensions
func (m *Manager) PublishEvent(eventName string, data any) {
	m.eventBus.Publish(eventName, data)
}

// SubscribeEvent allows a extension to subscribe to an event
func (m *Manager) SubscribeEvent(eventName string, handler func(any)) {
	m.eventBus.Subscribe(eventName, handler)
}
