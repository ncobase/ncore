package queue

import (
	"context"
	"errors"
	"fmt"
	"github.com/ncobase/ncore/pkg/worker"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrInvalidTask      = errors.New("invalid task")
	ErrQueueFull        = errors.New("task queue is full")
	ErrTaskTimeout      = errors.New("task execution timeout")
	ErrDuplicateTask    = errors.New("duplicate task ID")
	ErrTaskExists       = errors.New("task already exists")
	ErrTemporaryFailure = errors.New("temporary failure")
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCanceled  TaskStatus = "canceled"
	TaskStatusRetrying  TaskStatus = "retrying"
)

// TaskProcessor represents the actual task execution logic
type TaskProcessor interface {
	Process(task *QueuedTask) error
}

// defaultProcessor provides default task processing logic
type defaultProcessor struct{}

func (p *defaultProcessor) Process(task *QueuedTask) error {
	switch t := task.Data.(type) {
	case func() error:
		return t()
	case func():
		t()
		return nil
	default:
		return fmt.Errorf("unsupported task data type: %T", task.Data)
	}
}

// QueuedTask represents a task in queue
type QueuedTask struct {
	ID        string
	Type      string
	Priority  int
	Data      any
	TriggerAt time.Time

	// Retry configuration
	RetryCount    int
	MaxRetries    int
	RetryDelay    time.Duration
	LastRetryTime time.Time

	// Runtime status
	Status    TaskStatus
	LastError error
	Context   context.Context
	Cancel    context.CancelFunc
}

// Config represents queue configuration
type Config struct {
	// Worker pool configuration
	Workers     int
	QueueSize   int
	TaskTimeout time.Duration

	// Queue specific configuration
	MaxRetries int
	RetryDelay time.Duration

	// Priority configuration
	MaxPriority int // Maximum priority level (0 to MaxPriority)

	// Monitoring configuration
	MetricsWindow time.Duration
}

// Option defines the function type for queue options
type Option func(*TaskQueue)

// WithProcessor sets a custom task processor
func WithProcessor(processor TaskProcessor) Option {
	return func(q *TaskQueue) {
		if processor != nil {
			q.processor = processor
		}
	}
}

// WithConfig sets the queue configuration
func WithConfig(cfg *Config) Option {
	return func(q *TaskQueue) {
		if cfg != nil {
			if err := cfg.Validate(); err == nil {
				q.config = cfg
			}
		}
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Workers <= 0 {
		return errors.New("workers must be greater than 0")
	}
	if c.QueueSize <= 0 {
		return errors.New("queue size must be greater than 0")
	}
	if c.TaskTimeout <= 0 {
		return errors.New("task timeout must be greater than 0")
	}
	if c.MaxRetries < 0 {
		return errors.New("max retries must be greater than or equal to 0")
	}
	if c.RetryDelay <= 0 {
		return errors.New("retry delay must be greater than 0")
	}
	return nil
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Workers:       10,
		QueueSize:     1000,
		TaskTimeout:   time.Minute,
		MaxRetries:    3,
		RetryDelay:    time.Second * 5,
		MaxPriority:   10,
		MetricsWindow: time.Minute,
	}
}

// Metrics represents queue metrics
type Metrics struct {
	EnqueueCount   atomic.Int64
	DequeueCount   atomic.Int64
	ProcessCount   atomic.Int64
	SuccessCount   atomic.Int64
	FailureCount   atomic.Int64
	RetryCount     atomic.Int64
	CancelCount    atomic.Int64
	ProcessingTime atomic.Int64 // nanoseconds
	WaitTime       atomic.Int64 // nanoseconds
	QueueLength    atomic.Int64
	ActiveTasks    atomic.Int64
}

