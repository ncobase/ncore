package search

import "time"

// Engine represents search engine type
type Engine string

const (
	Elasticsearch Engine = "elasticsearch"
	OpenSearch    Engine = "opensearch"
	Meilisearch   Engine = "meilisearch"
)

// Request represents unified search request
type Request struct {
	Index  string         `json:"index"`
	Query  string         `json:"query"`
	Filter map[string]any `json:"filter,omitempty"`
	From   int            `json:"from,omitempty"`
	Size   int            `json:"size,omitempty"`
}

// Response represents unified search response
type Response struct {
	Total    int64         `json:"total"`
	Hits     []Hit         `json:"hits"`
	Duration time.Duration `json:"duration"`
	Engine   Engine        `json:"engine"`
}

// Hit represents search result item
type Hit struct {
	ID     string         `json:"id"`
	Score  float64        `json:"score"`
	Source map[string]any `json:"source"`
}

// IndexRequest represents document indexing request
type IndexRequest struct {
	Index      string `json:"index"`
	DocumentID string `json:"document_id,omitempty"`
	Document   any    `json:"document"`
}
