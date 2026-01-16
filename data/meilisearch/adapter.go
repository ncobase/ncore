package meilisearch

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ncobase/ncore/data/meilisearch/client"
	"github.com/ncobase/ncore/data/search"
)

func init() {
	search.RegisterAdapterFactory(search.Meilisearch, func(conn any) (search.Adapter, error) {
		c, ok := conn.(*client.Client)
		if !ok {
			return nil, fmt.Errorf("expected *client.Client, got %T", conn)
		}
		return NewAdapter(c), nil
	})
}

type Adapter struct {
	client *client.Client
}

func NewAdapter(c *client.Client) *Adapter {
	return &Adapter{client: c}
}

func (a *Adapter) Type() search.Engine {
	return search.Meilisearch
}

func (a *Adapter) Search(ctx context.Context, req *search.Request) (*search.Response, error) {
	if a.client == nil {
		return nil, errors.New("meilisearch client not available")
	}

	searchReq := &client.SearchParams{
		Offset: int64(req.From),
		Limit:  int64(req.Size),
	}

	if len(req.Filter) > 0 {
		filters := make([]string, 0, len(req.Filter))
		for field, value := range req.Filter {
			filters = append(filters, fmt.Sprintf("%s = '%v'", field, value))
		}
		filterStr := strings.Join(filters, " AND ")
		searchReq.Filter = filterStr
	}

	msResp, err := a.client.Search(req.Index, req.Query, searchReq)
	if err != nil {
		return nil, err
	}

	hits := make([]search.Hit, len(msResp.Hits))
	for i, hit := range msResp.Hits {
		hitMap := make(map[string]any)
		for k, v := range hit {
			hitMap[k] = v
		}

		var id string
		if idVal, exists := hitMap["id"]; exists {
			id = fmt.Sprintf("%v", idVal)
		}
		hits[i] = search.Hit{
			ID:     id,
			Score:  1.0,
			Source: hitMap,
		}
	}

	return &search.Response{
		Total: int64(msResp.EstimatedTotalHits),
		Hits:  hits,
	}, nil
}

func (a *Adapter) Index(ctx context.Context, req *search.IndexRequest) error {
	if a.client == nil {
		return errors.New("meilisearch client not available")
	}
	documents := []any{req.Document}
	if docMap, ok := req.Document.(map[string]any); ok && req.DocumentID != "" {
		docMap["id"] = req.DocumentID
	}
	return a.client.IndexDocuments(req.Index, documents)
}

func (a *Adapter) Delete(ctx context.Context, index, id string) error {
	if a.client == nil {
		return errors.New("meilisearch client not available")
	}
	return a.client.DeleteDocuments(index, id)
}

func (a *Adapter) BulkIndex(ctx context.Context, index string, documents []any) error {
	if a.client == nil {
		return errors.New("meilisearch client not available")
	}
	return a.client.IndexDocuments(index, documents)
}

func (a *Adapter) BulkDelete(ctx context.Context, index string, documentIDs []string) error {
	if a.client == nil {
		return errors.New("meilisearch client not available")
	}
	return a.client.DeleteDocuments(index, documentIDs...)
}

func (a *Adapter) IndexExists(ctx context.Context, indexName string) (bool, error) {
	if a.client == nil {
		return false, errors.New("meilisearch client not available")
	}
	_, err := a.client.GetIndexStats(indexName)
	if err != nil {
		if strings.Contains(err.Error(), "index_not_found") ||
			strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check meilisearch index existence: %w", err)
	}
	return true, nil
}

func (a *Adapter) CreateIndex(ctx context.Context, indexName string, settings *search.IndexSettings) error {
	if a.client == nil {
		return errors.New("meilisearch client not available")
	}
	dummyDoc := map[string]any{
		"id":    "init_doc_" + indexName,
		"_init": true,
		"type":  "initialization",
	}

	err := a.client.IndexDocuments(indexName, []any{dummyDoc}, "id")
	if err != nil {
		return fmt.Errorf("failed to create meilisearch index: %w", err)
	}

	time.Sleep(100 * time.Millisecond)
	_ = a.client.DeleteDocuments(indexName, "init_doc_"+indexName)

	nativeClient := a.client.GetClient()
	index := nativeClient.Index(indexName)

	searchableFields := []string{"title^2", "content", "details", "name", "description"}
	filterableFields := []string{"id", "user_id", "type", "status", "created_at", "updated_at"}

	if settings != nil {
		if len(settings.SearchableFields) > 0 {
			searchableFields = settings.SearchableFields
		}
		if len(settings.FilterableFields) > 0 {
			filterableFields = settings.FilterableFields
		}
	}

	_, err = index.UpdateSearchableAttributes(&searchableFields)
	if err != nil {
		fmt.Printf("Warning: failed to set searchable attributes for meilisearch index %s: %v\n", indexName, err)
	}

	filterableFieldsAny := make([]any, len(filterableFields))
	for i, f := range filterableFields {
		filterableFieldsAny[i] = f
	}
	_, err = index.UpdateFilterableAttributes(&filterableFieldsAny)
	if err != nil {
		fmt.Printf("Warning: failed to set filterable attributes for meilisearch index %s: %v\n", indexName, err)
	}

	return nil
}

func (a *Adapter) Health(ctx context.Context) error {
	if a.client == nil {
		return errors.New("meilisearch client not available")
	}
	_, err := a.client.Health()
	return err
}
