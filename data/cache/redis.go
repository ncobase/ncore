package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ncobase/ncore/data/metrics"
	"github.com/redis/go-redis/v9"
)

// ICache defines a general caching interface
type ICache[T any] interface {
	Get(context.Context, string) (*T, error)
	Set(context.Context, string, *T, ...time.Duration) error
	Delete(context.Context, string) error
	GetArray(context.Context, string, any) error
	SetArray(context.Context, string, any, ...time.Duration) error
	GetMultiple(context.Context, []string) (map[string]*T, error)
	SetMultiple(context.Context, map[string]*T, ...time.Duration) error
	Exists(context.Context, string) (bool, error)
	TTL(context.Context, string) (time.Duration, error)
	Expire(context.Context, string, time.Duration) error
}

// Cache implements the ICache interface
type Cache[T any] struct {
	rc        *redis.Client
	key       string
	useHash   bool
	collector metrics.CacheMetricsCollector
}

// Key defines the cache key
func (c *Cache[T]) Key(field string) string {
	if c.useHash {
		return field
	}
	if c.key != "" {
		return fmt.Sprintf("%s:%s", c.key, field)
	}
	return field
}

// NewCache creates a new Cache instance
func NewCache[T any](rc *redis.Client, key string, useHash ...bool) *Cache[T] {
	hash := false
	if len(useHash) > 0 {
		hash = useHash[0]
	}
	return &Cache[T]{
		rc:        rc,
		key:       key,
		useHash:   hash,
		collector: metrics.NoOpCollector{},
	}
}

// NewCacheWithMetrics creates a new Cache instance collector
func NewCacheWithMetrics[T any](rc *redis.Client, key string, collector metrics.CacheMetricsCollector, useHash ...bool) *Cache[T] {
	cache := NewCache[T](rc, key, useHash...)
	if collector != nil {
		cache.collector = collector
	}
	return cache
}

// Get retrieves a single item from cache
func (c *Cache[T]) Get(ctx context.Context, field string) (*T, error) {
	if c.rc == nil {
		err := errors.New("redis client is nil, cannot get cache")
		c.collector.RedisCommand("get", err)
		return nil, err
	}

	var result string
	var err error
	var command string

	if c.useHash {
		command = "hget"
		result, err = c.rc.HGet(ctx, c.Key(field), field).Result()
	} else {
		command = "get"
		result, err = c.rc.Get(ctx, c.Key(field)).Result()
	}

	c.collector.RedisCommand(command, err)

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	var row T
	if err = json.Unmarshal([]byte(result), &row); err != nil {
		c.collector.RedisCommand("unmarshal", err)
		return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}
	return &row, nil
}

// Set saves a single item into cache
func (c *Cache[T]) Set(ctx context.Context, field string, data *T, expire ...time.Duration) error {
	if c.rc == nil {
		err := errors.New("redis client is nil, cannot set cache")
		c.collector.RedisCommand("set", err)
		return err
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		c.collector.RedisCommand("marshal", err)
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	var command string
	if c.useHash {
		command = "hset"
		err = c.rc.HSet(ctx, c.Key(field), field, bytes).Err()
	} else {
		command = "set"
		exp := time.Duration(0)
		if len(expire) > 0 {
			exp = expire[0]
		}
		err = c.rc.Set(ctx, c.Key(field), bytes, exp).Err()
	}

	c.collector.RedisCommand(command, err)

	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}
	return nil
}

// GetArray retrieves an array of items from cache
func (c *Cache[T]) GetArray(ctx context.Context, field string, dest any) error {
	if c.rc == nil {
		err := errors.New("redis client is nil, cannot get array cache")
		c.collector.RedisCommand("get_array", err)
		return err
	}

	var result string
	var err error
	var command string

	if c.useHash {
		command = "hget"
		result, err = c.rc.HGet(ctx, c.Key(field), field).Result()
	} else {
		command = "get"
		result, err = c.rc.Get(ctx, c.Key(field)).Result()
	}

	c.collector.RedisCommand(command, err)

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil // Cache miss
		}
		return fmt.Errorf("failed to get array cache: %w", err)
	}

	if err = json.Unmarshal([]byte(result), dest); err != nil {
		c.collector.RedisCommand("unmarshal_array", err)
		return fmt.Errorf("failed to unmarshal array cache data: %w", err)
	}

	return nil
}

// SetArray saves an array of items into cache
func (c *Cache[T]) SetArray(ctx context.Context, field string, data any, expire ...time.Duration) error {
	if c.rc == nil {
		err := errors.New("redis client is nil, cannot set array cache")
		c.collector.RedisCommand("set_array", err)
		return err
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		c.collector.RedisCommand("marshal_array", err)
		return fmt.Errorf("failed to marshal array data: %w", err)
	}

	var command string
	if c.useHash {
		command = "hset"
		err = c.rc.HSet(ctx, c.Key(field), field, bytes).Err()
	} else {
		command = "set"
		exp := time.Duration(0)
		if len(expire) > 0 {
			exp = expire[0]
		}
		err = c.rc.Set(ctx, c.Key(field), bytes, exp).Err()
	}

	c.collector.RedisCommand(command, err)

	if err != nil {
		return fmt.Errorf("failed to set array cache: %w", err)
	}
	return nil
}

// Delete removes data from cache
func (c *Cache[T]) Delete(ctx context.Context, field string) error {
	if c.rc == nil {
		err := errors.New("redis client is nil, cannot delete cache")
		c.collector.RedisCommand("delete", err)
		return err
	}

	var err error
	var command string

	if c.useHash {
		command = "hdel"
		err = c.rc.HDel(ctx, c.Key(field), field).Err()
	} else {
		command = "del"
		err = c.rc.Del(ctx, c.Key(field)).Err()
	}

	c.collector.RedisCommand(command, err)

	if err != nil {
		log.Printf("failed to delete cache field: %s, error: %v", field, err)
		return fmt.Errorf("failed to delete cache: %w", err)
	}
	return nil
}

