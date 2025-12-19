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
	// Prefer `data.search.opensearch.*` but keep backward compatibility with `data.opensearch.*`.
	addresses := v.GetStringSlice("data.search.opensearch.addresses")
	if len(addresses) == 0 {
		addresses = v.GetStringSlice("data.opensearch.addresses")
	}

	username := v.GetString("data.search.opensearch.username")
	if username == "" {
		username = v.GetString("data.opensearch.username")
	}

	password := v.GetString("data.search.opensearch.password")
	if password == "" {
		password = v.GetString("data.opensearch.password")
	}

	insecureSkipTLS := v.GetBool("data.search.opensearch.insecure_skip_tls")
	if !v.IsSet("data.search.opensearch.insecure_skip_tls") {
		insecureSkipTLS = v.GetBool("data.opensearch.insecure_skip_tls")
	}

	return &OpenSearch{
		Addresses:       addresses,
		Username:        username,
		Password:        password,
		InsecureSkipTLS: insecureSkipTLS,
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
	// Prefer `data.search.elasticsearch.*` but keep backward compatibility with `data.elasticsearch.*`.
	addresses := v.GetStringSlice("data.search.elasticsearch.addresses")
	if len(addresses) == 0 {
		addresses = v.GetStringSlice("data.elasticsearch.addresses")
	}

	username := v.GetString("data.search.elasticsearch.username")
	if username == "" {
		username = v.GetString("data.elasticsearch.username")
	}

	password := v.GetString("data.search.elasticsearch.password")
	if password == "" {
		password = v.GetString("data.elasticsearch.password")
	}

	return &Elasticsearch{
		Addresses: addresses,
		Username:  username,
		Password:  password,
	}
}

// Meilisearch meilisearch config struct
type Meilisearch struct {
	Host   string `json:"host" yaml:"host"`
	APIKey string `json:"api_key" yaml:"api_key"`
}

// getMeilisearchConfigs reads Meilisearch configurations
func getMeilisearchConfigs(v *viper.Viper) *Meilisearch {
	// Prefer `data.search.meilisearch.*` but keep backward compatibility with `data.meilisearch.*`.
	host := v.GetString("data.search.meilisearch.host")
	if host == "" {
		host = v.GetString("data.meilisearch.host")
	}

	apiKey := v.GetString("data.search.meilisearch.api_key")
	if apiKey == "" {
		apiKey = v.GetString("data.meilisearch.api_key")
	}

	return &Meilisearch{
		Host:   host,
		APIKey: apiKey,
	}
}
