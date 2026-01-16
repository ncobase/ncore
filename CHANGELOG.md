# Changelog

All notable changes to NCore will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-01-17

### BREAKING CHANGES

**Modular Driver System - Import Path Changes Required**

NCore v0.2.0 introduces a modular driver architecture that significantly reduces binary size and improves dependency management. This requires explicit driver imports in your application.

**Before (v0.1.x):**

```go
import "github.com/ncobase/ncore/data"

// All drivers automatically included
```

**After (v0.2.0):**

```go
import (
    "github.com/ncobase/ncore/data"

    // Explicit driver registration required
    _ "github.com/ncobase/ncore/data/postgres"
    _ "github.com/ncobase/ncore/data/redis"
    _ "github.com/ncobase/ncore/data/elasticsearch"
)
```

**Storage Module Removed from Core**

The storage functionality has been extracted into a standalone `oss` module.

**Before (v0.1.x):**

```go
import "github.com/ncobase/ncore/data/storage"
storage := storage.NewS3Storage(conf)
```

**After (v0.2.0):**

```go
import "github.com/ncobase/ncore/oss"
storage, err := oss.NewStorage(&oss.Config{
    Provider: "s3",
    Bucket:   "mybucket",
    ID:       "access-key",
    Secret:   "secret-key",
})
```

**Search Initialization Simplified**

Search client initialization now uses a factory pattern, eliminating 14+ lines of boilerplate code.

**Before (v0.1.x):**

```go
import (
    esAdapter "github.com/ncobase/ncore/data/elasticsearch"
    esClient "github.com/ncobase/ncore/data/elasticsearch/client"
)

var adapters []search.Adapter
if es := d.GetElasticsearch(); es != nil {
    if c, ok := es.(*esClient.Client); ok {
        adapters = append(adapters, esAdapter.NewAdapter(c))
    }
}
searchClient := search.NewClient(metrics.NoOpCollector{}, adapters...)
```

**After (v0.2.0):**

```go
import "github.com/ncobase/ncore/data"
searchClient := data.NewSearchClient(d)
```

**Messaging Connections Now Optional**

RabbitMQ and Kafka connection failures are no longer fatal errors. Applications can start and run without messaging infrastructure.

**Migration Guide:** See [MIGRATION_v0.1_to_v0.2.md](MIGRATION_v0.1_to_v0.2.md) for complete migration instructions.

### Added

#### Data Layer (Core)

- **Modular Driver System**: Init-based auto-registration pattern (like `database/sql`)
  - Database drivers: `postgres`, `mysql`, `sqlite`, `mongodb`, `neo4j`
  - Cache driver: `redis`
  - Search drivers: `elasticsearch`, `opensearch`, `meilisearch`
  - Message queue drivers: `rabbitmq`, `kafka`
  - Compatibility layer: `data/all` (imports all drivers)

- **Search Factory Pattern**: `data.NewSearchClient()` auto-detects and initializes available search engines
  - Supports multiple search engines simultaneously
  - Automatic adapter registration via `init()`
  - Zero-configuration for common use cases

#### Logging Module

- **Modular Log Hooks**: Log-to-search-engine hooks now use driver registration pattern
  - `logging/hooks/elasticsearch` - Elasticsearch log hook (optional import)
  - `logging/hooks/meilisearch` - Meilisearch log hook (optional import)
  - `logging/hooks/opensearch` - OpenSearch log hook (optional import)
  - Hooks auto-register via `init()` and are initialized based on config
  - Zero dependencies in core logging module when hooks not imported

#### OSS Module (Standalone)

- **Independent Object Storage Module**: Extracted from core `data` package
  - 9 providers: AWS S3, Azure Blob, Aliyun OSS, Tencent COS, Google Cloud Storage, MinIO, Qiniu Kodo, Synology NAS, Local Filesystem
  - Direct use via `oss.NewStorage()` - no driver registration needed
  - See [oss/README.md](oss/README.md) for details

#### Configuration

