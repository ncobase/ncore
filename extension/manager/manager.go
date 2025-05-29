package manager

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/extension/discovery"
	"github.com/ncobase/ncore/extension/event"
	"github.com/ncobase/ncore/extension/grpc"
	"github.com/ncobase/ncore/extension/registry"
	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/sony/gobreaker"
)

// Manager manages extensions and provides unified service access
type Manager struct {
	extensions       map[string]*types.Wrapper
	conf             *config.Config
	mu               sync.RWMutex
	initialized      bool
	eventBus         *event.Bus
	serviceDiscovery *discovery.ServiceDiscovery
	grpcServer       *grpc.Server
	grpcRegistry     *grpc.ServiceRegistry
	circuitBreakers  map[string]*gobreaker.CircuitBreaker
	crossServices    map[string]any
	data             *data.Data
}

// NewManager creates a new extension manager
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
		extensions:       make(map[string]*types.Wrapper),
		conf:             conf,
		eventBus:         event.NewEventBus(),
		serviceDiscovery: svcDiscovery,
		circuitBreakers:  make(map[string]*gobreaker.CircuitBreaker),
		crossServices:    make(map[string]any),
		data:             d,
	}, nil
}

// GetConfig returns the manager's config
func (m *Manager) GetConfig() *config.Config {
	return m.conf
}

// RegisterExtension registers an extension
func (m *Manager) RegisterExtension(ext types.Interface) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("cannot register extension after initialization")
	}

	name := ext.Name()
	if _, exists := m.extensions[name]; exists {
		return fmt.Errorf("extension %s already registered", name)
	}

	m.extensions[name] = &types.Wrapper{
		Metadata: ext.GetMetadata(),
		Instance: ext,
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

	// Get registered extensions and their dependency graph
	registeredExtensions, dependencyGraph := registry.GetExtensionsAndDependencies()

	// Add registered extensions to the manager
	for name, ext := range registeredExtensions {
		if _, exists := m.extensions[name]; !exists {
			m.extensions[name] = &types.Wrapper{
				Metadata: ext.GetMetadata(),
				Instance: ext,
			}
		}
	}

	// Check dependencies and get initialization order
	if err := m.checkDependencies(); err != nil {
		m.mu.Unlock()
		return err
	}

	initOrder, err := getInitOrder(m.extensions, dependencyGraph)
	if err != nil {
		logger.Errorf(nil, "failed to determine initialization order: %v", err)
		m.mu.Unlock()
		return err
	}
	m.mu.Unlock()

	// Initialize extensions in phases
	if err := m.initializeExtensionsInPhases(initOrder); err != nil {
		return err
	}

	// Initialize gRPC if enabled
	if err := m.initGRPCSupport(); err != nil {
		return fmt.Errorf("failed to initialize gRPC: %v", err)
	}

	// Register services with service discovery
	m.registerServicesWithDiscovery()

	// Auto-register cross-module services
	m.refreshCrossServices()

	m.mu.Lock()
	m.initialized = true
	m.mu.Unlock()

	// Publish initialization complete event
	m.PublishEvent("exts.all.initialized", map[string]any{
		"status": "completed",
		"count":  len(m.extensions),
	})

	logger.Infof(nil, "Successfully initialized %d extensions", len(m.extensions))
	return nil
}

// GetExtensionByName returns a specific extension by name
func (m *Manager) GetExtensionByName(name string) (types.Interface, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ext, exists := m.extensions[name]
	if !exists {
		return nil, fmt.Errorf("extension %s not found", name)
	}

	return ext.Instance, nil
}

// ListExtensions returns all loaded extensions
func (m *Manager) ListExtensions() map[string]*types.Wrapper {
	m.mu.RLock()
	defer m.mu.RUnlock()

	extensions := make(map[string]*types.Wrapper)
	for name, extension := range m.extensions {
		extensions[name] = extension
	}
	return extensions
}

// GetHandlerByName returns a specific handler from an extension
func (m *Manager) GetHandlerByName(name string) (types.Handler, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ext, exists := m.extensions[name]
	if !exists {
		return nil, fmt.Errorf("extension %s not found", name)
	}

	handler := ext.Instance.GetHandlers()
	if handler == nil {
		return nil, fmt.Errorf("no handler found in extension %s", name)
	}

	return handler, nil
}

// ListHandlers returns all registered extension handlers
func (m *Manager) ListHandlers() map[string]types.Handler {
	m.mu.RLock()
	defer m.mu.RUnlock()

	handlers := make(map[string]types.Handler)
	for name, ext := range m.extensions {
		if handler := ext.Instance.GetHandlers(); handler != nil {
			handlers[name] = handler
		}
	}
	return handlers
}

