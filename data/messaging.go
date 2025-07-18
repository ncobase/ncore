package data

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// IsMessagingEnabled checks if messaging services
func (d *Data) IsMessagingEnabled() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed {
		return false
	}

	// Check if messaging is disabled in config
	if d.conf.Messaging != nil {
		return d.conf.Messaging.IsEnabled()
	}

	// Default to true if no messaging config
	return true
}

// IsQueueAvailable checks if external message queues are available
func (d *Data) IsQueueAvailable() bool {
	if !d.IsMessagingEnabled() {
		return false
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed {
		return false
	}

	return (d.RabbitMQ != nil && d.RabbitMQ.IsConnected()) ||
		(d.Kafka != nil && d.Kafka.IsConnected())
}

// ShouldUseMemoryFallback checks if should fallback to memory when queue unavailable
func (d *Data) ShouldUseMemoryFallback() bool {
	if !d.IsMessagingEnabled() {
		return false
	}

	if d.conf.Messaging != nil {
		return d.conf.Messaging.ShouldUseMemoryFallback()
	}

	return true
}

// IsMessagingAvailable checks if any messaging (queue or memory) is available
// Deprecated: Use IsMessagingEnabled() and IsQueueAvailable() separately
func (d *Data) IsMessagingAvailable() bool {
	return d.IsMessagingEnabled() && (d.IsQueueAvailable() || d.ShouldUseMemoryFallback())
}

// PublishToRabbitMQ publishes message to RabbitMQ with metrics
func (d *Data) PublishToRabbitMQ(exchange, routingKey string, body []byte) error {
	if !d.IsMessagingEnabled() {
		return errors.New("messaging is disabled")
	}

	start := time.Now()

	d.mu.RLock()
	closed := d.closed
	rabbitmq := d.RabbitMQ
	d.mu.RUnlock()

	err := errors.New("RabbitMQ service not initialized")

	if closed {
		err = errors.New("data layer is closed")
	} else if rabbitmq != nil {
		if !rabbitmq.IsConnected() {
			err = fmt.Errorf("RabbitMQ connection is not active")
		} else {
			err = rabbitmq.PublishMessage(exchange, routingKey, body)
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
	if !d.IsMessagingEnabled() {
		return errors.New("messaging is disabled")
	}

	d.mu.RLock()
	closed := d.closed
	rabbitmq := d.RabbitMQ
	d.mu.RUnlock()

	if closed {
		err := errors.New("data layer is closed")
		d.collector.MQConsume("rabbitmq", err)
		return err
	}

	if rabbitmq == nil {
		err := errors.New("RabbitMQ service not initialized")
		d.collector.MQConsume("rabbitmq", err)
		return err
	}

	if !rabbitmq.IsConnected() {
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

	return rabbitmq.ConsumeMessages(queue, wrappedHandler)
}

// PublishToKafka publishes message to Kafka with metrics
func (d *Data) PublishToKafka(ctx context.Context, topic string, key, value []byte) error {
	if !d.IsMessagingEnabled() {
		return errors.New("messaging is disabled")
	}

	start := time.Now()

	d.mu.RLock()
	closed := d.closed
	kafka := d.Kafka
	d.mu.RUnlock()

	err := errors.New("kafka service not initialized")

	if closed {
		err = errors.New("data layer is closed")
	} else if kafka != nil {
		if !kafka.IsConnected() {
			err = fmt.Errorf("kafka connection is not active")
		} else {
			err = kafka.PublishMessage(ctx, topic, key, value)
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
	if !d.IsMessagingEnabled() {
		return errors.New("messaging is disabled")
	}

	d.mu.RLock()
	closed := d.closed
	kafka := d.Kafka
	d.mu.RUnlock()

	if closed {
		err := errors.New("data layer is closed")
		d.collector.MQConsume("kafka", err)
		return err
	}

	if kafka == nil {
		err := errors.New("kafka service not initialized")
		d.collector.MQConsume("kafka", err)
		return err
	}

	if !kafka.IsConnected() {
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

	return kafka.ConsumeMessages(ctx, topic, groupID, wrappedHandler)
}
