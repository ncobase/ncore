package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Logger contains logging infrastructure settings.
type Logger struct {
	Level           int                  `json:"level" yaml:"level"`
	Path            string               `json:"path" yaml:"path"`
	Format          string               `json:"format" yaml:"format"`
	Output          string               `json:"output" yaml:"output"`
	OutputFile      string               `json:"output_file" yaml:"output_file"`
	IndexName       string               `json:"index_name" yaml:"index_name"`
	DateSuffix      string               `json:"date_suffix" yaml:"date_suffix"`
	RotateDaily     bool                 `json:"rotate_daily" yaml:"rotate_daily"`
	Desensitization *LogDesensitization  `json:"desensitization" yaml:"desensitization"`
	Meilisearch     *LoggerMeilisearch   `json:"meilisearch" yaml:"meilisearch"`
	Elasticsearch   *LoggerElasticsearch `json:"elasticsearch" yaml:"elasticsearch"`
	OpenSearch      *LoggerOpenSearch    `json:"opensearch" yaml:"opensearch"`
}

// LogDesensitization contains log masking settings.
type LogDesensitization struct {
	Enabled               bool     `json:"enabled" yaml:"enabled"`
	SensitiveFields       []string `json:"sensitive_fields" yaml:"sensitive_fields"`
	CustomPatterns        []string `json:"custom_patterns" yaml:"custom_patterns"`
	PreservePrefix        int      `json:"preserve_prefix" yaml:"preserve_prefix"`
	PreserveSuffix        int      `json:"preserve_suffix" yaml:"preserve_suffix"`
	MaskChar              string   `json:"mask_char" yaml:"mask_char"`
	UseFixedLength        bool     `json:"use_fixed_length" yaml:"use_fixed_length"`
	FixedMaskLength       int      `json:"fixed_mask_length" yaml:"fixed_mask_length"`
	ExactFieldMatch       bool     `json:"exact_field_match" yaml:"exact_field_match"`
	EnableDefaultPatterns bool     `json:"enable_default_patterns" yaml:"enable_default_patterns"`
}

// LoggerMeilisearch contains Meilisearch log hook settings.
type LoggerMeilisearch struct {
	Host   string `json:"host" yaml:"host"`
	APIKey string `json:"api_key" yaml:"api_key"`
}

// LoggerElasticsearch contains Elasticsearch log hook settings.
type LoggerElasticsearch struct {
	Addresses []string `json:"addresses" yaml:"addresses"`
	Username  string   `json:"username" yaml:"username"`
	Password  string   `json:"password" yaml:"password"`
}

// LoggerOpenSearch contains OpenSearch log hook settings.
type LoggerOpenSearch struct {
	Addresses       []string `json:"addresses" yaml:"addresses"`
	Username        string   `json:"username" yaml:"username"`
	Password        string   `json:"password" yaml:"password"`
	InsecureSkipTLS bool     `json:"insecure_skip_tls" yaml:"insecure_skip_tls"`
}

var defaultLogSensitiveFields = []string{
	"password", "passwd", "pwd",
	"token", "access_token", "refresh_token", "auth_token",
	"secret", "api_key", "apikey",
	"credit_card", "card_number",
}

// GetLoggerConfig returns logging configuration from viper.
func GetLoggerConfig(v *viper.Viper) *Logger {
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

	indexName := strings.ToLower(fmt.Sprintf("%s-%s-logs", appName, environment))
	if v.IsSet("logger.index_name") && v.GetString("logger.index_name") != "" {
		indexName = v.GetString("logger.index_name")
	}

	return &Logger{
		Level:           v.GetInt("logger.level"),
		Format:          v.GetString("logger.format"),
		Path:            v.GetString("logger.path"),
		Output:          v.GetString("logger.output"),
		OutputFile:      v.GetString("logger.output_file"),
		IndexName:       indexName,
		DateSuffix:      getLogDateSuffixPattern(v),
		RotateDaily:     getLogRotateDaily(v),
		Desensitization: getLogDesensitizationConfig(v),
		Meilisearch:     getLoggerMeilisearchConfig(v),
		Elasticsearch:   getLoggerElasticsearchConfig(v),
		OpenSearch:      getLoggerOpenSearchConfig(v),
	}
}

