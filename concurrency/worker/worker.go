package worker

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var ErrQueueFull = errors.New("task queue is full")

// Config represents pool configuration
type Config struct {
	MaxWorkers  int           // maximum number of workers
	QueueSize   int           // task queue size
	TaskTimeout time.Duration // timeout for single task
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		MaxWorkers:  10,          // default 10 workers
		QueueSize:   1000,        // default queue size 1000
		TaskTimeout: time.Minute, // default timeout 1 minute
	}
}

// Validate validates configuration
func (cfg *Config) Validate() error {
	if cfg.MaxWorkers < 1 {
		return errors.New("max workers must be greater than 0")
	}
	if cfg.QueueSize < 1 {
		return errors.New("queue size must be greater than 0")
	}
	if cfg.TaskTimeout < 0 {
		return errors.New("task timeout must be greater than or equal to 0")
	}
	return nil
}

// Processor represents a task processor
type Processor interface {
	Process(task any) error
}

// defaultProcessor provides default task processing logic
type defaultProcessor struct{}

func (p *defaultProcessor) Process(task any) error {
	switch t := task.(type) {
	case func() error:
		return t()
	case func():
		t()
		return nil
	default:
		return errors.New("unsupported task type")
	}
}

// Metrics tracks pool's operational metrics
type Metrics struct {
	ActiveWorkers  atomic.Int64
	PendingTasks   atomic.Int64
	CompletedTasks atomic.Int64
	FailedTasks    atomic.Int64
	ProcessingTime atomic.Int64 // nanoseconds
}

// Reset resets all metrics to zero
func (m *Metrics) Reset() {
	m.ActiveWorkers.Store(0)
	m.PendingTasks.Store(0)
	m.CompletedTasks.Store(0)
	m.FailedTasks.Store(0)
	m.ProcessingTime.Store(0)
}

// Pool represents a worker pool
type Pool struct {
	// Configuration
	maxWorkers  int
	queueSize   int
	taskTimeout time.Duration
	processor   Processor

	// Runtime components
	tasks  chan any
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Metrics
	metrics *Metrics
}

// NewPool creates a new worker pool
//
// Usage:
//
//	// Create a task processor
//	processor := func(task any) error {
//	    // Process the task here
//	    // ...
//	    return nil
//	}
//
//	// Create a worker pool configuration
//	cfg := &worker.Config{
//	    MaxWorkers:  10,   // Maximum number of worker goroutines
//	    QueueSize:   100,  // Maximum number of tasks that can be queued
//	    TaskTimeout: time.Minute, // Timeout for a single task execution
//	}
//
//	// Create a new worker pool
//	pool := worker.NewPool(cfg, processor)
//
//	// Start the worker pool
//	pool.Start()
//
//	// Submit tasks to the pool
//	err := pool.Submit("task1")
//	if err != nil {
//	    log.Printf("Failed to submit task: %v", err)
//	}
//
//	err = pool.Submit("task2")
//	if err != nil {
//	    log.Printf("Failed to submit task: %v", err)
//	}
//
//	// Get pool metrics
//	metrics := pool.GetMetrics()
//	log.Printf("Pool metrics: %+v", metrics)
//
//	// Check if the pool is busy
//	if pool.IsBusy() {
//	    log.Println("Pool is busy")
//	}
//
//	// Stop the worker pool
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	pool.Stop(ctx)
func NewPool(cfg *Config, processors ...Processor) *Pool {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Set default processor if none provided
	var processor Processor
	if len(processors) > 0 {
		processor = processors[0]
	} else {
		processor = &defaultProcessor{}
	}

	return &Pool{
		maxWorkers:  cfg.MaxWorkers,
		queueSize:   cfg.QueueSize,
		taskTimeout: cfg.TaskTimeout,
		processor:   processor,
		tasks:       make(chan any, cfg.QueueSize),
		ctx:         ctx,
		cancel:      cancel,
		metrics:     &Metrics{},
	}
}

// Start starts the worker pool
func (p *Pool) Start() {
	for i := 0; i < p.maxWorkers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

// Stop stops the worker pool
func (p *Pool) Stop(ctx context.Context) {
	p.cancel()
	close(p.tasks)

	// Wait for all workers to finish with timeout
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return
	case <-ctx.Done():
		return // timeout or cancelled
	}
}

// Submit submits a task to the pool
func (p *Pool) Submit(task any) error {
	select {
	case p.tasks <- task:
		p.metrics.PendingTasks.Add(1)
		return nil
	default:
		return ErrQueueFull
	}
}

// worker represents a worker goroutine
func (p *Pool) worker() {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return
		case task, ok := <-p.tasks:
			if !ok {
				return
			}
			p.processTask(task)
		}
	}
}

// processTask processes a single task
func (p *Pool) processTask(task any) {
	start := time.Now()
	p.metrics.ActiveWorkers.Add(1)
	p.metrics.PendingTasks.Add(-1)

	defer func() {
		p.metrics.ActiveWorkers.Add(-1)
		p.metrics.ProcessingTime.Add(time.Since(start).Nanoseconds())

		if r := recover(); r != nil {
			p.metrics.FailedTasks.Add(1)
		}
	}()

	// Create task context with timeout
	taskCtx, cancel := context.WithTimeout(p.ctx, p.taskTimeout)
	defer cancel()

	// Process task with context
	doneCh := make(chan error, 1)
	go func() {
		doneCh <- p.processor.Process(task)
	}()

	// Wait for completion or timeout
	select {
	case err := <-doneCh:
		if err != nil {
			p.metrics.FailedTasks.Add(1)
		} else {
			p.metrics.CompletedTasks.Add(1)
		}
	case <-taskCtx.Done():
		p.metrics.FailedTasks.Add(1)
	}
}

// GetMetrics returns the current metrics
func (p *Pool) GetMetrics() map[string]int64 {
	return map[string]int64{
		"active_workers":  p.metrics.ActiveWorkers.Load(),
		"pending_tasks":   p.metrics.PendingTasks.Load(),
		"completed_tasks": p.metrics.CompletedTasks.Load(),
		"failed_tasks":    p.metrics.FailedTasks.Load(),
		"processing_time": p.metrics.ProcessingTime.Load(),
	}
}

// IsBusy returns whether the pool is busy
func (p *Pool) IsBusy() bool {
	return p.metrics.ActiveWorkers.Load() >= int64(p.maxWorkers) ||
		p.metrics.PendingTasks.Load() >= int64(p.queueSize)
}

// IsIdle returns whether the pool is idle
func (p *Pool) IsIdle() bool {
	return p.metrics.ActiveWorkers.Load() == 0
}

// IsEmpty returns whether the pool is empty
func (p *Pool) IsEmpty() bool {
	return p.metrics.PendingTasks.Load() == 0
}
