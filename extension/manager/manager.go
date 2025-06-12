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

	// Metrics system
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

	if err := m.initSubsystems(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize subsystems: %v", err)
	}

	return m, nil
}

// initSubsystems initializes all manager subsystems
func (m *Manager) initSubsystems() error {
	// Initialize metrics system first
	if err := m.initMetricsSystem(); err != nil {
		return fmt.Errorf("failed to initialize metrics: %v", err)
	}

	// Initialize data layer
	if err := m.initDataLayer(); err != nil {
		return fmt.Errorf("failed to initialize data layer: %v", err)
	}

	// Initialize service discovery
	if err := m.initServiceDiscovery(); err != nil {
		return fmt.Errorf("failed to initialize service discovery: %v", err)
	}

	// Initialize optional components
	if err := m.initOptionalComponents(); err != nil {
		return fmt.Errorf("failed to initialize optional components: %v", err)
	}

	return nil
}

// initMetricsSystem initializes the metrics system
func (m *Manager) initMetricsSystem() error {
	// Create collector
	m.metricsCollector = metrics.NewCollector(m.conf.Extension.Metrics)
	return nil
}

// initDataLayer initializes the data layer
func (m *Manager) initDataLayer() error {
	d, _, err := data.New(m.conf.Data)
	if err != nil {
		return err
	}

	m.data = d

	// Upgrade metrics to Redis storage if available and enabled
	if m.metricsCollector != nil && m.metricsCollector.IsEnabled() {
		m.upgradeMetricsStorageIfNeeded()
	}

	return nil
}

// upgradeMetricsStorageIfNeeded upgrades metrics storage to Redis if configured and available
func (m *Manager) upgradeMetricsStorageIfNeeded() {
	if m.metricsCollector == nil || !m.metricsCollector.IsEnabled() {
		return
	}

	metricsConfig := m.conf.Extension.Metrics
	if metricsConfig == nil || metricsConfig.Storage == nil {
		return
	}

	storageType := metricsConfig.Storage.Type

	// Only upgrade if storage type is "redis" or "auto"
	if storageType != "redis" && storageType != "auto" {
		return
	}

	// Check if data layer and Redis are available
	if m.data == nil {
		return
	}

	redisClient := m.data.GetRedis()
	if redisClient == nil {
		if storageType == "redis" {
			logger.Warnf(nil, "Redis storage requested but Redis client not available, using memory storage")
		}
		return
	}

	// Get configuration values
	keyPrefix := metricsConfig.Storage.KeyPrefix
	if keyPrefix == "" {
		keyPrefix = "ncore_ext"
	}

	retention := 7 * 24 * time.Hour
	if metricsConfig.Retention != "" {
		if ret, err := time.ParseDuration(metricsConfig.Retention); err == nil {
			retention = ret
		}
	}

	// Upgrade to Redis storage
	if err := m.metricsCollector.UpgradeToRedisStorage(redisClient, keyPrefix, retention); err != nil {
		logger.Warnf(nil, "Failed to upgrade metrics to Redis storage: %v, continuing with memory storage", err)
	}
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

	// Initialize resource monitor - only if performance config exists
	if extConf.Performance != nil {
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

// Cleanup cleans up all loaded extensions and subsystems
func (m *Manager) Cleanup() {
	// Cancel context first to signal shutdown
	if m.cancel != nil {
		m.cancel()
	}

	// Stop metrics collector
	if m.metricsCollector != nil {
		m.metricsCollector.Stop()
	}

	// Cleanup subsystems in proper order
	m.cleanupSubsystems()

	// Reset state
	m.mu.Lock()
	m.extensions = make(map[string]*types.Wrapper)
	m.circuitBreakers = make(map[string]*gobreaker.CircuitBreaker)
	m.crossServices = make(map[string]any)
	m.initialized = false
	m.mu.Unlock()

	// Close data connections last (Redis client will be closed here)
	if m.data != nil {
		if errs := m.data.Close(); len(errs) > 0 {
			logger.Errorf(nil, "errors closing data connections: %v", errs)
		}
	}
}

// cleanupSubsystems cleans up all subsystems in proper order
func (m *Manager) cleanupSubsystems() {
	// Cleanup extensions first
	m.cleanupExtensions()

	// Stop gRPC server before closing registry
	if m.grpcServer != nil {
		_ = m.grpcServer.Stop(5 * time.Second)
		m.grpcServer = nil
	}

	// Close gRPC registry
	if m.grpcRegistry != nil {
		m.grpcRegistry.Close()
		m.grpcRegistry = nil
	}

	// Clear service discovery cache
	if m.serviceDiscovery != nil {
		m.serviceDiscovery.ClearCache()
	}

	// Cleanup optional components
	if m.resourceMonitor != nil && m.pm != nil {
		for pluginName := range m.extensions {
			m.resourceMonitor.Cleanup(pluginName)
			m.pm.RemovePluginConfig(pluginName)
		}
	}
}

// cleanupExtensions cleans up all loaded extensions
func (m *Manager) cleanupExtensions() {
	m.mu.RLock()
	extensions := make(map[string]*types.Wrapper)
	for name, ext := range m.extensions {
		extensions[name] = ext
	}
	m.mu.RUnlock()

	for _, ext := range extensions {
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
