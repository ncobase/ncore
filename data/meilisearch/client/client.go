package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/meilisearch/meilisearch-go"
)

// Client Meilisearch client wrapper
type Client struct {
	client meilisearch.ServiceManager
}

// SearchParams is an alias for meilisearch.SearchRequest type
type SearchParams = meilisearch.SearchRequest

// NewMeilisearch creates new Meilisearch client
func NewMeilisearch(host, apiKey string) *Client {
	if host == "" {
		return &Client{client: nil}
	}
	ms := meilisearch.New(host, meilisearch.WithAPIKey(apiKey))
	return &Client{client: ms}
}

// GetClient returns the underlying meilisearch client
func (c *Client) GetClient() meilisearch.ServiceManager {
	return c.client
}

// Search searches from Meilisearch
func (c *Client) Search(index, query string, options *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot perform search")
	}
	resp, err := c.client.Index(index).Search(query, options)
	if err != nil {
		return nil, fmt.Errorf("meilisearch search error: %v", err)
	}
	return resp, nil
}

// IndexDocuments indexes documents to Meilisearch (alias for AddDocuments)
func (c *Client) IndexDocuments(index string, documents any, primaryKey ...string) error {
	return c.AddDocuments(index, documents, primaryKey...)
}

// AddDocuments adds documents to Meilisearch
func (c *Client) AddDocuments(index string, documents any, primaryKey ...string) error {
	if c == nil || c.client == nil {
		return errors.New("meilisearch client is nil, cannot add documents")
	}

	// Convert documents to slice if it's not already
	var docs []any
	switch v := documents.(type) {
	case []any:
		docs = v
	case []map[string]any:
		docs = make([]any, len(v))
		for i, doc := range v {
			docs[i] = doc
		}
	case map[string]any:
		docs = []any{v}
	default:
		docs = []any{documents}
	}

	// Convert variadic primaryKey to *string
	var pk *string
	if len(primaryKey) > 0 && primaryKey[0] != "" {
		pk = &primaryKey[0]
	}

	_, err := c.client.Index(index).AddDocuments(docs, &meilisearch.DocumentOptions{PrimaryKey: pk})
	if err != nil {
		return fmt.Errorf("meilisearch add documents error: %v", err)
	}
	return nil
}

// UpdateDocuments updates documents in Meilisearch
func (c *Client) UpdateDocuments(index string, documents any, primaryKey ...string) error {
	if c == nil || c.client == nil {
		return errors.New("meilisearch client is nil, cannot update documents")
	}

	// Convert documents to slice if it's not already
	var docs []any
	switch v := documents.(type) {
	case []any:
		docs = v
	case []map[string]any:
		docs = make([]any, len(v))
		for i, doc := range v {
			docs[i] = doc
		}
	case map[string]any:
		docs = []any{v}
	default:
		docs = []any{documents}
	}

	// Convert variadic primaryKey to *string
	var pk *string
	if len(primaryKey) > 0 && primaryKey[0] != "" {
		pk = &primaryKey[0]
	}

	_, err := c.client.Index(index).UpdateDocuments(docs, &meilisearch.DocumentOptions{PrimaryKey: pk})
	if err != nil {
		return fmt.Errorf("meilisearch update documents error: %v", err)
	}
	return nil
}

// AddDocumentsInBatches adds documents in batches to Meilisearch
func (c *Client) AddDocumentsInBatches(index string, documents any, batchSize int, primaryKey ...string) error {
	if c == nil || c.client == nil {
		return errors.New("meilisearch client is nil, cannot add documents in batches")
	}

	// Convert documents to slice if it's not already
	var docs []any
	switch v := documents.(type) {
	case []any:
		docs = v
	case []map[string]any:
		docs = make([]any, len(v))
		for i, doc := range v {
			docs[i] = doc
		}
	default:
		return fmt.Errorf("unsupported document type for batch operation: %T", documents)
	}

	// Convert variadic primaryKey to *string
	var pk *string
	if len(primaryKey) > 0 && primaryKey[0] != "" {
		pk = &primaryKey[0]
	}

	_, err := c.client.Index(index).AddDocumentsInBatches(docs, batchSize, &meilisearch.DocumentOptions{PrimaryKey: pk})
	if err != nil {
		return fmt.Errorf("meilisearch add documents in batches error: %v", err)
	}
	return nil
}

