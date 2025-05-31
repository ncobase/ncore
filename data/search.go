package data

import (
	"context"
	"errors"

	"github.com/ncobase/ncore/data/search"
)

// Search Operations

// Search performs search using best available engine
func (d *Data) Search(ctx context.Context, req *search.Request) (*search.Response, error) {
	client := d.getSearchClient()
	if client == nil {
		return nil, errors.New("search client not available")
	}
	return client.Search(ctx, req)
}

// SearchWith performs search using specified engine
func (d *Data) SearchWith(ctx context.Context, engine search.Engine, req *search.Request) (*search.Response, error) {
	client := d.getSearchClient()
	if client == nil {
		return nil, errors.New("search client not available")
	}
	return client.SearchWith(ctx, engine, req)
}

// Document Operations

// IndexDocument indexes document using best available engine
func (d *Data) IndexDocument(ctx context.Context, req *search.IndexRequest) error {
	client := d.getSearchClient()
	if client == nil {
		return errors.New("search client not available")
	}
	return client.Index(ctx, req)
}

// IndexDocumentWith indexes document using specified engine
func (d *Data) IndexDocumentWith(ctx context.Context, engine search.Engine, req *search.IndexRequest) error {
	client := d.getSearchClient()
	if client == nil {
		return errors.New("search client not available")
	}
	return client.IndexWith(ctx, engine, req)
}

// DeleteDocument deletes document using best available engine
func (d *Data) DeleteDocument(ctx context.Context, index, documentID string) error {
	client := d.getSearchClient()
	if client == nil {
		return errors.New("search client not available")
	}
	return client.Delete(ctx, index, documentID)
}

// Bulk Operations

// BulkIndexDocuments indexes multiple documents using best available engine
func (d *Data) BulkIndexDocuments(ctx context.Context, index string, documents []any) error {
	client := d.getSearchClient()
	if client == nil {
		return errors.New("search client not available")
	}
	return client.BulkIndex(ctx, index, documents)
}

// BulkIndexDocumentsWith indexes multiple documents using specified engine
func (d *Data) BulkIndexDocumentsWith(ctx context.Context, engine search.Engine, index string, documents []any) error {
	client := d.getSearchClient()
	if client == nil {
		return errors.New("search client not available")
	}
	return client.BulkIndexWith(ctx, engine, index, documents)
}

// BulkDeleteDocuments deletes multiple documents using best available engine
func (d *Data) BulkDeleteDocuments(ctx context.Context, index string, documentIDs []string) error {
	client := d.getSearchClient()
	if client == nil {
		return errors.New("search client not available")
	}
	return client.BulkDelete(ctx, index, documentIDs)
}

// Engine Management

// GetAvailableSearchEngines returns available search engines
func (d *Data) GetAvailableSearchEngines() []search.Engine {
	client := d.getSearchClient()
	if client == nil {
		return nil
	}
	return client.GetAvailableEngines()
}

// GetSearchEngine returns search engine
func (d *Data) GetSearchEngine() search.Engine {
	client := d.getSearchClient()
	if client == nil {
		return ""
	}
	return client.GetEngine()
}

// SearchHealth checks search engines health
func (d *Data) SearchHealth(ctx context.Context) map[search.Engine]error {
	client := d.getSearchClient()
	if client == nil {
		return nil
	}
	return client.Health(ctx)
}

// Legacy methods for backward compatibility (deprecated)

// SearchWithMeilisearch searches with Meilisearch
// Deprecated: Use Search() or SearchWith() instead
func (d *Data) SearchWithMeilisearch(ctx context.Context, index, query string) error {
	req := &search.Request{Index: index, Query: query}
	_, err := d.SearchWith(ctx, search.Meilisearch, req)
	return err
}

// IndexWithMeilisearch indexes documents with Meilisearch
// Deprecated: Use IndexDocument() or IndexDocumentWith() instead
func (d *Data) IndexWithMeilisearch(index string, documents any, primaryKey ...string) error {
	req := &search.IndexRequest{Index: index, Document: documents}
	return d.IndexDocumentWith(context.Background(), search.Meilisearch, req)
}

// SearchWithElasticsearch searches with Elasticsearch
// Deprecated: Use Search() or SearchWith() instead
func (d *Data) SearchWithElasticsearch(ctx context.Context, index, query string) error {
	req := &search.Request{Index: index, Query: query}
	_, err := d.SearchWith(ctx, search.Elasticsearch, req)
	return err
}

// IndexWithElasticsearch indexes documents with Elasticsearch
// Deprecated: Use IndexDocument() or IndexDocumentWith() instead
func (d *Data) IndexWithElasticsearch(ctx context.Context, index, documentID string, document any) error {
	req := &search.IndexRequest{Index: index, DocumentID: documentID, Document: document}
	return d.IndexDocumentWith(ctx, search.Elasticsearch, req)
}

// SearchWithOpenSearch searches with OpenSearch
// Deprecated: Use Search() or SearchWith() instead
func (d *Data) SearchWithOpenSearch(ctx context.Context, index, query string) error {
	req := &search.Request{Index: index, Query: query}
	_, err := d.SearchWith(ctx, search.OpenSearch, req)
	return err
}

// IndexWithOpenSearch indexes documents with OpenSearch
// Deprecated: Use IndexDocument() or IndexDocumentWith() instead
func (d *Data) IndexWithOpenSearch(ctx context.Context, index, documentID string, document any) error {
	req := &search.IndexRequest{Index: index, DocumentID: documentID, Document: document}
	return d.IndexDocumentWith(ctx, search.OpenSearch, req)
}
