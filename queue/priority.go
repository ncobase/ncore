package queue

import (
	"sync"
	"sync/atomic"
	"time"
)

// PriorityLevel represents task priority level
type PriorityLevel struct {
	tasks     []*QueuedTask // Tasks at this priority level
	lastEmpty time.Time     // Last time this level became empty
}

// PriorityQueue implements an optimized priority queue
type PriorityQueue struct {
	// Queue configuration
	capacity int // Maximum total capacity
	maxPrio  int // Maximum priority level

	// Queue data
	levels []*PriorityLevel       // Priority levels
	lookup map[string]*QueuedTask // Task ID to task mapping
	size   int                    // Current total size

	// Statistics
	enqueueCount atomic.Int64
	dequeueCount atomic.Int64

	// Concurrency control
	mu sync.RWMutex
}

// NewPriorityQueue creates a new priority queue
//
// Usage:
//
// // Initialize a priority queue with capacity 1000 and maximum priority 10
// pq := NewPriorityQueue(1000, 10)
//
// // Add tasks to the queue
//
//	task1 := &QueuedTask{
//	    ID:       "task1",
//	    Priority: 5,
//	    Data:     "data1",
//	}
//
// pq.Push(task1)
//
//	task2 := &QueuedTask{
//	    ID:       "task2",
//	    Priority: 8,
//	    Data:     "data2",
//	}
//
// pq.Push(task2)
//
// // Retrieve the highest priority task
// highestTask := pq.Pop() // Returns task2
//
// // Update the priority of a specific task
// pq.UpdatePriority("task1", 9)
//
// // Retrieve a batch of tasks (up to 5)
// tasks := pq.PopBatch(5)
//
// // Retrieve queue metrics
// metrics := pq.GetMetrics()
func NewPriorityQueue(capacity int, maxPriority int) *PriorityQueue {
	if maxPriority < 1 {
		maxPriority = 1
	}

	pq := &PriorityQueue{
		capacity: capacity,
		maxPrio:  maxPriority,
		levels:   make([]*PriorityLevel, maxPriority+1),
		lookup:   make(map[string]*QueuedTask),
	}

	// Initialize priority levels
	for i := 0; i <= maxPriority; i++ {
		pq.levels[i] = &PriorityLevel{
			tasks:     make([]*QueuedTask, 0),
			lastEmpty: time.Now(),
		}
	}

	return pq
}

// Push pushes a task to priority queue
func (pq *PriorityQueue) Push(task *QueuedTask) error {
	if task == nil {
		return ErrInvalidTask
	}

	pq.mu.Lock()
	defer pq.mu.Unlock()

	// Check capacity
	if pq.size >= pq.capacity {
		return ErrQueueFull
	}

	// Check for duplicate task ID
	if _, exists := pq.lookup[task.ID]; exists {
		return ErrDuplicateTask
	}

	// Normalize priority
	priority := task.Priority
	if priority < 0 {
		priority = 0
	}
	if priority > pq.maxPrio {
		priority = pq.maxPrio
	}
	task.Priority = priority

	// Add task
	level := pq.levels[priority]
	level.tasks = append(level.tasks, task)
	pq.lookup[task.ID] = task
	pq.size++
	pq.enqueueCount.Add(1)

	return nil
}

// Pop removes and returns the highest priority task
func (pq *PriorityQueue) Pop() *QueuedTask {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if pq.size == 0 {
		return nil
	}

	// Find highest non-empty priority level
	for i := pq.maxPrio; i >= 0; i-- {
		level := pq.levels[i]
		if len(level.tasks) > 0 {
			// Get first task
			task := level.tasks[0]

			// Remove task
			level.tasks = level.tasks[1:]
			delete(pq.lookup, task.ID)
			pq.size--
			pq.dequeueCount.Add(1)

			// Update empty time if level is now empty
			if len(level.tasks) == 0 {
				level.lastEmpty = time.Now()
			}

			return task
		}
	}

	return nil
}

// PopBatch removes and returns up to n highest priority tasks
func (pq *PriorityQueue) PopBatch(n int) []*QueuedTask {
	if n <= 0 {
		return nil
	}

	pq.mu.Lock()
	defer pq.mu.Unlock()

	if pq.size == 0 {
		return nil
	}

	tasks := make([]*QueuedTask, 0, n)
	remaining := n

	// Get tasks from highest to lowest priority
	for i := pq.maxPrio; i >= 0 && remaining > 0; i-- {
		level := pq.levels[i]
		count := len(level.tasks)
		if count == 0 {
			continue
		}

		// Take up to remaining number of tasks
		take := count
		if take > remaining {
			take = remaining
		}

		// Add tasks to result
		tasks = append(tasks, level.tasks[:take]...)

		// Remove taken tasks
		for _, task := range level.tasks[:take] {
			delete(pq.lookup, task.ID)
		}
		level.tasks = level.tasks[take:]

		// Update counters
		pq.size -= take
		pq.dequeueCount.Add(int64(take))
		remaining -= take

		// Update empty time if level is now empty
		if len(level.tasks) == 0 {
			level.lastEmpty = time.Now()
		}
	}

	return tasks
}