// GetServiceByName returns a specific service from an extension
func (m *Manager) GetServiceByName(extensionName string) (types.Service, error) {
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

// ListServices returns all registered extension services
func (m *Manager) ListServices() map[string]types.Service {
	m.mu.RLock()
	defer m.mu.RUnlock()

	services := make(map[string]types.Service)
	for name, ext := range m.extensions {
		if service := ext.Instance.GetServices(); service != nil {
			services[name] = service
		}
	}
	return services
}

// GetExtensionPublisher returns a specific extension publisher
func (m *Manager) GetExtensionPublisher(name string, publisherType reflect.Type) (any, error) {
	ext, err := m.GetExtensionByName(name)
	if err != nil {
		return nil, err
	}

	publisher := ext.GetPublisher()
	if publisher == nil {
		return nil, fmt.Errorf("extension %s does not provide a publisher", name)
	}

	pubValue := reflect.ValueOf(publisher)
	if !pubValue.Type().ConvertibleTo(publisherType) {
		return nil, fmt.Errorf("extension %s publisher type %s is not compatible with requested type %s",
			name, pubValue.Type().String(), publisherType.String())
	}

	return publisher, nil
}

// GetExtensionSubscriber returns a specific extension subscriber
func (m *Manager) GetExtensionSubscriber(name string, subscriberType reflect.Type) (any, error) {
	ext, err := m.GetExtensionByName(name)
	if err != nil {
		return nil, err
	}

	subscriber := ext.GetSubscriber()
	if subscriber == nil {
		return nil, fmt.Errorf("extension %s does not provide a subscriber", name)
	}

	subValue := reflect.ValueOf(subscriber)
	if !subValue.Type().ConvertibleTo(subscriberType) {
		return nil, fmt.Errorf("extension %s subscriber type %s is not compatible with requested type %s",
			name, subValue.Type().String(), subscriberType.String())
	}

	return subscriber, nil
}

// GetMetadata returns the metadata of all registered extensions
func (m *Manager) GetMetadata() map[string]types.Metadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metadata := make(map[string]types.Metadata)
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

// Cleanup cleans up all loaded extensions
func (m *Manager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear service cache
	if m.serviceDiscovery != nil {
		m.serviceDiscovery.ClearCache()
	}

	// Stop gRPC server
	if m.grpcServer != nil {
		m.grpcServer.Stop(5 * time.Second)
	}

	// Close gRPC registry
	if m.grpcRegistry != nil {
		m.grpcRegistry.Close()
	}

	// Cleanup extensions
	for _, ext := range m.extensions {
		if err := ext.Instance.PreCleanup(); err != nil {
			logger.Errorf(nil, "failed pre-cleanup of extension %s: %v", ext.Metadata.Name, err)
		}
		if err := ext.Instance.Cleanup(); err != nil {
			logger.Errorf(nil, "failed to cleanup extension %s: %v", ext.Metadata.Name, err)
		}

		// Deregister from service discovery
		if m.serviceDiscovery != nil && ext.Instance.NeedServiceDiscovery() {
			if err := m.serviceDiscovery.DeregisterService(ext.Metadata.Name); err != nil {
				logger.Errorf(nil, "failed to deregister service %s: %v", ext.Metadata.Name, err)
			}
		}
	}

	// Reset state
	m.extensions = make(map[string]*types.Wrapper)
	m.circuitBreakers = make(map[string]*gobreaker.CircuitBreaker)
	m.crossServices = make(map[string]any)
	m.initialized = false

	// Close data connections
	if errs := m.data.Close(); len(errs) > 0 {
		logger.Errorf(nil, "errors closing data connections: %v", errs)
	}
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

// initializeExtensionsInPhases initializes extensions in three phases
func (m *Manager) initializeExtensionsInPhases(initOrder []string) error {
	var initErrors []error
	var successfulExtensions []string

	// Phase 1: Pre-initialization
	for _, name := range initOrder {
		ext := m.extensions[name]
		if err := ext.Instance.PreInit(); err != nil {
			logger.Errorf(nil, "failed pre-initialization of extension %s: %v", name, err)
			initErrors = append(initErrors, fmt.Errorf("pre-initialization of extension %s failed: %w", name, err))
		}
	}

	// Phase 2: Initialization
	for _, name := range initOrder {
		ext := m.extensions[name]
		err := ext.Instance.Init(m.conf, m)
		if err != nil {
			logger.Errorf(nil, "failed to initialize extension %s: %v", name, err)
			initErrors = append(initErrors, fmt.Errorf("initialization of extension %s failed: %w", name, err))
		}
	}

	// Phase 3: Post-initialization
	for _, name := range initOrder {
		ext := m.extensions[name]
		err := ext.Instance.PostInit()
		if err != nil {
			logger.Errorf(nil, "failed post-initialization of extension %s: %v", name, err)
			initErrors = append(initErrors, fmt.Errorf("post-initialization of extension %s failed: %w", name, err))
		} else {
			successfulExtensions = append(successfulExtensions, name)
			m.PublishEvent(fmt.Sprintf("exts.%s.ready", name), map[string]any{
				"name":     name,
				"status":   "ready",
				"metadata": ext.Instance.GetMetadata(),
			})
		}
	}

	if len(initErrors) > 0 {
		return fmt.Errorf("failed to initialize extensions: %v", initErrors)
	}

	logger.Debugf(nil, "Successfully initialized %d extensions: %s",
		len(successfulExtensions), successfulExtensions)

	return nil
}

// registerServicesWithDiscovery registers services with service discovery
func (m *Manager) registerServicesWithDiscovery() {
	if m.serviceDiscovery == nil {
		return
	}

	for name, ext := range m.extensions {
		if ext.Instance.NeedServiceDiscovery() {
			svcInfo := ext.Instance.GetServiceInfo()
			if svcInfo != nil {
				if err := m.serviceDiscovery.RegisterService(name, svcInfo); err != nil {
					logger.Warnf(nil, "failed to register extension %s with service discovery: %v", name, err)
				}
			}
		}
	}
}

// Deprecated methods for backward compatibility - will be removed in future versions

// Register is deprecated, use RegisterExtension instead
func (m *Manager) Register(ext types.Interface) error {
	return m.RegisterExtension(ext)
}

// GetExtension is deprecated, use GetExtensionByName instead
func (m *Manager) GetExtension(name string) (types.Interface, error) {
	return m.GetExtensionByName(name)
}

// GetExtensions is deprecated, use ListExtensions instead
func (m *Manager) GetExtensions() map[string]*types.Wrapper {
	return m.ListExtensions()
}

// GetHandler is deprecated, use GetHandlerByName instead
func (m *Manager) GetHandler(name string) (types.Handler, error) {
	return m.GetHandlerByName(name)
}

// GetHandlers is deprecated, use ListHandlers instead
func (m *Manager) GetHandlers() map[string]types.Handler {
	return m.ListHandlers()
}

// GetService is deprecated, use GetServiceByName instead
func (m *Manager) GetService(extensionName string) (types.Service, error) {
	return m.GetServiceByName(extensionName)
}

// GetServices is deprecated, use ListServices instead
func (m *Manager) GetServices() map[string]types.Service {
	return m.ListServices()
}

// RefreshCrossServices is deprecated, automatic refresh
func (m *Manager) RefreshCrossServices() {
	m.RefreshCrossServices()
}

// GetExtensionMetrics returns detailed extension metrics
func (m *Manager) GetExtensionMetrics() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := map[string]any{
		"total":            len(m.extensions),
		"initialized":      m.initialized,
		"by_status":        make(map[string]int),
		"by_group":         make(map[string]int),
		"by_type":          make(map[string]int),
		"cross_services":   len(m.crossServices),
		"circuit_breakers": len(m.circuitBreakers),
	}

	// Count by status, group, and type
	statusCount := make(map[string]int)
	groupCount := make(map[string]int)
	typeCount := make(map[string]int)

	for _, ext := range m.extensions {
		// Count by status
		status := ext.Instance.Status()
		statusCount[status]++

		// Count by group
		group := ext.Metadata.Group
		if group == "" {
			group = "default"
		}
		groupCount[group]++

		// Count by type
		extType := ext.Metadata.Type
		if extType == "" {
			extType = "unknown"
		}
		typeCount[extType]++
	}

	metrics["by_status"] = statusCount
	metrics["by_group"] = groupCount
	metrics["by_type"] = typeCount

	return metrics
}

// GetSystemMetrics returns comprehensive system metrics
func (m *Manager) GetSystemMetrics() map[string]any {
	return map[string]any{
		"events":        m.GetEventsMetrics(),
		"service_cache": m.GetServiceCacheStats(),
		"extensions":    m.GetExtensionMetrics(),
		"timestamp":     time.Now(),
	}
}
