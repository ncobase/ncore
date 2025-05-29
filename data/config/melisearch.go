package config

import "github.com/spf13/viper"

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
