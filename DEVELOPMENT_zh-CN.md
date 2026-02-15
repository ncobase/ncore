# NCore 开发指南

## 快速开始

### 克隆项目

```bash
git clone https://github.com/ncobase/ncore.git
cd ncore
```

### 同步依赖

```bash
# 使用 Makefile（推荐）
make sync

# 或直接使用 go 命令
go work sync
```

### 运行测试

```bash
# 使用 Makefile（推荐）
make test              # 运行所有测试
make test-v            # 详细输出
make test-cover        # 带覆盖率

# 使用脚本
./scripts/test.sh
./scripts/test.sh -v
./scripts/test.sh -cover

# 测试特定模块
cd data
go test ./...
```

### 添加新模块

```bash
mkdir newmodule && cd newmodule
go mod init github.com/ncobase/ncore/newmodule
cd .. && echo " ./newmodule" >> go.work
go work sync
```

## 模块开发规范

### 模块命名与版本管理

- **命名**：小写，单词连接或下划线分隔（如 `ctxutil`、`data`）
- **版本**：使用 git tag 独立管理（如 `data/v0.1.0`）

```bash
./scripts/tag.sh v0.1.0  # 批量发布所有模块
```

### 依赖管理

```bash
# 更新所有模块
make update && make sync

# 更新特定模块
./scripts/update-deps.sh data

# 检查过期依赖
make check-outdated
```

**注意**：根目录无 go.mod - 使用 `make update` 或脚本，不要使用 `go get -u ./...`

### 测试与格式化

```bash
go test ./...              # 运行测试
go test -cover ./...       # 带覆盖率
go fmt ./...               # 格式化代码
golangci-lint run          # 代码检查（需安装）
```

### 依赖注入 (Google Wire)

模块提供 `ProviderSet` 用于 Wire 集成：

```go
//go:build wireinject

func InitializeApp() (*App, func(), error) {
    panic(wire.Build(
        config.ProviderSet,
        logger.ProviderSet,
        data.ProviderSet,
        NewApp,
    ))
}
```

生成 wire 代码：`wire ./...`

## 集成到应用

### 方式 1: 使用发布版本

在应用的 `go.mod` 中：

```go
require (
    github.com/ncobase/ncore/data v0.1.0
    github.com/ncobase/ncore/config v0.1.0
)
```

### 方式 2: replace 本地开发

在应用的 `go.mod` 中添加：

```go
replace (
    github.com/ncobase/ncore/data => /path/to/ncore/data
    github.com/ncobase/ncore/config => /path/to/ncore/config
)
```

然后：

```bash
cd <your-app>
go mod tidy
```

### 方式 3: workspace（推荐）

在应用目录创建 `go.work`：

```text
go 1.24

use (
    .
    /path/to/ncore/data
    /path/to/ncore/config
    // 添加需要的模块
)
```

## 常用命令

```bash
# 环境设置
git clone https://github.com/ncobase/ncore.git && cd ncore && make sync

# 测试
make test           # 所有测试
make test-v         # 详细输出
make test-cover     # 带覆盖率

# 依赖管理
make update && make sync        # 更新所有
./scripts/update-deps.sh data   # 更新特定模块
make check-outdated             # 检查过期

# 代码质量
make fmt            # 格式化
make lint           # 代码检查（需要 golangci-lint）
make clean          # 清理构建产物

# 版本管理
make tag VERSION=v0.1.0         # 标记所有模块
git push origin --tags          # 推送标签
```

## 可用脚本

```bash
./scripts/update-deps.sh [module]   # 更新依赖
./scripts/check-outdated.sh         # 检查过期依赖
./scripts/test.sh [-v] [--cover]    # 运行测试
./scripts/tag.sh v0.1.0             # 批量标记模块
```

## Makefile 目标

运行 `make help` 查看所有可用目标。

## 常见问题

**`go work sync` 报错：** 运行 `go clean -modcache && go work sync`

**查看依赖关系：** `cd <module> && go mod graph`

**循环依赖：** 提取共享代码到公共模块或使用接口

## CI/CD 示例

```yaml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with: { go-version: "1.24" }
      - run: go work sync && bash scripts/test.sh
```

## 贡献指南

1. Fork → 2. 创建功能分支 → 3. 提交更改 → 4. 推送 → 5. 创建 Pull Request

## 技巧

```bash
# 竞态检测
go test -race ./...

# 覆盖率报告
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# 清理缓存
go clean -modcache && go work sync
```