// UpdateDocumentsInBatches updates documents in batches in Meilisearch
func (c *Client) UpdateDocumentsInBatches(index string, documents any, batchSize int, primaryKey ...string) error {
	if c == nil || c.client == nil {
		return errors.New("meilisearch client is nil, cannot update documents in batches")
	}

	// Convert documents to slice if it's not already
	var docs []any
	switch v := documents.(type) {
	case []any:
		docs = v
	case []map[string]any:
		docs = make([]any, len(v))
		for i, doc := range v {
			docs[i] = doc
		}
	default:
		return fmt.Errorf("unsupported document type for batch operation: %T", documents)
	}

	// Convert variadic primaryKey to *string
	var pk *string
	if len(primaryKey) > 0 && primaryKey[0] != "" {
		pk = &primaryKey[0]
	}

	_, err := c.client.Index(index).UpdateDocumentsInBatches(docs, batchSize, &meilisearch.DocumentOptions{PrimaryKey: pk})
	if err != nil {
		return fmt.Errorf("meilisearch update documents in batches error: %v", err)
	}
	return nil
}

// GetDocument gets a single document from Meilisearch
func (c *Client) GetDocument(index, documentID string, request *meilisearch.DocumentQuery, documentPtr any) error {
	if c == nil || c.client == nil {
		return errors.New("meilisearch client is nil, cannot get document")
	}

	err := c.client.Index(index).GetDocument(documentID, request, documentPtr)
	if err != nil {
		return fmt.Errorf("meilisearch get document error: %v", err)
	}
	return nil
}

// GetDocuments gets multiple documents from Meilisearch
func (c *Client) GetDocuments(index string, request *meilisearch.DocumentsQuery) (*meilisearch.DocumentsResult, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot get documents")
	}

	var result meilisearch.DocumentsResult
	err := c.client.Index(index).GetDocuments(request, &result)
	if err != nil {
		return nil, fmt.Errorf("meilisearch get documents error: %v", err)
	}
	return &result, nil
}

// DeleteDocument deletes a single document from Meilisearch
func (c *Client) DeleteDocument(index, documentID string) error {
	if c == nil || c.client == nil {
		return errors.New("meilisearch client is nil, cannot delete document")
	}

	_, err := c.client.Index(index).DeleteDocument(documentID, nil)
	if err != nil {
		return fmt.Errorf("meilisearch delete document error: %v", err)
	}
	return nil
}

// DeleteDocuments deletes multiple documents from Meilisearch
func (c *Client) DeleteDocuments(index string, documentIDs ...string) error {
	if c == nil || c.client == nil {
		return errors.New("meilisearch client is nil, cannot delete documents")
	}

	if len(documentIDs) == 0 {
		return errors.New("no document IDs provided")
	}

	// For single document deletion
	if len(documentIDs) == 1 {
		return c.DeleteDocument(index, documentIDs[0])
	}

	// For multiple document deletion
	_, err := c.client.Index(index).DeleteDocuments(documentIDs, nil)
	if err != nil {
		return fmt.Errorf("meilisearch delete documents error: %v", err)
	}
	return nil
}

// DeleteAllDocuments deletes all documents from an index
func (c *Client) DeleteAllDocuments(index string) error {
	if c == nil || c.client == nil {
		return errors.New("meilisearch client is nil, cannot delete all documents")
	}

	_, err := c.client.Index(index).DeleteAllDocuments(nil)
	if err != nil {
		return fmt.Errorf("meilisearch delete all documents error: %v", err)
	}
	return nil
}

