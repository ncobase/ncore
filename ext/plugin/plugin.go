package plugin

import (
	"context"
	"fmt"
	"github.com/ncobase/ncore/ext/core"
	"github.com/ncobase/ncore/pkg/logger"
	plg "plugin"
	"sync"
)

// PluginRegistry manages the loaded plugins
type PluginRegistry struct {
	mu      sync.RWMutex
	plugins map[string]*core.Wrapper
}

var registry = &PluginRegistry{
	plugins: make(map[string]*core.Wrapper),
}

var plugins []*core.Wrapper

// RegisterPlugin registers a new plugin
func RegisterPlugin(c core.Interface, metadata core.Metadata) {
	plugins = append(plugins, &core.Wrapper{
		Metadata: metadata,
		Instance: c,
	})
}

// GetRegisteredPlugins returns the registered plugins
func GetRegisteredPlugins() []*core.Wrapper {
	return plugins
}

// LoadPlugin loads a single plugin
func LoadPlugin(path string, m core.ManagerInterface) error {
	p, err := plg.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin %s: %v", path, err)
	}

	symPlugin, err := p.Lookup("Instance")
	if err != nil {
		return fmt.Errorf("plugin %s does not export 'Instance' symbol: %v", path, err)
	}

	sc, ok := symPlugin.(core.Interface)
	if !ok {
		return fmt.Errorf("plugin %s does not implement interface, got %T", path, sc)
	}

	if err := sc.PreInit(); err != nil {
		return fmt.Errorf("failed pre-initialization of plugin %s: %v", path, err)
	}

	if err := sc.Init(m.GetConfig(), m); err != nil {
		return fmt.Errorf("failed to initialize plugin %s: %v", path, err)
	}

	if err := sc.PostInit(); err != nil {
		return fmt.Errorf("failed post-initialization of plugin %s: %v", path, err)
	}

	metadata := sc.GetMetadata()

	registry.mu.Lock()
	defer registry.mu.Unlock()

	name := sc.Name()
	if _, exists := registry.plugins[name]; exists {
		logger.Warnf(context.Background(), "Plugin %s is being overwritten", name)
	}
	registry.plugins[name] = &core.Wrapper{
		Metadata: metadata,
		Instance: sc,
	}
	logger.Debugf(context.Background(), "Plugin %s loaded and initialized successfully", name)

	return nil
}

// UnloadPlugin unloads a single plugin
func UnloadPlugin(name string) error {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	c, exists := registry.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if err := c.Instance.PreCleanup(); err != nil {
		logger.Warnf(context.Background(), "Failed pre-cleanup of plugin %s: %v", name, err)
	}

	if err := c.Instance.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup plugin %s: %v", name, err)
	}

	delete(registry.plugins, name)
	logger.Infof(context.Background(), "plugin %s unloaded successfully", name)
	return nil
}

// GetPlugin returns a single plugin
func GetPlugin(name string) *core.Wrapper {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	return registry.plugins[name]
}

// GetPlugins returns all loaded plugins
func GetPlugins() map[string]*core.Wrapper {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	plugins := make(map[string]*core.Wrapper)
	for name, c := range registry.plugins {
		plugins[name] = c
	}
	return plugins
}
