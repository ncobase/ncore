# NCore 多模块架构说明

## 架构设计

NCore 采用多模块（multi-module）架构设计，每个子包都是独立的 Go 模块。这样设计的好处：

1. **减少依赖**：使用方只需引入需要的子模块，不会引入不必要的依赖
2. **降低构建大小**：避免将整个 NCore 的所有依赖都打包进应用
3. **独立版本管理**：每个模块可以独立升级，互不影响
4. **更清晰的模块边界**：强制模块之间的依赖关系清晰

## 模块列表

```text
github.com/ncobase/ncore/
├── concurrency    - 并发工具
├── config         - 配置管理
├── consts         - 常量定义
├── ctxutil        - Context 工具
├── data           - 数据层（数据库、缓存、搜索引擎等）
├── ecode          - 错误码
├── extension      - 扩展和插件系统
├── logging        - 日志
├── messaging      - 消息队列
├── net            - 网络工具
├── security       - 安全相关
├── types          - 通用类型
├── utils          - 工具函数
├── validation     - 数据验证
└── version        - 版本管理
```

## 使用方式

### 在应用中使用

在你的应用 `go.mod` 中，只引入需要的模块：

```go
require (
    github.com/ncobase/ncore/config v0.0.0-20251022025300-781956ac0776
    github.com/ncobase/ncore/data v0.0.0-20251022025300-781956ac0776
    github.com/ncobase/ncore/logging v0.0.0-20251022025300-781956ac0776
    // 只引入需要的模块
)
```

### 本地开发

#### 1. 使用 go.work（推荐）

项目根目录提供了 `go.work` 文件，方便本地开发和测试：

```bash
# 在 ncore 根目录
go work sync  # 同步所有模块依赖
bash scripts/test.sh # 测试所有模块
```

#### 2. 在应用中使用本地 ncore

在应用的 `go.mod` 中使用 `replace` 指令：

```go
replace (
    github.com/ncobase/ncore/data => /path/to/ncore/data
    github.com/ncobase/ncore/config => /path/to/ncore/config
    // 替换你需要本地调试的模块
)
```

## 发布流程

### 发布单个模块

```bash
cd data
git tag data/v0.1.0
git push origin data/v0.1.0
```

### 批量发布

```bash
# 为所有模块打上相同版本的 tag
./scripts/tag.sh v0.1.0
```

## 模块依赖原则

1. **最小化依赖**：每个模块只引入必需的依赖
2. **避免循环依赖**：模块之间不能循环依赖
3. **通用模块优先**：`types`、`consts`、`ecode` 等通用模块应该零依赖或最少依赖
4. **大型依赖隔离**：如 `data` 模块包含数据库、搜索引擎等大型依赖，独立成模块

## FAQ

### Q: 为什么根目录没有 go.mod？

A: 因为每个子包都是独立模块，根目录不需要 go.mod。go.work 文件已经足够管理本地开发。

### Q: go.work 需要提交到 git 吗？

A: 可以提交。go.work 方便团队成员本地开发，但 `go.work.sum` 不应该提交（已在 .gitignore 中）。

### Q: 如何添加新模块？

A:

1. 创建新目录
2. 在目录中运行 `go mod init github.com/ncobase/ncore/新模块名`
3. 在根目录 `go.work` 中添加 `./新模块名`

### Q: 模块之间如何引用？

A: 直接使用完整的模块路径：

```go
import (
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/logging"
)
```

在本地开发时，go.work 会自动解析这些引用。

## 依赖注入支持 (Google Wire)

NCore 模块提供原生的 [Google Wire](https://github.com/google/wire) 支持，每个模块都暴露了 `ProviderSet` 用于依赖注入。

### 支持的模块

| 模块                | ProviderSet               | 提供内容            | 清理函数 |
| ------------------- | ------------------------- | ------------------- | -------- |
| `config`            | `config.ProviderSet`      | `*Config` 及子配置  | 否       |
| `logging/logger`    | `logger.ProviderSet`      | `*Logger`           | 是       |
| `data`              | `data.ProviderSet`        | `*Data`             | 是       |
| `extension/manager` | `manager.ProviderSet`     | `*Manager`          | 是       |
| `security`          | `security.ProviderSet`    | JWT `*TokenManager` | 否       |
| `messaging`         | `messaging.ProviderSet`   | Email `Sender`      | 否       |
| `concurrency`       | `concurrency.ProviderSet` | Worker `*Pool`      | 是       |

### Wire 使用示例

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

// InitializeApp 初始化应用程序
func InitializeApp() (*App, func(), error) {
    panic(wire.Build(
        // 配置管理
        config.ProviderSet,

        // 核心组件
        logger.ProviderSet,
        data.ProviderSet,
        manager.ProviderSet,

        // 应用构造函数
        NewApp,
    ))
}
```

### 特性

1. **清理函数支持**: `data`、`logger`、`manager` 和 `concurrency` 模块的 Provider 返回清理函数
2. **配置提取**: `config.ProviderSet` 自动提取各模块所需的子配置
3. **接口绑定**: `security` 模块使用 `wire.Bind` 进行接口绑定
4. **错误处理**: 所有 Provider 正确处理和传播错误

详细文档请参见：

- [English README](README.md#dependency-injection-google-wire)
- [中文 README](README_zh-CN.md#依赖注入-google-wire)
- [DEVELOPMENT.md](DEVELOPMENT.md#6-dependency-injection-google-wire)
- [示例代码](examples/09-wire)

## 相关资源

- [Go Modules 官方文档](https://go.dev/doc/modules/managing-dependencies)
- [Go Workspaces 官方文档](https://go.dev/doc/tutorial/workspaces)
- [Multi-module repositories 最佳实践](https://github.com/golang/go/wiki/Modules#faqs--multi-module-repositories)
- [Google Wire 官方文档](https://github.com/google/wire)
