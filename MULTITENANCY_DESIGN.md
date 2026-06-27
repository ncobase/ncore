# ncore/data 多租户隔离方案设计

> 版本：v1.0
> 日期：2026-02-28
> 适用范围：ncore v0.2.x 及以上版本

---

## 一、背景与现状

### 1.1 ncore 定位

ncore 是跨项目的通用基础库，当前引用它的项目包括：

| 项目          | 多租户需求           | 主数据库           | ORM |
| ------------- | -------------------- | ------------------ | --- |
| ncobase       | 完整（Space = 租户） | PostgreSQL         | Ent |
| datidy        | 完整（Space 模块）   | PostgreSQL + MySQL | Ent |
| gainmats      | 无                   | PostgreSQL         | Ent |
| yonfile       | 轻量（Header 级别）  | PostgreSQL         | Ent |
| single/server | 待定                 | PostgreSQL         | Ent |

ncore/data 目前的直接依赖极少（`google/wire`、`spf13/viper`），这一轻量特性必须保持。

### 1.2 现状问题

**ncore/data 层**：单一共享 DBManager，完全无 Space 感知。

**ncobase 等应用层**：租户隔离仅靠应用层 middleware + service 检查实现：

```text
请求 → middleware 验证 space_id → service 层校验 → repository 直接用共享 DB
```

**风险**：应用层一旦出现逻辑漏洞，不存在数据库级别的兜底隔离。

### 1.3 目标

- PostgreSQL：每个 Space 使用独立 schema（`SET LOCAL search_path`）
- MySQL：每个 Space 使用独立 database（独立连接池）
- **不破坏现有 API**，不为无多租户需求的项目增加任何负担
- **不引入新的 ncore/data 核心依赖**

---

## 二、设计原则

```text
1. 可选（Opt-in）   — 不需要多租户的项目，零感知、零开销
2. 最小依赖         — ncore/data 核心层仅用标准库实现接口，无第三方包
3. 接口驱动         — ncore 定义抽象，具体实现由应用层注入
4. 向后兼容         — 现有的 WithTx / GetMasterDB 等 API 保持不变
5. 上下文传播       — Space ID 通过 context.Context 流动，与现有模式一致
6. 事务绑定安全     — Space 上下文必须绑定到事务（transaction-scoped），
                       禁止绑定到连接（connection-scoped），防止连接池污染
```

---

## 三、隔离策略对比

### 3.1 三种模式

```text
┌────────────────────────────────────────────────────────────────────┐
│  模式 A：共享表 + space_id 列（当前现状）                           │
│                                                                    │
│  DB: ncobase                                                       │
│  └── Schema: public                                                │
│       ├── users     (space_id 外键)                                │
│       ├── topics    (space_id 外键)                                │
│       └── settings  (space_id 外键)                                │
│                                                                    │
│  隔离层：应用层   强度：弱   复杂度：低                             │
└────────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────────┐
│  模式 B：Schema-per-Space（PostgreSQL）/ DB-per-Space（MySQL）      │
│                                                                    │
│  PostgreSQL:                    MySQL:                             │
│  DB: ncobase                    DB: ncobase_global (系统表)        │
│  ├── Schema: public (全局)      DB: ncobase_acme   (Space: acme)   │
│  │    ├── users                 DB: ncobase_globex (Space: globex) │
│  │    ├── spaces                                                   │
│  │    └── auth_tokens                                              │
│  ├── Schema: space_acme                                            │
│  │    ├── settings                                                 │
│  │    └── topics                                                   │
│  └── Schema: space_globex                                          │
│       ├── settings                                                 │
│       └── topics                                                   │
│                                                                    │
│  隔离层：数据库   强度：强   复杂度：中                             │
└────────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────────┐
│  模式 C：独立实例（每个 Space 独立部署）                             │
│                                                                    │
│  隔离层：物理     强度：最强  复杂度：极高（不在本方案范围内）        │
└────────────────────────────────────────────────────────────────────┘
```

### 3.2 推荐选择

**ncore 支持 A + B 两种模式，通过注入不同的 SpaceRouter 实现切换。**

