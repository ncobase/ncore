# NCore

A comprehensive Go application components library for building modern, scalable applications.

## Features

- **Modular Architecture**
  - Core and business modules
  - Plugin system
  - Event-driven communication
- **Rich Integrations**
  - Database support (PostgreSQL, MySQL, MongoDB, etc.)
  - Search engines (Elasticsearch, OpenSearch, Meilisearch)
  - Message queues (RabbitMQ, Kafka, Redis)
  - Storage solutions (AWS S3, cloud storage)
- **Security & Authentication**
  - JWT and OAuth utilities
  - Encryption and crypto utilities
  - Security middleware
- **Observability**
  - OpenTelemetry integration
  - Logging and monitoring
  - Error tracking

## Installation

```bash
go get github.com/ncobase/ncore
```

## Quick Start

```go
package main

import (
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/data/databases"
    "github.com/ncobase/ncore/security/jwt"
)

func main() {
    // Load configuration
    cfg, err := config.LoadConfig("config.yaml")
    if err != nil {
        panic(err)
    }

    // Initialize database
    db, err := databases.NewPostgreSQL(cfg.Database.DSN)
    if err != nil {
        panic(err)
    }

    // Create JWT manager
    jwtManager := jwt.NewManager(cfg.Security.JWTSecret)
}
```

## Code Generation

For scaffolding new projects and components, use the separate CLI tool:

```bash
# Install CLI tool
go install github.com/ncobase/cli@latest

# Generate new components
nco create core auth-service
nco create business payment --use-mongo --with-test
```

## Library Structure

```plaintext
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

## Core Components

### Database Support

- **PostgreSQL**: Full-featured support with connection pooling
- **MySQL**: High-performance MySQL integration
- **MongoDB**: Document database support
- **SQLite**: Embedded database for development/testing
- **Neo4j**: Graph database integration

### Search Engines

- **Elasticsearch**: Full-text search and analytics
- **OpenSearch**: Open-source search and analytics
- **Meilisearch**: Fast, typo-tolerant search

### Messaging Systems

- **RabbitMQ**: Message queuing and pub/sub
- **Kafka**: Event streaming platform
- **Redis**: Caching and real-time messaging

### Security Features

- **JWT**: Token generation and validation
- **OAuth**: Multiple provider integration
- **Cryptography**: Encryption, hashing, signing utilities

## Support

[Issues](https://github.com/ncobase/ncore/issues)

## License

See [LICENSE](LICENSE) file for details.
