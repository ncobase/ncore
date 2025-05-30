package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Collector defines interface for data layer metrics collection
type Collector interface {
	// Database metrics

	DBQuery(duration time.Duration, err error)
	DBTransaction(err error)
	DBConnections(count int)

	// Redis metrics

	RedisCommand(command string, err error)
	RedisConnections(count int)

	// MongoDB metrics

	MongoOperation(operation string, err error)

	// Search metrics

	SearchQuery(engine string, err error)
	SearchIndex(engine, operation string)

	// Message queue metrics

	MQPublish(system string, err error)
	MQConsume(system string, err error)

	// Health metrics

	HealthCheck(component string, healthy bool)
}

// NoOpCollector implements Collector with no-op methods
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

// DefaultCollector provides basic metrics collection with atomic counters
type DefaultCollector struct {
	// Database metrics
	dbQueries      atomic.Int64
	dbQueryErrors  atomic.Int64
	dbTransactions atomic.Int64
	dbTxErrors     atomic.Int64
	dbConnections  atomic.Int32
	dbSlowQueries  atomic.Int64

	// Redis metrics
	redisCommands    atomic.Int64
	redisErrors      atomic.Int64
	redisConnections atomic.Int32

	// MongoDB metrics
	mongoOperations atomic.Int64
	mongoErrors     atomic.Int64

	// Search metrics
	searchQueries  atomic.Int64
	searchErrors   atomic.Int64
	searchIndexOps atomic.Int64

	// Message queue metrics
	mqPublished     atomic.Int64
	mqPublishErrors atomic.Int64
	mqConsumed      atomic.Int64
	mqConsumeErrors atomic.Int64

	// Health metrics
	healthChecks map[string]*atomic.Bool
	healthMu     sync.RWMutex

	// Timing metrics
	lastDBQuery      atomic.Value // time.Time
	lastRedisCommand atomic.Value // time.Time
	lastMongoOp      atomic.Value // time.Time
	lastSearchQuery  atomic.Value // time.Time
	lastMQOperation  atomic.Value // time.Time
}

// NewDefaultCollector creates a new default collector
func NewDefaultCollector() *DefaultCollector {
	c := &DefaultCollector{
		healthChecks: make(map[string]*atomic.Bool),
	}

	now := time.Now()
	c.lastDBQuery.Store(now)
	c.lastRedisCommand.Store(now)
	c.lastMongoOp.Store(now)
	c.lastSearchQuery.Store(now)
	c.lastMQOperation.Store(now)

	return c
}

// DBQuery Database query metrics
func (c *DefaultCollector) DBQuery(duration time.Duration, err error) {
	c.dbQueries.Add(1)
	c.lastDBQuery.Store(time.Now())

	if err != nil {
		c.dbQueryErrors.Add(1)
	}

	// Track slow queries (>1 second)
	if duration > time.Second {
		c.dbSlowQueries.Add(1)
	}
}

// DBTransaction Database transaction metrics
func (c *DefaultCollector) DBTransaction(err error) {
	c.dbTransactions.Add(1)

	if err != nil {
		c.dbTxErrors.Add(1)
	}
}

// DBConnections Database connection metrics
func (c *DefaultCollector) DBConnections(count int) {
	c.dbConnections.Store(int32(count))
}

// RedisCommand Redis command metrics
func (c *DefaultCollector) RedisCommand(command string, err error) {
	c.redisCommands.Add(1)
	c.lastRedisCommand.Store(time.Now())

	if err != nil {
		c.redisErrors.Add(1)
	}
}

// RedisConnections Redis connection metrics
func (c *DefaultCollector) RedisConnections(count int) {
	c.redisConnections.Store(int32(count))
}

// MongoOperation MongoDB operation metrics
func (c *DefaultCollector) MongoOperation(operation string, err error) {
	c.mongoOperations.Add(1)
	c.lastMongoOp.Store(time.Now())

	if err != nil {
		c.mongoErrors.Add(1)
	}
}

