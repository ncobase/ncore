package search

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ncobase/ncore/data/metrics"
	"github.com/ncobase/ncore/data/search/elastic"
	"github.com/ncobase/ncore/data/search/meili"
	"github.com/ncobase/ncore/data/search/opensearch"
	"github.com/ncobase/ncore/utils/convert"
)

var (
	ErrNoEngineAvailable = errors.New("no search engine available")
	ErrEngineNotFound    = errors.New("search engine not found")
)

// Client unified search client
type Client struct {
	elasticsearch *elastic.Client
	opensearch    *opensearch.Client
	meilisearch   *meili.Client
	collector     metrics.Collector
	engine        Engine
	indexCache    map[string]bool
	cacheMu       sync.RWMutex
}

// NewClient creates new unified search client
func NewClient(es *elastic.Client, os *opensearch.Client, ms *meili.Client, collector metrics.Collector) *Client {
	c := &Client{
		elasticsearch: es,
		opensearch:    os,
		meilisearch:   ms,
		collector:     collector,
		indexCache:    make(map[string]bool),
	}

	if collector == nil {
		c.collector = metrics.NoOpCollector{}
	}

	c.setEngine()
	return c
}

// setEngine determines the engine based on availability
func (c *Client) setEngine() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Priority:  OpenSearch > Elasticsearch > Meilisearch
	if c.opensearch != nil && c.healthOpenSearch(ctx) == nil {
		c.engine = OpenSearch
	} else if c.elasticsearch != nil && c.healthElasticsearch(ctx) == nil {
		c.engine = Elasticsearch
	} else if c.meilisearch != nil && c.healthMeilisearch(ctx) == nil {
		c.engine = Meilisearch
	}
}

// Search performs unified search
func (c *Client) Search(ctx context.Context, req *Request) (*Response, error) {
	if c.engine == "" {
		return nil, ErrNoEngineAvailable
	}
	return c.SearchWith(ctx, c.engine, req)
}

// SearchWith searches using specified engine
func (c *Client) SearchWith(ctx context.Context, engine Engine, req *Request) (*Response, error) {
	start := time.Now()

	var resp *Response
	var err error

	switch engine {
	case Elasticsearch:
		resp, err = c.searchElasticsearch(ctx, req)
	case OpenSearch:
		resp, err = c.searchOpenSearch(ctx, req)
	case Meilisearch:
		resp, err = c.searchMeilisearch(ctx, req)
	default:
		err = fmt.Errorf("%w: %s", ErrEngineNotFound, engine)
	}

	// Collect metrics
	duration := time.Since(start)
	c.collector.SearchQuery(string(engine), err)

	if resp != nil {
		resp.Duration = duration
		resp.Engine = engine
	}

	return resp, err
}

// Index indexes document
func (c *Client) Index(ctx context.Context, req *IndexRequest) error {
	if c.engine == "" {
		return ErrNoEngineAvailable
	}
	return c.IndexWith(ctx, c.engine, req)
}

// IndexWith indexes document
func (c *Client) IndexWith(ctx context.Context, engine Engine, req *IndexRequest) error {
	start := time.Now()

	// Ensure index exists before indexing
	if err := c.ensureIndex(ctx, engine, req.Index); err != nil {
		return fmt.Errorf("failed to ensure index exists: %w", err)
	}

	var err error
	switch engine {
	case Elasticsearch:
		err = c.elasticsearch.IndexDocument(ctx, req.Index, req.DocumentID, req.Document)
	case OpenSearch:
		err = c.opensearch.IndexDocument(ctx, req.Index, req.DocumentID, req.Document)
	case Meilisearch:
		documents := []interface{}{req.Document}
		if docMap, ok := req.Document.(map[string]interface{}); ok && req.DocumentID != "" {
			docMap["id"] = req.DocumentID
		}
		err = c.meilisearch.IndexDocuments(req.Index, documents)
	default:
		err = fmt.Errorf("%w: %s", ErrEngineNotFound, engine)
	}

	// Collect metrics
	duration := time.Since(start)
	c.collectMetrics("index", err, duration)
	return err
}

// Delete deletes document
func (c *Client) Delete(ctx context.Context, index, documentID string) error {
	if c.engine == "" {
		return ErrNoEngineAvailable
	}

	start := time.Now()
	var err error

	switch c.engine {
	case Elasticsearch:
		err = c.elasticsearch.DeleteDocument(ctx, index, documentID)
	case OpenSearch:
		err = c.opensearch.DeleteDocument(ctx, index, documentID)
	case Meilisearch:
		err = c.meilisearch.DeleteDocuments(index, documentID)
	default:
		err = fmt.Errorf("%w: %s", ErrEngineNotFound, c.engine)
	}

	// Collect metrics
	duration := time.Since(start)
	c.collectMetrics("delete", err, duration)
	return err
}

