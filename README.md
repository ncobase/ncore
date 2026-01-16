# NCore

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat\u0026logo=go)](https://go.dev/doc/devel/release)
[![Release](https://img.shields.io/badge/release-v0.2.0-blue)](https://github.com/ncobase/ncore/releases/tag/v0.2.0)
[![License](https://img.shields.io/badge/license-Apache%202.0-green)](LICENSE)

A comprehensive Go application components library for building modern, scalable applications.

## Features

- **Modular Architecture**: Import only the modules you need
- **Modular Driver System** (v0.2.0+): Opt-in database, cache, search, and messaging drivers for minimal binary size
- **Rich Integrations**: PostgreSQL, MySQL, MongoDB, Redis, Elasticsearch, Kafka, and more
- **Security & Authentication**: JWT, OAuth, encryption utilities
- **Observability**: OpenTelemetry, logging, and monitoring
- **Dependency Injection**: Native support for Google Wire

## Multi-Module Architecture

NCore uses a **multi-module architecture** where each sub-package is an independent Go module, providing minimal
dependencies and independent versioning.

### Available Modules

```text
github.com/ncobase/ncore/concurrency    - Concurrency utilities
github.com/ncobase/ncore/config         - Configuration management
github.com/ncobase/ncore/consts         - Constants
github.com/ncobase/ncore/ctxutil        - Context utilities
github.com/ncobase/ncore/data           - Data layer (DB, cache, search)
github.com/ncobase/ncore/ecode          - Error codes
github.com/ncobase/ncore/extension      - Extension system
github.com/ncobase/ncore/logging        - Logging
github.com/ncobase/ncore/messaging      - Message queue
github.com/ncobase/ncore/net            - Network utilities
github.com/ncobase/ncore/oss            - Object Storage Service (S3, Azure, MinIO, etc.)
github.com/ncobase/ncore/security       - Security features
github.com/ncobase/ncore/types          - Common types
github.com/ncobase/ncore/utils          - Utility functions
github.com/ncobase/ncore/validation     - Validation
github.com/ncobase/ncore/version        - Version info
```

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

If you're upgrading from v0.1.x, add the driver imports to your main.go or any initialization file:

```go
import (
    _ "github.com/ncobase/ncore/data/postgres"  // Add your required drivers
    _ "github.com/ncobase/ncore/data/redis"
)
```

For search functionality, import the search module:

```go
import (
    "github.com/ncobase/ncore/data/search"
    _ "github.com/ncobase/ncore/data/elasticsearch"
)
```

For log-to-search-engine functionality:

```go
import (
    _ "github.com/ncobase/ncore/logging/hooks/elasticsearch"
    // or _ "github.com/ncobase/ncore/logging/hooks/meilisearch"
    // or _ "github.com/ncobase/ncore/logging/hooks/opensearch"
)
```

### Why Modular Drivers?

v0.2.0 introduces opt-in drivers that dramatically reduce binary size and dependencies:

| Metric                  | v0.1.x | v0.2.0 | Improvement |
|-------------------------|--------|--------|-------------|
| Binary Size (basic app) | ~92MB  | ~43MB  | **-53%**    |
| Dependencies            | 466    | ~100   | **-78%**    |
| Compilation Time        | ~45s   | ~20s   | **-56%**    |

**Available Drivers:**

- **Database**: postgres, mysql, sqlite, mongodb, neo4j
- **Cache**: redis
- **Search**: elasticsearch, opensearch, meilisearch (optional module)
- **Messaging**: kafka, rabbitmq
- **Storage**: s3, aliyun, minio, azure, tencent, qiniu, gcs, synology, local (via `oss` standalone module)

**Object Storage (OSS):**

Starting from v0.2.0, object storage has been extracted into a standalone `oss` module:
- **Direct use**: `import "github.com/ncobase/ncore/oss"` - no driver registration needed
- **9 providers**: AWS S3, Azure Blob, Aliyun OSS, Tencent COS, Google Cloud Storage, MinIO, Qiniu Kodo, Synology NAS, Local Filesystem

The modular driver system uses an init-based registration pattern for automatic driver discovery.

## Dependency Injection (Google Wire)

NCore provides native support for [Google Wire](https://github.com/google/wire). You can use the pre-defined
`ProviderSet` in each module to easily wire up your application.

### Available ProviderSets

| Module              | ProviderSet               | Provides                                     |
|---------------------|---------------------------|----------------------------------------------|
| `config`            | `config.ProviderSet`      | `*Config`, `*Logger`, `*Data`, `*Auth`, etc. |
| `logging/logger`    | `logger.ProviderSet`      | `*Logger` with cleanup                       |
| `data`              | `data.ProviderSet`        | `*Data` with cleanup                         |
| `extension/manager` | `manager.ProviderSet`     | `*Manager` with cleanup                      |
| `security`          | `security.ProviderSet`    | JWT `*TokenManager`                          |
| `messaging`         | `messaging.ProviderSet`   | Email `Sender`                               |
| `concurrency`       | `concurrency.ProviderSet` | Worker `*Pool` with cleanup                  |

### Basic Usage

```go
//go:build wireinject
// +build wireinject

package main

import (
    "github.com/google/wire"
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/logging/logger"
    "github.com/ncobase/ncore/data"
    "github.com/ncobase/ncore/extension/manager"
)

func InitializeApp() (*App, func(), error) {
    panic(wire.Build(
        // Import NCore's core ProviderSets
        config.ProviderSet,
        logger.ProviderSet,
        data.ProviderSet,
        manager.ProviderSet,

        // Your own providers
        NewApp,
    ))
}
```

### With Security and Messaging

```go
//go:build wireinject

package main

import (
    "github.com/google/wire"
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/data"
    "github.com/ncobase/ncore/security"
    "github.com/ncobase/ncore/messaging"
)

func InitializeApp() (*App, func(), error) {
    panic(wire.Build(
        config.ProviderSet,
        data.ProviderSet,
        security.ProviderSet,
        messaging.ProviderSet,
        NewApp,
    ))
}
```

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

## License

See [LICENSE](LICENSE) file for details.
