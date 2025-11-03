# NCore 开发指南

## 快速开始

### 克隆项目

```bash
git clone https://github.com/ncobase/ncore.git
cd ncore
```

### 同步依赖

```bash
# 方式 1: 使用 Makefile（推荐）
make sync

# 方式 2: 直接使用 go 命令
go work sync
```

### 运行测试

```bash
# 方式 1: 使用 Makefile（推荐）
make test              # 运行所有测试
make test-v            # 详细输出
make test-cover        # 带覆盖率

# 方式 2: 使用脚本
./scripts/test.sh
./scripts/test.sh -v
./scripts/test.sh -cover

# 测试特定模块
cd data
go test ./...
```

### 添加新模块

```bash
# 1. 创建模块目录
mkdir newmodule

# 2. 初始化模块
cd newmodule
go mod init github.com/ncobase/ncore/newmodule

# 3. 添加到 workspace
cd ..
echo " ./newmodule" >> go.work

# 4. 同步依赖
go work sync
```

## 模块开发规范

### 1. 模块命名

- 使用小写字母
- 多个单词用下划线或直接连接
- 示例：`ctxutil`, `data`, `messaging`

### 2. 版本管理

每个模块独立版本：

```bash
# 发布单个模块
cd data
git tag data/v0.1.0
git push origin data/v0.1.0

# 批量发布所有模块
./scripts/tag.sh v0.1.0
git push origin --tags
```

### 3. 依赖管理

#### 添加依赖

```bash
cd <module-name>
go get <dependency>
go mod tidy
```

#### 更新依赖

```bash
# 方式 1: 更新所有模块的所有依赖（推荐）
make update            # 升级所有模块的依赖
make sync              # 同步 workspace

# 方式 2: 使用脚本
./scripts/update-deps.sh           # 更新所有模块
./scripts/update-deps.sh data      # 只更新 data 模块

# 方式 3: 手动更新特定模块
cd <module-name>
go get -u ./...        # 升级所有依赖到最新 minor/patch
go get -u <dependency> # 升级特定依赖
go mod tidy

# 检查过期依赖
make check-outdated
# 或
./scripts/check-outdated.sh
```

**重要提示**:

- ⚠️ 由于根目录没有 go.mod，**不能**在根目录直接运行 `go get -u ./...`
- ✅ 必须使用 `make update` 或脚本来更新所有模块
- ✅ 或者进入单个模块目录手动更新

#### 模块间依赖

```go
// 在 go.mod 中
require (
    github.com/ncobase/ncore/types v0.0.0-20251022025300-781956ac0776
)

// 在代码中
import "github.com/ncobase/ncore/types"
```

### 4. 测试

每个模块都应该有充分的测试：

```bash
# 运行测试
go test ./...

# 带覆盖率
go test -cover ./...

# 详细输出
go test -v ./...
```

### 5. 代码格式

```bash
# 格式化代码
go fmt ./...

# 运行 linter（如果配置了）
golangci-lint run
```

## 与应用集成

### 方式 1：使用发布版本

在应用的 `go.mod` 中：

```go
require (
    github.com/ncobase/ncore/data v0.1.0
    github.com/ncobase/ncore/config v0.1.0
)
```

### 方式 2：使用 replace 进行本地开发

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

### 方式 3：使用 workspace（推荐用于开发）

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

### 开发环境设置

```bash
# 克隆项目
git clone https://github.com/ncobase/ncore.git
cd ncore

# 同步依赖
make sync
```

### 依赖管理

#### ❌ 错误做法

```bash
# 这不会工作！根目录没有 go.mod
go get -u ./...
```

#### ✅ 正确做法

```bash
# 升级所有模块的所有依赖
make update
make sync

# 升级特定模块
./scripts/update-deps.sh data

# 检查哪些依赖过期了
make check-outdated

# 手动升级单个模块
cd data
go get -u ./...
go mod tidy
cd ..
```

### 测试

```bash
# 运行所有测试
make test

# 详细输出
make test-v

# 带覆盖率
make test-cover

# 测试单个模块
cd data
go test -v ./...
```

### 代码质量

```bash
# 格式化代码
make fmt

# 运行 linter（需要先安装 golangci-lint）
make lint

# 清理构建产物
make clean
```

### 版本发布

```bash
# 为所有模块打标签
make tag VERSION=v0.1.0

# 推送标签
git push origin --tags

# 只为单个模块打标签
cd data
git tag data/v0.1.0
git push origin data/v0.1.0
```

