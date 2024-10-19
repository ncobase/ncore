package feature

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

// Manager represents a feature / plugin manager
type Manager struct {
	features        map[string]*Wrapper
	conf            *config.Config
	mu              sync.RWMutex
	initialized     bool
	eventBus        *EventBus
	consul          *api.Client
	circuitBreakers map[string]*gobreaker.CircuitBreaker
	data            *data.Data
}

// NewManager creates a new feature / plugin manager
func NewManager(conf *config.Config) (*Manager, error) {
	d, cleanup, err := data.New(conf.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to create data connections: %v", err)
	}

	consulClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("failed to create Consul client: %v", err)
	}

	return &Manager{
		features:        make(map[string]*Wrapper),
		conf:            conf,
		eventBus:        NewEventBus(),
		consul:          consulClient,
		circuitBreakers: make(map[string]*gobreaker.CircuitBreaker),
		data:            d,
	}, nil
}

// Register registers a feature
func (m *Manager) Register(f Interface) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("cannot register feature after initialization")
	}

	name := f.Name()
	if _, exists := m.features[name]; exists {
		return fmt.Errorf("feature %s already registered", name)
	}

	m.features[name] = &Wrapper{
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
	fc := m.conf.Feature
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
		m.features[c.Metadata.Name] = c
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
	fc := m.conf.Feature

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

	if _, exists := m.features[name]; exists {
		return nil // plugin already loaded
	}

	if err := LoadPlugin(path, m); err != nil {
		log.Errorf(context.Background(), "failed to load plugin %s: %v", name, err)
		return err
	}

	loadedPlugin := GetPlugin(name)
	if loadedPlugin != nil {
		m.features[name] = loadedPlugin
		log.Infof(context.Background(), "Plugin %s loaded successfully", name)
	}

	return nil
}

// UnloadPlugin unloads a single feature
func (m *Manager) UnloadPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	feature, exists := m.features[name]
	if !exists {
		return fmt.Errorf("feature %s not found", name)
	}

	if err := feature.Instance.PreCleanup(); err != nil {
		log.Errorf(context.Background(), "failed pre-cleanup of feature %s: %v", name, err)
	}

	if err := feature.Instance.Cleanup(); err != nil {
		log.Errorf(context.Background(), "failed to cleanup feature %s: %v", name, err)
		return err
	}

	delete(m.features, name)
	delete(m.circuitBreakers, name)

	if err := m.DeregisterConsulService(name); err != nil {
		log.Errorf(context.Background(), "failed to deregister service %s from Consul: %v", name, err)
	}

	return nil
}

// InitFeatures initializes all registered features
func (m *Manager) InitFeatures() error {
	m.mu.Lock()
	if m.initialized {
		m.mu.Unlock()
		return fmt.Errorf("features already initialized")
	}
	// Check dependencies before determining initialization order
	if err := m.checkDependencies(); err != nil {
		m.mu.Unlock()
		return err
	}
	initOrder, err := getInitOrder(m.features)
	if err != nil {
		log.Errorf(context.Background(), "failed to determine initialization order: %v", err)
		m.mu.Unlock()
		return err
	}
	m.mu.Unlock() // Unlock after dependencies check and order determination

	// Pre-initialization
	for _, name := range initOrder {
		feature := m.features[name]
		if err := feature.Instance.PreInit(); err != nil {
			log.Errorf(context.Background(), "failed pre-initialization of feature %s: %v", name, err)
			continue // Skip current feature and move to the next one
		}
	}

	// Initialization
	for _, name := range initOrder {
		feature := m.features[name]
		if err := feature.Instance.Init(m.conf, m); err != nil {
			log.Errorf(context.Background(), "failed to initialize feature %s: %v", name, err)
			continue // Skip current feature and move to the next one
		}
	}

	// Post-initialization
	for _, name := range initOrder {
		feature := m.features[name]
		if err := feature.Instance.PostInit(); err != nil {
			log.Errorf(context.Background(), "failed post-initialization of feature %s: %v", name, err)
			continue // Skip current feature and move to the next one
		}
	}

	// Ensure all services are initialized
	for _, feature := range m.features {
		_ = feature.Instance.GetServices()
	}

	// Lock again to safely update the initialized flag
	m.mu.Lock()
	m.initialized = true
	m.mu.Unlock()

	// log.Infof(context.Background(), " All features initialized successfully")
	return nil
}

