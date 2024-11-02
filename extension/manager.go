package extension

import (
	"context"
	"fmt"
	"ncobase/common/config"
	"ncobase/common/data"
	"ncobase/common/ecode"
	"ncobase/common/log"
	"ncobase/common/resp"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
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

// LoadPlugins loads all plugins based on the current configuration
func (m *Manager) LoadPlugins() error {
	if isIncludePluginMode(m.conf) {
		return m.loadPluginsInBuilt()
	}
	return m.loadPluginsInFile()
}

// loadPluginsInFile loads plugins in production mode
func (m *Manager) loadPluginsInFile() error {
	fc := m.conf.Extension
	fd := fc.Path

	pds, err := filepath.Glob(filepath.Join(fd, "*.so"))
	if err != nil {
		log.Errorf(context.Background(), "failed to list plugin files: %v", err)
		return err
	}

	for _, pp := range pds {
		pluginName := strings.TrimSuffix(filepath.Base(pp), ".so")
		if !m.shouldLoadPlugin(pluginName) {
			log.Infof(context.Background(), "🚧 Skipping plugin %s based on configuration", pluginName)
			continue
		}
		if err := m.loadPlugin(pp); err != nil {
			log.Errorf(context.Background(), "Failed to load plugin %s: %v", pluginName, err)
			return err
		}
	}

	return nil
}

// loadPluginsInBuilt built-in all plugins.
func (m *Manager) loadPluginsInBuilt() error {
	plugins := GetRegisteredPlugins()

	for _, c := range plugins {
		if err := m.initializePlugin(c); err != nil {
			log.Errorf(context.Background(), "Failed to initialize plugin %s: %v", c.Metadata.Name, err)
			continue
		}
		m.extensions[c.Metadata.Name] = c
		log.Infof(context.Background(), "Plugin %s loaded and initialized successfully", c.Metadata.Name)
	}

	return nil
}

// initializePlugin initializes a single plugin
func (m *Manager) initializePlugin(c *Wrapper) error {
	if err := c.Instance.PreInit(); err != nil {
		return fmt.Errorf("failed pre-initialization: %v", err)
	}
	if err := c.Instance.Init(m.conf, m); err != nil {
		return fmt.Errorf("failed initialization: %v", err)
	}
	if err := c.Instance.PostInit(); err != nil {
		return fmt.Errorf("failed post-initialization: %v", err)
	}
	return nil
}

// shouldLoadPlugin returns true if the plugin should be loaded
func (m *Manager) shouldLoadPlugin(name string) bool {
	fc := m.conf.Extension

	if len(fc.Includes) > 0 {
		for _, include := range fc.Includes {
			if include == name {
				return true
			}
		}
		return false
	}

	if len(fc.Excludes) > 0 {
		for _, exclude := range fc.Excludes {
			if exclude == name {
				return false
			}
		}
	}

	return true
}

// loadPlugin loads a single plugin
func (m *Manager) loadPlugin(path string) error {
	name := strings.TrimSuffix(filepath.Base(path), ".so")
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.extensions[name]; exists {
		return nil // plugin already loaded
	}

	if err := LoadPlugin(path, m); err != nil {
		log.Errorf(context.Background(), "failed to load plugin %s: %v", name, err)
		return err
	}

	loadedPlugin := GetPlugin(name)
	if loadedPlugin != nil {
		m.extensions[name] = loadedPlugin
		log.Infof(context.Background(), "Plugin %s loaded successfully", name)
	}

	return nil
}

// UnloadPlugin unloads a single extension
func (m *Manager) UnloadPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	extension, exists := m.extensions[name]
	if !exists {
		return fmt.Errorf("extension %s not found", name)
	}

	if err := extension.Instance.PreCleanup(); err != nil {
		log.Errorf(context.Background(), "failed pre-cleanup of extension %s: %v", name, err)
	}

	if err := extension.Instance.Cleanup(); err != nil {
		log.Errorf(context.Background(), "failed to cleanup extension %s: %v", name, err)
		return err
	}

	delete(m.extensions, name)
	delete(m.circuitBreakers, name)

	if err := m.DeregisterConsulService(name); err != nil {
		log.Errorf(context.Background(), "failed to deregister service %s from Consul: %v", name, err)
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

// RegisterConsulService registers a service with Consul
func (m *Manager) RegisterConsulService(name string, address string, port int) error {
	return m.consul.Agent().ServiceRegister(&api.AgentServiceRegistration{
		Name:    name,
		Address: address,
		Port:    port,
	})
}

// DeregisterConsulService deregisters a service from Consul
func (m *Manager) DeregisterConsulService(name string) error {
	if m.conf.Consul == nil || m.consul == nil {
		// Consul is not configured
		// log.Info(context.Background(), "Consul is not configured")
		return nil
	}
	return m.consul.Agent().ServiceDeregister(name)
}

// GetConsulService returns a service from Consul
func (m *Manager) GetConsulService(name string) (*api.AgentService, error) {
	services, err := m.consul.Agent().Services()
	if err != nil {
		return nil, err
	}
	service, ok := services[name]
	if !ok {
		return nil, fmt.Errorf("service %s not found in Consul", name)
	}
	return service, nil
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

// ManageRoutes manages routes for all extensions / plugins
func (m *Manager) ManageRoutes(r *gin.RouterGroup) {
	r.GET("/exts", func(c *gin.Context) {
		extensions := m.GetExtensions()
		result := make(map[string]map[string][]Metadata)

		for _, extension := range extensions {
			group := extension.Metadata.Group
			if group == "" {
				group = extension.Metadata.Name
			}
			if _, ok := result[group]; !ok {
				result[group] = make(map[string][]Metadata)
			}
			result[group][extension.Metadata.Type] = append(result[group][extension.Metadata.Type], extension.Metadata)
		}

		resp.Success(c.Writer, result)
	})

	r.POST("/exts/load", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("name")))
			return
		}
		fc := m.conf.Extension
		fp := filepath.Join(fc.Path, name+".so")
		if err := m.loadPlugin(fp); err != nil {
			resp.Fail(c.Writer, resp.InternalServer(fmt.Sprintf("Failed to load extension %s: %v", name, err)))
			return
		}
		resp.Success(c.Writer, fmt.Sprintf("%s loaded successfully", name))
	})

	r.POST("/exts/unload", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("name")))
			return
		}
		if err := m.UnloadPlugin(name); err != nil {
			resp.Fail(c.Writer, resp.InternalServer(fmt.Sprintf("Failed to unload extension %s: %v", name, err)))
			return
		}
		resp.Success(c.Writer, fmt.Sprintf("%s unloaded successfully", name))
	})

	r.POST("/exts/reload", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("name")))
			return
		}
		if err := m.ReloadPlugin(name); err != nil {
			resp.Fail(c.Writer, resp.InternalServer(fmt.Sprintf("Failed to reload extension %s: %v", name, err)))
			return
		}
		resp.Success(c.Writer, fmt.Sprintf("%s reloaded successfully", name))
	})
}

