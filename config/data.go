package config

import (
	"time"

	"github.com/spf13/viper"
)

// Data data config struct
type Data struct {
	Database
	Redis
	Meilisearch
	Elasticsearch
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

func getDataConfig(v *viper.Viper) Data {
	return Data{
		Database: Database{
			Driver:          v.GetString("data.database.driver"),
			Source:          v.GetString("data.database.source"),
			Migrate:         v.GetBool("data.database.migrate"),
			Logging:         v.GetBool("data.database.logging"),
			MaxIdleConn:     v.GetInt("data.database.max_idle_conn"),
			MaxOpenConn:     v.GetInt("data.database.max_open_conn"),
			ConnMaxLifeTime: v.GetDuration("data.database.max_life_time"),
		},
		Redis: Redis{
			Addr:         v.GetString("data.redis.addr"),
			Username:     v.GetString("data.redis.username"),
			Password:     v.GetString("data.redis.password"),
			Db:           v.GetInt("data.redis.db"),
			ReadTimeout:  v.GetDuration("data.redis.read_timeout"),
			WriteTimeout: v.GetDuration("data.redis.write_timeout"),
			DialTimeout:  v.GetDuration("data.redis.dial_timeout"),
		},
		Meilisearch: Meilisearch{
			Host:   v.GetString("data.meilisearch.host"),
			APIKey: v.GetString("data.meilisearch.api_key"),
		},
		Elasticsearch: Elasticsearch{
			Addresses: v.GetStringSlice("data.elasticsearch.addresses"),
			Username:  v.GetString("data.elasticsearch.username"),
			Password:  v.GetString("data.elasticsearch.password"),
		},
	}
}
