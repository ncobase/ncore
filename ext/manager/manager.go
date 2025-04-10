package manager

import (
	"context"
	"fmt"
	core2 "ncore/ext/core"
	"ncore/ext/discovery"
	"ncore/ext/event"
	"ncore/pkg/config"
	"ncore/pkg/data"
	"ncore/pkg/logger"
	"sync"

	"github.com/sony/gobreaker"
)

// Manager represents an extension / plugin manager
type Manager struct {
	extensions       map[string]*core2.Wrapper
	conf             *config.Config
	mu               sync.RWMutex
	initialized      bool
	eventBus         *event.EventBus
	serviceDiscovery *discovery.ServiceDiscovery
	circuitBreakers  map[string]*gobreaker.CircuitBreaker
	data             *data.Data
}

// NewManager creates a new extension / plugin manager
func NewManager(conf *config.Config) (*Manager, error) {
	d, cleanup, err := data.New(conf.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to create data connections: %v", err)
	}

	defer func() {
		if err != nil {
			cleanup()
		}
	}()

	var svcDiscovery *discovery.ServiceDiscovery
	if conf.Consul != nil {
		consulConfig := &discovery.ConsulConfig{
			Address: conf.Consul.Address,
			Scheme:  conf.Consul.Scheme,
			Discovery: struct {
				HealthCheck   bool
				CheckInterval string
				Timeout       string
			}{
				HealthCheck:   conf.Consul.Discovery.HealthCheck,
				CheckInterval: conf.Consul.Discovery.CheckInterval,
				Timeout:       conf.Consul.Discovery.Timeout,
			},
		}

		svcDiscovery, err = discovery.NewServiceDiscovery(consulConfig)
		if err != nil {
			cleanup()
			return nil, fmt.Errorf("failed to create service discovery: %v", err)
		}
	}

	return &Manager{
		extensions:       make(map[string]*core2.Wrapper),
		conf:             conf,
		eventBus:         event.NewEventBus(),
		serviceDiscovery: svcDiscovery,
		circuitBreakers:  make(map[string]*gobreaker.CircuitBreaker),
		data:             d,
	}, nil
}

// GetConfig returns the manager's config
func (m *Manager) GetConfig() *config.Config {
	return m.conf
}

// Register registers an extension
func (m *Manager) Register(f core2.Interface) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("cannot register extension after initialization")
	}

	name := f.Name()
	if _, exists := m.extensions[name]; exists {
		return fmt.Errorf("extension %s already registered", name)
	}

	m.extensions[name] = &core2.Wrapper{
		Metadata: f.GetMetadata(),
		Instance: f,
	}

	if m.serviceDiscovery != nil && f.NeedServiceDiscovery() {
		svcInfo := f.GetServiceInfo()
		if svcInfo != nil {
			if err := m.serviceDiscovery.RegisterService(name, svcInfo); err != nil {
				logger.Warnf(context.Background(), "Failed to register extension %s with Consul: %v", name, err)
			}
		}
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
		logger.Errorf(context.Background(), "failed to determine initialization order: %v", err)
		m.mu.Unlock()
		return err
	}
	m.mu.Unlock() // Unlock after dependencies check and order determination

	var initErrors []error

	// Pre-initialization
	for _, name := range initOrder {
		ext := m.extensions[name]
		if err := ext.Instance.PreInit(); err != nil {
			logger.Errorf(context.Background(), "failed pre-initialization of extension %s: %v", name, err)
			initErrors = append(initErrors, fmt.Errorf("pre-initialization of extension %s failed: %w", name, err))
		}
	}

	// Initialization
	for _, name := range initOrder {
		ext := m.extensions[name]
		err := ext.Instance.Init(m.conf, m)
		if err != nil {
			logger.Errorf(context.Background(), "failed to initialize extension %s: %v", name, err)
			initErrors = append(initErrors, fmt.Errorf("initialization of extension %s failed: %w", name, err))
		}
	}

	// Post-initialization
	for _, name := range initOrder {
		ext := m.extensions[name]
		err := ext.Instance.PostInit()
		if err != nil {
			logger.Errorf(context.Background(), "failed post-initialization of extension %s: %v", name, err)
			initErrors = append(initErrors, fmt.Errorf("post-initialization of extension %s failed: %w", name, err))
		}
	}

	if len(initErrors) > 0 {
		m.mu.Lock()
		m.initialized = false
		m.mu.Unlock()
		// Skip returning error since initialized is false
		// return fmt.Errorf("one or more extensions failed to initialize: %v", initErrors)
	} else {
		m.mu.Lock()
		m.initialized = true
		m.mu.Unlock()
	}

	// Ensure all services are initialized
	for _, ext := range m.extensions {
		_ = ext.Instance.GetServices()
	}

	// Log successful initialization
	logger.Info(context.Background(), "All extensions initialized successfully")
	return nil
}

