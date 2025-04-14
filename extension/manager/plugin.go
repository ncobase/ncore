package manager

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/extension/plugin"
	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils"
)

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
		filepath.Join(basePath, "*"+utils.GetPlatformExt()),            // extension/*
		filepath.Join(basePath, "plugins", "*"+utils.GetPlatformExt()), // extension/plugins/*
	}

	for _, pattern := range pluginPaths {
		pds, err := filepath.Glob(pattern)
		if err != nil {
			logger.Errorf(context.Background(), "failed to list plugin files in %s: %v", pattern, err)
			continue
		}

		for _, pp := range pds {
			pluginName := strings.TrimSuffix(filepath.Base(pp), utils.GetPlatformExt())
			if !m.shouldLoadPlugin(pluginName) {
				logger.Infof(context.Background(), "🚧 Skipping plugin %s based on configuration", pluginName)
				continue
			}
			if err := m.LoadPlugin(pp); err != nil {
				logger.Errorf(context.Background(), "Failed to load plugin %s: %v", pluginName, err)
				return err
			}
		}
	}

	return nil
}

// loadPluginsInBuilt built-in all plugins.
func (m *Manager) loadPluginsInBuilt() error {
	plugins := plugin.GetRegisteredPlugins()

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

// LoadPlugin loads a single plugin
func (m *Manager) LoadPlugin(path string) error {
	name := strings.TrimSuffix(filepath.Base(path), utils.GetPlatformExt())
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.extensions[name]; exists {
		return nil // plugin already loaded
	}

	if err := plugin.LoadPlugin(path, m); err != nil {
		logger.Errorf(context.Background(), "failed to load plugin %s: %v", name, err)
		return err
	}

	loadedPlugin := plugin.GetPlugin(name)
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
	fp := filepath.Join(fd, name+utils.GetPlatformExt())

	if err := m.UnloadPlugin(name); err != nil {
		return err
	}

	return m.LoadPlugin(fp)
}

// UnloadPlugin unloads a single extension
func (m *Manager) UnloadPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ext, exists := m.extensions[name]
	if !exists {
		return fmt.Errorf("extension %s not found", name)
	}

	if err := ext.Instance.PreCleanup(); err != nil {
		logger.Errorf(context.Background(), "failed pre-cleanup of extension %s: %v", name, err)
	}

	if err := ext.Instance.Cleanup(); err != nil {
		logger.Errorf(context.Background(), "failed to cleanup extension %s: %v", name, err)
		return err
	}

	delete(m.extensions, name)
	delete(m.circuitBreakers, name)

	if m.serviceDiscovery != nil {
		if err := m.serviceDiscovery.DeregisterService(name); err != nil {
			logger.Errorf(context.Background(), "failed to deregister service %s from Consul: %v", name, err)
		}
	}

	return nil
}

// ReloadPlugins reloads all extensions / plugins
func (m *Manager) ReloadPlugins() error {
	fc := m.conf.Extension
	fd := fc.Path
	pds, err := filepath.Glob(filepath.Join(fd, "*"+utils.GetPlatformExt()))
	if err != nil {
		logger.Errorf(context.Background(), "failed to list plugin files: %v", err)
		return err
	}
	for _, fp := range pds {
		if err := m.ReloadPlugin(strings.TrimSuffix(filepath.Base(fp), utils.GetPlatformExt())); err != nil {
			return err
		}
	}
	return nil
}

// initializePlugin initializes a single plugin
func (m *Manager) initializePlugin(c *types.Wrapper) error {
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

// isIncludePluginMode returns true if the mode is "c2hlbgo"
func isIncludePluginMode(conf *config.Config) bool {
	return conf.Extension.Mode == "c2hlbgo"
}
