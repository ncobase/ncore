package manager

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/ncobase/ncore/extension/plugin"
	"github.com/ncobase/ncore/extension/security"
	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils"
)

// LoadPlugins loads all plugins based on configuration
func (m *Manager) LoadPlugins() error {
	if m.isBuiltInMode() {
		return m.loadBuiltInPlugins()
	}
	return m.loadFilePlugins()
}

// loadFilePlugins loads plugins from files
func (m *Manager) loadFilePlugins() error {
	basePath := m.conf.Extension.Path
	if basePath == "" {
		logger.Warnf(nil, "No plugin path configured, skipping file plugin loading")
		return nil
	}

	// Search for plugin files
	searchPaths := []string{
		filepath.Join(basePath, "*"+utils.GetPlatformExt()),
		filepath.Join(basePath, "plugins", "*"+utils.GetPlatformExt()),
	}

	var loaded []string
	for _, pattern := range searchPaths {
		files, err := filepath.Glob(pattern)
		if err != nil {
			logger.Errorf(nil, "failed to search plugin files in %s: %v", pattern, err)
			continue
		}

		for _, filePath := range files {
			pluginName := strings.TrimSuffix(filepath.Base(filePath), utils.GetPlatformExt())

			if !m.shouldLoadPlugin(pluginName) {
				logger.Infof(nil, "Skipping plugin %s based on configuration", pluginName)
				continue
			}

			if err := m.LoadPlugin(filePath); err != nil {
				logger.Errorf(nil, "Failed to load plugin %s: %v", pluginName, err)
				return err
			}

			loaded = append(loaded, pluginName)
		}
	}

	if len(loaded) > 0 {
		logger.Debugf(nil, "Loaded %d file plugins: %v", len(loaded), loaded)
	}

	return nil
}

// loadBuiltInPlugins loads built-in registered plugins
func (m *Manager) loadBuiltInPlugins() error {
	plugins := plugin.GetRegisteredPlugins()
	var loaded []string

	for _, pluginWrapper := range plugins {
		pluginName := pluginWrapper.Metadata.Name

		if !m.shouldLoadPlugin(pluginName) {
			logger.Infof(nil, "Skipping built-in plugin %s based on configuration", pluginName)
			continue
		}

		if err := m.initializePlugin(pluginWrapper); err != nil {
			logger.Errorf(nil, "Failed to initialize built-in plugin %s: %v", pluginName, err)
			continue
		}

		m.extensions[pluginName] = pluginWrapper
		loaded = append(loaded, pluginName)
	}

	if len(loaded) > 0 {
		logger.Debugf(nil, "Loaded %d built-in plugins: %v", len(loaded), loaded)
	}

	return nil
}

// LoadPlugin loads a single plugin from file with security checks and metrics
func (m *Manager) LoadPlugin(path string) error {
	ctx := context.Background()
	start := time.Now()
	pluginName := extractPluginName(path)

	// Security validation if sandbox enabled
	if m.sandbox != nil {
		if err := m.sandbox.ValidatePluginPath(path); err != nil {
			return fmt.Errorf("security validation failed: %v", err)
		}
		if err := m.sandbox.ValidatePluginSignature(path); err != nil {
			return fmt.Errorf("signature validation failed: %v", err)
		}
	}

	// Resource limit check if monitoring enabled
	if m.resourceMonitor != nil {
		if err := m.resourceMonitor.CheckResourceLimits(pluginName); err != nil {
			return fmt.Errorf("resource limit check failed: %v", err)
		}
	}

	// Plugin limit check
	if m.pm != nil {
		currentCount := len(m.extensions)
		if err := m.pm.ValidatePluginLimit(currentCount); err != nil {
			return err
		}
	}

	// Load plugin with timeout if available
	loadFunc := func(timeoutCtx context.Context) error {
		return m.loadPluginInternal(path)
	}

	var err error
	if m.timeoutManager != nil {
		err = m.timeoutManager.WithLoadTimeout(ctx, loadFunc)
	} else {
		err = loadFunc(ctx)
	}

	duration := time.Since(start)

	if err != nil {
		return fmt.Errorf("plugin loading failed: %v", err)
	}

	// Track successful load metrics
	m.trackExtensionLoaded(pluginName, duration)

	// Record metrics if monitoring enabled
	if m.resourceMonitor != nil {
		metrics := &security.PluginMetrics{
			LoadTime:   duration,
			LastAccess: time.Now(),
		}
		m.resourceMonitor.RecordPluginMetrics(pluginName, metrics)
	}

	logger.Infof(ctx, "plugin loaded: %s (took %v)", pluginName, duration)
	return nil
}

