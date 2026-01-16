package search

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrNoEngineAvailable = errors.New("no search engine available")
	ErrEngineNotFound    = errors.New("search engine not found")
)

type Engine string

const (
	Elasticsearch Engine = "elasticsearch"
	OpenSearch    Engine = "opensearch"
	Meilisearch   Engine = "meilisearch"
)

// Config represents search engine configuration
type Config struct {
	IndexPrefix     string
	DefaultEngine   string
	AutoCreateIndex bool
	IndexSettings   *IndexSettings
}

// IndexSettings represents default index configuration
type IndexSettings struct {
	Shards           int
	Replicas         int
	RefreshInterval  string
	SearchableFields []string
	FilterableFields []string
}

// Collector interface for metrics
type Collector interface {
	SearchQuery(engine string, err error)
	SearchIndex(engine, operation string)
}

// NoOpCollector implementation
type NoOpCollector struct{}

func (NoOpCollector) SearchQuery(string, error)  {}
func (NoOpCollector) SearchIndex(string, string) {}

type Request struct {
	Index  string         `json:"index"`
	Query  string         `json:"query"`
	Filter map[string]any `json:"filter,omitempty"`
	From   int            `json:"from,omitempty"`
	Size   int            `json:"size,omitempty"`
}

type Response struct {
	Total    int64         `json:"total"`
	Hits     []Hit         `json:"hits"`
	Duration time.Duration `json:"duration"`
	Engine   Engine        `json:"engine"`
}

type Hit struct {
	ID     string         `json:"id"`
	Score  float64        `json:"score"`
	Source map[string]any `json:"source"`
}

type IndexRequest struct {
	Index      string `json:"index"`
	DocumentID string `json:"document_id,omitempty"`
	Document   any    `json:"document"`
}

// Adapter interface for search engine implementations
type Adapter interface {
	Search(ctx context.Context, req *Request) (*Response, error)
	Index(ctx context.Context, req *IndexRequest) error
	Delete(ctx context.Context, index, id string) error
	BulkIndex(ctx context.Context, index string, documents []any) error
	BulkDelete(ctx context.Context, index string, documentIDs []string) error
	IndexExists(ctx context.Context, indexName string) (bool, error)
	CreateIndex(ctx context.Context, indexName string, settings *IndexSettings) error
	Health(ctx context.Context) error
	Type() Engine
}

type Client struct {
	adapters     map[Engine]Adapter
	collector    Collector
	engine       Engine
	indexCache   map[string]bool
	cacheMu      sync.RWMutex
	indexPrefix  string
	searchConfig *Config
}

// NewClient creates a new search client with provided adapters
func NewClient(collector Collector, adapters ...Adapter) *Client {
	return NewClientWithConfig(collector, nil, adapters...)
}

// NewClientWithPrefix creates a new search client with index prefix
func NewClientWithPrefix(collector Collector, prefix string, adapters ...Adapter) *Client {
	searchConfig := &Config{
		IndexPrefix:     prefix,
		DefaultEngine:   "elasticsearch",
		AutoCreateIndex: true,
		IndexSettings:   nil,
	}
	return NewClientWithConfig(collector, searchConfig, adapters...)
}

// NewClientWithConfig creates a new search client with configuration
func NewClientWithConfig(collector Collector, searchConfig *Config, adapters ...Adapter) *Client {
	if searchConfig == nil {
		searchConfig = &Config{
			IndexPrefix:     "",
			DefaultEngine:   "elasticsearch",
			AutoCreateIndex: true,
			IndexSettings:   nil,
		}
	}

	adapterMap := make(map[Engine]Adapter)
	for _, a := range adapters {
		adapterMap[a.Type()] = a
	}

	if collector == nil {
		collector = NoOpCollector{}
	}

	c := &Client{
		adapters:     adapterMap,
		collector:    collector,
		indexCache:   make(map[string]bool),
		indexPrefix:  searchConfig.IndexPrefix,
		searchConfig: searchConfig,
	}

	c.setEngine()
	return c
}

