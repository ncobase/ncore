package search

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ncobase/ncore/data/config"
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

// Client unified search client with configuration support
type Client struct {
	elasticsearch *elastic.Client
	opensearch    *opensearch.Client
	meilisearch   *meili.Client
	collector     metrics.Collector
	engine        Engine
	indexCache    map[string]bool
	cacheMu       sync.RWMutex
	indexPrefix   string
	searchConfig  *config.Search
}

// NewClient creates new unified search client
func NewClient(es *elastic.Client, os *opensearch.Client, ms *meili.Client, collector metrics.Collector) *Client {
	return NewClientWithConfig(es, os, ms, collector, nil)
}

// NewClientWithPrefix creates new unified search client with index prefix
func NewClientWithPrefix(es *elastic.Client, os *opensearch.Client, ms *meili.Client, collector metrics.Collector, prefix string) *Client {
	searchConfig := &config.Search{
		IndexPrefix:     prefix,
		DefaultEngine:   "elasticsearch",
		AutoCreateIndex: true,
		IndexSettings:   nil,
	}
	return NewClientWithConfig(es, os, ms, collector, searchConfig)
}

// NewClientWithConfig creates new unified search client with search config
func NewClientWithConfig(es *elastic.Client, os *opensearch.Client, ms *meili.Client, collector metrics.Collector, searchConfig *config.Search) *Client {
	if searchConfig == nil {
		searchConfig = &config.Search{
			IndexPrefix:     "",
			DefaultEngine:   "elasticsearch",
			AutoCreateIndex: true,
			IndexSettings:   nil,
		}
	}

	c := &Client{
		elasticsearch: es,
		opensearch:    os,
		meilisearch:   ms,
		collector:     collector,
		indexCache:    make(map[string]bool),
		indexPrefix:   searchConfig.IndexPrefix,
		searchConfig:  searchConfig,
	}

	if collector == nil {
		c.collector = metrics.NoOpCollector{}
	}

	c.setEngine()
	return c
}

// SetIndexPrefix sets the index prefix
func (c *Client) SetIndexPrefix(prefix string) {
	c.indexPrefix = prefix
	if c.searchConfig != nil {
		c.searchConfig.IndexPrefix = prefix
	}
	c.cacheMu.Lock()
	c.indexCache = make(map[string]bool)
	c.cacheMu.Unlock()
}

// GetIndexPrefix returns the current index prefix
func (c *Client) GetIndexPrefix() string {
	return c.indexPrefix
}

// GetSearchConfig returns current search configuration
func (c *Client) GetSearchConfig() *config.Search {
	return c.searchConfig
}

// UpdateSearchConfig updates search configuration
func (c *Client) UpdateSearchConfig(searchConfig *config.Search) {
	c.searchConfig = searchConfig
	if searchConfig != nil {
		c.SetIndexPrefix(searchConfig.IndexPrefix)
	}
}

// buildIndexName builds full index name with prefix
func (c *Client) buildIndexName(index string) string {
	if c.indexPrefix == "" {
		return index
	}
	return fmt.Sprintf("%s-%s", c.indexPrefix, index)
}

