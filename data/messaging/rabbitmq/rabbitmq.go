package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ represents RabbitMQ implementation
type RabbitMQ struct {
	conn *amqp.Connection
	mu   sync.Mutex // Add mutex for thread safety
}

// NewRabbitMQ creates new RabbitMQ connection
func NewRabbitMQ(conn *amqp.Connection) *RabbitMQ {
	return &RabbitMQ{conn: conn}
}

// IsConnected checks if the RabbitMQ connection is valid
func (s *RabbitMQ) IsConnected() bool {
	return s.conn != nil && !s.conn.IsClosed()
}

// ensureExchangeAndQueue ensures exchange and queue exist and are bound
func (s *RabbitMQ) ensureExchangeAndQueue(ch *amqp.Channel, exchangeName, queueName string) error {
	// Declare exchange - using topic exchange for flexibility
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

	// Important: Use defer with a named function to safely handle channel closing
	// This prevents attempting to close an already closed channel
	var once sync.Once
	closeChannel := func() {
		once.Do(func() {
			if ch != nil {
				_ = ch.Close() // Ignore close errors
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

	// Publish the message with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	err = ch.PublishWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		true,       // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // Ensure message persistence
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
	case <-time.After(120 * time.Second):
		return fmt.Errorf("publish confirmation timed out")
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

	// Ensure exchange and queue exist
	if err := s.ensureExchangeAndQueue(ch, queue, queue); err != nil {
		_ = ch.Close() // Close channel on error
		return err
	}

	// Set QoS (prefetch count)
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		_ = ch.Close() // Close channel on error
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
		_ = ch.Close() // Close channel on error
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	// Handle messages in a separate goroutine
	go func() {
		defer func() {
			// Safely close the channel
			if ch != nil {
				_ = ch.Close() // Ignore close errors
			}
		}()

		for d := range msgs {
			if err := handler(d.Body); err != nil {
				// Log error but don't nack to avoid redelivery loops
				fmt.Printf("Failed to process message: %v\n", err)
			}

			// Acknowledge message
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