| 项目           | 推荐策略       | 说明                             |
| -------------- | -------------- | -------------------------------- |
| gainmats       | 模式 A（不变） | 无多租户需求，不接入 SpaceRouter |
| yonfile        | 模式 A（不变） | 当前 Header 级别已足够           |
| datidy/ncobase | 模式 B         | 企业级 SaaS，需数据库级隔离      |

---

## 四、职责边界

```text
┌─────────────────────────────────────────────────────────┐
│                    应用层（ncobase 等）                   │
│                                                         │
│  ① 实现 SpaceRouter 接口（PG schema / MySQL DB 切换）    │
│  ② Middleware 从请求中提取 space_id 写入 context         │
│  ③ Space 生命周期（创建 Schema/DB、运行迁移、删除）       │
│  ④ Ent Client 工厂（per-space schema 的 Ent 封装）       │
│  ⑤ 数据分层（哪些表全局，哪些表 Space 级）               │
└─────────────────────────┬───────────────────────────────┘
                          │ 注入 SpaceRouter
┌─────────────────────────▼───────────────────────────────┐
│                  ncore/data 层（本库）                    │
│                                                         │
│  ① 定义 SpaceRouter 接口（仅标准库）                     │
│  ② context 工具：SetSpaceID / GetSpaceID                │
│  ③ WithSpaceTx：有 router 时走 Space 路径，否则降级普通Tx │
│  ④ 现有 API 完全不变：WithTx / GetMasterDB 等            │
└─────────────────────────────────────────────────────────┘
```

---

## 五、ncore/data API 设计（新增部分）

### 5.1 新增文件：`data/space.go`

> **零新依赖，仅使用标准库**

```go
package data

import (
    "context"
    "database/sql"
    "fmt"
    "regexp"
    "time"
)

// ──────────────────────────────────────────
// Context 工具
// ──────────────────────────────────────────

const ContextKeySpaceID ContextKey = "space_id"

// SetSpaceID 将 Space ID 存入 context。由应用层 middleware 调用。
func SetSpaceID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, ContextKeySpaceID, id)
}

// GetSpaceID 从 context 中取出 Space ID。
func GetSpaceID(ctx context.Context) (string, bool) {
    id, ok := ctx.Value(ContextKeySpaceID).(string)
    return id, ok && id != ""
}

// ──────────────────────────────────────────
// SpaceRouter 接口
// ──────────────────────────────────────────

// SpaceRouter 是 Space 级数据库路由的核心抽象。
// 应用层实现此接口并通过 WithSpaceRouter Option 注入到 Data。
//
// 实现约定：
//   - Space 上下文必须绑定到事务（SET LOCAL），禁止绑定到连接（SET）。
//   - fn 执行期间，写入 context 的 *sql.Tx 应指向该 Space 的专属命名空间。
//   - 实现负责事务的 Begin / Commit / Rollback 生命周期。
type SpaceRouter interface {
    // WithSpace 在指定 Space 的命名空间内执行 fn（读写）。
    WithSpace(ctx context.Context, spaceID string, fn func(ctx context.Context) error) error

    // WithSpaceRead 在指定 Space 的命名空间内执行 fn（只读，可使用从库）。
    WithSpaceRead(ctx context.Context, spaceID string, fn func(ctx context.Context) error) error
}

// ──────────────────────────────────────────
// Data 层集成
// ──────────────────────────────────────────

// WithSpaceRouter 注入 Space 路由器的 Option。
func WithSpaceRouter(router SpaceRouter) Option {
    return func(d *Data) {
        d.spaceRouter = router
    }
}

// WithSpaceTx 在 Space 上下文中执行读写事务。
//
// 行为：
//   - 若 context 中有 space_id 且已注入 SpaceRouter → 走 Space 专属路径
//   - 否则 → 降级为普通 WithTx（向后兼容，gainmats 等项目无感知）
func (d *Data) WithSpaceTx(ctx context.Context, fn func(ctx context.Context) error) error {
    spaceID, hasSpace := GetSpaceID(ctx)

    d.mu.RLock()
    router := d.spaceRouter
    d.mu.RUnlock()

    if !hasSpace || router == nil {
        return d.WithTx(ctx, fn)
    }
    return router.WithSpace(ctx, spaceID, fn)
}

// WithSpaceTxRead 在 Space 上下文中执行只读事务。
func (d *Data) WithSpaceTxRead(ctx context.Context, fn func(ctx context.Context) error) error {
    spaceID, hasSpace := GetSpaceID(ctx)

    d.mu.RLock()
    router := d.spaceRouter
    d.mu.RUnlock()

    if !hasSpace || router == nil {
        return d.WithTxRead(ctx, fn)
    }
    return router.WithSpaceRead(ctx, spaceID, fn)
}

// ──────────────────────────────────────────
// 安全工具
// ──────────────────────────────────────────

// spaceIDPattern 白名单正则：只允许小写字母、数字、连字符、下划线。
// 防止通过 space_id 进行 SQL 注入（schema/db 名会拼接到 DDL/DML 中）。
var spaceIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`)

