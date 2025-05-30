package metrics

import (
	"time"
)

// DataCollector collects data layer specific metrics
type DataCollector struct {
	collector *Collector
}

// NewDataCollector creates a new data collector
func NewDataCollector(collector *Collector) *DataCollector {
	return &DataCollector{collector: collector}
}

// DBQuery Database query metrics
func (dc *DataCollector) DBQuery(duration time.Duration, err error) {
	labels := map[string]string{"operation": "query"}
	if err != nil {
		labels["status"] = "error"
		dc.collector.Inc("database", "queries_total", labels)
		dc.collector.Inc("database", "errors_total", labels)
	} else {
		labels["status"] = "success"
		dc.collector.Inc("database", "queries_total", labels)
	}

	// Record duration
	dc.collector.Observe("database", "query_duration_seconds", duration.Seconds(), nil)

	// Track slow queries
	if duration > time.Second {
		dc.collector.Inc("database", "slow_queries_total", nil)
	}
}

// DBTransaction Database transaction metrics
func (dc *DataCollector) DBTransaction(err error) {
	labels := map[string]string{"operation": "transaction"}
	if err != nil {
		labels["status"] = "error"
		dc.collector.Inc("database", "transactions_total", labels)
		dc.collector.Inc("database", "errors_total", labels)
	} else {
		labels["status"] = "success"
		dc.collector.Inc("database", "transactions_total", labels)
	}
}

// DBConnections Database connections metrics
func (dc *DataCollector) DBConnections(count int) {
	dc.collector.Set("database", "connections_active", float64(count), nil)
}

// RedisCommand Redis commands metrics
func (dc *DataCollector) RedisCommand(command string, err error) {
	labels := map[string]string{"command": command}
	if err != nil {
		labels["status"] = "error"
		dc.collector.Inc("redis", "commands_total", labels)
		dc.collector.Inc("redis", "errors_total", labels)
	} else {
		labels["status"] = "success"
		dc.collector.Inc("redis", "commands_total", labels)
	}
}

// RedisConnections Redis connections metrics
func (dc *DataCollector) RedisConnections(count int) {
	dc.collector.Set("redis", "connections_active", float64(count), nil)
}

// MongoOperation MongoDB operation metrics
func (dc *DataCollector) MongoOperation(operation string, err error) {
	labels := map[string]string{"operation": operation}
	if err != nil {
		labels["status"] = "error"
		dc.collector.Inc("mongodb", "operations_total", labels)
		dc.collector.Inc("mongodb", "errors_total", labels)
	} else {
		labels["status"] = "success"
		dc.collector.Inc("mongodb", "operations_total", labels)
	}
}

// SearchQuery Search query metrics
func (dc *DataCollector) SearchQuery(engine string, err error) {
	labels := map[string]string{"engine": engine}
	if err != nil {
		labels["status"] = "error"
		dc.collector.Inc("search", "queries_total", labels)
		dc.collector.Inc("search", "errors_total", labels)
	} else {
		labels["status"] = "success"
		dc.collector.Inc("search", "queries_total", labels)
	}
}

// SearchIndex Search index metrics
func (dc *DataCollector) SearchIndex(engine, operation string) {
	labels := map[string]string{"engine": engine, "operation": operation}
	dc.collector.Inc("search", "index_operations_total", labels)
}

// MQPublish Message queue metrics
func (dc *DataCollector) MQPublish(system string, err error) {
	labels := map[string]string{"system": system}
	if err != nil {
		labels["status"] = "error"
		dc.collector.Inc("messaging", "published_total", labels)
		dc.collector.Inc("messaging", "errors_total", labels)
	} else {
		labels["status"] = "success"
		dc.collector.Inc("messaging", "published_total", labels)
	}
}

// MQConsume Message queue metrics
func (dc *DataCollector) MQConsume(system string, err error) {
	labels := map[string]string{"system": system}
	if err != nil {
		labels["status"] = "error"
		dc.collector.Inc("messaging", "consumed_total", labels)
		dc.collector.Inc("messaging", "errors_total", labels)
	} else {
		labels["status"] = "success"
		dc.collector.Inc("messaging", "consumed_total", labels)
	}
}

// HealthCheck Health checks
func (dc *DataCollector) HealthCheck(component string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}

	labels := map[string]string{"component": component}
	dc.collector.Set("health", "status", value, labels)
}
