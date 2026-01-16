package data

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type rabbitMQ interface {
	IsConnected() bool
	PublishMessage(exchange, routingKey string, body []byte) error
	ConsumeMessages(queue string, handler func([]byte) error) error
}

type kafkaMQ interface {
	IsConnected() bool
	PublishMessage(ctx context.Context, topic string, key, value []byte) error
	ConsumeMessages(ctx context.Context, topic, groupID string, handler func([]byte) error) error
}

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

	if d.Conn == nil {
		return false
	}

	if rmq, ok := d.Conn.RMQ.(rabbitMQ); ok && rmq.IsConnected() {
		return true
	}
	if kfk, ok := d.Conn.KFK.(kafkaMQ); ok && kfk.IsConnected() {
		return true
	}
	return false
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
	conn := d.Conn
	d.mu.RUnlock()

	err := errors.New("RabbitMQ service not initialized")

	if closed {
		err = errors.New("data layer is closed")
	} else if conn != nil {
		if rmq, ok := conn.RMQ.(rabbitMQ); ok {
			if !rmq.IsConnected() {
				err = fmt.Errorf("RabbitMQ connection is not active")
			} else {
				err = rmq.PublishMessage(exchange, routingKey, body)
			}
		}
	}

	duration := time.Since(start)
	d.collector.MQPublish("rabbitmq", err)

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
	conn := d.Conn
	d.mu.RUnlock()

	if closed {
		err := errors.New("data layer is closed")
		d.collector.MQConsume("rabbitmq", err)
		return err
	}

	if conn == nil || conn.RMQ == nil {
		err := errors.New("rabbitmq service not initialized")
		d.collector.MQConsume("rabbitmq", err)
		return err
	}

	rmq, ok := conn.RMQ.(rabbitMQ)
	if !ok || !rmq.IsConnected() {
		err := fmt.Errorf("RabbitMQ connection is not active")
		d.collector.MQConsume("rabbitmq", err)
		return err
	}

	wrappedHandler := func(data []byte) error {
		start := time.Now()
		err := handler(data)
		duration := time.Since(start)

		d.collector.MQConsume("rabbitmq", err)

		if duration > 10*time.Second {
			d.collector.MQConsume("rabbitmq", errors.New("slow_consume"))
		}

		return err
	}

	return rmq.ConsumeMessages(queue, wrappedHandler)
}

// PublishToKafka publishes message to Kafka with metrics
func (d *Data) PublishToKafka(ctx context.Context, topic string, key, value []byte) error {
	if !d.IsMessagingEnabled() {
		return errors.New("messaging is disabled")
	}

	start := time.Now()

	d.mu.RLock()
	closed := d.closed
	conn := d.Conn
	d.mu.RUnlock()

	err := errors.New("kafka service not initialized")

	if closed {
		err = errors.New("data layer is closed")
	} else if conn != nil {
		if kfk, ok := conn.KFK.(kafkaMQ); ok {
			if !kfk.IsConnected() {
				err = fmt.Errorf("kafka connection is not active")
			} else {
				err = kfk.PublishMessage(ctx, topic, key, value)
			}
		}
	}

	duration := time.Since(start)
	d.collector.MQPublish("kafka", err)

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
	conn := d.Conn
	d.mu.RUnlock()

	if closed {
		err := errors.New("data layer is closed")
		d.collector.MQConsume("kafka", err)
		return err
	}

	if conn == nil || conn.KFK == nil {
		err := errors.New("kafka service not initialized")
		d.collector.MQConsume("kafka", err)
		return err
	}

	kfk, ok := conn.KFK.(kafkaMQ)
	if !ok || !kfk.IsConnected() {
		err := fmt.Errorf("kafka connection is not active")
		d.collector.MQConsume("kafka", err)
		return err
	}

	wrappedHandler := func(data []byte) error {
		start := time.Now()
		err := handler(data)
		duration := time.Since(start)

		d.collector.MQConsume("kafka", err)

		if duration > 10*time.Second {
			d.collector.MQConsume("kafka", errors.New("slow_consume"))
		}

		return err
	}

	return kfk.ConsumeMessages(ctx, topic, groupID, wrappedHandler)
}
