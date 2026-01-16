package data

import (
	"fmt"

	"github.com/ncobase/ncore/data/config"
	"github.com/ncobase/ncore/data/connection"
	"github.com/ncobase/ncore/data/metrics"
)

// New creates new data layer
func New(cfg *config.Config, createNewInstance ...bool) (*Data, func(name ...string), error) {
	var createNew bool
	if len(createNewInstance) > 0 {
		createNew = createNewInstance[0]
	}

	// Return shared instance if exists and not creating new
	if !createNew && sharedInstance != nil {
		cleanup := func(name ...string) {
			if errs := sharedInstance.Close(); len(errs) > 0 {
				fmt.Printf("cleanup errors: %v\n", errs)
			}
		}
		return sharedInstance, cleanup, nil
	}

	conn, err := connection.New(cfg)
	if err != nil {
		return nil, nil, err
	}

	d := &Data{
		Conn:      conn,
		conf:      cfg,
		collector: metrics.NoOpCollector{},
	}

	// Initialize metrics collector if enabled
	if cfg.Metrics != nil && cfg.Metrics.Enabled {
		if err := d.initMetricsCollector(cfg.Metrics); err != nil {
			fmt.Printf("failed to initialize metrics collector: %v\n", err)
		}
	}

	// Set as shared instance if not creating new
	if !createNew {
		sharedInstance = d
	}

	cleanup := func(name ...string) {
		if errs := d.Close(); len(errs) > 0 {
			fmt.Printf("cleanup errors: %v\n", errs)
		}
	}

	return d, cleanup, nil
}

// initMetricsCollector initializes metrics collector based on config
func (d *Data) initMetricsCollector(cfg *config.Metrics) error {
	collector := metrics.NewDataCollector(cfg.BatchSize)
	d.collector = collector
	return nil
}
