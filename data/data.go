package data

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ncobase/ncore/data/config"
	"github.com/ncobase/ncore/data/connection"
	"github.com/ncobase/ncore/data/messaging/kafka"
	"github.com/ncobase/ncore/data/messaging/rabbitmq"
	"github.com/ncobase/ncore/data/metrics"
	"github.com/ncobase/ncore/data/search"
	"github.com/ncobase/ncore/data/search/elastic"
	"github.com/ncobase/ncore/data/search/meili"
	"github.com/ncobase/ncore/data/search/opensearch"
	"github.com/redis/go-redis/v9"
)

type ContextKey string

const (
	ContextKeyTransaction ContextKey = "tx"
)

var sharedInstance *Data

// Data represents the data layer implementation
type Data struct {
	Conn         *connection.Connections
	RabbitMQ     *rabbitmq.RabbitMQ
	Kafka        *kafka.Kafka
	searchClient *search.Client
	collector    metrics.Collector
	searchOnce   sync.Once
}

// Option function type for configuring Data
type Option func(*Data)

// WithMetricsCollector sets the metrics collector
func WithMetricsCollector(collector metrics.Collector) Option {
	return func(d *Data) {
		if collector != nil {
			d.collector = collector
		}
	}
}

// WithExtensionCollector sets extension layer collector using adapter
func WithExtensionCollector(collector metrics.ExtensionCollector) Option {
	return func(d *Data) {
		if collector != nil {
			d.collector = metrics.NewExtensionCollectorAdapter(collector)
		}
	}
}

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
		RabbitMQ:  rabbitmq.NewRabbitMQ(conn.RMQ),
		Kafka:     kafka.New(conn.KFK),
		collector: metrics.NoOpCollector{},
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

// NewWithOptions creates new data layer with options
func NewWithOptions(cfg *config.Config, opts ...Option) (*Data, func(name ...string), error) {
	conn, err := connection.New(cfg)
	if err != nil {
		return nil, nil, err
	}

	d := &Data{
		Conn:      conn,
		RabbitMQ:  rabbitmq.NewRabbitMQ(conn.RMQ),
		Kafka:     kafka.New(conn.KFK),
		collector: metrics.NoOpCollector{},
	}

	// Apply options
	for _, opt := range opts {
		opt(d)
	}

	cleanup := func(name ...string) {
		if errs := d.Close(); len(errs) > 0 {
			fmt.Printf("cleanup errors: %v\n", errs)
		}
	}

	return d, cleanup, nil
}

// initSearchClient initializes search client lazily and safely
func (d *Data) initSearchClient() {
	d.searchOnce.Do(func() {
		d.searchClient = search.NewClient(
			d.GetElasticsearch(),
			d.GetOpenSearch(),
			d.GetMeilisearch(),
			d.collector,
		)
	})
}

// getSearchClient returns initialized search client
func (d *Data) getSearchClient() *search.Client {
	d.initSearchClient()
	return d.searchClient
}

// Database Access Methods

// GetDBManager returns the database manager
func (d *Data) GetDBManager() *connection.DBManager {
	if d.Conn != nil {
		return d.Conn.DBM
	}
	return nil
}

// GetMasterDB returns the master database connection for write operations
func (d *Data) GetMasterDB() *sql.DB {
	if d.Conn != nil {
		return d.Conn.DB()
	}
	return nil
}

// GetSlaveDB returns slave database connection for read operations
func (d *Data) GetSlaveDB() (*sql.DB, error) {
	if d.Conn != nil {
		return d.Conn.ReadDB()
	}
	return nil, errors.New("no database connection available")
}

// Cache Access Methods

// GetRedis returns the Redis client
func (d *Data) GetRedis() *redis.Client {
	if d.Conn != nil {
		// Update connection metrics
		if d.Conn.RC != nil {
			d.collector.RedisConnections(int(d.Conn.RC.PoolStats().TotalConns))
		}
		return d.Conn.RC
	}
	return nil
}

// Search Engine Access Methods

// GetMeilisearch returns the Meilisearch client
func (d *Data) GetMeilisearch() *meili.Client {
	if d.Conn != nil {
		return d.Conn.MS
	}
	return nil
}

// GetElasticsearch returns the Elasticsearch client
func (d *Data) GetElasticsearch() *elastic.Client {
	if d.Conn != nil {
		return d.Conn.ES
	}
	return nil
}

// GetOpenSearch returns the OpenSearch client
func (d *Data) GetOpenSearch() *opensearch.Client {
	if d.Conn != nil {
		return d.Conn.OS
	}
	return nil
}

// MongoDB Access Methods

// GetMongoManager returns the MongoDB manager
func (d *Data) GetMongoManager() *connection.MongoManager {
	if d.Conn != nil {
		return d.Conn.MGM
	}
	return nil
}

// Metrics Access Methods

// GetMetricsCollector returns the metrics collector
func (d *Data) GetMetricsCollector() metrics.Collector {
	return d.collector
}

// GetStats returns data layer statistics
func (d *Data) GetStats() map[string]any {
	if defaultCollector, ok := d.collector.(*metrics.DefaultCollector); ok {
		return defaultCollector.GetStats()
	}

	return map[string]any{
		"status":    "metrics_unavailable",
		"timestamp": time.Now(),
	}
}

// Close closes all data connections
func (d *Data) Close() []error {
	var errs []error

	if d.Conn != nil {
		if connErrs := d.Conn.Close(); len(connErrs) > 0 {
			errs = append(errs, connErrs...)
		}
	}

	return errs
}

// Deprecated methods for backward compatibility

// DB returns the master database connection
// Deprecated: Use GetMasterDB() for better clarity
func (d *Data) DB() *sql.DB { return d.GetMasterDB() }

// DBRead returns a slave database connection
// Deprecated: Use GetSlaveDB() for better clarity
func (d *Data) DBRead() (*sql.DB, error) { return d.GetSlaveDB() }