// DeleteDocumentsByFilter deletes documents by filter
func (c *Client) DeleteDocumentsByFilter(index string, filter any) error {
	if c == nil || c.client == nil {
		return errors.New("meilisearch client is nil, cannot delete documents by filter")
	}

	_, err := c.client.Index(index).DeleteDocumentsByFilter(filter, nil)
	if err != nil {
		return fmt.Errorf("meilisearch delete documents by filter error: %v", err)
	}
	return nil
}

// Health checks if Meilisearch is healthy
func (c *Client) Health() (*meilisearch.Health, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot check health")
	}

	health, err := c.client.Health()
	if err != nil {
		return nil, fmt.Errorf("meilisearch health check error: %v", err)
	}
	return health, nil
}

// IsHealthy checks if Meilisearch is healthy (convenience method)
func (c *Client) IsHealthy() bool {
	if c == nil || c.client == nil {
		return false
	}
	return c.client.IsHealthy()
}

// GetIndexes gets all indexes from Meilisearch
func (c *Client) GetIndexes(query *meilisearch.IndexesQuery) (*meilisearch.IndexesResults, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot get indexes")
	}

	// Use the correct method from the official SDK
	indexes, err := c.client.ListIndexes(query)
	if err != nil {
		return nil, fmt.Errorf("meilisearch get indexes error: %v", err)
	}
	return indexes, nil
}

// ListIndexes is an alias for GetIndexes for consistency with official SDK
func (c *Client) ListIndexes(query *meilisearch.IndexesQuery) (*meilisearch.IndexesResults, error) {
	return c.GetIndexes(query)
}

// GetIndex gets a specific index from Meilisearch
func (c *Client) GetIndex(indexUID string) (*meilisearch.IndexResult, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot get index")
	}

	index, err := c.client.GetIndex(indexUID)
	if err != nil {
		return nil, fmt.Errorf("meilisearch get index error: %v", err)
	}
	return index, nil
}

// CreateIndex creates a new index in Meilisearch
func (c *Client) CreateIndex(config *meilisearch.IndexConfig) (*meilisearch.TaskInfo, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot create index")
	}

	taskInfo, err := c.client.CreateIndex(config)
	if err != nil {
		return nil, fmt.Errorf("meilisearch create index error: %v", err)
	}
	return taskInfo, nil
}

// DeleteIndex deletes an index from Meilisearch
func (c *Client) DeleteIndex(indexUID string) (*meilisearch.TaskInfo, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot delete index")
	}

	taskInfo, err := c.client.DeleteIndex(indexUID)
	if err != nil {
		return nil, fmt.Errorf("meilisearch delete index error: %v", err)
	}
	return taskInfo, nil
}

// UpdateIndex updates an index in Meilisearch
func (c *Client) UpdateIndex(indexUID, primaryKey string) (*meilisearch.TaskInfo, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot update index")
	}

	// Create UpdateIndexRequestParams with the primary key
	params := &meilisearch.UpdateIndexRequestParams{
		PrimaryKey: primaryKey,
	}

	taskInfo, err := c.client.Index(indexUID).UpdateIndex(params)
	if err != nil {
		return nil, fmt.Errorf("meilisearch update index error: %v", err)
	}
	return taskInfo, nil
}

// GetTask gets task information
func (c *Client) GetTask(taskUID int64) (*meilisearch.Task, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot get task")
	}

	task, err := c.client.GetTask(taskUID)
	if err != nil {
		return nil, fmt.Errorf("meilisearch get task error: %v", err)
	}
	return task, nil
}

// GetTasks gets multiple tasks
func (c *Client) GetTasks(param *meilisearch.TasksQuery) (*meilisearch.TaskResult, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot get tasks")
	}

	tasks, err := c.client.GetTasks(param)
	if err != nil {
		return nil, fmt.Errorf("meilisearch get tasks error: %v", err)
	}
	return tasks, nil
}

