package data

import (
	"context"
	"errors"
)

var (
	ErrRabbitMQNotInitialized = errors.New("RabbitMQ service not initialized")
	ErrKafkaNotInitialized    = errors.New("kafka service not initialized")
)

func (d *Data) PublishToRabbitMQ(exchange, routingKey string, body []byte) error {
	if d.Svc.RabbitMQ == nil {
		return ErrRabbitMQNotInitialized
	}
	return d.Svc.RabbitMQ.PublishMessage(exchange, routingKey, body)
}

func (d *Data) ConsumeFromRabbitMQ(queue string, handler func([]byte) error) error {
	if d.Svc.RabbitMQ == nil {
		return ErrRabbitMQNotInitialized
	}
	return d.Svc.RabbitMQ.ConsumeMessages(queue, handler)
}

func (d *Data) PublishToKafka(ctx context.Context, topic string, key, value []byte) error {
	if d.Svc.Kafka == nil {
		return ErrKafkaNotInitialized
	}
	return d.Svc.Kafka.PublishMessage(ctx, topic, key, value)
}

func (d *Data) ConsumeFromKafka(ctx context.Context, topic, groupID string, handler func([]byte) error) error {
	if d.Svc.Kafka == nil {
		return ErrKafkaNotInitialized
	}
	return d.Svc.Kafka.ConsumeMessages(ctx, topic, groupID, handler)
}