func (c *Client) SetIndexPrefix(prefix string) {
	c.indexPrefix = prefix
	if c.searchConfig != nil {
		c.searchConfig.IndexPrefix = prefix
	}
	c.cacheMu.Lock()
	c.indexCache = make(map[string]bool)
	c.cacheMu.Unlock()
}

func (c *Client) GetIndexPrefix() string {
	return c.indexPrefix
}

func (c *Client) GetSearchConfig() *Config {
	return c.searchConfig
}

func (c *Client) UpdateSearchConfig(searchConfig *Config) {
	c.searchConfig = searchConfig
	if searchConfig != nil {
		c.SetIndexPrefix(searchConfig.IndexPrefix)
	}
	c.setEngine()
}

func (c *Client) buildIndexName(index string) string {
	if c.indexPrefix == "" {
		return index
	}
	return fmt.Sprintf("%s-%s", c.indexPrefix, index)
}

func (c *Client) setEngine() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use configured default engine if specified and available
	if c.searchConfig != nil && c.searchConfig.DefaultEngine != "" {
		eng := Engine(c.searchConfig.DefaultEngine)
		if adapter, ok := c.adapters[eng]; ok {
			if adapter.Health(ctx) == nil {
				c.engine = eng
				return
			}
		}
	}

	// Fallback to priority order
	// Priority: OpenSearch > Elasticsearch > Meilisearch
	priority := []Engine{OpenSearch, Elasticsearch, Meilisearch}

	for _, eng := range priority {
		if adapter, ok := c.adapters[eng]; ok {
			if adapter.Health(ctx) == nil {
				c.engine = eng
				return
			}
		}
	}

	// Fallback to any available
	for eng, adapter := range c.adapters {
		if adapter.Health(ctx) == nil {
			c.engine = eng
			return
		}
	}
}

func (c *Client) getAdapter() (Adapter, error) {
	if c.engine == "" {
		c.setEngine() // Try to set engine if not set
		if c.engine == "" {
			return nil, ErrNoEngineAvailable
		}
	}

	if adapter, ok := c.adapters[c.engine]; ok {
		return adapter, nil
	}
	return nil, ErrEngineNotFound
}

func (c *Client) Search(ctx context.Context, req *Request) (*Response, error) {
	_, err := c.getAdapter()
	if err != nil {
		return nil, err
	}
	return c.SearchWith(ctx, c.engine, req)
}

func (c *Client) SearchWith(ctx context.Context, engine Engine, req *Request) (*Response, error) {
	adapter, ok := c.adapters[engine]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrEngineNotFound, engine)
	}

	start := time.Now()

	fullIndex := c.buildIndexName(req.Index)
	prefixedReq := *req
	prefixedReq.Index = fullIndex

	resp, err := adapter.Search(ctx, &prefixedReq)

	// Collect metrics
	duration := time.Since(start)
	c.collector.SearchQuery(string(engine), err)

	if resp != nil {
		resp.Duration = duration
		resp.Engine = engine
	}

	return resp, err
}

func (c *Client) Index(ctx context.Context, req *IndexRequest) error {
	_, err := c.getAdapter()
	if err != nil {
		return err
	}
	return c.IndexWith(ctx, c.engine, req)
}

func (c *Client) IndexWith(ctx context.Context, engine Engine, req *IndexRequest) error {
	adapter, ok := c.adapters[engine]
	if !ok {
		return fmt.Errorf("%w: %s", ErrEngineNotFound, engine)
	}

	start := time.Now()

	fullIndex := c.buildIndexName(req.Index)
	// Create a copy of request with full index name
	prefixedReq := *req
	prefixedReq.Index = fullIndex

	if c.shouldAutoCreateIndex() {
		if err := c.ensureIndex(ctx, engine, fullIndex); err != nil {
			return fmt.Errorf("failed to ensure index exists: %w", err)
		}
	}

	err := adapter.Index(ctx, &prefixedReq)

	// Collect metrics
	duration := time.Since(start)
	c.collectMetrics("index", err, duration)
	return err
}

