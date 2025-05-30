package metrics

import (
	"fmt"
	"sync"
	"time"
)

// Collector manages metric collection and storage
type Collector struct {
	collections map[string]*MetricCollection
	storage     Storage
	retention   time.Duration
	mu          sync.RWMutex
	enabled     bool
}

// NewCollector creates a new metrics collector
func NewCollector(storage Storage, retention time.Duration) *Collector {
	return &Collector{
		collections: make(map[string]*MetricCollection),
		storage:     storage,
		retention:   retention,
		enabled:     true,
	}
}

// SetEnabled enables or disables metric collection
func (c *Collector) SetEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = enabled
}

// Counter creates or updates a counter metric
func (c *Collector) Counter(collection, name string, labels map[string]string, help string) *Metric {
	return c.metric(collection, name, Counter, labels, help, "")
}

// Gauge creates or updates a gauge metric
func (c *Collector) Gauge(collection, name string, labels map[string]string, help string) *Metric {
	return c.metric(collection, name, Gauge, labels, help, "")
}

// Histogram creates or updates a histogram metric
func (c *Collector) Histogram(collection, name string, labels map[string]string, help, unit string) *Metric {
	return c.metric(collection, name, Histogram, labels, help, unit)
}

// metric creates or retrieves a metric
func (c *Collector) metric(collection, name string, metricType MetricType, labels map[string]string, help, unit string) *Metric {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.enabled {
		return &Metric{} // Return dummy metric if disabled
	}

	coll, exists := c.collections[collection]
	if !exists {
		coll = &MetricCollection{
			Name:    collection,
			Metrics: make(map[string]*Metric),
		}
		c.collections[collection] = coll
	}

	key := c.metricKey(name, labels)
	metric, exists := coll.Metrics[key]
	if !exists {
		metric = &Metric{
			Name:      name,
			Type:      metricType,
			Labels:    labels,
			Timestamp: time.Now(),
			Help:      help,
			Unit:      unit,
		}
		coll.Metrics[key] = metric
	}

	coll.LastUpdated = time.Now()
	return metric
}

// Inc increments a counter by 1
func (c *Collector) Inc(collection, name string, labels map[string]string) {
	c.Add(collection, name, 1, labels)
}

// Add adds value to a counter
func (c *Collector) Add(collection, name string, value int64, labels map[string]string) {
	metric := c.Counter(collection, name, labels, "")
	current := c.getInt64Value(metric)
	metric.Value.Store(current + value)
	metric.Timestamp = time.Now()

	c.storeSnapshot(collection, metric, current+value)
}

// Set sets a gauge value
func (c *Collector) Set(collection, name string, value float64, labels map[string]string) {
	metric := c.Gauge(collection, name, labels, "")
	metric.Value.Store(value)
	metric.Timestamp = time.Now()

	c.storeSnapshot(collection, metric, value)
}

// Observe records a histogram observation
func (c *Collector) Observe(collection, name string, value float64, labels map[string]string) {
	metric := c.Histogram(collection, name, labels, "", "")
	metric.Value.Store(value)
	metric.Timestamp = time.Now()

	c.storeSnapshot(collection, metric, value)
}

// GetCollection returns a metric collection
func (c *Collector) GetCollection(name string) (*MetricCollection, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	coll, exists := c.collections[name]
	return coll, exists
}

// GetAllCollections returns all metric collections
func (c *Collector) GetAllCollections() map[string]*MetricCollection {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]*MetricCollection)
	for name, coll := range c.collections {
		result[name] = coll
	}
	return result
}

// Snapshot creates snapshots of all metrics
func (c *Collector) Snapshot() map[string][]*MetricSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string][]*MetricSnapshot)
	for collName, coll := range c.collections {
		snapshots := make([]*MetricSnapshot, 0, len(coll.Metrics))

		for _, metric := range coll.Metrics {
			snapshot := &MetricSnapshot{
				Name:      metric.Name,
				Type:      c.typeString(metric.Type),
				Value:     metric.Value.Load(),
				Labels:    metric.Labels,
				Timestamp: metric.Timestamp,
				Help:      metric.Help,
				Unit:      metric.Unit,
			}
			snapshots = append(snapshots, snapshot)
		}

		result[collName] = snapshots
	}

	return result
}

// Cleanup removes old metrics and performs storage cleanup
func (c *Collector) Cleanup() error {
	if c.storage != nil {
		before := time.Now().Add(-c.retention)
		return c.storage.Cleanup(before)
	}
	return nil
}

// Query retrieves historical metrics
func (c *Collector) Query(collection string, start, end time.Time) ([]*MetricSnapshot, error) {
	if c.storage == nil {
		return nil, fmt.Errorf("storage not configured")
	}
	return c.storage.Query(collection, start, end)
}

// QueryLatest retrieves latest metrics
func (c *Collector) QueryLatest(collection string, limit int) ([]*MetricSnapshot, error) {
	if c.storage == nil {
		return nil, fmt.Errorf("storage not configured")
	}
	return c.storage.QueryLatest(collection, limit)
}

// Helper methods

func (c *Collector) metricKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}

	key := name
	for k, v := range labels {
		key += fmt.Sprintf(",%s=%s", k, v)
	}
	return key
}

func (c *Collector) getInt64Value(metric *Metric) int64 {
	if val := metric.Value.Load(); val != nil {
		if i64, ok := val.(int64); ok {
			return i64
		}
	}
	return 0
}

func (c *Collector) storeSnapshot(collection string, metric *Metric, value any) {
	if c.storage == nil {
		return
	}

	snapshot := &MetricSnapshot{
		Name:      metric.Name,
		Type:      c.typeString(metric.Type),
		Value:     value,
		Labels:    metric.Labels,
		Timestamp: metric.Timestamp,
		Help:      metric.Help,
		Unit:      metric.Unit,
	}

	go c.storage.Store(collection, snapshot)
}

func (c *Collector) typeString(t MetricType) string {
	switch t {
	case Counter:
		return "counter"
	case Gauge:
		return "gauge"
	case Histogram:
		return "histogram"
	case Summary:
		return "summary"
	default:
		return "unknown"
	}
}
