package data

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ncobase/ncore/data/config"
	"github.com/ncobase/ncore/data/connection"
	"github.com/ncobase/ncore/data/messaging/kafka"
	"github.com/ncobase/ncore/data/messaging/rabbitmq"
	"github.com/ncobase/ncore/data/metrics"
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
	Conn     *connection.Connections
	RabbitMQ *rabbitmq.RabbitMQ
	Kafka    *kafka.Kafka

	collector      metrics.Collector
	redisCollector *metrics.CacheCollector
	healthMonitor  *metrics.HealthMonitor
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

	// If not creating new and shared instance exists, return it
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
		collector: metrics.NoOpCollector{}, // Default no-op collector
	}

	// Initialize metrics components
	d.initMetricsComponents()

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
	// Default to shared instance for options-based creation
	return NewWithCreateNewAndOptions(cfg, false, opts...)
}

// NewWithCreateNewAndOptions creates new data layer with explicit control and options
func NewWithCreateNewAndOptions(cfg *config.Config, createNewInstance bool, opts ...Option) (*Data, func(name ...string), error) {
	// If not creating new and shared instance exists, return it
	if !createNewInstance && sharedInstance != nil {
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
		collector: metrics.NoOpCollector{}, // Default no-op collector
	}

	// Apply options
	for _, opt := range opts {
		opt(d)
	}

	// Initialize metrics components
	d.initMetricsComponents()

	// Set as shared instance if not creating new
	if !createNewInstance {
		sharedInstance = d
	}

	cleanup := func(name ...string) {
		if errs := d.Close(); len(errs) > 0 {
			fmt.Printf("cleanup errors: %v\n", errs)
		}
	}

	return d, cleanup, nil
}

// initMetricsComponents initializes components with metrics collection
func (d *Data) initMetricsComponents() {
	// Wrap Redis client with metrics collector
	if d.Conn != nil && d.Conn.RC != nil {
		d.redisCollector = metrics.NewCacheCollector(d.Conn.RC, d.collector)
	}

	// Initialize health monitor
	d.healthMonitor = metrics.NewHealthMonitor(d.collector)

	// Register health checkers
	d.registerHealthCheckers()
}

// registerHealthCheckers registers components for health monitoring
func (d *Data) registerHealthCheckers() {
	if d.healthMonitor == nil {
		return
	}

	// Database health checker
	if d.Conn != nil && d.Conn.DBM != nil {
		d.healthMonitor.RegisterComponent(&DatabaseHealthChecker{data: d})
	}

	// Redis health checker
	if d.redisCollector != nil {
		d.healthMonitor.RegisterComponent(&RedisHealthChecker{collector: d.redisCollector})
	}

	// MongoDB health checker
	if d.Conn != nil && d.Conn.MGM != nil {
		d.healthMonitor.RegisterComponent(&MongoHealthChecker{data: d})
	}

	// Messaging health checkers
	if d.RabbitMQ != nil && d.RabbitMQ.IsConnected() {
		d.healthMonitor.RegisterComponent(&RabbitMQHealthChecker{rabbitmq: d.RabbitMQ})
	}

	if d.Kafka != nil && d.Kafka.IsConnected() {
		d.healthMonitor.RegisterComponent(&KafkaHealthChecker{kafka: d.Kafka})
	}

	// Search engine health checkers
	if d.Conn != nil {
		if d.Conn.ES != nil {
			d.healthMonitor.RegisterComponent(&ElasticsearchHealthChecker{client: d.Conn.ES})
		}
		if d.Conn.OS != nil {
			d.healthMonitor.RegisterComponent(&OpenSearchHealthChecker{client: d.Conn.OS})
		}
		if d.Conn.MS != nil {
			d.healthMonitor.RegisterComponent(&MeilisearchHealthChecker{client: d.Conn.MS})
		}
	}
}

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

// DB returns the master database connection for write operations
// Deprecated: Use GetMasterDB() for better clarity
func (d *Data) DB() *sql.DB {
	return d.GetMasterDB()
}

// DBRead returns slave database connection for read operations
// Deprecated: Use GetSlaveDB() for better clarity
func (d *Data) DBRead() (*sql.DB, error) {
	return d.GetSlaveDB()
}

// GetRedis returns the Redis client with metrics
func (d *Data) GetRedis() *redis.Client {
	if d.redisCollector != nil {
		// Update connection metrics
		stats := d.redisCollector.PoolStats()
		if stats != nil {
			d.collector.RedisConnections(int(stats.TotalConns))
		}
		return d.redisCollector.GetClient()
	}

	if d.Conn != nil {
		return d.Conn.RC
	}
	return nil
}

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

// GetMongoManager returns the MongoDB client
func (d *Data) GetMongoManager() *connection.MongoManager {
	if d.Conn != nil {
		return d.Conn.MGM
	}
	return nil
}

// GetRedisCollector returns the Redis collector for direct metrics-aware operations
func (d *Data) GetRedisCollector() *metrics.CacheCollector {
	return d.redisCollector
}

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

	// Close connections
	if d.Conn != nil {
		if connErrs := d.Conn.Close(); len(connErrs) > 0 {
			errs = append(errs, connErrs...)
		}
	}

	return errs
}
