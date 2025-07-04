package manager

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/ncobase/ncore/extension/registry"
	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/sony/gobreaker"
)

// InitExtensions initializes all registered extensions
func (m *Manager) InitExtensions() error {
	// Get timeout from config
	timeout := 300 * time.Second // Default 5 minutes
	if m.conf.Extension.InitTimeout != "" {
		if parsed, err := time.ParseDuration(m.conf.Extension.InitTimeout); err == nil {
			timeout = parsed
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("initialization panic: %v", r)
			}
		}()
		done <- m.initExtensionsInternal(ctx)
	}()

	select {
	case err := <-done:
		if err != nil {
			m.cleanupPartialInitialization()
			return fmt.Errorf("extension initialization failed: %v", err)
		}
		return nil
	case <-ctx.Done():
		m.cleanupPartialInitialization()
		return fmt.Errorf("extension initialization timeout after %v", timeout)
	}
}

// initExtensionsInternal performs the actual initialization
func (m *Manager) initExtensionsInternal(ctx context.Context) error {
	m.mu.Lock()
	if m.initialized {
		m.mu.Unlock()
		return fmt.Errorf("extensions already initialized")
	}

	// Get registered extensions and dependency graph
	registeredExtensions, dependencyGraph := registry.GetExtensionsAndDependencies()

	// Add registered extensions to the manager
	for name, ext := range registeredExtensions {
		select {
		case <-ctx.Done():
			m.mu.Unlock()
			return ctx.Err()
		default:
		}

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
		m.mu.Unlock()
		return err
	}
	m.mu.Unlock()

	// Initialize extensions in phases
	if err := m.initializeExtensionsInPhases(ctx, initOrder); err != nil {
		return err
	}

	// Initialize optional services asynchronously
	go m.initOptionalServicesAsync()

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
func (m *Manager) initializeExtensionsInPhases(ctx context.Context, initOrder []string) error {
	var initErrors []error
	var successfulExtensions []string

	// Phase 1: Pre-initialization
	for _, name := range initOrder {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		ext := m.extensions[name]
		if err := m.runWithTimeout(ctx, 30*time.Second, ext.Instance.PreInit); err != nil {
			logger.Errorf(nil, "failed pre-initialization of extension %s: %v", name, err)
			initErrors = append(initErrors, fmt.Errorf("pre-initialization of extension %s failed: %w", name, err))
		}
	}

	// Phase 2: Main initialization
	for _, name := range initOrder {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		ext := m.extensions[name]
		start := time.Now()

		err := m.runWithTimeout(ctx, 120*time.Second, func() error {
			return ext.Instance.Init(m.conf, m)
		})

		duration := time.Since(start)

		if err != nil {
			logger.Errorf(nil, "failed to initialize extension %s: %v", name, err)
			initErrors = append(initErrors, fmt.Errorf("initialization of extension %s failed: %w", name, err))
		} else {
			m.trackExtensionInitialized(name, duration, nil)
		}
	}

	// Phase 3: Post-initialization
	for _, name := range initOrder {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		ext := m.extensions[name]
		if err := m.runWithTimeout(ctx, 30*time.Second, ext.Instance.PostInit); err != nil {
			logger.Errorf(nil, "failed post-initialization of extension %s: %v", name, err)
			initErrors = append(initErrors, fmt.Errorf("post-initialization of extension %s failed: %w", name, err))
		} else {
			successfulExtensions = append(successfulExtensions, name)

			// Publish extension ready event
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

// runWithTimeout runs a function with timeout
func (m *Manager) runWithTimeout(ctx context.Context, timeout time.Duration, fn func() error) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("method panic: %v", r)
			}
		}()
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-timeoutCtx.Done():
		return timeoutCtx.Err()
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

// initOptionalServicesAsync initializes optional services asynchronously
func (m *Manager) initOptionalServicesAsync() {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf(nil, "optional services init panic: %v", r)
		}
	}()

	// gRPC services
	if err := m.initGRPCSupport(); err != nil {
		logger.Warnf(nil, "gRPC initialization failed: %v", err)
	}

	// Service discovery registration
	m.registerServicesWithDiscovery()

	// Cross-module services
	m.refreshCrossServices()
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

// cleanupPartialInitialization cleans up partial initialization state
func (m *Manager) cleanupPartialInitialization() {
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.Warnf(nil, "cleaning up partial initialization state")

	for name, ext := range m.extensions {
		if ext.Instance.Status() == types.StatusInitializing {
			if err := ext.Instance.Cleanup(); err != nil {
				logger.Errorf(nil, "failed to cleanup extension %s: %v", name, err)
			}
		}
	}

	m.initialized = false
	m.circuitBreakers = make(map[string]*gobreaker.CircuitBreaker)
	m.crossServices = make(map[string]any)
}
