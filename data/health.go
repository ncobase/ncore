package data

import (
	"context"
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
func (d *Data) checkRedisHealth(ctx context.Context, services map[string]any) bool {
	redis := d.GetRedis()
	if redis == nil {
		return true // No Redis configured
	}

	start := time.Now()
	err := redis.Ping(ctx).Err()
	duration := time.Since(start)

	healthy := err == nil
	services["redis"] = map[string]any{
		"healthy":     healthy,
		"response_ms": duration.Milliseconds(),
		"error":       getErrorString(err),
	}

	d.collector.RedisCommand("ping", err)
	d.collector.HealthCheck("redis", healthy)

	return healthy
}

// checkMongoHealth checks MongoDB health
func (d *Data) checkMongoHealth(ctx context.Context, services map[string]any) bool {
	if d.Conn == nil || d.Conn.MGM == nil {
		return true // No MongoDB configured
	}

	start := time.Now()
	err := d.Conn.MGM.Health(ctx)
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
	if d.RabbitMQ != nil {
		healthy := d.RabbitMQ.IsConnected()
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
	if d.Kafka != nil {
		healthy := d.Kafka.IsConnected()
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
func (d *Data) checkSearchHealth(ctx context.Context, services map[string]any) bool {
	overallHealthy := true

	// Get search health if client is available
	searchHealth := d.SearchHealth(ctx)
	if searchHealth != nil {
		for engine, err := range searchHealth {
			healthy := err == nil
			services[string(engine)] = map[string]any{
				"healthy": healthy,
				"error":   getErrorString(err),
			}
			d.collector.HealthCheck(string(engine), healthy)
			if !healthy {
				overallHealthy = false
			}
		}
	}

	return overallHealthy
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
