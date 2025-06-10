package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// MemoryStorage stores metrics in memory
type MemoryStorage struct {
	metrics []Metric
	mu      sync.RWMutex
}

// NewMemoryStorage creates a new memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		metrics: make([]Metric, 0),
	}
}

// Store stores metrics in memory
func (m *MemoryStorage) Store(metrics []Metric) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics = append(m.metrics, metrics...)
	return nil
}

// Query queries metrics from memory
func (m *MemoryStorage) Query(query QueryRequest) ([]Metric, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Metric
	for _, metric := range m.metrics {
		if m.matchesQuery(metric, query) {
			result = append(result, metric)
		}
	}

	// Sort by timestamp
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})

	// Apply limit
	if query.Limit > 0 && len(result) > query.Limit {
		result = result[:query.Limit]
	}

	return result, nil
}

// Close closes the memory storage
func (m *MemoryStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics = nil
	return nil
}

// matchesQuery checks if metric matches query criteria
func (m *MemoryStorage) matchesQuery(metric Metric, query QueryRequest) bool {
	// Check type
	if query.Type != "" && metric.Type != query.Type {
		return false
	}

	// Check time range
	if !query.StartTime.IsZero() && metric.Timestamp.Before(query.StartTime) {
		return false
	}
	if !query.EndTime.IsZero() && metric.Timestamp.After(query.EndTime) {
		return false
	}

	// Check labels
	for key, value := range query.Labels {
		if metric.Labels == nil || metric.Labels[key] != value {
			return false
		}
	}

	return true
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

// Store stores metrics in Redis
func (r *RedisStorage) Store(metrics []Metric) error {
	ctx := context.Background()
	pipe := r.client.Pipeline()

	for _, metric := range metrics {
		key := fmt.Sprintf("%s:metrics:%s", r.keyPrefix, metric.Type)
		data, err := json.Marshal(metric)
		if err != nil {
			continue
		}

		// Use timestamp as score for sorted set
		score := float64(metric.Timestamp.Unix())
		pipe.ZAdd(ctx, key, redis.Z{Score: score, Member: string(data)})
		pipe.Expire(ctx, key, r.retention)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Query queries metrics from Redis
func (r *RedisStorage) Query(query QueryRequest) ([]Metric, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:metrics:%s", r.keyPrefix, query.Type)

	// Build score range
	var minValut, maxValue string
	if !query.StartTime.IsZero() {
		minValut = fmt.Sprintf("%d", query.StartTime.Unix())
	} else {
		minValut = "-inf"
	}
	if !query.EndTime.IsZero() {
		maxValue = fmt.Sprintf("%d", query.EndTime.Unix())
	} else {
		maxValue = "+inf"
	}

	// Query with score range
	members, err := r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: minValut,
		Max: maxValue,
	}).Result()
	if err != nil {
		return nil, err
	}

	var result []Metric
	for _, member := range members {
		var metric Metric
		if err := json.Unmarshal([]byte(member), &metric); err != nil {
			continue
		}

		// Check labels
		if r.matchesLabels(metric, query.Labels) {
			result = append(result, metric)
		}
	}

	// Apply limit
	if query.Limit > 0 && len(result) > query.Limit {
		result = result[:query.Limit]
	}

	return result, nil
}

// Close closes the Redis storage
func (r *RedisStorage) Close() error {
	return nil // Redis client is managed externally
}

// matchesLabels checks if metric labels match query labels
func (r *RedisStorage) matchesLabels(metric Metric, queryLabels Labels) bool {
	for key, value := range queryLabels {
		if metric.Labels == nil || metric.Labels[key] != value {
			return false
		}
	}
	return true
}
