package extension

import (
	"context"
	"fmt"
	"ncobase/ncore/logger"
	"path/filepath"
	"runtime"
	"strings"
)

// Platform-specific extensions
const (
	ExtWindows = ".dll"
	ExtDarwin  = ".dylib"
	ExtLinux   = ".so"
)

// GetPlatformExt returns the platform-specific extension
func GetPlatformExt() string {
	switch runtime.GOOS {
	case "windows":
		return ExtWindows
	case "darwin":
		return ExtDarwin
	default:
		return ExtLinux
	}
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
	basePath := fc.Path

	// multiple paths
	pluginPaths := []string{
		filepath.Join(basePath, "*"+GetPlatformExt()),            // extension/*
		filepath.Join(basePath, "plugins", "*"+GetPlatformExt()), // extension/plugins/*
	}

	for _, pattern := range pluginPaths {
		pds, err := filepath.Glob(pattern)
		if err != nil {
			logger.Errorf(context.Background(), "failed to list plugin files in %s: %v", pattern, err)
			continue
		}

		for _, pp := range pds {
			pluginName := strings.TrimSuffix(filepath.Base(pp), GetPlatformExt())
			if !m.shouldLoadPlugin(pluginName) {
				logger.Infof(context.Background(), "🚧 Skipping plugin %s based on configuration", pluginName)
				continue
			}
			if err := m.loadPlugin(pp); err != nil {
				logger.Errorf(context.Background(), "Failed to load plugin %s: %v", pluginName, err)
				return err
			}
		}
	}

	return nil
}

// loadPluginsInBuilt built-in all plugins.
func (m *Manager) loadPluginsInBuilt() error {
	plugins := GetRegisteredPlugins()

	for _, c := range plugins {
		if err := m.initializePlugin(c); err != nil {
			logger.Errorf(context.Background(), "Failed to initialize plugin %s: %v", c.Metadata.Name, err)
			continue
		}
		m.extensions[c.Metadata.Name] = c
		logger.Debugf(context.Background(), "Plugin %s loaded and initialized successfully", c.Metadata.Name)
	}

	return nil
}

// loadPlugin loads a single plugin
func (m *Manager) loadPlugin(path string) error {
	name := strings.TrimSuffix(filepath.Base(path), GetPlatformExt())
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.extensions[name]; exists {
		return nil // plugin already loaded
	}

	if err := LoadPlugin(path, m); err != nil {
		logger.Errorf(context.Background(), "failed to load plugin %s: %v", name, err)
		return err
	}

	loadedPlugin := GetPlugin(name)
	if loadedPlugin != nil {
		m.extensions[name] = loadedPlugin
		logger.Infof(context.Background(), "Plugin %s loaded successfully", name)
	}

	return nil
}

// ReloadPlugin reloads a single extension / plugin
func (m *Manager) ReloadPlugin(name string) error {
	fc := m.conf.Extension
	fd := fc.Path
	fp := filepath.Join(fd, name+GetPlatformExt())

	if err := m.UnloadPlugin(name); err != nil {
		return err
	}

	return m.loadPlugin(fp)
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
		logger.Errorf(context.Background(), "failed pre-cleanup of extension %s: %v", name, err)
	}

	if err := extension.Instance.Cleanup(); err != nil {
		logger.Errorf(context.Background(), "failed to cleanup extension %s: %v", name, err)
		return err
	}

	delete(m.extensions, name)
	delete(m.circuitBreakers, name)

	if err := m.DeregisterConsulService(name); err != nil {
		logger.Errorf(context.Background(), "failed to deregister service %s from Consul: %v", name, err)
	}

	return nil
}

// ReloadPlugins reloads all extensions / plugins
func (m *Manager) ReloadPlugins() error {
	fc := m.conf.Extension
	fd := fc.Path
	pds, err := filepath.Glob(filepath.Join(fd, "*"+GetPlatformExt()))
	if err != nil {
		logger.Errorf(context.Background(), "failed to list plugin files: %v", err)
		return err
	}
	for _, fp := range pds {
		if err := m.ReloadPlugin(strings.TrimSuffix(filepath.Base(fp), GetPlatformExt())); err != nil {
			return err
		}
	}
	return nil
}
