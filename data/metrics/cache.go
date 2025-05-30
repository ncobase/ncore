package metrics

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheMetricsCollector for Redis cache operations
type CacheMetricsCollector interface {
	RedisCommand(command string, err error)
}

// CacheCollector wraps Redis client with metrics
type CacheCollector struct {
	client    *redis.Client
	collector CacheMetricsCollector
}

// NewCacheCollector creates a cache collector wrapper
func NewCacheCollector(client *redis.Client, collector CacheMetricsCollector) *CacheCollector {
	return &CacheCollector{
		client:    client,
		collector: collector,
	}
}

// Get wraps Redis GET with metrics
func (c *CacheCollector) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := c.client.Get(ctx, key)
	if c.collector != nil {
		c.collector.RedisCommand("get", cmd.Err())
	}
	return cmd
}

// Set wraps Redis SET with metrics
func (c *CacheCollector) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	cmd := c.client.Set(ctx, key, value, expiration)
	if c.collector != nil {
		c.collector.RedisCommand("set", cmd.Err())
	}
	return cmd
}

// Del wraps Redis DEL with metrics
func (c *CacheCollector) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	cmd := c.client.Del(ctx, keys...)
	if c.collector != nil {
		c.collector.RedisCommand("del", cmd.Err())
	}
	return cmd
}

// HGet wraps Redis HGET with metrics
func (c *CacheCollector) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	cmd := c.client.HGet(ctx, key, field)
	if c.collector != nil {
		c.collector.RedisCommand("hget", cmd.Err())
	}
	return cmd
}

// HSet wraps Redis HSET with metrics
func (c *CacheCollector) HSet(ctx context.Context, key string, values ...any) *redis.IntCmd {
	cmd := c.client.HSet(ctx, key, values...)
	if c.collector != nil {
		c.collector.RedisCommand("hset", cmd.Err())
	}
	return cmd
}

// HDel wraps Redis HDEL with metrics
func (c *CacheCollector) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	cmd := c.client.HDel(ctx, key, fields...)
	if c.collector != nil {
		c.collector.RedisCommand("hdel", cmd.Err())
	}
	return cmd
}

// Ping wraps Redis PING with metrics
func (c *CacheCollector) Ping(ctx context.Context) *redis.StatusCmd {
	cmd := c.client.Ping(ctx)
	if c.collector != nil {
		c.collector.RedisCommand("ping", cmd.Err())
	}
	return cmd
}

// PoolStats returns pool statistics and updates metrics
func (c *CacheCollector) PoolStats() *redis.PoolStats {
	stats := c.client.PoolStats()
	if c.collector != nil && stats != nil {
		if connCollector, ok := c.collector.(Collector); ok {
			connCollector.RedisConnections(int(stats.TotalConns))
		}
	}
	return stats
}

// GetClient returns the underlying Redis client
func (c *CacheCollector) GetClient() *redis.Client {
	return c.client
}