// BulkIndex performs bulk indexing
func (c *Client) BulkIndex(ctx context.Context, index string, documents []any) error {
	if c.engine == "" {
		return ErrNoEngineAvailable
	}
	return c.BulkIndexWith(ctx, c.engine, index, documents)
}

// BulkIndexWith performs bulk indexing
func (c *Client) BulkIndexWith(ctx context.Context, engine Engine, index string, documents []any) error {
	start := time.Now()

	// Ensure index exists before bulk indexing
	if err := c.ensureIndex(ctx, engine, index); err != nil {
		return fmt.Errorf("failed to ensure index exists: %w", err)
	}

	var err error
	switch engine {
	case Elasticsearch:
		err = c.bulkIndexElasticsearch(ctx, index, documents)
	case OpenSearch:
		err = c.opensearch.BulkIndex(ctx, index, documents)
	case Meilisearch:
		err = c.meilisearch.IndexDocuments(index, documents)
	default:
		err = fmt.Errorf("%w: %s", ErrEngineNotFound, engine)
	}

	// Collect metrics
	duration := time.Since(start)
	c.collectMetrics("bulk_index", err, duration)
	return err
}

// BulkDelete performs bulk deletion
func (c *Client) BulkDelete(ctx context.Context, index string, documentIDs []string) error {
	if c.engine == "" {
		return ErrNoEngineAvailable
	}

	start := time.Now()
	var err error

	switch c.engine {
	case Elasticsearch:
		err = c.bulkDeleteElasticsearch(ctx, index, documentIDs)
	case OpenSearch:
		for _, docID := range documentIDs {
			if delErr := c.opensearch.DeleteDocument(ctx, index, docID); delErr != nil {
				err = delErr
				break
			}
		}
	case Meilisearch:
		for _, docID := range documentIDs {
			if delErr := c.meilisearch.DeleteDocuments(index, docID); delErr != nil {
				err = delErr
				break
			}
		}
	default:
		err = fmt.Errorf("%w: %s", ErrEngineNotFound, c.engine)
	}

	// Collect metrics
	duration := time.Since(start)
	c.collectMetrics("bulk_delete", err, duration)
	return err
}

// GetAvailableEngines returns available engines
func (c *Client) GetAvailableEngines() []Engine {
	var engines []Engine
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if c.meilisearch != nil && c.healthMeilisearch(ctx) == nil {
		engines = append(engines, Meilisearch)
	}
	if c.elasticsearch != nil && c.healthElasticsearch(ctx) == nil {
		engines = append(engines, Elasticsearch)
	}
	if c.opensearch != nil && c.healthOpenSearch(ctx) == nil {
		engines = append(engines, OpenSearch)
	}

	return engines
}

// GetEngine returns primary engine
func (c *Client) GetEngine() Engine {
	return c.engine
}

// Health checks all engines health
func (c *Client) Health(ctx context.Context) map[Engine]error {
	results := make(map[Engine]error)

	if c.elasticsearch != nil {
		results[Elasticsearch] = c.healthElasticsearch(ctx)
	}
	if c.opensearch != nil {
		results[OpenSearch] = c.healthOpenSearch(ctx)
	}
	if c.meilisearch != nil {
		results[Meilisearch] = c.healthMeilisearch(ctx)
	}

	return results
}

// searchElasticsearch performs Elasticsearch search
func (c *Client) searchElasticsearch(ctx context.Context, req *Request) (*Response, error) {
	if c.elasticsearch == nil {
		return nil, errors.New("elasticsearch client not available")
	}

	query := c.buildElasticsearchQuery(req)
	resp, err := c.elasticsearch.Search(ctx, req.Index, query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("elasticsearch returned status: %d", resp.StatusCode)
	}

	// Parse response
	var esResp struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				ID     string                 `json:"_id"`
				Score  float64                `json:"_score"`
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&esResp); err != nil {
		return nil, err
	}

	hits := make([]Hit, len(esResp.Hits.Hits))
	for i, hit := range esResp.Hits.Hits {
		hits[i] = Hit{
			ID:     hit.ID,
			Score:  hit.Score,
			Source: hit.Source,
		}
	}

	return &Response{
		Total: esResp.Hits.Total.Value,
		Hits:  hits,
	}, nil
}

