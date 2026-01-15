// Package event provides the domain event bus and stores.
package event

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ncobase/ncore/logging/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EventType defines event types in the application.
type EventType string

const (
	// User events
	EventTypeUserCreated  EventType = "user.created"
	EventTypeUserUpdated  EventType = "user.updated"
	EventTypeUserDeleted  EventType = "user.deleted"
	EventTypeUserLoggedIn EventType = "user.logged_in"

	// Workspace events
	EventTypeWorkspaceCreated EventType = "workspace.created"
	EventTypeWorkspaceUpdated EventType = "workspace.updated"
	EventTypeMemberAdded      EventType = "workspace.member_added"
	EventTypeMemberRemoved    EventType = "workspace.member_removed"

	// Task events
	EventTypeTaskCreated  EventType = "task.created"
	EventTypeTaskUpdated  EventType = "task.updated"
	EventTypeTaskDeleted  EventType = "task.deleted"
	EventTypeTaskAssigned EventType = "task.assigned"

	// Comment events
	EventTypeCommentCreated EventType = "comment.created"
	EventTypeCommentUpdated EventType = "comment.updated"
	EventTypeCommentDeleted EventType = "comment.deleted"

	// Export events
	EventTypeExportRequested EventType = "export.requested"
	EventTypeExportCompleted EventType = "export.completed"
	EventTypeExportFailed    EventType = "export.failed"
)

