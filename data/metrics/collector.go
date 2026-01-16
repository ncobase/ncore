package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

type Collector interface {
	DBQuery(duration time.Duration, err error)
	DBTransaction(err error)
	DBConnections(count int)
	RedisCommand(command string, err error)
	RedisConnections(count int)
	MongoOperation(operation string, err error)
	SearchQuery(engine string, err error)
	SearchIndex(engine, operation string)
	MQPublish(system string, err error)
	MQConsume(system string, err error)
	HealthCheck(component string, healthy bool)
}

type CacheMetricsCollector interface {
	RedisCommand(command string, err error)
}

type NoOpCollector struct{}

func (NoOpCollector) DBQuery(time.Duration, error) {}
func (NoOpCollector) DBTransaction(error)          {}
func (NoOpCollector) DBConnections(int)            {}
func (NoOpCollector) RedisCommand(string, error)   {}
func (NoOpCollector) RedisConnections(int)         {}
func (NoOpCollector) MongoOperation(string, error) {}
func (NoOpCollector) SearchQuery(string, error)    {}
func (NoOpCollector) SearchIndex(string, string)   {}
func (NoOpCollector) MQPublish(string, error)      {}
func (NoOpCollector) MQConsume(string, error)      {}
func (NoOpCollector) HealthCheck(string, bool)     {}

type DataCollector struct {
	dbQueries      atomic.Int64
	dbQueryErrors  atomic.Int64
	dbTransactions atomic.Int64
	dbTxErrors     atomic.Int64
	dbConnections  atomic.Int32
	dbSlowQueries  atomic.Int64

	redisCommands    atomic.Int64
	redisErrors      atomic.Int64
	redisConnections atomic.Int32

	mongoOperations atomic.Int64
	mongoErrors     atomic.Int64

	searchQueries  atomic.Int64
	searchErrors   atomic.Int64
	searchIndexOps atomic.Int64

	mqPublished     atomic.Int64
	mqPublishErrors atomic.Int64
	mqConsumed      atomic.Int64
	mqConsumeErrors atomic.Int64

	healthChecks map[string]*atomic.Bool
	healthMu     sync.RWMutex

	lastDBQuery      atomic.Value
	lastRedisCommand atomic.Value
	lastMongoOp      atomic.Value
	lastSearchQuery  atomic.Value
	lastMQOperation  atomic.Value

	storage   Storage
	batchSize int
	buffer    []Metric
	bufferMu  sync.Mutex
}

type Metric struct {
	Type      string    `json:"type"`
	Value     int64     `json:"value"`
	Labels    Labels    `json:"labels"`
	Timestamp time.Time `json:"timestamp"`
}

type Labels map[string]string

type Storage interface {
	Store(metrics []Metric) error
	Query(query QueryRequest) ([]Metric, error)
	Close() error
}

