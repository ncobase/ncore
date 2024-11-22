package concurrency

import (
	"context"
	"fmt"
	"sync/atomic"
)

// Manager manages concurrent executions with enhanced features
type Manager struct {
	maxConcurrent int32
	current       atomic.Int32
	semaphore     chan struct{}
	// Add metrics tracking
	totalExecutions atomic.Int64
	rejectedCount   atomic.Int64
}

// NewManager creates a new concurrency manager with validation
//
// Usage:
//
//	// Create a manager that allows up to 10 concurrent operations
//	cm, err := NewManager(10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Example 1: Basic usage with context timeout
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	if err := cm.Acquire(ctx); err != nil {
//	    // Handle acquisition failure (timeout/cancellation)
//	    return err
//	}
//	defer cm.Release()  // Always release when done
//
//	// Example 2: Non-blocking attempt
//	if acquired := cm.TryAcquire(); acquired {
//	    defer cm.Release()
//	    // Do work...
//	} else {
//	    // Handle busy case
//	}
//
//	// Example 3: Check metrics
//	metrics := cm.GetMetrics()
//	log.Printf("Current usage: %d/%d", metrics["current"], 10)
func NewManager(max int32) (*Manager, error) {
	if max <= 0 {
		return nil, fmt.Errorf("max concurrent must be positive, got: %d", max)
	}

	return &Manager{
		maxConcurrent: max,
		semaphore:     make(chan struct{}, max),
	}, nil
}

// Acquire attempts to acquire a concurrency slot with timeout
func (m *Manager) Acquire(ctx context.Context) error {
	select {
	case m.semaphore <- struct{}{}:
		m.current.Add(1)
		m.totalExecutions.Add(1)
		return nil
	case <-ctx.Done():
		m.rejectedCount.Add(1)
		return fmt.Errorf("failed to acquire concurrency slot: %w", ctx.Err())
	}
}

// Release releases a concurrency slot
func (m *Manager) Release() {
	select {
	case <-m.semaphore:
		m.current.Add(-1)
	default:
		// Add error logging here
		panic("attempting to release more slots than acquired")
	}
}

// GetMetrics returns current metrics
func (m *Manager) GetMetrics() map[string]int64 {
	return map[string]int64{
		"current":          int64(m.current.Load()),
		"total_executions": m.totalExecutions.Load(),
		"rejected_count":   m.rejectedCount.Load(),
	}
}

// TryAcquire attempts to acquire without blocking
func (m *Manager) TryAcquire() bool {
	select {
	case m.semaphore <- struct{}{}:
		m.current.Add(1)
		m.totalExecutions.Add(1)
		return true
	default:
		return false
	}
}

// Available returns the number of available slots
func (m *Manager) Available() int32 {
	return m.maxConcurrent - m.current.Load()
}
