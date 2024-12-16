package config

import (
	"time"

	"github.com/spf13/viper"
)

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
