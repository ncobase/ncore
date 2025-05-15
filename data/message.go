package data

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrRabbitMQNotInitialized = errors.New("RabbitMQ service not initialized")
	ErrKafkaNotInitialized    = errors.New("kafka service not initialized")
	ErrConnectionUnavailable  = errors.New("messaging service connection unavailable")
)

// IsMessagingAvailable checks if any messaging system is available and properly connected
func (d *Data) IsMessagingAvailable() bool {
	return (d.RabbitMQ != nil && d.RabbitMQ.IsConnected()) ||
		(d.Kafka != nil && d.Kafka.IsConnected())
}

// PublishToRabbitMQ publishes message to RabbitMQ
func (d *Data) PublishToRabbitMQ(exchange, routingKey string, body []byte) error {
	if d.RabbitMQ == nil {
		return ErrRabbitMQNotInitialized
	}

	if !d.RabbitMQ.IsConnected() {
		return fmt.Errorf("%w: RabbitMQ connection is not active", ErrConnectionUnavailable)
	}

	return d.RabbitMQ.PublishMessage(exchange, routingKey, body)
}

// ConsumeFromRabbitMQ consumes messages from RabbitMQ
func (d *Data) ConsumeFromRabbitMQ(queue string, handler func([]byte) error) error {
	if d.RabbitMQ == nil {
		return ErrRabbitMQNotInitialized
	}

	if !d.RabbitMQ.IsConnected() {
		return fmt.Errorf("%w: RabbitMQ connection is not active", ErrConnectionUnavailable)
	}

	return d.RabbitMQ.ConsumeMessages(queue, handler)
}

// PublishToKafka publishes message to Kafka
func (d *Data) PublishToKafka(ctx context.Context, topic string, key, value []byte) error {
	if d.Kafka == nil {
		return ErrKafkaNotInitialized
	}

	if !d.Kafka.IsConnected() {
		return fmt.Errorf("%w: Kafka connection is not active", ErrConnectionUnavailable)
	}

	return d.Kafka.PublishMessage(ctx, topic, key, value)
}

// ConsumeFromKafka consumes messages from Kafka
func (d *Data) ConsumeFromKafka(ctx context.Context, topic, groupID string, handler func([]byte) error) error {
	if d.Kafka == nil {
		return ErrKafkaNotInitialized
	}

	if !d.Kafka.IsConnected() {
		return fmt.Errorf("%w: Kafka connection is not active", ErrConnectionUnavailable)
	}

	return d.Kafka.ConsumeMessages(ctx, topic, groupID, handler)
}
