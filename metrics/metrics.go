package metrics

import (
	"container/ring"
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Config represents metrics configuration
type Config struct {
	Enabled              bool          // Enable metrics collection
	FlushInterval        time.Duration // Interval to flush metrics
	MaxSamples           int           // Maximum samples for histograms
	EnableProcessMetrics bool          // Enable process level metrics
	EnableNodeMetrics    bool          // Enable node level metrics
	EnableTaskMetrics    bool          // Enable task level metrics
	EnableSystemMetrics  bool          // Enable system level metrics
	MetricsPrefix        string        // Prefix for all metric names
}

// DefaultConfig returns the default metrics configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:              true,
		FlushInterval:        time.Second * 10,
		MaxSamples:           1000,
		EnableProcessMetrics: true,
		EnableNodeMetrics:    true,
		EnableTaskMetrics:    true,
		EnableSystemMetrics:  true,
	}
}

// Validate validates the metrics configuration
func (c *Config) Validate() error {
	if c.FlushInterval <= 0 {
		return fmt.Errorf("flush interval must be greater than 0, got %v", c.FlushInterval)
	}

	if c.MaxSamples <= 0 {
		return fmt.Errorf("max samples must be greater than 0, got %d", c.MaxSamples)
	}

	return nil
}

// Collector represents the metrics collector
type Collector struct {
	// Configuration
	config *Config

	// Core metrics components
	process      *ProcessMetrics
	node         *NodeMetrics
	task         *TaskMetrics
	system       *SystemMetrics
	retryMetrics *RetryMetrics

	// Runtime state
	startTime time.Time
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	mu        sync.RWMutex

	// Base metrics storage
	counters   map[string]*atomic.Int64
	gauges     map[string]*atomic.Int64
	histograms map[string]*Histogram

	// Registered metrics
	metrics map[string]any
	pool    sync.Pool
}

// NewCollector creates a new metrics collector
func NewCollector(cfg *Config) (*Collector, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if cfg == nil {
		cfg = DefaultConfig()
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid metrics config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := &Collector{
		config:       cfg,
		startTime:    time.Now(),
		ctx:          ctx,
		cancel:       cancel,
		counters:     make(map[string]*atomic.Int64),
		gauges:       make(map[string]*atomic.Int64),
		histograms:   make(map[string]*Histogram),
		metrics:      make(map[string]any),
		retryMetrics: newRetryMetrics(cfg.MaxSamples),
		pool: sync.Pool{
			New: func() any {
				return &strings.Builder{}
			},
		},
	}

	// Initialize metrics components
	if cfg.EnableProcessMetrics {
		c.process = newProcessMetrics(cfg.MaxSamples)
	}
	if cfg.EnableNodeMetrics {
		c.node = newNodeMetrics(cfg.MaxSamples)
	}
	if cfg.EnableTaskMetrics {
		c.task = newTaskMetrics(cfg.MaxSamples)
	}
	if cfg.EnableSystemMetrics {
		c.system = newSystemMetrics(cfg.MaxSamples)
	}

	return c, nil
}

// Start starts metrics collection
func (c *Collector) Start(ctx context.Context) error {
	if !c.config.Enabled {
		return nil
	}

	// Create a new context for the collector
	collectorCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	// Start periodic flush
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.flushLoop(collectorCtx)
	}()

	return nil
}

// Stop stops metrics collection
func (c *Collector) Stop() {
	c.cancel()
	c.wg.Wait()
}

// flushLoop periodically flushes metrics
func (c *Collector) flushLoop(ctx context.Context) {
	ticker := time.NewTicker(c.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.flush()
		}
	}
}

// flush flushes current metrics
func (c *Collector) flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Reset histograms
	for _, h := range c.histograms {
		h.Reset()
	}
}

// Accessor methods

// Process returns the process metrics
func (c *Collector) Process() *ProcessMetrics {
	return c.process
}

// Node returns the node metrics
func (c *Collector) Node() *NodeMetrics {
	return c.node
}

// Task returns the task metrics
func (c *Collector) Task() *TaskMetrics {
	return c.task
}

// System returns the system metrics
func (c *Collector) System() *SystemMetrics {
	return c.system
}

// Basic metrics operations

// RegisterCounter registers a new counter metric
func (c *Collector) RegisterCounter(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	name = c.prefix(name)
	if _, exists := c.counters[name]; !exists {
		c.counters[name] = &atomic.Int64{}
	}
}

// RegisterGauge registers a new gauge metric
func (c *Collector) RegisterGauge(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	name = c.prefix(name)
	if _, exists := c.gauges[name]; !exists {
		c.gauges[name] = &atomic.Int64{}
	}
}

