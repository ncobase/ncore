package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStorage stores metrics in Redis with time series support
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

// Store single metric snapshot
func (r *RedisStorage) Store(snapshot *Snapshot) error {
	ctx := context.Background()

	// Store as sorted set with timestamp as score
	key := fmt.Sprintf("%s:metrics:%s:%s", r.keyPrefix, snapshot.ExtensionName, snapshot.MetricType)

	data, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	score := float64(snapshot.Timestamp.Unix())

	pipe := r.client.Pipeline()
	pipe.ZAdd(ctx, key, redis.Z{Score: score, Member: string(data)})
	pipe.Expire(ctx, key, r.retention)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to store snapshot in Redis: %w", err)
	}

	return nil
}

// StoreBatch stores multiple snapshots efficiently
func (r *RedisStorage) StoreBatch(snapshots []*Snapshot) error {
	if len(snapshots) == 0 {
		return nil
	}

	ctx := context.Background()
	pipe := r.client.Pipeline()

	// Group by key for batch operations
	keyGroups := make(map[string][]redis.Z)

	for _, snapshot := range snapshots {
		if snapshot == nil {
			continue // Skip nil snapshots
		}

		key := fmt.Sprintf("%s:metrics:%s:%s", r.keyPrefix, snapshot.ExtensionName, snapshot.MetricType)

		data, err := json.Marshal(snapshot)
		if err != nil {
			continue // Skip snapshots that can't be marshalled
		}

		score := float64(snapshot.Timestamp.Unix())
		keyGroups[key] = append(keyGroups[key], redis.Z{Score: score, Member: string(data)})
	}

	// Execute batch operations
	for key, members := range keyGroups {
		pipe.ZAdd(ctx, key, members...)
		pipe.Expire(ctx, key, r.retention)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to store batch in Redis: %w", err)
	}

	return nil
}

// Query historical metrics with proper filtering and aggregation
func (r *RedisStorage) Query(opts *QueryOptions) ([]*AggregatedMetrics, error) {
	if opts == nil {
		return nil, fmt.Errorf("query options cannot be nil")
	}

	ctx := context.Background()

	// Build key pattern with proper escaping
	pattern := r.buildKeyPattern(opts.ExtensionName, opts.MetricType)

	// Use SCAN instead of KEYS for better performance in production
	keys, err := r.scanKeys(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to scan keys: %w", err)
	}

	var results []*AggregatedMetrics

	for _, key := range keys {
		// Extract extension name and metric type from key
		extensionName, metricType, err := r.parseKey(key)
		if err != nil {
			continue // Skip malformed keys
		}

		// Skip if doesn't match filter criteria
		if opts.ExtensionName != "" && extensionName != opts.ExtensionName {
			continue
		}
		if opts.MetricType != "" && metricType != opts.MetricType {
			continue
		}

		// Query time range with proper score bounds
		members, err := r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
			Min: strconv.FormatInt(opts.StartTime.Unix(), 10),
			Max: strconv.FormatInt(opts.EndTime.Unix(), 10),
		}).Result()

		if err != nil || len(members) == 0 {
			continue
		}

		// Parse snapshots with error handling
		var snapshots []*Snapshot
		for _, member := range members {
			var snapshot Snapshot
			if err := json.Unmarshal([]byte(member), &snapshot); err == nil {
				// Apply label filtering if specified
				if r.matchesLabels(&snapshot, opts.Labels) {
					snapshots = append(snapshots, &snapshot)
				}
			}
		}

		if len(snapshots) == 0 {
			continue
		}

		// Apply limit before aggregation if no interval specified
		if opts.Interval == 0 && opts.Limit > 0 && len(snapshots) > opts.Limit {
			// Sort by timestamp descending and take the latest
			sort.Slice(snapshots, func(i, j int) bool {
				return snapshots[i].Timestamp.After(snapshots[j].Timestamp)
			})
			snapshots = snapshots[:opts.Limit]
		}

		// Aggregate snapshots
		aggregated := r.aggregateSnapshots(snapshots, opts)
		if aggregated != nil {
			results = append(results, aggregated)
		}
	}

	return results, nil
}

