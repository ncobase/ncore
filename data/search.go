package data

import (
	"context"
	"errors"
	"time"
)

// SearchWithMeilisearch searches with Meilisearch
func (d *Data) SearchWithMeilisearch(_ context.Context, index, query string) error {
	start := time.Now()
	ms := d.GetMeilisearch()
	if ms == nil {
		err := errors.New("meilisearch not available")
		d.collector.SearchQuery("meilisearch", err)
		return err
	}

	_, err := ms.Search(index, query, nil)
	duration := time.Since(start)

	// Track metrics
	d.collector.SearchQuery("meilisearch", err)
	if err == nil {
		d.collector.SearchIndex("meilisearch", "search")
	}

	if duration > time.Second {
		d.collector.SearchQuery("meilisearch", errors.New("slow_search"))
	}

	return err
}

// IndexWithMeilisearch indexes documents with Meilisearch
func (d *Data) IndexWithMeilisearch(index string, documents any, primaryKey ...string) error {
	start := time.Now()
	ms := d.GetMeilisearch()
	if ms == nil {
		err := errors.New("meilisearch not available")
		d.collector.SearchIndex("meilisearch", err.Error())
		return err
	}

	err := ms.IndexDocuments(index, documents, primaryKey...)
	duration := time.Since(start)

	// Track metrics
	d.collector.SearchIndex("meilisearch", "index")
	if duration > time.Second {
		// Track slow indexing operations
		d.collector.SearchQuery("meilisearch", errors.New("slow_index"))
	}

	return err
}

// SearchWithElasticsearch searches with Elasticsearch
func (d *Data) SearchWithElasticsearch(ctx context.Context, index, query string) error {
	start := time.Now()
	es := d.GetElasticsearch()
	if es == nil {
		err := errors.New("elasticsearch not available")
		d.collector.SearchQuery("elasticsearch", err)
		return err
	}

	_, err := es.Search(ctx, index, query)
	duration := time.Since(start)

	// Track metrics
	d.collector.SearchQuery("elasticsearch", err)
	if err == nil {
		d.collector.SearchIndex("elasticsearch", "search")
	}
	if duration > time.Second {
		d.collector.SearchQuery("elasticsearch", errors.New("slow_search"))
	}

	return err
}

// IndexWithElasticsearch indexes documents with Elasticsearch
func (d *Data) IndexWithElasticsearch(ctx context.Context, index, documentID string, document any) error {
	start := time.Now()
	es := d.GetElasticsearch()
	if es == nil {
		err := errors.New("elasticsearch not available")
		d.collector.SearchIndex("elasticsearch", err.Error())
		return err
	}

	err := es.IndexDocument(ctx, index, documentID, document)
	duration := time.Since(start)

	// Track metrics
	d.collector.SearchIndex("elasticsearch", "index")
	if duration > time.Second {
		d.collector.SearchQuery("elasticsearch", errors.New("slow_index"))
	}

	return err
}

// SearchWithOpenSearch searches with OpenSearch
func (d *Data) SearchWithOpenSearch(ctx context.Context, index, query string) error {
	start := time.Now()
	os := d.GetOpenSearch()
	if os == nil {
		err := errors.New("opensearch not available")
		d.collector.SearchQuery("opensearch", err)
		return err
	}

	_, err := os.Search(ctx, index, query)
	duration := time.Since(start)

	// Track metrics
	d.collector.SearchQuery("opensearch", err)
	if err == nil {
		d.collector.SearchIndex("opensearch", "search")
	}
	if duration > time.Second {
		d.collector.SearchQuery("opensearch", errors.New("slow_search"))
	}

	return err
}

// IndexWithOpenSearch indexes documents with OpenSearch
func (d *Data) IndexWithOpenSearch(ctx context.Context, index, documentID string, document any) error {
	start := time.Time{}
	os := d.GetOpenSearch()
	if os == nil {
		err := errors.New("opensearch not available")
		d.collector.SearchIndex("opensearch", err.Error())
		return err
	}

	err := os.IndexDocument(ctx, index, documentID, document)
	duration := time.Since(start)

	// Track metrics
	d.collector.SearchIndex("opensearch", "index")
	if duration > time.Second {
		d.collector.SearchQuery("opensearch", errors.New("slow_index"))
	}

	return err
}
