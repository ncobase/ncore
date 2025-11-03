# NCore

一个全面的 Go 应用程序组件库，用于构建现代、可扩展的应用程序。

## 特性

- **模块化架构**：只导入您需要的模块
- **丰富的集成**：数据库、搜索、消息传递和存储解决方案
- **安全与认证**：JWT、OAuth、加密工具
- **可观测性**：OpenTelemetry、日志记录和监控

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
go get github.com/ncobase/ncore/config
go get github.com/ncobase/ncore/data
go get github.com/ncobase/ncore/security
```

## 快速开始

```go
package main

import (
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/logging"
)

func main() {
    // 加载配置
    cfg, err := config.LoadConfig("config.yaml")
    if err != nil {
        panic(err)
    }

    // 初始化日志记录器
    logger := logging.NewLogger(cfg.Logging)
    logger.Info("应用程序已启动")
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