// RegisterHistogram registers a new histogram metric
func (c *Collector) RegisterHistogram(name string, maxSamples int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	name = c.prefix(name)
	if _, exists := c.histograms[name]; !exists {
		c.histograms[name] = NewHistogram(maxSamples)
	}
}

// RecordTaskStart records a task start event
func (c *Collector) RecordTaskStart() {
	if c.task != nil {
		c.task.RecordTaskStart()
	}
}

// RecordTaskCompletion records a task completion event
func (c *Collector) RecordTaskCompletion(duration float64, cancelled bool) {
	if c.task != nil {
		c.task.RecordTaskCompletion(duration, cancelled)
	}
}

// RecordProcessStart records a process start event
func (c *Collector) RecordProcessStart() {
	if c.process != nil {
		c.process.RecordProcessStart()
	}
}

// RecordProcessCompletion records a process completion event
func (c *Collector) RecordProcessCompletion(duration float64, success bool) {
	if c.process != nil {
		c.process.RecordProcessCompletion(duration, success)
	}
}

// GetMetrics returns all metrics
func (c *Collector) GetMetrics() map[string]any {
	metrics := make(map[string]any)

	// Add process metrics
	if c.process != nil {
		metrics["process"] = c.process.getMetrics()
	}

	// Add node metrics
	if c.node != nil {
		metrics["node"] = c.node.getMetrics()
	}

	// Add task metrics
	if c.task != nil {
		metrics["task"] = c.task.getMetrics()
	}

	// Add system metrics
	if c.system != nil {
		metrics["system"] = c.system.getMetrics()
	}

	// Add runtime metrics
	metrics["runtime"] = map[string]any{
		"uptime":     time.Since(c.startTime).Seconds(),
		"start_time": c.startTime.Unix(),
	}

	return metrics
}

// ProcessMetrics tracks process-level metrics
type ProcessMetrics struct {
	// Counters
	totalProcesses     atomic.Int64
	activeProcesses    atomic.Int64
	completedProcesses atomic.Int64
	failedProcesses    atomic.Int64

	// Histograms
	processDurations *Histogram

	// States
	processStates sync.Map // string -> atomic.Int64
	mu            sync.RWMutex
}

func newProcessMetrics(maxSamples int) *ProcessMetrics {
	return &ProcessMetrics{
		processDurations: NewHistogram(maxSamples),
	}
}

// RecordProcessStart records process start
func (p *ProcessMetrics) RecordProcessStart() {
	p.totalProcesses.Add(1)
	p.activeProcesses.Add(1)
}

// RecordProcessCompletion records process completion
func (p *ProcessMetrics) RecordProcessCompletion(duration float64, success bool) {
	p.activeProcesses.Add(-1)
	if success {
		p.completedProcesses.Add(1)
	} else {
		p.failedProcesses.Add(1)
	}
	p.processDurations.Add(duration)
}

func (p *ProcessMetrics) getMetrics() map[string]any {
	return map[string]any{
		"total":     p.totalProcesses.Load(),
		"active":    p.activeProcesses.Load(),
		"completed": p.completedProcesses.Load(),
		"failed":    p.failedProcesses.Load(),
		"durations": p.processDurations.GetStats(),
	}
}

// NodeMetrics tracks node-level metrics
type NodeMetrics struct {
	// Execution counts
	totalNodes     atomic.Int64
	executionCount atomic.Int64
	successCount   atomic.Int64
	failureCount   atomic.Int64
	retryCount     atomic.Int64

	// Timing
	nodeDurations *Histogram

	// Node types
	nodeTypes sync.Map // string -> atomic.Int64
	mu        sync.RWMutex
}

func newNodeMetrics(maxSamples int) *NodeMetrics {
	return &NodeMetrics{
		nodeDurations: NewHistogram(maxSamples),
	}
}

// RecordNodeExecution records a node execution
func (n *NodeMetrics) RecordNodeExecution(nodeType string, duration float64, success bool, retried bool) {
	n.executionCount.Add(1)
	if success {
		n.successCount.Add(1)
	} else {
		n.failureCount.Add(1)
	}
	if retried {
		n.retryCount.Add(1)
	}
	n.nodeDurations.Add(duration)

	// Record node type
	if counter, ok := n.nodeTypes.Load(nodeType); ok {
		counter.(*atomic.Int64).Add(1)
	} else {
		counter := &atomic.Int64{}
		counter.Add(1)
		n.nodeTypes.Store(nodeType, counter)
	}
}

