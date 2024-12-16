package data

import (
	"context"
	"errors"
)

var (
	ErrRabbitMQNotInitialized = errors.New("RabbitMQ service not initialized")
	ErrKafkaNotInitialized    = errors.New("kafka service not initialized")
)

// PublishToRabbitMQ publishes message to RabbitMQ
func (d *Data) PublishToRabbitMQ(exchange, routingKey string, body []byte) error {
	if d.RabbitMQ == nil {
		return ErrRabbitMQNotInitialized
	}
	return d.RabbitMQ.PublishMessage(exchange, routingKey, body)
}

// ConsumeFromRabbitMQ consumes messages from RabbitMQ
func (d *Data) ConsumeFromRabbitMQ(queue string, handler func([]byte) error) error {
	if d.RabbitMQ == nil {
		return ErrRabbitMQNotInitialized
	}
	return d.RabbitMQ.ConsumeMessages(queue, handler)
}

// PublishToKafka publishes message to Kafka
func (d *Data) PublishToKafka(ctx context.Context, topic string, key, value []byte) error {
	if d.Kafka == nil {
		return ErrKafkaNotInitialized
	}
	return d.Kafka.PublishMessage(ctx, topic, key, value)
}

// ConsumeFromKafka consumes messages from Kafka
func (d *Data) ConsumeFromKafka(ctx context.Context, topic, groupID string, handler func([]byte) error) error {
	if d.Kafka == nil {
		return ErrKafkaNotInitialized
	}
	return d.Kafka.ConsumeMessages(ctx, topic, groupID, handler)
}
