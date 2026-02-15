// Package metrics provides comprehensive metrics collection, aggregation,
// and monitoring for extensions and application components.
//
// This package offers:
//   - Real-time metrics collection (memory, CPU, goroutines)
//   - Extension-specific performance tracking
//   - Time-series data aggregation
//   - Metric storage (Redis, in-memory)
//   - Query and analysis capabilities
//   - System resource monitoring
//
// # Basic Metrics Collection
//
// Create a collector and track metrics:
//
//	collector := metrics.NewCollector(cfg, logger, redisClient)
//
//	// Record extension metrics
//	collector.RecordExtensionMetric(ctx, &metrics.ExtensionMetric{
//	    ExtensionID: "my-plugin",
//	    MetricType:  metrics.MetricTypeMemory,
//	    Value:       125.5,  // MB
//	    Timestamp:   time.Now(),
//	})
//
// # Metric Types
//
// Supported metric types:
//   - MetricTypeMemory: Memory usage in MB
//   - MetricTypeCPU: CPU usage percentage
//   - MetricTypeGoroutines: Active goroutine count
//   - MetricTypeRequests: Request count
//   - MetricTypeErrors: Error count
//   - MetricTypeLatency: Operation latency in ms
//
// # Aggregation Queries
//
// Query aggregated metrics over time:
//
//	metrics, err := collector.QueryMetrics(ctx, &metrics.Query{
//	    ExtensionID: "my-plugin",
//	    MetricType:  metrics.MetricTypeMemory,
//	    TimeRange: metrics.TimeRange{
//	        Start: time.Now().Add(-1 * time.Hour),
//	        End:   time.Now(),
//	    },
//	    Aggregation: metrics.AggregationAverage,
//	    Interval:    5 * time.Minute,
//	})
//
// # Aggregation Types
//
//   - AggregationSum: Sum of all values
//   - AggregationAverage: Mean value
//   - AggregationMin: Minimum value
//   - AggregationMax: Maximum value
//   - AggregationCount: Number of data points
//
// # System Monitoring
//
// Track overall system health:
//
//	stats := metrics.GetSystemStats()
//	// Returns: memory, CPU, goroutines, load averages
//
//	// Monitor specific extension
//	extensionStats := collector.GetExtensionStats(ctx, "my-plugin")
//
// # Resource Monitoring
//
// Automatic resource tracking for extensions:
//
//	monitor := metrics.NewResourceMonitor(cfg)
//
//	// Record extension resource usage
//	monitor.RecordMetrics("my-plugin", &metrics.ResourceMetrics{
//	    MemoryMB:    50.2,
//	    CPUPercent:  15.5,
//	    Goroutines:  10,
//	})
//
//	// Check if resources are within limits
//	if err := monitor.CheckLimits("my-plugin"); err != nil {
//	    log.Printf("Resource limit exceeded: %v", err)
//	}
//
// # Storage Backends
//
// Metrics can be stored in Redis (persistent) or memory (ephemeral):
//
//	// With Redis (recommended for production)
//	collector := metrics.NewCollector(cfg, logger, redisClient)
//
//	// In-memory only (for testing/development)
//	collector := metrics.NewCollector(cfg, logger, nil)
//
// # Metric Retention
//
// Configure automatic cleanup of old metrics:
//
//	cfg := &config.MetricsConfig{
//	    RetentionPeriod: 7 * 24 * time.Hour,  // Keep 7 days
//	    CleanupInterval: 1 * time.Hour,        // Cleanup hourly
//	}
//
// # Best Practices
//
//   - Use appropriate aggregation intervals (1m, 5m, 1h)
//   - Set reasonable retention periods (7-30 days)
//   - Monitor metric collection overhead
//   - Use Redis for production deployments
//   - Implement alerting based on metric thresholds
//   - Track trends over time, not just current values
//   - Clean up metrics for removed extensions
package metrics
