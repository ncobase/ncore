# ncore/data

面向 Go 应用的生产级数据访问层。提供统一的连接管理、读写分离、事务、健康检查，以及面向关系型数据库、缓存、搜索引擎和消息队列的可选 Space 级多租户隔离。

## 安装

```bash
go get github.com/ncobase/ncore/data
```

通过空白导入注册驱动，按需引入：

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

## 初始化

`New` 默认返回进程级共享单例。传入 `true` 可创建独立实例。

```go
cfg := config.GetConfig(viper.GetViper())

d, cleanup, err := data.New(cfg)
if err != nil {
    log.Fatal(err)
}
defer cleanup()
```

功能选项在构造后应用：

```go
data.WithMetricsCollector(myCollector)(d)
data.WithSpaceRouter(myRouter)(d)
```

对于使用 [google/wire](https://github.com/google/wire) 的项目，可使用内置 Provider Set：

```go
wire.Build(data.ProviderSet, ...)
```

## 配置

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

  # 可选 — 不配置则禁用 Space 隔离。
  space:
    strategy: schema # schema（PostgreSQL）| database（MySQL）
    global_schema: public # 默认值："public"
    space_prefix: space_ # 默认值："space_"
    max_cached_pools: 100 # 仅 MySQL 有效，默认值：100
```

## 驱动系统

驱动在 `init()` 中自注册，模式与 `database/sql` 相同。核心 `data` 包不含任何后端特定代码。

| 类别         | 注册函数                                 |
| ------------ | ---------------------------------------- |
| 关系型数据库 | `RegisterDatabaseDriver(DatabaseDriver)` |
| 缓存         | `RegisterCacheDriver(CacheDriver)`       |
| 搜索引擎     | `RegisterSearchDriver(SearchDriver)`     |
| 消息队列     | `RegisterMessageDriver(MessageDriver)`   |
| 对象存储     | `RegisterStorageDriver(StorageDriver)`   |

每个驱动实现 `Name() string`、`Connect(ctx, cfg) (any, error)` 和 `Close(conn) error` 方法。`DatabaseDriver` 和 `CacheDriver` 额外实现 `Ping(ctx, conn) error`。

运行时查看所有已注册驱动：

```go
drivers := data.ListRegisteredDrivers()
// map[string][]string{"database": ["postgres"], "cache": ["redis"], ...}
```

## 数据库访问

```go
master := d.GetMasterDB()          // *sql.DB — 写节点
slave, err := d.GetSlaveDB()       // *sql.DB — 负载均衡读节点
mgr := d.GetDBManager()            // *connection.DBManager — 完整管理器

// 简写别名
master = d.DB()
slave, err = d.DBRead()
```

未配置副本时，`GetSlaveDB` 自动回退至主节点。

### 读写分离

`DBManager` 将写操作路由至主节点，读操作路由至副本，策略由 `data.database.strategy` 配置：

| 策略          | 行为                           |
| ------------- | ------------------------------ |
| `round_robin` | 轮询副本（默认）               |
| `random`      | 随机选择副本                   |
| `weight`      | 按节点 `weight` 按比例分配流量 |

### 其他后端

```go
d.GetRedis()         // any — Redis 客户端（类型断言为具体类型）
d.GetMongoManager()  // any — MongoDB 管理器
d.GetElasticsearch() // any
d.GetOpenSearch()    // any
d.GetMeilisearch()   // any
```

## 事务

所有事务方法将 `*sql.Tx` 以 `ContextKeyTransaction` 为键存入 context，下游通过 `GetTx(ctx)` 取用。

```go
// 主节点读写事务。
err := d.WithTx(ctx, func(ctx context.Context) error {
    tx, _ := data.GetTx(ctx) // *sql.Tx
    _, err := tx.ExecContext(ctx, "UPDATE ...")
    return err
})

// 副本只读事务。
err := d.WithTxRead(ctx, func(ctx context.Context) error {
    tx, _ := data.GetTx(ctx)
    // ...
    return nil
})
```

## Space 隔离

Space 隔离为多租户应用提供数据库级别的数据分离。每个 Space 在独立的 PostgreSQL Schema 或 MySQL Database 中运行，作为应用层权限检查的补充。

该功能**按需启用**。不需要多租户的项目（如 gainmats、yonfile）不注入 `SpaceRouter` 即可，Space 感知方法将自动回退为普通事务，无任何额外开销。

### 架构

```text
HTTP 中间件
  └── data.SetSpaceID(ctx, space.Slug)

Service / Repository
  └── d.WithSpaceTx(ctx, fn)
        ├── [ctx 含 Space ID 且已注入 SpaceRouter]
        │     └── SpaceRouter.WithSpace(ctx, spaceID, fn)
        │           └── ExecInTx: BEGIN → SET LOCAL … → fn → COMMIT
        └── [无 Space ID 或无 SpaceRouter]
              └── d.WithTx(ctx, fn)   // 透明回退
```

### Context 辅助函数

```go
// 中间件：将当前 Space 写入请求 context。
ctx = data.SetSpaceID(ctx, space.Slug)

// Repository：读取 Space ID（不存在时返回 "", false）。
spaceID, ok := data.GetSpaceID(ctx)
```

### Space 感知事务

```go
// 当前 Space 的读写事务。
err := d.WithSpaceTx(ctx, func(ctx context.Context) error {
    tx, _ := data.GetTx(ctx) // *sql.Tx，已限定至 Space 的 Schema/Database
    return repo.Create(ctx, entity)
})

// 当前 Space 的只读事务。
err := d.WithSpaceTxRead(ctx, func(ctx context.Context) error {
    return repo.List(ctx, filters)
})
```

### SpaceRouter 接口

`ncore` 不提供任何具体实现，由各应用自行实现路由策略：

```go
type SpaceRouter interface {
    WithSpace(ctx context.Context, spaceID string, fn func(context.Context) error) error
    WithSpaceRead(ctx context.Context, spaceID string, fn func(context.Context) error) error
}
```

构造后注入路由器：

```go
data.WithSpaceRouter(myRouter)(d)
```

获取已注入的路由器（如在 Space 创建时进行 Schema 预置）：

```go
router := d.GetSpaceRouter() // SpaceRouter 或 nil
```

### ExecInTx

`ExecInTx` 是为 `SpaceRouter` 实现者提供的底层工具函数。它开启事务、执行可选的 setup SQL（如 `SET LOCAL search_path`）、将 `*sql.Tx` 存入 context、执行函数，并在成功时提交、失败时回滚。

```go
func ExecInTx(
    ctx      context.Context,
    db       *sql.DB,
    opts     *sql.TxOptions, // nil 表示默认读写
    setupSQL string,         // "" 表示跳过；必须使用 SET LOCAL，禁止使用 SET
    fn       func(context.Context) error,
) error
```

### PostgreSQL — Schema per Space

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

Space Schema 预置：

```go
schemaIdent, err := data.QuotePostgresIdentifier(schema)
if err != nil {
    return err
}
_, err = masterDB.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaIdent))
```

**关键约束：** 必须使用 `SET LOCAL`（事务级），禁止使用 `SET`（会话级）。`SET` 在事务结束后会保留在连接上，污染后续从连接池借用该连接的调用者。

### MySQL — Database per Space

每个 Space 使用独立的 `*sql.DB` 连接池。在 `sync.Map` 或 LRU map 中维护连接池，并设定上限进行淘汰。无需 setup SQL，连接池本身已绑定至对应 Database。

```go
func (r *MySQLDatabaseRouter) WithSpace(ctx context.Context, spaceID string, fn func(context.Context) error) error {
    db, err := r.getOrCreatePool(spaceID)
    if err != nil {
        return err
    }
    return data.ExecInTx(ctx, db, nil, "", fn)
}
```

### 标识符安全

Schema 名和 Database 名无法使用 SQL 参数绑定，必须在字符串插值前对 Space 标识符进行验证：

```go
// ValidateSpaceID 强制要求满足 ^[a-zA-Z0-9][a-zA-Z0-9_-]{0,62}$
if err := data.ValidateSpaceID(slug); err != nil {
    return err
}

