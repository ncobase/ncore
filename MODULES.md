# NCore Multi-Module Architecture

[English](./MODULES.md) | [中文](./MODULES_zh-CN.md)

## Architecture Design

NCore adopts a multi-module architecture design where each sub-package is an independent Go module. The benefits of this
design are:

1. **Reduced Dependencies**: Users only need to import the sub-modules they require, avoiding unnecessary dependencies
2. **Smaller Build Size**: Avoids bundling all NCore dependencies into the application
3. **Independent Version Management**: Each module can be upgraded independently without affecting others
4. **Clearer Module Boundaries**: Enforces clear dependency relationships between modules

## Module List

```text
github.com/ncobase/ncore/
├── concurrency    - Concurrency utilities
├── config         - Configuration management
├── consts         - Constants definitions
├── ctxutil        - Context utilities
│   ├── email          - Optional email sender helpers
│   └── storage        - Optional object storage helpers
├── data           - Data layer (core abstraction, driver registration, optional Space isolation)
│   ├── postgres       - PostgreSQL driver
│   ├── mysql          - MySQL driver
│   ├── sqlite         - SQLite driver
│   ├── mongodb        - MongoDB driver
│   ├── redis          - Redis driver
│   ├── neo4j          - Neo4j driver
│   ├── elasticsearch  - Elasticsearch driver
│   ├── opensearch     - OpenSearch driver
│   ├── meilisearch    - Meilisearch driver
│   ├── kafka          - Kafka driver
│   └── rabbitmq       - RabbitMQ driver
├── ecode          - Error codes
├── extension      - Extension and plugin system
├── logging        - Logging
│   ├── tracing        - OpenTelemetry tracing
│   └── sentry         - Sentry error tracking
├── messaging      - Message queues
├── net            - Network utilities
├── oss            - Object Storage Service base module
│   ├── s3             - AWS S3 compatible provider
│   ├── azure          - Azure Blob provider
│   ├── aliyun         - Aliyun OSS provider
│   ├── tencent        - Tencent COS provider
│   ├── gcs            - Google Cloud Storage provider
│   ├── minio          - MinIO provider
│   ├── qiniu          - Qiniu Kodo provider
│   └── synology       - Synology NAS provider
├── security       - Security features
├── types          - Common types
├── utils          - Utility functions
├── validation     - Data validation
└── version        - Version management
```

## Modular Driver System (v0.2.0+)

Starting from v0.2.0, NCore implements a **modular driver system** where database, cache, search, messaging, and storage
drivers are separate, opt-in modules. This design significantly reduces binary size and dependencies.

### Available Drivers

#### Database Drivers

- `github.com/ncobase/ncore/data/postgres` - PostgreSQL using pgx/v5
- `github.com/ncobase/ncore/data/mysql` - MySQL
- `github.com/ncobase/ncore/data/sqlite` - SQLite
- `github.com/ncobase/ncore/data/mongodb` - MongoDB
- `github.com/ncobase/ncore/data/neo4j` - Neo4j graph database

#### Cache Driver

- `github.com/ncobase/ncore/data/redis` - Redis cache

#### Search Drivers

- `github.com/ncobase/ncore/data/elasticsearch` - Elasticsearch
- `github.com/ncobase/ncore/data/opensearch` - OpenSearch
- `github.com/ncobase/ncore/data/meilisearch` - Meilisearch

#### Message Queue Drivers

- `github.com/ncobase/ncore/data/kafka` - Apache Kafka
- `github.com/ncobase/ncore/data/rabbitmq` - RabbitMQ

### Object Storage Service (OSS Module)

Object storage uses the base module `github.com/ncobase/ncore/oss` plus explicit provider modules:

