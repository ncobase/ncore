// Package event implements the in-process event bus.
package event

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ncobase/ncore/logging/logger"
)

// EventType defines event types.
type EventType string

const (
	EventTypeUserRegistered EventType = "user.registered"
	EventTypeUserUpdated    EventType = "user.updated"
	EventTypeOrderPlaced    EventType = "order.placed"
	EventTypeOrderPaid      EventType = "order.paid"
	EventTypeEmailSent      EventType = "email.sent"
	EventTypeCustom         EventType = "custom"
)

// Event represents a domain event.
type Event struct {
	ID            string            `json:"id"`
	Type          EventType         `json:"type"`
	AggregateID   string            `json:"aggregate_id,omitempty"`
	AggregateName string            `json:"aggregate_name,omitempty"`
	Payload       map[string]any    `json:"payload"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Timestamp     time.Time         `json:"timestamp"`
	Version       int               `json:"version"`
}

// EventHandler defines the event handler function type.
type EventHandler func(ctx context.Context, event *Event) error

// Bus represents the event bus.
type Bus struct {
	handlers map[EventType][]EventHandler
	buffer   chan *Event
	mu       sync.RWMutex
	logger   *logger.Logger
	store    EventStore
}

// NewBus creates a new event bus.
func NewBus(bufferSize int, logger *logger.Logger, store EventStore) *Bus {
	return &Bus{
		handlers: make(map[EventType][]EventHandler),
		buffer:   make(chan *Event, bufferSize),
		logger:   logger,
		store:    store,
	}
}

// Subscribe subscribes a handler to an event type.
func (b *Bus) Subscribe(eventType EventType, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
	b.logger.Info(context.Background(), "Event handler subscribed", "event_type", eventType)
}

// Publish publishes an event to the bus.
func (b *Bus) Publish(ctx context.Context, event *Event) error {
	event.Timestamp = time.Now()
	if event.ID == "" {
		event.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	// Store event if store is available
	if b.store != nil {
		if err := b.store.Save(ctx, event); err != nil {
			b.logger.Error(ctx, "Failed to store event", "error", err)
			// Continue even if storage fails
		}
	}

	// Send to buffer
	select {
	case b.buffer <- event:
		b.logger.Debug(ctx, "Event published", "type", event.Type, "id", event.ID)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("event buffer full")
	}
}

// Start starts the event bus workers.
func (b *Bus) Start(ctx context.Context, numWorkers int) {
	for i := range numWorkers {
		go b.worker(ctx, i)
	}
	b.logger.Info(ctx, "Event bus started", "workers", numWorkers)
}

// worker processes events from the buffer.
func (b *Bus) worker(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done():
			b.logger.Info(ctx, "Event worker stopped", "worker_id", id)
			return
		case event := <-b.buffer:
			b.dispatch(ctx, event)
		}
	}
}

// dispatch dispatches an event to all subscribed handlers.
func (b *Bus) dispatch(ctx context.Context, event *Event) {
	b.mu.RLock()
	handlers := b.handlers[event.Type]
	b.mu.RUnlock()

	if len(handlers) == 0 {
		b.logger.Debug(ctx, "No handlers for event", "type", event.Type)
		return
	}

	for _, handler := range handlers {
		go func(h EventHandler) {
			if err := h(ctx, event); err != nil {
				b.logger.Error(ctx, "Event handler failed",
					"type", event.Type,
					"id", event.ID,
					"error", err)
			}
		}(handler)
	}
}

// GetStats returns event bus statistics.
func (b *Bus) GetStats() map[string]any {
	b.mu.RLock()
	defer b.mu.RUnlock()

	subscribers := make(map[string]int)
	for eventType, handlers := range b.handlers {
		subscribers[string(eventType)] = len(handlers)
	}

	return map[string]any{
		"buffer_size":    cap(b.buffer),
		"buffer_used":    len(b.buffer),
		"event_types":    len(b.handlers),
		"total_handlers": b.countHandlers(),
		"subscribers":    subscribers,
	}
}

func (b *Bus) countHandlers() int {
	count := 0
	for _, handlers := range b.handlers {
		count += len(handlers)
	}
	return count
}

// EventStore defines the interface for event persistence.
type EventStore interface {
	Save(ctx context.Context, event *Event) error
	Load(ctx context.Context, eventID string) (*Event, error)
	LoadByAggregate(ctx context.Context, aggregateID string) ([]*Event, error)
	LoadByType(ctx context.Context, eventType EventType) ([]*Event, error)
	LoadSince(ctx context.Context, since time.Time) ([]*Event, error)
}

// MarshalEvent marshals an event to JSON.
func MarshalEvent(event *Event) ([]byte, error) {
	return json.Marshal(event)
}

// UnmarshalEvent unmarshals an event from JSON.
func UnmarshalEvent(data []byte) (*Event, error) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}
	return &event, nil
}
