package config

import (
	nconfig "github.com/ncobase/ncore/config"
	"github.com/spf13/viper"
)

// Config contains logging infrastructure settings.
type Config = nconfig.Logger

// Desensitization contains log masking settings.
type Desensitization = nconfig.LogDesensitization

// Meilisearch contains Meilisearch log hook settings.
type Meilisearch = nconfig.LoggerMeilisearch

// Elasticsearch contains Elasticsearch log hook settings.
type Elasticsearch = nconfig.LoggerElasticsearch

// OpenSearch contains OpenSearch log hook settings.
type OpenSearch = nconfig.LoggerOpenSearch

// GetConfig returns logging configuration from viper.
func GetConfig(v *viper.Viper) *Config {
	return nconfig.GetLoggerConfig(v)
}
