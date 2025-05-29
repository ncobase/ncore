package config

import "github.com/spf13/viper"

// OpenSearch opensearch config struct
type OpenSearch struct {
	Addresses       []string `json:"addresses" yaml:"addresses"`
	Username        string   `json:"username" yaml:"username"`
	Password        string   `json:"password" yaml:"password"`
	InsecureSkipTLS bool     `json:"insecure_skip_tls" yaml:"insecure_skip_tls"`
}

// getOpenSearchConfigs reads OpenSearch configurations
func getOpenSearchConfigs(v *viper.Viper) *OpenSearch {
	return &OpenSearch{
		Addresses:       v.GetStringSlice("data.opensearch.addresses"),
		Username:        v.GetString("data.opensearch.username"),
		Password:        v.GetString("data.opensearch.password"),
		InsecureSkipTLS: v.GetBool("data.opensearch.insecure_skip_tls"),
	}
}
