package config

import "github.com/spf13/viper"

// Elasticsearch elasticsearch config struct
type Elasticsearch struct {
	Addresses []string `json:"addresses" yaml:"addresses"`
	Username  string   `json:"username" yaml:"username"`
	Password  string   `json:"password" yaml:"password"`
}

// getElasticsearchConfigs reads Elasticsearch configurations
func getElasticsearchConfigs(v *viper.Viper) *Elasticsearch {
	if !v.IsSet("logger.elasticsearch") {
		return nil
	}
	return &Elasticsearch{
		Addresses: v.GetStringSlice("logger.elasticsearch.addresses"),
		Username:  v.GetString("logger.elasticsearch.username"),
		Password:  v.GetString("logger.elasticsearch.password"),
	}
}
