package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ represents RabbitMQ implementation
type RabbitMQ struct {
	conn *amqp.Connection
}

// NewRabbitMQ creates new RabbitMQ connection
func NewRabbitMQ(conn *amqp.Connection) *RabbitMQ {
	return &RabbitMQ{conn: conn}
}

// PublishMessage publishes message to RabbitMQ
func (s *RabbitMQ) PublishMessage(exchange, routingKey string, body []byte) error {
	ch, err := s.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	defer func(ch *amqp.Channel) {
		_ = ch.Close()
	}(ch)

	err = ch.Publish(
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// ConsumeMessages consumes messages from RabbitMQ
func (s *RabbitMQ) ConsumeMessages(queue string, handler func([]byte) error) error {
	ch, err := s.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	defer func(ch *amqp.Channel) {
		_ = ch.Close()
	}(ch)

	msgs, err := ch.Consume(
		queue, // queue
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for d := range msgs {
			if err := handler(d.Body); err != nil {
				fmt.Printf("Failed to process message: %v\n", err)
			}
		}
	}()

	return nil
}

// Close closes the RabbitMQ service
func (s *RabbitMQ) Close() error {
	if err := s.conn.Close(); err != nil {
		return fmt.Errorf("failed to close RabbitMQ connection: %w", err)
	}
	return nil
}
