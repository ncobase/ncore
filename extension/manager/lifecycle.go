package manager

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/ncobase/ncore/extension/registry"
	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// InitExtensions initializes all registered extensions
func (m *Manager) InitExtensions() error {
	ctx := context.Background()

	// Initialize with timeout if available
	initFunc := func(timeoutCtx context.Context) error {
		return m.initExtensionsInternal()
	}

	var err error
	if m.timeoutManager != nil {
		err = m.timeoutManager.WithInitTimeout(ctx, initFunc)
	} else {
		err = initFunc(ctx)
	}

	return err
}

// initExtensionsInternal performs the actual initialization
func (m *Manager) initExtensionsInternal() error {
	m.mu.Lock()
	if m.initialized {
		m.mu.Unlock()
		return fmt.Errorf("extensions already initialized")
	}

	// Get registered extensions and dependency graph
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

	// Check dependencies
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
	return nil
}

// initializeExtensionsInPhases initializes extensions in three phases
func (m *Manager) initializeExtensionsInPhases(initOrder []string) error {
	var initErrors []error
	var successfulExtensions []string

	// Phase 1: Pre-initialization
	for _, name := range initOrder {
		ext := m.extensions[name]
		start := time.Now()
		err := ext.Instance.PreInit()
		duration := time.Since(start)

		if err != nil {
			logger.Errorf(nil, "failed pre-initialization of extension %s: %v", name, err)
			initErrors = append(initErrors, fmt.Errorf("pre-initialization of extension %s failed: %w", name, err))
		}

		// Track metrics if available
		if m.metricsManager != nil {
			m.metricsManager.ExtensionPhase(name, "pre_init", duration, err)
		}
	}

	// Phase 2: Initialization
	for _, name := range initOrder {
		ext := m.extensions[name]
		start := time.Now()
		err := ext.Instance.Init(m.conf, m)
		duration := time.Since(start)

		if err != nil {
			logger.Errorf(nil, "failed to initialize extension %s: %v", name, err)
			initErrors = append(initErrors, fmt.Errorf("initialization of extension %s failed: %w", name, err))
		}

		// Track metrics if available
		if m.metricsManager != nil {
			m.metricsManager.ExtensionInitialized(name, duration, err)
		}
	}

	// Phase 3: Post-initialization
	for _, name := range initOrder {
		ext := m.extensions[name]
		start := time.Now()
		err := ext.Instance.PostInit()
		duration := time.Since(start)

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

		// Track metrics if available
		if m.metricsManager != nil {
			m.metricsManager.ExtensionPhase(name, "post_init", duration, err)
		}
	}

	if len(initErrors) > 0 {
		return fmt.Errorf("failed to initialize extensions: %v", initErrors)
	}

	logger.Debugf(nil, "Successfully initialized %d extensions: %s",
		len(successfulExtensions), successfulExtensions)

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

// refreshCrossServices clears and re-registers all cross services
func (m *Manager) refreshCrossServices() {
	m.mu.Lock()
	m.crossServices = make(map[string]any)
	m.mu.Unlock()

	// Re-register all extension services
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

// getSecurityStatus returns current security status
func (m *Manager) getSecurityStatus() map[string]any {
	status := map[string]any{
		"sandbox_enabled": m.sandbox != nil,
	}

	if m.conf.Extension.Security != nil {
		status["signature_required"] = m.conf.Extension.Security.RequireSignature
		status["trusted_sources"] = len(m.conf.Extension.Security.TrustedSources)
		status["allowed_paths"] = len(m.conf.Extension.Security.AllowedPaths)
		status["blocked_extensions"] = len(m.conf.Extension.Security.BlockedExtensions)
	}

	return status
}
