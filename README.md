# NCore

> A set of reusable components for Go applications.

## Features

- Extensible modular architecture
  - Core domain and business domain module development
  - Plugin system for feature extensions
  - Event-driven module communication
- Universal proxy system
  - Platform API integration
  - Third-party service adapters
  - Protocol transformation
- Comprehensive business logic helpers
- Flexible configuration management
- Secure authentication (JWT, OAuth)
- Multiple database support
- Search engine integrations
- File storage solutions

## Installation

```bash
# Install the latest version
go get github.com/ncobase/ncore

# Install a specific version
go get github.com/ncobase/ncore@v1.0.0
```

## Quick Start

```bash
# Install the CLI tool
go install github.com/ncobase/ncore/cmd/ncore@latest

# View available commands and options
ncore --help

# Start with example project
ncore create example --standalone
```

## Code Generation

Provides a powerful code generation tool to scaffold extensions and applications:

```bash
# Create a core extension
ncore create core auth-service

# Create with cmd directory (extension + executable)
ncore create core auth-service --with-cmd

# Create standalone application
ncore create core auth-service --standalone

# Additional options
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

- Go 1.21 or higher

## Support

- Issue Tracker: [https://github.com/ncobase/ncore/issues](https://github.com/ncobase/ncore/issues)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