// ValidateSpaceID 校验 space_id 格式是否安全。
// 应用层在构建 schema/db 名时必须先调用此函数。
func ValidateSpaceID(id string) error {
    if !spaceIDPattern.MatchString(id) {
        return fmt.Errorf("invalid space id %q: must match ^[a-z0-9][a-z0-9_-]{0,62}$", id)
    }
    return nil
}

// ──────────────────────────────────────────
// 工具函数（供 SpaceRouter 实现使用）
// ──────────────────────────────────────────

// ExecInTx 在给定 db 上开启事务，执行 setup SQL（如 SET LOCAL），再调用 fn。
// 是 PostgresSchemaRouter 和其他实现的底层工具，避免重复代码。
//
// setupSQL 示例（PostgreSQL）：
//   "SET LOCAL search_path TO space_acme, public"
//
// 若 setupSQL 为空则直接执行 fn（MySQL 模式）。
func ExecInTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions, setupSQL string, fn func(ctx context.Context) error) error {
    tx, err := db.BeginTx(ctx, opts)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }

    if setupSQL != "" {
        if _, err = tx.ExecContext(ctx, setupSQL); err != nil {
            _ = tx.Rollback()
            return fmt.Errorf("setup tx: %w", err)
        }
    }

    txCtx := context.WithValue(ctx, ContextKeyTransaction, tx)

    start := time.Now()
    fnErr := fn(txCtx)
    _ = start // 供扩展：应用层可在此加 metrics

    if fnErr != nil {
        if rbErr := tx.Rollback(); rbErr != nil {
            return fmt.Errorf("fn err: %w; rollback err: %v", fnErr, rbErr)
        }
        return fnErr
    }

    if err = tx.Commit(); err != nil {
        return fmt.Errorf("commit: %w", err)
    }
    return nil
}
```

### 5.2 `data/data.go` 修改（最小化）

在 `Data` 结构体中增加一个字段：

```go
type Data struct {
    Conn        *connection.Connections
    collector   metrics.Collector
    conf        *config.Config
    spaceRouter SpaceRouter  // 新增，nil 表示不启用 Space 级隔离
    closed      bool
    mu          sync.RWMutex
}
```

**其余现有代码不变。**

### 5.3 `data/config/database.go` 配置扩展（可选）

如需通过配置文件声明 Space 隔离策略：

```go
// Space Space 级数据库隔离策略配置（可选，不配置则不启用隔离）
type Space struct {
    // Strategy: "schema"（PostgreSQL）| "database"（MySQL）| ""（不启用）
    Strategy    string `json:"strategy"     yaml:"strategy"`
    // GlobalSchema: 全局表所在 schema，默认 "public"（仅 PG）
    GlobalSchema string `json:"global_schema" yaml:"global_schema"`
    // SpacePrefix: schema/db 名前缀，默认 "space_"
    SpacePrefix  string `json:"space_prefix"  yaml:"space_prefix"`
    // MaxCachedPools: 最大缓存连接池数（仅 MySQL），默认 100
    MaxCachedPools int  `json:"max_cached_pools" yaml:"max_cached_pools"`
}
```

在 `Config` 中增加此字段：

```go
type Config struct {
    *Database `yaml:"database" json:"database"`
    *Space    `yaml:"space"    json:"space"`    // 新增
    // ... 其余不变
}
```

---

## 六、应用层实现指南

### 6.1 PostgreSQL：Schema-per-Space 实现

> 在 ncobase/datidy 等应用层实现，**不放在 ncore**

```go
// internal/space/postgres_router.go（应用层）

