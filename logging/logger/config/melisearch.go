package config

import "github.com/spf13/viper"

// Meilisearch meilisearch config struct
type Meilisearch struct {
	Host   string `json:"host" yaml:"host"`
	APIKey string `json:"api_key" yaml:"api_key"`
}

// getMeilisearchConfigs reads Meilisearch configurations
func getMeilisearchConfigs(v *viper.Viper) *Meilisearch {
	if !v.IsSet("logger.meilisearch") {
		return nil
	}
	return &Meilisearch{
		Host:   v.GetString("logger.meilisearch.host"),
		APIKey: v.GetString("logger.meilisearch.api_key"),
	}
}