// TaskQueue represents a unified task queue
type TaskQueue struct {
	// Sub queues
	normalQueue   chan *QueuedTask
	priorityQueue *PriorityQueue
	timerQueue    *TimerQueue

	// Task execution
	processor TaskProcessor

	// Worker pool
	workerPool *worker.Pool

	// Configuration
	config *Config

	// Runtime components
	metrics    *Metrics
	processing sync.Map
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewTaskQueue creates a new task queue
//
// Usage:
//
//	// Define a custom task
//	type CustomTask struct {
//	    Name string
//	    Data interface{}
//	}
//
//	// Implement TaskProcessor for CustomTask
//	type CustomTaskProcessor struct{}
//
//	func (p *CustomTaskProcessor) Process(task *queue.QueuedTask) error {
//	    // Convert task data to CustomTask
//	    customTask, ok := task.Data.(*CustomTask)
//	    if !ok {
//	        return fmt.Errorf("invalid task data type")
//	    }
//
//	    // Process the task
//	    fmt.Printf("Processing task: %s with data: %v", customTask.Name, customTask.Data)
//	    time.Sleep(time.Second) // Simulate task execution
//	    return nil
//	}
//
//	// Create queue with default processor (handles func() and func() error)
//	q1 := NewTaskQueue(context.Background())
//
//	// Add a simple task using default processor
//	err := q1.Push(&QueuedTask{
//	    ID: "func-task",
//	    Data: func() error {
//	        // Do some work
//	        return nil
//	    },
//	})
//
//	// Configuration for the task queue
//	cfg := &Config{
//	    Workers:     10,
//	    QueueSize:   1000,
//	    TaskTimeout: time.Minute,
//	    MaxPriority: 5,
//	    MaxRetries:  3,
//	    RetryDelay:  time.Second * 5,
//	}
//
//	// Create queue with custom processor and configuration
//	processor := &CustomTaskProcessor{}
//	q2 := NewTaskQueue(context.Background(),
//	    WithProcessor(processor),
//	    WithConfig(cfg),
//	)
//
//	// Start the task queue
//	q2.Start()
//
//	// Example tasks
//	normalTask := &QueuedTask{
//	    ID:   "task-1",
//	    Type: "normal",
//	    Data: &CustomTask{
//	        Name: "Normal Task",
//	        Data: "some data",
//	    },
//	}
//
//	priorityTask := &QueuedTask{
//	    ID:       "task-2",
//	    Type:     "priority",
//	    Priority: 5,
//	    Data: &CustomTask{
//	        Name: "Priority Task",
//	        Data: "priority data",
//	    },
//	}
//
//	timerTask := &QueuedTask{
//	    ID:        "task-3",
//	    Type:      "timer",
//	    TriggerAt: time.Now().Add(time.Minute),
//	    Data: &CustomTask{
//	        Name: "Timer Task",
//	        Data: "timer data",
//	    },
//	}
//
//	// Add tasks to the queue
//	if err := q2.Push(normalTask); err != nil {
//	    fmt.Printf("Failed to push normal task: %v", err)
//	}
//
//	if err := q2.Push(priorityTask); err != nil {
//	    fmt.Printf("Failed to push priority task: %v", err)
//	}
//
//	if err := q2.Push(timerTask); err != nil {
//	    fmt.Printf("Failed to push timer task: %v", err)
//	}
//
//	// Retrieve queue metrics
//	metrics := q2.GetMetrics()
//
//	// Stop the task queue with a timeout
//	stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	q2.Stop(stopCtx)
func NewTaskQueue(ctx context.Context, opts ...Option) *TaskQueue {
	ctx, cancel := context.WithCancel(ctx)

	// Initialize with defaults
	q := &TaskQueue{
		processor: &defaultProcessor{},
		config:    DefaultConfig(),
		metrics:   &Metrics{},
		ctx:       ctx,
		cancel:    cancel,
	}

	// Apply options
	for _, opt := range opts {
		opt(q)
	}

	// Initialize queues
	q.normalQueue = make(chan *QueuedTask, q.config.QueueSize)
	q.priorityQueue = NewPriorityQueue(q.config.QueueSize, q.config.MaxPriority)
	q.timerQueue = NewTimerQueue(q.config.QueueSize)

	// Initialize worker pool
	workerCfg := &worker.Config{
		MaxWorkers:  q.config.Workers,
		QueueSize:   q.config.QueueSize,
		TaskTimeout: q.config.TaskTimeout,
	}
	q.workerPool = worker.NewPool(workerCfg, q)

	return q
}

// Process implements TaskProcessor interface for worker pool
func (q *TaskQueue) Process(task any) error {
	qTask, ok := task.(*QueuedTask)
	if !ok {
		return ErrInvalidTask
	}
	return q.executeTask(qTask)
}

// Start starts the queue
func (q *TaskQueue) Start() {
	// Start worker pool
	q.workerPool.Start()

	// Start timer processor
	go q.processTimers()

	// Start metrics collector
	go q.collectMetrics()
}

// Stop stops the queue
func (q *TaskQueue) Stop(ctx context.Context) error {
	q.cancel()
	close(q.normalQueue)
	q.workerPool.Stop(ctx)
	return nil
}

// Push pushes a task to queue
func (q *TaskQueue) Push(task *QueuedTask) error {
	if task == nil {
		return ErrInvalidTask
	}

	// Set initial task status
	task.Status = TaskStatusPending

	// Create task context
	taskCtx, taskCancel := context.WithCancel(q.ctx)
	task.Context = taskCtx
	task.Cancel = taskCancel

	// Set retry configuration if not specified
	if task.MaxRetries == 0 {
		task.MaxRetries = q.config.MaxRetries
	}
	if task.RetryDelay == 0 {
		task.RetryDelay = q.config.RetryDelay
	}

	startTime := time.Now()
	q.metrics.EnqueueCount.Add(1)

	var err error
	switch {
	case !task.TriggerAt.IsZero():
		err = q.timerQueue.Push(task)
	case task.Priority > 0:
		err = q.priorityQueue.Push(task)
	default:
		select {
		case q.normalQueue <- task:
			err = nil
		default:
			err = ErrQueueFull
		}
	}

	if err == nil {
		q.metrics.WaitTime.Add(time.Since(startTime).Nanoseconds())
		q.metrics.QueueLength.Add(1)
	}

	return err
}

// PushBatch pushes multiple tasks to queue
func (q *TaskQueue) PushBatch(tasks []*QueuedTask) []error {
	errs := make([]error, len(tasks))
	for i, task := range tasks {
		errs[i] = q.Push(task)
	}
	return errs
}

// Cancel cancels a task by ID
func (q *TaskQueue) Cancel(taskID string) bool {
	if taskData, ok := q.processing.Load(taskID); ok {
		if task, ok := taskData.(*QueuedTask); ok {
			task.Cancel()
			task.Status = TaskStatusCanceled
			q.metrics.CancelCount.Add(1)
			return true
		}
	}
	return false
}

// executeTask executes a single task
func (q *TaskQueue) executeTask(task *QueuedTask) error {
	start := time.Now()
	task.Status = TaskStatusRunning
	q.metrics.ProcessCount.Add(1)
	q.metrics.ActiveTasks.Add(1)
	defer q.metrics.ActiveTasks.Add(-1)

	// Store task in processing map
	q.processing.Store(task.ID, task)
	defer q.processing.Delete(task.ID)

	// Execute task (to be implemented by user)
	err := q.processTaskWithRetry(task)

	if err != nil {
		task.LastError = err
		task.Status = TaskStatusFailed
		q.metrics.FailureCount.Add(1)
	} else {
		task.Status = TaskStatusCompleted
		q.metrics.SuccessCount.Add(1)
	}

	q.metrics.ProcessingTime.Add(time.Since(start).Nanoseconds())
	return err
}

// processTaskWithRetry processes a task with retry mechanism
func (q *TaskQueue) processTaskWithRetry(task *QueuedTask) error {
	var lastErr error

	for attempt := 0; attempt <= task.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-time.After(task.RetryDelay):
			case <-task.Context.Done():
				return context.Canceled
			}

			task.RetryCount++
			task.LastRetryTime = time.Now()
			task.Status = TaskStatusRetrying
			q.metrics.RetryCount.Add(1)
		}

		// Execute task
		err := q.processor.Process(task)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if we should retry
		if !isRetryableError(lastErr) {
			return lastErr
		}
	}

	return lastErr
}