// ReloadPlugin reloads a single extension / plugin
func (m *Manager) ReloadPlugin(name string) error {
	fc := m.conf.Extension
	fd := fc.Path
	fp := filepath.Join(fd, name+".so")

	if err := m.UnloadPlugin(name); err != nil {
		return err
	}

	return m.loadPlugin(fp)
}

// ReloadPlugins reloads all extensions / plugins
func (m *Manager) ReloadPlugins() error {
	fc := m.conf.Extension
	fd := fc.Path
	pds, err := filepath.Glob(filepath.Join(fd, "*.so"))
	if err != nil {
		log.Errorf(context.Background(), "failed to list plugin files: %v", err)
		return err
	}
	for _, fp := range pds {
		if err := m.ReloadPlugin(strings.TrimSuffix(filepath.Base(fp), ".so")); err != nil {
			return err
		}
	}
	return nil
}

// RegisterRoutes registers all extension routes with the provided router
func (m *Manager) RegisterRoutes(router *gin.Engine) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, f := range m.extensions {
		if f.Instance.GetHandlers() != nil {
			m.registerExtensionRoutes(router, f)
		}
	}
}

// registerExtensionRoutes registers routes for a single extension
func (m *Manager) registerExtensionRoutes(router *gin.Engine, f *Wrapper) {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        f.Metadata.Name,
		MaxRequests: 100,
		Interval:    5 * time.Second,
		Timeout:     3 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
	})
	m.circuitBreakers[f.Metadata.Name] = cb
	group := router.Group("")
	f.Instance.RegisterRoutes(group)
	// log.Infof(context.Background(), "Registered routes for %s with circuit breaker", f.Metadata.Name)
}

