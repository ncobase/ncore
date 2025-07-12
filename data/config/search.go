package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Search represents search engine configuration
type Search struct {
	IndexPrefix     string         `yaml:"index_prefix" json:"index_prefix"`
	DefaultEngine   string         `yaml:"default_engine" json:"default_engine"`
	AutoCreateIndex bool           `yaml:"auto_create_index" json:"auto_create_index"`
	IndexSettings   *IndexSettings `yaml:"index_settings" json:"index_settings"`
	Meilisearch     *Meilisearch   `yaml:"meilisearch" json:"meilisearch"`
	Elasticsearch   *Elasticsearch `yaml:"elasticsearch" json:"elasticsearch"`
	OpenSearch      *OpenSearch    `yaml:"opensearch" json:"opensearch"`
}

// IndexSettings represents default index configuration
type IndexSettings struct {
	Shards           int      `yaml:"shards" json:"shards"`
	Replicas         int      `yaml:"replicas" json:"replicas"`
	RefreshInterval  string   `yaml:"refresh_interval" json:"refresh_interval"`
	SearchableFields []string `yaml:"searchable_fields" json:"searchable_fields"`
	FilterableFields []string `yaml:"filterable_fields" json:"filterable_fields"`
}

// getSearchConfig reads search configurations
func getSearchConfig(v *viper.Viper) *Search {
	if !v.IsSet("data.search") {
		return &Search{
			IndexPrefix:     getDefaultIndexPrefix(v),
			DefaultEngine:   "elasticsearch",
			AutoCreateIndex: true,
			IndexSettings:   getDefaultIndexSettings(),
			Meilisearch:     getMeilisearchConfigs(v),
			Elasticsearch:   getElasticsearchConfigs(v),
			OpenSearch:      getOpenSearchConfigs(v),
		}
	}

	return &Search{
		IndexPrefix:     getSearchIndexPrefix(v),
		DefaultEngine:   getSearchDefaultEngine(v),
		AutoCreateIndex: getSearchAutoCreateIndex(v),
		IndexSettings:   getSearchIndexSettings(v),
		Meilisearch:     getMeilisearchConfigs(v),
		Elasticsearch:   getElasticsearchConfigs(v),
		OpenSearch:      getOpenSearchConfigs(v),
	}
}

// getSearchIndexPrefix gets search index prefix
func getSearchIndexPrefix(v *viper.Viper) string {
	if v.IsSet("data.search.index_prefix") {
		return v.GetString("data.search.index_prefix")
	}
	return getDefaultIndexPrefix(v)
}

// getDefaultIndexPrefix builds default index prefix from app info
func getDefaultIndexPrefix(v *viper.Viper) string {
	appName := v.GetString("app_name")
	environment := v.GetString("environment")

	if appName != "" && environment != "" {
		return strings.ToLower(fmt.Sprintf("%s-%s", appName, environment))
	}

	if appName != "" {
		return strings.ToLower(appName)
	}

	return ""
}

// getSearchDefaultEngine gets default search engine
func getSearchDefaultEngine(v *viper.Viper) string {
	if v.IsSet("data.search.default_engine") {
		return v.GetString("data.search.default_engine")
	}
	return "elasticsearch"
}

// getSearchAutoCreateIndex gets auto create index setting
func getSearchAutoCreateIndex(v *viper.Viper) bool {
	if v.IsSet("data.search.auto_create_index") {
		return v.GetBool("data.search.auto_create_index")
	}
	return true
}

// getSearchIndexSettings gets search index settings
func getSearchIndexSettings(v *viper.Viper) *IndexSettings {
	if !v.IsSet("data.search.index_settings") {
		return getDefaultIndexSettings()
	}

	searchableFields := v.GetStringSlice("data.search.index_settings.searchable_fields")
	if len(searchableFields) == 0 {
		searchableFields = []string{"title", "content", "details", "name", "description"}
	}

	filterableFields := v.GetStringSlice("data.search.index_settings.filterable_fields")
	if len(filterableFields) == 0 {
		filterableFields = []string{"id", "user_id", "type", "status", "created_at", "updated_at"}
	}

	return &IndexSettings{
		Shards:           getIntOrDefault(v, "data.search.index_settings.shards", 1),
		Replicas:         getIntOrDefault(v, "data.search.index_settings.replicas", 0),
		RefreshInterval:  getStringOrDefault(v, "data.search.index_settings.refresh_interval", "1s"),
		SearchableFields: searchableFields,
		FilterableFields: filterableFields,
	}
}

// getDefaultIndexSettings returns default index settings
func getDefaultIndexSettings() *IndexSettings {
	return &IndexSettings{
		Shards:          1,
		Replicas:        0,
		RefreshInterval: "1s",
		SearchableFields: []string{
			"title", "content", "details", "name", "description",
		},
		FilterableFields: []string{
			"id", "user_id", "type", "status", "created_at", "updated_at",
		},
	}
}

// OpenSearch opensearch config struct
type OpenSearch struct {
	Addresses       []string `json:"addresses" yaml:"addresses"`
	Username        string   `json:"username" yaml:"username"`
	Password        string   `json:"password" yaml:"password"`
	InsecureSkipTLS bool     `json:"insecure_skip_tls" yaml:"insecure_skip_tls"`
}

// getOpenSearchConfigs reads OpenSearch configurations
func getOpenSearchConfigs(v *viper.Viper) *OpenSearch {
	return &OpenSearch{
		Addresses:       v.GetStringSlice("data.opensearch.addresses"),
		Username:        v.GetString("data.opensearch.username"),
		Password:        v.GetString("data.opensearch.password"),
		InsecureSkipTLS: v.GetBool("data.opensearch.insecure_skip_tls"),
	}
}

// Elasticsearch elasticsearch config struct
type Elasticsearch struct {
	Addresses []string `json:"addresses" yaml:"addresses"`
	Username  string   `json:"username" yaml:"username"`
	Password  string   `json:"password" yaml:"password"`
}

// getElasticsearchConfigs reads Elasticsearch configurations
func getElasticsearchConfigs(v *viper.Viper) *Elasticsearch {
	return &Elasticsearch{
		Addresses: v.GetStringSlice("data.elasticsearch.addresses"),
		Username:  v.GetString("data.elasticsearch.username"),
		Password:  v.GetString("data.elasticsearch.password"),
	}
}

// Meilisearch meilisearch config struct
type Meilisearch struct {
	Host   string `json:"host" yaml:"host"`
	APIKey string `json:"api_key" yaml:"api_key"`
}

// getMeilisearchConfigs reads Meilisearch configurations
func getMeilisearchConfigs(v *viper.Viper) *Meilisearch {
	return &Meilisearch{
		Host:   v.GetString("data.meilisearch.host"),
		APIKey: v.GetString("data.meilisearch.api_key"),
	}
}
