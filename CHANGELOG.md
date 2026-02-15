# Changelog

All notable changes to NCore will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- **MongoDB Driver**: Upgraded to v2.5.0
  - Updated import paths to `go.mongodb.org/mongo-driver/v2`
  - Improved context handling following v2 best practices
  - **No user code changes required** - backward compatibility maintained

### Removed

- **data/all package**: Removed bulk driver import package
  - The `data/all` compatibility layer has been removed to encourage best practices
  - Applications should explicitly import only the drivers they need
  - This reduces binary size and prevents unused dependencies from being included
  - **Migration**: Replace `import _ "github.com/ncobase/ncore/data/all"` with explicit driver imports:

    ```go
    import (
        _ "github.com/ncobase/ncore/data/postgres"
        _ "github.com/ncobase/ncore/data/redis"
        // Add only what you need
    )
    ```

## [0.2.3] - 2026-01-18

### Added

#### OSS Module

- **Interface Methods**: Added `Exists` and `Stat` methods to the `Interface`
  - `Exists(path string) (bool, error)` - Check if an object exists without downloading
  - `Stat(path string) (*Object, error)` - Retrieve object metadata without downloading content
  - Implemented in all 9 storage adapters with proper error handling

- **Driver Auto-Registration**: Added missing driver registration for 4 adapters
  - Azure Blob Storage (`azure`)
  - Google Cloud Storage (`gcs`)
  - Qiniu Kodo (`qiniu`)
  - Synology NAS (`synology`)
  - All drivers now auto-register via `init()` functions

### Changed

#### OSS Module

- **Documentation**: Updated README.md and README_zh-CN.md to formal release format
  - Accurate Interface definition with all 9 methods
  - Complete provider configuration examples
  - Advanced usage examples for Exists, Stat, and stream operations

- **Code Quality**: Added comprehensive English comments to all adapter implementations
  - Struct documentation for all adapters
  - Method documentation following Go conventions
  - Consistent error handling patterns across all adapters

### Fixed

#### OSS Module

- **S3 Adapter**: Added proper error type imports for `NotFound` and `NoSuchKey` handling
- **Azure Adapter**: Added `bloberror` import for proper blob not found detection
- **Qiniu Adapter**: Added `strings` import and `isQiniuNotFound` helper for error detection

## [0.2.2] - 2026-01-18

### Fixed

- **Module Structure**: Removed `data/search` as an independent module
  - `data/search` is now correctly treated as a subpackage of the `data` module
  - Fixes "ambiguous import" errors when using `github.com/ncobase/ncore/data/search`
  - This was a packaging error in v0.2.0 where `data/search/go.mod` was incorrectly included
  - Applications should import `github.com/ncobase/ncore/data` (which includes the search subpackage), not `github.com/ncobase/ncore/data/search` as a separate module

**Migration from v0.2.0:**

If you explicitly added `github.com/ncobase/ncore/data/search v0.2.0` to your `go.mod`, remove it:

```bash
go mod edit -droprequire=github.com/ncobase/ncore/data/search
go get github.com/ncobase/ncore/data@v0.2.1
go mod tidy
```

The import statement in your code remains unchanged:

import "github.com/ncobase/ncore/data/search" // Correct - this is a subpackage

````

## [0.2.0] - 2026-01-17

### BREAKING CHANGES

**Modular Driver System**: Drivers must be explicitly imported:
```go
import (
    _ "github.com/ncobase/ncore/data/postgres"
    _ "github.com/ncobase/ncore/data/redis"
)
```

**Storage Module**: Moved to standalone `oss` module:
```go
import "github.com/ncobase/ncore/oss"
storage, _ := oss.NewStorage(&oss.Config{Provider: "s3", ...})
```

**Search Initialization**: Simplified to factory pattern:
```go
searchClient := data.NewSearchClient(d)
```

**Messaging**: Connection failures no longer fatal

See [MIGRATION_v0.1_to_v0.2.md](MIGRATION_v0.1_to_v0.2.md) for migration instructions.

### Added

- **Modular Driver System**: Init-based auto-registration (like `database/sql`)
  - Drivers: `postgres`, `mysql`, `sqlite`, `mongodb`, `neo4j`, `redis`
  - Search: `elasticsearch`, `opensearch`, `meilisearch`
  - Messaging: `rabbitmq`, `kafka`

- **Search Factory Pattern**: `data.NewSearchClient()` auto-detects available search engines

- **Modular Log Hooks**: Optional logging hooks for search engines
  - `logging/hooks/elasticsearch`, `logging/hooks/meilisearch`, `logging/hooks/opensearch`

- **OSS Module (Standalone)**: Independent object storage module with 9 providers
  - AWS S3, Azure Blob, Aliyun OSS, Tencent COS, GCS, MinIO, Qiniu, Synology, Local

- **Optional Service Connections**: Messaging (RabbitMQ/Kafka) failures now graceful

### Changed

- **RabbitMQ Vhost**: Fixed URL encoding for vhosts with leading slash (`/vhost`)
- **Go Workspace**: Multi-module development with `go.work` (42 sub-modules)

### Removed

- **Storage Package**: Moved `data/storage/*` to standalone `oss` module
- **Logging Hooks**: Moved to modular packages (`logging/hooks/*`)

### Fixed

- **RabbitMQ Vhost 403**: Fixed connection failures for `/vhost` format

### Performance

| Metric | v0.1.x | v0.2.0 | Improvement |
|--------|--------|--------|-------------|
| Binary Size | ~92MB | ~43MB | **-53%** |
| Dependencies | ~25 | 5 | **-80%** |
| Build Time | ~45s | ~20s | **-56%** |

### Documentation

Updated: MODULES.md, README.md, DEVELOPMENT.md (English & Chinese versions)

### Dependencies

- **Added**: Standalone driver modules (`data/postgres`, `data/redis`, etc.)
- **Removed**: Cloud storage SDKs moved to `oss` module

**Requirements:** Go 1.21+, tested on Linux, macOS, Windows, FreeBSD

**Migration:** See [MIGRATION_v0.1_to_v0.2.md](MIGRATION_v0.1_to_v0.2.md)

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

[0.2.3]: https://github.com/ncobase/ncore/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/ncobase/ncore/compare/v0.2.0...v0.2.2
[0.1.24]: https://github.com/ncobase/ncore/compare/v0.1.23...v0.1.24
[0.1.23]: https://github.com/ncobase/ncore/compare/v0.1.22...v0.1.23
[0.1.22]: https://github.com/ncobase/ncore/compare/v0.1.0...v0.1.22
[0.1.0]: https://github.com/ncobase/ncore/releases/tag/v0.1.0
