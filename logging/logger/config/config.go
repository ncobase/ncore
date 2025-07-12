package config

import (
	"fmt"
	"strings"
	"time"

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
	DateSuffix      string           `json:"date_suffix" yaml:"date_suffix"`
	RotateDaily     bool             `json:"rotate_daily" yaml:"rotate_daily"`
	Desensitization *Desensitization `json:"desensitization" yaml:"desensitization"`
	Meilisearch     *Meilisearch     `json:"meilisearch" yaml:"meilisearch"`
	Elasticsearch   *Elasticsearch   `json:"elasticsearch" yaml:"elasticsearch"`
	OpenSearch      *OpenSearch      `json:"opensearch" yaml:"opensearch"`
}

// GetConfig returns the logger configuration with date suffix support
func GetConfig(v *viper.Viper) *Config {
	if !v.IsSet("logger") {
		return nil
	}

	appName := v.GetString("app_name")
	environment := v.GetString("environment")

	if appName == "" {
		appName = "app"
	}
	if environment == "" {
		environment = "default"
	}

	baseIndexName := strings.ToLower(fmt.Sprintf("%s-%s-logs", appName, environment))

	if v.IsSet("logger.index_name") && v.GetString("logger.index_name") != "" {
		baseIndexName = v.GetString("logger.index_name")
	}

	return &Config{
		Level:           v.GetInt("logger.level"),
		Format:          v.GetString("logger.format"),
		Path:            v.GetString("logger.path"),
		Output:          v.GetString("logger.output"),
		OutputFile:      v.GetString("logger.output_file"),
		IndexName:       baseIndexName,
		DateSuffix:      getDateSuffixPattern(v),
		RotateDaily:     getRotateDaily(v),
		Desensitization: getDesensitizationConfigs(v),
		Meilisearch:     getMeilisearchConfigs(v),
		Elasticsearch:   getElasticsearchConfigs(v),
		OpenSearch:      getOpenSearchConfigs(v),
	}
}

// getDateSuffixPattern gets date suffix pattern
func getDateSuffixPattern(v *viper.Viper) string {
	if v.IsSet("logger.date_suffix") {
		return v.GetString("logger.date_suffix")
	}
	return "2006-01-02" // Default: YYYY-MM-DD
}

// getRotateDaily gets daily rotation setting
func getRotateDaily(v *viper.Viper) bool {
	if v.IsSet("logger.rotate_daily") {
		return v.GetBool("logger.rotate_daily")
	}
	return true // Default: enable daily rotation
}

// BuildIndexName builds full index name with date suffix
func (c *Config) BuildIndexName(t time.Time) string {
	if !c.RotateDaily {
		return c.IndexName
	}

	dateSuffix := t.Format(c.DateSuffix)
	return fmt.Sprintf("%s-%s", c.IndexName, dateSuffix)
}

// GetCurrentIndexName gets current index name with today's date
func (c *Config) GetCurrentIndexName() string {
	return c.BuildIndexName(time.Now())
}