// RetryableErrors are errors that can be retried
var RetryableErrors = []error{
	context.DeadlineExceeded,
	ErrTaskTimeout,
	ErrTemporaryFailure,
}

// RetryableError represents an error that can be retried
type RetryableError interface {
	error
	Temporary() bool // Returns true if the error is temporary and can be retried
}

// isRetryableError determines if an error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	// Check if error is in the predefined retryable errors list
	for _, retryableErr := range RetryableErrors {
		if errors.Is(err, retryableErr) {
			return true
		}
	}

	// Check if error implements RetryableError interface
	var retryErr RetryableError
	if errors.As(err, &retryErr) {
		return retryErr.Temporary()
	}

	// Check if error is network or I/O related
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Temporary()
	}

	// Check error type for specific conditions
	switch {
	case strings.Contains(err.Error(), "connection refused"):
		return true
	case strings.Contains(err.Error(), "no such host"):
		return false // DNS errors are usually not retryable
	case strings.Contains(err.Error(), "timeout"):
		return true
	}

	// Default to not retrying for unknown errors
	return false
}

// processTimers processes timer queue
func (q *TaskQueue) processTimers() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-q.ctx.Done():
			return
		case <-ticker.C:
			tasks := q.timerQueue.DueTasks()
			for _, task := range tasks {
				if err := q.workerPool.Submit(task); err != nil {
					// Handle submission error
					task.Status = TaskStatusFailed
					task.LastError = err
					q.metrics.FailureCount.Add(1)
				}
			}
		}
	}
}

// collectMetrics collects queue metrics periodically
func (q *TaskQueue) collectMetrics() {
	ticker := time.NewTicker(q.config.MetricsWindow)
	defer ticker.Stop()

	for {
		select {
		case <-q.ctx.Done():
			return
		case <-ticker.C:
			// Update queue length
			q.metrics.QueueLength.Store(int64(len(q.normalQueue) +
				q.priorityQueue.Len() + q.timerQueue.Len()))
		}
	}
}

// GetMetrics returns current queue metrics
func (q *TaskQueue) GetMetrics() map[string]int64 {
	return map[string]int64{
		"enqueue_count":   q.metrics.EnqueueCount.Load(),
		"dequeue_count":   q.metrics.DequeueCount.Load(),
		"process_count":   q.metrics.ProcessCount.Load(),
		"success_count":   q.metrics.SuccessCount.Load(),
		"failure_count":   q.metrics.FailureCount.Load(),
		"retry_count":     q.metrics.RetryCount.Load(),
		"cancel_count":    q.metrics.CancelCount.Load(),
		"processing_time": q.metrics.ProcessingTime.Load(),
		"wait_time":       q.metrics.WaitTime.Load(),
		"queue_length":    q.metrics.QueueLength.Load(),
		"active_tasks":    q.metrics.ActiveTasks.Load(),
	}
}

// IsBusy returns whether the queue is currently busy
func (q *TaskQueue) IsBusy() bool {
	return q.workerPool.IsBusy()
}

// IsEmpty returns whether the queue is empty
func (q *TaskQueue) IsEmpty() bool {
	return q.workerPool.IsEmpty()
}