// GetLatest retrieves latest metrics for an extension with proper sorting
func (r *RedisStorage) GetLatest(extensionName string, limit int) ([]*Snapshot, error) {
	if extensionName == "" {
		return nil, fmt.Errorf("extension name cannot be empty")
	}

	ctx := context.Background()

	pattern := fmt.Sprintf("%s:metrics:%s:*", r.keyPrefix, extensionName)
	keys, err := r.scanKeys(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to scan keys: %w", err)
	}

	var allSnapshots []*Snapshot

	for _, key := range keys {
		// Get latest entries (highest scores) with proper limit
		queryLimit := int64(limit)
		if limit <= 0 {
			queryLimit = 100 // Default limit to prevent memory issues
		}

		members, err := r.client.ZRevRange(ctx, key, 0, queryLimit-1).Result()
		if err != nil {
			continue
		}

		for _, member := range members {
			var snapshot Snapshot
			if err := json.Unmarshal([]byte(member), &snapshot); err == nil {
				allSnapshots = append(allSnapshots, &snapshot)
			}
		}
	}

	// Sort by timestamp descending and apply limit
	sort.Slice(allSnapshots, func(i, j int) bool {
		return allSnapshots[i].Timestamp.After(allSnapshots[j].Timestamp)
	})

	if limit > 0 && len(allSnapshots) > limit {
		allSnapshots = allSnapshots[:limit]
	}

	return allSnapshots, nil
}

// Cleanup removes old metrics
func (r *RedisStorage) Cleanup(before time.Time) error {
	ctx := context.Background()

	pattern := fmt.Sprintf("%s:metrics:*", r.keyPrefix)
	keys, err := r.scanKeys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to scan keys for cleanup: %w", err)
	}

	score := strconv.FormatInt(before.Unix(), 10)
	pipe := r.client.Pipeline()

	// Batch cleanup operations
	for _, key := range keys {
		pipe.ZRemRangeByScore(ctx, key, "-inf", score)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to cleanup old metrics: %w", err)
	}

	return nil
}

// GetStats retrieves comprehensive storage statistics
func (r *RedisStorage) GetStats() map[string]any {
	ctx := context.Background()

	pattern := fmt.Sprintf("%s:metrics:*", r.keyPrefix)
	keys, err := r.scanKeys(ctx, pattern)
	if err != nil {
		return map[string]any{
			"type":  "redis",
			"error": err.Error(),
		}
	}

	total := int64(0)
	memUsage := int64(0)

	// Use pipeline for efficient stats collection
	pipe := r.client.Pipeline()
	for _, key := range keys {
		pipe.ZCard(ctx, key)
		pipe.MemoryUsage(ctx, key)
	}

	results, err := pipe.Exec(ctx)
	if err == nil {
		for i := 0; i < len(results); i += 2 {
			if cardCmd, ok := results[i].(*redis.IntCmd); ok {
				if count, err := cardCmd.Result(); err == nil {
					total += count
				}
			}
			if memCmd, ok := results[i+1].(*redis.IntCmd); ok {
				if mem, err := memCmd.Result(); err == nil {
					memUsage += mem
				}
			}
		}
	}

	return map[string]any{
		"type":      "redis",
		"total":     total,
		"keys":      len(keys),
		"memory_mb": float64(memUsage) / 1024 / 1024,
		"retention": r.retention.String(),
	}
}

// scanKeys uses SCAN instead of KEYS for better performance
func (r *RedisStorage) scanKeys(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("scan iteration failed: %w", err)
	}

	return keys, nil
}

// buildKeyPattern constructs Redis key pattern for queries
func (r *RedisStorage) buildKeyPattern(extensionName, metricType string) string {
	if extensionName != "" && metricType != "" {
		return fmt.Sprintf("%s:metrics:%s:%s", r.keyPrefix, extensionName, metricType)
	} else if extensionName != "" {
		return fmt.Sprintf("%s:metrics:%s:*", r.keyPrefix, extensionName)
	}
	return fmt.Sprintf("%s:metrics:*", r.keyPrefix)
}

// parseKey extracts extension name and metric type from Redis key
func (r *RedisStorage) parseKey(key string) (string, string, error) {
	// Expected format: prefix:metrics:extension_name:metric_type
	parts := strings.Split(key, ":")
	if len(parts) < 4 {
		return "", "", fmt.Errorf("invalid key format: %s", key)
	}

	// Reconstruct in case extension name or metric type contains colons
	prefixParts := strings.Split(r.keyPrefix, ":")
	expectedPrefixLen := len(prefixParts) + 1 // +1 for "metrics"

	if len(parts) < expectedPrefixLen+2 {
		return "", "", fmt.Errorf("insufficient key parts: %s", key)
	}

	// Find where extension name starts and metric type ends
	extensionStart := expectedPrefixLen
	extensionName := parts[extensionStart]
	metricType := strings.Join(parts[extensionStart+1:], ":")

	return extensionName, metricType, nil
}