func (c *Client) Delete(ctx context.Context, index, documentID string) error {
	adapter, err := c.getAdapter()
	if err != nil {
		return err
	}

	start := time.Now()
	fullIndex := c.buildIndexName(index)

	err = adapter.Delete(ctx, fullIndex, documentID)

	// Collect metrics
	duration := time.Since(start)
	c.collectMetrics("delete", err, duration)
	return err
}

func (c *Client) BulkIndex(ctx context.Context, index string, documents []any) error {
	_, err := c.getAdapter()
	if err != nil {
		return err
	}
	return c.BulkIndexWith(ctx, c.engine, index, documents)
}

func (c *Client) BulkIndexWith(ctx context.Context, engine Engine, index string, documents []any) error {
	adapter, ok := c.adapters[engine]
	if !ok {
		return fmt.Errorf("%w: %s", ErrEngineNotFound, engine)
	}

	start := time.Now()
	fullIndex := c.buildIndexName(index)

	if c.shouldAutoCreateIndex() {
		if err := c.ensureIndex(ctx, engine, fullIndex); err != nil {
			return fmt.Errorf("failed to ensure index exists: %w", err)
		}
	}

	err := adapter.BulkIndex(ctx, fullIndex, documents)

	// Collect metrics
	duration := time.Since(start)
	c.collectMetrics("bulk_index", err, duration)
	return err
}

func (c *Client) BulkDelete(ctx context.Context, index string, documentIDs []string) error {
	adapter, err := c.getAdapter()
	if err != nil {
		return err
	}

	start := time.Now()
	fullIndex := c.buildIndexName(index)

	err = adapter.BulkDelete(ctx, fullIndex, documentIDs)

	// Collect metrics
	duration := time.Since(start)
	c.collectMetrics("bulk_delete", err, duration)
	return err
}

func (c *Client) shouldAutoCreateIndex() bool {
	return c.searchConfig != nil && c.searchConfig.AutoCreateIndex
}

func (c *Client) GetAvailableEngines() []Engine {
	var engines []Engine
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for eng, adapter := range c.adapters {
		if adapter.Health(ctx) == nil {
			engines = append(engines, eng)
		}
	}

	return engines
}

func (c *Client) GetEngine() Engine {
	return c.engine
}

func (c *Client) Health(ctx context.Context) map[Engine]error {
	results := make(map[Engine]error)

	for eng, adapter := range c.adapters {
		results[eng] = adapter.Health(ctx)
	}

	return results
}

func (c *Client) collectMetrics(operation string, err error, duration time.Duration) {
	if c.collector == nil {
		return
	}

	c.collector.SearchQuery(string(c.engine), err)
	if err == nil {
		c.collector.SearchIndex(string(c.engine), operation)
	}
}

func (c *Client) ensureIndex(ctx context.Context, engine Engine, indexName string) error {
	cacheKey := fmt.Sprintf("%s:%s", engine, indexName)

	c.cacheMu.RLock()
	exists := c.indexCache[cacheKey]
	c.cacheMu.RUnlock()

	if exists {
		return nil
	}

	adapter, ok := c.adapters[engine]
	if !ok {
		return fmt.Errorf("%w: %s", ErrEngineNotFound, engine)
	}

	indexExists, err := adapter.IndexExists(ctx, indexName)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	if indexExists {
		c.cacheMu.Lock()
		c.indexCache[cacheKey] = true
		c.cacheMu.Unlock()
		return nil
	}

	if err := adapter.CreateIndex(ctx, indexName, c.searchConfig.IndexSettings); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	c.cacheMu.Lock()
	c.indexCache[cacheKey] = true
	c.cacheMu.Unlock()

	return nil
}
