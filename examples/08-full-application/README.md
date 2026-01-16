# Example 08: Full Application

A production-ready collaborative task management platform demonstrating comprehensive NCore features including
multi-tenancy, real-time communication, event-driven architecture, authentication, background jobs, and WebSocket
integration.

## Overview

This example showcases a complete, production-ready application combining all NCore patterns from previous examples into
a cohesive system. It implements a collaborative task management platform with:

- **Multi-tenant Architecture**: Workspaces for isolation and team management
- **Real-time Updates**: WebSocket-based live collaboration
- **Event-Driven Communication**: Modules communicate via event bus
- **JWT Authentication**: Secure token-based auth with refresh tokens
- **Background Processing**: Async job execution for exports
- **Email Notifications**: Event-driven email alerts

## Architecture

```text
┌─────────────────────────────────────────────────────────────────────────┐
│                         Application Server                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────┐     ┌──────────┐     ┌──────────┐    │
│  │   Core    │     │   Biz    │     │  Plugin   │    │
│  │  Modules  │     │  Modules  │     │  Modules  │    │
│  │           │     │           │     │           │    │
│  │ • auth    │     │ • task    │     │ • notify  │    │
│  │ • user    │◄───►│ • comment │◄───►│ • export  │    │
│  │ • workspace│     │ • realtime│     │           │    │
│  └──────────┘     └──────────┘     └──────────┘    │
│         │                │                   │              │
│         │                │                   │              │
│         ▼                ▼                   ▼              │
│  ┌───────────────────────────────────────────────────────┐        │
│  │              Event Bus (Pub/Sub)               │        │
│  └───────────────────────────────────────────────────────┘        │
│         │                │                   │                       │
│         ▼                ▼                   ▼                       │
│  ┌──────────┐     ┌──────────┐     ┌──────────┐        │
│  │WebSocket │     │  Worker   │     │   Email   │        │
│  │   Hub    │     │   Pool    │     │  Sender   │        │
│  └──────────┘     └──────────┘     └──────────┘        │
└─────────────────────────────────────────────────────────────────────────┘
```

## Features Demonstrated

### 1. Multi-Tenancy with Workspaces

- Workspace isolation for different teams/organizations
- Member management with role-based permissions (owner, admin, member)
- Workspace-scoped resources (tasks, comments)

### 2. Real-Time WebSocket Communication

- Room-based broadcasting by workspace
- Live updates for tasks and comments
- Connection management and graceful disconnects
- Ping/pong for connection health

### 3. Event-Driven Architecture

- Inter-module communication via event bus
- Async event processing with worker pool
- Event sourcing capabilities
- Workspace-scoped events

### 4. JWT Authentication & RBAC

- Token-based authentication with JWT
- Access and refresh token pattern
- Role-based authorization (user, admin)
- Middleware for protected routes

### 5. Background Job Processing

- Export jobs (CSV, JSON)
- Job status tracking (pending, running, completed, failed)
- Worker pool for concurrent execution
- Progress updates via events

### 6. Email Notifications

- Event-driven email notifications
- Task creation alerts
- Task assignment notifications
- Comment notifications

## Project Structure

```text
08-full-application/
├── main.go                         # Application entry point
├── core/                           # Core domain modules
│   ├── auth/
│   │   ├── auth.go                 # Auth module
│   │   ├── service/
│   │   │   └── service.go          # Auth business logic
│   │   └── middleware/
│   │       └── middleware.go       # JWT middleware
│   ├── user/
│   │   ├── user.go                 # User module
│   │   ├── handler/
│   │   │   └── handler.go          # User HTTP handlers
│   │   ├── service/
│   │   │   └── service.go          # User business logic
│   │   ├── structs/
│   │   │   └── structs.go          # User structs
│   │   └── data/
│   │       ├── generate.go         # Ent generation entrypoint
│   │       ├── schema/
│   │       │   └── user.go         # Ent schema
│   │       ├── ent/                # Generated Ent client
│   │       └── repository/
│   │           └── repository.go   # Postgres user repo (Ent)
│   └── workspace/
│       ├── workspace.go            # Workspace module
│       ├── structs/
│       │   └── structs.go          # Workspace structs
│       └── data/repository/
│           └── repository.go       # Postgres workspace repo
├── biz/                            # Business domain modules
│   ├── task/
│   │   ├── task.go                 # Task module
│   │   ├── structs/
│   │   │   └── structs.go          # Task structs
│   │   └── data/repository/
│   │       └── repository.go       # Postgres task repo
│   ├── comment/
│   │   ├── comment.go              # Comment module
│   │   ├── structs/
│   │   │   └── structs.go          # Comment structs
│   │   └── data/repository/
│   │       └── repository.go       # Postgres comment repo
│   └── realtime/
│       └── realtime.go             # WebSocket hub & handlers
├── plugin/                         # Plugin modules
│   ├── notification/
│   │   └── notification.go         # Email notification plugin
│   └── export/
│       ├── export.go               # Export job plugin
│       ├── structs/
│       │   └── structs.go          # Export job structs
│       └── data/repository/
│           └── repository.go       # Mongo export job repo
├── internal/
│   ├── event/
│   │   └── bus.go                  # Event bus implementation
│   └── server/
│       └── server.go               # Server & extension manager
├── config.yaml                      # Application configuration
├── go.mod
└── README.md
```

