package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ncobase/ncore/data/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ represents RabbitMQ implementation
type RabbitMQ struct {
	conn      *amqp.Connection
	messaging *config.Messaging
	mu        sync.Mutex
}

// NewRabbitMQ creates new RabbitMQ connection
func NewRabbitMQ(conn *amqp.Connection) *RabbitMQ {
	return &RabbitMQ{
		conn: conn,
		messaging: &config.Messaging{
			PublishTimeout:   30 * time.Second,
			CrossRegionMode:  false,
			RetryAttempts:    3,
			RetryBackoffMax:  30 * time.Second,
			FallbackToMemory: true,
		},
	}
}

// NewRabbitMQWithConfig creates new RabbitMQ with messaging config
func NewRabbitMQWithConfig(conn *amqp.Connection, messaging *config.Messaging) *RabbitMQ {
	return &RabbitMQ{
		conn:      conn,
		messaging: messaging,
	}
}

// IsConnected checks if the RabbitMQ connection is valid
func (s *RabbitMQ) IsConnected() bool {
	return s.conn != nil && !s.conn.IsClosed()
}

// ensureExchangeAndQueue ensures exchange and queue exist and are bound
func (s *RabbitMQ) ensureExchangeAndQueue(ch *amqp.Channel, exchangeName, queueName string) error {
	// Declare exchange
	err := ch.ExchangeDeclare(
		exchangeName, // exchange name
		"topic",      // exchange type
		true,         // durable
		false,        // auto-delete
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	q, err := ch.QueueDeclare(
		queueName, // queue name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = ch.QueueBind(
		q.Name,       // queue name
		queueName,    // routing key (same as queue name for simplicity)
		exchangeName, // exchange
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	return nil
}

// PublishMessage publishes message to RabbitMQ
func (s *RabbitMQ) PublishMessage(exchange, routingKey string, body []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.IsConnected() {
		return fmt.Errorf("rabbitmq connection is not available")
	}

	ch, err := s.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	var once sync.Once
	closeChannel := func() {
		once.Do(func() {
			if ch != nil {
				_ = ch.Close()
			}
		})
	}
	defer closeChannel()

	// Ensure exchange and queue exist
	if err = s.ensureExchangeAndQueue(ch, exchange, routingKey); err != nil {
		return err
	}

	// Set up confirmation mode for reliable publishing
	if err = ch.Confirm(false); err != nil {
		return fmt.Errorf("failed to put channel in confirm mode: %w", err)
	}

	// Create confirmation channel with buffer
	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	timeout := s.messaging.PublishTimeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = ch.PublishWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		true,       // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	// Wait for confirmation with a timeout
	select {
	case confirmed, ok := <-confirms:
		if !ok {
			return fmt.Errorf("confirmation channel closed")
		}
		if !confirmed.Ack {
			return fmt.Errorf("failed to receive publish confirmation")
		}
	case <-time.After(timeout):
		return fmt.Errorf("publish confirmation timed out after %v", timeout)
	}

	return nil
}

// ConsumeMessages consumes messages from RabbitMQ
func (s *RabbitMQ) ConsumeMessages(queue string, handler func([]byte) error) error {
	if !s.IsConnected() {
		return fmt.Errorf("rabbitmq connection is not available")
	}

	ch, err := s.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	if err := s.ensureExchangeAndQueue(ch, queue, queue); err != nil {
		_ = ch.Close()
		return err
	}

	// Set QoS
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		_ = ch.Close()
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Start consuming
	msgs, err := ch.Consume(
		queue, // queue
		"",    // consumer (empty means auto-generated)
		false, // auto-ack (set to false for manual ack)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		_ = ch.Close()
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		defer func() {
			if ch != nil {
				_ = ch.Close()
			}
		}()

		for d := range msgs {
			if err := handler(d.Body); err != nil {
				fmt.Printf("Failed to process message: %v\n", err)
			}

			if err := d.Ack(false); err != nil {
				fmt.Printf("Failed to acknowledge message: %v\n", err)
			}
		}
	}()

	return nil
}

// Close closes the RabbitMQ service
func (s *RabbitMQ) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.IsConnected() {
		return nil
	}

	if err := s.conn.Close(); err != nil {
		return fmt.Errorf("failed to close RabbitMQ connection: %w", err)
	}
	return nil
}
