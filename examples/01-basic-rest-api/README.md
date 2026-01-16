# Example 01: Basic REST API

A simple REST API demonstrating the fundamental usage of NCore with Google Wire dependency injection, PostgreSQL
database using Ent ORM, and basic CRUD operations.

## Features

- **Google Wire Integration**: Demonstrates Wire-based dependency injection
- **PostgreSQL + Ent ORM**: Database operations with type-safe schema
- **RESTful API**: Standard CRUD endpoints for a Task resource
- **Clean Architecture**: Handler → Service → Repository pattern
- **Configuration Management**: YAML-based configuration
- **Logging**: Structured logging with NCore logger
- **Error Handling**: Consistent error responses

## Project Structure

```text
01-basic-rest-api/
├── main.go                  # Application entry point
├── handler/                 # HTTP handlers
│   ├── handler.go
│   └── task.go
├── service/                 # Business logic
│   ├── service.go
│   └── task.go
├── data/                    # Data access layer
│   ├── data.go
│   ├── repository/
│   │   └── task.go
│   └── schema/              # Ent schema
│       └── task.go
├── wire.go                  # Wire injector
├── wire_gen.go              # Wire generated code
├── config.yaml              # Configuration file
├── go.mod
└── README.md
```

## Prerequisites

- Go 1.21+
- PostgreSQL 14+
- [Google Wire](https://github.com/google/wire) CLI tool

## Installation

```bash
# Install Wire
go install github.com/google/wire/cmd/wire@latest

# Install Ent CLI
go install entgo.io/ent/cmd/ent@latest

# Install dependencies
go mod download
```

## Setup

### 1. Configure Database

Edit `config.yaml` and update the database connection:

```yaml
data:
  database:
    master:
      driver: postgres
      source: "host=localhost port=5432 user=postgres password=postgres dbname=taskdb sslmode=disable"
```

### 2. Generate Ent Code

```bash
go generate ./data
```

### 3. Generate Wire Code

```bash
wire ./...
```

### 4. Run Database Migrations

The application will automatically create tables on startup.

## Running

```bash
go run main.go
```

The server starts on `http://localhost:8080` by default.

## API Endpoints

### Create Task

```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Complete documentation",
    "description": "Write comprehensive API docs",
    "status": "pending"
  }'
```

### List Tasks

```bash
curl http://localhost:8080/tasks
```

### Get Task

```bash
curl http://localhost:8080/tasks/1
```

### Update Task

```bash
curl -X PUT http://localhost:8080/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Complete documentation",
    "description": "Write comprehensive API docs",
    "status": "completed"
  }'
```

### Delete Task

```bash
curl -X DELETE http://localhost:8080/tasks/1
```

## Key Learning Points

### 1. Wire Dependency Injection

The `wire.go` file defines how dependencies are wired together:

```go
func InitializeApp() (*App, func(), error) {
    panic(wire.Build(
        config.ProviderSet,
        logger.ProviderSet,
        data.ProviderSet,
        NewApp,
    ))
}
```

Wire automatically generates the initialization code in `wire_gen.go`.

### 2. Clean Architecture Layers

**Handler Layer** (`handler/task.go`):

- Handles HTTP requests/responses
- Validates input
- Calls service layer

**Service Layer** (`service/task.go`):

- Contains business logic
- Orchestrates repository operations
- Returns domain errors

**Repository Layer** (`data/repository/task.go`):

- Abstracts database operations
- Uses Ent client for queries
- Returns models or errors

### 3. NCore Module Usage

- **config**: Centralized configuration management
- **logger**: Structured logging
- **data**: Database connection pooling and management
- **net/resp**: Standardized API responses

## Configuration

The `config.yaml` file controls all application settings:

```yaml
server:
  host: 0.0.0.0
  port: 8080
  mode: debug

data:
  database:
    master:
      driver: postgres
      source: "postgres://user:pass@localhost:5432/db"
      max_idle_conns: 10
      max_open_conns: 100

logger:
  level: debug
  format: json
  output: stdout
```

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...
```

## Next Steps

- Explore [02-mongodb-api](../02-mongodb-api) for MongoDB integration without Wire
- Explore [03-multi-module](../03-multi-module) for extension-based architecture
- Add authentication (see [07-authentication](../07-authentication))
- Add background jobs (see [05-background-jobs](../05-background-jobs))

## License

This example is part of the NCore project.