// searchOpenSearch performs OpenSearch search
func (c *Client) searchOpenSearch(ctx context.Context, req *Request) (*Response, error) {
	if c.opensearch == nil {
		return nil, errors.New("opensearch client not available")
	}

	query := c.buildElasticsearchQuery(req) // OpenSearch uses same query format
	osResp, err := c.opensearch.Search(ctx, req.Index, query)
	if err != nil {
		return nil, err
	}

	if osResp.Errors {
		return nil, errors.New("opensearch returned errors")
	}

	hits := make([]Hit, len(osResp.Hits.Hits))
	for i, hit := range osResp.Hits.Hits {
		source, _ := convert.ToJSONMap(hit.Source)
		hits[i] = Hit{
			ID:     hit.ID,
			Score:  float64(hit.Score),
			Source: source,
		}
	}

	return &Response{
		Total: int64(osResp.Hits.Total.Value),
		Hits:  hits,
	}, nil
}

// searchMeilisearch performs Meilisearch search
func (c *Client) searchMeilisearch(ctx context.Context, req *Request) (*Response, error) {
	if c.meilisearch == nil {
		return nil, errors.New("meilisearch client not available")
	}

	searchReq := &meili.SearchParams{
		Offset: int64(req.From),
		Limit:  int64(req.Size),
	}

	// Add filters
	if len(req.Filter) > 0 {
		filters := make([]string, 0, len(req.Filter))
		for field, value := range req.Filter {
			filters = append(filters, fmt.Sprintf("%s = '%v'", field, value))
		}
		filterStr := strings.Join(filters, " AND ")
		searchReq.Filter = &filterStr
	}

	msResp, err := c.meilisearch.Search(req.Index, req.Query, searchReq)
	if err != nil {
		return nil, err
	}

	hits := make([]Hit, len(msResp.Hits))
	for i, hit := range msResp.Hits {
		if hitMap, ok := hit.(map[string]interface{}); ok {
			var id string
			if idVal, exists := hitMap["id"]; exists {
				id = fmt.Sprintf("%v", idVal)
			}
			hits[i] = Hit{
				ID:     id,
				Score:  1.0,
				Source: hitMap,
			}
		}
	}

	total := int64(msResp.EstimatedTotalHits)
	return &Response{
		Total: total,
		Hits:  hits,
	}, nil
}

// buildElasticsearchQuery builds Elasticsearch query
func (c *Client) buildElasticsearchQuery(req *Request) string {
	if len(req.Filter) == 0 {
		return fmt.Sprintf(`{
			"query": {
				"multi_match": {
					"query": "%s",
					"fields": ["title^2", "content", "details"]
				}
			},
			"from": %d,
			"size": %d
		}`, req.Query, req.From, req.Size)
	}

	filterQueries := make([]string, 0, len(req.Filter))
	for field, value := range req.Filter {
		filterQueries = append(filterQueries, fmt.Sprintf(`{"term": {"%s": "%v"}}`, field, value))
	}

	return fmt.Sprintf(`{
		"query": {
			"bool": {
				"must": {
					"multi_match": {
						"query": "%s",
						"fields": ["title^2", "content", "details"]
					}
				},
				"filter": [%s]
			}
		},
		"from": %d,
		"size": %d
	}`, req.Query, strings.Join(filterQueries, ","), req.From, req.Size)
}

// bulkIndexElasticsearch indexes documents to Elasticsearch
func (c *Client) bulkIndexElasticsearch(ctx context.Context, index string, documents []any) error {
	client := c.elasticsearch.GetClient()
	if client == nil {
		return errors.New("elasticsearch client is nil")
	}

	var bulkBody strings.Builder
	for _, doc := range documents {
		bulkBody.WriteString(fmt.Sprintf(`{"index":{"_index":"%s"}}`, index))
		bulkBody.WriteString("\n")

		docBytes, err := json.Marshal(doc)
		if err != nil {
			return err
		}
		bulkBody.Write(docBytes)
		bulkBody.WriteString("\n")
	}

	res, err := client.Bulk(strings.NewReader(bulkBody.String()),
		client.Bulk.WithIndex(index),
		client.Bulk.WithRefresh("true"))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk index error: %s", res.Status())
	}

	return nil
}

