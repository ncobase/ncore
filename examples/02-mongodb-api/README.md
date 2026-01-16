# Example 02: MongoDB API

A REST API demonstrating NCore usage with MongoDB (without Google Wire), showcasing manual dependency injection patterns
used in production applications.

## Features

- **MongoDB Integration**: Native MongoDB driver with connection pooling
- **Manual Dependency Injection**: No Wire, explicit constructor injection
- **User Management**: Complete CRUD for user resources
- **Clean Architecture**: Handler → Service → Repository pattern
- **Configuration Management**: YAML-based configuration
- **Validation**: Request validation and error handling

## Project Structure

```text
02-mongodb-api/
├── main.go              # Application entry point
├── handler/             # HTTP handlers
│   ├── handler.go
│   └── user.go
├── service/             # Business logic
│   ├── service.go
│   └── user.go
└── data/                # Data access layer
    ├── data.go
    └── repository/
        └── user.go
├── config.yaml              # Configuration file
├── go.mod
└── README.md
```

## Key Differences from Example 01

| Feature     | Example 01              | Example 02                  |
|-------------|-------------------------|-----------------------------|
| Database    | PostgreSQL + Ent ORM    | MongoDB (native driver)     |
| DI Strategy | Google Wire (automatic) | Manual (explicit)           |
| Pattern     | Wire-based injection    | Constructor-based injection |
| Schema      | Code-first (Ent)        | Schema-less (MongoDB)       |

## Prerequisites

- Go 1.21+
- MongoDB 6.0+

## Installation

```bash
# Install dependencies
go mod download
```

## Setup

### 1. Configure Database

Edit `config.yaml`:

```yaml
data:
  database:
    master:
      driver: mongodb
      source: "mongodb://localhost:27017/userdb"
```

### 2. Start MongoDB

```bash
# Using Docker
docker run -d -p 27017:27017 --name mongodb mongo:7

# Or using local installation
mongod --dbpath /path/to/data
```

## Running

```bash
go run main.go
```

The server starts on `http://localhost:8080`.

## API Endpoints

### Create User

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "role": "user"
  }'
```

### List Users

```bash
curl http://localhost:8080/api/v1/users
```

### Get User

```bash
curl http://localhost:8080/api/v1/users/507f1f77bcf86cd799439011
```

### Update User

```bash
curl -X PUT http://localhost:8080/api/v1/users/507f1f77bcf86cd799439011 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Smith",
    "email": "john.smith@example.com",
    "role": "admin"
  }'
```

### Delete User

```bash
curl -X DELETE http://localhost:8080/api/v1/users/507f1f77bcf86cd799439011
```

## Key Learning Points

### 1. Manual Dependency Injection

Unlike Example 01 which uses Wire, this example demonstrates explicit dependency injection:

```go
func main() {
    // Load config
    cfg, err := config.LoadConfig("config.yaml")

    // Create logger
    logger, cleanup := logger.New(cfg.Logger)
    defer cleanup()

    // Create data layer
    dataLayer, err := data.New(cfg, logger)
    defer dataLayer.Close()

    // Create service layer
    svc := service.NewService(dataLayer, logger)

    // Create handler layer
    handler := handler.NewHandler(svc, logger)
}
```

### 2. MongoDB Pattern

The repository pattern with MongoDB:

```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id string) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    List(ctx context.Context, skip, limit int64) ([]*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}
```

### 3. Two-Phase Initialization

Data layer initialization is split into two phases (used in production systems):

- **Phase 1**: Connect to database, validate configuration
- **Phase 2**: Initialize repositories with the connection

```bash
go test ./...
```

## Next Steps

- Explore [03-multi-module](../03-multi-module) for extension-based architecture
- Add authentication (see [07-authentication](../07-authentication))
- Integrate with Wire if needed (see [01-basic-rest-api](../01-basic-rest-api))

## License

This example is part of the NCore project.
