# Example 06: Event-Driven Architecture

Demonstrates event-driven patterns with NCore's event bus, pub/sub system, and async event handling from the extension
system.

## Features

- **Event Bus**: Central event distribution
- **Pub/Sub Pattern**: Publish events, subscribe to topics
- **Async Processing**: Non-blocking event handling
- **Event Types**: System events, domain events, integration events
- **Event Persistence**: Optional event sourcing
- **Event Replay**: Replay events for debugging

## Architecture

```text
┌─────────────┐                    ┌──────────────┐
│  Publisher  │───── Event ───────►│  Event Bus   │
└─────────────┘                    └──────┬───────┘
                                          │
                            ┌─────────────┼─────────────┐
                            │             │             │
                            ▼             ▼             ▼
                      ┌──────────┐  ┌──────────┐  ┌──────────┐
                      │ Handler  │  │ Handler  │  │ Handler  │
                      │    A     │  │    B     │  │    C     │
                      └──────────┘  └──────────┘  └──────────┘
```

## Event Types

### 1. Domain Events

```go
// User registered event
type UserRegisteredEvent struct {
    UserID    string
    Email     string
    Timestamp time.Time
}
```

### 2. System Events

```go
// Extension loaded event
type ExtensionLoadedEvent struct {
    Name      string
    Version   string
    Timestamp time.Time
}
```

### 3. Integration Events

```go
// External service notification
type PaymentReceivedEvent struct {
    OrderID   string
    Amount    float64
    Currency  string
    Timestamp time.Time
}
```

## Project Structure

```text
06-event-driven/
├── data/
│   ├── data.go          # SQLite connection
│   └── repository/
│       └── user.go      # User repository
├── event/
│   ├── bus.go           # Event bus implementation
│   └── store_sqlite.go  # SQLite event store
├── service/
│   ├── handlers.go      # Notification/analytics/audit services
│   └── user.go          # Publishes events
└── handler/
│   └── handler.go       # HTTP handlers
├── main.go
├── config.yaml
└── README.md
```

## Publishing Events

```go
// In user service
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserReq) (*User, error) {
    user, err := s.repo.Create(ctx, req)
    if err != nil {
        return nil, err
    }

    // Publish event
    s.eventBus.Publish("user.registered", &UserRegisteredEvent{
        UserID:    user.ID,
        Email:     user.Email,
        Timestamp: time.Now(),
    })

    return user, nil
}
```

## Subscribing to Events

```go
// In notification service
func (s *NotificationService) Init() {
    s.eventBus.Subscribe("user.registered", s.handleUserRegistered)
    s.eventBus.Subscribe("order.placed", s.handleOrderPlaced)
}

func (s *NotificationService) handleUserRegistered(event *UserRegisteredEvent) error {
    return s.sendWelcomeEmail(event.Email)
}
```

## Event Bus Implementation

```go
type EventBus struct {
    handlers map[string][]EventHandler
    mu       sync.RWMutex
}

func (b *EventBus) Publish(topic string, event interface{}) {
    b.mu.RLock()
    handlers := b.handlers[topic]
    b.mu.RUnlock()

    for _, handler := range handlers {
        go handler(event) // Async execution
    }
}

func (b *EventBus) Subscribe(topic string, handler EventHandler) {
    b.mu.Lock()
    defer b.mu.Unlock()

    b.handlers[topic] = append(b.handlers[topic], handler)
}
```

## Event Persistence

```go
type EventStore interface {
    Save(event *Event) error
    Load(eventID string) (*Event, error)
    LoadByAggregate(aggregateID string) ([]*Event, error)
    Replay(from time.Time) error
}

// Example: Save to database
func (s *PostgresEventStore) Save(event *Event) error {
    _, err := s.db.Exec(`
        INSERT INTO events (id, type, aggregate_id, payload, timestamp)
        VALUES ($1, $2, $3, $4, $5)
    `, event.ID, event.Type, event.AggregateID, event.Payload, event.Timestamp)
    return err
}
```

## Event Patterns

### 1. Fan-Out Pattern

```
User Created Event
    ├── Send welcome email
    ├── Create user profile
    ├── Log analytics
    └── Update cache
```

### 2. Saga Pattern

```
Order Placed
    ├── Reserve inventory → Success/Failure
    ├── Process payment → Success/Failure
    └── Send confirmation → Success/Failure
```

### 3. CQRS Pattern

```
Command: CreateUser
    └── Event: UserCreated
        ├── Write Model (Update database)
        └── Read Model (Update search index)
```

## API Examples

### Publish Event

```bash
curl -X POST http://localhost:8080/events \
  -d '{
    "type": "user.registered",
    "data": {
      "user_id": "123",
      "email": "user@example.com"
    }
  }'
```

### Query Events

```bash
# Get all events for aggregate
curl http://localhost:8080/events?aggregate_id=user-123

# Get events by type
curl http://localhost:8080/events?type=user.registered

# Get events since timestamp
curl http://localhost:8080/events?since=2024-01-01T00:00:00Z
```

## Event Handlers

### Email Handler

```go
func (h *EmailHandler) Handle(event Event) error {
    switch e := event.(type) {
    case *UserRegisteredEvent:
        return h.sendWelcomeEmail(e.Email)
    case *PasswordResetEvent:
        return h.sendPasswordResetEmail(e.Email, e.Token)
    }
    return nil
}
```

### Analytics Handler

```go
func (h *AnalyticsHandler) Handle(event Event) error {
    return h.tracker.Track(event.Type, map[string]interface{}{
        "timestamp": event.Timestamp,
        "data":      event.Data,
    })
}
```

```yaml
event:
  bus:
    buffer_size: 1000
    workers: 10
  store:
    enabled: true
    driver: postgres
    retention_days: 90
  handlers:
    async: true
    timeout: 30s
    retry_count: 3
```

## Monitoring

```bash
# Event bus stats
curl http://localhost:8080/events/stats

# Response:
{
  "published": 15234,
  "processed": 15180,
  "failed": 54,
  "subscribers": {
    "user.registered": 3,
    "order.placed": 5
  }
}
```

## Testing

```go
func TestUserService_CreateUser(t *testing.T) {
    bus := event.NewMockBus()
    svc := NewUserService(repo, bus)

    user, err := svc.CreateUser(ctx, req)
    assert.NoError(t, err)

    // Verify event published
    assert.Equal(t, 1, bus.PublishedCount("user.registered"))
}
```

## Use Cases

- User registration workflows
- Order processing pipelines
- Notification systems
- Audit logging
- Analytics tracking
- Cache invalidation
- Microservice communication

## Next Steps

- Integrate with [message queue](../08-full-application) for reliability
- Add [authentication](../07-authentication) for event publishing
- Combine with [background jobs](../05-background-jobs) for heavy processing

## License

This example is part of the NCore project.
