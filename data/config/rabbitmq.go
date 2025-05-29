package config

import (
	"time"

	"github.com/spf13/viper"
)

// RabbitMQ rabbitmq config struct
type RabbitMQ struct {
	URL               string        `json:"url" yaml:"url"`
	Username          string        `json:"username" yaml:"username"`
	Password          string        `json:"password" yaml:"password"`
	Vhost             string        `json:"vhost" yaml:"vhost"`
	ConnectionTimeout time.Duration `json:"connection_timeout" yaml:"connection_timeout"`
	HeartbeatInterval time.Duration `json:"heartbeat_interval" yaml:"heartbeat_interval"`
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