type QueryRequest struct {
	Type      string    `json:"type"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Labels    Labels    `json:"labels"`
	Limit     int       `json:"limit"`
}

func NewDataCollector(batchSize int) *DataCollector {
	if batchSize <= 0 {
		batchSize = 100
	}

	c := &DataCollector{
		healthChecks: make(map[string]*atomic.Bool),
		storage:      NewMemoryStorage(),
		batchSize:    batchSize,
		buffer:       make([]Metric, 0, batchSize),
	}

	now := time.Now()
	c.lastDBQuery.Store(now)
	c.lastRedisCommand.Store(now)
	c.lastMongoOp.Store(now)
	c.lastSearchQuery.Store(now)
	c.lastMQOperation.Store(now)

	return c
}

func (c *DataCollector) DBQuery(duration time.Duration, err error) {
	c.dbQueries.Add(1)
	c.lastDBQuery.Store(time.Now())

	if err != nil {
		c.dbQueryErrors.Add(1)
	}

	if duration > time.Second {
		c.dbSlowQueries.Add(1)
	}

	c.recordMetric("db_query", 1, Labels{
		"success": boolToString(err == nil),
		"slow":    boolToString(duration > time.Second),
	})
}

func (c *DataCollector) DBTransaction(err error) {
	c.dbTransactions.Add(1)
	if err != nil {
		c.dbTxErrors.Add(1)
	}

	c.recordMetric("db_transaction", 1, Labels{
		"success": boolToString(err == nil),
	})
}

func (c *DataCollector) DBConnections(count int) {
	c.dbConnections.Store(int32(count))
	c.recordMetric("db_connections", int64(count), nil)
}

func (c *DataCollector) RedisCommand(command string, err error) {
	c.redisCommands.Add(1)
	c.lastRedisCommand.Store(time.Now())

	if err != nil {
		c.redisErrors.Add(1)
	}

	c.recordMetric("redis_command", 1, Labels{
		"command": command,
		"success": boolToString(err == nil),
	})
}

func (c *DataCollector) RedisConnections(count int) {
	c.redisConnections.Store(int32(count))
	c.recordMetric("redis_connections", int64(count), nil)
}

func (c *DataCollector) MongoOperation(operation string, err error) {
	c.mongoOperations.Add(1)
	c.lastMongoOp.Store(time.Now())

	if err != nil {
		c.mongoErrors.Add(1)
	}

	c.recordMetric("mongo_operation", 1, Labels{
		"operation": operation,
		"success":   boolToString(err == nil),
	})
}

func (c *DataCollector) SearchQuery(engine string, err error) {
	c.searchQueries.Add(1)
	c.lastSearchQuery.Store(time.Now())

	if err != nil {
		c.searchErrors.Add(1)
	}

	c.recordMetric("search_query", 1, Labels{
		"engine":  engine,
		"success": boolToString(err == nil),
	})
}

func (c *DataCollector) SearchIndex(engine, operation string) {
	c.searchIndexOps.Add(1)

	c.recordMetric("search_index", 1, Labels{
		"engine":    engine,
		"operation": operation,
	})
}

func (c *DataCollector) MQPublish(system string, err error) {
	c.mqPublished.Add(1)
	c.lastMQOperation.Store(time.Now())

	if err != nil {
		c.mqPublishErrors.Add(1)
	}

	c.recordMetric("mq_publish", 1, Labels{
		"system":  system,
		"success": boolToString(err == nil),
	})
}

func (c *DataCollector) MQConsume(system string, err error) {
	c.mqConsumed.Add(1)
	c.lastMQOperation.Store(time.Now())

	if err != nil {
		c.mqConsumeErrors.Add(1)
	}

	c.recordMetric("mq_consume", 1, Labels{
		"system":  system,
		"success": boolToString(err == nil),
	})
}

func (c *DataCollector) HealthCheck(component string, healthy bool) {
	c.healthMu.Lock()
	if _, exists := c.healthChecks[component]; !exists {
		c.healthChecks[component] = &atomic.Bool{}
	}
	healthCheck := c.healthChecks[component]
	c.healthMu.Unlock()

	healthCheck.Store(healthy)

	c.recordMetric("health_check", boolToInt(healthy), Labels{
		"component": component,
	})
}

func (c *DataCollector) recordMetric(metricType string, value int64, labels Labels) {
	metric := Metric{
		Type:      metricType,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
	}

	c.bufferMu.Lock()
	c.buffer = append(c.buffer, metric)
	shouldFlush := len(c.buffer) >= c.batchSize
	c.bufferMu.Unlock()

	if shouldFlush {
		c.flush()
	}
}

func (c *DataCollector) flush() {
	c.bufferMu.Lock()
	if len(c.buffer) == 0 {
		c.bufferMu.Unlock()
		return
	}

	metrics := make([]Metric, len(c.buffer))
	copy(metrics, c.buffer)
	c.buffer = c.buffer[:0]
	c.bufferMu.Unlock()

	if c.storage != nil {
		_ = c.storage.Store(metrics)
	}
}

func (c *DataCollector) GetStats() map[string]any {
	c.healthMu.RLock()
	healthStatus := make(map[string]bool)
	for component, status := range c.healthChecks {
		healthStatus[component] = status.Load()
	}
	c.healthMu.RUnlock()

	return map[string]any{
		"database": map[string]any{
			"connections":  c.dbConnections.Load(),
			"queries":      c.dbQueries.Load(),
			"errors":       c.dbQueryErrors.Load(),
			"slow_queries": c.dbSlowQueries.Load(),
			"transactions": c.dbTransactions.Load(),
			"tx_errors":    c.dbTxErrors.Load(),
			"last_query":   c.lastDBQuery.Load(),
		},
		"redis": map[string]any{
			"connections":  c.redisConnections.Load(),
			"commands":     c.redisCommands.Load(),
			"errors":       c.redisErrors.Load(),
			"last_command": c.lastRedisCommand.Load(),
		},
		"mongodb": map[string]any{
			"operations":     c.mongoOperations.Load(),
			"errors":         c.mongoErrors.Load(),
			"last_operation": c.lastMongoOp.Load(),
		},
		"search": map[string]any{
			"queries":    c.searchQueries.Load(),
			"errors":     c.searchErrors.Load(),
			"index_ops":  c.searchIndexOps.Load(),
			"last_query": c.lastSearchQuery.Load(),
		},
		"messaging": map[string]any{
			"published":      c.mqPublished.Load(),
			"publish_errors": c.mqPublishErrors.Load(),
			"consumed":       c.mqConsumed.Load(),
			"consume_errors": c.mqConsumeErrors.Load(),
			"last_operation": c.lastMQOperation.Load(),
		},
		"health":    healthStatus,
		"timestamp": time.Now(),
	}
}

func (c *DataCollector) Close() error {
	c.flush()
	if c.storage != nil {
		return c.storage.Close()
	}
	return nil
}

func (c *DataCollector) SetStorage(storage Storage) {
	if storage != nil {
		c.storage = storage
	}
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