## 可用脚本

### `scripts/update-deps.sh`

升级依赖的脚本

```bash
# 升级所有模块
./scripts/update-deps.sh

# 只升级特定模块
./scripts/update-deps.sh data
```

### `scripts/check-outdated.sh`

检查过期依赖

```bash
./scripts/check-outdated.sh
```

### `scripts/test.sh`

运行所有测试

```bash
# 基本测试
./scripts/test.sh

# 详细输出
./scripts/test.sh -v

# 带覆盖率
./scripts/test.sh -cover
```

### `scripts/tag.sh`

批量打标签

```bash
./scripts/tag.sh v0.1.0
```

## Makefile 目标

| 命令 | 说明 |
|------|------|
| `make help` | 显示帮助信息 |
| `make sync` | 同步 workspace 依赖 |
| `make test` | 运行所有测试 |
| `make test-v` | 运行所有测试（详细） |
| `make test-cover` | 运行测试带覆盖率 |
| `make update` | 更新所有依赖 |
| `make check-outdated` | 检查过期依赖 |
| `make tag VERSION=v0.1.0` | 打标签 |
| `make fmt` | 格式化代码 |
| `make lint` | 运行 linter |
| `make clean` | 清理构建产物 |

## 常见问题

### Q: go work sync 报错怎么办？

A: 尝试以下步骤：

```bash
# 清理模块缓存
go clean -modcache

# 重新同步
go work sync

# 如果还有问题，单独更新每个模块
cd <module-name>
go mod tidy
```

### Q: 如何查看模块依赖关系？

```bash
cd <module-name>
go mod graph
```

### Q: 如何升级所有模块的依赖？

```bash
# 创建脚本或手动执行
for dir in */; do
    if [ -f "$dir/go.mod" ]; then
        echo "Updating $dir"
        cd "$dir"
        go get -u ./...
        go mod tidy
        cd ..
    fi
done
```

### Q: 模块之间有循环依赖怎么办？

A: 重新设计模块结构，可能的解决方案：

1. 将共享代码提取到新的通用模块（如 `types`）
2. 使用接口而不是具体实现
3. 调整模块职责划分

## CI/CD 建议

### GitHub Actions 示例

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Sync workspace
        run: go work sync

      - name: Run tests
        run: bash scripts/test.sh

      - name: Test each module
        run: |
          for dir in */; do
            if [ -f "$dir/go.mod" ]; then
              echo "Testing $dir"
              cd "$dir"
              go test -v ./...
              cd ..
            fi
          done
```

## 性能优化建议

1. **最小化依赖**：每个模块只引入必要的依赖
2. **延迟加载**：大型依赖（如数据库驱动）放在独立模块
3. **接口优先**：模块间通过接口交互，减少耦合
4. **文档完善**：清晰的模块职责和 API 文档

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 工具推荐

- **golangci-lint**: 代码检查
- **go-mod-outdated**: 检查过期依赖
- **go-mod-upgrade**: 批量升级依赖
- **air**: 热重载（开发时）

## 工作流示例

### 添加新功能

```bash
# 1. 同步依赖
make sync

# 2. 开发功能
cd data
# ... 编写代码 ...

# 3. 如果需要新依赖
go get github.com/some/package
go mod tidy

# 4. 运行测试
go test ./...

# 5. 返回根目录，测试所有模块
cd ..
make test

# 6. 格式化代码
make fmt

# 7. 提交
git add .
git commit -m "Add new feature"
```

### 修复 Bug

```bash
# 1. 定位 bug 所在模块
cd <module>

# 2. 修复代码

# 3. 运行测试
go test ./...

# 4. 返回根目录
cd ..

# 5. 运行所有测试
make test

# 6. 如果是重要修复，发布 patch 版本
make tag VERSION=v0.1.1
git push origin --tags
```

### 升级依赖

```bash
# 1. 检查过期依赖
make check-outdated

# 2. 升级依赖
make update

# 3. 同步 workspace
make sync

# 4. 运行测试确保一切正常
make test

# 5. 提交更改
git add .
git commit -m "Update dependencies"
```

## 技巧

### 只测试特定包

```bash
cd data/databases
go test -v .
```

### 带竞态检测的测试

```bash
cd <module>
go test -race ./...
```

### 查看测试覆盖率报告

```bash
cd <module>
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 清理模块缓存

```bash
go clean -modcache
go work sync
```