func (n *NodeMetrics) getMetrics() map[string]any {
	return map[string]any{
		"total":      n.totalNodes.Load(),
		"executions": n.executionCount.Load(),
		"successes":  n.successCount.Load(),
		"failures":   n.failureCount.Load(),
		"retries":    n.retryCount.Load(),
		"durations":  n.nodeDurations.GetStats(),
	}
}

// TaskMetrics tracks task-level metrics
type TaskMetrics struct {
	// Task counts
	totalTasks     atomic.Int64
	activeTasks    atomic.Int64
	completedTasks atomic.Int64
	cancelledTasks atomic.Int64

	// Task timing
	taskDurations *Histogram

	// Task states
	taskStates sync.Map // string -> atomic.Int64
	mu         sync.RWMutex
}

func newTaskMetrics(maxSamples int) *TaskMetrics {
	return &TaskMetrics{
		taskDurations: NewHistogram(maxSamples),
	}
}

// RecordTaskStart records task start
func (t *TaskMetrics) RecordTaskStart() {
	t.totalTasks.Add(1)
	t.activeTasks.Add(1)
}

// RecordTaskCompletion records task completion
func (t *TaskMetrics) RecordTaskCompletion(duration float64, cancelled bool) {
	t.activeTasks.Add(-1)
	if cancelled {
		t.cancelledTasks.Add(1)
	} else {
		t.completedTasks.Add(1)
	}
	t.taskDurations.Add(duration)
}

func (t *TaskMetrics) getMetrics() map[string]any {
	return map[string]any{
		"total":     t.totalTasks.Load(),
		"active":    t.activeTasks.Load(),
		"completed": t.completedTasks.Load(),
		"cancelled": t.cancelledTasks.Load(),
		"durations": t.taskDurations.GetStats(),
	}
}

// SystemMetrics tracks system-level metrics
type SystemMetrics struct {
	// System resources
	cpuUsage       atomic.Int64
	memoryUsage    atomic.Int64
	goroutineCount atomic.Int64

	// Request stats
	requestCount atomic.Int64
	errorCount   atomic.Int64
	responseTime *Histogram

	mu sync.RWMutex
}

func newSystemMetrics(maxSamples int) *SystemMetrics {
	return &SystemMetrics{
		responseTime: NewHistogram(maxSamples),
	}
}

// RecordRequest records a system request
func (s *SystemMetrics) RecordRequest(duration float64, isError bool) {
	s.requestCount.Add(1)
	if isError {
		s.errorCount.Add(1)
	}
	s.responseTime.Add(duration)
}

func (s *SystemMetrics) getMetrics() map[string]any {
	return map[string]any{
		"cpu":        float64(s.cpuUsage.Load()) / 100.0,
		"memory":     float64(s.memoryUsage.Load()) / 100.0,
		"goroutines": s.goroutineCount.Load(),
		"requests":   s.requestCount.Load(),
		"errors":     s.errorCount.Load(),
		"response":   s.responseTime.GetStats(),
	}
}

// RecordExecutorAttemptDuration records an executor attempt duration
func (c *Collector) RecordExecutorAttemptDuration(executorType string, duration time.Duration, err error) {
	labels := []Label{
		{Name: "executor", Value: executorType},
		{Name: "success", Value: fmt.Sprintf("%t", err == nil)},
	}
	c.RecordDuration("executor_attempt_duration", duration, labels...)

	if err == nil {
		c.AddCounter(c.WithLabels("executor_success_total",
			Label{Name: "executor", Value: executorType}), 1)
	} else {
		c.AddCounter(c.WithLabels("executor_failure_total",
			Label{Name: "executor", Value: executorType}), 1)
	}
}

// RecordExecutorRetryCount records an executor retry count
func (c *Collector) RecordExecutorRetryCount(executorType string, attempt int, err error) {
	labels := []Label{
		{Name: "executor", Value: executorType},
		{Name: "attempt", Value: fmt.Sprintf("%d", attempt)},
	}

	c.AddCounter(c.WithLabels("executor_retry_total", labels...), 1)

	if err != nil {
		errLabels := append(labels, Label{Name: "error", Value: err.Error()})
		c.AddCounter(c.WithLabels("executor_retry_error_total", errLabels...), 1)
	}
}

// RecordHandlerExecutionDuration records a handler execution duration
func (c *Collector) RecordHandlerExecutionDuration(handlerType string, duration time.Duration, err error) {
	labels := []Label{
		{Name: "handler", Value: handlerType},
		{Name: "success", Value: fmt.Sprintf("%t", err == nil)},
	}
	c.RecordDuration("handler_execution_duration", duration, labels...)

	if err != nil {
		errLabels := append(labels, Label{Name: "error", Value: err.Error()})
		c.AddCounter(c.WithLabels("handler_execution_error_total", errLabels...), 1)
	}
}

