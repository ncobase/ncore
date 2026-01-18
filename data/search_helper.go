package data

import (
	"github.com/ncobase/ncore/data/config"
	"github.com/ncobase/ncore/data/metrics"
	"github.com/ncobase/ncore/data/search"
)

// SearchCollectorAdapter adapts data/metrics.Collector to data/search.Collector
type SearchCollectorAdapter struct {
	collector metrics.Collector
}

// SearchQuery records search query metrics
func (a *SearchCollectorAdapter) SearchQuery(engine string, err error) {
	a.collector.SearchQuery(engine, err)
}

// SearchIndex records search index operation metrics
func (a *SearchCollectorAdapter) SearchIndex(engine, operation string) {
	a.collector.SearchIndex(engine, operation)
}

// NewSearchClient creates a search client from ncore data layer.
// It automatically detects and creates adapters for available search engines.
//
// Returns nil if no search engines are available.
// Applications should check if the returned client is nil to support optional search functionality.
func NewSearchClient(d *Data, collector ...metrics.Collector) *search.Client {
	var adapters []search.Adapter
	var c search.Collector

	if len(collector) > 0 && collector[0] != nil {
		c = &SearchCollectorAdapter{collector: collector[0]}
	} else {
		c = &search.NoOpCollector{}
	}

	// Try to initialize Elasticsearch
	if factory, err := search.GetAdapterFactory(search.Elasticsearch); err == nil {
		if es := d.GetElasticsearch(); es != nil {
			if adapter, err := factory(es); err == nil {
				adapters = append(adapters, adapter)
			}
		}
	}

	// Try to initialize OpenSearch
	if factory, err := search.GetAdapterFactory(search.OpenSearch); err == nil {
		if os := d.GetOpenSearch(); os != nil {
			if adapter, err := factory(os); err == nil {
				adapters = append(adapters, adapter)
			}
		}
	}

	// Try to initialize Meilisearch
	if factory, err := search.GetAdapterFactory(search.Meilisearch); err == nil {
		if ms := d.GetMeilisearch(); ms != nil {
			if adapter, err := factory(ms); err == nil {
				adapters = append(adapters, adapter)
			}
		}
	}

	// Return nil if no adapters are available
	if len(adapters) == 0 {
		return nil
	}

	client := search.NewClient(c, adapters...)

	// Apply search configuration if it exists
	if d.conf != nil && d.conf.Search != nil {
		client.UpdateSearchConfig(adaptSearchConfig(d.conf.Search))
	}

	return client
}

// adaptSearchConfig converts config layer search config to search module config
func adaptSearchConfig(cfg *config.Search) *search.Config {
	if cfg == nil {
		return nil
	}
	return &search.Config{
		IndexPrefix:     cfg.IndexPrefix,
		DefaultEngine:   cfg.DefaultEngine,
		AutoCreateIndex: cfg.AutoCreateIndex,
		IndexSettings:   adaptIndexSettings(cfg.IndexSettings),
	}
}

// adaptIndexSettings converts config layer index settings to search module index settings
func adaptIndexSettings(s *config.IndexSettings) *search.IndexSettings {
	if s == nil {
		return nil
	}
	return &search.IndexSettings{
		Shards:           s.Shards,
		Replicas:         s.Replicas,
		RefreshInterval:  s.RefreshInterval,
		SearchableFields: s.SearchableFields,
		FilterableFields: s.FilterableFields,
	}
}
