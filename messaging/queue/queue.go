package queue

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ncobase/ncore/concurrency/worker"
)

var (
	ErrInvalidTask      = errors.New("invalid task")
	ErrQueueFull        = errors.New("task queue is full")
	ErrTaskTimeout      = errors.New("task execution timeout")
	ErrDuplicateTask    = errors.New("duplicate task ID")
	ErrTaskExists       = errors.New("task already exists")
	ErrTemporaryFailure = errors.New("temporary failure")
	ErrQueueStopped     = errors.New("task queue is stopped")
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
	priorityReady chan struct{}
	timerReady    chan struct{}

	// Task execution
	processor TaskProcessor

	// Worker pool
	workerPool *worker.Pool

	// Configuration
	config *Config

	// Runtime components
	metrics    *Metrics
	processing sync.Map
	tasks      sync.Map
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.RWMutex
	startOnce  sync.Once
	stopOnce   sync.Once
	stopped    bool
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
	if ctx == nil {
		ctx = context.Background()
	}
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
	q.priorityReady = make(chan struct{}, 1)
	q.timerReady = make(chan struct{}, 1)

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
	q.startOnce.Do(func() {
		q.mu.RLock()
		stopped := q.stopped
		q.mu.RUnlock()
		if stopped {
			return
		}

		q.workerPool.Start()

		q.wg.Add(3)
		go q.dispatchReadyTasks()
		go q.processTimers()
		go q.collectMetrics()
	})
}

// Stop stops the queue
func (q *TaskQueue) Stop(ctx context.Context) error {
	q.stopOnce.Do(func() {
		q.mu.Lock()
		q.stopped = true
		q.cancel()
		q.cancelPendingTasks()
		close(q.normalQueue)
		q.mu.Unlock()
	})

	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}

	q.workerPool.Stop(ctx)
	return nil
}

func (q *TaskQueue) cancelPendingTasks() {
	var canceled int64
	q.tasks.Range(func(key, value any) bool {
		task, ok := value.(*QueuedTask)
		if !ok {
			q.tasks.Delete(key)
			return true
		}
		if task.Status == TaskStatusPending || task.Status == TaskStatusRetrying {
			task.Cancel()
			task.Status = TaskStatusCanceled
			task.LastError = context.Canceled
			q.tasks.Delete(key)
			canceled++
		}
		return true
	})
	if canceled > 0 {
		q.metrics.CancelCount.Add(canceled)
	}
	q.metrics.QueueLength.Store(0)
}

// Push pushes a task to queue
func (q *TaskQueue) Push(task *QueuedTask) error {
	if task == nil {
		return ErrInvalidTask
	}
	if task.ID == "" {
		return ErrInvalidTask
	}

	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.stopped || q.ctx.Err() != nil {
		return ErrQueueStopped
	}

	if _, loaded := q.tasks.LoadOrStore(task.ID, task); loaded {
		return ErrDuplicateTask
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

	var err error
	switch {
	case !task.TriggerAt.IsZero():
		err = q.timerQueue.Push(task)
		if err == nil {
			q.signalTimerReady()
		}
	case task.Priority > 0:
		err = q.priorityQueue.Push(task)
		if err == nil {
			q.signalPriorityReady()
		}
	default:
		select {
		case q.normalQueue <- task:
			err = nil
		case <-q.ctx.Done():
			err = ErrQueueStopped
		default:
			err = ErrQueueFull
		}
	}

	if err == nil {
		q.metrics.EnqueueCount.Add(1)
		q.metrics.WaitTime.Add(time.Since(startTime).Nanoseconds())
		q.metrics.QueueLength.Add(1)
		return nil
	}

	q.tasks.Delete(task.ID)
	task.Cancel()
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

	if taskData, ok := q.tasks.Load(taskID); ok {
		task, ok := taskData.(*QueuedTask)
		if !ok {
			q.tasks.Delete(taskID)
			return false
		}
		task.Cancel()
		task.Status = TaskStatusCanceled

		removed := false
		if !task.TriggerAt.IsZero() {
			removed = q.timerQueue.Cancel(taskID)
		} else if task.Priority > 0 {
			removed = q.priorityQueue.Cancel(taskID)
		}
		if removed {
			q.metrics.QueueLength.Add(-1)
		}

		q.tasks.Delete(taskID)
		q.metrics.CancelCount.Add(1)
		return true
	}
	return false
}

func (q *TaskQueue) dispatchReadyTasks() {
	defer q.wg.Done()

	for {
		for {
			task := q.priorityQueue.Pop()
			if task == nil {
				break
			}
			if !q.dispatchTask(task) {
				return
			}
		}

		select {
		case <-q.ctx.Done():
			return
		case <-q.priorityReady:
			continue
		case task, ok := <-q.normalQueue:
			if !ok {
				return
			}
			if !q.dispatchTask(task) {
				return
			}
		}
	}
}

func (q *TaskQueue) dispatchTask(task *QueuedTask) bool {
	if task == nil {
		return true
	}

	q.metrics.DequeueCount.Add(1)
	q.metrics.QueueLength.Add(-1)

	if task.Status == TaskStatusCanceled || task.Context.Err() != nil {
		task.Status = TaskStatusCanceled
		q.tasks.Delete(task.ID)
		return true
	}

	if err := q.workerPool.SubmitWait(q.ctx, task); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, worker.ErrPoolStopped) || errors.Is(err, ErrQueueStopped) {
			task.Status = TaskStatusCanceled
			task.LastError = err
			q.tasks.Delete(task.ID)
			return false
		}
		task.Status = TaskStatusFailed
		task.LastError = err
		q.tasks.Delete(task.ID)
		q.metrics.FailureCount.Add(1)
	}

	return true
}

func (q *TaskQueue) signalPriorityReady() {
	select {
	case q.priorityReady <- struct{}{}:
	default:
	}
}

func (q *TaskQueue) signalTimerReady() {
	select {
	case q.timerReady <- struct{}{}:
	default:
	}
}

// executeTask executes a single task
func (q *TaskQueue) executeTask(task *QueuedTask) error {
	if task.Context != nil && task.Context.Err() != nil {
		task.Status = TaskStatusCanceled
		q.tasks.Delete(task.ID)
		return context.Canceled
	}

	start := time.Now()
	task.Status = TaskStatusRunning
	q.metrics.ProcessCount.Add(1)
	q.metrics.ActiveTasks.Add(1)
	defer q.metrics.ActiveTasks.Add(-1)

	// Store task in processing map
	q.processing.Store(task.ID, task)
	defer func() {
		q.processing.Delete(task.ID)
		q.tasks.Delete(task.ID)
	}()

	// Execute task through the configured processor and retry policy.
	err := q.processTaskWithRetry(task)

	if task.Context != nil && task.Context.Err() != nil {
		task.LastError = task.Context.Err()
		task.Status = TaskStatusCanceled
		q.metrics.CancelCount.Add(1)
	} else if err != nil {
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
	defer q.wg.Done()

	for {
		tasks := q.timerQueue.DueTasks()
		for _, task := range tasks {
			if !q.dispatchTask(task) {
				return
			}
		}

		wait := q.timerQueue.NextDue()
		if wait < 0 {
			select {
			case <-q.ctx.Done():
				return
			case <-q.timerReady:
				continue
			}
		}
		if wait == 0 {
			continue
		}

		timer := time.NewTimer(wait)
		select {
		case <-q.ctx.Done():
			timer.Stop()
			return
		case <-q.timerReady:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
		case <-timer.C:
		}
	}
}

// collectMetrics collects queue metrics periodically
func (q *TaskQueue) collectMetrics() {
	defer q.wg.Done()

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
