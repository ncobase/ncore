# NCore

一个全面的 Go 应用程序组件库，用于构建现代、可扩展的应用程序。

## 特性

- **模块化架构**：只导入您需要的模块
- **模块化驱动系统** (v0.2.0+)：可选的数据库、缓存、搜索和消息驱动，最小化二进制体积
- **丰富的集成**：PostgreSQL、MySQL、MongoDB、Redis、Elasticsearch、Kafka 等
- **安全与认证**：JWT、OAuth、加密工具
- **可观测性**：OpenTelemetry、日志记录和监控
- **依赖注入**：原生支持 Google Wire

## 多模块架构

NCore 采用**多模块架构**，每个子包都是独立的 Go 模块，提供最小依赖和独立版本管理。

### 可用模块

```text
github.com/ncobase/ncore/concurrency    - 并发工具
github.com/ncobase/ncore/config         - 配置管理
github.com/ncobase/ncore/consts         - 常量
github.com/ncobase/ncore/ctxutil        - Context 工具
github.com/ncobase/ncore/data           - 数据层（数据库、缓存、搜索）
github.com/ncobase/ncore/ecode          - 错误码
github.com/ncobase/ncore/extension      - 扩展系统
github.com/ncobase/ncore/logging        - 日志
github.com/ncobase/ncore/messaging      - 消息队列
github.com/ncobase/ncore/net            - 网络工具
github.com/ncobase/ncore/security       - 安全功能
github.com/ncobase/ncore/types          - 通用类型
github.com/ncobase/ncore/utils          - 工具函数
github.com/ncobase/ncore/validation     - 验证
github.com/ncobase/ncore/version        - 版本信息
```

## 安装

只导入您需要的模块：

```bash
# 核心模块
go get github.com/ncobase/ncore/config
go get github.com/ncobase/ncore/data
go get github.com/ncobase/ncore/security

# 数据驱动 (v0.2.0+) - 只导入您使用的驱动
go get github.com/ncobase/ncore/data/postgres
go get github.com/ncobase/ncore/data/redis
go get github.com/ncobase/ncore/data/meilisearch
```

## 快速开始

```go
package main

import (
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/data"

    // 只导入您需要的驱动 (v0.2.0+)
    _ "github.com/ncobase/ncore/data/postgres"
    _ "github.com/ncobase/ncore/data/redis"
)

func main() {
    // 加载配置
    cfg, err := config.LoadConfig("config.yaml")
    if err != nil {
        panic(err)
    }

    // 初始化数据层（驱动在导入时自动注册）
    d, cleanup, err := data.New(cfg.Data)
    if err != nil {
        panic(err)
    }
    defer cleanup()

    // 使用数据库和缓存
    db := d.Conn.DB()
    redis := d.Conn.RC
}
```

### 从 v0.1.x 迁移

如果您正在从 v0.1.x 升级，请在您的 main.go 或任何初始化文件中添加驱动导入：

```go
import (
    _ "github.com/ncobase/ncore/data/postgres"  // 添加您需要的驱动
    _ "github.com/ncobase/ncore/data/redis"
)
```

对于搜索功能，导入搜索模块：

```go
import (
    "github.com/ncobase/ncore/data/search"
    _ "github.com/ncobase/ncore/data/elasticsearch"
)
```

### 为什么使用模块化驱动？

v0.2.0 引入了可选驱动，显著减少了二进制体积和依赖：

| 指标          | v0.1.x | v0.2.0 | 改进       |
|-------------|--------|--------|----------|
| 二进制大小（基础应用） | ~92MB  | ~43MB  | **-53%** |
| 依赖数量        | 466    | ~100   | **-78%** |
| 编译时间        | ~45s   | ~20s   | **-56%** |

**可用驱动：**

- **数据库**: postgres, mysql, sqlite, mongodb, neo4j
- **缓存**: redis
- **搜索**: elasticsearch, opensearch, meilisearch（可选模块）
- **消息队列**: kafka, rabbitmq
- **存储**: local, s3, aliyun, minio

模块化驱动系统使用基于 init 的注册模式来实现自动驱动发现。

## 依赖注入 (Google Wire)

NCore 原生支持 [Google Wire](https://github.com/google/wire)。您可以使用每个模块中预定义的 `ProviderSet` 轻松组装您的应用程序。

### 可用的 ProviderSets

| 模块                  | ProviderSet               | 提供内容                                     |
|---------------------|---------------------------|------------------------------------------|
| `config`            | `config.ProviderSet`      | `*Config`, `*Logger`, `*Data`, `*Auth` 等 |
| `logging/logger`    | `logger.ProviderSet`      | `*Logger` 带清理函数                          |
| `data`              | `data.ProviderSet`        | `*Data` 带清理函数                            |
| `extension/manager` | `manager.ProviderSet`     | `*Manager` 带清理函数                         |
| `security`          | `security.ProviderSet`    | JWT `*TokenManager`                      |
| `messaging`         | `messaging.ProviderSet`   | 邮件 `Sender`                              |
| `concurrency`       | `concurrency.ProviderSet` | Worker `*Pool` 带清理函数                     |

### 基础用法

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
        // 引入 NCore 的核心 ProviderSet
        config.ProviderSet,
        logger.ProviderSet,
        data.ProviderSet,
        manager.ProviderSet,

        // 您自己的 Provider
        NewApp,
    ))
}
```

### 带安全模块和消息模块

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

## 开发

```bash
# 克隆仓库
git clone https://github.com/ncobase/ncore.git
cd ncore

# 同步依赖
go work sync

# 运行测试
bash scripts/test.sh
```

## 示例

详见 [examples/README.md](./examples/README.md) 以获取详细概览和学习路线。

## 文档

- [DEVELOPMENT_zh-CN.md](DEVELOPMENT_zh-CN.md) - 开发指南
- [MODULES_zh-CN.md](MODULES_zh-CN.md) - 多模块架构说明

## 代码生成

用于搭建新项目和组件，使用 CLI 工具：

```bash
go install github.com/ncobase/cli@latest
nco create core auth-service
nco create business payment --use-mongo --with-test
```

## 许可证

详见 [LICENSE](LICENSE) 文件。
