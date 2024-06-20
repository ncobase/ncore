package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// ICache defines a general caching interface
type ICache[T any] interface {
	Get(context.Context, string) (*T, error)
	Set(context.Context, string, *T, ...time.Duration) error
	Delete(context.Context, string) error
	GetArray(context.Context, string, any) error
	SetArray(context.Context, string, any, ...time.Duration) error
}

// Cache implements the ICache interface
type Cache[T any] struct {
	rc      *redis.Client
	key     string
	useHash bool
}

// Key defines the cache key
func Key(key string) string {
	return fmt.Sprintf("%s", key)
}

// NewCache creates a new Cache instance
func NewCache[T any](rc *redis.Client, key string, useHash ...bool) *Cache[T] {
	hash := false
	if len(useHash) > 0 {
		hash = useHash[0]
	}
	return &Cache[T]{rc: rc, key: key, useHash: hash}
}

// Get retrieves a single item from cache
func (c *Cache[T]) Get(ctx context.Context, field string) (*T, error) {
	if c.rc == nil {
		log.Printf("redis client is nil, skipping Get operation")
		return nil, nil
	}

	var result string
	var err error

	if c.useHash {
		result, err = c.rc.HGet(ctx, c.key, field).Result()
	} else {
		result, err = c.rc.Get(ctx, field).Result()
	}

	if err != nil {
		// if errors.Is(err, redis.Nil) {
		// 	return nil, nil // Cache miss
		// }
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	var row T
	if err = json.Unmarshal([]byte(result), &row); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}
	return &row, nil
}

// Set saves a single item into cache
func (c *Cache[T]) Set(ctx context.Context, field string, data *T, expire ...time.Duration) error {
	if c.rc == nil {
		log.Printf("redis client is nil, skipping Set operation")
		return nil
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("failed to marshal data for cache set: %v, error: %v", data, err)
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if c.useHash {
		err = c.rc.HSet(ctx, c.key, field, bytes).Err()
	} else {
		exp := time.Duration(0)
		if len(expire) > 0 {
			exp = expire[0]
		}
		err = c.rc.Set(ctx, field, bytes, exp).Err()
	}

	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}
	return nil
}

// GetArray retrieves an array of items from cache
func (c *Cache[T]) GetArray(ctx context.Context, field string, dest any) error {
	if c.rc == nil {
		log.Printf("redis client is nil, skipping GetArray operation")
		return nil
	}

	var result string
	var err error

	if c.useHash {
		result, err = c.rc.HGet(ctx, c.key, field).Result()
	} else {
		result, err = c.rc.Get(ctx, field).Result()
	}

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return fmt.Errorf("cache miss for key %s", field)
		}
		return fmt.Errorf("failed to get cache: %w", err)
	}

	if err := json.Unmarshal([]byte(result), &dest); err != nil {
		return fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return nil
}

// SetArray saves an array of items into cache
func (c *Cache[T]) SetArray(ctx context.Context, field string, data any, expire ...time.Duration) error {
	if c.rc == nil {
		log.Printf("redis client is nil, skipping SetArray operation")
		return nil
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("failed to marshal data for cache set: %v, error: %v", data, err)
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if c.useHash {
		err = c.rc.HSet(ctx, c.key, field, bytes).Err()
	} else {
		exp := time.Duration(0)
		if len(expire) > 0 {
			exp = expire[0]
		}
		err = c.rc.Set(ctx, field, bytes, exp).Err()
	}

	if err != nil {
		log.Printf("failed to set cache: %v, error: %v", data, err)
		return fmt.Errorf("failed to set cache: %w", err)
	}
	return nil
}

// Delete removes data from cache
func (c *Cache[T]) Delete(ctx context.Context, field string) error {
	if c.rc == nil {
		log.Printf("redis client is nil, skipping Delete operation")
		return nil
	}

	var err error

	if c.useHash {
		err = c.rc.HDel(ctx, c.key, field).Err()
	} else {
		err = c.rc.Del(ctx, field).Err()
	}

	if err != nil {
		log.Printf("failed to delete cache field: %s, error: %v", field, err)
		return fmt.Errorf("failed to delete cache: %w", err)
	}
	return nil
}