// GetMultiple retrieves multiple items from cache
func (c *Cache[T]) GetMultiple(ctx context.Context, fields []string) (map[string]*T, error) {
	if c.rc == nil {
		err := errors.New("redis client is nil, cannot get multiple cache")
		c.collector.RedisCommand("get_multiple", err)
		return nil, err
	}

	result := make(map[string]*T)

	if c.useHash {
		// Use HMGET for hash-based cache
		keys := make([]string, len(fields))
		copy(keys, fields)

		// Get the first field's key for the hash
		hashKey := c.Key(fields[0])
		values, err := c.rc.HMGet(ctx, hashKey, keys...).Result()
		c.collector.RedisCommand("hmget", err)

		if err != nil {
			return nil, fmt.Errorf("failed to get multiple hash cache: %w", err)
		}

		for i, val := range values {
			if val != nil {
				if strVal, ok := val.(string); ok && strVal != "" {
					var item T
					if err := json.Unmarshal([]byte(strVal), &item); err == nil {
						result[fields[i]] = &item
					}
				}
			}
		}
	} else {
		// Use MGET for key-based cache
		keys := make([]string, len(fields))
		for i, field := range fields {
			keys[i] = c.Key(field)
		}

		values, err := c.rc.MGet(ctx, keys...).Result()
		c.collector.RedisCommand("mget", err)

		if err != nil {
			return nil, fmt.Errorf("failed to get multiple cache: %w", err)
		}

		for i, val := range values {
			if val != nil {
				if strVal, ok := val.(string); ok && strVal != "" {
					var item T
					if err := json.Unmarshal([]byte(strVal), &item); err == nil {
						result[fields[i]] = &item
					}
				}
			}
		}
	}

	return result, nil
}

// SetMultiple saves multiple items into cache
func (c *Cache[T]) SetMultiple(ctx context.Context, items map[string]*T, expire ...time.Duration) error {
	if c.rc == nil {
		err := errors.New("redis client is nil, cannot set multiple cache")
		c.collector.RedisCommand("set_multiple", err)
		return err
	}

	if c.useHash {
		// Use HMSET for hash-based cache
		if len(items) == 0 {
			return nil
		}

		// Get the first key for the hash
		var hashKey string
		values := make(map[string]any)

		for field, data := range items {
			if hashKey == "" {
				hashKey = c.Key(field)
			}

			bytes, err := json.Marshal(data)
			if err != nil {
				c.collector.RedisCommand("marshal_multiple", err)
				return fmt.Errorf("failed to marshal data for field %s: %w", field, err)
			}
			values[field] = bytes
		}

		err := c.rc.HMSet(ctx, hashKey, values).Err()
		c.collector.RedisCommand("hmset", err)

		if err != nil {
			return fmt.Errorf("failed to set multiple hash cache: %w", err)
		}
	} else {
		// Use pipeline for key-based cache
		pipe := c.rc.Pipeline()
		exp := time.Duration(0)
		if len(expire) > 0 {
			exp = expire[0]
		}

		for field, data := range items {
			bytes, err := json.Marshal(data)
			if err != nil {
				c.collector.RedisCommand("marshal_multiple", err)
				return fmt.Errorf("failed to marshal data for field %s: %w", field, err)
			}
			pipe.Set(ctx, c.Key(field), bytes, exp)
		}

		_, err := pipe.Exec(ctx)
		c.collector.RedisCommand("pipeline_set", err)

		if err != nil {
			return fmt.Errorf("failed to set multiple cache: %w", err)
		}
	}

	return nil
}

// Exists checks if cache key exists
func (c *Cache[T]) Exists(ctx context.Context, field string) (bool, error) {
	if c.rc == nil {
		err := errors.New("redis client is nil, cannot check existence")
		c.collector.RedisCommand("exists", err)
		return false, err
	}

	var result bool
	var err error
	var command string

	if c.useHash {
		command = "hexists"
		result, err = c.rc.HExists(ctx, c.Key(field), field).Result()
	} else {
		command = "exists"
		count, existsErr := c.rc.Exists(ctx, c.Key(field)).Result()
		result = count > 0
		err = existsErr
	}

	c.collector.RedisCommand(command, err)

	if err != nil {
		return false, fmt.Errorf("failed to check cache existence: %w", err)
	}

	return result, nil
}

// TTL gets the time to live for a cache key
func (c *Cache[T]) TTL(ctx context.Context, field string) (time.Duration, error) {
	if c.rc == nil {
		err := errors.New("redis client is nil, cannot get TTL")
		c.collector.RedisCommand("ttl", err)
		return 0, err
	}

	duration, err := c.rc.TTL(ctx, c.Key(field)).Result()
	c.collector.RedisCommand("ttl", err)

	if err != nil {
		return 0, fmt.Errorf("failed to get cache TTL: %w", err)
	}

	return duration, nil
}

// Expire sets expiration for a cache key
func (c *Cache[T]) Expire(ctx context.Context, field string, expiration time.Duration) error {
	if c.rc == nil {
		err := errors.New("redis client is nil, cannot set expiration")
		c.collector.RedisCommand("expire", err)
		return err
	}

	err := c.rc.Expire(ctx, c.Key(field), expiration).Err()
	c.collector.RedisCommand("expire", err)

	if err != nil {
		return fmt.Errorf("failed to set cache expiration: %w", err)
	}

	return nil
}
