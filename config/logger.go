package config

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

func getLoggerConfig() Logger {
	return Logger{
		Level:      c.GetInt("logger.level"),
		Format:     c.GetString("logger.format"),
		Path:       c.GetString("logger.path"),
		Output:     c.GetString("logger.output"),
		OutputFile: c.GetString("logger.output_file"),
		Meilisearch: Meilisearch{
			Host:   c.GetString("data.meilisearch.host"),
			APIKey: c.GetString("data.meilisearch.api_key"),
		},
		Elasticsearch: Elasticsearch{
			Addresses: c.GetStringSlice("data.elasticsearch.addresses"),
			Username:  c.GetString("data.elasticsearch.username"),
			Password:  c.GetString("data.elasticsearch.password"),
		},
		IndexName: c.GetString("app_name") + "_log",
	}
}
