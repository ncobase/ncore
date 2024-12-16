package meili

import (
	"errors"
	"log"

	"github.com/meilisearch/meilisearch-go"
)

// Client Meilisearch client
type Client struct {
	client meilisearch.ServiceManager
}

// NewMeilisearch new Meilisearch client
func NewMeilisearch(host, apiKey string) *Client {
	if host == "" {
		return &Client{client: nil}
	}
	ms := meilisearch.New(host, meilisearch.WithAPIKey(apiKey))
	return &Client{client: ms}
}

// Search search from Meilisearch
func (c *Client) Search(index, query string, options *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error) {
	if c == nil || c.client == nil {
		log.Printf("Meilisearch client is nil, cannot perform search")
		return nil, errors.New("meilisearch client is nil")
	}
	resp, err := c.client.Index(index).Search(query, options)
	if err != nil {
		log.Printf("Meilisearch search error: %v", err)
		return nil, err
	}
	return resp, nil
}

// IndexDocuments index document to Meilisearch
func (c *Client) IndexDocuments(index string, document any, primaryKey ...string) error {
	if c == nil || c.client == nil {
		log.Printf("Meilisearch client is nil, cannot index documents")
		return errors.New("meilisearch client is nil")
	}
	_, err := c.client.Index(index).AddDocuments(document, primaryKey...)
	if err != nil {
		log.Printf("Meilisearch index document error: %v", err)
		return err
	}
	return nil
}

// UpdateDocuments update document to Meilisearch
func (c *Client) UpdateDocuments(index string, document any, documentID string) error {
	if c == nil || c.client == nil {
		log.Printf("Meilisearch client is nil, cannot update documents")
		return errors.New("meilisearch client is nil")
	}
	_, err := c.client.Index(index).UpdateDocuments(document, documentID)
	if err != nil {
		log.Printf("Meilisearch update document error: %v", err)
		return err
	}
	return nil
}

// DeleteDocuments delete document from Meilisearch
func (c *Client) DeleteDocuments(index, documentID string) error {
	if c == nil || c.client == nil {
		log.Printf("Meilisearch client is nil, cannot delete documents")
		return errors.New("meilisearch client is nil")
	}
	_, err := c.client.Index(index).DeleteDocument(documentID)
	if err != nil {
		log.Printf("Meilisearch delete document error: %v", err)
		return err
	}
	return nil
}

// GetClient get Meilisearch client
func (c *Client) GetClient() meilisearch.ServiceManager {
	return c.client
}
