package manager

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// Cross service registry
type crossRegistry struct {
	mu       sync.RWMutex
	services map[string]any
}

var globalCrossRegistry = &crossRegistry{
	services: make(map[string]any),
}

// GetCrossService gets service by extension name and service path
func (m *Manager) GetCrossService(extensionName, servicePath string) (any, error) {
	key := fmt.Sprintf("%s.%s", extensionName, servicePath)

	// Try from registered services first
	globalCrossRegistry.mu.RLock()
	if service, exists := globalCrossRegistry.services[key]; exists {
		globalCrossRegistry.mu.RUnlock()
		return service, nil
	}
	globalCrossRegistry.mu.RUnlock()

	// Fallback to direct extraction if not in registry
	extensionService, err := m.GetService(extensionName)
	if err != nil {
		return nil, fmt.Errorf("extension %s not found: %v", extensionName, err)
	}

	return extractServiceByPath(extensionService, servicePath)
}

// RegisterCrossService manually registers a service
func (m *Manager) RegisterCrossService(key string, service any) {
	globalCrossRegistry.mu.Lock()
	defer globalCrossRegistry.mu.Unlock()
	globalCrossRegistry.services[key] = service
	logger.Debugf(nil, "Cross service registered: %s", key)
}

// RefreshCrossServices clears and re-registers all services
func (m *Manager) RefreshCrossServices() {
	globalCrossRegistry.mu.Lock()
	globalCrossRegistry.services = make(map[string]any)
	globalCrossRegistry.mu.Unlock()

	// Re-register all module services
	m.autoRegisterAllServices()
}

// autoRegisterAllServices auto-registers services from all modules
func (m *Manager) autoRegisterAllServices() {
	extensions := m.GetExtensions()
	for name := range extensions {
		m.autoRegisterModuleServices(name)
	}
	// logger.Debugf(nil, "Auto-registered services for all modules")
}

// autoRegisterModuleServices auto-registers services from a module
func (m *Manager) autoRegisterModuleServices(extensionName string) {
	extensionService, err := m.GetService(extensionName)
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

		globalCrossRegistry.mu.Lock()
		globalCrossRegistry.services[serviceKey] = field.Interface()
		globalCrossRegistry.mu.Unlock()

		// logger.Debugf(nil, "Auto-registered service: %s", serviceKey)
	}
}

// extractServiceByPath extracts service by path
func extractServiceByPath(service any, path string) (any, error) {
	parts := strings.Split(path, ".")
	current := service

	for _, part := range parts {
		if field, ok := types.GetServiceInterface(current, part); ok {
			current = field
		} else {
			return nil, fmt.Errorf("service path %s not found", part)
		}
	}

	return current, nil
}