// GetFeature returns a specific feature
func (m *Manager) GetFeature(name string) (Interface, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	feature, exists := m.features[name]
	if !exists {
		return nil, fmt.Errorf("feature %s not found", name)
	}

	return feature.Instance, nil
}

// GetFeatures returns the loaded features
func (m *Manager) GetFeatures() map[string]*Wrapper {
	m.mu.RLock()
	defer m.mu.RUnlock()

	features := make(map[string]*Wrapper)
	for name, feature := range m.features {
		features[name] = feature
	}
	return features
}

// Cleanup cleans up all loaded features
func (m *Manager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, feature := range m.features {
		if err := feature.Instance.PreCleanup(); err != nil {
			log.Errorf(context.Background(), "failed pre-cleanup of feature %s: %v", feature.Metadata.Name, err)
		}
		if err := feature.Instance.Cleanup(); err != nil {
			log.Errorf(context.Background(), "failed to cleanup feature %s: %v", feature.Metadata.Name, err)
		}
		if err := m.DeregisterConsulService(feature.Metadata.Name); err != nil {
			log.Errorf(context.Background(), "failed to deregister service %s from Consul: %v", feature.Metadata.Name, err)
		}
	}

	m.features = make(map[string]*Wrapper)
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

// GetHandler returns a specific handler from a feature
func (m *Manager) GetHandler(f string) (Handler, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	feature, exists := m.features[f]
	if !exists {
		return nil, fmt.Errorf("feature %s not found", f)
	}

	handler := feature.Instance.GetHandlers()
	if handler == nil {
		return nil, fmt.Errorf("no handler found in feature %s", f)
	}

	return handler, nil
}

// GetHandlers returns all registered feature handlers
func (m *Manager) GetHandlers() map[string]Handler {
	m.mu.RLock()
	defer m.mu.RUnlock()
	handlers := make(map[string]Handler)
	for name, feature := range m.features {
		handlers[name] = feature.Instance.GetHandlers()
	}
	return handlers
}

// GetService returns a specific service from a feature
func (m *Manager) GetService(featureName string) (Service, error) {
	m.mu.RLock()
	feature, exists := m.features[featureName]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("feature %s not found", featureName)
	}

	service := feature.Instance.GetServices()
	if service == nil {
		return nil, fmt.Errorf("no service found in feature %s", featureName)
	}

	return service, nil
}

// GetServices returns all registered feature services
func (m *Manager) GetServices() map[string]Service {
	m.mu.RLock()
	defer m.mu.RUnlock()

	services := make(map[string]Service)
	for name, feature := range m.features {
		services[name] = feature.Instance.GetServices()
	}
	return services
}

// GetMetadata returns the metadata of all registered features
func (m *Manager) GetMetadata() map[string]Metadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metadata := make(map[string]Metadata)
	for name, feature := range m.features {
		metadata[name] = feature.Metadata
	}
	return metadata
}

// GetStatus returns the status of all registered features
func (m *Manager) GetStatus() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]string)
	for name, feature := range m.features {
		status[name] = feature.Instance.Status()
	}
	return status
}