package space

import (
    "context"
    "database/sql"
    "fmt"

    "github.com/ncobase/ncore/data"
)

// PostgresSchemaRouter 实现 data.SpaceRouter。
// 每个 Space 对应 PostgreSQL 中的一个独立 schema。
// 通过 SET LOCAL search_path 实现事务级 schema 切换（连接池安全）。
type PostgresSchemaRouter struct {
    masterDB     *sql.DB
    slaveDB      *sql.DB
    globalSchema string // 全局 schema，默认 "public"
    spacePrefix  string // schema 名前缀，默认 "space_"
}

func NewPostgresSchemaRouter(masterDB, slaveDB *sql.DB, globalSchema, spacePrefix string) *PostgresSchemaRouter {
    if globalSchema == "" {
        globalSchema = "public"
    }
    if spacePrefix == "" {
        spacePrefix = "space_"
    }
    return &PostgresSchemaRouter{
        masterDB:     masterDB,
        slaveDB:      slaveDB,
        globalSchema: globalSchema,
        spacePrefix:  spacePrefix,
    }
}

func (r *PostgresSchemaRouter) schemaName(spaceID string) (string, error) {
    if err := data.ValidateSpaceID(spaceID); err != nil {
        return "", err
    }
    return r.spacePrefix + spaceID, nil
}

// WithSpace 在 Space schema 内执行读写事务。
// SET LOCAL search_path 只在当前事务有效，事务结束后连接自动重置，对连接池安全。
func (r *PostgresSchemaRouter) WithSpace(ctx context.Context, spaceID string, fn func(ctx context.Context) error) error {
    schema, err := r.schemaName(spaceID)
    if err != nil {
        return err
    }
    // search_path 不支持参数化，通过 ValidateSpaceID 白名单校验后拼接。
    setupSQL := fmt.Sprintf("SET LOCAL search_path TO %s, %s", schema, r.globalSchema)
    return data.ExecInTx(ctx, r.masterDB, nil, setupSQL, fn)
}

// WithSpaceRead 在 Space schema 内执行只读事务（优先使用从库）。
func (r *PostgresSchemaRouter) WithSpaceRead(ctx context.Context, spaceID string, fn func(ctx context.Context) error) error {
    schema, err := r.schemaName(spaceID)
    if err != nil {
        return err
    }
    db := r.slaveDB
    if db == nil {
        db = r.masterDB
    }
    setupSQL := fmt.Sprintf("SET LOCAL search_path TO %s, %s", schema, r.globalSchema)
    return data.ExecInTx(ctx, db, &sql.TxOptions{ReadOnly: true}, setupSQL, fn)
}

// ──────────────────────────────────────────
// Space 生命周期（DDL 操作，在应用层实现）
// ──────────────────────────────────────────

// CreateSchema 为新 Space 创建独立 schema。在创建 Space 时调用。
func (r *PostgresSchemaRouter) CreateSchema(ctx context.Context, spaceID string) error {
    schema, err := r.schemaName(spaceID)
    if err != nil {
        return err
    }
    _, err = r.masterDB.ExecContext(ctx,
        fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema))
    return err
}

// DropSchema 删除 Space schema（慎用，不可逆）。
func (r *PostgresSchemaRouter) DropSchema(ctx context.Context, spaceID string) error {
    schema, err := r.schemaName(spaceID)
    if err != nil {
        return err
    }
    _, err = r.masterDB.ExecContext(ctx,
        fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema))
    return err
}
```

### 6.2 MySQL：Database-per-Space 实现

```go
// internal/space/mysql_router.go（应用层）

package space

import (
    "container/list"
    "context"
    "database/sql"
    "fmt"
    "strings"
    "sync"

    "github.com/ncobase/ncore/data"
)

// MySQLDatabaseRouter 实现 data.SpaceRouter。
// 每个 Space 对应 MySQL 中的一个独立 database，使用独立连接池。
// 采用 LRU 策略控制最大并发连接池数量（仅用标准库 container/list 实现）。
type MySQLDatabaseRouter struct {
    baseDSN     string // DSN 模板，不含 dbname，如 "user:pass@tcp(host:3306)/"
    spacePrefix string // db 名前缀，默认 "space_"
    maxPools    int    // LRU 最大数量，默认 100
    maxIdle     int
    maxOpen     int

    mu      sync.Mutex
    pools   map[string]*sql.DB    // spaceID -> pool
    lruList *list.List            // LRU 驱逐顺序
    lruMap  map[string]*list.Element
}

