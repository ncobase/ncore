package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/extension/discovery"
	"github.com/ncobase/ncore/extension/event"
	"github.com/ncobase/ncore/extension/grpc"
	"github.com/ncobase/ncore/extension/metrics"
	"github.com/ncobase/ncore/extension/plugin"
	"github.com/ncobase/ncore/extension/security"
	"github.com/ncobase/ncore/extension/timeout"
	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/sony/gobreaker"
)

// Manager manages extensions and provides unified service access
type Manager struct {
	// Core components
	extensions  map[string]*types.Wrapper
	conf        *config.Config
	mu          sync.RWMutex
	initialized bool
	ctx         context.Context
	cancel      context.CancelFunc

	// Service components
	eventDispatcher  *event.Dispatcher
	serviceDiscovery *discovery.ServiceDiscovery
	grpcServer       *grpc.Server
	grpcRegistry     *grpc.ServiceRegistry
	circuitBreakers  map[string]*gobreaker.CircuitBreaker
	crossServices    map[string]any
	data             *data.Data

	// Simplified metrics system
	metricsCollector *metrics.Collector

	// Optional components
	sandbox         *security.Sandbox
	resourceMonitor *security.ResourceMonitor
	timeoutManager  *timeout.Manager
	pm              *plugin.Manager
}

// NewManager creates a new extension manager
func NewManager(conf *config.Config) (*Manager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		extensions:      make(map[string]*types.Wrapper),
		conf:            conf,
		eventDispatcher: event.NewEventDispatcher(),
		circuitBreakers: make(map[string]*gobreaker.CircuitBreaker),
		crossServices:   make(map[string]any),
		ctx:             ctx,
		cancel:          cancel,
	}

	// Initialize all subsystems
	if err := m.initSubsystems(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize subsystems: %v", err)
	}

	return m, nil
}

// initSubsystems initializes all manager subsystems
func (m *Manager) initSubsystems() error {
	// 1. Initialize metrics system first
	if err := m.initMetricsSystem(); err != nil {
		return fmt.Errorf("failed to initialize metrics: %v", err)
	}

	// 2. Initialize data layer
	if err := m.initDataLayer(); err != nil {
		return fmt.Errorf("failed to initialize data layer: %v", err)
	}

	// 3. Initialize service discovery
	if err := m.initServiceDiscovery(); err != nil {
		return fmt.Errorf("failed to initialize service discovery: %v", err)
	}

	// 4. Initialize optional components
	if err := m.initOptionalComponents(); err != nil {
		return fmt.Errorf("failed to initialize optional components: %v", err)
	}

	return nil
}

// initMetricsSystem initializes the metrics system
func (m *Manager) initMetricsSystem() error {
	extConf := m.conf.Extension
	enabled := extConf.Performance != nil && extConf.Performance.EnableMetrics

	// Initialize with memory storage first
	m.metricsCollector = metrics.NewCollectorWithMemoryStorage(enabled)

	return nil
}

// initDataLayer initializes the data layer
func (m *Manager) initDataLayer() error {
	d, _, err := data.New(m.conf.Data)
	if err != nil {
		return err
	}

	m.data = d

	// Upgrade metrics to Redis storage if available
	if m.metricsCollector != nil && m.metricsCollector.IsEnabled() {
		if redisClient := m.data.GetRedis(); redisClient != nil {
			retention := 7 * 24 * time.Hour // 7 days default
			if m.conf.Extension.Performance != nil && m.conf.Extension.Performance.MetricsInterval != "" {
				if parsed, err := time.ParseDuration(m.conf.Extension.Performance.MetricsInterval); err == nil {
					retention = parsed * 168 // 7 days worth of intervals
				}
			}

			// Use Redis client directly from data layer
			if err := m.metricsCollector.UpgradeToRedisStorage(redisClient, "ncore:extension", retention); err != nil {
				logger.Warnf(nil, "Failed to upgrade metrics to Redis storage: %v, continuing with memory storage", err)
			}
		}
	}

	return nil
}

