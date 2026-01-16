package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/data/config"
	"github.com/ncobase/ncore/data/metrics"
	"github.com/redis/go-redis/v9"
)

type driver struct{}

func (d *driver) Name() string {
	return "redis"
}

func (d *driver) Connect(ctx context.Context, cfg any) (any, error) {
	redisCfg, ok := cfg.(*config.Redis)
	if !ok {
		return nil, fmt.Errorf("redis: invalid configuration type, expected *config.Redis")
	}

	if redisCfg.Addr == "" {
		return nil, fmt.Errorf("redis: address is empty")
	}

	client := redis.NewClient(&redis.Options{
		Addr:         redisCfg.Addr,
		Username:     redisCfg.Username,
		Password:     redisCfg.Password,
		DB:           redisCfg.Db,
		ReadTimeout:  redisCfg.ReadTimeout,
		WriteTimeout: redisCfg.WriteTimeout,
		DialTimeout:  redisCfg.DialTimeout,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("redis: failed to ping server: %w", err)
	}

	return client, nil
}

func (d *driver) Close(conn any) error {
	client, ok := conn.(*redis.Client)
	if !ok {
		return fmt.Errorf("redis: invalid connection type, expected *redis.Client")
	}

	if err := client.Close(); err != nil {
		return fmt.Errorf("redis: failed to close connection: %w", err)
	}

	return nil
}

func (d *driver) Ping(ctx context.Context, conn any) error {
	client, ok := conn.(*redis.Client)
	if !ok {
		return fmt.Errorf("redis: invalid connection type, expected *redis.Client")
	}

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis: ping failed: %w", err)
	}

	return nil
}

type CacheCollector struct {
	client    *redis.Client
	collector metrics.CacheMetricsCollector
}

func NewCacheCollector(client *redis.Client, collector metrics.CacheMetricsCollector) *CacheCollector {
	return &CacheCollector{
		client:    client,
		collector: collector,
	}
}

func (c *CacheCollector) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := c.client.Get(ctx, key)
	if c.collector != nil {
		c.collector.RedisCommand("get", cmd.Err())
	}
	return cmd
}

func (c *CacheCollector) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	cmd := c.client.Set(ctx, key, value, expiration)
	if c.collector != nil {
		c.collector.RedisCommand("set", cmd.Err())
	}
	return cmd
}

func (c *CacheCollector) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	cmd := c.client.Del(ctx, keys...)
	if c.collector != nil {
		c.collector.RedisCommand("del", cmd.Err())
	}
	return cmd
}

func (c *CacheCollector) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	cmd := c.client.HGet(ctx, key, field)
	if c.collector != nil {
		c.collector.RedisCommand("hget", cmd.Err())
	}
	return cmd
}

func (c *CacheCollector) HSet(ctx context.Context, key string, values ...any) *redis.IntCmd {
	cmd := c.client.HSet(ctx, key, values...)
	if c.collector != nil {
		c.collector.RedisCommand("hset", cmd.Err())
	}
	return cmd
}

func (c *CacheCollector) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	cmd := c.client.HDel(ctx, key, fields...)
	if c.collector != nil {
		c.collector.RedisCommand("hdel", cmd.Err())
	}
	return cmd
}

func (c *CacheCollector) Ping(ctx context.Context) *redis.StatusCmd {
	cmd := c.client.Ping(ctx)
	if c.collector != nil {
		c.collector.RedisCommand("ping", cmd.Err())
	}
	return cmd
}

func (c *CacheCollector) PoolStats() *redis.PoolStats {
	stats := c.client.PoolStats()
	if c.collector != nil && stats != nil {
		if connCollector, ok := c.collector.(metrics.Collector); ok {
			connCollector.RedisConnections(int(stats.TotalConns))
		}
	}
	return stats
}

func (c *CacheCollector) GetClient() *redis.Client {
	return c.client
}

type RedisStorage struct {
	client    *redis.Client
	keyPrefix string
	retention time.Duration
}