func NewMySQLDatabaseRouter(baseDSN, spacePrefix string, maxPools, maxIdle, maxOpen int) *MySQLDatabaseRouter {
    if spacePrefix == "" {
        spacePrefix = "space_"
    }
    if maxPools <= 0 {
        maxPools = 100
    }
    return &MySQLDatabaseRouter{
        baseDSN:     baseDSN,
        spacePrefix: spacePrefix,
        maxPools:    maxPools,
        maxIdle:     maxIdle,
        maxOpen:     maxOpen,
        pools:       make(map[string]*sql.DB),
        lruList:     list.New(),
        lruMap:      make(map[string]*list.Element),
    }
}

func (r *MySQLDatabaseRouter) dbName(spaceID string) (string, error) {
    if err := data.ValidateSpaceID(spaceID); err != nil {
        return "", err
    }
    return r.spacePrefix + strings.ReplaceAll(spaceID, "-", "_"), nil
}

func (r *MySQLDatabaseRouter) getPool(spaceID string) (*sql.DB, error) {
    dbName, err := r.dbName(spaceID)
    if err != nil {
        return nil, err
    }

    r.mu.Lock()
    defer r.mu.Unlock()

    // LRU 命中
    if db, ok := r.pools[spaceID]; ok {
        r.lruList.MoveToFront(r.lruMap[spaceID])
        return db, nil
    }

    // 超过上限时驱逐最久未使用的池
    if len(r.pools) >= r.maxPools {
        oldest := r.lruList.Back()
        if oldest != nil {
            oldID := oldest.Value.(string)
            if oldDB, ok := r.pools[oldID]; ok {
                _ = oldDB.Close()
                delete(r.pools, oldID)
            }
            r.lruList.Remove(oldest)
            delete(r.lruMap, oldID)
        }
    }

    // 创建新连接池
    dsn := r.baseDSN + dbName + "?parseTime=true&multiStatements=true"
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("open space db %s: %w", dbName, err)
    }
    db.SetMaxIdleConns(r.maxIdle)
    db.SetMaxOpenConns(r.maxOpen)

    r.pools[spaceID] = db
    elem := r.lruList.PushFront(spaceID)
    r.lruMap[spaceID] = elem

    return db, nil
}

func (r *MySQLDatabaseRouter) WithSpace(ctx context.Context, spaceID string, fn func(ctx context.Context) error) error {
    db, err := r.getPool(spaceID)
    if err != nil {
        return err
    }
    return data.ExecInTx(ctx, db, nil, "", fn)
}

func (r *MySQLDatabaseRouter) WithSpaceRead(ctx context.Context, spaceID string, fn func(ctx context.Context) error) error {
    db, err := r.getPool(spaceID)
    if err != nil {
        return err
    }
    return data.ExecInTx(ctx, db, &sql.TxOptions{ReadOnly: true}, "", fn)
}

// CreateDatabase 为新 Space 创建 database。
func (r *MySQLDatabaseRouter) CreateDatabase(ctx context.Context, spaceID string) error {
    dbName, err := r.dbName(spaceID)
    if err != nil {
        return err
    }
    globalDSN := r.baseDSN + "?multiStatements=true"
    db, err := sql.Open("mysql", globalDSN)
    if err != nil {
        return err
    }
    defer db.Close()
    _, err = db.ExecContext(ctx,
        fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName))
    return err
}

// Close 关闭所有连接池。
func (r *MySQLDatabaseRouter) Close() {
    r.mu.Lock()
    defer r.mu.Unlock()
    for _, db := range r.pools {
        _ = db.Close()
    }
    r.pools = make(map[string]*sql.DB)
    r.lruList.Init()
    r.lruMap = make(map[string]*list.Element)
}
```

### 6.3 Ent ORM 接入模式

> Ent Client 的创建和 schema 切换由应用层负责，ncore 不感知 Ent。

**方案：Ent 使用原生 `*sql.Tx` 驱动，复用 SpaceRouter 设置好的事务**

```go
// internal/space/ent_factory.go（应用层）