// matchesLabels checks if snapshot labels match query criteria
func (r *RedisStorage) matchesLabels(snapshot *Snapshot, labels map[string]string) bool {
	if len(labels) == 0 {
		return true
	}

	for key, value := range labels {
		if snapshot.Labels == nil {
			return false
		}
		if snapValue, exists := snapshot.Labels[key]; !exists || snapValue != value {
			return false
		}
	}

	return true
}

// aggregateSnapshots performs aggregation
func (r *RedisStorage) aggregateSnapshots(snapshots []*Snapshot, opts *QueryOptions) *AggregatedMetrics {
	if len(snapshots) == 0 {
		return nil
	}

	first := snapshots[0]

	// Group by interval if specified
	if opts.Interval > 0 {
		return r.aggregateByInterval(snapshots, opts)
	}

	// Simple aggregation - return raw values
	var values []TimeSeriesPoint
	for _, snapshot := range snapshots {
		values = append(values, TimeSeriesPoint{
			Timestamp: snapshot.Timestamp,
			Value:     snapshot.Value,
		})
	}

	// Sort by timestamp
	sort.Slice(values, func(i, j int) bool {
		return values[i].Timestamp.Before(values[j].Timestamp)
	})

	aggregationType := opts.Aggregation
	if aggregationType == "" {
		aggregationType = "raw"
	}

	return &AggregatedMetrics{
		ExtensionName: first.ExtensionName,
		MetricType:    first.MetricType,
		Values:        values,
		Aggregation:   aggregationType,
	}
}

// aggregateByInterval performs time-based interval aggregation
func (r *RedisStorage) aggregateByInterval(snapshots []*Snapshot, opts *QueryOptions) *AggregatedMetrics {
	if len(snapshots) == 0 {
		return nil
	}

	// Group by interval
	intervals := make(map[int64][]*Snapshot)
	intervalSeconds := int64(opts.Interval.Seconds())

	for _, snapshot := range snapshots {
		intervalKey := snapshot.Timestamp.Unix() / intervalSeconds
		intervals[intervalKey] = append(intervals[intervalKey], snapshot)
	}

	var values []TimeSeriesPoint
	for intervalKey, intervalSnapshots := range intervals {
		timestamp := time.Unix(intervalKey*intervalSeconds, 0)
		value := r.calculateAggregation(intervalSnapshots, opts.Aggregation)

		values = append(values, TimeSeriesPoint{
			Timestamp: timestamp,
			Value:     value,
		})
	}

	// Sort by timestamp
	sort.Slice(values, func(i, j int) bool {
		return values[i].Timestamp.Before(values[j].Timestamp)
	})

	aggregationType := opts.Aggregation
	if aggregationType == "" {
		aggregationType = "avg"
	}

	return &AggregatedMetrics{
		ExtensionName: snapshots[0].ExtensionName,
		MetricType:    snapshots[0].MetricType,
		Values:        values,
		Aggregation:   aggregationType,
	}
}

// calculateAggregation performs the actual aggregation calculation
func (r *RedisStorage) calculateAggregation(snapshots []*Snapshot, aggType string) int64 {
	if len(snapshots) == 0 {
		return 0
	}

	switch aggType {
	case "sum":
		var sum int64
		for _, s := range snapshots {
			sum += s.Value
		}
		return sum
	case "avg":
		var sum int64
		for _, s := range snapshots {
			sum += s.Value
		}
		return sum / int64(len(snapshots))
	case "max":
		maxValue := snapshots[0].Value
		for _, s := range snapshots {
			if s.Value > maxValue {
				maxValue = s.Value
			}
		}
		return maxValue
	case "min":
		minValue := snapshots[0].Value
		for _, s := range snapshots {
			if s.Value < minValue {
				minValue = s.Value
			}
		}
		return minValue
	case "count":
		return int64(len(snapshots))
	default:
		// Return latest value for unknown aggregation types
		sort.Slice(snapshots, func(i, j int) bool {
			return snapshots[i].Timestamp.Before(snapshots[j].Timestamp)
		})
		return snapshots[len(snapshots)-1].Value
	}
}