// SearchQuery Search query metrics
func (c *DefaultCollector) SearchQuery(engine string, err error) {
	c.searchQueries.Add(1)
	c.lastSearchQuery.Store(time.Now())

	if err != nil {
		c.searchErrors.Add(1)
	}
}

// SearchIndex Search index metrics
func (c *DefaultCollector) SearchIndex(engine, operation string) {
	c.searchIndexOps.Add(1)
}

// MQPublish Message queue publish metrics
func (c *DefaultCollector) MQPublish(system string, err error) {
	c.mqPublished.Add(1)
	c.lastMQOperation.Store(time.Now())

	if err != nil {
		c.mqPublishErrors.Add(1)
	}
}

// MQConsume Message queue consume metrics
func (c *DefaultCollector) MQConsume(system string, err error) {
	c.mqConsumed.Add(1)
	c.lastMQOperation.Store(time.Now())

	if err != nil {
		c.mqConsumeErrors.Add(1)
	}
}

// HealthCheck Health check
func (c *DefaultCollector) HealthCheck(component string, healthy bool) {
	c.healthMu.Lock()
	if _, exists := c.healthChecks[component]; !exists {
		c.healthChecks[component] = &atomic.Bool{}
	}
	healthCheck := c.healthChecks[component]
	c.healthMu.Unlock()

	healthCheck.Store(healthy)
}

// GetStats returns comprehensive statistics
func (c *DefaultCollector) GetStats() map[string]any {
	dbQueries := c.dbQueries.Load()
	dbErrors := c.dbQueryErrors.Load()
	dbSuccessRate := calculateSuccessRate(dbQueries, dbErrors)

	redisCommands := c.redisCommands.Load()
	redisErrors := c.redisErrors.Load()
	redisSuccessRate := calculateSuccessRate(redisCommands, redisErrors)

	mongoOps := c.mongoOperations.Load()
	mongoErrors := c.mongoErrors.Load()
	mongoSuccessRate := calculateSuccessRate(mongoOps, mongoErrors)

	searchQueries := c.searchQueries.Load()
	searchErrors := c.searchErrors.Load()
	searchSuccessRate := calculateSuccessRate(searchQueries, searchErrors)

	c.healthMu.RLock()
	healthStatus := make(map[string]bool)
	for component, status := range c.healthChecks {
		healthStatus[component] = status.Load()
	}
	c.healthMu.RUnlock()

	return map[string]any{
		"database": map[string]any{
			"connections":  c.dbConnections.Load(),
			"queries":      dbQueries,
			"errors":       dbErrors,
			"slow_queries": c.dbSlowQueries.Load(),
			"transactions": c.dbTransactions.Load(),
			"tx_errors":    c.dbTxErrors.Load(),
			"success_rate": dbSuccessRate,
			"last_query":   c.lastDBQuery.Load(),
		},
		"redis": map[string]any{
			"connections":  c.redisConnections.Load(),
			"commands":     redisCommands,
			"errors":       redisErrors,
			"success_rate": redisSuccessRate,
			"last_command": c.lastRedisCommand.Load(),
		},
		"mongodb": map[string]any{
			"operations":     mongoOps,
			"errors":         mongoErrors,
			"success_rate":   mongoSuccessRate,
			"last_operation": c.lastMongoOp.Load(),
		},
		"search": map[string]any{
			"queries":      searchQueries,
			"errors":       searchErrors,
			"index_ops":    c.searchIndexOps.Load(),
			"success_rate": searchSuccessRate,
			"last_query":   c.lastSearchQuery.Load(),
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

// calculateSuccessRate calculates success rate percentage
func calculateSuccessRate(total, errors int64) float64 {
	if total == 0 {
		return 100.0
	}
	success := total - errors
	return (float64(success) / float64(total)) * 100.0
}
