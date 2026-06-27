# ncore/data

Production-grade data access layer for Go applications. Provides unified connection management, read/write splitting, transactions, health checking, and optional Space-level multi-tenant isolation for relational databases, caches, search engines, and message queues.

## Installation

```bash
go get github.com/ncobase/ncore/data
```

Drivers are registered via blank import. Import only the backends your project uses:

```go
import (
    _ "github.com/ncobase/ncore/data/postgres"     // PostgreSQL
    _ "github.com/ncobase/ncore/data/mysql"         // MySQL
    _ "github.com/ncobase/ncore/data/sqlite"        // SQLite
    _ "github.com/ncobase/ncore/data/redis"         // Redis
    _ "github.com/ncobase/ncore/data/mongodb"       // MongoDB
    _ "github.com/ncobase/ncore/data/elasticsearch" // Elasticsearch
    _ "github.com/ncobase/ncore/data/meilisearch"   // Meilisearch
    _ "github.com/ncobase/ncore/data/kafka"         // Kafka
    _ "github.com/ncobase/ncore/data/rabbitmq"      // RabbitMQ
)
```

## Initialization

`New` returns a process-wide shared instance by default. Pass `true` to create an independent instance.

```go
cfg := config.GetConfig(viper.GetViper())

d, cleanup, err := data.New(cfg)
if err != nil {
    log.Fatal(err)
}
defer cleanup()
```

Functional options are applied after construction:

```go
data.WithMetricsCollector(myCollector)(d)
data.WithSpaceRouter(myRouter)(d)
```

