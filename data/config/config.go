package config

import (
	"github.com/spf13/viper"
)

// Config data config struct
type Config struct {
	*Database  `yaml:"database" json:"database"`
	*Redis     `yaml:"redis" json:"redis"`
	*Search    `yaml:"search" json:"search"`
	*MongoDB   `yaml:"mongodb" json:"mongodb"`
	*Neo4j     `yaml:"neo4j" json:"neo4j"`
	*RabbitMQ  `yaml:"rabbitmq" json:"rabbitmq"`
	*Kafka     `yaml:"kafka" json:"kafka"`
	*Metrics   `yaml:"metrics" json:"metrics"`
	*Messaging `yaml:"messaging" json:"messaging"`
}

// GetConfig returns data config
func GetConfig(v *viper.Viper) *Config {
	return &Config{
		Database:  getDatabaseConfig(v),
		Redis:     getRedisConfigs(v),
		Search:    getSearchConfig(v),
		MongoDB:   getMongoDBConfigs(v),
		Neo4j:     getNeo4jConfigs(v),
		RabbitMQ:  getRabbitMQConfigs(v),
		Kafka:     getKafkaConfigs(v),
		Metrics:   getMetricsConfig(v),
		Messaging: getMessagingConfig(v),
	}
}
