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
go get github.com/ncobase/ncore
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

# More options
ncore --help
```

## Structure

```plaintext
├── pkg/               # Public packages that can be imported by other projects
│   ├── concurrency/   # Concurrency management
│   ├── config/        # Configuration
│   ├── consts/        # Constants
│   ├── cookie/        # Cookie handling
│   ├── crypto/        # Encryption utilities
│   ├── data/          # Data access
│   ├── ecode/         # Error codes
│   ├── email/         # Email functionality
│   ├── expression/    # Expression evaluation and parsing
│   ├── helper/        # Helper functions
│   ├── jwt/           # JWT handling
│   ├── logger/        # Logging
│   ├── monitor/       # Monitoring
│   ├── nanoid/        # ID generation
│   ├── oauth/         # OAuth utilities
│   ├── observes/      # Observability
│   ├── paging/        # Pagination
│   ├── queue/         # Queue management
│   ├── resp/          # HTTP responses
│   ├── types/         # Common types
│   ├── slug/          # URL slugs
│   ├── storage/       # Storage utilities
│   ├── uuid/          # UUID utilities
│   ├── validator/     # Validation utilities
│   └── worker/        # Worker pool
├── ext/               # Extension system
│   ├── core/          # Interfaces and types
│   ├── discovery/     # Service discovery
│   ├── event/         # Event
│   ├── manager/       # Extension manager
│   └── plugin/        # Plugin
└── cmd/               # Application entry points
    ├── commands/      # Command-line tools
    └── generator/     # Code generation utilities
```

## Dependencies

- Go 1.21 or higher

## Support

- Issue Tracker: [https://github.com/ncobase/ncore/issues](https://github.com/ncobase/ncore/issues)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
