package plugin

import (
	"fmt"
	"sync"

	"github.com/ncobase/ncore/extension/config"
)

// Manager manages plugin-specific configurations
type Manager struct {
	mu         sync.RWMutex
	cfg        map[string]any
	maxPlugins int
}

// NewManager creates a new plugin config manager
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		cfg:        cfg.PluginConfig,
		maxPlugins: cfg.MaxPlugins,
	}
}

// GetPluginConfig returns configuration for a specific plugin
func (cm *Manager) GetPluginConfig(pluginName string) (any, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	cfg, exists := cm.cfg[pluginName]
	return cfg, exists
}

// SetPluginConfig sets configuration for a specific plugin
func (cm *Manager) SetPluginConfig(pluginName string, config any) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.cfg == nil {
		cm.cfg = make(map[string]any)
	}
	cm.cfg[pluginName] = config
}

// RemovePluginConfig removes configuration for a specific plugin
func (cm *Manager) RemovePluginConfig(pluginName string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.cfg, pluginName)
}

// GetAllPluginConfigs returns all plugin configurations
func (cm *Manager) GetAllPluginConfigs() map[string]any {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make(map[string]any)
	for name, cfg := range cm.cfg {
		result[name] = cfg
	}
	return result
}

// ValidatePluginLimit checks if plugin limit is reached
func (cm *Manager) ValidatePluginLimit(currentCount int) error {
	if cm.maxPlugins > 0 && currentCount >= cm.maxPlugins {
		return fmt.Errorf("plugin limit reached: %d/%d", currentCount, cm.maxPlugins)
	}
	return nil
}

// GetMaxPlugins returns the maximum number of plugins allowed
func (cm *Manager) GetMaxPlugins() int {
	return cm.maxPlugins
}
