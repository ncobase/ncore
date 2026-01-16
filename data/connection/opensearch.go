package connection

import (
	"context"
	"fmt"

	"github.com/ncobase/ncore/data/config"
)

func newOpenSearchClient(conf *config.OpenSearch) (any, error) {
	if driverRegistry == nil {
		return nil, fmt.Errorf("driver registry not initialized, ensure drivers are imported")
	}

	driver, err := driverRegistry.GetSearchDriver("opensearch")
	if err != nil {
		return nil, fmt.Errorf("failed to get opensearch driver: %w", err)
	}

	conn, err := driver.Connect(context.Background(), conf)
	if err != nil {
		return nil, fmt.Errorf("failed to connect using opensearch driver: %w", err)
	}

	return conn, nil
}
