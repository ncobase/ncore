package config

import "github.com/spf13/viper"

// Logger logger config struct
type Logger struct {
	Level      int
	Path       string
	Format     string
	Output     string
	OutputFile string
	IndexName  string
	Meilisearch
	Elasticsearch
}

func getLoggerConfig(v *viper.Viper) Logger {
	return Logger{
		Level:      v.GetInt("logger.level"),
		Format:     v.GetString("logger.format"),
		Path:       v.GetString("logger.path"),
		Output:     v.GetString("logger.output"),
		OutputFile: v.GetString("logger.output_file"),
		Meilisearch: Meilisearch{
			Host:   v.GetString("data.meilisearch.host"),
			APIKey: v.GetString("data.meilisearch.api_key"),
		},
		Elasticsearch: Elasticsearch{
			Addresses: v.GetStringSlice("data.elasticsearch.addresses"),
			Username:  v.GetString("data.elasticsearch.username"),
			Password:  v.GetString("data.elasticsearch.password"),
		},
		IndexName: v.GetString("app_name") + "_log",
	}
}