- **Optional Service Connections**: Messaging services (RabbitMQ/Kafka) now gracefully degrade on connection failure
  - Applications can start without messaging infrastructure
  - Warning logs instead of fatal errors
  - Improved resilience for development environments

### Changed

#### Data Layer

- **RabbitMQ Vhost Encoding**: Fixed URL encoding for vhosts with leading slash
  - Now supports standard vhost format: `/vhost_name`
  - Generates double-slash URL path: `amqp://user:pass@host//vhost`

- **Connection Handling**: RabbitMQ and Kafka connections are now optional
  - Non-fatal errors allow application startup
  - Suitable for microservices that don't require messaging

#### Build System

- **Go Workspace**: Project uses `go.work` for multi-module development
  - 42 sub-modules in workspace
  - Improved local development experience
  - See [MODULES.md](MODULES.md) for architecture details

### Removed

#### Data Layer

- **Storage Package**: Removed `data/storage/*` - migrated to standalone `oss` module
  - Deleted 8 hardcoded storage implementations (S3, MinIO, Aliyun, Azure, GCP, Tencent, Qiniu, Synology)
  - Use `github.com/ncobase/ncore/oss` instead

- **Logging Hooks**: Moved hardcoded search engine logging hooks to modular packages
  - `logging/logger/elasticsearch_hook.go` → `logging/hooks/elasticsearch`
  - `logging/logger/meilisearch_hook.go` → `logging/hooks/meilisearch`
  - `logging/logger/opensearch_hook.go` → `logging/hooks/opensearch`
  - Now requires explicit import: `_ "github.com/ncobase/ncore/logging/hooks/elasticsearch"`

### Fixed

#### Data Layer

- **RabbitMQ Vhost 403 Errors**: Fixed connection failures for vhosts with leading slash
  - Issue: `vhost: /maidesk` generated incorrect URL `amqp://user:pass@host/maidesk`
  - Fix: Now generates `amqp://user:pass@host//maidesk` (double slash)
  - Supports both `/vhost` and `vhost` naming conventions

### Performance

**Binary Size Reduction**

| Metric                  | v0.1.x | v0.2.0 | Improvement |
| ----------------------- | ------ | ------ | ----------- |
| Binary Size (basic app) | ~92MB  | ~43MB  | **-53%**    |
| Dependencies (data mod) | ~25    | 5      | **-80%**    |
| Compilation Time        | ~45s   | ~20s   | **-56%**    |

**Dependency Reduction**

The `data` module now has only 5 direct dependencies (was ~25):

1. `github.com/google/wire` - Dependency injection
2. `github.com/spf13/viper` - Configuration management
3. Internal ncore modules

**Build Time Improvement**

Applications compile 56% faster due to eliminated unused SDK dependencies.

### Documentation

- Updated [MODULES.md](MODULES.md) - Multi-module architecture documentation
- Updated [MODULES_zh-CN.md](MODULES_zh-CN.md) - Chinese version
- Updated [README.md](README.md) - v0.2 quick start and migration guide
- Updated [README_zh-CN.md](README_zh-CN.md) - Chinese version
- Updated [DEVELOPMENT.md](DEVELOPMENT.md) - Development workflow guide

### CLI Tool

#### Templates

- **Data Template**: Updated to use `data.NewSearchClient()` pattern
  - Ent ORM template includes search client helpers
  - GORM template includes search client helpers
  - Automatic search method generation: `IndexDocument`, `DeleteDocument`, `Search`, `GetAvailableSearchEngines`

- **Command Template**: Generates correct v0.2 driver imports
  - Automatic driver import generation based on flags
  - Supports: `--use-postgres`, `--use-redis`, `--use-elastic`, etc.

### Examples

- Updated [01-basic-rest-api](examples/01-basic-rest-api) - Demonstrates v0.2 driver imports
- Updated [02-mongodb-api](examples/02-mongodb-api) - MongoDB with v0.2 architecture
- All 9 examples verified to work with v0.2

### Dependencies

#### Added

