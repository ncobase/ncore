package metrics

import (
	"sort"
	"sync"
)

type MemoryStorage struct {
	metrics []Metric
	mu      sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		metrics: make([]Metric, 0),
	}
}

func (m *MemoryStorage) Store(metrics []Metric) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics = append(m.metrics, metrics...)
	return nil
}

func (m *MemoryStorage) Query(query QueryRequest) ([]Metric, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Metric
	for _, metric := range m.metrics {
		if m.matchesQuery(metric, query) {
			result = append(result, metric)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})

	if query.Limit > 0 && len(result) > query.Limit {
		result = result[:query.Limit]
	}

	return result, nil
}

func (m *MemoryStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics = nil
	return nil
}

func (m *MemoryStorage) matchesQuery(metric Metric, query QueryRequest) bool {
	if query.Type != "" && metric.Type != query.Type {
		return false
	}

	if !query.StartTime.IsZero() && metric.Timestamp.Before(query.StartTime) {
		return false
	}
	if !query.EndTime.IsZero() && metric.Timestamp.After(query.EndTime) {
		return false
	}

	for key, value := range query.Labels {
		if metric.Labels == nil || metric.Labels[key] != value {
			return false
		}
	}

	return true
}
