package monitor

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sync/atomic"
	"time"
)

// RuntimeStatsConfig contains configuration for runtime stats monitor
type RuntimeStatsConfig struct {
	// Monitor
	MaxMemory     int64         `json:"max_memory"`     // Max memory usage in bytes
	MaxCPU        float64       `json:"max_cpu"`        // Max CPU usage percentage
	MaxGoroutines int32         `json:"max_goroutines"` // Max number of goroutines
	Interval      time.Duration `json:"interval"`       // Monitoring interval
	EnableCPU     bool          `json:"enable_cpu"`     // Enable CPU monitoring
	EnableMemory  bool          `json:"enable_memory"`  // Enable memory monitoring
	EnableGC      bool          `json:"enable_gc"`      // Enable GC stats monitoring
}

func (c *RuntimeStatsConfig) Validate() error {
	if c.Interval <= 0 {
		return fmt.Errorf("monitoring interval must be positive, got %v", c.Interval)
	}
	if c.MaxMemory <= 0 {
		return fmt.Errorf("max memory must be positive, got %d", c.MaxMemory)
	}
	if c.MaxCPU <= 0 || c.MaxCPU > 100 {
		return fmt.Errorf("max CPU must be between 0 and 100, got %f", c.MaxCPU)
	}
	if c.MaxGoroutines <= 0 {
		return fmt.Errorf("max goroutines must be positive, got %d", c.MaxGoroutines)
	}
	return nil
}

// DefaultConfig returns default configuration
func DefaultConfig() *RuntimeStatsConfig {
	return &RuntimeStatsConfig{
		MaxMemory:     1024 * 1024 * 1024, // 1GB
		MaxCPU:        80.0,               // 80%
		MaxGoroutines: 10000,              // 10k goroutines
		Interval:      time.Second,        // 1 second
		EnableCPU:     true,
		EnableMemory:  true,
		EnableGC:      true,
	}
}

// RuntimeStats monitors runtime metrics like memory, CPU, goroutines and GC
type RuntimeStats struct {
	// Configuration
	config *RuntimeStatsConfig

	// Current usage
	currentMemory int64  // Using sync/atomic
	currentCPU    uint64 // Using uint64 to store float64 bit pattern
	goroutines    int32  // Using sync/atomic

	// Peak usage
	peakMemory  int64
	peakCPU     uint64
	peakThreads int32

	// Monitoring state
	enabled atomic.Bool // Using atomic.Bool instead of bool + mutex

	// Context
	ctx    context.Context
	cancel context.CancelFunc
}

// Metrics tracks runtime metrics
type Metrics struct {
	Memory     int64   `json:"memory"`     // Memory usage in bytes
	CPU        float64 `json:"cpu"`        // CPU usage percentage
	Goroutines int32   `json:"goroutines"` // Number of goroutines
	HeapAlloc  uint64  `json:"heap_alloc"` // Heap allocation in bytes
	HeapSys    uint64  `json:"heap_sys"`   // Heap system memory in bytes
	HeapIdle   uint64  `json:"heap_idle"`  // Heap idle memory in bytes
	GCCount    uint32  `json:"gc_count"`   // Number of completed GC cycles
	GCPause    uint64  `json:"gc_pause"`   // Total GC pause in nanoseconds
}

// NewRuntimeStats creates a new runtime monitor
//
// Usage:
//
//	ctx := context.Background()
//	config := DefaultConfig()
//	config.MaxMemory = 2 * 1024 * 1024 * 1024  // 2GB
//
//	monitor := NewRuntimeStats(ctx, config)
//	monitor.Start()
//	defer monitor.Stop()
//
//	// Get current usage
//	usage := monitor.GetMetrics()
func NewRuntimeStats(ctx context.Context, config *RuntimeStatsConfig) (*RuntimeStats, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	monitorCtx, cancel := context.WithCancel(ctx)

	return &RuntimeStats{
		config: config,
		ctx:    monitorCtx,
		cancel: cancel,
	}, nil
}

// Start starts runtime monitoring
func (m *RuntimeStats) Start() {
	m.enabled.Store(true)
	go m.monitorRuntime()
}

// Stop stops runtime monitoring
func (m *RuntimeStats) Stop() {
	m.enabled.Store(false)
	m.cancel()
}

// IsEnabled returns if monitoring is enabled
func (m *RuntimeStats) IsEnabled() bool {
	return m.enabled.Load()
}

// monitorRuntime periodically monitors runtime usage
func (m *RuntimeStats) monitorRuntime() {
	ticker := time.NewTicker(m.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			if !m.enabled.Load() {
				return
			}
			m.updateMetrics()
		}
	}
}

