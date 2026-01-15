# Example 03: Multi-Module Application

Demonstrates NCore's extension system with multiple modules communicating through the Extension Manager, showcasing patterns from real-world applications.

## Features

- **Extension System**: Dynamic module loading and lifecycle management
- **Inter-Module Communication**: Cross-service calls via Extension Manager
- **Wrapper Pattern**: Safe cross-module service access
- **Two-Phase Initialization**: `Init()` and `PostInit()` lifecycle
- **Weak Dependencies**: Optional module dependencies
- **Event-Driven**: Pub/sub for module coordination

## Architecture

```text
┌──────────────────────────────────────┐
│       Extension Manager              │
│  (Service Discovery & Lifecycle)     │
└──────────────────────────────────────┘
         ▲          ▲          ▲
         │          │          │
    ┌────┴───┐ ┌───┴────┐ ┌───┴─────┐
    │  User  │ │  Post  │ │Comment  │
    │ Module │ │ Module │ │ Module  │
    └────────┘ └────────┘ └─────────┘
```

## Modules

### Core Modules

- **user**: User management and authentication
- **post**: Content creation and management

### Business Modules

- **comment**: Comments on posts (depends on user & post)

## Key Patterns

### 1. Module Registration

```go
func init() {
    exr.RegisterToGroupWithWeakDeps(
        exr.ModuleCore,
        Meta(),
        []string{}, // Hard dependencies
        []string{"auth"}, // Weak dependencies
    )
}
```

### 2. Two-Phase Initialization

```go
func (m *Module) Init(cfg *config.Config, em types.ManagerInterface) error {
    // Phase 1: Initialize own resources
    m.data = data.New(cfg)
    return nil
}

func (m *Module) PostInit() error {
    // Phase 2: Wire up dependencies (other modules are ready)
    m.service = service.New(m.data, m.em)
    m.handler = handler.New(m.service)
    return nil
}
```

### 3. Cross-Module Communication (Wrapper Pattern)

```go
// wrapper/user_service.go
type UserServiceWrapper struct {
    em types.ManagerInterface
}

func (w *UserServiceWrapper) GetUser(ctx context.Context, id string) (*User, error) {
    svc, err := w.em.GetCrossService("user", "UserService")
    if err != nil {
        return nil, err
    }
    userSvc := svc.(UserService)
    return userSvc.GetUser(ctx, id)
}
```

## Project Structure

```text
03-multi-module/
├── core/
│   ├── user/                # User module
│   │   ├── user.go          # Module definition
│   │   ├── handler/
│   │   ├── service/
│   │   ├── structs/
│   │   └── data/
│   │       └── repository/
│   └── post/                # Post module
│       ├── post.go
│       ├── handler/
│       ├── service/
│       ├── structs/
│       ├── data/
│       │   └── repository/
│       └── wrapper/         # Cross-module wrappers
│           └── user_service.go
├── biz/
│   └── comment/             # Comment module
│       ├── comment.go
│       ├── handler/
│       ├── service/
│       ├── data/
│       └── wrapper/
│           ├── user_service.go
│           └── post_service.go
├── internal/
│   └── server/
│       └── server.go        # Extension Manager setup
├── main.go
├── config.yaml
└── README.md
```

## Running

```bash
# Run the application
go run main.go

# Modules auto-register via init()
# Extension Manager coordinates lifecycle
# HTTP routes aggregated from all modules
```

## API Examples

```bash
# Create user
curl -X POST http://localhost:8080/api/v1/users \
  -d '{"name":"Alice","email":"alice@example.com"}'

# Create post (requires user)
curl -X POST http://localhost:8080/api/v1/posts \
  -d '{"user_id":"123","title":"Hello","content":"World"}'

# Create comment (requires user & post via wrappers)
curl -X POST http://localhost:8080/api/v1/comments \
  -d '{"user_id":"123","post_id":"456","content":"Nice post!"}'
```

1. **Extension Architecture**: How to structure modular applications
2. **Dependency Management**: Avoiding circular imports
3. **Service Discovery**: Runtime service location
4. **Lifecycle Hooks**: Two-phase initialization
5. **Weak Dependencies**: Optional module loading

## Next Steps

- Add [event-driven communication](../06-event-driven)
- Integrate [authentication](../07-authentication)
- Scale with [background jobs](../05-background-jobs)

## License

This example is part of the NCore project.
