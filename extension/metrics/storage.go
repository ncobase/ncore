package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// MemoryStorage stores metrics in memory (for testing/development)
type MemoryStorage struct {
	data map[string][]*MetricSnapshot
	mu   sync.RWMutex
}

// NewMemoryStorage creates a new memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string][]*MetricSnapshot),
	}
}

func (m *MemoryStorage) Store(collection string, snapshot *MetricSnapshot) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[collection] = append(m.data[collection], snapshot)
	return nil
}

func (m *MemoryStorage) Query(collection string, start, end time.Time) ([]*MetricSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshots, exists := m.data[collection]
	if !exists {
		return []*MetricSnapshot{}, nil
	}

	var result []*MetricSnapshot
	for _, snapshot := range snapshots {
		if snapshot.Timestamp.After(start) && snapshot.Timestamp.Before(end) {
			result = append(result, snapshot)
		}
	}

	return result, nil
}

func (m *MemoryStorage) QueryLatest(collection string, limit int) ([]*MetricSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshots, exists := m.data[collection]
	if !exists {
		return []*MetricSnapshot{}, nil
	}

	start := len(snapshots) - limit
	if start < 0 {
		start = 0
	}

	return snapshots[start:], nil
}

func (m *MemoryStorage) Cleanup(before time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for collection, snapshots := range m.data {
		var filtered []*MetricSnapshot
		for _, snapshot := range snapshots {
			if snapshot.Timestamp.After(before) {
				filtered = append(filtered, snapshot)
			}
		}
		m.data[collection] = filtered
	}

	return nil
}

func (m *MemoryStorage) GetStats() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := 0
	collections := make(map[string]int)

	for collection, snapshots := range m.data {
		count := len(snapshots)
		collections[collection] = count
		total += count
	}

	return map[string]any{
		"type":        "memory",
		"total":       total,
		"collections": collections,
	}
}

// RedisStorage stores metrics in Redis
type RedisStorage struct {
	client    *redis.Client
	keyPrefix string
	retention time.Duration
}

// NewRedisStorage creates a new Redis storage
func NewRedisStorage(client *redis.Client, keyPrefix string, retention time.Duration) *RedisStorage {
	return &RedisStorage{
		client:    client,
		keyPrefix: keyPrefix,
		retention: retention,
	}
}

func (r *RedisStorage) Store(collection string, snapshot *MetricSnapshot) error {
	ctx := context.Background()
	key := fmt.Sprintf("%s:metrics:%s", r.keyPrefix, collection)

	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	// Use sorted set with timestamp as score
	score := float64(snapshot.Timestamp.Unix())
	member := string(data)

	pipe := r.client.Pipeline()
	pipe.ZAdd(ctx, key, redis.Z{Score: score, Member: member})
	pipe.Expire(ctx, key, r.retention)

	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStorage) Query(collection string, start, end time.Time) ([]*MetricSnapshot, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:metrics:%s", r.keyPrefix, collection)

	min := float64(start.Unix())
	max := float64(end.Unix())

	results, err := r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: fmt.Sprintf("%f", min),
		Max: fmt.Sprintf("%f", max),
	}).Result()

	if err != nil {
		return nil, err
	}

	snapshots := make([]*MetricSnapshot, 0, len(results))
	for _, result := range results {
		var snapshot MetricSnapshot
		if err := json.Unmarshal([]byte(result), &snapshot); err == nil {
			snapshots = append(snapshots, &snapshot)
		}
	}

	return snapshots, nil
}

func (r *RedisStorage) QueryLatest(collection string, limit int) ([]*MetricSnapshot, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:metrics:%s", r.keyPrefix, collection)

	// Get latest entries (highest scores)
	results, err := r.client.ZRevRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	snapshots := make([]*MetricSnapshot, 0, len(results))
	for _, result := range results {
		var snapshot MetricSnapshot
		if err := json.Unmarshal([]byte(result), &snapshot); err == nil {
			snapshots = append(snapshots, &snapshot)
		}
	}

	return snapshots, nil
}

func (r *RedisStorage) Cleanup(before time.Time) error {
	ctx := context.Background()
	pattern := fmt.Sprintf("%s:metrics:*", r.keyPrefix)

	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	score := float64(before.Unix())
	pipe := r.client.Pipeline()

	for _, key := range keys {
		pipe.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%f", score))
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStorage) GetStats() map[string]any {
	ctx := context.Background()
	pattern := fmt.Sprintf("%s:metrics:*", r.keyPrefix)

	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return map[string]any{
			"type":  "redis",
			"error": err.Error(),
		}
	}

	total := int64(0)
	collections := make(map[string]int64)

	for _, key := range keys {
		count, err := r.client.ZCard(ctx, key).Result()
		if err == nil {
			total += count
			// Extract collection name from key
			collection := key[len(r.keyPrefix)+9:] // Remove prefix + ":metrics:"
			collections[collection] = count
		}
	}

	return map[string]any{
		"type":        "redis",
		"total":       total,
		"collections": collections,
		"keys":        len(keys),
	}
}