// float64ToUint64 converts float64 to uint64 for atomic operations
func float64ToUint64(f float64) uint64 {
	return math.Float64bits(f)
}

// uint64ToFloat64 converts uint64 back to float64
func uint64ToFloat64(u uint64) float64 {
	return math.Float64frombits(u)
}

// updateMetrics updates current runtime metrics
func (m *RuntimeStats) updateMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Update memory usage if enabled
	if m.config.EnableMemory {
		currentMem := int64(memStats.Alloc)
		atomic.StoreInt64(&m.currentMemory, currentMem)

		peakMem := atomic.LoadInt64(&m.peakMemory)
		if currentMem > peakMem {
			atomic.StoreInt64(&m.peakMemory, currentMem)
		}
	}

	// Update CPU usage if enabled
	if m.config.EnableCPU {
		cpuUsage := m.calculateCPUUsage()
		currentCPU := float64ToUint64(cpuUsage)
		atomic.StoreUint64(&m.currentCPU, currentCPU)

		peakCPU := atomic.LoadUint64(&m.peakCPU)
		if currentCPU > peakCPU {
			atomic.StoreUint64(&m.peakCPU, currentCPU)
		}
	}

	// Update goroutine count
	goroutines := int32(runtime.NumGoroutine())
	atomic.StoreInt32(&m.goroutines, goroutines)

	peakThreads := atomic.LoadInt32(&m.peakThreads)
	if goroutines > peakThreads {
		atomic.StoreInt32(&m.peakThreads, goroutines)
	}
}

// calculateCPUUsage calculates current CPU usage
func (m *RuntimeStats) calculateCPUUsage() float64 {
	var cpuStats runtime.MemStats
	runtime.ReadMemStats(&cpuStats)

	// Calculate CPU usage based on GC stats
	gcTime := float64(cpuStats.PauseTotalNs) / 1e9 // Convert to seconds
	timeSinceLastGC := time.Since(time.Unix(0, int64(cpuStats.LastGC))).Seconds()

	if timeSinceLastGC == 0 {
		return 0
	}

	return (gcTime / timeSinceLastGC) * 100
}

// GetMetrics gets current runtime metrics
func (m *RuntimeStats) GetMetrics() Metrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	usage := Metrics{
		Memory:     atomic.LoadInt64(&m.currentMemory),
		CPU:        uint64ToFloat64(atomic.LoadUint64(&m.currentCPU)),
		Goroutines: atomic.LoadInt32(&m.goroutines),
	}

	// Add GC stats if enabled
	if m.config.EnableGC {
		usage.HeapAlloc = memStats.HeapAlloc
		usage.HeapSys = memStats.HeapSys
		usage.HeapIdle = memStats.HeapIdle
		usage.GCCount = memStats.NumGC
		usage.GCPause = memStats.PauseTotalNs
	}

	return usage
}

// GetPeakUsage gets peak runtime usage
func (m *RuntimeStats) GetPeakUsage() Metrics {
	return Metrics{
		Memory:     atomic.LoadInt64(&m.peakMemory),
		CPU:        uint64ToFloat64(atomic.LoadUint64(&m.peakCPU)),
		Goroutines: atomic.LoadInt32(&m.peakThreads),
	}
}

// CheckThresholds checks if current usage exceeds thresholds
func (m *RuntimeStats) CheckThresholds() []error {
	var errors []error

	// Check memory if enabled
	if m.config.EnableMemory {
		currentMem := atomic.LoadInt64(&m.currentMemory)
		if currentMem > m.config.MaxMemory {
			errors = append(errors, fmt.Errorf("memory usage exceeded: %d > %d",
				currentMem, m.config.MaxMemory))
		}
	}

	// Check CPU if enabled
	if m.config.EnableCPU {
		currentCPU := uint64ToFloat64(atomic.LoadUint64(&m.currentCPU))
		if currentCPU > m.config.MaxCPU {
			errors = append(errors, fmt.Errorf("CPU usage exceeded: %f > %f",
				currentCPU, m.config.MaxCPU))
		}
	}

	// Check goroutines
	currentGoroutines := atomic.LoadInt32(&m.goroutines)
	if currentGoroutines > m.config.MaxGoroutines {
		errors = append(errors, fmt.Errorf("goroutine count exceeded: %d > %d",
			currentGoroutines, m.config.MaxGoroutines))
	}

	return errors
}

// Reset resets peak usage metrics
func (m *RuntimeStats) Reset() {
	atomic.StoreInt64(&m.peakMemory, 0)
	atomic.StoreUint64(&m.peakCPU, 0)
	atomic.StoreInt32(&m.peakThreads, 0)
}

// GetConfig returns current configuration
func (m *RuntimeStats) GetConfig() *RuntimeStatsConfig {
	return m.config
}
