package connection

import (
	"context"
	"fmt"

	"github.com/ncobase/ncore/data/config"
)

func newElasticsearchClient(conf *config.Elasticsearch) (any, error) {
	if driverRegistry == nil {
		return nil, fmt.Errorf("driver registry not initialized, ensure drivers are imported")
	}

	driver, err := driverRegistry.GetSearchDriver("elasticsearch")
	if err != nil {
		return nil, fmt.Errorf("failed to get elasticsearch driver: %w", err)
	}

	conn, err := driver.Connect(context.Background(), conf)
	if err != nil {
		return nil, fmt.Errorf("failed to connect using elasticsearch driver: %w", err)
	}

	return conn, nil
}
