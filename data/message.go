package data

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// IsMessagingAvailable checks if any messaging system is available and properly connected
func (d *Data) IsMessagingAvailable() bool {
	return (d.RabbitMQ != nil && d.RabbitMQ.IsConnected()) ||
		(d.Kafka != nil && d.Kafka.IsConnected())
}

// PublishToRabbitMQ publishes message to RabbitMQ with metrics
func (d *Data) PublishToRabbitMQ(exchange, routingKey string, body []byte) error {
	start := time.Now()
	err := errors.New("RabbitMQ service not initialized")

	if d.RabbitMQ != nil {
		if !d.RabbitMQ.IsConnected() {
			err = fmt.Errorf("RabbitMQ connection is not active")
		} else {
			err = d.RabbitMQ.PublishMessage(exchange, routingKey, body)
		}
	}

	duration := time.Since(start)
	d.collector.MQPublish("rabbitmq", err)

	// Track slow publishing
	if duration > 5*time.Second {
		d.collector.MQPublish("rabbitmq", errors.New("slow_publish"))
	}

	return err
}

// ConsumeFromRabbitMQ consumes messages from RabbitMQ with metrics
func (d *Data) ConsumeFromRabbitMQ(queue string, handler func([]byte) error) error {
	if d.RabbitMQ == nil {
		err := errors.New("RabbitMQ service not initialized")
		d.collector.MQConsume("rabbitmq", err)
		return err
	}

	if !d.RabbitMQ.IsConnected() {
		err := fmt.Errorf("RabbitMQ connection is not active")
		d.collector.MQConsume("rabbitmq", err)
		return err
	}

	// Wrap handler with metrics
	wrappedHandler := func(data []byte) error {
		start := time.Now()
		err := handler(data)
		duration := time.Since(start)

		d.collector.MQConsume("rabbitmq", err)

		// Track slow message processing
		if duration > 10*time.Second {
			d.collector.MQConsume("rabbitmq", errors.New("slow_consume"))
		}

		return err
	}

	return d.RabbitMQ.ConsumeMessages(queue, wrappedHandler)
}

// PublishToKafka publishes message to Kafka with metrics
func (d *Data) PublishToKafka(ctx context.Context, topic string, key, value []byte) error {
	start := time.Now()
	err := errors.New("kafka service not initialized")

	if d.Kafka != nil {
		if !d.Kafka.IsConnected() {
			err = fmt.Errorf("kafka connection is not active")
		} else {
			err = d.Kafka.PublishMessage(ctx, topic, key, value)
		}
	}

	duration := time.Since(start)
	d.collector.MQPublish("kafka", err)

	// Track slow publishing
	if duration > 5*time.Second {
		d.collector.MQPublish("kafka", errors.New("slow_publish"))
	}

	return err
}

// ConsumeFromKafka consumes messages from Kafka with metrics
func (d *Data) ConsumeFromKafka(ctx context.Context, topic, groupID string, handler func([]byte) error) error {
	if d.Kafka == nil {
		err := errors.New("kafka service not initialized")
		d.collector.MQConsume("kafka", err)
		return err
	}

	if !d.Kafka.IsConnected() {
		err := fmt.Errorf("kafka connection is not active")
		d.collector.MQConsume("kafka", err)
		return err
	}

	// Wrap handler with metrics
	wrappedHandler := func(data []byte) error {
		start := time.Now()
		err := handler(data)
		duration := time.Since(start)

		d.collector.MQConsume("kafka", err)

		// Track slow message processing
		if duration > 10*time.Second {
			d.collector.MQConsume("kafka", errors.New("slow_consume"))
		}

		return err
	}

	return d.Kafka.ConsumeMessages(ctx, topic, groupID, wrappedHandler)
}
