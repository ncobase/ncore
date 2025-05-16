package event

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// EventBus represents a simple event bus for inter-extension communication
type EventBus struct {
	subscribers map[string][]func(any)
	mu          sync.RWMutex
	metrics     struct {
		processed     atomic.Int64
		failed        atomic.Int64
		lastEventTime atomic.Value
		queueSize     atomic.Int32
	}
}

// NewEventBus creates a new EventBus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]func(any)),
	}
}

// GetMetrics returns metrics
func (eb *EventBus) GetMetrics() map[string]any {
	return map[string]any{
		"processed_events": eb.metrics.processed.Load(),
		"failed_events":    eb.metrics.failed.Load(),
		"last_event_time":  eb.metrics.lastEventTime.Load(),
		"queue_size":       eb.metrics.queueSize.Load(),
	}
}

// Subscribe adds a subscriber for a specific event
func (eb *EventBus) Subscribe(eventName string, handler func(any)) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if handler == nil {
		return
	}

	wrappedHandler := func(data any) {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf(nil, "panic in event handler: %v", r)
			}
		}()
		handler(data)
	}

	eb.subscribers[eventName] = append(eb.subscribers[eventName], wrappedHandler)
}

// Publish sends an event to all subscribers
func (eb *EventBus) Publish(eventName string, data any) {
	eb.mu.RLock()
	handlers, exists := eb.subscribers[eventName]
	eb.mu.RUnlock()

	if !exists {
		return
	}

	eventData := types.EventData{
		Time:      time.Now(),
		Source:    "extension",
		EventType: eventName,
		Data:      data,
	}

	for _, handler := range handlers {
		go func(h func(any)) {
			defer func() {
				if r := recover(); r != nil {
					eb.metrics.failed.Add(1)
					logger.Errorf(nil, "event handler panic: %v", r)
				}
			}()

			h(eventData)
			eb.metrics.processed.Add(1)
		}(handler)
	}
}

// PublishWithRetry retry publish event
func (eb *EventBus) PublishWithRetry(eventName string, data any, maxRetries int) {
	var attempts int
	for attempts < maxRetries {
		if err := eb.publishWithError(eventName, data); err != nil {
			attempts++
			time.Sleep(time.Duration(attempts) * time.Second)
			continue
		}
		return
	}
	eb.metrics.failed.Add(1)
}

// publishWithError publish with error
func (eb *EventBus) publishWithError(eventName string, data any) error {
	eb.mu.RLock()
	handlers, exists := eb.subscribers[eventName]
	eb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no handlers for event: %s", eventName)
	}

	eventData := types.EventData{
		Time:      time.Now(),
		Source:    "extension",
		EventType: eventName,
		Data:      data,
	}

	eb.metrics.lastEventTime.Store(eventData.Time)
	eb.metrics.queueSize.Add(int32(len(handlers)))
	defer eb.metrics.queueSize.Add(int32(-len(handlers)))

	for _, handler := range handlers {
		go func(h func(any)) {
			defer func() {
				if r := recover(); r != nil {
					eb.metrics.failed.Add(1)
					logger.Errorf(nil, "event handler panic: %v", r)
				}
			}()

			h(eventData)
			eb.metrics.processed.Add(1)
		}(handler)
	}

	return nil
}