## Module Descriptions

### Core Modules

#### Auth Module (`core/auth`)

- User registration with password validation
- JWT token generation and validation
- Login with email/password
- Token refresh mechanism
- Authentication middleware
- Role-based authorization middleware

#### User Module (`core/user`)

- User CRUD operations
- User listing with pagination
- Role management (user, admin)

#### Workspace Module (`core/workspace`)

- Workspace CRUD operations
- Member management (add/remove)
- Role-based access control (owner, admin, member)
- Workspace listing for users

### Business Modules

#### Task Module (`biz/task`)

- Task CRUD operations
- Task assignment to users
- Status tracking (pending, in_progress, completed)
- Priority management (low, medium, high)
- Workspace-scoped tasks
- Event emission on all operations

#### Comment Module (`biz/comment`)

- Comment CRUD operations
- Task-comment relationship
- Ownership-based update/delete
- Event emission on all operations

#### Realtime Module (`biz/realtime`)

- WebSocket connection management
- Room-based broadcasting by workspace
- Event subscription and forwarding
- Connection health monitoring
- Hub statistics

### Plugin Modules

#### Notification Plugin (`plugin/notification`)

- Event-driven email notifications
- Task creation alerts
- Task assignment notifications
- Comment notifications
- Email template support

#### Export Plugin (`plugin/export`)

- Background export jobs (CSV, JSON)
- Task and comment export
- Job status tracking
- Progress updates via events
- Worker pool integration

## Configuration

```yaml
server:
  host: 0.0.0.0
  port: 8080
  mode: debug

data:
  database:
    master:
      driver: postgres
      source: "host=localhost port=5432 user=postgres password=postgres dbname=fullappdb sslmode=disable"
      max_idle_conns: 10
      max_open_conns: 100
  mongodb:
    master:
      uri: "mongodb://localhost:27017/fullappdb"
      logging: false
    strategy: "round_robin"
    max_retry: 3
  redis:
    addr: "localhost:6379"
    db: 0
    read_timeout: 3s
    write_timeout: 3s
    dial_timeout: 5s

auth:
  jwt:
    secret: "your-secret-key-change-in-production"
    access_token_ttl: 900 # 15 minutes
    refresh_token_ttl: 604800 # 7 days
  password:
    min_length: 8
    require_uppercase: true
    require_number: true

worker:
  pool_size: 10
  queue_size: 1000
  shutdown_timeout: 30

email:
  smtp:
    host: "smtp.gmail.com"
    port: "587"
    username: "noreply@example.com"
    password: "your-smtp-password"
    from: "noreply@example.com"

logger:
  level: debug
  format: json
  output: stdout
```

## API Endpoints

### Authentication

```text
POST   /auth/register    # Register new user
POST   /auth/login       # Login and get tokens
POST   /auth/refresh     # Refresh access token
POST   /auth/logout      # Logout
```

### Users

```text
POST   /users                # Create user (admin)
GET    /users                # List users (admin)
GET    /users/:user_id            # Get user by ID
PUT    /users/:user_id            # Update user
DELETE /users/:user_id            # Delete user
```

### Workspaces

