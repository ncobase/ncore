package config

import (
	"fmt"
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
	Master   *DBNode   `json:"master"`
	Slaves   []*DBNode `json:"slaves"`
	Migrate  bool      `json:"migrate"`
	Strategy string    `json:"strategy"`
	MaxRetry int       `json:"max_retry"`
}

// DBNode represents a single database node configuration
type DBNode struct {
	Driver          string        `json:"driver"`
	Source          string        `json:"source"`
	Logging         bool          `json:"logging"`
	MaxIdleConn     int           `json:"max_idle_conn"`
	MaxOpenConn     int           `json:"max_open_conn"`
	ConnMaxLifeTime time.Duration `json:"conn_max_life_time"`
	Weight          int           `json:"weight"`
}

func getDataConfig(v *viper.Viper) *Data {
	return &Data{
		Database: &Database{
			Master: &DBNode{
				Driver:          v.GetString("data.database.master.driver"),
				Source:          v.GetString("data.database.master.source"),
				Logging:         v.GetBool("data.database.master.logging"),
				MaxIdleConn:     v.GetInt("data.database.master.max_idle_conn"),
				MaxOpenConn:     v.GetInt("data.database.master.max_open_conn"),
				ConnMaxLifeTime: v.GetDuration("data.database.master.max_life_time"),
				Weight:          v.GetInt("data.database.master.weight"),
			},
			Slaves:   getSlaveConfigs(v),
			Migrate:  v.GetBool("data.database.migrate"),
			Strategy: v.GetString("data.database.strategy"),
			MaxRetry: v.GetInt("data.database.max_retry"),
		},
		Redis:         getRedisConfigs(v),
		Meilisearch:   getMeilisearchConfigs(v),
		Elasticsearch: getElasticsearchConfigs(v),
		MongoDB:       getMongoDBConfigs(v),
		Neo4j:         getNeo4jConfigs(v),
		RabbitMQ:      getRabbitMQConfigs(v),
		Kafka:         getKafkaConfigs(v),
	}
}

