package config

import (
	"time"

	"github.com/spf13/viper"
)

// Data data config struct
type Data struct {
	*Database
	*Redis
	*Meilisearch
	*Elasticsearch
	*MongoDB
	*Neo4j
	*RabbitMQ
	*Kafka
}

// Database database config struct
type Database struct {
	Driver          string
	Source          string
	Migrate         bool
	Logging         bool
	MaxIdleConn     int
	MaxOpenConn     int
	ConnMaxLifeTime time.Duration
}

// Redis redis config struct
type Redis struct {
	Addr         string
	Username     string
	Password     string
	Db           int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	DialTimeout  time.Duration
}

// Meilisearch meilisearch config struct
type Meilisearch struct {
	Host   string `json:"host"`
	APIKey string `json:"api_key"`
}

// Elasticsearch elasticsearch config struct
type Elasticsearch struct {
	Addresses []string `json:"addresses"`
	Username  string   `json:"username"`
	Password  string   `json:"password"`
}

// MongoDB mongodb config struct
type MongoDB struct {
	URI      string
	Username string
	Password string
}

// Neo4j neo4j config struct
type Neo4j struct {
	URI      string
	Username string
	Password string
}

// RabbitMQ rabbitmq config struct
type RabbitMQ struct {
	URL               string
	Username          string
	Password          string
	Vhost             string
	ConnectionTimeout time.Duration
	HeartbeatInterval time.Duration
}

// Kafka kafka config struct
type Kafka struct {
	Brokers        []string
	ClientID       string
	ConsumerGroup  string
	Topic          string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	ConnectTimeout time.Duration
}

func getDataConfig(v *viper.Viper) *Data {
	return &Data{
		Database: &Database{
			Driver:          v.GetString("data.database.driver"),
			Source:          v.GetString("data.database.source"),
			Migrate:         v.GetBool("data.database.migrate"),
			Logging:         v.GetBool("data.database.logging"),
			MaxIdleConn:     v.GetInt("data.database.max_idle_conn"),
			MaxOpenConn:     v.GetInt("data.database.max_open_conn"),
			ConnMaxLifeTime: v.GetDuration("data.database.max_life_time"),
		},
		Redis: &Redis{
			Addr:         v.GetString("data.redis.addr"),
			Username:     v.GetString("data.redis.username"),
			Password:     v.GetString("data.redis.password"),
			Db:           v.GetInt("data.redis.db"),
			ReadTimeout:  v.GetDuration("data.redis.read_timeout"),
			WriteTimeout: v.GetDuration("data.redis.write_timeout"),
			DialTimeout:  v.GetDuration("data.redis.dial_timeout"),
		},
		Meilisearch: &Meilisearch{
			Host:   v.GetString("data.meilisearch.host"),
			APIKey: v.GetString("data.meilisearch.api_key"),
		},
		Elasticsearch: &Elasticsearch{
			Addresses: v.GetStringSlice("data.elasticsearch.addresses"),
			Username:  v.GetString("data.elasticsearch.username"),
			Password:  v.GetString("data.elasticsearch.password"),
		},
		MongoDB: &MongoDB{
			URI:      v.GetString("data.mongodb.uri"),
			Username: v.GetString("data.mongodb.username"),
			Password: v.GetString("data.mongodb.password"),
		},
		Neo4j: &Neo4j{
			URI:      v.GetString("data.neo4j.uri"),
			Username: v.GetString("data.neo4j.username"),
			Password: v.GetString("data.neo4j.password"),
		},
		RabbitMQ: &RabbitMQ{
			URL:               v.GetString("data.rabbitmq.url"),
			Username:          v.GetString("data.rabbitmq.username"),
			Password:          v.GetString("data.rabbitmq.password"),
			Vhost:             v.GetString("data.rabbitmq.vhost"),
			ConnectionTimeout: v.GetDuration("data.rabbitmq.connection_timeout"),
			HeartbeatInterval: v.GetDuration("data.rabbitmq.heartbeat_interval"),
		},
		Kafka: &Kafka{
			Brokers:        v.GetStringSlice("data.kafka.brokers"),
			ClientID:       v.GetString("data.kafka.client_id"),
			ConsumerGroup:  v.GetString("data.kafka.consumer_group"),
			Topic:          v.GetString("data.kafka.topic"),
			ReadTimeout:    v.GetDuration("data.kafka.read_timeout"),
			WriteTimeout:   v.GetDuration("data.kafka.write_timeout"),
			ConnectTimeout: v.GetDuration("data.kafka.connect_timeout"),
		},
	}
}
