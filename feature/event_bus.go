package feature

import (
	"sync"
)

// EventBus represents a simple event bus for inter-feature communication
type EventBus struct {
	subscribers map[string][]func(any)
	mu          sync.RWMutex
}

// NewEventBus creates a new EventBus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]func(any)),
	}
}

// Subscribe adds a subscriber for a specific event
func (eb *EventBus) Subscribe(eventName string, handler func(any)) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.subscribers[eventName] = append(eb.subscribers[eventName], handler)
}

// Publish sends an event to all subscribers
func (eb *EventBus) Publish(eventName string, data any) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if handlers, ok := eb.subscribers[eventName]; ok {
		for _, handler := range handlers {
			go handler(data)
		}
	}
}