// bulkDeleteElasticsearch deletes documents from Elasticsearch
func (c *Client) bulkDeleteElasticsearch(ctx context.Context, index string, documentIDs []string) error {
	client := c.elasticsearch.GetClient()
	if client == nil {
		return errors.New("elasticsearch client is nil")
	}

	var bulkBody strings.Builder
	for _, docID := range documentIDs {
		bulkBody.WriteString(fmt.Sprintf(`{"delete":{"_index":"%s","_id":"%s"}}`, index, docID))
		bulkBody.WriteString("\n")
	}

	res, err := client.Bulk(strings.NewReader(bulkBody.String()),
		client.Bulk.WithIndex(index),
		client.Bulk.WithRefresh("true"))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk delete error: %s", res.Status())
	}

	return nil
}

// collectMetrics collects metrics
func (c *Client) collectMetrics(operation string, err error, duration time.Duration) {
	if c.collector == nil {
		return
	}

	c.collector.SearchQuery(string(c.engine), err)
	if err == nil {
		c.collector.SearchIndex(string(c.engine), operation)
	}

	if duration > time.Second {
		c.collector.SearchQuery(string(c.engine), errors.New("slow_"+operation))
	}
}

// healthElasticsearch checks Elasticsearch health
func (c *Client) healthElasticsearch(ctx context.Context) error {
	if c.elasticsearch == nil {
		return errors.New("elasticsearch client not available")
	}
	client := c.elasticsearch.GetClient()
	if client == nil {
		return errors.New("elasticsearch client is nil")
	}
	res, err := client.Info()
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("elasticsearch error: %s", res.Status())
	}
	return nil
}

// healthOpenSearch checks OpenSearch health
func (c *Client) healthOpenSearch(ctx context.Context) error {
	if c.opensearch == nil {
		return errors.New("opensearch client not available")
	}
	_, err := c.opensearch.Health(ctx)
	return err
}

// healthMeilisearch checks Meilisearch health
func (c *Client) healthMeilisearch(ctx context.Context) error {
	if c.meilisearch == nil {
		return errors.New("meilisearch client not available")
	}
	client := c.meilisearch.GetClient()
	if client == nil {
		return errors.New("meilisearch client is nil")
	}
	_, err := client.Health()
	return err
}

// ensureIndex ensures the index exists for the specified engine
func (c *Client) ensureIndex(ctx context.Context, engine Engine, indexName string) error {
	cacheKey := fmt.Sprintf("%s:%s", engine, indexName)

	// Check cache first
	c.cacheMu.RLock()
	exists := c.indexCache[cacheKey]
	c.cacheMu.RUnlock()

	if exists {
		return nil
	}

	// Check if index actually exists
	indexExists, err := c.checkIndexExists(ctx, engine, indexName)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	if indexExists {
		// Update cache
		c.cacheMu.Lock()
		c.indexCache[cacheKey] = true
		c.cacheMu.Unlock()
		return nil
	}

	// Create index if it doesn't exist
	if err := c.createIndexForEngine(ctx, engine, indexName); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	// Update cache on success
	c.cacheMu.Lock()
	c.indexCache[cacheKey] = true
	c.cacheMu.Unlock()

	return nil
}

// checkIndexExists checks if index exists for specific engine
func (c *Client) checkIndexExists(ctx context.Context, engine Engine, indexName string) (bool, error) {
	switch engine {
	case Elasticsearch:
		return c.indexExistsElasticsearch(ctx, indexName)
	case OpenSearch:
		return c.indexExistsOpenSearch(ctx, indexName)
	case Meilisearch:
		return c.indexExistsMeilisearch(ctx, indexName)
	default:
		return false, fmt.Errorf("unsupported engine: %s", engine)
	}
}

// createIndexForEngine creates index for specific engine
func (c *Client) createIndexForEngine(ctx context.Context, engine Engine, indexName string) error {
	switch engine {
	case Elasticsearch:
		return c.createElasticsearchIndex(ctx, indexName)
	case OpenSearch:
		return c.createOpenSearchIndex(ctx, indexName)
	case Meilisearch:
		return c.createMeilisearchIndex(ctx, indexName)
	default:
		return fmt.Errorf("unsupported engine: %s", engine)
	}
}