// WaitForTask waits for a task to complete
func (c *Client) WaitForTask(taskUID int64) (*meilisearch.Task, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot wait for task")
	}

	// Use a reasonable interval for waiting
	task, err := c.client.WaitForTask(taskUID, 50*time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("meilisearch wait for task error: %v", err)
	}
	return task, nil
}

// WaitForTaskWithInterval waits for a task to complete with custom interval
func (c *Client) WaitForTaskWithInterval(taskUID int64, interval time.Duration) (*meilisearch.Task, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot wait for task")
	}

	task, err := c.client.WaitForTask(taskUID, interval)
	if err != nil {
		return nil, fmt.Errorf("meilisearch wait for task error: %v", err)
	}
	return task, nil
}

// GetVersion gets Meilisearch version
func (c *Client) GetVersion() (*meilisearch.Version, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot get version")
	}

	version, err := c.client.Version()
	if err != nil {
		return nil, fmt.Errorf("meilisearch get version error: %v", err)
	}
	return version, nil
}

// GetStats gets global stats
func (c *Client) GetStats() (*meilisearch.Stats, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot get stats")
	}

	stats, err := c.client.GetStats()
	if err != nil {
		return nil, fmt.Errorf("meilisearch get stats error: %v", err)
	}
	return stats, nil
}

// GetIndexStats gets stats for a specific index
func (c *Client) GetIndexStats(indexUID string) (*meilisearch.StatsIndex, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot get index stats")
	}

	stats, err := c.client.Index(indexUID).GetStats()
	if err != nil {
		return nil, fmt.Errorf("meilisearch get index stats error: %v", err)
	}
	return stats, nil
}

// GetSettings gets index settings
func (c *Client) GetSettings(indexUID string) (*meilisearch.Settings, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot get settings")
	}

	settings, err := c.client.Index(indexUID).GetSettings()
	if err != nil {
		return nil, fmt.Errorf("meilisearch get settings error: %v", err)
	}
	return settings, nil
}

// UpdateSettings updates index settings
func (c *Client) UpdateSettings(indexUID string, settings *meilisearch.Settings) (*meilisearch.TaskInfo, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot update settings")
	}

	taskInfo, err := c.client.Index(indexUID).UpdateSettings(settings)
	if err != nil {
		return nil, fmt.Errorf("meilisearch update settings error: %v", err)
	}
	return taskInfo, nil
}

// ResetSettings resets index settings to default
func (c *Client) ResetSettings(indexUID string) (*meilisearch.TaskInfo, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot reset settings")
	}

	taskInfo, err := c.client.Index(indexUID).ResetSettings()
	if err != nil {
		return nil, fmt.Errorf("meilisearch reset settings error: %v", err)
	}
	return taskInfo, nil
}

// SearchWithContext performs search with context
func (c *Client) SearchWithContext(ctx context.Context, index, query string, options *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot perform search")
	}
	resp, err := c.client.Index(index).SearchWithContext(ctx, query, options)
	if err != nil {
		return nil, fmt.Errorf("meilisearch search error: %v", err)
	}
	return resp, nil
}

// MultiSearch performs multi-index search
func (c *Client) MultiSearch(queries *meilisearch.MultiSearchRequest) (*meilisearch.MultiSearchResponse, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot perform multi-search")
	}

	resp, err := c.client.MultiSearch(queries)
	if err != nil {
		return nil, fmt.Errorf("meilisearch multi-search error: %v", err)
	}
	return resp, nil
}

// MultiSearchWithContext performs multi-index search with context
func (c *Client) MultiSearchWithContext(ctx context.Context, queries *meilisearch.MultiSearchRequest) (*meilisearch.MultiSearchResponse, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("meilisearch client is nil, cannot perform multi-search")
	}

	resp, err := c.client.MultiSearchWithContext(ctx, queries)
	if err != nil {
		return nil, fmt.Errorf("meilisearch multi-search error: %v", err)
	}
	return resp, nil
}
