package connection

import (
	"context"
	"fmt"

	"github.com/ncobase/ncore/data/config"
)

func newMeilisearchClient(conf *config.Meilisearch) (any, error) {
	if driverRegistry == nil {
		return nil, fmt.Errorf("driver registry not initialized, ensure drivers are imported")
	}

	driver, err := driverRegistry.GetSearchDriver("meilisearch")
	if err != nil {
		return nil, fmt.Errorf("failed to get meilisearch driver: %w", err)
	}

	conn, err := driver.Connect(context.Background(), conf)
	if err != nil {
		return nil, fmt.Errorf("failed to connect using meilisearch driver: %w", err)
	}

	return conn, nil
}