package space

import (
    "database/sql"
    "sync"

    entsql "entgo.io/ent/dialect/sql"
    "entgo.io/ent/dialect"
    "<project>/ent"
)

// EntClientFactory 按 Space schema 缓存 Ent Client。
// Ent Client 本身不绑定 schema；schema 切换由 SpaceRouter 在事务层完成。
// PostgreSQL 模式：所有 Space 共用同一个 Ent Client，search_path 由 WithSpaceTx 设置。
// MySQL 模式：每个 Space 对应不同的 *sql.DB，需要不同 Ent Client。
type EntClientFactory struct {
    masterDB *sql.DB
    slaveDB  *sql.DB
    master   *ent.Client
    slave    *ent.Client
    spaceMap map[string]*ent.Client // 仅 MySQL 模式使用
    mu       sync.RWMutex
}

func NewEntClientFactory(masterDB, slaveDB *sql.DB) *EntClientFactory {
    masterDrv := entsql.OpenDB(dialect.Postgres, masterDB)
    slaveDrv := entsql.OpenDB(dialect.Postgres, slaveDB)
    return &EntClientFactory{
        masterDB: masterDB,
        slaveDB:  slaveDB,
        master:   ent.NewClient(ent.Driver(masterDrv)),
        slave:    ent.NewClient(ent.Driver(slaveDrv)),
        spaceMap: make(map[string]*ent.Client),
    }
}

// GetMasterClient 返回主库 Ent Client（PostgreSQL 模式）。
// search_path 由 WithSpaceTx 设置，Ent 复用事务中的连接，自动在 Space schema 内操作。
func (f *EntClientFactory) GetMasterClient() *ent.Client {
    return f.master
}

// GetSlaveClient 返回从库 Ent Client（PostgreSQL 模式）。
func (f *EntClientFactory) GetSlaveClient() *ent.Client {
    return f.slave
}

// GetSpaceClient 返回 Space 专属 Ent Client（MySQL database-per-space 模式）。
func (f *EntClientFactory) GetSpaceClient(db *sql.DB) *ent.Client {
    drv := entsql.OpenDB(dialect.MySQL, db)
    return ent.NewClient(ent.Driver(drv))
}
```

**Repository 中的使用模式（PostgreSQL schema 模式）：**

```go
// core/space/data/repository/setting.go

func (r *settingRepo) Create(ctx context.Context, body *structs.CreateSetting) (*structs.ReadSetting, error) {
    // PostgreSQL 模式：Ent Client 复用 ctx 中的 *sql.Tx（由 WithSpaceTx 设置）
    // Ent 会自动在当前事务中执行，而该事务已通过 SET LOCAL search_path 切换到 Space schema
    ec := r.data.GetMasterEntClient()

    row, err := ec.Setting.Create().
        SetSpaceID(body.SpaceID).
        SetKey(body.Key).
        SetValue(body.Value).
        Save(ctx)
    if err != nil {
        return nil, err
    }
    return toStruct(row), nil
}
```

**Service 层调用模式：**

```go
func (s *settingService) Create(ctx context.Context, body *structs.CreateSetting) (*structs.ReadSetting, error) {
    var result *structs.ReadSetting
    err := s.data.WithSpaceTx(ctx, func(ctx context.Context) error {
        var err error
        result, err = s.repo.Create(ctx, body)
        return err
    })
    return result, err
}
```

### 6.4 Middleware 层（注入 space_id 到 context）

```go
// internal/middleware/space.go（应用层）

func SpaceMiddleware(spaceService SpaceServiceInterface) gin.HandlerFunc {
    return func(c *gin.Context) {
        spaceID := c.GetHeader("X-Space-ID")

        // 验证 space_id 属于当前用户（已有逻辑）
        if spaceID != "" {
            userID := ctxutil.GetUserID(c.Request.Context())
            if valid, _ := spaceService.IsSpaceInUser(c.Request.Context(), spaceID, userID); !valid {
                spaceID = ""
            }
        }

        if spaceID != "" {
            // 将验证通过的 spaceID 写入 context，供 WithSpaceTx 使用
            ctx := data.SetSpaceID(c.Request.Context(), spaceID)
            c.Request = c.Request.WithContext(ctx)
        }

        c.Next()
    }
}
```

### 6.5 Space 生命周期（创建/删除时）

```go
// core/space/service/space.go

