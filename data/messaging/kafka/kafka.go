package kafka

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ncobase/ncore/data/config"
	"github.com/segmentio/kafka-go"
)

// Kafka represents Kafka implementation
type Kafka struct {
	conn      *kafka.Conn
	brokers   []string
	messaging *config.Messaging
	mu        sync.Mutex
	writer    *kafka.Writer
	readers   map[string]*kafka.Reader
	readersMu sync.RWMutex
}

// New creates new Kafka service
func New(conn *kafka.Conn) *Kafka {
	if conn == nil {
		return nil
	}

	var brokers []string
	remoteBroker := conn.RemoteAddr().String()
	if remoteBroker != "" {
		brokers = []string{remoteBroker}
	}

	return &Kafka{
		conn:    conn,
		brokers: brokers,
		readers: make(map[string]*kafka.Reader),
		messaging: &config.Messaging{
			PublishTimeout:   30 * time.Second,
			CrossRegionMode:  false,
			RetryAttempts:    3,
			RetryBackoffMax:  30 * time.Second,
			FallbackToMemory: true,
		},
		writer: &kafka.Writer{
			Addr:         kafka.TCP(remoteBroker),
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 10 * time.Millisecond,
			RequiredAcks: kafka.RequireAll,
			Async:        false,
		},
	}
}

// NewWithConfig creates new Kafka service with messaging config
func NewWithConfig(conn *kafka.Conn, messaging *config.Messaging) *Kafka {
	k := New(conn)
	if k != nil {
		k.messaging = messaging
	}
	return k
}

// IsConnected checks if the Kafka connection is valid
func (s *Kafka) IsConnected() bool {
	if s == nil || s.conn == nil {
		return false
	}

	_, err := s.conn.Controller()
	return err == nil
}

// PublishMessage publishes message to Kafka
func (s *Kafka) PublishMessage(ctx context.Context, topic string, key, value []byte) error {
	if !s.IsConnected() {
		return fmt.Errorf("kafka connection is not available")
	}

	writer := s.getWriter()
	if writer == nil {
		return errors.New("kafka writer is not initialized")
	}

	timeout := s.messaging.PublishTimeout
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	msg := kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
		Time:  time.Now(),
	}

	maxRetries := s.messaging.RetryAttempts
	backoff := 100 * time.Millisecond

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := writer.WriteMessages(timeoutCtx, msg)
		if err == nil {
			return nil
		}

		if timeoutCtx.Err() != nil {
			return fmt.Errorf("publish context timeout: %w", timeoutCtx.Err())
		}

		if attempt < maxRetries {
			time.Sleep(backoff)
			backoff *= 2
			if backoff > s.messaging.RetryBackoffMax {
				backoff = s.messaging.RetryBackoffMax
			}
		}
	}

	return fmt.Errorf("failed to write message after %d attempts", maxRetries+1)
}

// getWriter ensures a valid writer exists and returns it
func (s *Kafka) getWriter() *kafka.Writer {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.writer == nil && len(s.brokers) > 0 {
		s.writer = &kafka.Writer{
			Addr:         kafka.TCP(s.brokers...),
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 10 * time.Millisecond,
			RequiredAcks: kafka.RequireAll,
			Async:        false,
		}
	}

	return s.writer
}

// ConsumeMessages consumes messages from Kafka
func (s *Kafka) ConsumeMessages(ctx context.Context, topic string, groupID string, handler func([]byte) error) error {
	if !s.IsConnected() {
		return fmt.Errorf("kafka connection is not available")
	}

	// Get or create reader
	reader := s.getReader(topic, groupID)
	if reader == nil {
		return errors.New("failed to create Kafka reader")
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered from panic in Kafka consumer: %v\n", r)
			}
		}()

		for {
			// Check if context is done
			select {
			case <-ctx.Done():
				// Clean up reader
				s.closeReader(topic, groupID)
				return
			default:
				// Continue processing
			}

			readCtx, cancel := context.WithTimeout(ctx, s.messaging.PublishTimeout)
			m, err := reader.FetchMessage(readCtx)
			cancel()

			if err != nil {
				if err == io.EOF || err == context.Canceled {
					// Reader closed or context canceled
					return
				}

				if !errors.Is(err, context.DeadlineExceeded) {
					// Only log non-timeout errors
					fmt.Printf("Error reading Kafka message: %v\n", err)
					time.Sleep(1 * time.Second)
				}
				continue
			}

			// Process message
			err = handler(m.Value)
			if err != nil {
				fmt.Printf("Error processing Kafka message: %v\n", err)
				// Continue without committing - message will be redelivered
				continue
			}

			// Commit message offset
			commitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			if err := reader.CommitMessages(commitCtx, m); err != nil {
				fmt.Printf("Failed to commit Kafka message: %v\n", err)
			}
			cancel()
		}
	}()

	return nil
}

// getReader gets or creates a reader for the specified topic
func (s *Kafka) getReader(topic, groupID string) *kafka.Reader {
	key := topic + ":" + groupID

	s.readersMu.RLock()
	reader, exists := s.readers[key]
	s.readersMu.RUnlock()

	if exists && reader != nil {
		return reader
	}

	s.readersMu.Lock()
	defer s.readersMu.Unlock()

	if reader, exists = s.readers[key]; exists && reader != nil {
		return reader
	}

	reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:         s.brokers,
		GroupID:         groupID,
		Topic:           topic,
		MinBytes:        10e3,
		MaxBytes:        10e6,
		MaxWait:         500 * time.Millisecond,
		StartOffset:     kafka.LastOffset,
		CommitInterval:  1 * time.Second,
		ReadLagInterval: -1,
		ReadBackoffMin:  100 * time.Millisecond,
		ReadBackoffMax:  5 * time.Second,
		ErrorLogger:     kafka.LoggerFunc(logKafkaError),
	})

	s.readers[key] = reader
	return reader
}

// closeReader safely closes a reader and removes it from the map
func (s *Kafka) closeReader(topic, groupID string) {
	key := topic + ":" + groupID

	s.readersMu.Lock()
	defer s.readersMu.Unlock()

	if reader, exists := s.readers[key]; exists && reader != nil {
		_ = reader.Close()
		delete(s.readers, key)
	}
}

// Close closes the Kafka service
func (s *Kafka) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var errs []error

	// Close writer
	if s.writer != nil {
		if err := s.writer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Kafka writer: %w", err))
		}
		s.writer = nil
	}

	// Close all readers
	s.readersMu.Lock()
	for key, reader := range s.readers {
		if err := reader.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Kafka reader %s: %w", key, err))
		}
		delete(s.readers, key)
	}
	s.readersMu.Unlock()

	// Close connection
	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Kafka connection: %w", err))
		}
		s.conn = nil
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing Kafka service: %v", errs)
	}

	return nil
}

func logKafkaError(msg string, args ...any) {
	fmt.Printf("Kafka error: "+msg+"\n", args...)
}