// indexExistsElasticsearch checks if Elasticsearch index exists
func (c *Client) indexExistsElasticsearch(ctx context.Context, indexName string) (bool, error) {
	if c.elasticsearch == nil {
		return false, errors.New("elasticsearch client not available")
	}

	client := c.elasticsearch.GetClient()
	if client == nil {
		return false, errors.New("elasticsearch client is nil")
	}

	res, err := client.Indices.Exists([]string{indexName})
	if err != nil {
		return false, fmt.Errorf("failed to check elasticsearch index existence: %w", err)
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

// indexExistsOpenSearch checks if OpenSearch index exists
func (c *Client) indexExistsOpenSearch(ctx context.Context, indexName string) (bool, error) {
	if c.opensearch == nil {
		return false, errors.New("opensearch client not available")
	}

	return c.opensearch.IndexExists(ctx, indexName)
}

// indexExistsMeilisearch checks if Meilisearch index exists
func (c *Client) indexExistsMeilisearch(ctx context.Context, indexName string) (bool, error) {
	if c.meilisearch == nil {
		return false, errors.New("meilisearch client not available")
	}

	client := c.meilisearch.GetClient()
	if client == nil {
		return false, errors.New("meilisearch client is nil")
	}

	// Try to get index info - if it fails, index doesn't exist
	_, err := client.Index(indexName).GetStats()
	if err != nil {
		// Check if it's a "index not found" error
		if strings.Contains(err.Error(), "index_not_found") ||
			strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check meilisearch index existence: %w", err)
	}

	return true, nil
}

// createElasticsearchIndex creates Elasticsearch index
func (c *Client) createElasticsearchIndex(ctx context.Context, indexName string) error {
	if c.elasticsearch == nil {
		return errors.New("elasticsearch client not available")
	}

	client := c.elasticsearch.GetClient()
	if client == nil {
		return errors.New("elasticsearch client is nil")
	}

	// Create index with default mapping
	mapping := `{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0,
			"refresh_interval": "1s"
		},
		"mappings": {
			"properties": {
				"id": {"type": "keyword"},
				"user_id": {"type": "keyword"},
				"type": {"type": "keyword"},
				"title": {"type": "text"},
				"content": {"type": "text"},
				"details": {"type": "text"},
				"metadata": {"type": "object"},
				"created_at": {"type": "long"},
				"updated_at": {"type": "long"}
			}
		}
	}`

	createRes, err := client.Indices.Create(indexName, client.Indices.Create.WithBody(strings.NewReader(mapping)))
	if err != nil {
		return fmt.Errorf("failed to create elasticsearch index: %w", err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		return fmt.Errorf("elasticsearch index creation failed: %s", createRes.Status())
	}

	return nil
}

// createOpenSearchIndex creates OpenSearch index
func (c *Client) createOpenSearchIndex(ctx context.Context, indexName string) error {
	if c.opensearch == nil {
		return errors.New("opensearch client not available")
	}

	// Create index with default mapping
	mapping := `{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0,
			"refresh_interval": "1s"
		},
		"mappings": {
			"properties": {
				"id": {"type": "keyword"},
				"user_id": {"type": "keyword"},
				"type": {"type": "keyword"},
				"title": {"type": "text"},
				"content": {"type": "text"},
				"details": {"type": "text"},
				"metadata": {"type": "object"},
				"created_at": {"type": "long"},
				"updated_at": {"type": "long"}
			}
		}
	}`

	return c.opensearch.CreateIndex(ctx, indexName, mapping)
}

// createMeilisearchIndex creates Meilisearch index
func (c *Client) createMeilisearchIndex(ctx context.Context, indexName string) error {
	if c.meilisearch == nil {
		return errors.New("meilisearch client not available")
	}

	client := c.meilisearch.GetClient()
	if client == nil {
		return errors.New("meilisearch client is nil")
	}

	// For Meilisearch, we create index by indexing a dummy document
	// This is because Meilisearch creates indexes automatically when documents are added
	dummyDoc := map[string]interface{}{
		"id":    "init_doc_" + indexName,
		"_init": true,
		"type":  "initialization",
	}

	err := c.meilisearch.IndexDocuments(indexName, []interface{}{dummyDoc}, "id")
	if err != nil {
		return fmt.Errorf("failed to create meilisearch index: %w", err)
	}

	// Wait a bit for index creation and then delete the dummy document
	time.Sleep(100 * time.Millisecond)
	_ = c.meilisearch.DeleteDocuments(indexName, "init_doc_"+indexName)

	// Configure searchable attributes and other settings
	index := client.Index(indexName)

	// Set searchable attributes
	_, err = index.UpdateSearchableAttributes(&[]string{
		"title", "content", "details", "type", "user_id",
	})
	if err != nil {
		// Log warning but don't fail - index is still usable
		fmt.Printf("Warning: failed to set searchable attributes for meilisearch index %s: %v\n", indexName, err)
	}

	// Set filterable attributes
	_, err = index.UpdateFilterableAttributes(&[]string{
		"id", "user_id", "type", "created_at", "updated_at",
	})
	if err != nil {
		// Log warning but don't fail
		fmt.Printf("Warning: failed to set filterable attributes for meilisearch index %s: %v\n", indexName, err)
	}

	return nil
}
