package metrics

import (
	"sync/atomic"
	"time"
)

// MetricType defines metric value types
type MetricType int

const (
	Counter MetricType = iota
	Gauge
	Histogram
	Summary
)

// Metric represents a single metric
type Metric struct {
	Name      string            `json:"name"`
	Type      MetricType        `json:"type"`
	Value     atomic.Value      `json:"value"` // stores int64 or float64
	Labels    map[string]string `json:"labels,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Help      string            `json:"help,omitempty"`
	Unit      string            `json:"unit,omitempty"`
}

// MetricSnapshot represents a point-in-time metric value
type MetricSnapshot struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Value     any               `json:"value"`
	Labels    map[string]string `json:"labels,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Help      string            `json:"help,omitempty"`
	Unit      string            `json:"unit,omitempty"`
}

// MetricCollection groups related metrics
type MetricCollection struct {
	Name        string             `json:"name"`
	Metrics     map[string]*Metric `json:"metrics"`
	Snapshots   []*MetricSnapshot  `json:"snapshots,omitempty"`
	LastUpdated time.Time          `json:"last_updated"`
}

// Storage defines metric storage interface
type Storage interface {
	Store(collection string, snapshot *MetricSnapshot) error
	Query(collection string, start, end time.Time) ([]*MetricSnapshot, error)
	QueryLatest(collection string, limit int) ([]*MetricSnapshot, error)
	Cleanup(before time.Time) error
	GetStats() map[string]any
}
