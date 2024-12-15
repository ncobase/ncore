package connection

import (
	"context"
	"ncobase/common/config"
	"ncobase/common/logger"

	"github.com/segmentio/kafka-go"
)

// newKafkaConnection creates a new Kafka connection
func newKafkaConnection(conf *config.Kafka) (*kafka.Conn, error) {
	if conf == nil || len(conf.Brokers) == 0 {
		logger.Infof(context.Background(), "Kafka configuration is nil or empty")
		return nil, nil
	}

	conn, err := kafka.DialContext(context.Background(), "tcp", conf.Brokers[0])
	if err != nil {
		logger.Errorf(context.Background(), "Failed to connect to Kafka: %v", err)
		return nil, err
	}

	logger.Infof(context.Background(), "Connected to Kafka")
	return conn, nil
}
