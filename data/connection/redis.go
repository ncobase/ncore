package connection

import (
	"context"
	"fmt"

	"github.com/ncobase/ncore/data/config"
)

func newRedisClient(conf *config.Redis) (any, error) {
	if driverRegistry == nil {
		return nil, fmt.Errorf("driver registry not initialized, ensure drivers are imported")
	}

	driver, err := driverRegistry.GetCacheDriver("redis")
	if err != nil {
		return nil, fmt.Errorf("failed to get redis driver: %w", err)
	}

	conn, err := driver.Connect(context.Background(), conf)
	if err != nil {
		return nil, fmt.Errorf("failed to connect using redis driver: %w", err)
	}

	return conn, nil
}
