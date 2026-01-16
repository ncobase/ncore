package opensearch

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ncobase/ncore/data/opensearch/client"
	"github.com/ncobase/ncore/data/search"
	"github.com/ncobase/ncore/utils/convert"
)

func init() {
	search.RegisterAdapterFactory(search.OpenSearch, func(conn any) (search.Adapter, error) {
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
	return search.OpenSearch
}

func (a *Adapter) Search(ctx context.Context, req *search.Request) (*search.Response, error) {
	if a.client == nil {
		return nil, errors.New("opensearch client not available")
	}

	query := a.buildQuery(req)
	osResp, err := a.client.Search(ctx, req.Index, query)
	if err != nil {
		return nil, err
	}

	if osResp.Errors {
		return nil, errors.New("opensearch returned errors")
	}

	hits := make([]search.Hit, len(osResp.Hits.Hits))
	for i, hit := range osResp.Hits.Hits {
		source, _ := convert.ToJSONMap(hit.Source)
		hits[i] = search.Hit{
			ID:     hit.ID,
			Score:  float64(hit.Score),
			Source: source,
		}
	}

	return &search.Response{
		Total: int64(osResp.Hits.Total.Value),
		Hits:  hits,
	}, nil
}

func (a *Adapter) Index(ctx context.Context, req *search.IndexRequest) error {
	if a.client == nil {
		return errors.New("opensearch client not available")
	}
	return a.client.IndexDocument(ctx, req.Index, req.DocumentID, req.Document)
}

func (a *Adapter) Delete(ctx context.Context, index, id string) error {
	if a.client == nil {
		return errors.New("opensearch client not available")
	}
	return a.client.DeleteDocument(ctx, index, id)
}

func (a *Adapter) BulkIndex(ctx context.Context, index string, documents []any) error {
	if a.client == nil {
		return errors.New("opensearch client not available")
	}
	return a.client.BulkIndex(ctx, index, documents)
}

func (a *Adapter) BulkDelete(ctx context.Context, index string, documentIDs []string) error {
	if a.client == nil {
		return errors.New("opensearch client not available")
	}
	// Opensearch wrapper doesn't have BulkDelete?
	// search.go implemented it by looping DeleteDocument
	for _, docID := range documentIDs {
		if delErr := a.client.DeleteDocument(ctx, index, docID); delErr != nil {
			return delErr
		}
	}
	return nil
}

func (a *Adapter) IndexExists(ctx context.Context, indexName string) (bool, error) {
	if a.client == nil {
		return false, errors.New("opensearch client not available")
	}
	return a.client.IndexExists(ctx, indexName)
}

func (a *Adapter) CreateIndex(ctx context.Context, indexName string, settings *search.IndexSettings) error {
	if a.client == nil {
		return errors.New("opensearch client not available")
	}
	settingsBody := a.buildSettings(settings)
	return a.client.CreateIndex(ctx, indexName, settingsBody)
}

func (a *Adapter) Health(ctx context.Context) error {
	if a.client == nil {
		return errors.New("opensearch client not available")
	}
	_, err := a.client.Health(ctx)
	return err
}

func (a *Adapter) buildQuery(req *search.Request) string {
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
