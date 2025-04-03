# NCore Library

> NCore library provides a set of reusable components and utilities for building modern Go applications.

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

## Structure

```plaintext
├── biz                 # Business logic, helper functions
├── config              # Configuration helpers and management
├── consts              # Constants and predefined values
├── cookie              # Cookie handling and management
├── crypto              # Encryption/decryption utilities and security tools
├── data                # Data handling and persistence
│   ├── cache           # Cache management and operations
│   ├── connection      # Database connections and connection pool management
│   ├── elastic         # Elasticsearch integration and operations
│   ├── entgo           # Entgo ORM support and schema management
│   │   └── mixin       # Entgo mixins for common fields and behaviors
│   ├── meili           # Meilisearch integration and search operations
│   └── service         # Direct service support and implementations
├── ecode               # Error codes and error handling utilities
├── email               # Email templates, sending and management
├── expression          # Expression evaluation and parsing
├── extension           # Extension system for module and plugin development, event management
├── helper              # Common helper functions and utilities
├── jwt                 # JWT generation, validation and management
├── log                 # Logging infrastructure and formatters
├── metrics             # Metrics collection and reporting
├── nanoid              # NanoID generation for unique identifiers
├── oauth               # OAuth2 authentication and authorization
├── observes            # Observers, monitoring and metrics collection
├── paging              # Pagination utilities and cursor implementation
├── proxy               # Universal proxy system for platform and third-party service integration
├── queue               # Queue management and operations
├── resp                # HTTP response handling and formatting
├── router              # Router configuration and middleware management
├── slug                # URL-friendly slug generation and validation
├── storage             # File storage and management (local/cloud)
├── time                # Time formatting, parsing and timezone utilities
├── types               # Common type definitions and interfaces
├── util                # General utility functions and tools
├── uuid                # UUID generation and validation
├── validator           # Data validation and sanitization
└── worker              # Background worker support
```

## Dependencies

- Go 1.21 or higher

## Support

- Issue Tracker: [https://github.com/ncobase/ncore/issues](https://github.com/ncobase/ncore/issues)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
