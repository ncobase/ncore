package monitor

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestRuntimeStats_Basic(t *testing.T) {
	ctx := context.Background()
	monitor, err := NewRuntimeStats(ctx, nil) // Test with default config
	if err != nil {
		t.Fatalf("failed to create monitor: %v", err)
	}

	if !monitor.config.EnableMemory {
		t.Error("memory monitoring should be enabled by default")
	}

	if !monitor.config.EnableCPU {
		t.Error("CPU monitoring should be enabled by default")
	}

	if monitor.IsEnabled() {
		t.Error("monitor should not be enabled before Start()")
	}

	monitor.Start()
	defer monitor.Stop()

	if !monitor.IsEnabled() {
		t.Error("monitor should be enabled after Start()")
	}

	// Wait for first metrics collection
	time.Sleep(2 * time.Second)

	usage := monitor.GetMetrics()

	if usage.Memory == 0 {
		t.Error("memory usage should not be zero")
	}

	if usage.Goroutines == 0 {
		t.Error("goroutine count should not be zero")
	}

	if usage.HeapAlloc == 0 {
		t.Error("heap allocation should not be zero")
	}
}

func TestRuntimeStats_InvalidConfig(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name   string
		config *RuntimeStatsConfig
	}{
		{
			name: "zero interval",
			config: &RuntimeStatsConfig{
				MaxMemory:     1024 * 1024,
				MaxCPU:        50,
				MaxGoroutines: 100,
				Interval:      0,
			},
		},
		{
			name: "negative interval",
			config: &RuntimeStatsConfig{
				MaxMemory:     1024 * 1024,
				MaxCPU:        50,
				MaxGoroutines: 100,
				Interval:      -time.Second,
			},
		},
		{
			name: "negative memory",
			config: &RuntimeStatsConfig{
				MaxMemory:     -1,
				MaxCPU:        50,
				MaxGoroutines: 100,
				Interval:      time.Second,
			},
		},
		{
			name: "invalid CPU percentage",
			config: &RuntimeStatsConfig{
				MaxMemory:     1024 * 1024,
				MaxCPU:        150,
				MaxGoroutines: 100,
				Interval:      time.Second,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewRuntimeStats(ctx, tc.config)
			if err == nil {
				t.Error("expected error for invalid config, got nil")
			}
		})
	}
}

func TestRuntimeStats_Config(t *testing.T) {
	ctx := context.Background()
	config := &RuntimeStatsConfig{
		MaxMemory:     100 * 1024 * 1024, // 100MB
		MaxCPU:        50.0,
		MaxGoroutines: 100,
		Interval:      time.Millisecond * 100,
		EnableCPU:     false,
		EnableMemory:  true,
		EnableGC:      false,
	}

	monitor, err := NewRuntimeStats(ctx, config)
	if err != nil {
		t.Fatalf("failed to create monitor: %v", err)
	}

	monitor.Start()
	defer monitor.Stop()

	// Wait for metrics collection
	time.Sleep(200 * time.Millisecond)

	usage := monitor.GetMetrics()

	if usage.CPU != 0 {
		t.Error("CPU usage should be zero when CPU monitoring is disabled")
	}

	if usage.GCCount != 0 || usage.GCPause != 0 {
		t.Error("GC stats should be zero when GC monitoring is disabled")
	}
}

func TestRuntimeStats_PeakUsage(t *testing.T) {
	ctx := context.Background()
	monitor, err := NewRuntimeStats(ctx, nil)
	if err != nil {
		t.Fatalf("failed to create monitor: %v", err)
	}

	monitor.Start()
	defer monitor.Stop()

	// Generate some memory allocations
	var memoryHog [][]byte
	for i := 0; i < 100; i++ {
		memoryHog = append(memoryHog, make([]byte, 1024*1024)) // Allocate 1MB each
	}

	// Force GC to get accurate memory stats
	runtime.GC()

	// Wait for metrics collection
	time.Sleep(2 * time.Second)

	peak := monitor.GetPeakUsage()
	if peak.Memory == 0 {
		t.Error("peak memory usage should not be zero")
	}

	// Clean up to prevent memory leak in tests
	memoryHog = nil
	runtime.GC()
}

// Benchmarks
func BenchmarkRuntimeStats_GetMetrics(b *testing.B) {
	ctx := context.Background()
	monitor, err := NewRuntimeStats(ctx, nil)
	if err != nil {
		b.Fatalf("failed to create monitor: %v", err)
	}

	monitor.Start()
	defer monitor.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = monitor.GetMetrics()
	}
}

func BenchmarkRuntimeStats_UpdateMetrics(b *testing.B) {
	ctx := context.Background()
	monitor, err := NewRuntimeStats(ctx, nil)
	if err != nil {
		b.Fatalf("failed to create monitor: %v", err)
	}

	monitor.Start()
	defer monitor.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.updateMetrics()
	}
}

func BenchmarkRuntimeStats_Parallel(b *testing.B) {
	ctx := context.Background()
	monitor, err := NewRuntimeStats(ctx, nil)
	if err != nil {
		b.Fatalf("failed to create monitor: %v", err)
	}

	monitor.Start()
	defer monitor.Stop()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = monitor.GetMetrics()
			_ = monitor.CheckThresholds()
		}
	})
}

func BenchmarkRuntimeStats_HighLoad(b *testing.B) {
	ctx := context.Background()
	monitor, err := NewRuntimeStats(ctx, nil)
	if err != nil {
		b.Fatalf("failed to create monitor: %v", err)
	}

	monitor.Start()
	defer monitor.Stop()

	// Create some background load
	done := make(chan struct{})
	defer close(done)

	for i := 0; i < 10; i++ {
		go func() {
			var mem [][]byte
			for {
				select {
				case <-done:
					return
				default:
					mem = append(mem, make([]byte, 1024*1024))
					time.Sleep(time.Millisecond)
				}
			}
		}()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = monitor.GetMetrics()
		_ = monitor.CheckThresholds()
	}
}
