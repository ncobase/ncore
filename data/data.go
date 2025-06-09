package data

import (
	"sync"

	"github.com/ncobase/ncore/data/connection"
	"github.com/ncobase/ncore/data/messaging/kafka"
	"github.com/ncobase/ncore/data/messaging/rabbitmq"
	"github.com/ncobase/ncore/data/metrics"
	"github.com/ncobase/ncore/data/search"
)

type ContextKey string

const (
	ContextKeyTransaction ContextKey = "tx"
)

var sharedInstance *Data

// Data represents the data layer implementation
type Data struct {
	Conn         *connection.Connections
	RabbitMQ     *rabbitmq.RabbitMQ
	Kafka        *kafka.Kafka
	searchClient *search.Client
	collector    metrics.Collector
	searchOnce   sync.Once
	closed       bool
	mu           sync.RWMutex // Protects all fields from concurrent access
}

// Option function type for configuring Data
type Option func(*Data)

// WithMetricsCollector sets the metrics collector
func WithMetricsCollector(collector metrics.Collector) Option {
	return func(d *Data) {
		if collector != nil {
			d.collector = collector
		}
	}
}

// WithExtensionCollector sets extension layer collector using adapter
func WithExtensionCollector(collector metrics.ExtensionCollector) Option {
	return func(d *Data) {
		if collector != nil {
			d.collector = metrics.NewExtensionCollectorAdapter(collector)
		}
	}
}