// RecordQueueOperationDuration records a queue operation duration
func (c *Collector) RecordQueueOperationDuration(operation string, duration time.Duration, err error) {
	labels := []Label{
		{Name: "operation", Value: operation},
		{Name: "success", Value: fmt.Sprintf("%t", err == nil)},
	}
	c.RecordDuration("queue_operation_duration", duration, labels...)

	if err != nil {
		errLabels := append(labels, Label{Name: "error", Value: err.Error()})
		c.AddCounter(c.WithLabels("queue_operation_error_total", errLabels...), 1)
	}
}

// RecordSchedulerOperationDuration records a scheduler operation duration
func (c *Collector) RecordSchedulerOperationDuration(operation string, duration time.Duration, err error) {
	labels := []Label{
		{Name: "operation", Value: operation},
		{Name: "success", Value: fmt.Sprintf("%t", err == nil)},
	}
	c.RecordDuration("scheduler_operation_duration", duration, labels...)

	if err != nil {
		errLabels := append(labels, Label{Name: "error", Value: err.Error()})
		c.AddCounter(c.WithLabels("scheduler_operation_error_total", errLabels...), 1)
	}
}

// RetryEvent represents a retry attempt event
type RetryEvent struct {
	Attempt  int
	Duration time.Duration
	Error    error
}

// RetrySuccessEvent represents a successful retry event
type RetrySuccessEvent struct {
	Attempts int
	Duration time.Duration
}

// RetryMetrics tracks retry-related metrics
type RetryMetrics struct {
	attempts       atomic.Int64
	successes      atomic.Int64
	failures       atomic.Int64
	lastRetryTime  atomic.Int64
	retryDurations *Histogram
	mu             sync.RWMutex
}

func newRetryMetrics(maxSamples int) *RetryMetrics {
	return &RetryMetrics{
		retryDurations: NewHistogram(maxSamples),
	}
}

// RecordRetry records a retry attempt
func (c *Collector) RecordRetry(event RetryEvent) {
	// Update retry counts
	c.retryMetrics.attempts.Add(1)
	if event.Error != nil {
		c.retryMetrics.failures.Add(1)
	}

	// Record duration
	c.retryMetrics.retryDurations.Add(event.Duration.Seconds())

	// Update last retry time
	c.retryMetrics.lastRetryTime.Store(time.Now().UnixNano())
}

// RecordRetrySuccess records a successful retry
func (c *Collector) RecordRetrySuccess(event RetrySuccessEvent) {
	c.retryMetrics.successes.Add(1)
	c.retryMetrics.retryDurations.Add(event.Duration.Seconds())
}

// Helper methods for metrics operations

// SetGauge sets a gauge value
func (c *Collector) SetGauge(name string, value float64) {
	c.mu.RLock()
	gauge, exists := c.gauges[name]
	c.mu.RUnlock()

	if exists {
		gauge.Store(int64(value))
	}
}

// SetCounter sets a counter value
func (c *Collector) SetCounter(name string, value int64) {
	c.mu.RLock()
	counter, exists := c.counters[name]
	c.mu.RUnlock()

	if exists {
		counter.Store(value)
	}
}

// AddCounter adds to counter value
func (c *Collector) AddCounter(name string, delta int64) {
	c.mu.RLock()
	counter, exists := c.counters[name]
	c.mu.RUnlock()

	if exists {
		counter.Add(delta)
	}
}

// GetCounter gets a counter value
func (c *Collector) GetCounter(name string) int64 {
	c.mu.RLock()
	counter, exists := c.counters[name]
	c.mu.RUnlock()

	if exists {
		return counter.Load()
	}
	return 0
}

// GetGauge gets a gauge value
func (c *Collector) GetGauge(name string) float64 {
	c.mu.RLock()
	gauge, exists := c.gauges[name]
	c.mu.RUnlock()

	if exists {
		return float64(gauge.Load())
	}
	return 0
}

// SetHistogram sets a histogram value
func (c *Collector) SetHistogram(name string, value float64) {
	c.mu.RLock()
	histogram, exists := c.histograms[name]
	c.mu.RUnlock()

	if exists {
		histogram.Add(value)
	}
}

// GetHistogram gets a histogram value
func (c *Collector) GetHistogram(name string) HistogramStats {
	c.mu.RLock()
	histogram, exists := c.histograms[name]
	c.mu.RUnlock()

	if exists {
		return histogram.GetStats()
	}
	return HistogramStats{}
}