// Cancel cancels and removes a task by ID
func (pq *PriorityQueue) Cancel(taskID string) bool {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	task, exists := pq.lookup[taskID]
	if !exists {
		return false
	}

	// Remove from priority level
	level := pq.levels[task.Priority]
	for i, t := range level.tasks {
		if t.ID == taskID {
			// Remove task
			level.tasks = append(level.tasks[:i], level.tasks[i+1:]...)
			delete(pq.lookup, taskID)
			pq.size--

			// Update empty time if level is now empty
			if len(level.tasks) == 0 {
				level.lastEmpty = time.Now()
			}

			return true
		}
	}

	return false
}

// UpdatePriority updates task priority
func (pq *PriorityQueue) UpdatePriority(taskID string, newPriority int) bool {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	task, exists := pq.lookup[taskID]
	if !exists {
		return false
	}

	// Normalize new priority
	if newPriority < 0 {
		newPriority = 0
	}
	if newPriority > pq.maxPrio {
		newPriority = pq.maxPrio
	}

	// No change needed if same priority
	if task.Priority == newPriority {
		return true
	}

	// Remove from old priority level
	oldLevel := pq.levels[task.Priority]
	for i, t := range oldLevel.tasks {
		if t.ID == taskID {
			oldLevel.tasks = append(oldLevel.tasks[:i], oldLevel.tasks[i+1:]...)

			// Update empty time if level is now empty
			if len(oldLevel.tasks) == 0 {
				oldLevel.lastEmpty = time.Now()
			}

			// Add to new priority level
			task.Priority = newPriority
			newLevel := pq.levels[newPriority]
			newLevel.tasks = append(newLevel.tasks, task)

			return true
		}
	}

	return false
}

// GetTask returns a task by ID without removing it
func (pq *PriorityQueue) GetTask(taskID string) *QueuedTask {
	pq.mu.RLock()
	defer pq.mu.RUnlock()

	if task, exists := pq.lookup[taskID]; exists {
		return task
	}
	return nil
}

// Peek returns highest priority task without removing it
func (pq *PriorityQueue) Peek() *QueuedTask {
	pq.mu.RLock()
	defer pq.mu.RUnlock()

	for i := pq.maxPrio; i >= 0; i-- {
		level := pq.levels[i]
		if len(level.tasks) > 0 {
			return level.tasks[0]
		}
	}
	return nil
}

// PeekBatch returns up to n highest priority tasks without removing them
func (pq *PriorityQueue) PeekBatch(n int) []*QueuedTask {
	if n <= 0 {
		return nil
	}

	pq.mu.RLock()
	defer pq.mu.RUnlock()

	tasks := make([]*QueuedTask, 0, n)
	remaining := n

	for i := pq.maxPrio; i >= 0 && remaining > 0; i-- {
		level := pq.levels[i]
		count := len(level.tasks)
		if count == 0 {
			continue
		}

		take := count
		if take > remaining {
			take = remaining
		}

		tasks = append(tasks, level.tasks[:take]...)
		remaining -= take
	}

	return tasks
}

// Len returns the current number of tasks
func (pq *PriorityQueue) Len() int {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return pq.size
}

// LenAtPriority returns number of tasks at given priority level
func (pq *PriorityQueue) LenAtPriority(priority int) int {
	if priority < 0 || priority > pq.maxPrio {
		return 0
	}

	pq.mu.RLock()
	defer pq.mu.RUnlock()

	return len(pq.levels[priority].tasks)
}

// Clear removes all tasks
func (pq *PriorityQueue) Clear() {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	now := time.Now()
	for _, level := range pq.levels {
		level.tasks = level.tasks[:0]
		level.lastEmpty = now
	}

	pq.lookup = make(map[string]*QueuedTask)
	pq.size = 0
}

// GetMetrics returns queue metrics
func (pq *PriorityQueue) GetMetrics() map[string]any {
	pq.mu.RLock()
	defer pq.mu.RUnlock()

	metrics := map[string]any{
		"capacity":      pq.capacity,
		"size":          pq.size,
		"enqueue_count": pq.enqueueCount.Load(),
		"dequeue_count": pq.dequeueCount.Load(),
	}

	// Add per-priority metrics
	levelMetrics := make([]map[string]any, pq.maxPrio+1)
	for i := 0; i <= pq.maxPrio; i++ {
		level := pq.levels[i]
		levelMetrics[i] = map[string]any{
			"priority":   i,
			"size":       len(level.tasks),
			"last_empty": level.lastEmpty,
		}
	}
	metrics["levels"] = levelMetrics

	return metrics
}