- `data/postgres`, `data/mysql`, `data/sqlite`, `data/mongodb`, `data/neo4j` - Database drivers (standalone modules)
- `data/redis` - Redis cache driver (standalone module)
- `data/elasticsearch`, `data/opensearch`, `data/meilisearch` - Search drivers (standalone modules)
- `data/rabbitmq`, `data/kafka` - Message queue drivers (standalone modules)
- `data/all` - Compatibility layer importing all drivers

#### Changed

- `data` module dependencies reduced from ~25 to 5 direct dependencies

#### Removed

- AWS SDK, Aliyun SDK, Azure SDK, GCP SDK, Tencent SDK, Qiniu SDK - Moved to `oss` module
- Elasticsearch, OpenSearch, Meilisearch SDKs from `data` - Now in optional driver modules

### Compatibility

| Platform | Support Status     | Notes                         |
| -------- | ------------------ | ----------------------------- |
| Linux    | ✅ Fully Supported | x86_64, arm64                 |
| macOS    | ✅ Fully Supported | arm64 (Apple Silicon), x86_64 |
| Windows  | ✅ Fully Supported | x86_64                        |
| FreeBSD  | ✅ Supported       | x86_64 (community tested)     |

**Go Version Requirements:**

- Minimum: Go 1.21
- Recommended: Go 1.25.3
- Tested: Go 1.21, 1.22, 1.23, 1.24, 1.25

### Upgrading from v0.1.x

**Quick Migration Steps:**

1. **Add driver imports** to your `main.go`:

   ```go
   import (
       _ "github.com/ncobase/ncore/data/postgres"
       _ "github.com/ncobase/ncore/data/redis"
   )
   ```

2. **Update search initialization** (if using search):

   ```go
   // Replace old adapter code with:
   searchClient := data.NewSearchClient(d)
   ```

3. **Migrate storage** (if using `data/storage`):

   ```go
   import "github.com/ncobase/ncore/oss"
   storage, err := oss.NewStorage(&oss.Config{...})
   ```

4. **Update RabbitMQ vhost config** (if using custom vhost):

   ```yaml
   data:
     rabbitmq:
       vhost: /myapp # Leading slash now supported
   ```

5. **Test and deploy**:

   ```bash
   go mod tidy
   go build
   go test ./...
   ```

**Detailed Migration:** See [MIGRATION_v0.1_to_v0.2.md](MIGRATION_v0.1_to_v0.2.md)

### Applications Updated

- **maidesk**: Fully migrated to v0.2.0 - all 14 modules using new search factory pattern
- **ncore examples**: All 9 examples updated and verified

### Testing

- ✅ All driver modules compile independently
- ✅ Search factory auto-detection verified
- ✅ Multiple search engines work simultaneously (ES + Meilisearch tested)
- ✅ RabbitMQ vhost connection verified
- ✅ Optional messaging degradation tested
- ✅ Binary size reduction confirmed (92MB → 43MB)
- ✅ Build time improvement verified (45s → 20s)

---

## [0.1.24] - 2025-12-20

### Fixed

- Minor bug fixes and improvements
- Documentation updates

## [0.1.23] - 2025-12-15

### Changed

- Internal refactorings
- Dependency updates

## [0.1.22] - 2025-12-10

### Fixed

- Bug fixes in extension system
- Logging improvements

## [0.1.0] - 2025-11-01

### Added

- Initial stable release of NCore
- Core modules: config, data, logging, extension, security, types, utils, validation
- Database support: PostgreSQL, MySQL, MongoDB
- Cache support: Redis
- Search support: Elasticsearch
- Message queue support: RabbitMQ, Kafka
- Extension system for modular applications
- Wire dependency injection support
- Comprehensive documentation and examples

[unreleased]: https://github.com/ncobase/ncore/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/ncobase/ncore/compare/v0.1.24...v0.2.0
[0.1.24]: https://github.com/ncobase/ncore/compare/v0.1.23...v0.1.24
[0.1.23]: https://github.com/ncobase/ncore/compare/v0.1.22...v0.1.23
[0.1.22]: https://github.com/ncobase/ncore/compare/v0.1.0...v0.1.22
[0.1.0]: https://github.com/ncobase/ncore/releases/tag/v0.1.0
