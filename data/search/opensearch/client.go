package opensearch

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

// Client OpenSearch client
type Client struct {
	client *opensearchapi.Client
}

// NewClient creates a new OpenSearch client
func NewClient(addresses []string, username, password string, insecure bool) (*Client, error) {
	if len(addresses) == 0 {
		return &Client{client: nil}, nil
	}

	// Configure transport with TLS options
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}

	// Create the client
	client, err := opensearchapi.NewClient(
		opensearchapi.Config{
			Client: opensearch.Config{
				Addresses:  addresses,
				Username:   username,
				Password:   password,
				Transport:  transport,
				MaxRetries: 3,
			},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("opensearch client creation error: %w", err)
	}

	return &Client{client: client}, nil
}

// Search performs a search in OpenSearch
func (c *Client) Search(ctx context.Context, indexName, query string) (*opensearchapi.SearchResp, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("opensearch client is nil, cannot perform search")
	}

	// Create search request
	searchReq := opensearchapi.SearchReq{
		Indices: []string{indexName},
		Body:    strings.NewReader(query),
	}

	// Execute search
	res, err := c.client.Search(ctx, &searchReq)
	if err != nil {
		log.Printf("OpenSearch search error: %s", err)
		return nil, err
	}

	return res, nil
}

// IndexDocument indexes a document in OpenSearch
func (c *Client) IndexDocument(ctx context.Context, indexName string, documentID string, document any) error {
	if c == nil || c.client == nil {
		return errors.New("opensearch client is nil, cannot index documents")
	}

	// Marshal document to JSON
	data, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("error encoding document: %w", err)
	}

	// Create index request
	indexReq := opensearchapi.IndexReq{
		Index:      indexName,
		DocumentID: documentID,
		Body:       strings.NewReader(string(data)),
		Params:     opensearchapi.IndexParams{Refresh: "true"},
	}

	// Execute index request
	_, err = c.client.Index(ctx, indexReq)
	if err != nil {
		return fmt.Errorf("opensearch indexing error: %w", err)
	}

	return nil
}

// DeleteDocument deletes a document from OpenSearch
func (c *Client) DeleteDocument(ctx context.Context, indexName, documentID string) error {
	if c == nil || c.client == nil {
		return errors.New("opensearch client is nil, cannot delete documents")
	}

	// Create delete request
	deleteReq := opensearchapi.DocumentDeleteReq{
		Index:      indexName,
		DocumentID: documentID,
		Params:     opensearchapi.DocumentDeleteParams{Refresh: "true"},
	}

	// Execute delete request
	_, err := c.client.Document.Delete(ctx, deleteReq)
	if err != nil {
		return fmt.Errorf("opensearch deletion error: %w", err)
	}

	return nil
}

// BulkIndex indexes multiple documents in OpenSearch
func (c *Client) BulkIndex(ctx context.Context, indexName string, documents []any) error {
	if c == nil || c.client == nil {
		return errors.New("opensearch client is nil, cannot perform bulk indexing")
	}

	// Prepare bulk request body
	var bulkRequestBody strings.Builder
	for _, doc := range documents {
		// Create action line
		actionLine := fmt.Sprintf(`{"index":{}}%s`, "\n")
		bulkRequestBody.WriteString(actionLine)

		// Marshal document
		docBytes, err := json.Marshal(doc)
		if err != nil {
			return fmt.Errorf("error encoding document: %w", err)
		}

		// Add document line
		bulkRequestBody.Write(docBytes)
		bulkRequestBody.WriteString("\n")
	}

	// Create bulk request
	bulkReq := opensearchapi.BulkReq{
		Index: indexName,
		Body:  strings.NewReader(bulkRequestBody.String()),
	}

	// Execute bulk request
	_, err := c.client.Bulk(ctx, bulkReq)
	if err != nil {
		return fmt.Errorf("opensearch bulk index error: %w", err)
	}

	return nil
}

// CreateIndex creates a new index with optional mappings
func (c *Client) CreateIndex(ctx context.Context, indexName string, mappings string) error {
	if c == nil || c.client == nil {
		return errors.New("opensearch client is nil, cannot create index")
	}

	// Create index request
	createReq := opensearchapi.IndicesCreateReq{
		Index: indexName,
	}

	// Add mappings if provided
	if mappings != "" {
		createReq.Body = strings.NewReader(mappings)
	}

	// Execute index creation
	_, err := c.client.Indices.Create(ctx, createReq)
	if err != nil {
		// Check if error is "index already exists"
		var opensearchError *opensearch.StructError
		if errors.As(err, &opensearchError) {
			if opensearchError.Err.Type == "resource_already_exists_exception" {
				return nil // Ignore if index already exists
			}
		}
		return fmt.Errorf("opensearch create index error: %w", err)
	}

	return nil
}

// IndexExists checks if an index exists
func (c *Client) IndexExists(ctx context.Context, indexName string) (bool, error) {
	if c == nil || c.client == nil {
		return false, errors.New("opensearch client is nil, cannot check index")
	}

	// Create index exists request
	existsReq := opensearchapi.IndicesExistsReq{
		Indices: []string{indexName},
	}

	// Execute request
	res, err := c.client.Indices.Exists(ctx, existsReq)
	if err != nil {
		return false, fmt.Errorf("opensearch index exists error: %w", err)
	}

	// Check status code
	return res.StatusCode == 200, nil
}

// DeleteIndex deletes an index
func (c *Client) DeleteIndex(ctx context.Context, indexName string) error {
	if c == nil || c.client == nil {
		return errors.New("opensearch client is nil, cannot delete index")
	}

	// Create delete index request
	deleteReq := opensearchapi.IndicesDeleteReq{
		Indices: []string{indexName},
	}

	// Execute request
	_, err := c.client.Indices.Delete(ctx, deleteReq)
	if err != nil {
		// Check if error is "index not found"
		var opensearchError *opensearch.StructError
		if errors.As(err, &opensearchError) {
			if opensearchError.Err.Type == "index_not_found_exception" {
				return nil // Ignore if index doesn't exist
			}
		}
		return fmt.Errorf("opensearch delete index error: %w", err)
	}

	return nil
}

// GetClient returns the OpenSearch client
func (c *Client) GetClient() *opensearchapi.Client {
	return c.client
}

// Health checks cluster health
func (c *Client) Health(ctx context.Context) (string, error) {
	if c == nil || c.client == nil {
		return "", errors.New("opensearch client is nil, cannot check health")
	}

	// Create cluster health request
	healthReq := opensearchapi.ClusterHealthReq{}

	// Execute request
	res, err := c.client.Cluster.Health(ctx, &healthReq)
	if err != nil {
		return "", fmt.Errorf("opensearch health check error: %w", err)
	}

	return res.Status, nil
}
