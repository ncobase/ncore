package elastic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// Client Elasticsearch client
type Client struct {
	client *elasticsearch.Client
}

// NewClient new Elasticsearch client
func NewClient(addresses []string, username, password string) (*Client, error) {
	if len(addresses) == 0 {
		return &Client{client: nil}, nil
	}

	cfg := elasticsearch.Config{
		Addresses: addresses,
		Username:  username,
		Password:  password,
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Printf("Elasticsearch client creation error: %s", err)
		return nil, err
	}

	return &Client{client: es}, nil
}

// Search search from Elasticsearch
func (c *Client) Search(ctx context.Context, indexName, query string) (*esapi.Response, error) {
	if c == nil || c.client == nil {
		log.Printf("Elasticsearch client is nil, cannot perform search")
		return nil, errors.New("elasticsearch client is nil")
	}

	res, err := c.client.Search(
		c.client.Search.WithContext(ctx),
		c.client.Search.WithIndex(indexName),
		c.client.Search.WithBody(strings.NewReader(query)),
		c.client.Search.WithTrackTotalHits(true),
		c.client.Search.WithPretty(),
	)
	if err != nil {
		log.Printf("Elasticsearch search error: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	var sr esapi.Response
	if err := json.NewDecoder(res.Body).Decode(&sr); err != nil {
		log.Printf("Error parsing the response body: %s", err)
		return nil, err
	}

	return &sr, nil
}

// IndexDocument index document to Elasticsearch
func (c *Client) IndexDocument(ctx context.Context, indexName string, documentID string, document any) error {
	if c == nil || c.client == nil {
		log.Printf("Elasticsearch client is nil, cannot index documents")
		return errors.New("elasticsearch client is nil")
	}

	var b strings.Builder
	enc := json.NewEncoder(&b)
	if err := enc.Encode(document); err != nil {
		log.Printf("Error encoding document: %s", err)
		return err
	}

	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: documentID,
		Body:       strings.NewReader(b.String()),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		log.Printf("Error indexing document: %s", err)
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		var respBody map[string]any
		if err := json.NewDecoder(res.Body).Decode(&respBody); err != nil {
			log.Printf("Error parsing the response body: %s", err)
		} else {
			log.Printf("Elasticsearch indexing error: %s: %s", res.Status(), respBody["error"])
		}
		return fmt.Errorf("elasticsearch indexing error: %s", res.Status())
	}

	return nil
}

// DeleteDocument delete document from Elasticsearch
func (c *Client) DeleteDocument(ctx context.Context, indexName, documentID string) error {
	if c == nil || c.client == nil {
		log.Printf("Elasticsearch client is nil, cannot delete documents")
		return errors.New("elasticsearch client is nil")
	}

	req := esapi.DeleteRequest{
		Index:      indexName,
		DocumentID: documentID,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		log.Printf("Error deleting document: %s", err)
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		var respBody map[string]any
		if err := json.NewDecoder(res.Body).Decode(&respBody); err != nil {
			log.Printf("Error parsing the response body: %s", err)
		} else {
			log.Printf("Elasticsearch deletion error: %s: %s", res.Status(), respBody["error"])
		}
		return fmt.Errorf("elasticsearch deletion error: %s", res.Status())
	}

	return nil
}

// GetClient get Elasticsearch client
func (c *Client) GetClient() *elasticsearch.Client {
	return c.client
}
