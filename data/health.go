package data

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Health checks all components with comprehensive metrics collection
func (d *Data) Health(ctx context.Context) map[string]any {
	health := map[string]any{
		"timestamp": time.Now(),
		"services":  make(map[string]any),
	}

	services := health["services"].(map[string]any)
	overallHealthy := true

	// Database health
	if healthy := d.checkDatabaseHealth(ctx, services); !healthy {
		overallHealthy = false
	}

	// Redis health
	if healthy := d.checkRedisHealth(ctx, services); !healthy {
		overallHealthy = false
	}

	// MongoDB health
	if healthy := d.checkMongoHealth(ctx, services); !healthy {
		overallHealthy = false
	}

	// Messaging health
	if healthy := d.checkMessagingHealth(services); !healthy {
		overallHealthy = false
	}

	// Search engines health
	if healthy := d.checkSearchHealth(ctx, services); !healthy {
		overallHealthy = false
	}

	if overallHealthy {
		health["status"] = "healthy"
	} else {
		health["status"] = "degraded"
	}

	return health
}

// checkDatabaseHealth checks database health
func (d *Data) checkDatabaseHealth(ctx context.Context, services map[string]any) bool {
	if d.Conn == nil || d.Conn.DBM == nil {
		return true // No database configured
	}

	start := time.Now()
	err := d.Conn.Ping(ctx)
	duration := time.Since(start)

	healthy := err == nil
	services["database"] = map[string]any{
		"healthy":     healthy,
		"response_ms": duration.Milliseconds(),
		"error":       getErrorString(err),
	}

	d.collector.DBQuery(duration, err)
	d.collector.HealthCheck("database", healthy)

	return healthy
}

// checkRedisHealth checks Redis health
func (d *Data) checkRedisHealth(_ context.Context, services map[string]any) bool {
	if d.Conn == nil || d.Conn.RC == nil {
		return true
	}

	d.collector.RedisCommand("ping", errors.New("redis client not available"))
	d.collector.HealthCheck("redis", false)

	services["redis"] = map[string]any{
		"healthy":     false,
		"response_ms": int64(0),
		"error":       "redis client not available",
	}

	return false
}

// checkMongoHealth checks MongoDB health
func (d *Data) checkMongoHealth(ctx context.Context, services map[string]any) bool {
	if d.Conn == nil || d.Conn.MGM == nil {
		return true // No MongoDB configured
	}

	manager, ok := d.Conn.MGM.(interface{ Health(context.Context) error })
	if !ok {
		err := errors.New("mongodb manager not available")
		services["mongodb"] = map[string]any{
			"healthy":     false,
			"response_ms": int64(0),
			"error":       err.Error(),
		}
		d.collector.MongoOperation("health_check", err)
		d.collector.HealthCheck("mongodb", false)
		return false
	}

	start := time.Now()
	err := manager.Health(ctx)
	duration := time.Since(start)

	healthy := err == nil
	services["mongodb"] = map[string]any{
		"healthy":     healthy,
		"response_ms": duration.Milliseconds(),
		"error":       getErrorString(err),
	}

	d.collector.MongoOperation("health_check", err)
	d.collector.HealthCheck("mongodb", healthy)

	return healthy
}

// checkMessagingHealth checks messaging systems health
func (d *Data) checkMessagingHealth(services map[string]any) bool {
	overallHealthy := true

	// RabbitMQ health
	if d.Conn != nil && d.Conn.RMQ != nil {
		healthy := false
		if rmq, ok := d.Conn.RMQ.(interface{ IsConnected() bool }); ok {
			healthy = rmq.IsConnected()
		}
		services["rabbitmq"] = map[string]any{
			"healthy": healthy,
			"error":   getErrorString(getConnectivityError(!healthy, "rabbitmq")),
		}
		d.collector.HealthCheck("rabbitmq", healthy)
		if !healthy {
			overallHealthy = false
		}
	}

	// Kafka health
	if d.Conn != nil && d.Conn.KFK != nil {
		healthy := false
		if kfk, ok := d.Conn.KFK.(interface{ IsConnected() bool }); ok {
			healthy = kfk.IsConnected()
		}
		services["kafka"] = map[string]any{
			"healthy": healthy,
			"error":   getErrorString(getConnectivityError(!healthy, "kafka")),
		}
		d.collector.HealthCheck("kafka", healthy)
		if !healthy {
			overallHealthy = false
		}
	}

	return overallHealthy
}

// checkSearchHealth checks search engines health
func (d *Data) checkSearchHealth(_ context.Context, services map[string]any) bool {
	if d.Conn == nil {
		return true
	}

	if d.Conn.ES == nil && d.Conn.OS == nil && d.Conn.MS == nil {
		return true
	}

	services["search"] = map[string]any{
		"healthy": false,
		"error":   "search client not available",
	}
	d.collector.HealthCheck("search", false)

	return false
}

// getErrorString returns error string
func getErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// getConnectivityError returns connectivity error
func getConnectivityError(hasError bool, service string) error {
	if hasError {
		return fmt.Errorf("%s connection not available", service)
	}
	return nil
}
