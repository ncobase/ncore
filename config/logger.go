package config

import (
	"strings"

	dc "github.com/ncobase/ncore/data/config"
	"github.com/spf13/viper"
)

// Logger logger config struct
type Logger struct {
	Level         int
	Path          string
	Format        string
	Output        string
	OutputFile    string
	IndexName     string
	Meilisearch   *dc.Meilisearch
	Elasticsearch *dc.Elasticsearch
	OpenSearch    *dc.OpenSearch
}

func getLoggerConfig(v *viper.Viper) *Logger {
	indexName := strings.ToLower(v.GetString("app_name") + "-" + v.GetString("run_mode") + "-log")
	if v.IsSet("logger.index_name") && v.GetString("logger.index_name") != "" {
		indexName = v.GetString("logger.index_name")
	}
	return &Logger{
		Level:      v.GetInt("logger.level"),
		Format:     v.GetString("logger.format"),
		Path:       v.GetString("logger.path"),
		Output:     v.GetString("logger.output"),
		OutputFile: v.GetString("logger.output_file"),
		IndexName:  indexName,
		Meilisearch: &dc.Meilisearch{
			Host:   v.GetString("logger.meilisearch.host"),
			APIKey: v.GetString("logger.meilisearch.api_key"),
		},
		Elasticsearch: &dc.Elasticsearch{
			Addresses: v.GetStringSlice("logger.elasticsearch.addresses"),
			Username:  v.GetString("logger.elasticsearch.username"),
			Password:  v.GetString("logger.elasticsearch.password"),
		},
		OpenSearch: &dc.OpenSearch{
			Addresses:       v.GetStringSlice("logger.opensearch.addresses"),
			Username:        v.GetString("logger.opensearch.username"),
			Password:        v.GetString("logger.opensearch.password"),
			InsecureSkipTLS: v.GetBool("logger.opensearch.insecure_skip_tls"),
		},
	}
}
