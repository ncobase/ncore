package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ncobase/ncore/data/elasticsearch/client"
	"github.com/ncobase/ncore/data/search"
)

func init() {
	search.RegisterAdapterFactory(search.Elasticsearch, func(conn any) (search.Adapter, error) {
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
	return search.Elasticsearch
}

func (a *Adapter) Search(ctx context.Context, req *search.Request) (*search.Response, error) {
	if a.client == nil {
		return nil, errors.New("elasticsearch client not available")
	}

	query := a.buildQuery(req)
	resp, err := a.client.Search(ctx, req.Index, query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("elasticsearch returned status: %d", resp.StatusCode)
	}

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

	hits := make([]search.Hit, len(esResp.Hits.Hits))
	for i, hit := range esResp.Hits.Hits {
		hits[i] = search.Hit{
			ID:     hit.ID,
			Score:  hit.Score,
			Source: hit.Source,
		}
	}

	return &search.Response{
		Total: esResp.Hits.Total.Value,
		Hits:  hits,
	}, nil
}

func (a *Adapter) Index(ctx context.Context, req *search.IndexRequest) error {
	if a.client == nil {
		return errors.New("elasticsearch client not available")
	}
	return a.client.IndexDocument(ctx, req.Index, req.DocumentID, req.Document)
}

func (a *Adapter) Delete(ctx context.Context, index, id string) error {
	if a.client == nil {
		return errors.New("elasticsearch client not available")
	}
	return a.client.DeleteDocument(ctx, index, id)
}

func (a *Adapter) BulkIndex(ctx context.Context, index string, documents []any) error {
	if a.client == nil {
		return errors.New("elasticsearch client not available")
	}

	client := a.client.GetClient()
	if client == nil {
		return errors.New("elasticsearch raw client is nil")
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

func (a *Adapter) BulkDelete(ctx context.Context, index string, documentIDs []string) error {
	if a.client == nil {
		return errors.New("elasticsearch client not available")
	}

	client := a.client.GetClient()
	if client == nil {
		return errors.New("elasticsearch raw client is nil")
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

func (a *Adapter) IndexExists(ctx context.Context, indexName string) (bool, error) {
	if a.client == nil {
		return false, errors.New("elasticsearch client not available")
	}

	client := a.client.GetClient()
	if client == nil {
		return false, errors.New("elasticsearch raw client is nil")
	}

	res, err := client.Indices.Exists([]string{indexName})
	if err != nil {
		return false, fmt.Errorf("failed to check elasticsearch index existence: %w", err)
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

func (a *Adapter) CreateIndex(ctx context.Context, indexName string, settings *search.IndexSettings) error {
	if a.client == nil {
		return errors.New("elasticsearch client not available")
	}

	client := a.client.GetClient()
	if client == nil {
		return errors.New("elasticsearch raw client is nil")
	}

	settingsBody := a.buildSettings(settings)
	createRes, err := client.Indices.Create(indexName, client.Indices.Create.WithBody(strings.NewReader(settingsBody)))
	if err != nil {
		return fmt.Errorf("failed to create elasticsearch index: %w", err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		return fmt.Errorf("elasticsearch index creation failed: %s", createRes.Status())
	}

	return nil
}

func (a *Adapter) Health(ctx context.Context) error {
	if a.client == nil {
		return errors.New("elasticsearch client not available")
	}

	client := a.client.GetClient()
	if client == nil {
		return errors.New("elasticsearch raw client is nil")
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

func (a *Adapter) buildQuery(req *search.Request) string {
	// For now use defaults matching original code:
	searchableFields := []string{"title^2", "content", "details", "name", "description"}

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
		}`, req.Query, a.buildFieldsArray(searchableFields), req.From, req.Size)
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
	}`, req.Query, a.buildFieldsArray(searchableFields), strings.Join(filterQueries, ","), req.From, req.Size)
}

func (a *Adapter) buildFieldsArray(fields []string) string {
	quotedFields := make([]string, len(fields))
	for i, field := range fields {
		quotedFields[i] = fmt.Sprintf(`"%s"`, field)
	}
	return fmt.Sprintf("[%s]", strings.Join(quotedFields, ","))
}

func (a *Adapter) buildSettings(settings *search.IndexSettings) string {
	shards := 1
	replicas := 0
	refreshInterval := "1s"
	searchableFields := []string{"title^2", "content", "details", "name", "description"}
	filterableFields := []string{"id", "user_id", "type", "status", "created_at", "updated_at"}

	if settings != nil {
		if settings.Shards > 0 {
			shards = settings.Shards
		}
		if settings.Replicas >= 0 {
			replicas = settings.Replicas
		}
		if settings.RefreshInterval != "" {
			refreshInterval = settings.RefreshInterval
		}
		if len(settings.SearchableFields) > 0 {
			searchableFields = settings.SearchableFields
		}
		if len(settings.FilterableFields) > 0 {
			filterableFields = settings.FilterableFields
		}
	}

	properties := make(map[string]string)

	for _, field := range searchableFields {
		fieldName := strings.TrimSuffix(field, "^2")
		properties[fieldName] = `{"type": "text"}`
	}

	for _, field := range filterableFields {
		if field == "created_at" || field == "updated_at" {
			properties[field] = `{"type": "long"}`
		} else {
			properties[field] = `{"type": "keyword"}`
		}
	}

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
