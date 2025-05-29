package event

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// Dispatcher handles event publishing and subscription
type Dispatcher struct {
	subscribers map[string][]func(any)
	mu          sync.RWMutex
	metrics     struct {
		published        atomic.Int64
		delivered        atomic.Int64
		failed           atomic.Int64
		retries          atomic.Int64
		lastEventTime    atomic.Value // time.Time
		activeHandlers   atomic.Int32
		totalSubscribers atomic.Int32
	}
}

// NewEventDispatcher creates a new event dispatcher
func NewEventDispatcher() *Dispatcher {
	d := &Dispatcher{
		subscribers: make(map[string][]func(any)),
	}
	d.metrics.lastEventTime.Store(time.Time{})
	return d
}

// Subscribe adds a handler for specific event
func (d *Dispatcher) Subscribe(eventName string, handler func(any)) {
	if handler == nil {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Wrap handler with metrics
	wrappedHandler := d.wrapHandler(handler)
	d.subscribers[eventName] = append(d.subscribers[eventName], wrappedHandler)
	d.metrics.totalSubscribers.Add(1)
}

// Publish sends event to all subscribers
func (d *Dispatcher) Publish(eventName string, data any) {
	d.mu.RLock()
	handlers, exists := d.subscribers[eventName]
	handlerCount := len(handlers)
	d.mu.RUnlock()

	if !exists || handlerCount == 0 {
		return
	}

	d.metrics.published.Add(1)
	d.metrics.lastEventTime.Store(time.Now())

	eventData := types.EventData{
		Time:      time.Now(),
		Source:    "extension",
		EventType: eventName,
		Data:      data,
	}

	// Execute handlers concurrently
	for _, handler := range handlers {
		go handler(eventData)
	}
}

// PublishWithRetry publishes event with retry mechanism
func (d *Dispatcher) PublishWithRetry(eventName string, data any, maxRetries int) {
	d.mu.RLock()
	handlers, exists := d.subscribers[eventName]
	d.mu.RUnlock()

	if !exists || len(handlers) == 0 {
		d.metrics.failed.Add(1)
		return
	}

	attempt := 0
	for attempt <= maxRetries {
		d.metrics.retries.Add(1)

		if attempt > 0 {
			backoff := time.Duration(attempt) * time.Second
			time.Sleep(backoff)
		}

		d.Publish(eventName, data)

		// Simple success assumption - in real implementation,
		// would need confirmation mechanism
		return
	}

	d.metrics.failed.Add(1)
}

// wrapHandler wraps user handler with metrics and error handling
func (d *Dispatcher) wrapHandler(handler func(any)) func(any) {
	return func(data any) {
		d.metrics.activeHandlers.Add(1)
		defer d.metrics.activeHandlers.Add(-1)

		defer func() {
			if r := recover(); r != nil {
				d.metrics.failed.Add(1)
				logger.Errorf(nil, "event handler panic: %v", r)
				return
			}
			d.metrics.delivered.Add(1)
		}()

		handler(data)
	}
}

// GetMetrics returns comprehensive metrics
func (d *Dispatcher) GetMetrics() map[string]any {
	lastEventTime := d.metrics.lastEventTime.Load().(time.Time)

	published := d.metrics.published.Load()
	delivered := d.metrics.delivered.Load()
	failed := d.metrics.failed.Load()

	var successRate float64
	if published > 0 {
		successRate = (float64(delivered) / float64(published)) * 100.0
	}

	return map[string]any{
		"published":         published,
		"delivered":         delivered,
		"failed":            failed,
		"retries":           d.metrics.retries.Load(),
		"success_rate":      successRate,
		"last_event_time":   lastEventTime,
		"active_handlers":   d.metrics.activeHandlers.Load(),
		"total_subscribers": d.metrics.totalSubscribers.Load(),
		"events_per_second": d.calculateEventsPerSecond(published, lastEventTime),
	}
}

// calculateEventsPerSecond calculates recent event rate
func (d *Dispatcher) calculateEventsPerSecond(published int64, lastEventTime time.Time) float64 {
	if lastEventTime.IsZero() || published == 0 {
		return 0.0
	}

	duration := time.Since(lastEventTime)
	if duration > time.Minute {
		return 0.0
	}

	return float64(published) / duration.Seconds()
}
