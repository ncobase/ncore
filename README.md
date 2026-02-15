# NCore

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/doc/devel/release)
[![Release](https://img.shields.io/badge/release-v0.2.0-blue)](https://github.com/ncobase/ncore/releases/tag/v0.2.0)
[![License](https://img.shields.io/badge/license-Apache%202.0-green)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/ncobase/ncore)](https://goreportcard.com/report/github.com/ncobase/ncore)

NCore is a modular Go framework providing enterprise-grade components for building production-ready applications. It features a multi-module architecture with opt-in drivers, minimal dependencies, and comprehensive integrations.

## Features

- **Modular Architecture**: Import only the modules you need
- **Modular Driver System** (v0.2.0+): Opt-in database, cache, search, and messaging drivers for minimal binary size
- **Rich Integrations**: PostgreSQL, MySQL, MongoDB, Redis, Elasticsearch, Kafka, and more
- **Security & Authentication**: JWT, OAuth, encryption utilities
- **Observability**: OpenTelemetry, logging, and monitoring
- **Dependency Injection**: Native support for Google Wire

## Multi-Module Architecture

Each sub-package is an independent Go module with minimal dependencies and independent versioning.

**Core Modules:** config, data, logging, security, extension, oss, validation, messaging

See [MODULES.md](MODULES.md) for complete module list and architecture details.

## Installation

Import only the modules you need:

```bash
# Core modules
go get github.com/ncobase/ncore/config
go get github.com/ncobase/ncore/data
go get github.com/ncobase/ncore/security
go get github.com/ncobase/ncore/oss

# Data drivers (v0.2.0+) - import only what you use
go get github.com/ncobase/ncore/data/postgres
go get github.com/ncobase/ncore/data/redis
go get github.com/ncobase/ncore/data/meilisearch
```

## Quick Start

```go
package main

import (
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/data"

    // Import only the drivers you need (v0.2.0+)
    _ "github.com/ncobase/ncore/data/postgres"
    _ "github.com/ncobase/ncore/data/redis"
)

func main() {
    // Load configuration
    cfg, err := config.LoadConfig("config.yaml")
    if err != nil {
        panic(err)
    }

    // Initialize data layer (drivers auto-register on import)
    d, cleanup, err := data.New(cfg.Data)
    if err != nil {
        panic(err)
    }
    defer cleanup()

    // Use your database and cache
    db := d.Conn.DB()
    redis := d.Conn.RC
}
```

### Migration from v0.1.x

Add driver imports to main.go:

```go
import (
    _ "github.com/ncobase/ncore/data/postgres"
    _ "github.com/ncobase/ncore/data/redis"
    _ "github.com/ncobase/ncore/data/elasticsearch"
)
```

See [MIGRATION_v0.1_to_v0.2.md](MIGRATION_v0.1_to_v0.2.md) for complete migration guide.

### Performance

| Metric       | v0.1.x | v0.2.0 | Improvement |
| ------------ | ------ | ------ | ----------- |
| Binary Size  | ~92MB  | ~43MB  | **-53%**    |
| Dependencies | 466    | ~100   | **-78%**    |
| Build Time   | ~45s   | ~20s   | **-56%**    |

## Dependency Injection (Google Wire)

NCore provides native [Google Wire](https://github.com/google/wire) support with `ProviderSet` in each module.

```go
//go:build wireinject

func InitializeApp() (*App, func(), error) {
    panic(wire.Build(
        config.ProviderSet,
        logger.ProviderSet,
        data.ProviderSet,
        security.ProviderSet,
        NewApp,
    ))
}
```

See [DEVELOPMENT.md](DEVELOPMENT.md#6-dependency-injection-google-wire) for details.

## Development

```bash
# Clone the repository
git clone https://github.com/ncobase/ncore.git
cd ncore

# Sync dependencies
go work sync

# Run tests
bash scripts/test.sh
```

## Examples

See [examples/README.md](./examples/README.md) for a detailed overview and learning paths.

## Documentation

- [DEVELOPMENT.md](DEVELOPMENT.md) - Development guide
- [MODULES.md](MODULES.md) - Multi-module architecture explanation

## Code Generation

For scaffolding new projects and components, use the CLI tool:

```bash
go install github.com/ncobase/cli@latest
nco create core auth-service
nco create business payment --use-mongo --with-test
```

## Contributing

Contributions are welcome! Please read our [Contributing Guidelines](CONTRIBUTING.md) before submitting pull requests.

## License

Copyright 2023-present Ncobase

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.