func NewRedisStorage(client *redis.Client, keyPrefix string, retention time.Duration) *RedisStorage {
	return &RedisStorage{
		client:    client,
		keyPrefix: keyPrefix,
		retention: retention,
	}
}

func (r *RedisStorage) Store(metricsData []metrics.Metric) error {
	ctx := context.Background()
	pipe := r.client.Pipeline()

	for _, metric := range metricsData {
		key := fmt.Sprintf("%s:metrics:%s", r.keyPrefix, metric.Type)
		data, err := json.Marshal(metric)
		if err != nil {
			continue
		}

		score := float64(metric.Timestamp.Unix())
		pipe.ZAdd(ctx, key, redis.Z{Score: score, Member: string(data)})
		pipe.Expire(ctx, key, r.retention)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisStorage) Query(query metrics.QueryRequest) ([]metrics.Metric, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:metrics:%s", r.keyPrefix, query.Type)

	minValue := "-inf"
	if !query.StartTime.IsZero() {
		minValue = fmt.Sprintf("%d", query.StartTime.Unix())
	}
	maxValue := "+inf"
	if !query.EndTime.IsZero() {
		maxValue = fmt.Sprintf("%d", query.EndTime.Unix())
	}

	members, err := r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: minValue,
		Max: maxValue,
	}).Result()
	if err != nil {
		return nil, err
	}

	result := make([]metrics.Metric, 0, len(members))
	for _, member := range members {
		var metric metrics.Metric
		if err := json.Unmarshal([]byte(member), &metric); err != nil {
			continue
		}
		if matchesLabels(metric, query.Labels) {
			result = append(result, metric)
		}
	}

	if query.Limit > 0 && len(result) > query.Limit {
		result = result[:query.Limit]
	}

	return result, nil
}

func (r *RedisStorage) Close() error {
	return nil
}

type RedisDataCollector struct {
	base *metrics.DataCollector
}

func NewDataCollectorWithRedis(client *redis.Client, keyPrefix string, retention time.Duration, batchSize int) *RedisDataCollector {
	if batchSize <= 0 {
		batchSize = 100
	}

	collector := metrics.NewDataCollector(batchSize)
	collector.SetStorage(NewRedisStorage(client, keyPrefix, retention))
	return &RedisDataCollector{base: collector}
}

func (c *RedisDataCollector) DBQuery(duration time.Duration, err error) {
	c.base.DBQuery(duration, err)
}

func (c *RedisDataCollector) DBTransaction(err error) {
	c.base.DBTransaction(err)
}

func (c *RedisDataCollector) DBConnections(count int) {
	c.base.DBConnections(count)
}

func (c *RedisDataCollector) RedisCommand(command string, err error) {
	c.base.RedisCommand(command, err)
}

func (c *RedisDataCollector) RedisConnections(count int) {
	c.base.RedisConnections(count)
}

func (c *RedisDataCollector) MongoOperation(operation string, err error) {
	c.base.MongoOperation(operation, err)
}

func (c *RedisDataCollector) SearchQuery(engine string, err error) {
	c.base.SearchQuery(engine, err)
}

func (c *RedisDataCollector) SearchIndex(engine, operation string) {
	c.base.SearchIndex(engine, operation)
}

func (c *RedisDataCollector) MQPublish(system string, err error) {
	c.base.MQPublish(system, err)
}

func (c *RedisDataCollector) MQConsume(system string, err error) {
	c.base.MQConsume(system, err)
}

func (c *RedisDataCollector) HealthCheck(component string, healthy bool) {
	c.base.HealthCheck(component, healthy)
}

func (c *RedisDataCollector) GetStats() map[string]any {
	return c.base.GetStats()
}

func (c *RedisDataCollector) Close() error {
	return c.base.Close()
}

func matchesLabels(metric metrics.Metric, labels metrics.Labels) bool {
	for key, value := range labels {
		if metric.Labels == nil || metric.Labels[key] != value {
			return false
		}
	}
	return true
}

func init() {
	data.RegisterCacheDriver(&driver{})
}