// Histogram represents a histogram metric
type Histogram struct {
	mu         sync.RWMutex
	samples    *ring.Ring
	maxSamples int
	count      atomic.Int64
	sum        atomic.Int64
	min        atomic.Int64
	max        atomic.Int64
	buckets    []float64
}

// NewHistogram creates a new histogram
func NewHistogram(maxSamples int, buckets ...float64) *Histogram {
	if len(buckets) == 0 {
		buckets = []float64{50, 75, 90, 95, 99}
	}

	return &Histogram{
		samples:    ring.New(maxSamples),
		maxSamples: maxSamples,
		buckets:    buckets,
	}
}

// Add adds a value to the histogram
func (h *Histogram) Add(v float64) {
	h.count.Add(1)
	h.sum.Add(int64(v))

	// Update min
	for {
		curMin := math.Float64frombits(uint64(h.min.Load()))
		if curMin == 0 || v < curMin {
			if h.min.CompareAndSwap(int64(math.Float64bits(curMin)),
				int64(math.Float64bits(v))) {
				break
			}
		} else {
			break
		}
	}

	// Update max
	for {
		curMax := math.Float64frombits(uint64(h.max.Load()))
		if v > curMax {
			if h.max.CompareAndSwap(int64(math.Float64bits(curMax)),
				int64(math.Float64bits(v))) {
				break
			}
		} else {
			break
		}
	}

	h.mu.Lock()
	h.samples.Value = v
	h.samples = h.samples.Next()
	h.mu.Unlock()
}

// Reset resets the histogram
func (h *Histogram) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.samples = ring.New(h.maxSamples)
	h.count.Store(0)
	h.sum.Store(0)
	h.min.Store(0)
	h.max.Store(0)
}

// HistogramStats returns histogram statistics
type HistogramStats struct {
	Count       int64               `json:"count"`
	Min         float64             `json:"min"`
	Max         float64             `json:"max"`
	Mean        float64             `json:"mean"`
	Percentiles map[float64]float64 `json:"percentiles"`
}

// GetStats returns current histogram statistics
func (h *Histogram) GetStats() HistogramStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := h.count.Load()
	if count == 0 {
		return HistogramStats{}
	}

	var samples []float64
	h.samples.Do(func(v any) {
		if v != nil {
			samples = append(samples, v.(float64))
		}
	})

	if len(samples) == 0 {
		return HistogramStats{}
	}

	sort.Float64s(samples)

	stats := HistogramStats{
		Count:       count,
		Min:         math.Float64frombits(uint64(h.min.Load())),
		Max:         math.Float64frombits(uint64(h.max.Load())),
		Mean:        float64(h.sum.Load()) / float64(count),
		Percentiles: make(map[float64]float64),
	}

	// Calculate configured percentiles
	for _, p := range h.buckets {
		idx := int(float64(len(samples)) * p / 100)
		if idx >= len(samples) {
			idx = len(samples) - 1
		}
		stats.Percentiles[p] = samples[idx]
	}

	return stats
}

// Label represents a metric label
type Label struct {
	Name  string
	Value string
}

// WithLabels adds labels to a metric
func (c *Collector) WithLabels(name string, labels ...Label) string {
	sb := c.getStringBuilder()
	defer c.putStringBuilder(sb)

	sb.WriteString(name)

	if len(labels) > 0 {
		sb.WriteByte('{')
		for i, label := range labels {
			if i > 0 {
				sb.WriteByte(',')
			}
			sb.WriteString(label.Name)
			sb.WriteByte('=')
			sb.WriteString(label.Value)
		}
		sb.WriteByte('}')
	}

	return sb.String()
}

// RecordValue records a value with labels
func (c *Collector) RecordValue(name string, value float64, labels ...Label) {
	metricName := c.WithLabels(name, labels...)

	c.mu.RLock()
	if hist, ok := c.histograms[metricName]; ok {
		hist.Add(value)
	}
	c.mu.RUnlock()
}

// RecordDuration records a duration with labels
func (c *Collector) RecordDuration(name string, duration time.Duration, labels ...Label) {
	c.RecordValue(name, duration.Seconds(), labels...)
}

// Prefix returns a metric name with the configured prefix
func (c *Collector) prefix(name string) string {
	return c.config.MetricsPrefix + name
}

// GetStringBuilder returns a string builder from the pool
func (c *Collector) getStringBuilder() *strings.Builder {
	return c.pool.Get().(*strings.Builder)
}

// PutStringBuilder returns a string builder to the pool
func (c *Collector) putStringBuilder(sb *strings.Builder) {
	sb.Reset()
	c.pool.Put(sb)
}