// ManageRoutes manages routes for all features / plugins
func (m *Manager) ManageRoutes(e *gin.Engine) {
	e.GET("/features", func(c *gin.Context) {
		resp.Success(c.Writer, m.GetFeatures())
	})

	e.POST("/features/load", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("name")))
			return
		}
		fc := m.conf.Feature
		fp := filepath.Join(fc.Path, name+".so")
		if err := m.loadPlugin(fp); err != nil {
			resp.Fail(c.Writer, resp.InternalServer(fmt.Sprintf("Failed to load feature %s: %v", name, err)))
			return
		}
		resp.Success(c.Writer, fmt.Sprintf("%s loaded successfully", name))
	})

	e.POST("/features/unload", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("name")))
			return
		}
		if err := m.UnloadPlugin(name); err != nil {
			resp.Fail(c.Writer, resp.InternalServer(fmt.Sprintf("Failed to unload feature %s: %v", name, err)))
			return
		}
		resp.Success(c.Writer, fmt.Sprintf("%s unloaded successfully", name))
	})

	e.POST("/features/reload", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("name")))
			return
		}
		if err := m.ReloadPlugin(name); err != nil {
			resp.Fail(c.Writer, resp.InternalServer(fmt.Sprintf("Failed to reload feature %s: %v", name, err)))
			return
		}
		resp.Success(c.Writer, fmt.Sprintf("%s reloaded successfully", name))
	})
}

// ReloadPlugin reloads a single feature / plugin
func (m *Manager) ReloadPlugin(name string) error {
	fc := m.conf.Feature
	fd := fc.Path
	fp := filepath.Join(fd, name+".so")

	if err := m.UnloadPlugin(name); err != nil {
		return err
	}

	return m.loadPlugin(fp)
}

// ReloadPlugins reloads all features / plugins
func (m *Manager) ReloadPlugins() error {
	fc := m.conf.Feature
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

// RegisterRoutes registers all feature routes with the provided router
func (m *Manager) RegisterRoutes(router *gin.Engine) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, f := range m.features {
		if f.Instance.GetHandlers() != nil {
			m.registerFeatureRoutes(router, f)
		}
	}
}

// registerFeatureRoutes registers routes for a single feature
func (m *Manager) registerFeatureRoutes(router *gin.Engine, f *Wrapper) {
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
	group := router.Group("/" + f.Metadata.Name)
	f.Instance.RegisterRoutes(group)
	log.Infof(context.Background(), "Registered routes for %s with circuit breaker", f.Metadata.Name)
}

// ExecuteWithCircuitBreaker executes a function with circuit breaker protection
func (m *Manager) ExecuteWithCircuitBreaker(featureName string, fn func() (any, error)) (any, error) {
	cb, ok := m.circuitBreakers[featureName]
	if !ok {
		return nil, fmt.Errorf("circuit breaker not found for feature %s", featureName)
	}

	return cb.Execute(fn)
}

// getInitOrder returns the initialization order based on dependencies
//
// noDeps - modules with no dependencies, first to initialize
// withDeps - modules with dependencies
// special - special modules that should be ordered last
func getInitOrder(features map[string]*Wrapper) ([]string, error) {
	var noDeps, withDeps, special []string
	specialModules := []string{"socket", "linker"} // exclude these modules from dependency check
	specialSet := make(map[string]bool)
	for _, m := range specialModules {
		specialSet[m] = true
	}

	dependencies := make(map[string]map[string]bool)
	initialized := make(map[string]bool)

	// analyze dependencies, classify modules into noDeps and withDeps
	for name, feature := range features {
		if specialSet[name] {
			special = append(special, name)
			continue
		}

		deps := make(map[string]bool)
		for _, dep := range feature.Metadata.Dependencies {
			if !specialSet[dep] {
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

// PublishEvent publishes an event to all features
func (m *Manager) PublishEvent(eventName string, data any) {
	m.eventBus.Publish(eventName, data)
}

// SubscribeEvent allows a feature to subscribe to an event
func (m *Manager) SubscribeEvent(eventName string, handler func(any)) {
	m.eventBus.Subscribe(eventName, handler)
}

// checkDependencies checks if all dependencies are loaded
func (m *Manager) checkDependencies() error {
	for name, feature := range m.features {
		for _, dep := range feature.Instance.Dependencies() {
			if _, ok := m.features[dep]; !ok {
				return fmt.Errorf("feature '%s' depends on '%s', which is not available", name, dep)
			}
		}
	}
	return nil
}

// isIncludePluginMode returns true if the mode is "c2hlbgo"
func isIncludePluginMode(conf *config.Config) bool {
	return conf.Feature.Mode == "c2hlbgo"
}