// setEngine determines the engine based on availability and config
func (c *Client) setEngine() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use configured default engine if specified and available
	if c.searchConfig != nil && c.searchConfig.DefaultEngine != "" {
		switch c.searchConfig.DefaultEngine {
		case "elasticsearch":
			if c.elasticsearch != nil && c.healthElasticsearch(ctx) == nil {
				c.engine = Elasticsearch
				return
			}
		case "opensearch":
			if c.opensearch != nil && c.healthOpenSearch(ctx) == nil {
				c.engine = OpenSearch
				return
			}
		case "meilisearch":
			if c.meilisearch != nil && c.healthMeilisearch(ctx) == nil {
				c.engine = Meilisearch
				return
			}
		}
	}

	// Fallback to priority order
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

	fullIndex := c.buildIndexName(req.Index)
	prefixedReq := *req
	prefixedReq.Index = fullIndex

	var resp *Response
	var err error

	switch engine {
	case Elasticsearch:
		resp, err = c.searchElasticsearch(ctx, &prefixedReq)
	case OpenSearch:
		resp, err = c.searchOpenSearch(ctx, &prefixedReq)
	case Meilisearch:
		resp, err = c.searchMeilisearch(ctx, &prefixedReq)
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

	fullIndex := c.buildIndexName(req.Index)

	if c.shouldAutoCreateIndex() {
		if err := c.ensureIndex(ctx, engine, fullIndex); err != nil {
			return fmt.Errorf("failed to ensure index exists: %w", err)
		}
	}

	var err error
	switch engine {
	case Elasticsearch:
		err = c.elasticsearch.IndexDocument(ctx, fullIndex, req.DocumentID, req.Document)
	case OpenSearch:
		err = c.opensearch.IndexDocument(ctx, fullIndex, req.DocumentID, req.Document)
	case Meilisearch:
		documents := []any{req.Document}
		if docMap, ok := req.Document.(map[string]any); ok && req.DocumentID != "" {
			docMap["id"] = req.DocumentID
		}
		err = c.meilisearch.IndexDocuments(fullIndex, documents)
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
	fullIndex := c.buildIndexName(index)
	var err error

	switch c.engine {
	case Elasticsearch:
		err = c.elasticsearch.DeleteDocument(ctx, fullIndex, documentID)
	case OpenSearch:
		err = c.opensearch.DeleteDocument(ctx, fullIndex, documentID)
	case Meilisearch:
		err = c.meilisearch.DeleteDocuments(fullIndex, documentID)
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

	fullIndex := c.buildIndexName(index)

	if c.shouldAutoCreateIndex() {
		if err := c.ensureIndex(ctx, engine, fullIndex); err != nil {
			return fmt.Errorf("failed to ensure index exists: %w", err)
		}
	}

	var err error
	switch engine {
	case Elasticsearch:
		err = c.bulkIndexElasticsearch(ctx, fullIndex, documents)
	case OpenSearch:
		err = c.opensearch.BulkIndex(ctx, fullIndex, documents)
	case Meilisearch:
		err = c.meilisearch.IndexDocuments(fullIndex, documents)
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
	fullIndex := c.buildIndexName(index)
	var err error

	switch c.engine {
	case Elasticsearch:
		err = c.bulkDeleteElasticsearch(ctx, fullIndex, documentIDs)
	case OpenSearch:
		for _, docID := range documentIDs {
			if delErr := c.opensearch.DeleteDocument(ctx, fullIndex, docID); delErr != nil {
				err = delErr
				break
			}
		}
	case Meilisearch:
		for _, docID := range documentIDs {
			if delErr := c.meilisearch.DeleteDocuments(fullIndex, docID); delErr != nil {
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

// shouldAutoCreateIndex checks if auto index creation is enabled
func (c *Client) shouldAutoCreateIndex() bool {
	return c.searchConfig != nil && c.searchConfig.AutoCreateIndex
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
				ID     string         `json:"_id"`
				Score  float64        `json:"_score"`
				Source map[string]any `json:"_source"`
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

	query := c.buildElasticsearchQuery(req)
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
		if hitMap, ok := hit.(map[string]any); ok {
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

	return &Response{
		Total: int64(msResp.EstimatedTotalHits),
		Hits:  hits,
	}, nil
}

// buildElasticsearchQuery builds Elasticsearch query with config-aware fields
func (c *Client) buildElasticsearchQuery(req *Request) string {
	searchableFields := c.getSearchableFields()

	if len(req.Filter) == 0 {
		return fmt.Sprintf(`{
			"query": {
				"multi_match": {
					"query": "%s",
					"fields": %s
				}
			},
			"from": %d,
			"size": %d
		}`, req.Query, c.buildFieldsArray(searchableFields), req.From, req.Size)
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
						"fields": %s
					}
				},
				"filter": [%s]
			}
		},
		"from": %d,
		"size": %d
	}`, req.Query, c.buildFieldsArray(searchableFields), strings.Join(filterQueries, ","), req.From, req.Size)
}

// getSearchableFields returns searchable fields from config or defaults
func (c *Client) getSearchableFields() []string {
	if c.searchConfig != nil && c.searchConfig.IndexSettings != nil && len(c.searchConfig.IndexSettings.SearchableFields) > 0 {
		return c.searchConfig.IndexSettings.SearchableFields
	}
	return []string{"title^2", "content", "details", "name", "description"}
}

// buildFieldsArray builds JSON array string from fields
func (c *Client) buildFieldsArray(fields []string) string {
	quotedFields := make([]string, len(fields))
	for i, field := range fields {
		quotedFields[i] = fmt.Sprintf(`"%s"`, field)
	}
	return fmt.Sprintf("[%s]", strings.Join(quotedFields, ","))
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

	c.cacheMu.RLock()
	exists := c.indexCache[cacheKey]
	c.cacheMu.RUnlock()

	if exists {
		return nil
	}

	indexExists, err := c.checkIndexExists(ctx, engine, indexName)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	if indexExists {
		c.cacheMu.Lock()
		c.indexCache[cacheKey] = true
		c.cacheMu.Unlock()
		return nil
	}

	if err := c.createIndexForEngine(ctx, engine, indexName); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

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

	_, err := client.Index(indexName).GetStats()
	if err != nil {
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

	settings := c.buildElasticsearchSettings()
	createRes, err := client.Indices.Create(indexName, client.Indices.Create.WithBody(strings.NewReader(settings)))
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

	settings := c.buildElasticsearchSettings()
	return c.opensearch.CreateIndex(ctx, indexName, settings)
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

	dummyDoc := map[string]any{
		"id":    "init_doc_" + indexName,
		"_init": true,
		"type":  "initialization",
	}

	err := c.meilisearch.IndexDocuments(indexName, []any{dummyDoc}, "id")
	if err != nil {
		return fmt.Errorf("failed to create meilisearch index: %w", err)
	}

	time.Sleep(100 * time.Millisecond)
	_ = c.meilisearch.DeleteDocuments(indexName, "init_doc_"+indexName)

	index := client.Index(indexName)

	// Configure index settings from config
	searchableFields := c.getSearchableFields()
	_, err = index.UpdateSearchableAttributes(&searchableFields)
	if err != nil {
		fmt.Printf("Warning: failed to set searchable attributes for meilisearch index %s: %v\n", indexName, err)
	}

	filterableFields := c.getFilterableFields()
	_, err = index.UpdateFilterableAttributes(&filterableFields)
	if err != nil {
		fmt.Printf("Warning: failed to set filterable attributes for meilisearch index %s: %v\n", indexName, err)
	}

	return nil
}

// buildElasticsearchSettings builds Elasticsearch/OpenSearch index settings
func (c *Client) buildElasticsearchSettings() string {
	shards := 1
	replicas := 0
	refreshInterval := "1s"
	searchableFields := c.getSearchableFields()
	filterableFields := c.getFilterableFields()

	if c.searchConfig != nil && c.searchConfig.IndexSettings != nil {
		if c.searchConfig.IndexSettings.Shards > 0 {
			shards = c.searchConfig.IndexSettings.Shards
		}
		if c.searchConfig.IndexSettings.Replicas >= 0 {
			replicas = c.searchConfig.IndexSettings.Replicas
		}
		if c.searchConfig.IndexSettings.RefreshInterval != "" {
			refreshInterval = c.searchConfig.IndexSettings.RefreshInterval
		}
	}

	// Build properties for searchable and filterable fields
	properties := make(map[string]string)

	// Add searchable fields as text fields
	for _, field := range searchableFields {
		fieldName := strings.TrimSuffix(field, "^2") // Remove boost notation
		properties[fieldName] = `{"type": "text"}`
	}

	// Add filterable fields as keyword fields
	for _, field := range filterableFields {
		if field == "created_at" || field == "updated_at" {
			properties[field] = `{"type": "long"}`
		} else {
			properties[field] = `{"type": "keyword"}`
		}
	}

	// Build properties string
	var propsBuilder strings.Builder
	i := 0
	for field, mapping := range properties {
		if i > 0 {
			propsBuilder.WriteString(",")
		}
		propsBuilder.WriteString(fmt.Sprintf(`"%s": %s`, field, mapping))
		i++
	}

	return fmt.Sprintf(`{
		"settings": {
			"number_of_shards": %d,
			"number_of_replicas": %d,
			"refresh_interval": "%s"
		},
		"mappings": {
			"properties": {
				%s
			}
		}
	}`, shards, replicas, refreshInterval, propsBuilder.String())
}

// getFilterableFields returns filterable fields from config or defaults
func (c *Client) getFilterableFields() []string {
	if c.searchConfig != nil && c.searchConfig.IndexSettings != nil && len(c.searchConfig.IndexSettings.FilterableFields) > 0 {
		return c.searchConfig.IndexSettings.FilterableFields
	}
	return []string{"id", "user_id", "type", "status", "created_at", "updated_at"}
}
