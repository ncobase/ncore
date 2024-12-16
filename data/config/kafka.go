package config

import (
	"time"

	"github.com/spf13/viper"
)

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
