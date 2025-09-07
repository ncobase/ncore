# NCore

Go application components library.

## Features

- Modular architecture
  - Core and business modules
  - Plugin system
  - Event-driven communication
- Proxy system
  - API integration
  - Service adapters
  - Protocol transformation
- Business logic helpers
- Configuration management
- Authentication (JWT, OAuth)
- Database support
- Search engine integrations
- File storage

## Installation

```bash
go get github.com/ncobase/ncore
go get github.com/ncobase/ncore@v1.0.0
```

## Quick Start

```bash
go install github.com/ncobase/ncore/cmd@latest
ncore --help
ncore create example --standalone
```

## Code Generation

```bash
ncore create core auth-service
ncore create core auth-service --with-cmd
ncore create core auth-service --standalone
ncore create business payment --use-mongo --with-test
```

## Structure

```plaintext
├── cmd/               # Application entry points and code generation
│   ├── commands/      # Command-line tools
│   ├── generator/     # Code generation utilities
│   └── main.go        # Main application entry point
├── concurrency/       # Concurrency management tools
│   └── worker/        # Worker pool implementation
├── config/            # Configuration management
├── consts/            # Constant definitions
├── ctxutil/           # Context utilities and helpers
├── data/              # Data access and management
│   ├── config/        # Data source configurations
│   ├── connection/    # Connection management for data sources
│   ├── databases/     # Database specific implementations
│   ├── messaging/     # Data messaging implementations
│   ├── paging/        # Pagination utilities
│   ├── search/        # Search engine integrations
│   └── storage/       # File storage solutions
├── ecode/             # Error codes and error handling
├── extension/         # Extension system
│   ├── discovery/     # Service discovery
│   ├── event/         # Event handling
│   ├── manager/       # Extension management
│   ├── plugin/        # Plugin system
│   └── types/         # Extension system types and interfaces
├── logging/           # Logging and monitoring
│   ├── logger/        # Logging utilities
│   ├── monitor/       # System monitoring
│   └── observes/      # Observability tools
├── messaging/         # Messaging services
│   ├── email/         # Email functionality
│   └── queue/         # Queue management
├── net/               # Network utilities
│   ├── cookie/        # Cookie handling
│   └── resp/          # HTTP responses
├── security/          # Security utilities
│   ├── crypto/        # Encryption utilities
│   ├── jwt/           # JWT handling
│   └── oauth/         # OAuth utilities
├── types/             # Common types and type utilities
├── utils/             # Utility functions
│   ├── nanoid/        # ID generation
│   ├── slug/          # URL slugs
│   └── uuid/          # UUID utilities
├── validation/        # Validation utilities
│   ├── expression/    # Expression evaluation and parsing
│   └── validator/     # Validation tools
└── version/           # Version information
```

## Dependencies

Go 1.21+

## Support

[Issues](https://github.com/ncobase/ncore/issues)

## License

Apache License 2.0