// loadPluginInternal performs the actual plugin loading
func (m *Manager) loadPluginInternal(path string) error {
	name := extractPluginName(path)

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.extensions[name]; exists {
		logger.Debugf(nil, "Plugin %s already loaded, skipping", name)
		return nil
	}

	if err := plugin.LoadPlugin(path, m); err != nil {
		return fmt.Errorf("failed to load plugin %s: %v", name, err)
	}

	loadedPlugin := plugin.GetPlugin(name)
	if loadedPlugin != nil {
		m.extensions[name] = loadedPlugin
		logger.Infof(nil, "Plugin %s loaded successfully", name)
	}

	return nil
}

// ReloadPlugin reloads a single plugin
func (m *Manager) ReloadPlugin(name string) error {
	basePath := m.conf.Extension.Path
	filePath := filepath.Join(basePath, name+utils.GetPlatformExt())

	if err := m.UnloadPlugin(name); err != nil {
		return fmt.Errorf("failed to unload plugin %s: %v", name, err)
	}

	if err := m.LoadPlugin(filePath); err != nil {
		return fmt.Errorf("failed to reload plugin %s: %v", name, err)
	}

	logger.Infof(nil, "Plugin %s reloaded successfully", name)
	return nil
}

// UnloadPlugin unloads a single plugin
func (m *Manager) UnloadPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ext, exists := m.extensions[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Cleanup extension
	if err := ext.Instance.PreCleanup(); err != nil {
		logger.Errorf(nil, "failed pre-cleanup of plugin %s: %v", name, err)
	}

	if err := ext.Instance.Cleanup(); err != nil {
		logger.Errorf(nil, "failed cleanup of plugin %s: %v", name, err)
		return err
	}

	// Remove from collections
	delete(m.extensions, name)
	delete(m.circuitBreakers, name)

	// Remove cross services for this extension
	m.removeCrossServicesForExtension(name)

	// Deregister from service discovery
	if m.serviceDiscovery != nil && ext.Instance.NeedServiceDiscovery() {
		if err := m.serviceDiscovery.DeregisterService(name); err != nil {
			logger.Errorf(nil, "failed to deregister service %s: %v", name, err)
		}
	}

	// Cleanup optional components
	if m.resourceMonitor != nil {
		m.resourceMonitor.Cleanup(name)
	}
	if m.pm != nil {
		m.pm.RemovePluginConfig(name)
	}

	// Track unload
	m.trackExtensionUnloaded(name)

	logger.Infof(nil, "Plugin %s unloaded successfully", name)
	return nil
}

// initializePlugin initializes a single plugin
func (m *Manager) initializePlugin(pluginWrapper *types.Wrapper) error {
	instance := pluginWrapper.Instance

	if err := instance.PreInit(); err != nil {
		return fmt.Errorf("pre-initialization failed: %v", err)
	}

	if err := instance.Init(m.conf, m); err != nil {
		return fmt.Errorf("initialization failed: %v", err)
	}

	if err := instance.PostInit(); err != nil {
		return fmt.Errorf("post-initialization failed: %v", err)
	}

	return nil
}

// shouldLoadPlugin checks if a plugin should be loaded based on configuration
func (m *Manager) shouldLoadPlugin(name string) bool {
	fc := m.conf.Extension

	// If includes list is specified, only load plugins in the list
	if len(fc.Includes) > 0 {
		for _, include := range fc.Includes {
			if include == name {
				return true
			}
		}
		return false
	}

	// If excludes list is specified, skip plugins in the list
	if len(fc.Excludes) > 0 {
		for _, exclude := range fc.Excludes {
			if exclude == name {
				return false
			}
		}
	}

	return true
}

// isBuiltInMode checks if we're in built-in plugin mode
func (m *Manager) isBuiltInMode() bool {
	return m.conf.Extension.Mode == "c2hlbgo" // built-in mode
}

// removeCrossServicesForExtension removes all cross services for an extension
func (m *Manager) removeCrossServicesForExtension(extensionName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	keysToRemove := make([]string, 0)
	prefix := extensionName + "."

	for key := range m.crossServices {
		if strings.HasPrefix(key, prefix) {
			keysToRemove = append(keysToRemove, key)
		}
	}

	for _, key := range keysToRemove {
		delete(m.crossServices, key)
	}

	if len(keysToRemove) > 0 {
		logger.Debugf(nil, "Removed %d cross services for extension %s", len(keysToRemove), extensionName)
	}
}

// extractPluginName extracts plugin name from path
func extractPluginName(path string) string {
	return filepath.Base(strings.TrimSuffix(path, filepath.Ext(path)))
}