```text
POST   /workspaces                    # Create workspace
GET    /workspaces                    # List workspaces for user
GET    /workspaces/:workspace_id      # Get workspace by ID
POST   /workspaces/:workspace_id/members        # Add member
GET    /workspaces/:workspace_id/members        # List members
```

### Tasks

```text
POST   /workspaces/:workspace_id/tasks   # Create task in workspace
GET    /workspaces/:workspace_id/tasks   # List tasks in workspace
GET    /tasks/:task_id                      # Get task by ID
PUT    /tasks/:task_id                      # Update task
DELETE /tasks/:task_id                      # Delete task
POST   /tasks/:task_id/assign               # Assign task to user
```

### Comments

```text
POST   /workspaces/:workspace_id/comments   # Create comment
GET    /tasks/:task_id/comments         # List comments for task
GET    /comments/:comment_id                   # Get comment by ID
PUT    /comments/:comment_id                   # Update comment
DELETE /comments/:comment_id                   # Delete comment
```

### Real-time (WebSocket)

```text
GET    /ws/:workspace_id    # WebSocket connection for workspace
```

### Exports

```text
POST   /workspaces/:workspace_id/exports   # Create export job
GET    /workspaces/:workspace_id/exports   # List export jobs
GET    /exports/:job_id                  # Get export job status
```

### System

```text
GET    /health               # Health check
GET    /events/stats         # Event bus statistics
GET    /realtime/stats      # WebSocket hub statistics
```

## Event Flow Examples

### Task Creation Flow

```text
User creates task
    ↓
Task Service saves task to database
    ↓
Task Service publishes "task.created" event
    ↓
┌─────────────────┬─────────────────┐
│                 │                 │
▼                 ▼                 ▼
WebSocket Hub   Email Notification  Event Store
Broadcasts to    Sends email       Persists
Workspace        to members
```

### Comment Creation Flow

```text
User adds comment
    ↓
Comment Service saves comment to database
    ↓
Comment Service publishes "comment.created" event
    ↓
┌─────────────────┬─────────────────┐
│                 │                 │
▼                 ▼                 ▼
WebSocket Hub   Email Notification  Event Store
Broadcasts to    Sends email       Persists
Workspace
```

### Export Job Flow

```text
User requests export
    ↓
Export Service creates job (status: pending)
    ↓
Job submitted to Worker Pool
    ↓
Job status: running
    ↓
Worker executes export (CSV/JSON)
    ↓
Job status: completed (or failed)
    ↓
Service publishes "export.completed" event
```

## Running the Application

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- MongoDB 6+
- Redis 7+

### Installation

```bash
# Navigate to example directory
cd 08-full-application

# Download dependencies
go mod download

# Generate Ent client for user module
go generate ./core/user/data

# Run the application
go run main.go -config config.yaml
```

### Usage

```bash
# Start the application
go run main.go

# Application will start on http://localhost:8080

# Check health
curl http://localhost:8080/health

# View event bus stats
curl http://localhost:8080/events/stats

# View WebSocket stats
curl http://localhost:8080/realtime/stats
```

## API Examples

### Register User

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "SecurePass123"
  }'

# Response:
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "John Doe",
  "email": "john@example.com",
  "role": "user",
  "created_at": "2024-01-15T12:00:00Z"
}
```

### Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123"
  }'

# Response:
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

### Create Workspace

```bash
curl -X POST http://localhost:8080/workspaces \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "name": "Engineering Team",
    "description": "Engineering workspace"
  }'
```

### Create Task

```bash
curl -X POST http://localhost:8080/workspaces/{workspace_id}/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "title": "Implement new feature",
    "description": "Add the new feature to the application",
    "priority": "high",
    "status": "pending"
  }'
```

### Add Comment

```bash
curl -X POST http://localhost:8080/workspaces/{workspace_id}/comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "task_id": "{task_id}",
    "content": "This looks good, let's proceed"
  }'
```

### Request Export

```bash
curl -X POST http://localhost:8080/workspaces/{workspace_id}/exports \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "type": "tasks",
    "format": "csv"
  }'

# Response:
{
  "id": "job-id",
  "type": "tasks",
  "format": "csv",
  "status": "pending",
  "created_at": "2024-01-15T12:00:00Z"
}
```

### WebSocket Connection

```javascript
// Connect to workspace WebSocket
const ws = new WebSocket(
  "ws://localhost:8080/ws/{workspace_id}?token={access_token}"
);