- **Standalone Base Module**: Shared interface, configuration, registry, validation, and local filesystem storage.
- **Opt-In Cloud Providers**: Cloud SDK dependencies live in provider submodules and register through blank imports.
- **Provider Modules**:
  - `github.com/ncobase/ncore/oss/s3` for AWS S3, Cloudflare R2, and Backblaze B2 compatible APIs
  - `github.com/ncobase/ncore/oss/azure` for Azure Blob Storage
  - `github.com/ncobase/ncore/oss/aliyun` for Aliyun OSS
  - `github.com/ncobase/ncore/oss/tencent` for Tencent Cloud COS
  - `github.com/ncobase/ncore/oss/gcs` for Google Cloud Storage
  - `github.com/ncobase/ncore/oss/minio` for MinIO
  - `github.com/ncobase/ncore/oss/qiniu` for Qiniu Kodo
  - `github.com/ncobase/ncore/oss/synology` for Synology NAS

**Usage:**

```go
import (
    "github.com/ncobase/ncore/oss"

    _ "github.com/ncobase/ncore/oss/minio"
)

storage, err := oss.NewStorage(&oss.Config{
    Provider: "minio",
    Endpoint: "http://localhost:9000",
    Bucket:   "mybucket",
    ID:       "minioadmin",
    Secret:   "minioadmin",
})
```

See [oss/README.md](oss/README.md) for detailed documentation.

### Optional Space Isolation in the Data Module

The `data` module includes an opt-in `SpaceRouter` abstraction for multi-tenant applications that
need database-level isolation. It provides context helpers, `WithSpaceTx`, `WithSpaceTxRead`,
identifier validation, and `ExecInTx`; it does not include concrete PostgreSQL or MySQL routing
implementations.

Applications provide the router and own the database strategy:

- PostgreSQL applications can implement schema-per-space with transaction-scoped `SET LOCAL
  search_path`.
- MySQL applications can implement database-per-space with separate connection pools.
- Applications also own Space lifecycle provisioning, migrations, table classification, access
  policy, audit, backup, and recovery.

