package queue

import (
	"container/heap"
	"sync"
	"sync/atomic"
	"time"
)

// TimerMetrics tracks operational metrics for the timer queue
type TimerMetrics struct {
	EnqueueCount atomic.Int64
	DequeueCount atomic.Int64
	CancelCount  atomic.Int64
	TimeoutCount atomic.Int64
	OverdueCount atomic.Int64
}

// TimerQueue implements an optimized timer queue using min heap
type TimerQueue struct {
	items    *timerHeap
	capacity int
	lookup   map[string]*QueuedTask
	metrics  *TimerMetrics
	mu       sync.RWMutex
}

// timerHeap implements heap.Interface for QueuedTasks ordered by TriggerAt
type timerHeap []*QueuedTask

func (h *timerHeap) Len() int {
	return len(*h)
}

func (h *timerHeap) Less(i, j int) bool {
	return (*h)[i].TriggerAt.Before((*h)[j].TriggerAt)
}

func (h *timerHeap) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

func (h *timerHeap) Push(x any) {
	*h = append(*h, x.(*QueuedTask))
}

func (h *timerHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

// NewTimerQueue creates a new timer queue with the specified capacity
//
// Usage:
//
// // Initialize a timer queue with capacity 1000
// tq := NewTimerQueue(1000)
//
// // Add a timer task
//
//	task := &QueuedTask{
//	    ID:        "timer-1",
//	    TriggerAt: time.Now().Add(time.Minute),
//	    Data:      "timer data",
//	}
//
// tq.Push(task)
//
// // Retrieve the duration until the next task is due
// nextDue := tq.NextDue()
// fmt.Printf("Next task due in: %v\n", nextDue)
func NewTimerQueue(capacity int) *TimerQueue {
	if capacity <= 0 {
		capacity = 1000 // default capacity
	}

	return &TimerQueue{
		items:    &timerHeap{},
		capacity: capacity,
		lookup:   make(map[string]*QueuedTask),
		metrics:  &TimerMetrics{},
	}
}

// Push adds a task to the timer queue
func (tq *TimerQueue) Push(task *QueuedTask) error {
	if task == nil {
		return ErrInvalidTask
	}

	if task.ID == "" {
		return ErrInvalidTask
	}

	if task.TriggerAt.IsZero() {
		return ErrInvalidTask
	}

	tq.mu.Lock()
	defer tq.mu.Unlock()

	// Check capacity
	if len(*tq.items) >= tq.capacity {
		return ErrQueueFull
	}

	// Check duplicate
	if _, exists := tq.lookup[task.ID]; exists {
		return ErrTaskExists
	}

	// Initialize task status
	task.Status = TaskStatusPending

	// Add to heap and lookup
	heap.Push(tq.items, task)
	tq.lookup[task.ID] = task

	// Update metrics
	tq.metrics.EnqueueCount.Add(1)
	if task.TriggerAt.Before(time.Now()) {
		tq.metrics.OverdueCount.Add(1)
	}

	return nil
}

// Pop removes and returns the earliest due task
func (tq *TimerQueue) Pop() *QueuedTask {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	if len(*tq.items) == 0 {
		return nil
	}

	task := heap.Pop(tq.items).(*QueuedTask)
	delete(tq.lookup, task.ID)
	tq.metrics.DequeueCount.Add(1)

	return task
}

// Cancel removes a task by ID and returns true if successful
func (tq *TimerQueue) Cancel(taskID string) bool {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	task, exists := tq.lookup[taskID]
	if !exists {
		return false
	}

	// Find and remove the task from heap
	for i, t := range *tq.items {
		if t.ID == taskID {
			heap.Remove(tq.items, i)
			delete(tq.lookup, taskID)
			task.Status = TaskStatusCanceled
			tq.metrics.CancelCount.Add(1)
			return true
		}
	}
	return false
}

// DueTasks returns all tasks that are due for execution
func (tq *TimerQueue) DueTasks() []*QueuedTask {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	var tasks []*QueuedTask
	now := time.Now()

	for len(*tq.items) > 0 && (*tq.items)[0].TriggerAt.Before(now) {
		task := heap.Pop(tq.items).(*QueuedTask)
		delete(tq.lookup, task.ID)
		tasks = append(tasks, task)
		tq.metrics.DequeueCount.Add(1)
	}

	return tasks
}

// GetTask returns a task by ID without removing it
func (tq *TimerQueue) GetTask(taskID string) *QueuedTask {
	tq.mu.RLock()
	defer tq.mu.RUnlock()

	return tq.lookup[taskID]
}

// Peek returns the next due task without removing it
func (tq *TimerQueue) Peek() *QueuedTask {
	tq.mu.RLock()
	defer tq.mu.RUnlock()

	if len(*tq.items) == 0 {
		return nil
	}
	return (*tq.items)[0]
}

// NextDue returns the duration until the next task is due
// Returns -1 if queue is empty
func (tq *TimerQueue) NextDue() time.Duration {
	tq.mu.RLock()
	defer tq.mu.RUnlock()

	if len(*tq.items) == 0 {
		return time.Duration(-1)
	}

	next := (*tq.items)[0].TriggerAt
	now := time.Now()
	if next.Before(now) {
		return 0
	}
	return next.Sub(now)
}

// Clear removes all tasks from the queue
func (tq *TimerQueue) Clear() {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	*tq.items = (*tq.items)[:0]
	tq.lookup = make(map[string]*QueuedTask)
}

// Len returns the current number of tasks in queue
func (tq *TimerQueue) Len() int {
	tq.mu.RLock()
	defer tq.mu.RUnlock()
	return len(*tq.items)
}

// GetMetrics returns current queue metrics
func (tq *TimerQueue) GetMetrics() map[string]int64 {
	return map[string]int64{
		"enqueue_count": tq.metrics.EnqueueCount.Load(),
		"dequeue_count": tq.metrics.DequeueCount.Load(),
		"cancel_count":  tq.metrics.CancelCount.Load(),
		"timeout_count": tq.metrics.TimeoutCount.Load(),
		"overdue_count": tq.metrics.OverdueCount.Load(),
		"queue_length":  int64(tq.Len()),
	}
}
