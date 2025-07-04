package manager

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/ncobase/ncore/extension/registry"
	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/sony/gobreaker"
)

var (
	eventFallbackMode bool
	eventFallbackMu   sync.RWMutex
)

// InitExtensions initializes all registered extensions
func (m *Manager) InitExtensions() error {
	return m.initExtensionsInternal(context.Background())
}

// initExtensionsInternal performs the actual initialization
func (m *Manager) initExtensionsInternal(ctx context.Context) error {
	m.mu.Lock()
	if m.initialized {
		m.mu.Unlock()
		return fmt.Errorf("extensions already initialized")
	}
	m.mu.Unlock()

	// Check infrastructure readiness
	if err := m.checkInfrastructure(ctx); err != nil {
		return fmt.Errorf("infrastructure check failed: %v", err)
	}

	// Prepare extensions
	registeredExtensions, dependencyGraph := registry.GetExtensionsAndDependencies()
	m.mu.Lock()
	for name, ext := range registeredExtensions {
		if _, exists := m.extensions[name]; !exists {
			m.extensions[name] = &types.Wrapper{
				Metadata: ext.GetMetadata(),
				Instance: ext,
			}
		}
	}

	if err := m.checkDependencies(); err != nil {
		m.mu.Unlock()
		return err
	}

	initOrder, err := getInitOrder(m.extensions, dependencyGraph)
	if err != nil {
		m.mu.Unlock()
		return err
	}
	m.mu.Unlock()

	// Initialize extensions in phases
	if err := m.initializeExtensionsInPhases(ctx, initOrder); err != nil {
		return err
	}

	// Start optional services async
	go m.initOptionalServicesAsync()

	// Publish ready events
	m.publishReadyEvents()

	m.mu.Lock()
	m.initialized = true
	m.mu.Unlock()

	return nil
}

// checkInfrastructure checks basic infrastructure readiness
func (m *Manager) checkInfrastructure(ctx context.Context) error {
	// Test data layer if available
	if m.data != nil {
		if err := m.data.Ping(ctx); err != nil {
			logger.Warnf(nil, "Data layer ping failed: %v", err)
		}
	}

	// Test messaging and set fallback mode
	if m.data != nil && m.data.IsMessagingAvailable() {
		if err := m.testMessaging(); err != nil {
			logger.Warnf(nil, "Messaging test failed: %v, using memory-only events", err)
			m.setEventFallbackMode(true)
		}
	} else {
		m.setEventFallbackMode(true)
	}

	return nil
}

// testMessaging tests messaging connectivity
func (m *Manager) testMessaging() error {
	testData := []byte(`{"test":true}`)
	return m.data.PublishToRabbitMQ("test", "test", testData)
}

// initializeExtensionsInPhases initializes extensions in three phases
func (m *Manager) initializeExtensionsInPhases(ctx context.Context, initOrder []string) error {
	phases := []struct {
		name string
		fn   func(types.Interface) error
	}{
		{"PreInit", func(ext types.Interface) error { return ext.PreInit() }},
		{"Init", func(ext types.Interface) error { return ext.Init(m.conf, m) }},
		{"PostInit", func(ext types.Interface) error { return ext.PostInit() }},
	}

	for _, phase := range phases {
		for _, name := range initOrder {
			ext := m.extensions[name]
			start := time.Now()

			if err := phase.fn(ext.Instance); err != nil {
				logger.Errorf(nil, "Failed %s of extension %s: %v", phase.name, name, err)
				return fmt.Errorf("%s of extension %s failed: %w", phase.name, name, err)
			}

			duration := time.Since(start)
			if phase.name == "Init" {
				m.trackExtensionInitialized(name, duration, nil)
			}

			// Publish extension ready event after PostInit
			if phase.name == "PostInit" {
				m.publishExtensionReadyEvent(name, ext)
			}
		}
	}

	logger.Debugf(nil, "Successfully initialized %d extensions: %v", len(initOrder), initOrder)
	return nil
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

// initOptionalServicesAsync initializes optional services
func (m *Manager) initOptionalServicesAsync() {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf(nil, "Optional services init panic: %v", r)
		}
	}()

	if err := m.initGRPCSupport(); err != nil {
		logger.Warnf(nil, "gRPC initialization failed: %v", err)
	}

	m.registerServicesWithDiscovery()
	m.refreshCrossServices()
}

