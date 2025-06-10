package data

import (
	"fmt"
	"time"

	"github.com/ncobase/ncore/data/config"
	"github.com/ncobase/ncore/data/connection"
	"github.com/ncobase/ncore/data/messaging/kafka"
	"github.com/ncobase/ncore/data/messaging/rabbitmq"
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

	// Initialize messaging systems
	d.initMessaging()

	cleanup := func(name ...string) {
		if errs := d.Close(); len(errs) > 0 {
			fmt.Printf("cleanup errors: %v\n", errs)
		}
	}

	return d, cleanup, nil
}

// NewWithOptions creates new data layer with options
func NewWithOptions(cfg *config.Config, opts ...Option) (*Data, func(name ...string), error) {
	conn, err := connection.New(cfg)
	if err != nil {
		return nil, nil, err
	}

	d := &Data{
		Conn:      conn,
		collector: metrics.NoOpCollector{},
	}

	// Apply options
	for _, opt := range opts {
		opt(d)
	}

	// Initialize metrics collector if enabled
	if cfg.Metrics != nil && cfg.Metrics.Enabled {
		if err := d.initMetricsCollector(cfg.Metrics); err != nil {
			fmt.Printf("failed to initialize metrics collector: %v\n", err)
		}
	}

	// Initialize messaging systems
	d.initMessaging()

	cleanup := func(name ...string) {
		if errs := d.Close(); len(errs) > 0 {
			fmt.Printf("cleanup errors: %v\n", errs)
		}
	}

	return d, cleanup, nil
}

// initMetricsCollector initializes metrics collector based on config
func (d *Data) initMetricsCollector(cfg *config.Metrics) error {
	// Use Redis storage if configured and available
	if cfg.StorageType == "redis" && d.Conn != nil && d.Conn.RC != nil {
		keyPrefix := cfg.KeyPrefix
		retention := time.Duration(cfg.RetentionDays) * 24 * time.Hour
		batchSize := cfg.BatchSize

		collector := metrics.NewDataCollectorWithRedis(d.Conn.RC, keyPrefix, retention, batchSize)
		d.collector = collector
		return nil
	}

	// Fallback to memory storage
	collector := metrics.NewDataCollector(cfg.BatchSize)
	d.collector = collector
	return nil
}

// initMessaging initializes messaging systems
func (d *Data) initMessaging() {
	if d.Conn.RMQ != nil {
		d.RabbitMQ = rabbitmq.NewRabbitMQ(d.Conn.RMQ)
	}

	if d.Conn.KFK != nil {
		d.Kafka = kafka.New(d.Conn.KFK)
	}
}
