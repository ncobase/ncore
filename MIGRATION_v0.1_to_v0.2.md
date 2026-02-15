# Migration Guide: NCore v0.1.x to v0.2.0

This guide provides instructions for migrating applications from NCore v0.1.x to v0.2.0.

## Overview

NCore v0.2.0 introduces a modular driver architecture that significantly reduces binary size and dependency complexity.

**Key Changes:**

- **Modular Drivers**: Explicit driver imports required (like `database/sql`)
- **Search Factory Pattern**: Simplified search client initialization
- **Storage Extraction**: Object storage moved to standalone `oss` module
- **Optional Messaging**: RabbitMQ/Kafka failures no longer fatal

**Impact:**

| Metric                  | v0.1.x | v0.2.0 | Improvement |
| ----------------------- | ------ | ------ | ----------- |
| Binary Size (basic app) | ~92MB  | ~43MB  | **-53%**    |
| Dependencies            | ~466   | ~100   | **-78%**    |
| Build Time              | ~45s   | ~20s   | **-56%**    |

**Estimated Migration Time:** 30 minutes to 2 hours for most projects

---

## Quick Migration

### Step 1: Add Driver Imports

Add explicit driver imports to your `main.go`:

```go
import (
    "github.com/ncobase/ncore/data"

    // Add only the drivers you need
    _ "github.com/ncobase/ncore/data/postgres"
    _ "github.com/ncobase/ncore/data/redis"
    _ "github.com/ncobase/ncore/data/elasticsearch"
)
```

### Step 2: Update Search Initialization

Replace search adapter boilerplate with factory pattern:

```go
// Old (v0.1.x) - REMOVE
// var adapters []search.Adapter
// if es := d.GetElasticsearch(); es != nil {
//     adapters = append(adapters, esAdapter.NewAdapter(es))
// }
// searchClient := search.NewClient(metrics.NoOpCollector{}, adapters...)

// New (v0.2.0) - USE THIS
searchClient := data.NewSearchClient(d)
```

### Step 3: Migrate Storage

Replace `data/storage` with `oss` module:

```go
// Old
// import "github.com/ncobase/ncore/data/storage"

// New
import "github.com/ncobase/ncore/oss"

storage, err := oss.NewStorage(&oss.Config{
    Provider: "s3",
    Endpoint: "s3.amazonaws.com",
    Bucket:   "my-bucket",
    ID:       "access-key",
    Secret:   "secret-key",
})
```

### Step 4: Update Dependencies

```bash
go get github.com/ncobase/ncore/data@v0.2.0
go get github.com/ncobase/ncore/config@v0.2.0
go mod tidy
go build
go test ./...
```

---

## Migration Checklist

### Pre-Migration

- [ ] Backup: `git checkout -b migrate-to-v0.2`
- [ ] Tag: `git tag backup/pre-v0.2-migration`
- [ ] Baseline: Run all tests
- [ ] Audit: Identify which drivers you're using

### Code Changes

- [ ] Add driver imports to `main.go`
- [ ] Update search initialization (if using search)
- [ ] Migrate storage code (if using `data/storage`)
- [ ] Update RabbitMQ vhost config (if using custom vhost with leading slash)

### Testing

- [ ] Compile: `go build`
- [ ] Unit Tests: `go test ./...`
- [ ] Integration tests
- [ ] Manual testing

---

## Breaking Changes

### 1. Driver Import Requirements

**What Changed:** Drivers must be explicitly imported.

**Migration:**

```go
// Add to main.go or init file
import (
    // Database (choose one)
    _ "github.com/ncobase/ncore/data/postgres"
    // _ "github.com/ncobase/ncore/data/mysql"
    // _ "github.com/ncobase/ncore/data/mongodb"

    // Cache
    _ "github.com/ncobase/ncore/data/redis"

    // Search (optional)
    _ "github.com/ncobase/ncore/data/elasticsearch"
    // _ "github.com/ncobase/ncore/data/meilisearch"

    // Messaging (optional)
    _ "github.com/ncobase/ncore/data/rabbitmq"
    // _ "github.com/ncobase/ncore/data/kafka"
)
```

