package proxy

import (
	"context"
	"sync"
)

// Event represents a proxy event
type Event struct {
	Type    string
	Payload any
}

// EventHandler is a function that handles proxy events
type EventHandler func(Event)

// EventBusInterface interface for handling events
type EventBusInterface interface {
	Publish(context context.Context, event Event)
	Subscribe(eventType string, handler EventHandler)
	Unsubscribe(eventType string, handler EventHandler)
}

// EventBus is a basic implementation of the EventBus interface
type EventBus struct {
	subscribers map[string][]EventHandler
	mu          sync.RWMutex
}

// NewEventBus creates a new eventBus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]EventHandler),
	}
}

// Publish publishes an event to all subscribers
func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if handlers, ok := eb.subscribers[event.Type]; ok {
		for _, handler := range handlers {
			go handler(event)
		}
	}
}

// Subscribe adds a new subscriber for a specific event type
func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.subscribers[eventType] = append(eb.subscribers[eventType], handler)
}

// Unsubscribe removes a subscriber for a specific event type
func (eb *EventBus) Unsubscribe(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if handlers, ok := eb.subscribers[eventType]; ok {
		for i, h := range handlers {
			if &h == &handler {
				eb.subscribers[eventType] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
	}
}
