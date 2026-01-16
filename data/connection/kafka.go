package connection

import (
	"context"
	"fmt"

	"github.com/ncobase/ncore/data/config"
)

func newKafkaConnection(conf *config.Kafka) (any, error) {
	if driverRegistry == nil {
		return nil, fmt.Errorf("driver registry not initialized, ensure drivers are imported")
	}

	driver, err := driverRegistry.GetMessageDriver("kafka")
	if err != nil {
		return nil, fmt.Errorf("failed to get kafka driver: %w", err)
	}

	conn, err := driver.Connect(context.Background(), conf)
	if err != nil {
		return nil, fmt.Errorf("failed to connect using kafka driver: %w", err)
	}

	return conn, nil
}