// initServiceDiscovery initializes service discovery
func (m *Manager) initServiceDiscovery() error {
	if m.conf.Consul == nil {
		return nil
	}

	consulConfig := &discovery.ConsulConfig{
		Address: m.conf.Consul.Address,
		Scheme:  m.conf.Consul.Scheme,
		Discovery: struct {
			HealthCheck   bool
			CheckInterval string
			Timeout       string
		}{
			HealthCheck:   m.conf.Consul.Discovery.HealthCheck,
			CheckInterval: m.conf.Consul.Discovery.CheckInterval,
			Timeout:       m.conf.Consul.Discovery.Timeout,
		},
	}

	var err error
	m.serviceDiscovery, err = discovery.NewServiceDiscovery(consulConfig)
	return err
}

// initOptionalComponents initializes optional components
func (m *Manager) initOptionalComponents() error {
	extConf := m.conf.Extension

	// Initialize security sandbox
	if extConf.Security != nil && extConf.Security.EnableSandbox {
		m.sandbox = security.NewSandbox(extConf.Security)
	}

	// Initialize resource monitor
	if extConf.Performance != nil && extConf.Performance.EnableMetrics {
		m.resourceMonitor = security.NewResourceMonitor(extConf.Performance)
	}

	// Initialize timeout manager
	if extConf.LoadTimeout != "" || extConf.InitTimeout != "" || extConf.DependencyTimeout != "" {
		var err error
		m.timeoutManager, err = timeout.NewManager(extConf)
		if err != nil {
			return fmt.Errorf("failed to create timeout manager: %v", err)
		}
	}

	// Initialize plugin manager
	m.pm = plugin.NewManager(extConf)

	return nil
}

// Core interface methods

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

// GetData returns the data layer instance
func (m *Manager) GetData() *data.Data {
	return m.data
}

// Cleanup cleans up all loaded extensions
func (m *Manager) Cleanup() {
	// Cancel context
	if m.cancel != nil {
		m.cancel()
	}

	// Cleanup subsystems in reverse order
	m.cleanupSubsystems()

	// Reset state
	m.extensions = make(map[string]*types.Wrapper)
	m.circuitBreakers = make(map[string]*gobreaker.CircuitBreaker)
	m.crossServices = make(map[string]any)
	m.initialized = false

	// Close data connections
	if m.data != nil {
		if errs := m.data.Close(); len(errs) > 0 {
			logger.Errorf(nil, "errors closing data connections: %v", errs)
		}
	}
}

// cleanupSubsystems cleans up all subsystems
func (m *Manager) cleanupSubsystems() {
	// Cleanup optional components
	if m.resourceMonitor != nil && m.pm != nil {
		for pluginName := range m.extensions {
			m.resourceMonitor.Cleanup(pluginName)
			m.pm.RemovePluginConfig(pluginName)
		}
	}

	// Clear service cache
	if m.serviceDiscovery != nil {
		m.serviceDiscovery.ClearCache()
	}

	// Stop gRPC server
	if m.grpcServer != nil {
		_ = m.grpcServer.Stop(5 * time.Second)
	}

	// Close gRPC registry
	if m.grpcRegistry != nil {
		m.grpcRegistry.Close()
	}

	// Cleanup extensions
	m.cleanupExtensions()

	// Stop and cleanup metrics collector
	if m.metricsCollector != nil {
		// Stop background routines gracefully
		m.metricsCollector.Stop()
	}
}

// cleanupExtensions cleans up all loaded extensions
func (m *Manager) cleanupExtensions() {
	for _, ext := range m.extensions {
		if err := ext.Instance.PreCleanup(); err != nil {
			logger.Errorf(nil, "failed pre-cleanup of extension %s: %v", ext.Metadata.Name, err)
		}
		if err := ext.Instance.Cleanup(); err != nil {
			logger.Errorf(nil, "failed to cleanup extension %s: %v", ext.Metadata.Name, err)
		}

		// Track extension unloading
		m.trackExtensionUnloaded(ext.Metadata.Name)

		// Deregister from service discovery
		if m.serviceDiscovery != nil && ext.Instance.NeedServiceDiscovery() {
			if err := m.serviceDiscovery.DeregisterService(ext.Metadata.Name); err != nil {
				logger.Errorf(nil, "failed to deregister service %s: %v", ext.Metadata.Name, err)
			}
		}
	}
}

// Deprecated methods for backward compatibility

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
	m.refreshCrossServices()
}
