package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// Kafka represents Kafka implementation
type Kafka struct {
	conn   *kafka.Conn
	writer *kafka.Writer
	reader *kafka.Reader
}

// New creates new Kafka service
func New(conn *kafka.Conn) *Kafka {
	if conn == nil {
		return nil
	}
	return &Kafka{
		conn: conn,
		writer: &kafka.Writer{
			Addr:         kafka.TCP(conn.RemoteAddr().String()),
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 10 * time.Millisecond,
		},
	}
}

// PublishMessage publishes message to Kafka
func (s *Kafka) PublishMessage(ctx context.Context, topic string, key, value []byte) error {
	err := s.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("failed to write messages: %w", err)
	}

	return nil
}

// ConsumeMessages consumes messages from Kafka
func (s *Kafka) ConsumeMessages(ctx context.Context, topic string, groupID string, handler func([]byte) error) error {
	if s.reader == nil {
		s.reader = kafka.NewReader(kafka.ReaderConfig{
			Brokers:  []string{s.conn.RemoteAddr().String()},
			GroupID:  groupID,
			Topic:    topic,
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		})
	}

	for {
		m, err := s.reader.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("failed to read message: %w", err)
		}

		if err := handler(m.Value); err != nil {
			fmt.Printf("Failed to process message: %v\n", err)
		}

		select {
		case <-ctx.Done():
			return nil
		default:
			// continue next message
			continue
		}
	}
}

// Close closes the Kafka service
func (s *Kafka) Close() error {
	var errs []error

	if s.writer != nil {
		if err := s.writer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Kafka writer: %w", err))
		}
	}

	if s.reader != nil {
		if err := s.reader.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Kafka reader: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing Kafka service: %v", errs)
	}

	return nil
}