func getLoggerConfig(v *viper.Viper) *Logger {
	return GetLoggerConfig(v)
}

// BuildIndexName builds the date-qualified log index name.
func (c *Logger) BuildIndexName(t time.Time) string {
	if c == nil {
		return ""
	}
	if !c.RotateDaily {
		return c.IndexName
	}
	return fmt.Sprintf("%s-%s", c.IndexName, t.Format(c.DateSuffix))
}

// GetCurrentIndexName returns the log index name for the current date.
func (c *Logger) GetCurrentIndexName() string {
	return c.BuildIndexName(time.Now())
}

func getLogDateSuffixPattern(v *viper.Viper) string {
	if v.IsSet("logger.date_suffix") {
		return v.GetString("logger.date_suffix")
	}
	return "2006-01-02"
}

func getLogRotateDaily(v *viper.Viper) bool {
	if v.IsSet("logger.rotate_daily") {
		return v.GetBool("logger.rotate_daily")
	}
	return true
}

func getLogDesensitizationConfig(v *viper.Viper) *LogDesensitization {
	if !v.IsSet("logger.desensitization") {
		return &LogDesensitization{
			Enabled:               true,
			SensitiveFields:       defaultLogSensitiveFields,
			PreservePrefix:        0,
			PreserveSuffix:        0,
			MaskChar:              "*",
			UseFixedLength:        true,
			FixedMaskLength:       6,
			ExactFieldMatch:       false,
			EnableDefaultPatterns: false,
		}
	}

	config := &LogDesensitization{
		Enabled:               v.GetBool("logger.desensitization.enabled"),
		SensitiveFields:       v.GetStringSlice("logger.desensitization.sensitive_fields"),
		CustomPatterns:        v.GetStringSlice("logger.desensitization.custom_patterns"),
		PreservePrefix:        v.GetInt("logger.desensitization.preserve_prefix"),
		PreserveSuffix:        v.GetInt("logger.desensitization.preserve_suffix"),
		MaskChar:              v.GetString("logger.desensitization.mask_char"),
		UseFixedLength:        v.GetBool("logger.desensitization.use_fixed_length"),
		FixedMaskLength:       v.GetInt("logger.desensitization.fixed_mask_length"),
		ExactFieldMatch:       v.GetBool("logger.desensitization.exact_field_match"),
		EnableDefaultPatterns: v.GetBool("logger.desensitization.enable_default_patterns"),
	}

	if len(config.SensitiveFields) == 0 {
		config.SensitiveFields = defaultLogSensitiveFields
	}
	if config.MaskChar == "" {
		config.MaskChar = "*"
	}
	if config.FixedMaskLength == 0 {
		config.FixedMaskLength = 8
	}
	if !v.IsSet("logger.desensitization.exact_field_match") {
		config.ExactFieldMatch = false
	}
	if !v.IsSet("logger.desensitization.enable_default_patterns") {
		config.EnableDefaultPatterns = false
	}

	return config
}

func getLoggerMeilisearchConfig(v *viper.Viper) *LoggerMeilisearch {
	if !v.IsSet("logger.meilisearch") {
		return nil
	}
	return &LoggerMeilisearch{
		Host:   v.GetString("logger.meilisearch.host"),
		APIKey: v.GetString("logger.meilisearch.api_key"),
	}
}

func getLoggerElasticsearchConfig(v *viper.Viper) *LoggerElasticsearch {
	if !v.IsSet("logger.elasticsearch") {
		return nil
	}
	return &LoggerElasticsearch{
		Addresses: v.GetStringSlice("logger.elasticsearch.addresses"),
		Username:  v.GetString("logger.elasticsearch.username"),
		Password:  v.GetString("logger.elasticsearch.password"),
	}
}

func getLoggerOpenSearchConfig(v *viper.Viper) *LoggerOpenSearch {
	if !v.IsSet("logger.opensearch") {
		return nil
	}
	return &LoggerOpenSearch{
		Addresses:       v.GetStringSlice("logger.opensearch.addresses"),
		Username:        v.GetString("logger.opensearch.username"),
		Password:        v.GetString("logger.opensearch.password"),
		InsecureSkipTLS: v.GetBool("logger.opensearch.insecure_skip_tls"),
	}
}
