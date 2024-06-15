package config

import "time"

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

func getDataConfig() Data {
	return Data{
		Database: Database{
			Driver:          c.GetString("data.database.driver"),
			Source:          c.GetString("data.database.source"),
			Migrate:         c.GetBool("data.database.migrate"),
			Logging:         c.GetBool("data.database.logging"),
			MaxIdleConn:     c.GetInt("data.database.max_idle_conn"),
			MaxOpenConn:     c.GetInt("data.database.max_open_conn"),
			ConnMaxLifeTime: c.GetDuration("data.database.max_life_time"),
		},
		Redis: Redis{
			Addr:         c.GetString("data.redis.addr"),
			Username:     c.GetString("data.redis.username"),
			Password:     c.GetString("data.redis.password"),
			Db:           c.GetInt("data.redis.db"),
			ReadTimeout:  c.GetDuration("data.redis.read_timeout"),
			WriteTimeout: c.GetDuration("data.redis.write_timeout"),
			DialTimeout:  c.GetDuration("data.redis.dial_timeout"),
		},
		Meilisearch: Meilisearch{
			Host:   c.GetString("data.meilisearch.host"),
			APIKey: c.GetString("data.meilisearch.api_key"),
		},
		Elasticsearch: Elasticsearch{
			Addresses: c.GetStringSlice("data.elasticsearch.addresses"),
			Username:  c.GetString("data.elasticsearch.username"),
			Password:  c.GetString("data.elasticsearch.password"),
		},
	}
}
