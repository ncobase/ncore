package service

import (
	"context"
	"fmt"
	"ncobase/common/log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQService represents a RabbitMQ service
type RabbitMQService struct {
	conn *amqp.Connection
}

// NewRabbitMQService creates a new RabbitMQ service
func NewRabbitMQService(conn *amqp.Connection) *RabbitMQService {
	return &RabbitMQService{conn: conn}
}

// PublishMessage publishes a message to RabbitMQ
func (s *RabbitMQService) PublishMessage(exchange, routingKey string, body []byte) error {
	ch, err := s.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

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
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	return nil
}

// ConsumeMessages consumes messages from RabbitMQ
func (s *RabbitMQService) ConsumeMessages(queue string, handler func([]byte) error) error {
	ch, err := s.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

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
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	go func() {
		for d := range msgs {
			if err := handler(d.Body); err != nil {
				log.Errorf(context.Background(), "Failed to process a message: %v", err)
				fmt.Println(err)
			}
		}
	}()

	return nil
}

// Close closes the RabbitMQ service
func (s *RabbitMQService) Close() error {
	if err := s.conn.Close(); err != nil {
		return fmt.Errorf("failed to close RabbitMQ connection: %w", err)
	}
	return s.conn.Close()
}
