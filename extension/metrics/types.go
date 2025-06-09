package metrics

import (
	"fmt"
	"sync/atomic"
	"time"
)

// Constants for aggregation types
const (
	AggregationRaw   = "raw"
	AggregationSum   = "sum"
	AggregationAvg   = "avg"
	AggregationMax   = "max"
	AggregationMin   = "min"
	AggregationCount = "count"
)

// ExtensionMetrics tracks real-time metrics for a single extension
type ExtensionMetrics struct {
	Name          string    `json:"name"`
	LoadTime      int64     `json:"load_time_ms"`   // Load time in milliseconds
	InitTime      int64     `json:"init_time_ms"`   // Init time in milliseconds
	LoadedAt      time.Time `json:"loaded_at"`      // When extension was loaded
	InitializedAt time.Time `json:"initialized_at"` // When extension was initialized
	Status        string    `json:"status"`         // "loading", "active", "failed", "stopped"

	// Atomic counters for concurrent access (use atomic.Int64 internally but convert for JSON)
	ServiceCalls        int64 `json:"service_calls"`
	ServiceErrors       int64 `json:"service_errors"`
	EventsPublished     int64 `json:"events_published"`
	EventsReceived      int64 `json:"events_received"`
	CircuitBreakerTrips int64 `json:"circuit_breaker_trips"`

	// Internal atomic counters (not exported for JSON)
	serviceCalls        atomic.Int64 `json:"-"`
	serviceErrors       atomic.Int64 `json:"-"`
	eventsPublished     atomic.Int64 `json:"-"`
	eventsReceived      atomic.Int64 `json:"-"`
	circuitBreakerTrips atomic.Int64 `json:"-"`
}

// SystemMetrics tracks system-wide metrics
type SystemMetrics struct {
	StartTime          time.Time `json:"start_time"`
	MemoryUsageMB      int64     `json:"memory_usage_mb"`
	GoroutineCount     int       `json:"goroutine_count"`
	GCCycles           uint32    `json:"gc_cycles"`
	ServicesRegistered int       `json:"services_registered"`
	ServiceCacheHits   int64     `json:"service_cache_hits"`
	ServiceCacheMisses int64     `json:"service_cache_misses"`
}

// Snapshot represents a point-in-time metric measurement
type Snapshot struct {
	ExtensionName string            `json:"extension_name"`
	MetricType    string            `json:"metric_type"`
	Value         int64             `json:"value"`
	Labels        map[string]string `json:"labels,omitempty"`
	Timestamp     time.Time         `json:"timestamp"`
}

// AggregatedMetrics represents query results with aggregation
type AggregatedMetrics struct {
	ExtensionName string            `json:"extension_name"`
	MetricType    string            `json:"metric_type"`
	Values        []TimeSeriesPoint `json:"values"`
	Aggregation   string            `json:"aggregation"` // "sum", "avg", "max", "min", "count", "raw"
}

// TimeSeriesPoint represents a single point in time series
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     int64     `json:"value"`
}