// publishReadyEvents publishes system ready events
func (m *Manager) publishReadyEvents() {
	eventData := map[string]any{
		"status": "completed",
		"count":  len(m.extensions),
	}

	// Always publish to memory
	m.eventDispatcher.Publish("exts.all.initialized", eventData)

	// Async publish to queue if available
	if !m.isEventFallbackMode() {
		go func() {
			time.Sleep(2 * time.Second) // Give messaging some time
			m.PublishEvent("exts.all.initialized", eventData, types.EventTargetQueue)
		}()
	}
}

// publishExtensionReadyEvent publishes extension ready event
func (m *Manager) publishExtensionReadyEvent(name string, ext *types.Wrapper) {
	eventData := map[string]any{
		"name":     name,
		"status":   "ready",
		"metadata": ext.Instance.GetMetadata(),
	}

	// Always publish to memory
	m.eventDispatcher.Publish(fmt.Sprintf("exts.%s.ready", name), eventData)

	// Async publish to queue if available
	if !m.isEventFallbackMode() {
		go func() {
			m.PublishEvent(fmt.Sprintf("exts.%s.ready", name), eventData, types.EventTargetQueue)
		}()
	}
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
					logger.Warnf(nil, "Failed to register extension %s with service discovery: %v", name, err)
				}
			}
		}
	}
}

// refreshCrossServices clears and re-registers all cross services
func (m *Manager) refreshCrossServices() {
	m.mu.Lock()
	m.crossServices = make(map[string]any)
	m.mu.Unlock()

	for name := range m.extensions {
		m.autoRegisterExtensionServices(name)
	}
}

// autoRegisterExtensionServices auto-registers services from an extension
func (m *Manager) autoRegisterExtensionServices(extensionName string) {
	extensionService, err := m.GetServiceByName(extensionName)
	if err != nil {
		return
	}

	if extensionService == nil {
		return
	}

	m.discoverAndRegisterServices(extensionName, extensionService)
}

// discoverAndRegisterServices uses reflection to find and register service fields
func (m *Manager) discoverAndRegisterServices(extensionName string, service any) {
	serviceValue := reflect.ValueOf(service)

	if serviceValue.Kind() == reflect.Ptr {
		if serviceValue.IsNil() {
			return
		}
		serviceValue = serviceValue.Elem()
	}

	if serviceValue.Kind() != reflect.Struct {
		return
	}

	serviceType := serviceValue.Type()

	for i := 0; i < serviceValue.NumField(); i++ {
		field := serviceValue.Field(i)
		fieldType := serviceType.Field(i)

		if !field.CanInterface() {
			continue
		}

		if field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface {
			if field.IsNil() {
				continue
			}
		}

		fieldName := fieldType.Name
		serviceKey := fmt.Sprintf("%s.%s", extensionName, fieldName)

		m.mu.Lock()
		m.crossServices[serviceKey] = field.Interface()
		m.mu.Unlock()
	}
}

// setEventFallbackMode sets event fallback mode
func (m *Manager) setEventFallbackMode(fallback bool) {
	eventFallbackMu.Lock()
	defer eventFallbackMu.Unlock()
	eventFallbackMode = fallback
}

// isEventFallbackMode returns event fallback mode status
func (m *Manager) isEventFallbackMode() bool {
	eventFallbackMu.RLock()
	defer eventFallbackMu.RUnlock()
	return eventFallbackMode
}

// cleanupPartialInitialization cleans up partial initialization state
func (m *Manager) cleanupPartialInitialization() {
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.Warnf(nil, "Cleaning up partial initialization state")

	for name, ext := range m.extensions {
		if ext.Instance.Status() == types.StatusInitializing {
			if err := ext.Instance.Cleanup(); err != nil {
				logger.Errorf(nil, "Failed to cleanup extension %s: %v", name, err)
			}
		}
	}

	m.initialized = false
	m.circuitBreakers = make(map[string]*gobreaker.CircuitBreaker)
	m.crossServices = make(map[string]any)
}