// getSlaveConfigs reads slave database configurations
func getSlaveConfigs(v *viper.Viper) []*DBNode {
	var slaves []*DBNode

	slavesConfig := v.Get("data.database.slaves")
	if slavesConfig == nil {
		return slaves
	}

	slavesList, ok := slavesConfig.([]any)
	if !ok {
		return slaves
	}

	slavesCount := len(slavesList)
	for i := 0; i < slavesCount; i++ {
		slave := &DBNode{
			Driver:          v.GetString(fmt.Sprintf("data.database.slaves.%d.driver", i)),
			Source:          v.GetString(fmt.Sprintf("data.database.slaves.%d.source", i)),
			Logging:         v.GetBool(fmt.Sprintf("data.database.slaves.%d.logging", i)),
			MaxIdleConn:     v.GetInt(fmt.Sprintf("data.database.slaves.%d.max_idle_conn", i)),
			MaxOpenConn:     v.GetInt(fmt.Sprintf("data.database.slaves.%d.max_open_conn", i)),
			ConnMaxLifeTime: v.GetDuration(fmt.Sprintf("data.database.slaves.%d.max_life_time", i)),
			Weight:          v.GetInt(fmt.Sprintf("data.database.slaves.%d.weight", i)),
		}
		slaves = append(slaves, slave)
	}
	return slaves
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

// getRedisConfigs reads Redis configurations
func getRedisConfigs(v *viper.Viper) *Redis {
	return &Redis{
		Addr:         v.GetString("data.redis.addr"),
		Username:     v.GetString("data.redis.username"),
		Password:     v.GetString("data.redis.password"),
		Db:           v.GetInt("data.redis.db"),
		ReadTimeout:  v.GetDuration("data.redis.read_timeout"),
		WriteTimeout: v.GetDuration("data.redis.write_timeout"),
		DialTimeout:  v.GetDuration("data.redis.dial_timeout"),
	}
}

// Meilisearch meilisearch config struct
type Meilisearch struct {
	Host   string `json:"host"`
	APIKey string `json:"api_key"`
}

// getMeilisearchConfigs reads Meilisearch configurations
func getMeilisearchConfigs(v *viper.Viper) *Meilisearch {
	return &Meilisearch{
		Host:   v.GetString("data.meilisearch.host"),
		APIKey: v.GetString("data.meilisearch.api_key"),
	}
}

// Elasticsearch elasticsearch config struct
type Elasticsearch struct {
	Addresses []string `json:"addresses"`
	Username  string   `json:"username"`
	Password  string   `json:"password"`
}

// getElasticsearchConfigs reads Elasticsearch configurations
func getElasticsearchConfigs(v *viper.Viper) *Elasticsearch {
	return &Elasticsearch{
		Addresses: v.GetStringSlice("data.elasticsearch.addresses"),
		Username:  v.GetString("data.elasticsearch.username"),
		Password:  v.GetString("data.elasticsearch.password"),
	}
}

// MongoDB mongodb config struct
type MongoDB struct {
	Master   *MongoNode   `json:"master"`
	Slaves   []*MongoNode `json:"slaves"`
	Strategy string       `json:"strategy"`
	MaxRetry int          `json:"max_retry"`
}

// MongoNode mongodb node config
type MongoNode struct {
	URI     string `json:"uri"`
	Logging bool   `json:"logging"`
	Weight  int    `json:"weight"`
}

// getMongoDBConfigs reads MongoDB configurations
func getMongoDBConfigs(v *viper.Viper) *MongoDB {
	return &MongoDB{
		Master: &MongoNode{
			URI:     v.GetString("data.mongodb.master.uri"),
			Logging: v.GetBool("data.mongodb.master.logging"),
		},
		Slaves:   getMongoSlaveConfigs(v),
		Strategy: v.GetString("data.mongodb.strategy"),
		MaxRetry: v.GetInt("data.mongodb.max_retry"),
	}
}

// getMongoSlaveConfigs reads MongoDB slave configurations
func getMongoSlaveConfigs(v *viper.Viper) []*MongoNode {
	var slaves []*MongoNode

	// get mongodb slaves
	slavesConfig := v.Get("data.mongodb.slaves")
	if slavesConfig == nil {
		return slaves
	}

	// check if the slaves config is a slice
	slavesInterface, ok := slavesConfig.([]interface{})
	if !ok {
		fmt.Println("Invalid mongodb slaves configuration format")
		return slaves
	}

	// parse each slave
	for i := 0; i < len(slavesInterface); i++ {
		slave := &MongoNode{
			URI:     v.GetString(fmt.Sprintf("data.mongodb.slaves.%d.uri", i)),
			Logging: v.GetBool(fmt.Sprintf("data.mongodb.slaves.%d.logging", i)),
			Weight:  v.GetInt(fmt.Sprintf("data.mongodb.slaves.%d.weight", i)),
		}

		// check if the slave is valid
		if slave.URI != "" {
			// set default values
			if slave.Weight <= 0 {
				slave.Weight = 1
			}
			slaves = append(slaves, slave)
		}
	}

	return slaves
}

// Neo4j neo4j config struct
type Neo4j struct {
	URI      string
	Username string
	Password string
}

// getNeo4jConfigs reads Neo4j configurations
func getNeo4jConfigs(v *viper.Viper) *Neo4j {
	return &Neo4j{
		URI:      v.GetString("data.neo4j.uri"),
		Username: v.GetString("data.neo4j.username"),
		Password: v.GetString("data.neo4j.password"),
	}
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

// getRabbitMQConfigs reads RabbitMQ configurations
func getRabbitMQConfigs(v *viper.Viper) *RabbitMQ {
	return &RabbitMQ{
		URL:               v.GetString("data.rabbitmq.url"),
		Username:          v.GetString("data.rabbitmq.username"),
		Password:          v.GetString("data.rabbitmq.password"),
		Vhost:             v.GetString("data.rabbitmq.vhost"),
		ConnectionTimeout: v.GetDuration("data.rabbitmq.connection_timeout"),
		HeartbeatInterval: v.GetDuration("data.rabbitmq.heartbeat_interval"),
	}
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

// getKafkaConfigs reads Kafka configurations
func getKafkaConfigs(v *viper.Viper) *Kafka {
	return &Kafka{
		Brokers:        v.GetStringSlice("data.kafka.brokers"),
		ClientID:       v.GetString("data.kafka.client_id"),
		ConsumerGroup:  v.GetString("data.kafka.consumer_group"),
		Topic:          v.GetString("data.kafka.topic"),
		ReadTimeout:    v.GetDuration("data.kafka.read_timeout"),
		WriteTimeout:   v.GetDuration("data.kafka.write_timeout"),
		ConnectTimeout: v.GetDuration("data.kafka.connect_timeout"),
	}
}