For [google/wire](https://github.com/google/wire) projects, use the provided provider set:

```go
wire.Build(data.ProviderSet, ...)
```

## Configuration

```yaml
data:
  database:
    strategy: round_robin # round_robin | random | weight
    max_retry: 3
    master:
      driver: postgres
      source: "postgres://user:pass@host:5432/db?sslmode=disable"
      max_idle_conn: 5
      max_open_conn: 50
      max_life_time: 1h
      logging: false
    slaves:
      - driver: postgres
        source: "postgres://user:pass@replica:5432/db?sslmode=disable"
        weight: 2

  redis:
    addrs: ["localhost:6379"]
    password: ""
    db: 0

  mongodb:
    master:
      uri: "mongodb://localhost:27017/mydb"

  search:
    driver: meilisearch
    url: "http://localhost:7700"
    api_key: ""
    index_prefix: ""

  metrics:
    enabled: false
    batch_size: 100

  # Optional — omit to disable Space isolation.
  space:
    strategy: schema # schema (PostgreSQL) | database (MySQL)
    global_schema: public # default: "public"
    space_prefix: space_ # default: "space_"
    max_cached_pools: 100 # MySQL only, default: 100
```

## Driver System

Drivers self-register in `init()` using a pattern identical to `database/sql`. The core `data` package contains no backend-specific code.

Available driver categories and their registration functions:

| Category            | Registration function                    |
| ------------------- | ---------------------------------------- |
| Relational database | `RegisterDatabaseDriver(DatabaseDriver)` |
| Cache               | `RegisterCacheDriver(CacheDriver)`       |
| Search engine       | `RegisterSearchDriver(SearchDriver)`     |
| Message queue       | `RegisterMessageDriver(MessageDriver)`   |
| Object storage      | `RegisterStorageDriver(StorageDriver)`   |

Each driver implements the `Name() string`, `Connect(ctx, cfg) (any, error)`, and `Close(conn) error` methods. `DatabaseDriver` and `CacheDriver` additionally implement `Ping(ctx, conn) error`.

To list all registered drivers at runtime:

```go
drivers := data.ListRegisteredDrivers()
// map[string][]string{"database": ["postgres"], "cache": ["redis"], ...}
```

## Database Access

```go
master := d.GetMasterDB()          // *sql.DB — write node
slave, err := d.GetSlaveDB()       // *sql.DB — load-balanced read node
mgr := d.GetDBManager()            // *connection.DBManager — full manager

// Convenience aliases
master = d.DB()
slave, err = d.DBRead()
```

When no replicas are configured, `GetSlaveDB` falls back to the master.

### Read/Write Splitting

`DBManager` routes writes to the master and reads to replicas using the strategy configured in `data.database.strategy`:

| Strategy      | Behaviour                                        |
| ------------- | ------------------------------------------------ |
| `round_robin` | Cycles through replicas sequentially (default)   |
| `random`      | Selects a replica uniformly at random            |
| `weight`      | Distributes load proportionally by node `weight` |

### Other Backends

```go
d.GetRedis()        // any — Redis client (cast to your client type)
d.GetMongoManager() // any — MongoDB manager
d.GetElasticsearch() // any
d.GetOpenSearch()    // any
d.GetMeilisearch()   // any
```

## Transactions

All transaction methods store `*sql.Tx` in the context under `ContextKeyTransaction`. Retrieve it downstream with `GetTx(ctx)`.

```go
// Read-write transaction on master.
err := d.WithTx(ctx, func(ctx context.Context) error {
    tx, _ := data.GetTx(ctx) // *sql.Tx
    _, err := tx.ExecContext(ctx, "UPDATE ...")
    return err
})

// Read-only transaction on a replica.
err := d.WithTxRead(ctx, func(ctx context.Context) error {
    tx, _ := data.GetTx(ctx)
    // ...
    return nil
})
```

## Space Isolation

Space isolation provides database-level separation for multi-tenant applications. Each Space operates in its own PostgreSQL schema or MySQL database, complementing application-layer permission checks.

The feature is **opt-in**. Projects that do not require multi-tenancy simply omit `WithSpaceRouter`; all Space-aware methods fall back to ordinary transactions automatically.

### Architecture

```text
HTTP middleware
  └── data.SetSpaceID(ctx, space.Slug)

Service / repository
  └── d.WithSpaceTx(ctx, fn)
        ├── [space ID present + router injected]
        │     └── SpaceRouter.WithSpace(ctx, spaceID, fn)
        │           └── ExecInTx: BEGIN → SET LOCAL … → fn → COMMIT
        └── [no space ID or no router]
              └── d.WithTx(ctx, fn)   // transparent fallback
```

### Context helpers

```go
// Middleware: attach the current space to the request context.
ctx = data.SetSpaceID(ctx, space.Slug)

// Repository: read the space ID (returns "", false when absent).
spaceID, ok := data.GetSpaceID(ctx)
```

### Space-aware transactions

```go
// Read-write transaction in the current space.
err := d.WithSpaceTx(ctx, func(ctx context.Context) error {
    tx, _ := data.GetTx(ctx) // *sql.Tx scoped to the space schema/database
    return repo.Create(ctx, entity)
})

// Read-only transaction in the current space.
err := d.WithSpaceTxRead(ctx, func(ctx context.Context) error {
    return repo.List(ctx, filters)
})
```

### SpaceRouter interface

`ncore` provides no concrete implementation; each application supplies its own routing strategy:

```go
type SpaceRouter interface {
    WithSpace(ctx context.Context, spaceID string, fn func(context.Context) error) error
    WithSpaceRead(ctx context.Context, spaceID string, fn func(context.Context) error) error
}
```

Inject the router after construction:

```go
data.WithSpaceRouter(myRouter)(d)
```

Retrieve the injected router (e.g. for schema provisioning during Space creation):

```go
router := d.GetSpaceRouter() // SpaceRouter or nil
```

### ExecInTx

`ExecInTx` is the low-level helper for `SpaceRouter` implementors. It begins a transaction, executes an optional setup statement (e.g. `SET LOCAL search_path`), stores `*sql.Tx` in the context, runs the function, and commits or rolls back.

```go
func ExecInTx(
    ctx      context.Context,
    db       *sql.DB,
    opts     *sql.TxOptions, // nil for read-write defaults
    setupSQL string,         // "" to skip; use SET LOCAL, never SET
    fn       func(context.Context) error,
) error
```

### PostgreSQL — schema-per-space

```go
func (r *PostgresSchemaRouter) WithSpace(ctx context.Context, spaceID string, fn func(context.Context) error) error {
    schema, err := data.BuildSchemaName(r.prefix, spaceID)
    if err != nil {
        return err
    }
    schemaIdent, err := data.QuotePostgresIdentifier(schema)
    if err != nil {
        return err
    }
    globalIdent, err := data.QuotePostgresIdentifier(r.globalSchema)
    if err != nil {
        return err
    }
    setup := fmt.Sprintf("SET LOCAL search_path TO %s, %s", schemaIdent, globalIdent)
    return data.ExecInTx(ctx, r.masterDB, nil, setup, fn)
}

func (r *PostgresSchemaRouter) WithSpaceRead(ctx context.Context, spaceID string, fn func(context.Context) error) error {
    schema, err := data.BuildSchemaName(r.prefix, spaceID)
    if err != nil {
        return err
    }
    schemaIdent, err := data.QuotePostgresIdentifier(schema)
    if err != nil {
        return err
    }
    globalIdent, err := data.QuotePostgresIdentifier(r.globalSchema)
    if err != nil {
        return err
    }
    db := r.slaveDB
    if db == nil {
        db = r.masterDB
    }
    setup := fmt.Sprintf("SET LOCAL search_path TO %s, %s", schemaIdent, globalIdent)
    return data.ExecInTx(ctx, db, &sql.TxOptions{ReadOnly: true}, setup, fn)
}
```

Space schema provisioning:

```go
schemaIdent, err := data.QuotePostgresIdentifier(schema)
if err != nil {
    return err
}
_, err = masterDB.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaIdent))
```

**Critical:** always use `SET LOCAL` (transaction-scoped), never `SET` (session-scoped). `SET` persists on the connection after the transaction ends and contaminates subsequent pool borrowers.

### MySQL — database-per-space

Each Space uses its own `*sql.DB` connection pool. Maintain pools in a `sync.Map` or LRU map and evict stale entries to cap total pool count. No setup SQL is needed; the pool is already database-scoped.

```go
func (r *MySQLDatabaseRouter) WithSpace(ctx context.Context, spaceID string, fn func(context.Context) error) error {
    db, err := r.getOrCreatePool(spaceID)
    if err != nil {
        return err
    }
    return data.ExecInTx(ctx, db, nil, "", fn)
}
```

### Identifier security

Schema and database names cannot use SQL bind parameters. Always validate the space identifier before interpolating it into SQL:

```go
// ValidateSpaceID enforces ^[a-zA-Z0-9][a-zA-Z0-9_-]{0,62}$
if err := data.ValidateSpaceID(slug); err != nil {
    return err
}

// BuildSchemaName validates + normalises (lowercase, hyphens → underscores).
schema, err := data.BuildSchemaName("space_", slug)
// "Acme-Corp" → "space_acme_corp"
```

Never pass raw HTTP input as a space identifier. Resolve and validate it via an authoritative Space lookup that confirms the caller's membership.

### Space isolation configuration

Helper methods on `*config.Space`:

```go
cfg.Space.IsEnabled()              // true when Strategy != ""
cfg.Space.EffectiveGlobalSchema()  // "public" when unset
cfg.Space.EffectiveSpacePrefix()   // "space_" when unset
cfg.Space.EffectiveMaxCachedPools() // 100 when unset
```

## Health Checks

`Health` probes every configured backend and returns a composite status map:

```go
result := d.Health(ctx)
// map[string]any{
//   "status":    "healthy" | "degraded",
//   "timestamp": time.Time,
//   "services": map[string]any{
//     "database": {"healthy": true,  "response_ms": 2,    "error": ""},
//     "mongodb":  {"healthy": false, "response_ms": 0,    "error": "..."},
//     "rabbitmq": {"healthy": true,  "error": ""},
//     "kafka":    {"healthy": true,  "error": ""},
//     "search":   {"healthy": false, "error": "..."},
//   },
// }
```

Backends that are not configured are omitted from `services` and do not affect the overall status.

For lightweight connectivity checks use `Ping`:

```go
err := d.Ping(ctx) // pings the relational database only
```

## Metrics

Implement `metrics.Collector` and apply it as an option:

```go
data.WithMetricsCollector(myCollector)(d)
// or, for the extension adapter:
data.WithExtensionCollector(myExtCollector)(d)
```

`metrics.Collector` interface:

```go
type Collector interface {
    DBQuery(duration time.Duration, err error)
    DBTransaction(err error)
    DBConnections(count int)
    RedisCommand(command string, err error)
    RedisConnections(count int)
    MongoOperation(operation string, err error)
    SearchQuery(engine string, err error)
    SearchIndex(engine, operation string)
    MQPublish(system string, err error)
    MQConsume(system string, err error)
    HealthCheck(component string, healthy bool)
}
```

The built-in `DataCollector` accumulates counters in memory and supports pluggable `Storage` backends. Retrieve aggregated statistics:

```go
stats := d.GetStats()
// map with sub-maps: "database", "redis", "mongodb", "search", "messaging", "health"
```

When metrics are not needed, the default `NoOpCollector` imposes zero overhead.

## Ent ORM Mixins

The `data/entgo/mixin` sub-package provides reusable [Ent](https://entgo.io/) schema mixins for projects that use Ent ORM.

**Timestamp mixins:**

```go
import "github.com/ncobase/ncore/data/entgo/mixin"

func (MySchema) Mixin() []ent.Mixin {
    return []ent.Mixin{
        mixin.TimeAt{},    // created_at + updated_at (Unix milliseconds)
    }
}
```

Available timestamp fields: `CreatedAt`, `UpdatedAt`, `DeletedAt`, `ExpiredAt`, `Expires`, `Released`, `StartedAt`, `CompletedAt`.

**Primary key mixins:**

```go
mixin.PrimaryKey                       // nanoid string primary key
mixin.PrimaryKeyAlias{...}            // named alias for a different storage key
mixin.CustomPrimaryKey{Length, DefaultFunc} // configurable key
```

**Field mixins** (via `StringMixin`, `BoolMixin`, `IntMixin`, etc.) provide typed, composable schema fields with configurable optionality, uniqueness, default values, and comments.

## Lifecycle

```go
// Close shuts down all connections and releases resources.
// Returns a slice of errors (one per failing backend).
errs := d.Close()
```

After `Close`, any method on the same `*Data` instance returns an error immediately. The cleanup function returned by `New` calls `Close` and is safe to call multiple times.