func (s *spaceService) Create(ctx context.Context, body *structs.CreateSpaceBody) (*structs.ReadSpace, error) {
    // Step 1: 在全局 schema 创建 Space 元数据（使用普通事务）
    space, err := s.spaceRepo.Create(ctx, body)
    if err != nil {
        return nil, err
    }

    // Step 2: 创建 Space 专属 schema/database
    if err := s.spaceProvisioner.CreateSchema(ctx, space.Slug); err != nil {
        _ = s.spaceRepo.Delete(ctx, space.ID)
        return nil, fmt.Errorf("provision space schema: %w", err)
    }

    // Step 3: 在 Space schema 内运行迁移（Ent Schema.Create）
    if err := s.spaceMigrator.MigrateSchema(ctx, space.Slug); err != nil {
        _ = s.spaceProvisioner.DropSchema(ctx, space.Slug)
        _ = s.spaceRepo.Delete(ctx, space.ID)
        return nil, fmt.Errorf("migrate space schema: %w", err)
    }

    return space, nil
}
```

---

## 七、数据分层规划

### 7.1 全局表（保留在 public/global schema）

这些表跨 Space 共享，不进入 Space schema：

```
users               — 用户身份，全局唯一
auth_tokens         — JWT/Session token
oauth_accounts      — OAuth 绑定
spaces              — Space 元数据（slug、name、状态）
access_roles        — 全局 RBAC 角色定义
access_permissions  — 全局权限定义
system_options      — 全局系统配置
```

### 7.2 Space 表（迁移到 Space schema）

```
space_settings      — Space 配置项
space_menus         — Space 自定义菜单
space_options       — Space 特性开关
space_quotas        — Space 资源配额
space_billing       — 计费/订阅数据
user_space_roles    — 用户在 Space 内的角色
content_topics      — 内容（CMS）
content_media       — 媒体资源
notifications       — 通知消息
payment_orders      — 支付订单
```

---

## 八、迁移路径（渐进式，不停机）

```
阶段 0（当前）
  └── 纯应用层隔离，space_id 外键，共享 public schema

阶段 1：ncore 扩展（无破坏性变更）
  ├── 新增 data/space.go：SpaceRouter 接口和工具函数
  ├── Data 结构体增加 spaceRouter 字段（nil = 不启用）
  └── 所有现有调用路径不变

阶段 2：应用层接入（先非核心模块）
  ├── 实现 PostgresSchemaRouter / MySQLDatabaseRouter
  ├── Middleware 注入 space_id 到 context
  ├── content、payment、resource 等模块 repository 改用 WithSpaceTx
  └── 新建 Space 时自动创建 schema 并运行迁移

阶段 3：全量切换
  ├── 历史数据迁移脚本（按 space_id 分发到对应 schema）
  ├── 所有 Space 业务表完成 schema 隔离
  └── 验证隔离效果

阶段 4：增强防护（可选）
  └── 在 Space schema 上添加 PostgreSQL RLS 策略（双重保险）
```

### 8.1 历史数据迁移脚本模板

```sql
-- PostgreSQL：将现有 space_settings 按 space_id 分发到各 Space schema
DO $$
DECLARE
    space_rec RECORD;
BEGIN
    FOR space_rec IN
        SELECT s.id, s.slug FROM public.spaces s WHERE s.slug IS NOT NULL
    LOOP
        -- 创建 Space schema
        EXECUTE format('CREATE SCHEMA IF NOT EXISTS space_%I', space_rec.slug);

        -- 迁移数据（以 space_settings 为例）
        EXECUTE format(
            'INSERT INTO space_%I.space_settings
             SELECT * FROM public.space_settings
             WHERE space_id = %L
             ON CONFLICT DO NOTHING',
            space_rec.slug, space_rec.id
        );

        RAISE NOTICE 'Migrated space: %', space_rec.slug;
    END LOOP;