// ExecuteWithCircuitBreaker executes a function with circuit breaker protection
func (m *Manager) ExecuteWithCircuitBreaker(extensionName string, fn func() (any, error)) (any, error) {
	cb, ok := m.circuitBreakers[extensionName]
	if !ok {
		return nil, fmt.Errorf("circuit breaker not found for extension %s", extensionName)
	}

	return cb.Execute(fn)
}

// getInitOrder returns the initialization order based on dependencies
//
// noDeps - modules with no dependencies, first to initialize
// withDeps - modules with dependencies
// special - special modules that should be ordered last
func getInitOrder(extensions map[string]*Wrapper) ([]string, error) {
	var noDeps, withDeps, special []string
	specialModules := []string{"relation", "relations", "linker", "linkers"} // exclude these modules from dependency check
	specialSet := make(map[string]bool)
	for _, m := range specialModules {
		specialSet[m] = true
	}

	dependencies := make(map[string]map[string]bool)
	initialized := make(map[string]bool)

	// Collect all available extension names
	availableExtensions := make(map[string]bool)
	for name := range extensions {
		availableExtensions[name] = true
	}

	// analyze dependencies, classify modules into noDeps and withDeps
	for name, extension := range extensions {
		if specialSet[name] {
			special = append(special, name)
			continue
		}

		deps := make(map[string]bool)

		// Check if all dependencies exist
		for _, dep := range extension.Metadata.Dependencies {
			if !specialSet[dep] {
				// Check if dependency exists in extensions
				if !availableExtensions[dep] {
					return nil, fmt.Errorf("extension '%s' depends on '%s' which does not exist", name, dep)
				}
				deps[dep] = true
			}
		}

		if len(deps) == 0 {
			noDeps = append(noDeps, name)
			initialized[name] = true
		} else {
			withDeps = append(withDeps, name)
			dependencies[name] = deps
		}
	}

	// sort noDeps modules
	sort.Strings(noDeps)

	// sort withDeps modules
	var order []string
	order = append(order, noDeps...)

	for len(withDeps) > 0 {
		progress := false
		remainingDeps := withDeps[:0]

		for _, name := range withDeps {
			canInitialize := true
			for dep := range dependencies[name] {
				if !initialized[dep] {
					canInitialize = false
					break
				}
			}

			if canInitialize {
				order = append(order, name)
				initialized[name] = true
				progress = true
			} else {
				remainingDeps = append(remainingDeps, name)
			}
		}

		if !progress {
			return nil, fmt.Errorf("cyclic dependency detected")
		}

		withDeps = remainingDeps
	}

	// add special modules
	order = append(order, special...)

	return order, nil
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

// checkDependencies checks if all dependencies are loaded
func (m *Manager) checkDependencies() error {
	for name, extension := range m.extensions {
		for _, dep := range extension.Instance.Dependencies() {
			if _, ok := m.extensions[dep]; !ok {
				return fmt.Errorf("extension '%s' depends on '%s', which is not available", name, dep)
			}
		}
	}
	return nil
}

// isIncludePluginMode returns true if the mode is "c2hlbgo"
func isIncludePluginMode(conf *config.Config) bool {
	return conf.Extension.Mode == "c2hlbgo"
}
