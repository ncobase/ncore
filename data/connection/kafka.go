package connection

import (
	"context"
	"errors"
	"fmt"

	"github.com/ncobase/ncore/data/config"
	"github.com/segmentio/kafka-go"
)

// newKafkaConnection creates a new Kafka connection
func newKafkaConnection(conf *config.Kafka) (*kafka.Conn, error) {
	if conf == nil || len(conf.Brokers) == 0 {
		return nil, errors.New("kafka configuration is nil or empty")
	}

	conn, err := kafka.DialContext(context.Background(), "tcp", conf.Brokers[0])
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Kafka: %w", err)
	}

	return conn, nil
}
