package connection

import (
	"context"
	"fmt"

	"github.com/ncobase/ncore/data/config"
)

func newRabbitMQConnection(conf *config.RabbitMQ) (any, error) {
	if driverRegistry == nil {
		return nil, fmt.Errorf("driver registry not initialized, ensure drivers are imported")
	}

	driver, err := driverRegistry.GetMessageDriver("rabbitmq")
	if err != nil {
		return nil, fmt.Errorf("failed to get rabbitmq driver: %w", err)
	}

	conn, err := driver.Connect(context.Background(), conf)
	if err != nil {
		return nil, fmt.Errorf("failed to connect using rabbitmq driver: %w", err)
	}

	return conn, nil
}