// QueryOptions specifies parameters for historical metric queries
type QueryOptions struct {
	ExtensionName string            `json:"extension_name,omitempty"`
	MetricType    string            `json:"metric_type,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	StartTime     time.Time         `json:"start_time"`
	EndTime       time.Time         `json:"end_time"`
	Aggregation   string            `json:"aggregation"` // "sum", "avg", "max", "min", "count", "raw"
	Interval      time.Duration     `json:"interval"`    // aggregation interval (0 = no interval aggregation)
	Limit         int               `json:"limit"`       // maximum number of results
}

// Validate checks if QueryOptions are valid
func (q *QueryOptions) Validate() error {
	if q == nil {
		return fmt.Errorf("query options cannot be nil")
	}
	if q.StartTime.IsZero() {
		return fmt.Errorf("start time cannot be zero")
	}
	if q.EndTime.IsZero() {
		return fmt.Errorf("end time cannot be zero")
	}
	if q.EndTime.Before(q.StartTime) {
		return fmt.Errorf("end time cannot be before start time")
	}
	if q.Limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}
	if q.Interval < 0 {
		return fmt.Errorf("interval cannot be negative")
	}

	// Validate aggregation type
	validAggregations := map[string]bool{
		"":      true, // empty is valid (defaults to "raw")
		"raw":   true,
		"sum":   true,
		"avg":   true,
		"max":   true,
		"min":   true,
		"count": true,
	}
	if !validAggregations[q.Aggregation] {
		return fmt.Errorf("invalid aggregation type: %s", q.Aggregation)
	}

	return nil
}

// Storage interface for metrics persistence
type Storage interface {
	// Store single metric snapshot
	Store(snapshot *Snapshot) error

	// StoreBatch stores multiple snapshots efficiently
	StoreBatch(snapshots []*Snapshot) error

	// Query historical metrics with aggregation and filtering
	Query(opts *QueryOptions) ([]*AggregatedMetrics, error)

	// GetLatest retrieves latest metrics for an extension
	GetLatest(extensionName string, limit int) ([]*Snapshot, error)

	// Cleanup removes old metrics before the specified time
	Cleanup(before time.Time) error

	// GetStats returns storage statistics and health information
	GetStats() map[string]any
}

// CollectorConfig represents configuration for the metrics collector
type CollectorConfig struct {
	Enabled       bool          `json:"enabled"`
	BatchSize     int           `json:"batch_size"`
	FlushInterval time.Duration `json:"flush_interval"`
	Retention     time.Duration `json:"retention"`
	Storage       StorageConfig `json:"storage"`
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	Type      string            `json:"type"`       // "memory", "redis"
	KeyPrefix string            `json:"key_prefix"` // Redis key prefix
	Options   map[string]string `json:"options"`    // Storage-specific options
}

// DefaultCollectorConfig provides default configuration
var DefaultCollectorConfig = CollectorConfig{
	Enabled:       true,
	BatchSize:     100,
	FlushInterval: 30 * time.Second,
	Retention:     7 * 24 * time.Hour, // 7 days
	Storage: StorageConfig{
		Type:      "memory",
		KeyPrefix: "ncore",
		Options:   make(map[string]string),
	},
}

// StorageStats represents common storage statistics
type StorageStats struct {
	Type        string     `json:"type"`                   // "memory", "redis", etc.
	Total       int64      `json:"total"`                  // Total number of stored snapshots
	Keys        int        `json:"keys"`                   // Number of storage keys
	MemoryMB    float64    `json:"memory_mb"`              // Memory usage in MB
	Retention   string     `json:"retention"`              // Retention policy
	LastCleanup *time.Time `json:"last_cleanup,omitempty"` // Last cleanup time
	Errors      int64      `json:"errors"`                 // Number of storage errors
}

// MetricError represents an error that occurred during metric operations
type MetricError struct {
	Operation string    `json:"operation"` // "store", "query", "cleanup", etc.
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Err       error     `json:"-"` // Original error (not serialized)
}

func (e *MetricError) Error() string {
	return fmt.Sprintf("metric %s error: %s", e.Operation, e.Message)
}

func (e *MetricError) Unwrap() error {
	return e.Err
}

// NewMetricError creates a new metric error
func NewMetricError(operation, message string, err error) *MetricError {
	return &MetricError{
		Operation: operation,
		Message:   message,
		Timestamp: time.Now(),
		Err:       err,
	}
}

// HealthStatus represents the health status of the metrics system
type HealthStatus struct {
	Enabled       bool              `json:"enabled"`
	StorageType   string            `json:"storage_type"`
	StorageHealth string            `json:"storage_health"` // "healthy", "degraded", "unhealthy"
	LastError     *MetricError      `json:"last_error,omitempty"`
	Stats         map[string]any    `json:"stats"`
	Extensions    map[string]string `json:"extensions"` // extension_name -> status
}
