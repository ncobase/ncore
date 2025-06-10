package data

import (
	"time"

	"github.com/ncobase/ncore/data/metrics"
)

// GetMetricsCollector returns the metrics collector
func (d *Data) GetMetricsCollector() metrics.Collector {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.collector
}

// GetStats returns data layer statistics
func (d *Data) GetStats() map[string]any {
	d.mu.RLock()
	collector := d.collector
	d.mu.RUnlock()

	// Check if collector is DataCollector and has GetStats method
	if dataCollector, ok := collector.(*metrics.DataCollector); ok {
		return dataCollector.GetStats()
	}

	return map[string]any{
		"status":    "metrics_unavailable",
		"timestamp": time.Now(),
	}
}

// Close closes all data connections
func (d *Data) Close() []error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil
	}

	d.closed = true
	var errs []error

	// Close metrics collector if it has Close method
	if dataCollector, ok := d.collector.(*metrics.DataCollector); ok {
		if err := dataCollector.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// Close connections through connection manager
	if d.Conn != nil {
		if connErrs := d.Conn.Close(); len(connErrs) > 0 {
			errs = append(errs, connErrs...)
		}
		d.Conn = nil
	}

	// Clear other references
	d.RabbitMQ = nil
	d.Kafka = nil
	d.searchClient = nil
	d.collector = metrics.NoOpCollector{} // Reset to no-op collector

	return errs
}
