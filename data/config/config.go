package config

import (
	"github.com/spf13/viper"
)

// Config data config struct
type Config struct {
	Environment    string `yaml:"environment" json:"environment"`
	*Database      `yaml:"database" json:"database"`
	*Redis         `yaml:"redis" json:"redis"`
	*Meilisearch   `yaml:"meilisearch" json:"meilisearch"`
	*Elasticsearch `yaml:"elasticsearch" json:"elasticsearch"`
	*OpenSearch    `yaml:"opensearch" json:"opensearch"`
	*MongoDB       `yaml:"mongodb" json:"mongodb"`
	*Neo4j         `yaml:"neo4j" json:"neo4j"`
	*RabbitMQ      `yaml:"rabbitmq" json:"rabbitmq"`
	*Kafka         `yaml:"kafka" json:"kafka"`
	*Metrics       `yaml:"metrics" json:"metrics"`
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
		Metrics:       getMetricsConfig(v),
	}
}
