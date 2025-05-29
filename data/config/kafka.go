package config

import (
	"time"

	"github.com/spf13/viper"
)

// Kafka kafka config struct
type Kafka struct {
	Brokers        []string      `json:"brokers" yaml:"brokers"`
	ClientID       string        `json:"client_id" yaml:"client_id"`
	ConsumerGroup  string        `json:"consumer_group" yaml:"consumer_group"`
	Topic          string        `json:"topic" yaml:"topic"`
	ReadTimeout    time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout   time.Duration `json:"write_timeout" yaml:"write_timeout"`
	ConnectTimeout time.Duration `json:"connect_timeout" yaml:"connect_timeout"`
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
