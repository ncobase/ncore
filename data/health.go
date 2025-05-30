package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ncobase/ncore/data/messaging/kafka"
	"github.com/ncobase/ncore/data/messaging/rabbitmq"
	"github.com/ncobase/ncore/data/metrics"
	"github.com/ncobase/ncore/data/search/elastic"
	"github.com/ncobase/ncore/data/search/meili"
	"github.com/ncobase/ncore/data/search/opensearch"
)

// Health checkers implementation

type DatabaseHealthChecker struct {
	data *Data
}

func (d *DatabaseHealthChecker) Name() string {
	return "database"
}

func (d *DatabaseHealthChecker) Check(ctx context.Context) error {
	if d.data.Conn != nil {
		return d.data.Conn.Ping(ctx)
	}
	return errors.New("database connection not available")
}

type RedisHealthChecker struct {
	collector *metrics.CacheCollector
}

func (r *RedisHealthChecker) Name() string {
	return "redis"
}

func (r *RedisHealthChecker) Check(ctx context.Context) error {
	if r.collector != nil {
		return r.collector.Ping(ctx).Err()
	}
	return errors.New("redis collector not available")
}

type MongoHealthChecker struct {
	data *Data
}

func (m *MongoHealthChecker) Name() string {
	return "mongodb"
}

func (m *MongoHealthChecker) Check(ctx context.Context) error {
	if m.data.Conn != nil && m.data.Conn.MGM != nil {
		return m.data.Conn.MGM.Health(ctx)
	}
	return errors.New("mongodb connection not available")
}

type RabbitMQHealthChecker struct {
	rabbitmq *rabbitmq.RabbitMQ
}

func (r *RabbitMQHealthChecker) Name() string {
	return "rabbitmq"
}

func (r *RabbitMQHealthChecker) Check(_ context.Context) error {
	if r.rabbitmq != nil && r.rabbitmq.IsConnected() {
		return nil
	}
	return errors.New("rabbitmq connection not available")
}

type KafkaHealthChecker struct {
	kafka *kafka.Kafka
}

func (k *KafkaHealthChecker) Name() string {
	return "kafka"
}

func (k *KafkaHealthChecker) Check(_ context.Context) error {
	if k.kafka != nil && k.kafka.IsConnected() {
		return nil
	}
	return errors.New("kafka connection not available")
}

type ElasticsearchHealthChecker struct {
	client *elastic.Client
}

func (e *ElasticsearchHealthChecker) Name() string {
	return "elasticsearch"
}

func (e *ElasticsearchHealthChecker) Check(_ context.Context) error {
	if e.client == nil {
		return errors.New("elasticsearch client not available")
	}

	client := e.client.GetClient()
	if client == nil {
		return errors.New("elasticsearch client is nil")
	}

	res, err := client.Info()
	if err != nil {
		return fmt.Errorf("elasticsearch info request failed: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch returned error: %s", res.Status())
	}

	return nil
}

type OpenSearchHealthChecker struct {
	client *opensearch.Client
}

func (o *OpenSearchHealthChecker) Name() string {
	return "opensearch"
}

func (o *OpenSearchHealthChecker) Check(ctx context.Context) error {
	if o.client == nil {
		return errors.New("opensearch client not available")
	}

	_, err := o.client.Health(ctx)
	if err != nil {
		return fmt.Errorf("opensearch health check failed: %v", err)
	}

	return nil
}

type MeilisearchHealthChecker struct {
	client *meili.Client
}

func (m *MeilisearchHealthChecker) Name() string {
	return "meilisearch"
}

func (m *MeilisearchHealthChecker) Check(_ context.Context) error {
	if m.client == nil {
		return errors.New("meilisearch client not available")
	}

	client := m.client.GetClient()
	if client == nil {
		return errors.New("meilisearch client is nil")
	}

	_, err := client.Health()
	if err != nil {
		return fmt.Errorf("meilisearch health check failed: %v", err)
	}

	return nil
}

