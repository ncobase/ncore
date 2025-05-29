package event

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// Bus represents a simple event bus for inter-extension communication
type Bus struct {
	subscribers map[string][]func(any)
	mu          sync.RWMutex
	metrics     struct {
		published        atomic.Int64
		delivered        atomic.Int64
		failed           atomic.Int64
		lastEventTime    atomic.Value
		activeHandlers   atomic.Int32
		totalSubscribers atomic.Int32
	}
}

// NewEventBus creates a new EventBus
func NewEventBus() *Bus {
	eb := &Bus{
		subscribers: make(map[string][]func(any)),
	}
	eb.metrics.lastEventTime.Store(time.Time{})
	return eb
}

// GetMetrics returns event bus metrics
func (eb *Bus) GetMetrics() map[string]any {
	lastEventTime := eb.metrics.lastEventTime.Load().(time.Time)
	return map[string]any{
		"published_events":  eb.metrics.published.Load(),
		"delivered_events":  eb.metrics.delivered.Load(),
		"failed_events":     eb.metrics.failed.Load(),
		"last_event_time":   lastEventTime,
		"active_handlers":   eb.metrics.activeHandlers.Load(),
		"total":             eb.metrics.totalSubscribers.Load(),
		"events_per_second": eb.calculateEventsPerSecond(lastEventTime),
		"failure_rate":      eb.calculateFailureRate(),
	}
}

// calculateEventsPerSecond calculates events per second based on recent activity
func (eb *Bus) calculateEventsPerSecond(lastEventTime time.Time) float64 {
	if lastEventTime.IsZero() {
		return 0.0
	}

	duration := time.Since(lastEventTime)
	if duration > time.Minute {
		return 0.0 // No recent activity
	}

	published := eb.metrics.published.Load()
	if published == 0 {
		return 0.0
	}

	// Simple approximation - actual implementation would need time window tracking
	return float64(published) / duration.Seconds()
}

// calculateFailureRate calculates the failure rate percentage
func (eb *Bus) calculateFailureRate() float64 {
	total := eb.metrics.published.Load()
	if total == 0 {
		return 0.0
	}

	failed := eb.metrics.failed.Load()
	return (float64(failed) / float64(total)) * 100.0
}

// Subscribe adds a subscriber for a specific event
func (eb *Bus) Subscribe(eventName string, handler func(any)) {
	if handler == nil {
		return
	}

	eb.mu.Lock()
	defer eb.mu.Unlock()

	wrappedHandler := func(data any) {
		eb.metrics.activeHandlers.Add(1)
		defer eb.metrics.activeHandlers.Add(-1)

		defer func() {
			if r := recover(); r != nil {
				eb.metrics.failed.Add(1)
				logger.Errorf(nil, "panic in event handler: %v", r)
			}
		}()

		handler(data)
		eb.metrics.delivered.Add(1)
	}

	eb.subscribers[eventName] = append(eb.subscribers[eventName], wrappedHandler)
	eb.metrics.totalSubscribers.Add(1)
}

// Publish sends an event to all subscribers
func (eb *Bus) Publish(eventName string, data any) {
	eb.mu.RLock()
	handlers, exists := eb.subscribers[eventName]
	eb.mu.RUnlock()

	if !exists || len(handlers) == 0 {
		return
	}

	eb.metrics.published.Add(1)
	eb.metrics.lastEventTime.Store(time.Now())

	eventData := types.EventData{
		Time:      time.Now(),
		Source:    "extension",
		EventType: eventName,
		Data:      data,
	}

	for _, handler := range handlers {
		go handler(eventData)
	}
}

// PublishWithRetry publishes event with retry logic
func (eb *Bus) PublishWithRetry(eventName string, data any, maxRetries int) {
	var attempts int
	for attempts <= maxRetries {
		if err := eb.publishWithError(eventName, data); err == nil {
			return
		}

		attempts++
		if attempts <= maxRetries {
			backoff := time.Duration(attempts) * time.Second
			time.Sleep(backoff)
		}
	}

	eb.metrics.failed.Add(1)
	logger.Errorf(nil, "failed to publish event %s after %d retries", eventName, maxRetries)
}

// publishWithError publishes event and returns error if no handlers
func (eb *Bus) publishWithError(eventName string, data any) error {
	eb.mu.RLock()
	handlers, exists := eb.subscribers[eventName]
	eb.mu.RUnlock()

	if !exists || len(handlers) == 0 {
		return fmt.Errorf("no handlers for event: %s", eventName)
	}

	eb.metrics.published.Add(1)
	eb.metrics.lastEventTime.Store(time.Now())

	eventData := types.EventData{
		Time:      time.Now(),
		Source:    "extension",
		EventType: eventName,
		Data:      data,
	}

	for _, handler := range handlers {
		go handler(eventData)
	}

	return nil
}