**Identify Required Drivers:**

```bash
grep -r "d.Conn.DB()" .          # Database
grep -r "d.Conn.RC" .            # Redis
grep -r "GetElasticsearch" .     # Elasticsearch
grep -r "GetRabbitMQ" .          # RabbitMQ
```

### 2. Storage Module Migration

**What Changed:** Storage moved from `data/storage` to standalone `oss` module.

**Migration:**

1. **Update imports:**

   ```bash
   # Find usage
   grep -r "data/storage" . --include="*.go"

   # Replace import
   # macOS: sed -i '' 's|data/storage|oss|g' file.go
   # Linux: sed -i 's|data/storage|oss|g' file.go
   ```

2. **Update constructor calls:**

   ```go
   // Old
   storage.NewS3Storage(&storage.S3Config{
       AccessKey: "key",
       SecretKey: "secret",
   })

   // New
   oss.NewStorage(&oss.Config{
       Provider: "s3",
       ID:       "key",    // Renamed from AccessKey
       Secret:   "secret", // Renamed from SecretKey
   })
   ```

**Supported Providers:** `s3`, `minio`, `aliyun`, `tencent`, `gcs`, `azure`, `qiniu`, `synology`, `local`

### 3. Search Initialization

**What Changed:** Simplified from 14+ lines to 1 line using factory pattern.

**Migration:**

```go
// Delete old adapter boilerplate
// var adapters []search.Adapter
// if es := d.GetElasticsearch(); es != nil { ... }
// searchClient := search.NewClient(metrics.NoOpCollector{}, adapters...)

// Replace with factory
searchClient := data.NewSearchClient(d)

// With custom metrics collector
searchClient := data.NewSearchClient(d, myCollector)
```

**Add search driver import to main.go:**

```go
import _ "github.com/ncobase/ncore/data/elasticsearch"
```

### 4. Log Hooks (if using)

**What Changed:** Log hooks are now modular.

**Migration:**

Add hook imports to `main.go`:

```go
import (
    _ "github.com/ncobase/ncore/logging/hooks/elasticsearch"
    // _ "github.com/ncobase/ncore/logging/hooks/meilisearch"
    // _ "github.com/ncobase/ncore/logging/hooks/opensearch"
)
```

Config remains unchanged.

---

## Complete Example

### Before (v0.1.x)

```go
package main

import (
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/data"
    "github.com/ncobase/ncore/data/storage"
)

func main() {
    conf, _ := config.LoadConfig("config.yaml")
    d, cleanup, _ := data.New(conf.Data)
    defer cleanup()

    s3 := storage.NewS3Storage(&storage.S3Config{
        Endpoint:  conf.S3.Endpoint,
        Bucket:    conf.S3.Bucket,
        AccessKey: conf.S3.AccessKey,
        SecretKey: conf.S3.SecretKey,
    })
}
```

### After (v0.2.0)

```go
package main

import (
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/data"
    "github.com/ncobase/ncore/oss"

    // Driver imports
    _ "github.com/ncobase/ncore/data/postgres"
    _ "github.com/ncobase/ncore/data/redis"
)

func main() {
    conf, _ := config.LoadConfig("config.yaml")
    d, cleanup, _ := data.New(conf.Data)
    defer cleanup()

    s3, err := oss.NewStorage(&oss.Config{
        Provider: "s3",
        Endpoint: conf.S3.Endpoint,
        Bucket:   conf.S3.Bucket,
        ID:       conf.S3.AccessKey,
        Secret:   conf.S3.SecretKey,
    })
    if err != nil {
        panic(err)
    }
}
```

**Changes:**