See [data/README.md](data/README.md#space-isolation) and
[MULTITENANCY_DESIGN.md](MULTITENANCY_DESIGN.md) for the full contract.

### How It Works

Drivers follow the `database/sql` pattern:

1. Each driver is a separate Go module with its own dependencies
2. Drivers self-register via `init()` when imported
3. Users explicitly import only the drivers they need using blank imports

```go
import (
    "github.com/ncobase/ncore/data"

    // Import only the drivers you need
    _ "github.com/ncobase/ncore/data/postgres"
    _ "github.com/ncobase/ncore/data/redis"
)
```

### Benefits

| Metric                   | Before v0.2.0 | After v0.2.0 | Improvement |
| ------------------------ | ------------- | ------------ | ----------- |
| Binary Size (basic app)  | ~92MB         | ~43MB        | **-53%**    |
| Dependencies (basic app) | 466           | ~100         | **-78%**    |
| Compilation Time         | ~45s          | ~20s         | **-56%**    |

### Migration Guide

See [README.md](README.md) for migration instructions from v0.1.x.

## Usage

### Using in Applications

Install only the modules you need:

```bash
go get github.com/ncobase/ncore/config@latest
go get github.com/ncobase/ncore/data@latest
go get github.com/ncobase/ncore/logging@latest
go get github.com/ncobase/ncore/data/postgres@latest
go get github.com/ncobase/ncore/data/redis@latest
```

In your code, use blank imports to register the drivers:

```go
package main

import (
    "github.com/ncobase/ncore/data"

    // Blank import to register drivers
    _ "github.com/ncobase/ncore/data/postgres"
    _ "github.com/ncobase/ncore/data/redis"
)

func main() {
    // Use data layer normally - drivers are auto-registered
    d, cleanup, _ := data.New(cfg.Data)
    defer cleanup()
}
```

### Local Development

#### 1. Using go.work (Recommended)

The project root provides a `go.work` file for convenient local development and testing:

```bash
# In ncore root directory
go work sync  # Sync all module dependencies
bash scripts/test.sh # Test all modules
```

#### 2. Using local ncore in applications

Use the `replace` directive in your application's `go.mod`:

```go
replace (
    github.com/ncobase/ncore/data => /path/to/ncore/data
    github.com/ncobase/ncore/config => /path/to/ncore/config
    // Replace the modules you need to debug locally
)
```

## Release Process

### Release Single Module

```bash
cd data
git tag data/v0.1.0
git push origin data/v0.1.0
```

### Batch Release

```bash
# Tag all modules with the same version
./scripts/tag.sh v0.1.0
```

## Module Dependency Principles

1. **Minimal Dependencies**: Each module should only import necessary dependencies
2. **Avoid Circular Dependencies**: Modules cannot have circular dependencies
3. **Common Modules First**: `types`, `consts`, `ecode` and other common modules should have zero or minimal
   dependencies
4. **Large Dependency Isolation**: Modules like `data` containing large dependencies (databases, search engines) should
   be isolated

## FAQ

### Q: Why is there no go.mod in the root directory?

A: Because each sub-package is an independent module, the root directory doesn't need go.mod. The go.work file is
sufficient for managing local development.

### Q: Should go.work be committed to git?

A: It can be committed. go.work facilitates local development for team members, but `go.work.sum` should not be
committed (already in .gitignore).

### Q: How to add a new module?

A:

1. Create a new directory
2. Run `go mod init github.com/ncobase/ncore/new-module-name` in the directory
3. Add `./new-module-name` to the root directory's `go.work`

### Q: How do modules reference each other?

A: Use the full module path directly:

```go
import (
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/logging"
)
```

During local development, go.work will automatically resolve these references.

## Dependency Injection Support (Google Wire)

NCore modules provide native [Google Wire](https://github.com/google/wire) support, with each module exposing a
`ProviderSet` for dependency injection.

### Supported Modules

| Module              | ProviderSet               | Provides                         | Cleanup |
| ------------------- | ------------------------- | -------------------------------- | ------- |
| `config`            | `config.ProviderSet`      | `*Config` and sub-configurations | No      |
| `logging/logger`    | `logger.ProviderSet`      | `*Logger`                        | Yes     |
| `data`              | `data.ProviderSet`        | `*Data`                          | Yes     |
| `extension/manager` | `manager.ProviderSet`     | `*Manager`                       | Yes     |
| `security`          | `security.ProviderSet`    | JWT `*TokenManager`              | No      |
| `messaging`         | `messaging.ProviderSet`   | Email `Sender`                   | No      |
| `concurrency`       | `concurrency.ProviderSet` | Worker `*Pool`                   | Yes     |

### Wire Usage Example

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

// InitializeApp initializes the application
func InitializeApp() (*App, func(), error) {
    panic(wire.Build(
        // Configuration management
        config.ProviderSet,

        // Core components
        logger.ProviderSet,
        data.ProviderSet,
        manager.ProviderSet,

        // Application constructor
        NewApp,
    ))
}
```

### Features

1. **Cleanup Function Support**: Providers for `data`, `logger`, `manager`, and `concurrency` modules return cleanup
   functions
2. **Configuration Extraction**: `config.ProviderSet` automatically extracts sub-configurations required by other
   modules
3. **Interface Binding**: `security` module uses `wire.Bind` for interface binding
4. **Error Handling**: All providers properly handle and propagate errors

For detailed documentation, see:

- [README](README.md#dependency-injection-google-wire)
- [DEVELOPMENT.md](DEVELOPMENT.md#6-dependency-injection-google-wire)
- [Example Code](examples/09-wire)

## Related Resources

- [Go Modules Official Documentation](https://go.dev/doc/modules/managing-dependencies)
- [Go Workspaces Official Documentation](https://go.dev/doc/tutorial/workspaces)
- [Multi-module repositories Best Practices](https://github.com/golang/go/wiki/Modules#faqs--multi-module-repositories)
- [Google Wire Official Documentation](https://github.com/google/wire)
