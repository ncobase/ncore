// Package kafka provides a Kafka driver for ncore/data.
//
// This driver uses kafka-go (github.com/segmentio/kafka-go) as the underlying
// client library. It registers itself automatically when imported:
//
//	import _ "github.com/ncobase/ncore/data/kafka"
//
// The driver supports Kafka connection configuration including brokers,
// consumer groups, topics, and timeout settings.
//
// Example usage:
//
//	driver, err := data.GetMessageDriver("kafka")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	cfg := &config.Kafka{
//	    Brokers:       []string{"localhost:9092"},
//	    ClientID:      "my-app",
//	    ConsumerGroup: "my-group",
//	    Topic:         "events",
//	}
//
//	conn, err := driver.Connect(ctx, cfg)
//	kafkaConn := conn.(*kafka.Conn)
package kafka

import (
	"context"
	"fmt"

	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/data/config"
	"github.com/segmentio/kafka-go"
)

// driver implements data.MessageDriver for Kafka.
type driver struct{}

// Name returns the driver identifier used in configuration files.
func (d *driver) Name() string {
	return "kafka"
}

// Connect establishes a Kafka connection using the provided configuration.
//
// The configuration must be a *config.Kafka containing:
//   - Brokers: List of Kafka broker addresses (e.g., ["localhost:9092"])
//   - ClientID: Client identifier for the connection
//   - ConsumerGroup: Consumer group name for consuming messages
//   - Topic: Default topic for producing/consuming
//   - ReadTimeout: Optional read timeout duration
//   - WriteTimeout: Optional write timeout duration
//   - ConnectTimeout: Optional connection timeout duration
//
// Returns a *kafka.Conn that can be used for producing and consuming messages.
func (d *driver) Connect(ctx context.Context, cfg any) (any, error) {
	kafkaCfg, ok := cfg.(*config.Kafka)
	if !ok {
		return nil, fmt.Errorf("kafka: invalid configuration type, expected *config.Kafka")
	}

	if len(kafkaCfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka: brokers are empty")
	}

	conn, err := kafka.DialContext(ctx, "tcp", kafkaCfg.Brokers[0])
	if err != nil {
		return nil, fmt.Errorf("kafka: failed to connect: %w", err)
	}

	return conn, nil
}

// Close terminates the Kafka connection and releases resources.
func (d *driver) Close(conn any) error {
	kafkaConn, ok := conn.(*kafka.Conn)
	if !ok {
		return fmt.Errorf("kafka: invalid connection type, expected *kafka.Conn")
	}

	if err := kafkaConn.Close(); err != nil {
		return fmt.Errorf("kafka: failed to close connection: %w", err)
	}

	return nil
}

// init registers the Kafka driver with the data package.
// This function is called automatically when the package is imported.
func init() {
	data.RegisterMessageDriver(&driver{})
}