// Event represents a domain event in the system.
type Event struct {
	ID            string            `json:"id"`
	Type          EventType         `json:"type"`
	AggregateID   string            `json:"aggregate_id,omitempty"`
	AggregateName string            `json:"aggregate_name,omitempty"`
	WorkspaceID   string            `json:"workspace_id,omitempty"`
	UserID        string            `json:"user_id,omitempty"`
	Payload       map[string]any    `json:"payload"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Timestamp     time.Time         `json:"timestamp"`
	Version       int               `json:"version"`
}

// EventHandler defines the event handler function type.
type EventHandler func(ctx context.Context, event *Event) error

// Bus represents the event bus for inter-module communication.
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
	b.logger.Info(context.Background(), "Event handler subscribed",
		"event_type", eventType,
		"total_handlers", len(b.handlers[eventType]))
}

// Publish publishes an event to the bus.
func (b *Bus) Publish(ctx context.Context, event *Event) error {
	// Set event metadata
	event.Timestamp = time.Now()
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Version == 0 {
		event.Version = 1
	}

	// Store event if store is available
	if b.store != nil {
		if err := b.store.Save(ctx, event); err != nil {
			b.logger.Error(ctx, "Failed to store event",
				"error", err,
				"event_id", event.ID,
				"event_type", event.Type)
			// Continue even if storage fails
		}
	}

	// Send to buffer (non-blocking with timeout)
	select {
	case b.buffer <- event:
		b.logger.Debug(ctx, "Event published",
			"type", event.Type,
			"id", event.ID,
			"workspace_id", event.WorkspaceID)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return fmt.Errorf("event buffer full, timeout publishing event")
	}
}

// Start starts the event bus workers.
func (b *Bus) Start(ctx context.Context, numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go b.worker(ctx, i)
	}
	b.logger.Info(ctx, "Event bus started", "workers", numWorkers)
}

// worker processes events from the buffer.
func (b *Bus) worker(ctx context.Context, id int) {
	b.logger.Info(ctx, "Event bus worker started", "worker_id", id)

	for {
		select {
		case <-ctx.Done():
			b.logger.Info(ctx, "Event bus worker stopped", "worker_id", id)
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
		b.logger.Debug(ctx, "No handlers for event", "type", event.Type, "id", event.ID)
		return
	}

	b.logger.Debug(ctx, "Dispatching event",
		"type", event.Type,
		"id", event.ID,
		"handlers", len(handlers))

	// Execute all handlers asynchronously
	var wg sync.WaitGroup
	for i, handler := range handlers {
		wg.Add(1)
		go func(h EventHandler, idx int) {
			defer wg.Done()

			// Create a timeout context for handler execution
			handlerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			startTime := time.Now()
			if err := h(handlerCtx, event); err != nil {
				b.logger.Error(ctx, "Event handler failed",
					"type", event.Type,
					"id", event.ID,
					"handler_index", idx,
					"duration", time.Since(startTime),
					"error", err)
			} else {
				b.logger.Debug(ctx, "Event handler completed",
					"type", event.Type,
					"id", event.ID,
					"handler_index", idx,
					"duration", time.Since(startTime))
			}
		}(handler, i)
	}

	// Wait for all handlers to complete (with timeout)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All handlers completed
	case <-time.After(1 * time.Minute):
		b.logger.Warn(ctx, "Event dispatch timeout", "type", event.Type, "id", event.ID)
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

// Shutdown gracefully shuts down the event bus.
func (b *Bus) Shutdown(ctx context.Context) error {
	b.logger.Info(ctx, "Shutting down event bus", "pending_events", len(b.buffer))

	// Drain remaining events with timeout
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-timeout:
			b.logger.Warn(ctx, "Event bus shutdown timeout", "remaining_events", len(b.buffer))
			return fmt.Errorf("shutdown timeout with %d events remaining", len(b.buffer))
		case <-ctx.Done():
			return ctx.Err()
		default:
			if len(b.buffer) == 0 {
				b.logger.Info(ctx, "Event bus shutdown complete")
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// EventStore defines the interface for event persistence.
type EventStore interface {
	Save(ctx context.Context, event *Event) error
	Load(ctx context.Context, eventID string) (*Event, error)
	LoadByAggregate(ctx context.Context, aggregateID string) ([]*Event, error)
	LoadByType(ctx context.Context, eventType EventType) ([]*Event, error)
	LoadByWorkspace(ctx context.Context, workspaceID string) ([]*Event, error)
	LoadSince(ctx context.Context, since time.Time) ([]*Event, error)
}

// MemoryStore is an in-memory implementation of EventStore.
type MemoryStore struct {
	events map[string]*Event
	mu     sync.RWMutex
	logger *logger.Logger
}

// NewMemoryStore creates a new memory-based event store.
func NewMemoryStore(logger *logger.Logger) *MemoryStore {
	return &MemoryStore{
		events: make(map[string]*Event),
		logger: logger,
	}
}

// Save saves an event to memory.
func (s *MemoryStore) Save(ctx context.Context, event *Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a copy to avoid mutations
	eventCopy := *event
	s.events[event.ID] = &eventCopy
	s.logger.Debug(ctx, "Event stored", "id", event.ID, "type", event.Type)
	return nil
}

// Load loads an event by ID.
func (s *MemoryStore) Load(ctx context.Context, eventID string) (*Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, ok := s.events[eventID]
	if !ok {
		return nil, fmt.Errorf("event not found: %s", eventID)
	}
	return event, nil
}

// LoadByAggregate loads events for a specific aggregate.
func (s *MemoryStore) LoadByAggregate(ctx context.Context, aggregateID string) ([]*Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var events []*Event
	for _, event := range s.events {
		if event.AggregateID == aggregateID {
			events = append(events, event)
		}
	}
	return events, nil
}

// LoadByType loads events of a specific type.
func (s *MemoryStore) LoadByType(ctx context.Context, eventType EventType) ([]*Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var events []*Event
	for _, event := range s.events {
		if event.Type == eventType {
			events = append(events, event)
		}
	}
	return events, nil
}

// LoadByWorkspace loads events for a specific workspace.
func (s *MemoryStore) LoadByWorkspace(ctx context.Context, workspaceID string) ([]*Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var events []*Event
	for _, event := range s.events {
		if event.WorkspaceID == workspaceID {
			events = append(events, event)
		}
	}
	return events, nil
}

// LoadSince loads events since a specific time.
func (s *MemoryStore) LoadSince(ctx context.Context, since time.Time) ([]*Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var events []*Event
	for _, event := range s.events {
		if event.Timestamp.After(since) {
			events = append(events, event)
		}
	}
	return events, nil
}

type MongoStore struct {
	collection *mongo.Collection
	logger     *logger.Logger
}

func NewMongoStore(collection *mongo.Collection, logger *logger.Logger) (*MongoStore, error) {
	if collection == nil {
		return nil, fmt.Errorf("mongo collection is nil")
	}

	store := &MongoStore{collection: collection, logger: logger}
	if err := store.ensureIndexes(context.Background()); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *MongoStore) ensureIndexes(ctx context.Context) error {
	_, err := s.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "aggregate_id", Value: 1}, {Key: "timestamp", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "type", Value: 1}, {Key: "timestamp", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "timestamp", Value: -1}},
		},
	})
	return err
}

func (s *MongoStore) Save(ctx context.Context, event *Event) error {
	filter := bson.M{"id": event.ID}
	update := bson.M{"$set": event}
	_, err := s.collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil && s.logger != nil {
		s.logger.Error(ctx, "Failed to save event in Mongo", "error", err, "event_id", event.ID)
	}
	return err
}

func (s *MongoStore) Load(ctx context.Context, eventID string) (*Event, error) {
	result := &Event{}
	if err := s.collection.FindOne(ctx, bson.M{"id": eventID}).Decode(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *MongoStore) LoadByAggregate(ctx context.Context, aggregateID string) ([]*Event, error) {
	return s.loadMany(ctx, bson.M{"aggregate_id": aggregateID})
}

func (s *MongoStore) LoadByType(ctx context.Context, eventType EventType) ([]*Event, error) {
	return s.loadMany(ctx, bson.M{"type": eventType})
}

func (s *MongoStore) LoadByWorkspace(ctx context.Context, workspaceID string) ([]*Event, error) {
	return s.loadMany(ctx, bson.M{"workspace_id": workspaceID})
}

func (s *MongoStore) LoadSince(ctx context.Context, since time.Time) ([]*Event, error) {
	return s.loadMany(ctx, bson.M{"timestamp": bson.M{"$gte": since}})
}

func (s *MongoStore) loadMany(ctx context.Context, filter bson.M) ([]*Event, error) {
	cursor, err := s.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []*Event
	for cursor.Next(ctx) {
		evt := &Event{}
		if err := cursor.Decode(evt); err != nil {
			return nil, err
		}
		events = append(events, evt)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return events, nil
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