1. ✅ Added driver imports
2. ✅ Changed `data/storage` → `oss`
3. ✅ Updated config field names (`AccessKey` → `ID`, `SecretKey` → `Secret`)
4. ✅ Added error handling

---

## Troubleshooting

### "driver not registered" panic

**Cause:** Missing driver import.

**Solution:** Add to `main.go`:

```go
import _ "github.com/ncobase/ncore/data/postgres"
```

### Import errors: "cannot find package data/storage"

**Cause:** Package moved to `oss` module.

**Solution:**

```bash
# Find usage
grep -r "data/storage" . --include="*.go"

# Replace import
# macOS: find . -name "*.go" -exec sed -i '' 's|data/storage|oss|g' {} \;
# Linux: find . -name "*.go" -exec sed -i 's|data/storage|oss|g' {} \;
```

### Search returns no results

**Cause:** Search driver not imported.

**Solution:**

```go
import _ "github.com/ncobase/ncore/data/elasticsearch"
```

Verify:

```go
engines := searchClient.GetAvailableSearchEngines()
log.Printf("Available: %v", engines) // Should show ["elasticsearch"]
```

### RabbitMQ 403 "access to vhost" error

**Solution:** Ensure vhost has leading slash in config:

```yaml
data:
  rabbitmq:
    vhost: "/myapp" # ← Must have leading slash
```

### Binary size not reduced

**Causes:**

1. Importing unused drivers
2. Not running `go mod tidy`
3. Debug build

**Solution:**

```bash
# Remove unused drivers from imports
# Clean and rebuild
rm go.sum
go mod tidy
go build -ldflags="-s -w" -o app
```

### Mixed version conflicts

**Cause:** Mixing v0.1 and v0.2 modules.

**Solution:** Ensure all ncore modules are v0.2.0:

```bash
go get github.com/ncobase/ncore/config@v0.2.0
go get github.com/ncobase/ncore/data@v0.2.0
go get github.com/ncobase/ncore/logging@v0.2.0
go mod tidy
```

---

## FAQ

**Q: Can I mix v0.1 and v0.2?**
A: No. All ncore modules must be the same version.

**Q: Do I need to update all modules at once?**
A: Yes. Use:

```bash
go get github.com/ncobase/ncore/{config,data,logging,security,extension}@v0.2.0
go mod tidy
```

**Q: What if I don't know which drivers I'm using?**
A: Search your codebase:

```bash
grep -r "d.Conn.DB()" .      # Database
grep -r "d.Conn.RC" .        # Redis
grep -r "GetElasticsearch" . # Search
```

**Q: Will v0.1.x continue to be supported?**
A: Security updates for 6 months. No new features.

**Q: Performance difference?**
A: Build time: 56% faster, Binary: 53% smaller, Dependencies: 78% fewer

**Q: Can I migrate gradually in microservices?**
A: Yes. Migrate one service at a time. Each service must use either v0.1 or v0.2 completely.

---

## Performance Benefits

| Metric              | v0.1.24 | v0.2.0  | Improvement |
| ------------------- | ------- | ------- | ----------- |
| Binary Size         | 92.3 MB | 43.1 MB | **-53.3%**  |
| Dependencies        | 466     | 102     | **-78.1%**  |
| Build time (cold)   | 45.2s   | 19.8s   | **-56.2%**  |
| Build time (cached) | 8.1s    | 3.4s    | **-58.0%**  |
| Docker image        | 145 MB  | 68 MB   | **-53.1%**  |
| Startup time        | 1.2s    | 1.1s    | **-8.3%**   |

---

## Resources

- [CHANGELOG.md](CHANGELOG.md) - Complete list of changes
- [MODULES.md](MODULES.md) - Multi-module architecture
- [README.md](README.md) - Quick start guide
- [OSS Module](oss/README.md) - Object storage documentation

---

**Last Updated:** 2026-01-17
**For questions, please [open an issue](https://github.com/ncobase/ncore/issues/new).**
