# Migration Guide: NCore v0.1.x to v0.2.0

This guide helps you migrate from NCore v0.1.x to v0.2.0.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Migration Checklist](#migration-checklist)
- [Breaking Changes](#breaking-changes)
  - [Driver Import Requirements](#driver-import-requirements)
  - [Storage Module Migration](#storage-module-migration)
  - [Search Initialization](#search-initialization)
  - [Optional Messaging Connections](#optional-messaging-connections)
- [Step-by-Step Migration](#step-by-step-migration)
- [Troubleshooting](#troubleshooting)
- [FAQ](#faq)
- [Performance Benefits](#performance-benefits)

---

## Overview

NCore v0.2.0 introduces a **modular driver architecture** that significantly reduces binary size and dependency complexity. The main changes are:

- **Modular Drivers**: Explicit driver imports required (like `database/sql`)
- **Search Factory Pattern**: Simplified search client initialization
- **Storage Extraction**: Object storage moved to standalone `oss` module
- **Optional Messaging**: RabbitMQ/Kafka failures no longer fatal

**Impact Summary:**

| Metric | v0.1.x | v0.2.0 | Improvement |
|--------|--------|--------|-------------|
| Binary Size (basic app) | ~92MB | ~43MB | **-53%** |
| Dependencies | ~466 | ~100 | **-78%** |
| Build Time | ~45s | ~20s | **-56%** |

**Estimated Migration Time:**
- Simple projects: 30 minutes
- Medium projects: 2-4 hours
- Complex projects: 1-2 days

---

## Quick Start

If you want to migrate quickly, follow these 5 steps:

### Step 1: Add Driver Imports

Add explicit driver imports to your `main.go` or initialization file:

```go
import (
    "github.com/ncobase/ncore/data"
    
    // Add only the drivers you need
    _ "github.com/ncobase/ncore/data/postgres"
    _ "github.com/ncobase/ncore/data/redis"
    _ "github.com/ncobase/ncore/data/elasticsearch"
    // Add more as needed
)
```

### Step 2: Update Search Initialization (if using search)

Replace search adapter boilerplate with factory pattern:

```go
// Old (v0.1.x) - REMOVE THIS
// var adapters []search.Adapter
// if es := d.GetElasticsearch(); es != nil {
//     adapters = append(adapters, esAdapter.NewAdapter(es))
// }
// searchClient := search.NewClient(metrics.NoOpCollector{}, adapters...)

// New (v0.2.0) - USE THIS
import "github.com/ncobase/ncore/data"
searchClient := data.NewSearchClient(d)
```

### Step 3: Migrate Storage (if using `data/storage`)

Replace storage imports:

```go
// Old (v0.1.x)
// import "github.com/ncobase/ncore/data/storage"

// New (v0.2.0)
import "github.com/ncobase/ncore/oss"

storage, err := oss.NewStorage(&oss.Config{
    Provider: "s3", // or "minio", "aliyun", "local", etc.
    Endpoint: "s3.amazonaws.com",
    Bucket:   "my-bucket",
    ID:       "access-key-id",
    Secret:   "secret-access-key",
})
```

### Step 4: Update RabbitMQ Config (if using custom vhost)

If you use RabbitMQ with a custom vhost, ensure it has a leading slash:

```yaml
# config.yaml
data:
  rabbitmq:
    url: "localhost:5672"
    username: "user"
    password: "pass"
    vhost: "/myapp"  # Ensure leading slash
```

### Step 5: Build and Test

```bash
# Update dependencies
go get github.com/ncobase/ncore/data@v0.2.0
go get github.com/ncobase/ncore/config@v0.2.0
go mod tidy

# Build
go build

# Run tests
go test ./...
```

---

## Migration Checklist

Use this checklist to ensure a complete migration:

### Pre-Migration

- [ ] **Backup**: Create git branch `git checkout -b migrate-to-v0.2`
- [ ] **Tag**: Create backup tag `git tag backup/pre-v0.2-migration`
- [ ] **Baseline**: Run all tests to establish baseline `go test ./...`
- [ ] **Audit**: Identify which drivers you're using (postgres, redis, elasticsearch, etc.)
- [ ] **Review**: Read [Breaking Changes](#breaking-changes) section

### Code Changes

- [ ] **Drivers**: Add explicit driver imports to `main.go`
  - [ ] Database driver (postgres, mysql, mongodb, etc.)
  - [ ] Cache driver (redis)
  - [ ] Search driver (elasticsearch, opensearch, meilisearch)
  - [ ] Message queue driver (rabbitmq, kafka)
  
- [ ] **Search**: Update search initialization (if using search)
  - [ ] Import `github.com/ncobase/ncore/data/search`
  - [ ] Replace adapter boilerplate with `data.NewSearchClient(d)`
  - [ ] Remove old adapter imports
  
- [ ] **Storage**: Migrate storage code (if using `data/storage`)
  - [ ] Replace `import "github.com/ncobase/ncore/data/storage"`
  - [ ] Update to `import "github.com/ncobase/ncore/oss"`
  - [ ] Update constructor calls to `oss.NewStorage()`
  
- [ ] **Config**: Update RabbitMQ vhost config (if using custom vhost)
  - [ ] Ensure vhost has leading slash: `/vhost_name`

### Dependencies

- [ ] **Update go.mod**: Get v0.2.0 versions
  ```bash
  go get github.com/ncobase/ncore/config@v0.2.0
  go get github.com/ncobase/ncore/data@v0.2.0
  go get github.com/ncobase/ncore/logging@v0.2.0
  # ... update all ncore modules you use
  ```
  
- [ ] **Tidy**: Run `go mod tidy`
- [ ] **Verify**: Check `go.sum` for v0.2.0 versions

### Testing

- [ ] **Compile**: `go build` - fix all compilation errors
- [ ] **Unit Tests**: `go test ./...` - ensure all tests pass
- [ ] **Integration Tests**: Run integration test suite
- [ ] **Manual Testing**: Test critical application paths
- [ ] **Check Logs**: Verify no unexpected warnings/errors

### Deployment

- [ ] **Staging**: Deploy to staging environment
- [ ] **Monitor**: Watch metrics for 24 hours
  - [ ] Binary size reduced?
  - [ ] Performance improved?
  - [ ] No errors in logs?
- [ ] **Production**: Deploy to production
- [ ] **Verify**: Confirm application health

### Post-Migration

- [ ] **Cleanup**: Remove old commented code
- [ ] **Documentation**: Update project documentation
- [ ] **Team**: Notify team of migration completion
- [ ] **Delete Branch**: Merge migration branch `git merge migrate-to-v0.2`

---

## Breaking Changes

### Driver Import Requirements

**What Changed:** Drivers are no longer automatically included. You must explicitly import them.

**Why:** This allows applications to include only the drivers they need, dramatically reducing binary size and dependencies.

#### Before (v0.1.x)

```go
package main

import (
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/data"
)

func main() {
    conf, _ := config.LoadConfig("config.yaml")
    d, cleanup, _ := data.New(conf.Data)
    defer cleanup()
    
    // All drivers automatically available
    db := d.Conn.DB()        // Works
    redis := d.Conn.RC       // Works
    es := d.GetElasticsearch() // Works
}
```

#### After (v0.2.0)

```go
package main

import (
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/data"
    
    // Explicit driver registration
    _ "github.com/ncobase/ncore/data/postgres"
    _ "github.com/ncobase/ncore/data/redis"
    _ "github.com/ncobase/ncore/data/elasticsearch"
)

func main() {
    conf, _ := config.LoadConfig("config.yaml")
    d, cleanup, _ := data.New(conf.Data)
    defer cleanup()
    
    // Now drivers are available
    db := d.Conn.DB()        // Works - postgres driver registered
    redis := d.Conn.RC       // Works - redis driver registered
    es := d.GetElasticsearch() // Works - elasticsearch driver registered
}
```

#### Migration Steps

1. **Identify Required Drivers:**

```bash
# Search your codebase for data usage
grep -r "d.Conn.DB()" .          # Database
grep -r "d.Conn.RC" .            # Redis
grep -r "GetElasticsearch" .     # Elasticsearch
grep -r "GetRabbitMQ" .          # RabbitMQ
```

2. **Add Imports to main.go:**

```go
import (
    // Database (choose one)
    _ "github.com/ncobase/ncore/data/postgres"
    // _ "github.com/ncobase/ncore/data/mysql"
    // _ "github.com/ncobase/ncore/data/mongodb"
    // _ "github.com/ncobase/ncore/data/sqlite"
    
    // Cache (if using Redis)
    _ "github.com/ncobase/ncore/data/redis"
    
    // Search (if using search)
    _ "github.com/ncobase/ncore/data/elasticsearch"
    // _ "github.com/ncobase/ncore/data/opensearch"
    // _ "github.com/ncobase/ncore/data/meilisearch"
    
    // Messaging (if using message queues)
    _ "github.com/ncobase/ncore/data/rabbitmq"
    // _ "github.com/ncobase/ncore/data/kafka"
)
```

3. **Compile and Fix Errors:**

```bash
go build
# If you see "driver not registered" errors, add the missing driver import
```

#### Compatibility Layer

If you need all drivers (e.g., for testing), use the compatibility layer:

```go
import _ "github.com/ncobase/ncore/data/all"
```

**Warning:** This defeats the purpose of v0.2.0 optimization and should only be used temporarily.

---

### Storage Module Migration

**What Changed:** Storage functionality moved from `data/storage` to standalone `oss` module.

**Why:** Object storage is a specialized concern with heavy dependencies. Extracting it allows most applications to avoid these dependencies.

#### Before (v0.1.x)

```go
import "github.com/ncobase/ncore/data/storage"

// S3 Storage
s3Storage := storage.NewS3Storage(&storage.S3Config{
    Endpoint:  "s3.amazonaws.com",
    Bucket:    "my-bucket",
    AccessKey: "access-key",
    SecretKey: "secret-key",
})

// MinIO Storage
minioStorage := storage.NewMinIOStorage(&storage.MinIOConfig{
    Endpoint:  "localhost:9000",
    Bucket:    "my-bucket",
    AccessKey: "minioadmin",
    SecretKey: "minioadmin",
})
```

#### After (v0.2.0)

```go
import "github.com/ncobase/ncore/oss"

// S3 Storage
s3Storage, err := oss.NewStorage(&oss.Config{
    Provider: "s3",
    Endpoint: "s3.amazonaws.com",
    Bucket:   "my-bucket",
    ID:       "access-key",
    Secret:   "secret-key",
})
if err != nil {
    log.Fatal(err)
}

// MinIO Storage
minioStorage, err := oss.NewStorage(&oss.Config{
    Provider: "minio",
    Endpoint: "localhost:9000",
    Bucket:   "my-bucket",
    ID:       "minioadmin",
    Secret:   "minioadmin",
})
if err != nil {
    log.Fatal(err)
}
```

#### Migration Steps

1. **Find All Storage Usage:**

```bash
grep -r "data/storage" . --include="*.go"
```

2. **Update Imports:**

```bash
# Using sed (macOS)
find . -name "*.go" -exec sed -i '' 's|github.com/ncobase/ncore/data/storage|github.com/ncobase/ncore/oss|g' {} \;

# Using sed (Linux)
find . -name "*.go" -exec sed -i 's|github.com/ncobase/ncore/data/storage|github.com/ncobase/ncore/oss|g' {} \;
```

3. **Update Constructor Calls:**

Replace type-specific constructors with unified `oss.NewStorage()`:

```go
// Old
storage.NewS3Storage(&storage.S3Config{...})
storage.NewMinIOStorage(&storage.MinIOConfig{...})
storage.NewAliyunStorage(&storage.AliyunConfig{...})

// New
oss.NewStorage(&oss.Config{Provider: "s3", ...})
oss.NewStorage(&oss.Config{Provider: "minio", ...})
oss.NewStorage(&oss.Config{Provider: "aliyun", ...})
```

4. **Update Config Structs:**

```go
// Old field names
type S3Config struct {
    Endpoint  string
    Bucket    string
    AccessKey string  // ← old name
    SecretKey string  // ← old name
}

// New field names
type Config struct {
    Provider string   // ← new field
    Endpoint string
    Bucket   string
    ID       string   // ← new name
    Secret   string   // ← new name
}
```

#### Supported Providers

| Provider | Value | Notes |
|----------|-------|-------|
| AWS S3 | `"s3"` | |
| MinIO | `"minio"` | S3-compatible |
| Aliyun OSS | `"aliyun"` | |
| Tencent COS | `"tencent"` | |
| Google Cloud Storage | `"gcs"` | |
| Azure Blob Storage | `"azure"` | |
| Qiniu Kodo | `"qiniu"` | |
| Synology NAS | `"synology"` | |
| Local Filesystem | `"local"` | For development |

---

### Search Initialization

**What Changed:** Search client initialization simplified from 14+ lines of boilerplate to 1 line using factory pattern.

**Why:** The old pattern required importing multiple adapters and manually wiring them. The new factory auto-detects available search engines.

#### Before (v0.1.x)

```go
import (
    "github.com/ncobase/ncore/data"
    "github.com/ncobase/ncore/data/search"
    "github.com/ncobase/ncore/data/metrics"
    
    esAdapter "github.com/ncobase/ncore/data/elasticsearch"
    esClient "github.com/ncobase/ncore/data/elasticsearch/client"
    osAdapter "github.com/ncobase/ncore/data/opensearch"
    osClient "github.com/ncobase/ncore/data/opensearch/client"
)

func NewData(conf *config.Data) (*Data, error) {
    d, cleanup, err := data.New(conf)
    if err != nil {
        return nil, err
    }
    
    // Manual adapter creation (14+ lines)
    var adapters []search.Adapter
    
    if es := d.GetElasticsearch(); es != nil {
        if c, ok := es.(*esClient.Client); ok {
            adapters = append(adapters, esAdapter.NewAdapter(c))
        }
    }
    
    if os := d.GetOpenSearch(); os != nil {
        if c, ok := os.(*osClient.Client); ok {
            adapters = append(adapters, osAdapter.NewAdapter(c))
        }
    }
    
    searchClient := search.NewClient(metrics.NoOpCollector{}, adapters...)
    
    return &Data{
        Data:         d,
        SearchClient: searchClient,
    }, nil
}
```

#### After (v0.2.0)

```go
import (
    "github.com/ncobase/ncore/data"
    "github.com/ncobase/ncore/data/search"
    
    // Drivers auto-register their adapters
    _ "github.com/ncobase/ncore/data/elasticsearch"
    _ "github.com/ncobase/ncore/data/opensearch"
)

func NewData(conf *config.Data) (*Data, error) {
    d, cleanup, err := data.New(conf)
    if err != nil {
        return nil, err
    }
    
    // Auto-detection (1 line)
    searchClient := search.NewClientFromData(d)
    
    return &Data{
        Data:         d,
        SearchClient: searchClient,
    }, nil
}
```

#### Migration Steps

1. **Find All Search Initialization Code:**

```bash
grep -r "search.NewClient" . --include="*.go"
grep -r "esAdapter.NewAdapter" . --include="*.go"
```

2. **Update Imports:**

Remove adapter-specific imports:

```go
// Remove these
// import esAdapter "github.com/ncobase/ncore/data/elasticsearch"
// import esClient "github.com/ncobase/ncore/data/elasticsearch/client"
// import osAdapter "github.com/ncobase/ncore/data/opensearch"
// import osClient "github.com/ncobase/ncore/data/opensearch/client"

// Keep this
import "github.com/ncobase/ncore/data/search"
```

3. **Replace Boilerplate:**

```go
// Delete this entire section
// var adapters []search.Adapter
// if es := d.GetElasticsearch(); es != nil {
//     if c, ok := es.(*esClient.Client); ok {
//         adapters = append(adapters, esAdapter.NewAdapter(c))
//     }
// }
// searchClient := search.NewClient(metrics.NoOpCollector{}, adapters...)

// Replace with this
searchClient := data.NewSearchClient(d)
```

4. **Add Driver Imports to main.go:**

```go
import (
    _ "github.com/ncobase/ncore/data/elasticsearch"
    // _ "github.com/ncobase/ncore/data/opensearch"
    // _ "github.com/ncobase/ncore/data/meilisearch"
)
```

#### Advanced: Custom Metrics Collector

If you need a custom metrics collector:

```go
// With custom collector
import "github.com/ncobase/ncore/data/metrics"

myCollector := metrics.NewPrometheusCollector()
searchClient := data.NewSearchClient(d, myCollector)
```

---

### Optional Messaging Connections

**What Changed:** RabbitMQ and Kafka connection failures are now warnings instead of fatal errors.

**Why:** Many applications can function without messaging during development or when messaging infrastructure is temporarily unavailable.

#### Before (v0.1.x)

```go
// Application fails to start if RabbitMQ is unavailable
d, cleanup, err := data.New(conf.Data)
if err != nil {
    log.Fatal(err) // Fatal: RabbitMQ connection failed
}
```

#### After (v0.2.0)

```go
// Application starts even if RabbitMQ is unavailable
d, cleanup, err := data.New(conf.Data)
if err != nil {
    log.Fatal(err)
}

// Check logs for warnings:
// [WARN] RabbitMQ connection failed (optional): dial tcp: connection refused
```

#### Migration Steps

**No code changes required.** This change is automatic.

However, you should:

1. **Review Logs:** Check application logs for messaging warnings
2. **Graceful Degradation:** Ensure your application handles missing messaging gracefully:

```go
rmq := d.GetRabbitMQ()
if rmq == nil {
    log.Warn("RabbitMQ not available, events will not be published")
    return nil // Gracefully skip event publishing
}

// Publish event
err := rmq.Publish(ctx, "events", message)
```

3. **Production Checks:** Add health checks to verify messaging is available in production:

```go
func (h *HealthHandler) Check(c *gin.Context) {
    status := "healthy"
    checks := make(map[string]string)
    
    // Check RabbitMQ
    if rmq := h.data.GetRabbitMQ(); rmq == nil {
        checks["rabbitmq"] = "unavailable"
        status = "degraded"
    } else {
        checks["rabbitmq"] = "ok"
    }
    
    c.JSON(200, gin.H{
        "status": status,
        "checks": checks,
    })
}
```

---

### Log-to-Search-Engine Hooks

**What Changed:** Logging hooks for sending logs to search engines are now modular and optional.

**Why:** Reduces core logging module dependencies. Only import hooks you need.

#### Before (v0.1.x)

```go
// Hooks were automatically included in logging module
import "github.com/ncobase/ncore/logging/logger"

// Configure in config.yaml:
// logger:
//   elasticsearch:
//     addresses: ["http://localhost:9200"]
```

#### After (v0.2.0)

```go
import (
    "github.com/ncobase/ncore/logging/logger"

    // Explicitly import the hooks you need
    _ "github.com/ncobase/ncore/logging/hooks/elasticsearch"
    // _ "github.com/ncobase/ncore/logging/hooks/meilisearch"
    // _ "github.com/ncobase/ncore/logging/hooks/opensearch"
)

// Same config.yaml works - hooks auto-initialize based on config
```

#### Migration Steps

1. **Add hook imports** to your `main.go`:

```go
import (
    _ "github.com/ncobase/ncore/logging/hooks/elasticsearch"
)
```

2. **Keep your existing config** - no changes needed:

```yaml
logger:
  level: 4
  format: json
  elasticsearch:
    addresses: ["http://localhost:9200"]
    username: elastic
    password: changeme
```

3. **Verify logs are being sent** to your search engine after startup.

---

## Step-by-Step Migration

### Example: Complete Migration of a REST API

This example shows migrating a typical REST API application.

#### Original Application (v0.1.x)

**File: `main.go`**

```go
package main

import (
    "context"
    "log"
    
    "github.com/gin-gonic/gin"
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/data"
    "github.com/ncobase/ncore/data/storage"
    "github.com/ncobase/ncore/logging/logger"
)

func main() {
    // Load config
    conf, err := config.LoadConfig("config.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    // Initialize logger
    cleanupLogger := logger.Init(conf.Logger)
    defer cleanupLogger()
    
    // Initialize data layer
    d, cleanup, err := data.New(conf.Data)
    if err != nil {
        logger.Fatalf(context.Background(), "data init failed: %v", err)
    }
    defer cleanup()
    
    // Initialize storage
    s3 := storage.NewS3Storage(&storage.S3Config{
        Endpoint:  conf.S3.Endpoint,
        Bucket:    conf.S3.Bucket,
        AccessKey: conf.S3.AccessKey,
        SecretKey: conf.S3.SecretKey,
    })
    
    // Start server
    r := gin.Default()
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    r.Run(":8080")
}
```

**File: `internal/user/data/data.go`**

```go
package data

import (
    "github.com/ncobase/ncore/data"
    "github.com/ncobase/ncore/data/search"
    "github.com/ncobase/ncore/data/metrics"
    
    esAdapter "github.com/ncobase/ncore/data/elasticsearch"
    esClient "github.com/ncobase/ncore/data/elasticsearch/client"
)

type Data struct {
    *data.Data
    SearchClient *search.Client
}

func New(conf *config.Data) (*Data, error) {
    d, cleanup, err := data.New(conf)
    if err != nil {
        return nil, err
    }
    
    // Search initialization (old way)
    var adapters []search.Adapter
    if es := d.GetElasticsearch(); es != nil {
        if c, ok := es.(*esClient.Client); ok {
            adapters = append(adapters, esAdapter.NewAdapter(c))
        }
    }
    searchClient := search.NewClient(metrics.NoOpCollector{}, adapters...)
    
    return &Data{
        Data:         d,
        SearchClient: searchClient,
    }, nil
}
```

#### Migrated Application (v0.2.0)

**File: `main.go`**

```go
package main

import (
    "context"
    "log"
    
    "github.com/gin-gonic/gin"
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/data"
    "github.com/ncobase/ncore/oss"          // ← Changed: OSS module
    "github.com/ncobase/ncore/logging/logger"
    
    // ← Added: Driver imports
    _ "github.com/ncobase/ncore/data/postgres"
    _ "github.com/ncobase/ncore/data/redis"
    _ "github.com/ncobase/ncore/data/elasticsearch"
)

func main() {
    // Load config
    conf, err := config.LoadConfig("config.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    // Initialize logger
    cleanupLogger := logger.Init(conf.Logger)
    defer cleanupLogger()
    
    // Initialize data layer
    d, cleanup, err := data.New(conf.Data)
    if err != nil {
        logger.Fatalf(context.Background(), "data init failed: %v", err)
    }
    defer cleanup()
    
    // Initialize storage (← Changed: OSS API)
    s3, err := oss.NewStorage(&oss.Config{
        Provider: "s3",                    // ← New field
        Endpoint: conf.S3.Endpoint,
        Bucket:   conf.S3.Bucket,
        ID:       conf.S3.AccessKey,       // ← Renamed from AccessKey
        Secret:   conf.S3.SecretKey,       // ← Renamed from SecretKey
    })
    if err != nil {
        logger.Fatalf(context.Background(), "storage init failed: %v", err)
    }
    
    // Start server
    r := gin.Default()
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    r.Run(":8080")
}
```

**File: `internal/user/data/data.go`**

```go
package data

import (
    "github.com/ncobase/ncore/data"
    "github.com/ncobase/ncore/data/search"
    
    // ← Removed: All adapter imports deleted
)

type Data struct {
    *data.Data
    SearchClient *search.Client
}

func New(conf *config.Data) (*Data, error) {
    d, cleanup, err := data.New(conf)
    if err != nil {
        return nil, err
    }
    
    // ← Changed: 1 line replaces 14+ lines
    searchClient := data.NewSearchClient(d)
    
    return &Data{
        Data:         d,
        SearchClient: searchClient,
    }, nil
}
```

**Summary of Changes:**

1. ✅ Added driver imports to `main.go`
2. ✅ Changed `data/storage` → `oss`
3. ✅ Updated storage config struct fields
4. ✅ Simplified search initialization

---

## Troubleshooting

### Issue: "driver not registered" panic

**Symptoms:**

```
panic: postgres: driver not registered
```

**Cause:** Missing driver import in your application.

**Solution:**

Add the driver import to your `main.go`:

```go
import _ "github.com/ncobase/ncore/data/postgres"
```

**Verification:**

```bash
# Check if driver is imported
grep "data/postgres" main.go
```

---

### Issue: Import errors after `go mod tidy`

**Symptoms:**

```
cannot find module providing package github.com/ncobase/ncore/data/storage
```

**Cause:** The `data/storage` package no longer exists in v0.2.0.

**Solution:**

1. Find all usages:
```bash
grep -r "data/storage" . --include="*.go"
```

2. Replace with `oss`:
```bash
# macOS
find . -name "*.go" -exec sed -i '' 's|data/storage|oss|g' {} \;

# Linux
find . -name "*.go" -exec sed -i 's|data/storage|oss|g' {} \;
```

3. Update constructor calls:
```go
// Old
storage.NewS3Storage(&storage.S3Config{...})

// New
oss.NewStorage(&oss.Config{Provider: "s3", ...})
```

---

### Issue: Search returns no results after migration

**Symptoms:** Search client initialized but returns empty results.

**Cause:** Search driver not imported, so `NewSearchClient()` creates empty client.

**Solution:**

1. Add search driver import to `main.go`:
```go
import _ "github.com/ncobase/ncore/data/elasticsearch"
```

2. Verify search engine is configured:
```yaml
# config.yaml
data:
  elasticsearch:
    url: "http://localhost:9200"
```

3. Check available engines:
```go
engines := searchClient.GetAvailableSearchEngines()
log.Printf("Available search engines: %v", engines)
// Should output: ["elasticsearch"]
```

---

### Issue: RabbitMQ 403 "access to vhost" error

**Symptoms:**

```
[ERROR] RabbitMQ connection failed: access to vhost '/myapp' refused
```

**Cause:** Vhost path encoding issue.

**Solution:**

Ensure your vhost config has a leading slash:

```yaml
# config.yaml
data:
  rabbitmq:
    vhost: "/myapp"  # ← Ensure leading slash
```

If the error persists, verify RabbitMQ permissions:

```bash
# Check vhost exists
rabbitmqctl list_vhosts

# Check user permissions
rabbitmqctl list_permissions -p "/myapp"

# Grant permissions
rabbitmqctl set_permissions -p "/myapp" "username" ".*" ".*" ".*"
```

---

### Issue: Binary size not reduced after migration

**Symptoms:** Binary size still ~90MB after v0.2.0 migration.

**Possible Causes:**

1. **Using `data/all` compatibility layer:**
```go
import _ "github.com/ncobase/ncore/data/all"  // ← Imports all drivers
```

**Solution:** Replace with specific drivers:
```go
import (
    _ "github.com/ncobase/ncore/data/postgres"
    _ "github.com/ncobase/ncore/data/redis"
)
```

2. **Not running `go mod tidy`:**

**Solution:**
```bash
rm go.sum
go mod tidy
go build
```

3. **Debug build:**

**Solution:** Build with optimizations:
```bash
go build -ldflags="-s -w" -o app
```

---

### Issue: Tests fail with "unexpected type" errors

**Symptoms:**

```go
cannot use data.NewSearchClient(d) (value of type *search.Client) as *search.Client value in assignment
```

**Cause:** Import path confusion between v0.1 and v0.2.

**Solution:**

1. Check `go.mod` for mixed versions:
```bash
cat go.mod | grep ncobase/ncore
```

All should be v0.2.0:
```
github.com/ncobase/ncore/config v0.2.0
github.com/ncobase/ncore/data v0.2.0
```

2. Update all modules:
```bash
go get github.com/ncobase/ncore/config@v0.2.0
go get github.com/ncobase/ncore/data@v0.2.0
go mod tidy
```

---

## FAQ

### Q: Can I use v0.1 and v0.2 in the same project?

**A:** No, mixing versions is not recommended and may cause conflicts. Choose one version for your entire project.

---

### Q: Do I need to update all ncore modules at once?

**A:** Yes. All ncore modules should be the same version (all v0.1.x or all v0.2.0).

Update script:
```bash
go get github.com/ncobase/ncore/config@v0.2.0
go get github.com/ncobase/ncore/data@v0.2.0
go get github.com/ncobase/ncore/logging@v0.2.0
go get github.com/ncobase/ncore/security@v0.2.0
go get github.com/ncobase/ncore/extension@v0.2.0
go mod tidy
```

---

### Q: What if I don't know which drivers I'm using?

**A:** Use the `data/all` compatibility layer temporarily:

```go
import _ "github.com/ncobase/ncore/data/all"
```

Then audit your logs to see which drivers are actually used, and switch to specific imports.

---

### Q: Will v0.1.x continue to be supported?

**A:** v0.1.x will receive security updates for 6 months after v0.2.0 release. No new features will be added.

---

### Q: Is there a performance difference between v0.1 and v0.2?

**A:** Yes, v0.2 has improved:
- **Build time**: 56% faster
- **Binary size**: 53% smaller
- **Dependencies**: 78% fewer

Runtime performance is similar, but reduced binary size improves deployment and startup times.

---

### Q: Can I migrate gradually?

**A:** For a single application, no - you must migrate entirely to v0.2. 

However, in a microservices architecture, you can migrate services one at a time:
- Service A: v0.1.24
- Service B: v0.2.0 (newly migrated)
- Service C: v0.1.24 (not yet migrated)

---

### Q: What if I encounter issues not covered here?

**A:**

1. Check the [CHANGELOG](CHANGELOG.md) for all changes
2. Review [CHANGELOG.md](CHANGELOG.md) for complete list of changes
3. Search [GitHub Issues](https://github.com/ncobase/ncore/issues)
4. Create a new issue with:
   - Your code snippet
   - Error message
   - Steps to reproduce
   - Output of `go version` and `go list -m all`

---

## Performance Benefits

### Before and After Comparison

**Test Application:** Basic REST API with PostgreSQL and Redis

| Metric | v0.1.24 | v0.2.0 | Improvement |
|--------|---------|--------|-------------|
| Binary Size | 92.3 MB | 43.1 MB | **-53.3%** |
| `go.mod` dependencies | 466 | 102 | **-78.1%** |
| Build time (cold) | 45.2s | 19.8s | **-56.2%** |
| Build time (cached) | 8.1s | 3.4s | **-58.0%** |
| Docker image size | 145 MB | 68 MB | **-53.1%** |
| Startup time | 1.2s | 1.1s | **-8.3%** |

### Binary Size Breakdown

**v0.1.24 (92.3 MB):**
- Application code: 15 MB
- NCore framework: 25 MB
- Database drivers: 12 MB
- Storage SDKs (S3, Azure, etc.): 24 MB
- Search engines: 16 MB

**v0.2.0 (43.1 MB):**
- Application code: 15 MB
- NCore framework: 8 MB
- Database drivers: 12 MB
- Storage SDKs: 0 MB (moved to `oss`)
- Search engines: 8 MB (only imported drivers)

---

## Additional Resources

- [CHANGELOG.md](CHANGELOG.md) - Complete list of changes in v0.2.0
- [MODULES.md](MODULES.md) - Multi-module architecture documentation
- [README.md](README.md) - Quick start guide
- [OSS Module Documentation](oss/README.md) - Object storage service guide

---

**Last Updated:** 2026-01-17  
**For questions or improvements to this guide, please [open an issue](https://github.com/ncobase/ncore/issues/new).**