END $$;
```

---

## 九、安全考量

### 9.1 SQL 注入防护

schema/db 名不支持参数化查询（`?` 占位符），必须拼接。防护措施：

1. `ValidateSpaceID` 白名单正则：`^[a-z0-9][a-z0-9_-]{0,62}$`
2. 所有构建 schema 名的代码必须先调用此函数
3. 禁止在 HTTP 层将用户输入直接作为 space_id，必须经过 Space 查询验证

### 9.2 连接池污染防护

核心原则：**Space 上下文必须绑定到事务（`SET LOCAL`），不能绑定到连接（`SET`）**

```sql
-- ✅ 安全：事务级，事务结束连接归还时自动重置
BEGIN;
SET LOCAL search_path TO space_acme, public;
...
COMMIT; -- 连接归还池后 search_path 恢复默认

-- ❌ 危险：会话级，连接归还池后下一个请求可能继承错误的 search_path
SET search_path TO space_acme, public;
```

`ExecInTx` 工具函数通过在事务内执行 `SET LOCAL` 保证此原则。

### 9.3 跨 Space 查询防护

- Admin 统计等需要跨 Space 查询的操作，应在全局 schema 上执行，不经过 SpaceRouter
- 全局 schema 的访问需要额外的权限校验（admin 角色）

---

## 十、各项目接入决策

| 项目         | 接入方式         | 说明                                                        |
| ------------ | ---------------- | ----------------------------------------------------------- |
| **gainmats** | 不接入           | 无多租户需求，保持现状，零额外开销                          |
| **yonfile**  | 可选接入（轻量） | 如需强隔离可注入 PostgresSchemaRouter；否则保持 Header 级别 |
| **datidy**   | 完整接入         | 有 Space 模块，参考 ncobase 方案；多 DB 后端需额外适配      |
| **ncobase**  | 完整接入         | 优先级最高，Space = 租户，全量迁移到 schema-per-space       |

---

## 十一、接口变更汇总

### ncore/data 新增（不破坏现有 API）

```go
// data/space.go（全新文件，仅标准库依赖）
const ContextKeySpaceID                               // context key
func SetSpaceID(ctx, id) context.Context              // 写入 space_id
func GetSpaceID(ctx) (string, bool)                   // 读取 space_id
func ValidateSpaceID(id string) error                 // 安全校验
func ExecInTx(ctx, db, opts, setupSQL, fn) error      // 事务工具
type SpaceRouter interface { ... }                     // 核心抽象接口
func WithSpaceRouter(router) Option                    // 注入 router
func (d *Data) WithSpaceTx(ctx, fn) error             // Space 事务（读写）
func (d *Data) WithSpaceTxRead(ctx, fn) error         // Space 事务（只读）

// data/data.go 修改
type Data struct { spaceRouter SpaceRouter; ... }      // 增加一个字段

// data/config/database.go 修改
type Space struct { ... }                     // 可选配置项
type Config struct { *Space; ... }            // 增加一个字段
```

### 现有 API（完全不变）

```go
func (d *Data) WithTx(ctx, fn) error
func (d *Data) WithTxRead(ctx, fn) error
func (d *Data) GetMasterDB() *sql.DB
func (d *Data) GetSlaveDB() (*sql.DB, error)
func (d *Data) GetDBManager() *connection.DBManager
// ... 其余所有方法不变
```

---

## 十二、参考资料

- [Building scalable multi-tenant applications in Go (GopherCon Israel 2025)](https://atlasgo.io/blog/2025/05/26/gophercon-scalable-multi-tenant-apps-in-go)
- [Multi-Tenancy Database Patterns with examples in Go (2025)](https://www.glukhov.org/post/2025/11/multitenant-database-patterns/)
- [Why Tenant Context Must Be Scoped Per Transaction (2026)](https://dev.to/m_zinger_2fc60eb3f3897908/why-tenant-context-must-be-scoped-per-transaction-3aop)
- [How to Design Multi-Tenant Schemas in PostgreSQL (2026)](https://oneuptime.com/blog/post/2026-01-25-multi-tenant-schemas-postgresql/view)
- [pgxpool and multi-tenancy discussion (jackc/pgx #2384)](https://github.com/jackc/pgx/discussions/2384)
- [Ent ORM Privacy/Tenant example](https://pkg.go.dev/entgo.io/ent/examples/privacytenant/ent)
