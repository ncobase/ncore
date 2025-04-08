package connection

import (
	"errors"
	"fmt"
	"ncore/pkg/data/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

// newRabbitMQConnection creates a new RabbitMQ connection
func newRabbitMQConnection(conf *config.RabbitMQ) (*amqp.Connection, error) {
	if conf == nil || conf.URL == "" {
		return nil, errors.New("RabbitMQ configuration is nil or empty")
	}

	url := fmt.Sprintf("amqp://%s:%s@%s/%s", conf.Username, conf.Password, conf.URL, conf.Vhost)
	conn, err := amqp.DialConfig(url, amqp.Config{
		Heartbeat: conf.HeartbeatInterval,
		Vhost:     conf.Vhost,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return conn, nil
}
