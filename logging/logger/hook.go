package logger

import (
	"fmt"
	"sync"

	"github.com/ncobase/ncore/logging/logger/config"
	"github.com/sirupsen/logrus"
)

// HookType represents the type of logging hook
type HookType string

const (
	HookElasticsearch HookType = "elasticsearch"
	HookOpenSearch    HookType = "opensearch"
	HookMeilisearch   HookType = "meilisearch"
)

// HookFactory creates a logrus hook from configuration
type HookFactory func(cfg *config.Config) (logrus.Hook, error)

var (
	hookFactories = make(map[HookType]HookFactory)
	hookMu        sync.RWMutex
)

// RegisterHookFactory registers a hook factory for a given type.
// This is called by hook packages in their init() functions.
func RegisterHookFactory(hookType HookType, factory HookFactory) {
	hookMu.Lock()
	defer hookMu.Unlock()
	hookFactories[hookType] = factory
}

// GetHookFactory returns the factory for a given hook type
func GetHookFactory(hookType HookType) (HookFactory, bool) {
	hookMu.RLock()
	defer hookMu.RUnlock()
	factory, ok := hookFactories[hookType]
	return factory, ok
}

// GetRegisteredHooks returns list of registered hook types
func GetRegisteredHooks() []HookType {
	hookMu.RLock()
	defer hookMu.RUnlock()
	hooks := make([]HookType, 0, len(hookFactories))
	for hookType := range hookFactories {
		hooks = append(hooks, hookType)
	}
	return hooks
}

// initSearchHooks initializes all registered search engine hooks based on config
func (l *Logger) initSearchHooks(cfg *config.Config) error {
	if cfg == nil {
		return nil
	}

	hookMu.RLock()
	defer hookMu.RUnlock()

	// Try Elasticsearch
	if cfg.Elasticsearch != nil && len(cfg.Elasticsearch.Addresses) > 0 {
		if factory, ok := hookFactories[HookElasticsearch]; ok {
			hook, err := factory(cfg)
			if err != nil {
				return fmt.Errorf("failed to create elasticsearch hook: %w", err)
			}
			l.AddHook(hook)
		}
	}

	// Try OpenSearch
	if cfg.OpenSearch != nil && len(cfg.OpenSearch.Addresses) > 0 {
		if factory, ok := hookFactories[HookOpenSearch]; ok {
			hook, err := factory(cfg)
			if err != nil {
				return fmt.Errorf("failed to create opensearch hook: %w", err)
			}
			l.AddHook(hook)
		}
	}

	// Try Meilisearch
	if cfg.Meilisearch != nil && cfg.Meilisearch.Host != "" {
		if factory, ok := hookFactories[HookMeilisearch]; ok {
			hook, err := factory(cfg)
			if err != nil {
				return fmt.Errorf("failed to create meilisearch hook: %w", err)
			}
			l.AddHook(hook)
		}
	}

	return nil
}
