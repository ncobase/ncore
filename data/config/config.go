package config

import (
	"github.com/spf13/viper"
)

// Config data config struct
type Config struct {
	Environment string `json:"environment"`
	*Database
	*Redis
	*Meilisearch
	*Elasticsearch
	*OpenSearch
	*MongoDB
	*Neo4j
	*RabbitMQ
	*Kafka
}

// GetConfig returns data config
func GetConfig(v *viper.Viper) *Config {
	return &Config{
		Environment:   v.GetString("data.environment"),
		Database:      getDatabaseConfig(v),
		Redis:         getRedisConfigs(v),
		Meilisearch:   getMeilisearchConfigs(v),
		Elasticsearch: getElasticsearchConfigs(v),
		OpenSearch:    getOpenSearchConfigs(v),
		MongoDB:       getMongoDBConfigs(v),
		Neo4j:         getNeo4jConfigs(v),
		RabbitMQ:      getRabbitMQConfigs(v),
		Kafka:         getKafkaConfigs(v),
	}
}
