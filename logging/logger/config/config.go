package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config configuration struct
type Config struct {
	Level           int              `json:"level" yaml:"level"`
	Path            string           `json:"path" yaml:"path"`
	Format          string           `json:"format" yaml:"format"`
	Output          string           `json:"output" yaml:"output"`
	OutputFile      string           `json:"output_file" yaml:"output_file"`
	IndexName       string           `json:"index_name" yaml:"index_name"`
	Desensitization *Desensitization `json:"desensitization" yaml:"desensitization"`
	Meilisearch     *Meilisearch     `json:"meilisearch" yaml:"meilisearch"`
	Elasticsearch   *Elasticsearch   `json:"elasticsearch" yaml:"elasticsearch"`
	OpenSearch      *OpenSearch      `json:"opensearch" yaml:"opensearch"`
}

// GetConfig returns the logger configuration
func GetConfig(v *viper.Viper) *Config {
	if !v.IsSet("logger") {
		return nil
	}

	indexName := strings.ToLower(v.GetString("app_name") + "-" + v.GetString("run_mode") + "-log")
	if v.IsSet("logger.index_name") && v.GetString("logger.index_name") != "" {
		indexName = v.GetString("logger.index_name")
	}

	return &Config{
		Level:           v.GetInt("logger.level"),
		Format:          v.GetString("logger.format"),
		Path:            v.GetString("logger.path"),
		Output:          v.GetString("logger.output"),
		OutputFile:      v.GetString("logger.output_file"),
		IndexName:       indexName,
		Desensitization: getDesensitizationConfigs(v),
		Meilisearch:     getMeilisearchConfigs(v),
		Elasticsearch:   getElasticsearchConfigs(v),
		OpenSearch:      getOpenSearchConfigs(v),
	}

}