// Health checks with comprehensive metrics collection
func (d *Data) Health(ctx context.Context) map[string]any {
	health := map[string]any{
		"timestamp": time.Now(),
		"services":  make(map[string]any),
	}

	services := health["services"].(map[string]any)
	overallHealthy := true

	// Use health monitor if available
	if d.healthMonitor != nil {
		healthResults := d.healthMonitor.CheckAll(ctx)

		for component, healthy := range healthResults {
			services[component] = map[string]any{
				"healthy": healthy,
			}

			if !healthy {
				overallHealthy = false
			}
		}
	} else {
		// Fallback to manual health checks
		overallHealthy = d.performManualHealthChecks(ctx, services)
	}

	if overallHealthy {
		health["status"] = "healthy"
	} else {
		health["status"] = "degraded"
	}

	return health
}

// performManualHealthChecks performs health checks without health monitor
func (d *Data) performManualHealthChecks(ctx context.Context, services map[string]any) bool {
	overallHealthy := true

	// Database health
	if d.Conn != nil && d.Conn.DBM != nil {
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

		if !healthy {
			overallHealthy = false
		}
	}

	// Redis health
	if d.redisCollector != nil {
		start := time.Now()
		err := d.redisCollector.Ping(ctx).Err()
		duration := time.Since(start)

		healthy := err == nil
		services["redis"] = map[string]any{
			"healthy":     healthy,
			"response_ms": duration.Milliseconds(),
			"error":       getErrorString(err),
		}

		d.collector.HealthCheck("redis", healthy)

		if !healthy {
			overallHealthy = false
		}
	}

	// MongoDB health
	if d.Conn != nil && d.Conn.MGM != nil {
		start := time.Now()
		err := d.Conn.MGM.Health(ctx)
		duration := time.Since(start)

		healthy := err == nil
		services["mongodb"] = map[string]any{
			"healthy":     healthy,
			"response_ms": duration.Milliseconds(),
			"error":       getErrorString(err),
		}

		d.collector.HealthCheck("mongodb", healthy)

		if !healthy {
			overallHealthy = false
		}
	}

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

	// Search engines health
	d.checkSearchEnginesHealth(ctx, services, &overallHealthy)

	return overallHealthy
}

// checkSearchEnginesHealth checks health of search engines
func (d *Data) checkSearchEnginesHealth(ctx context.Context, services map[string]any, overallHealthy *bool) {
	// Elasticsearch health
	if d.Conn != nil && d.Conn.ES != nil {
		start := time.Now()
		client := d.Conn.ES.GetClient()
		var err error

		if client != nil {
			res, esErr := client.Info()
			if esErr != nil {
				err = esErr
			} else {
				defer res.Body.Close()
				if res.IsError() {
					err = fmt.Errorf("elasticsearch error: %s", res.Status())
				}
			}
		} else {
			err = errors.New("elasticsearch client is nil")
		}

		duration := time.Since(start)
		healthy := err == nil

		services["elasticsearch"] = map[string]any{
			"healthy":     healthy,
			"response_ms": duration.Milliseconds(),
			"error":       getErrorString(err),
		}

		d.collector.HealthCheck("elasticsearch", healthy)

		if !healthy {
			*overallHealthy = false
		}
	}

	// OpenSearch health
	if d.Conn != nil && d.Conn.OS != nil {
		start := time.Now()
		_, err := d.Conn.OS.Health(ctx)
		duration := time.Since(start)

		healthy := err == nil
		services["opensearch"] = map[string]any{
			"healthy":     healthy,
			"response_ms": duration.Milliseconds(),
			"error":       getErrorString(err),
		}

		d.collector.HealthCheck("opensearch", healthy)

		if !healthy {
			*overallHealthy = false
		}
	}

	// Meilisearch health
	if d.Conn != nil && d.Conn.MS != nil {
		start := time.Now()
		client := d.Conn.MS.GetClient()
		var err error

		if client != nil {
			_, msErr := client.Health()
			err = msErr
		} else {
			err = errors.New("meilisearch client is nil")
		}

		duration := time.Since(start)
		healthy := err == nil

		services["meilisearch"] = map[string]any{
			"healthy":     healthy,
			"response_ms": duration.Milliseconds(),
			"error":       getErrorString(err),
		}

		d.collector.HealthCheck("meilisearch", healthy)

		if !healthy {
			*overallHealthy = false
		}
	}
}

func getErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func getConnectivityError(hasError bool, service string) error {
	if hasError {
		return fmt.Errorf("%s connection not available", service)
	}
	return nil
}
