package data

import (
	"sync"

	"github.com/ncobase/ncore/data/config"
	"github.com/ncobase/ncore/data/connection"
	"github.com/ncobase/ncore/data/metrics"
)

type ContextKey string

const (
	ContextKeyTransaction ContextKey = "tx"
)

var sharedInstance *Data

type Data struct {
	Conn      *connection.Connections
	collector metrics.Collector
	conf      *config.Config
	closed    bool
	mu        sync.RWMutex
}

type Option func(*Data)

func WithMetricsCollector(collector metrics.Collector) Option {
	return func(d *Data) {
		if collector != nil {
			d.collector = collector
		}
	}
}

func WithExtensionCollector(collector metrics.ExtensionCollector) Option {
	return func(d *Data) {
		if collector != nil {
			d.collector = metrics.NewExtensionCollectorAdapter(collector)
		}
	}
}

func WithSearchConfig(searchConfig *config.Search) Option {
	return func(d *Data) {
		if d.conf == nil {
			d.conf = &config.Config{}
		}
		d.conf.Search = searchConfig
	}
}

func WithIndexPrefix(prefix string) Option {
	return func(d *Data) {
		if d.conf == nil {
			d.conf = &config.Config{}
		}
		if d.conf.Search == nil {
			d.conf.Search = &config.Search{}
		}
		d.conf.Search.IndexPrefix = prefix
	}
}
