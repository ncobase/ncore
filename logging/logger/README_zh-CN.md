# ncore/logger

基于 [logrus](https://github.com/sirupsen/logrus) 构建的强大日志系统，支持多输出目标、搜索引擎集成和数据脱敏功能。

## 功能特点

- 多级别结构化 JSON 日志
- 上下文感知追踪，自动 Trace ID 传播
- **数据脱敏，支持深层结构处理**
- 多输出目标：控制台、文件（自动轮换）、Elasticsearch、OpenSearch、Meilisearch
- 固定长度脱敏，防止敏感数据长度泄露

## 快速开始

```go
import (
    "context"
    "github.com/ncobase/ncore/logging/logger"
    "github.com/ncobase/ncore/logging/logger/config"
)

// 基本设置
cleanup, err := logger.New(&config.Config{
    Level:  4, // Info 级别
    Format: "json",
    Output: "stdout",
})
if err != nil {
    panic(err)
}
defer cleanup()

// 日志记录
ctx := context.Background()
logger.Info(ctx, "应用程序启动")
logger.WithFields(ctx, logrus.Fields{
    "user_id": "123",
    "action":  "login",
}).Info("用户登录")
```

## 数据脱敏

自动保护日志中的敏感数据，使用固定长度脱敏：

```go
// 安全配置（推荐）
&config.Config{
    Level:  4,
    Format: "json",
    Output: "file",
    OutputFile: "./logs/app.log",
    Desensitization: &config.Desensitization{
        Enabled:         true,
        UseFixedLength:  true,  // 所有敏感数据 → "********"
        FixedMaskLength: 8,
        MaskChar:        "*",
    },
}

// 使用 - 敏感字段自动脱敏
logger.WithFields(ctx, logrus.Fields{
    "username": "john",
    "password": "secret123",     // → "********"
    "email":    "john@test.com", // → "********"
    "token":    "eyJhbGci...",   // → "********"
}).Info("用户身份验证")
```

### 深层结构支持

自动处理嵌套对象、数组和映射：

```go
type User struct {
    Username string            `json:"username"`
    Password string            `json:"password"`
    Profile  map[string]string `json:"profile"`
    APIKeys  []string          `json:"api_keys"`
}

user := User{
    Username: "john",
    Password: "secret",
    Profile:  map[string]string{"email": "john@test.com"},
    APIKeys:  []string{"sk_test_123", "pk_live_456"},
}

// 所有嵌套敏感数据自动脱敏
logger.WithFields(ctx, logrus.Fields{
    "user": user, // 深层结构自动处理
}).Info("用户创建完成")
```

## 配置

### 文件输出

```go
&config.Config{
    Output:     "file",
    OutputFile: "./logs/app.log", // 每日轮换
}
```

### 搜索引擎

```go
// Elasticsearch
Elasticsearch: &config.Elasticsearch{
    Addresses: []string{"http://localhost:9200"},
    Username:  "elastic",
    Password:  "password",
}

// OpenSearch  
OpenSearch: &config.OpenSearch{
    Addresses: []string{"https://localhost:9200"},
    Username:  "admin",
    Password:  "admin",
}

// Meilisearch
Meilisearch: &config.Meilisearch{
    Host:   "http://localhost:7700",
    APIKey: "masterKey",
}
```

### 自定义脱敏

```go
Desensitization: &config.Desensitization{
    Enabled:         true,
    UseFixedLength:  true,
    FixedMaskLength: 8,
    SensitiveFields: []string{"password", "token", "secret", "api_key"},
    CustomPatterns:  []string{`\b\d{4}-\d{4}-\d{4}-\d{4}\b`}, // 信用卡号
}
```

## 请求追踪

```go
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    ctx, traceID := logger.EnsureTraceID(r.Context())
    w.Header().Set("X-Trace-ID", traceID)
    
    logger.Info(ctx, "请求开始")
    // 此上下文中的所有日志都包含相同的 trace ID
    processRequest(ctx)
    logger.Info(ctx, "请求完成")
}
```

## 生产环境配置

```yaml
logger:
  level: 4
  format: json
  output: file
  output_file: /var/log/app.log
  
  desensitization:
    enabled: true
    use_fixed_length: true
    fixed_mask_length: 8
    
  elasticsearch:
    addresses: ["http://es:9200"]
    username: elastic
    password: ${ES_PASSWORD}
```

## API 参考

```go
// 初始化
func New(c *config.Config) (func(), error)

// 日志记录
func Debug/Info/Warn/Error/Fatal/Panic(ctx context.Context, args ...any)
func Debugf/Infof/Warnf/Errorf/Fatalf/Panicf(ctx context.Context, format string, args ...any)
func WithFields(ctx context.Context, fields logrus.Fields) *logrus.Entry

// 追踪
func EnsureTraceID(ctx context.Context) (context.Context, string)
```

## 日志级别

| 级别 | 值 | 用途 |
|------|----|----|
| Trace | 6 | 详细调试 |
| Debug | 5 | 调试信息 |
| Info  | 4 | 一般信息 |
| Warn  | 3 | 警告 |
| Error | 2 | 错误 |
| Fatal | 1 | 严重错误 |
| Panic | 0 | 系统崩溃 |
