# ncore/loger

基于 [logrus](https://github.com/sirupsen/logrus) 构建的强大日志系统，支持多种输出目标和搜索引擎集成（Elasticsearch、OpenSearch 和 Meilisearch）。

## 功能特点

- 多级别日志（Trace、Debug、Info、Warn、Error、Fatal、Panic）
- 结构化 JSON 日志
- 上下文感知的追踪
- 文件轮换
- 多输出目标（控制台、文件、搜索引擎）
- 搜索引擎集成：
  - Elasticsearch
  - OpenSearch
  - Meilisearch

## 使用方法

### 初始化

```go
import (
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/logging/logger"
)

// 创建配置
loggerConfig := &config.Logger{
    Level:    4, // Info 级别
    Format:   "json",
    Output:   "stdout",
    IndexName: "application-logs",
}

// 初始化日志器
cleanup, err := logger.New(loggerConfig)
if err != nil {
    panic(err)
}
defer cleanup()

// 设置应用版本（可选）
logger.SetVersion("1.0.0")
```

### 基本日志记录

```go
import (
    "context"
    "github.com/ncobase/ncore/logging/logger"
    "github.com/sirupsen/logrus"
)

ctx := context.Background()

// 基本日志
logger.Debug(ctx, "调试信息")
logger.Info(ctx, "信息消息")
logger.Warn(ctx, "警告消息")
logger.Error(ctx, "错误消息")

// 格式化日志
logger.Infof(ctx, "用户 %s 以角色 %s 登录", "john", "admin")

// 带字段的日志
logger.WithFields(ctx, logrus.Fields{
    "user_id": "12345",
    "action":  "login",
    "ip":      "192.168.1.1",
}).Info("用户登录成功")
```

### 配置文件输出

```go
loggerConfig := &config.Logger{
    Level:      4,
    Format:     "json",
    Output:     "file",
    OutputFile: "./logs/app.log",
}
```

### 配置 Elasticsearch

```go
loggerConfig := &config.Logger{
    Level:      4,
    Format:     "json",
    IndexName:  "application-logs",
    Elasticsearch: struct {
        Addresses []string
        Username  string
        Password  string
    }{
        Addresses: []string{"http://elasticsearch:9200"},
        Username:  "elastic",
        Password:  "password",
    },
}
```

### 配置 OpenSearch

```go
loggerConfig := &config.Logger{
    Level:      4,
    Format:     "json",
    IndexName:  "application-logs",
    OpenSearch: struct {
        Addresses      []string
        Username       string
        Password       string
        InsecureSkipTLS bool
    }{
        Addresses:      []string{"https://opensearch:9200"},
        Username:       "admin",
        Password:       "admin",
        InsecureSkipTLS: true,
    },
}
```

### 配置 Meilisearch

```go
loggerConfig := &config.Logger{
    Level:      4,
    Format:     "json",
    IndexName:  "application-logs",
    Meilisearch: struct {
        Host   string
        APIKey string
    }{
        Host:   "http://meilisearch:7700",
        APIKey: "masterKey",
    },
}
```

### 请求追踪

```go
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    // 创建带有追踪 ID 的上下文
    ctx, traceID := logger.EnsureTraceID(r.Context())
    
    // 添加追踪 ID 到响应头
    w.Header().Set("X-Trace-ID", traceID)
    
    logger.Infof(ctx, "处理请求：%s %s", r.Method, r.URL.Path)
    
    // 请求处理逻辑
    
    logger.Infof(ctx, "请求完成：%s %s", r.Method, r.URL.Path)
}
```

### 错误处理

```go
func ProcessData(ctx context.Context, data []byte) error {
    if len(data) == 0 {
        logger.Warn(ctx, "接收到空数据")
        return nil
    }
    
    result, err := parseData(data)
    if err != nil {
        logger.WithFields(ctx, logrus.Fields{
            "error": err.Error(),
            "data_length": len(data),
        }).Error("数据解析失败")
        return err
    }
    
    logger.WithFields(ctx, logrus.Fields{
        "result_count": len(result),
    }).Info("数据处理成功")
    
    return nil
}
```

## 日志级别

- **Trace** (6): 极其详细的信息
- **Debug** (5): 详细的调试信息
- **Info** (4): 一般操作信息
- **Warn** (3): 警告，潜在的问题情况
- **Error** (2): 错误条件，操作失败
- **Fatal** (1): 严重错误导致应用终止
- **Panic** (0): 关键错误导致应用崩溃