// Listen for events
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log("Event received:", message);

  // Message format:
  // {
  //   "type": "task.created",
  //   "workspace_id": "...",
  //   "user_id": "...",
  //   "data": { ... },
  //   "timestamp": 1234567890
  // }
};

// Send message to workspace
ws.send(
  JSON.stringify({
    type: "custom.message",
    data: {
      text: "Hello workspace!",
    },
  })
);
```

## Key Patterns Demonstrated

### 1. Extension Architecture

All modules implement the `extension.Extension` interface:

```go
type Module struct {
    name    string
    service *Service
    logger  *logger.Logger
}

func (m *Module) Init(ctx context.Context, deps extension.Dependencies) error
func (m *Module) PostInit(ctx context.Context) error
func (m *Module) Name() string
func (m *Module) Routes() []extension.Route
func (m *Module) Cleanup(ctx context.Context) error
```

### 2. Event-Driven Communication

Modules communicate via the event bus:

```go
// Publish event
s.bus.Publish(ctx, &event.Event{
    Type:          event.EventTypeTaskCreated,
    AggregateID:   task.ID,
    WorkspaceID:   task.WorkspaceID,
    Payload: map[string]interface{}{
        "task_id": task.ID,
        "title": task.Title,
    },
})

// Subscribe to event
s.bus.Subscribe(event.EventTypeTaskCreated, func(ctx context.Context, evt *event.Event) error {
    // Handle event
    return nil
})
```

### 3. Multi-Tenancy Pattern

Workspace isolation ensures data separation:

```go
// All operations are workspace-scoped
func (s *Service) CreateTask(ctx context.Context, workspaceID string, req *CreateTaskRequest) (*Task, error) {
    task := &Task{
        WorkspaceID: workspaceID,
        // ... other fields
    }
    // ...
}
```

### 4. Two-Phase Initialization

Extensions support dependency resolution:

```go
// Phase 1: Init - All extensions initialized
if err := manager.Init(ctx); err != nil {
    return err
}

// Phase 2: PostInit - All extensions ready
if err := manager.PostInit(ctx); err != nil {
    return err
}
```

## Production Considerations

### Database Migration

Use Postgres for core data, MongoDB for events/exports, and Redis for caching. For production:

1. Create database migrations for Postgres tables
2. Validate MongoDB indexes for event and export collections
3. Use transaction support for complex operations
4. Add indexes for workspace and task queries

### Authentication Security

For production:

- Use strong JWT secrets (environment variables)
- Implement refresh token rotation
- Add rate limiting to auth endpoints
- Implement CAPTCHA for registration
- Add password reset flow

### WebSocket Scaling

For production:

- Use Redis Pub/Sub for multi-instance deployments
- Implement connection limits per user
- Add reconnection logic with backoff
- Use message queue for high-throughput scenarios

### Email Configuration

For production:

- Use SMTP with TLS
- Configure DKIM/SPF records
- Implement email queue for rate limiting
- Add unsubscribe functionality

### Monitoring

Add observability:

- Metrics for event bus (published/processed/failed)
- Metrics for WebSocket connections
- Metrics for job queue (pending/running/completed)
- Structured logging with correlation IDs

## Testing

```bash
# Unit tests
go test ./...

# Integration tests
go test -tags=integration ./...

# Coverage
go test -cover ./...

# Benchmark
go test -bench=. -benchmem
```

## Real-World Application Patterns

This example incorporates patterns from actual NCore-based applications:

| Application | Pattern | Implementation |
|-------------|---------|----------------|

## NCore Modules Used

| Module               | Usage                               |
|----------------------|-------------------------------------|
| `config`             | Configuration management            |
| `logging`            | Structured logging                  |
| `extension/manager`  | Module registration and lifecycle   |
| `security/jwt`       | JWT token generation and validation |
| `concurrency/worker` | Background job processing           |
| `email`              | Email sending (notifications)       |
| `event`              | Custom event bus implementation     |

## Next Steps

- Add [database persistence](../01-basic-rest-api) with Ent ORM
- Explore [multi-module patterns](../03-multi-module) in detail
- Add [OAuth 2.0](../07-authentication) providers
- See [NCore main documentation](../../README.md) for more topics

## License

This example is part of the NCore project.
