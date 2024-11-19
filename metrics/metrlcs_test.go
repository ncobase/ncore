package metrics

import (
	"context"
	"testing"
	"time"
)

func TestCollector(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FlushInterval = time.Millisecond * 10

	collector, err := NewCollector(cfg)
	if err != nil {
		t.Fatalf("Failed to create collector: %v", err)
	}

	ctx := context.Background()
	if err := collector.Start(ctx); err != nil {
		t.Fatalf("Failed to start collector: %v", err)
	}
	defer collector.Stop()

	collector.RecordProcessStart()
	time.Sleep(time.Millisecond)
	collector.RecordProcessCompletion(time.Since(time.Now()).Seconds(), true)

	collector.RecordExecutorAttemptDuration("task_executor", time.Millisecond, nil)
	collector.RecordExecutorRetryCount("task_executor", 1, nil)

	// Wait for metrics to be flushed
	time.Sleep(cfg.FlushInterval * 2)

	metrics := collector.GetMetrics()

	if metrics["process"].(map[string]any)["total"].(int64) != 1 {
		t.Errorf("Unexpected process total: got %d, want 1", metrics["process"].(map[string]any)["total"].(int64))
	}

	if metrics["process"].(map[string]any)["completed"].(int64) != 1 {
		t.Errorf("Unexpected process completed: got %d, want 1", metrics["process"].(map[string]any)["completed"].(int64))
	}

	// Add more assertions for other metrics...
}

func BenchmarkCollector(b *testing.B) {
	cfg := DefaultConfig()
	cfg.FlushInterval = time.Millisecond * 10

	collector, _ := NewCollector(cfg)

	ctx := context.Background()
	_ = collector.Start(ctx)
	defer collector.Stop()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		collector.RecordProcessStart()
		collector.RecordProcessCompletion(time.Since(time.Now()).Seconds(), true)

		collector.RecordExecutorAttemptDuration("task_executor", time.Millisecond, nil)
		collector.RecordExecutorRetryCount("task_executor", 1, nil)
	}
}
