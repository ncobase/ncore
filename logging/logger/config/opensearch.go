package config

import "github.com/spf13/viper"

// OpenSearch opensearch config struct
type OpenSearch struct {
	Addresses       []string `json:"addresses"`
	Username        string   `json:"username"`
	Password        string   `json:"password"`
	InsecureSkipTLS bool     `json:"insecure_skip_tls"`
}

// getOpenSearchConfigs reads OpenSearch configurations
func getOpenSearchConfigs(v *viper.Viper) *OpenSearch {
	if !v.IsSet("logger.opensearch") {
		return nil
	}
	return &OpenSearch{
		Addresses:       v.GetStringSlice("logger.opensearch.addresses"),
		Username:        v.GetString("logger.opensearch.username"),
		Password:        v.GetString("logger.opensearch.password"),
		InsecureSkipTLS: v.GetBool("logger.opensearch.insecure_skip_tls"),
	}
}
