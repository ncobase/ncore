package metrics

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// MemoryStorage stores metrics in memory with thread safety
type MemoryStorage struct {
	data        map[string][]*Snapshot // key: extension_name:metric_type
	mu          sync.RWMutex
	stats       StorageStats
	lastCleanup time.Time
	errors      int64
}

// NewMemoryStorage creates a new memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string][]*Snapshot),
		stats: StorageStats{
			Type: "memory",
		},
		lastCleanup: time.Now(),
	}
}

// Store single metric snapshot
func (m *MemoryStorage) Store(snapshot *Snapshot) error {
	if snapshot == nil {
		return NewMetricError("store", "snapshot cannot be nil", nil)
	}

	if snapshot.ExtensionName == "" {
		return NewMetricError("store", "extension name cannot be empty", nil)
	}

	if snapshot.MetricType == "" {
		return NewMetricError("store", "metric type cannot be empty", nil)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.buildKey(snapshot.ExtensionName, snapshot.MetricType)
	m.data[key] = append(m.data[key], snapshot)
	m.stats.Total++

	return nil
}

// StoreBatch stores multiple snapshots efficiently
func (m *MemoryStorage) StoreBatch(snapshots []*Snapshot) error {
	if len(snapshots) == 0 {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	validCount := 0
	for _, snapshot := range snapshots {
		if snapshot == nil || snapshot.ExtensionName == "" || snapshot.MetricType == "" {
			continue // Skip invalid snapshots
		}

		key := m.buildKey(snapshot.ExtensionName, snapshot.MetricType)
		m.data[key] = append(m.data[key], snapshot)
		validCount++
	}

	m.stats.Total += int64(validCount)
	m.updateMemoryStats()

	return nil
}

// Query historical metrics with proper filtering and aggregation
func (m *MemoryStorage) Query(opts *QueryOptions) ([]*AggregatedMetrics, error) {
	if err := opts.Validate(); err != nil {
		return nil, NewMetricError("query", "invalid query options", err)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*AggregatedMetrics

	for key, snapshots := range m.data {
		extensionName, metricType := m.parseKey(key)

		// Apply filters
		if opts.ExtensionName != "" && extensionName != opts.ExtensionName {
			continue
		}
		if opts.MetricType != "" && metricType != opts.MetricType {
			continue
		}

		// Filter by time range and labels
		var filtered []*Snapshot
		for _, snapshot := range snapshots {
			if m.isInTimeRange(snapshot, opts.StartTime, opts.EndTime) &&
				m.matchesLabels(snapshot, opts.Labels) {
				filtered = append(filtered, snapshot)
			}
		}

		if len(filtered) == 0 {
			continue
		}

		// Apply limit before aggregation if no interval specified
		if opts.Interval == 0 && opts.Limit > 0 && len(filtered) > opts.Limit {
			// Sort by timestamp descending and take the latest
			sort.Slice(filtered, func(i, j int) bool {
				return filtered[i].Timestamp.After(filtered[j].Timestamp)
			})
			filtered = filtered[:opts.Limit]
		}

		// Aggregate snapshots
		aggregated := m.aggregateSnapshots(filtered, opts)
		if aggregated != nil {
			results = append(results, aggregated)
		}
	}

	return results, nil
}

// GetLatest retrieves latest metrics for an extension
func (m *MemoryStorage) GetLatest(extensionName string, limit int) ([]*Snapshot, error) {
	if extensionName == "" {
		return nil, NewMetricError("get_latest", "extension name cannot be empty", nil)
	}

	if limit <= 0 {
		limit = 10 // Default limit
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var allSnapshots []*Snapshot
	for key, snapshots := range m.data {
		keyExtensionName, _ := m.parseKey(key)
		if keyExtensionName == extensionName {
			allSnapshots = append(allSnapshots, snapshots...)
		}
	}

	// Sort by timestamp descending
	sort.Slice(allSnapshots, func(i, j int) bool {
		return allSnapshots[i].Timestamp.After(allSnapshots[j].Timestamp)
	})

	if len(allSnapshots) > limit {
		allSnapshots = allSnapshots[:limit]
	}

	return allSnapshots, nil
}

// Cleanup removes old metrics before the specified time
func (m *MemoryStorage) Cleanup(before time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	removedCount := int64(0)
	for key, snapshots := range m.data {
		var filtered []*Snapshot
		for _, snapshot := range snapshots {
			if snapshot.Timestamp.After(before) {
				filtered = append(filtered, snapshot)
			} else {
				removedCount++
			}
		}

		if len(filtered) == 0 {
			delete(m.data, key)
		} else {
			m.data[key] = filtered
		}
	}

	m.stats.Total -= removedCount
	m.lastCleanup = time.Now()
	m.updateMemoryStats()

	return nil
}

// GetStats returns storage statistics
func (m *MemoryStorage) GetStats() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.updateMemoryStats()

	return map[string]any{
		"type":         m.stats.Type,
		"total":        m.stats.Total,
		"keys":         len(m.data),
		"memory_mb":    m.stats.MemoryMB,
		"last_cleanup": m.lastCleanup,
		"errors":       m.errors,
	}
}

// buildKey builds a unique key for a metric
func (m *MemoryStorage) buildKey(extensionName, metricType string) string {
	return fmt.Sprintf("%s:%s", extensionName, metricType)
}

func (m *MemoryStorage) parseKey(key string) (string, string) {
	parts := strings.SplitN(key, ":", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

// isInTimeRange checks if a snapshot is within the specified time range
func (m *MemoryStorage) isInTimeRange(snapshot *Snapshot, start, end time.Time) bool {
	return snapshot.Timestamp.After(start) && snapshot.Timestamp.Before(end)
}

// matchesLabels checks if snapshot labels match query criteria
func (m *MemoryStorage) matchesLabels(snapshot *Snapshot, labels map[string]string) bool {
	if len(labels) == 0 {
		return true
	}

	if snapshot.Labels == nil {
		return false
	}

	for key, value := range labels {
		if snapValue, exists := snapshot.Labels[key]; !exists || snapValue != value {
			return false
		}
	}

	return true
}

// updateMemoryStats updates memory stats
func (m *MemoryStorage) updateMemoryStats() {
	// Rough estimate: each snapshot is approximately 200 bytes
	m.stats.MemoryMB = float64(m.stats.Total*200) / 1024 / 1024
}

// aggregateSnapshots performs aggregation
func (m *MemoryStorage) aggregateSnapshots(snapshots []*Snapshot, opts *QueryOptions) *AggregatedMetrics {
	if len(snapshots) == 0 {
		return nil
	}

	first := snapshots[0]

	// Group by interval if specified
	if opts.Interval > 0 {
		return m.aggregateByInterval(snapshots, opts)
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
		aggregationType = AggregationRaw
	}

	return &AggregatedMetrics{
		ExtensionName: first.ExtensionName,
		MetricType:    first.MetricType,
		Values:        values,
		Aggregation:   aggregationType,
	}
}

// aggregateByInterval performs time-based interval aggregation
func (m *MemoryStorage) aggregateByInterval(snapshots []*Snapshot, opts *QueryOptions) *AggregatedMetrics {
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
		value := m.calculateAggregation(intervalSnapshots, opts.Aggregation)

		values = append(values, TimeSeriesPoint{
			Timestamp: timestamp,
			Value:     value,
		})
	}

	// Sort by timestamp
	sort.Slice(values, func(i, j int) bool {
		return values[i].Timestamp.Before(values[j].Timestamp)
	})

	// Apply limit after aggregation
	if opts.Limit > 0 && len(values) > opts.Limit {
		values = values[:opts.Limit]
	}

	aggregationType := opts.Aggregation
	if aggregationType == "" {
		aggregationType = AggregationAvg
	}

	return &AggregatedMetrics{
		ExtensionName: snapshots[0].ExtensionName,
		MetricType:    snapshots[0].MetricType,
		Values:        values,
		Aggregation:   aggregationType,
	}
}

// calculateAggregation performs aggregation
func (m *MemoryStorage) calculateAggregation(snapshots []*Snapshot, aggType string) int64 {
	if len(snapshots) == 0 {
		return 0
	}

	switch aggType {
	case AggregationSum, "":
		var sum int64
		for _, s := range snapshots {
			sum += s.Value
		}
		return sum

	case AggregationAvg:
		var sum int64
		for _, s := range snapshots {
			sum += s.Value
		}
		return sum / int64(len(snapshots))

	case AggregationMax:
		maxValue := snapshots[0].Value
		for _, s := range snapshots {
			if s.Value > maxValue {
				maxValue = s.Value
			}
		}
		return maxValue

	case AggregationMin:
		minValue := snapshots[0].Value
		for _, s := range snapshots {
			if s.Value < minValue {
				minValue = s.Value
			}
		}
		return minValue

	case AggregationCount:
		return int64(len(snapshots))

	default:
		// Return latest value for unknown aggregation types
		sort.Slice(snapshots, func(i, j int) bool {
			return snapshots[i].Timestamp.Before(snapshots[j].Timestamp)
		})
		return snapshots[len(snapshots)-1].Value
	}
}