// BuildSchemaName 验证 + 规范化（转小写，连字符替换为下划线）。
schema, err := data.BuildSchemaName("space_", slug)
// "Acme-Corp" → "space_acme_corp"
```

禁止将 HTTP 原始输入直接用作 Space 标识符。必须通过权威的 Space 查询进行解析和成员身份校验后方可使用。

### Space 隔离配置辅助方法

```go
cfg.Space.IsEnabled()               // Strategy != "" 时返回 true
cfg.Space.EffectiveGlobalSchema()   // 未设置时返回 "public"
cfg.Space.EffectiveSpacePrefix()    // 未设置时返回 "space_"
cfg.Space.EffectiveMaxCachedPools() // 未设置时返回 100
```

## 健康检查

`Health` 探测所有已配置的后端，返回聚合状态：

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

未配置的后端不出现在 `services` 中，也不影响整体状态。

轻量连接性探测使用 `Ping`：

```go
err := d.Ping(ctx) // 仅探测关系型数据库
```

## 指标采集

实现 `metrics.Collector` 接口并作为选项注入：

```go
data.WithMetricsCollector(myCollector)(d)
// 或使用扩展适配器：
data.WithExtensionCollector(myExtCollector)(d)
```

`metrics.Collector` 接口：

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

内置 `DataCollector` 在内存中累积计数器，并支持可插拔的 `Storage` 后端。获取聚合统计数据：

```go
stats := d.GetStats()
// 包含子 map："database"、"redis"、"mongodb"、"search"、"messaging"、"health"
```

不需要指标采集时，默认 `NoOpCollector` 零开销。

## Ent ORM Mixin

`data/entgo/mixin` 子包为使用 [Ent](https://entgo.io/) ORM 的项目提供可复用的 Schema Mixin。

**时间戳 Mixin：**

```go
import "github.com/ncobase/ncore/data/entgo/mixin"

func (MySchema) Mixin() []ent.Mixin {
    return []ent.Mixin{
        mixin.TimeAt{}, // created_at + updated_at（Unix 毫秒）
    }
}
```

可用时间字段：`CreatedAt`、`UpdatedAt`、`DeletedAt`、`ExpiredAt`、`Expires`、`Released`、`StartedAt`、`CompletedAt`。

**主键 Mixin：**

```go
mixin.PrimaryKey                                      // nanoid 字符串主键
mixin.PrimaryKeyAlias{...}                           // 带自定义存储键名的别名主键
mixin.CustomPrimaryKey{Length, DefaultFunc}          // 可配置主键
```

**字段 Mixin**（`StringMixin`、`BoolMixin`、`IntMixin` 等）提供带类型的、可组合的 Schema 字段，支持配置可选性、唯一性、默认值和字段注释。

## 生命周期

```go
// Close 关闭所有连接并释放资源。
// 返回错误切片（每个失败的后端对应一个错误）。
errs := d.Close()
```

`Close` 之后，同一 `*Data` 实例上的所有方法将立即返回错误。`New` 返回的 cleanup 函数内部调用 `Close`，可安全多次调用。
