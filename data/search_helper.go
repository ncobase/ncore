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

func (a *SearchCollectorAdapter) SearchQuery(engine string, err error) {
	a.collector.SearchQuery(engine, err)
}

func (a *SearchCollectorAdapter) SearchIndex(engine, operation string) {
	a.collector.SearchIndex(engine, operation)
}

// NewSearchClient creates a search client from ncore data layer
// Automatically detects and creates adapters for available search engines
func NewSearchClient(d *Data, collector ...metrics.Collector) *search.Client {
	var adapters []search.Adapter
	var c search.Collector

	if len(collector) > 0 && collector[0] != nil {
		c = &SearchCollectorAdapter{collector: collector[0]}
	} else {
		c = &search.NoOpCollector{}
	}

	// Try Elasticsearch
	if factory, err := search.GetAdapterFactory(search.Elasticsearch); err == nil {
		if es := d.GetElasticsearch(); es != nil {
			if adapter, err := factory(es); err == nil {
				adapters = append(adapters, adapter)
			}
		}
	}

	// Try OpenSearch
	if factory, err := search.GetAdapterFactory(search.OpenSearch); err == nil {
		if os := d.GetOpenSearch(); os != nil {
			if adapter, err := factory(os); err == nil {
				adapters = append(adapters, adapter)
			}
		}
	}

	// Try Meilisearch
	if factory, err := search.GetAdapterFactory(search.Meilisearch); err == nil {
		if ms := d.GetMeilisearch(); ms != nil {
			if adapter, err := factory(ms); err == nil {
				adapters = append(adapters, adapter)
			}
		}
	}

	client := search.NewClient(c, adapters...)

	// Configure client if config exists
	if d.conf != nil && d.conf.Search != nil {
		client.UpdateSearchConfig(adaptSearchConfig(d.conf.Search))
	}

	return client
}

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