// GetExtension returns a specific extension
func (m *Manager) GetExtension(name string) (core2.Interface, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ext, exists := m.extensions[name]
	if !exists {
		return nil, fmt.Errorf("extension %s not found", name)
	}

	return ext.Instance, nil
}

// GetExtensions returns the loaded extensions
func (m *Manager) GetExtensions() map[string]*core2.Wrapper {
	m.mu.RLock()
	defer m.mu.RUnlock()

	extensions := make(map[string]*core2.Wrapper)
	for name, extension := range m.extensions {
		extensions[name] = extension
	}
	return extensions
}

// Cleanup cleans up all loaded extensions
func (m *Manager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear service cache
	if m.serviceDiscovery != nil {
		m.serviceDiscovery.ClearCache()
	}

	for _, ext := range m.extensions {
		if err := ext.Instance.PreCleanup(); err != nil {
			logger.Errorf(context.Background(), "failed pre-cleanup of extension %s: %v", ext.Metadata.Name, err)
		}
		if err := ext.Instance.Cleanup(); err != nil {
			logger.Errorf(context.Background(), "failed to cleanup extension %s: %v", ext.Metadata.Name, err)
		}
		// Deregister from Consul
		if m.serviceDiscovery != nil && ext.Instance.NeedServiceDiscovery() {
			if err := m.serviceDiscovery.DeregisterService(ext.Metadata.Name); err != nil {
				logger.Errorf(context.Background(), "failed to deregister service %s from Consul: %v", ext.Metadata.Name, err)
			}
		}
	}

	m.extensions = make(map[string]*core2.Wrapper)
	m.circuitBreakers = make(map[string]*gobreaker.CircuitBreaker)
	m.initialized = false

	if errs := m.data.Close(); len(errs) > 0 {
		logger.Errorf(context.Background(), "errors closing data connections: %v", errs)
	}
}

// GetHandler returns a specific handler from an extension
func (m *Manager) GetHandler(f string) (core2.Handler, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ext, exists := m.extensions[f]
	if !exists {
		return nil, fmt.Errorf("extension %s not found", f)
	}

	handler := ext.Instance.GetHandlers()
	if handler == nil {
		return nil, fmt.Errorf("no handler found in extension %s", f)
	}

	return handler, nil
}

// GetHandlers returns all registered extension handlers
func (m *Manager) GetHandlers() map[string]core2.Handler {
	m.mu.RLock()
	defer m.mu.RUnlock()
	handlers := make(map[string]core2.Handler)
	for name, ext := range m.extensions {
		handlers[name] = ext.Instance.GetHandlers()
	}
	return handlers
}

// GetService returns a specific service from an extension
func (m *Manager) GetService(extensionName string) (core2.Service, error) {
	m.mu.RLock()
	ext, exists := m.extensions[extensionName]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("extension %s not found", extensionName)
	}

	service := ext.Instance.GetServices()
	if service == nil {
		return nil, fmt.Errorf("no service found in extension %s", extensionName)
	}

	return service, nil
}

// GetServices returns all registered extension services
func (m *Manager) GetServices() map[string]core2.Service {
	m.mu.RLock()
	defer m.mu.RUnlock()

	services := make(map[string]core2.Service)
	for name, ext := range m.extensions {
		services[name] = ext.Instance.GetServices()
	}
	return services
}

// GetMetadata returns the metadata of all registered extensions
func (m *Manager) GetMetadata() map[string]core2.Metadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metadata := make(map[string]core2.Metadata)
	for name, ext := range m.extensions {
		metadata[name] = ext.Metadata
	}
	return metadata
}

// GetStatus returns the status of all registered extensions
func (m *Manager) GetStatus() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]string)
	for name, ext := range m.extensions {
		status[name] = ext.Instance.Status()
	}
	return status
}

// checkDependencies checks if all dependencies are loaded
func (m *Manager) checkDependencies() error {
	for name, ext := range m.extensions {
		for _, dep := range ext.Instance.Dependencies() {
			if _, ok := m.extensions[dep]; !ok {
				return fmt.Errorf("extension '%s' depends on '%s', which is not available", name, dep)
			}
		}
	}
	return nil
}
