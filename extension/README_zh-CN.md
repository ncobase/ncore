# NCore 扩展系统

一个灵活且强大的扩展系统，提供动态加载、生命周期管理、依赖处理、服务间通信和企业级安全功能。

## 特性

- **动态加载**: 支持从文件或内置注册加载扩展
- **依赖管理**: 支持强弱依赖并自动解析依赖关系
- **服务发现**: 基于 Consul 的服务注册与发现
- **事件系统**: 统一的事件处理，支持内存和消息队列
- **gRPC 集成**: 可选的 gRPC 服务支持，用于分布式通信
- **熔断器**: 内置容错机制保护服务调用
- **热重载**: 运行时插件加载和卸载
- **跨服务调用**: 统一的本地和远程服务调用接口
- **安全沙盒**: 插件路径验证、签名检查、源信任验证
- **资源监控**: 内存和 CPU 使用限制、性能指标收集
- **插件配置**: 每个插件的个性化配置管理

## 基本使用

### 创建扩展

```go
package myext

import (
    "github.com/ncobase/ncore/extension/registry"
    "github.com/ncobase/ncore/extension/types"
)

type MyExtension struct {
    types.OptionalImpl
}

func init() {
    registry.RegisterToGroupWithWeakDeps(New(), "core", []string{"user"})
}

func (m *MyExtension) Name() string { return "my-extension" }
func (m *MyExtension) Version() string { return "1.0.0" }
func (m *MyExtension) Dependencies() []string { return []string{} }

func (m *MyExtension) Init(conf *config.Config, manager types.ManagerInterface) error {
    // 初始化扩展
    return nil
}

func (m *MyExtension) GetMetadata() types.Metadata {
    return types.Metadata{
        Name: m.Name(), Version: m.Version(),
        Description: "我的扩展", Type: "module", Group: "core",
    }
}
```

### 使用管理器

```go
func main() {
    mgr, err := manager.NewManager(config)
    if err != nil { panic(err) }
    defer mgr.Cleanup()

    if err := mgr.InitExtensions(); err != nil { panic(err) }
    
    ext, err := mgr.GetExtensionByName("my-extension")
    // 使用扩展...
}
```

## 扩展生命周期

扩展遵循结构化的初始化过程：

1. **注册** - 通过 `init()` 自动注册或使用 `manager.RegisterExtension()` 手动注册
2. **依赖解析** - 系统根据依赖关系计算初始化顺序
3. **预初始化** - `PreInit()` 方法进行早期设置
4. **初始化** - `Init(config, manager)` 方法提供完整上下文
5. **后初始化** - `PostInit()` 方法用于跨扩展通信
6. **运行时** - 扩展处于活跃状态并处理请求
7. **清理** - `PreCleanup()` 和 `Cleanup()` 方法进行资源清理

## 依赖管理

### 依赖类型

**强依赖** - 必需的，必须先初始化：

```go
func (m *MyExtension) Dependencies() []string {
    return []string{"required-module"}
}
```

**弱依赖** - 可选的，缺失时优雅降级：

```go
func (m *MyExtension) GetAllDependencies() []types.DependencyEntry {
    return []types.DependencyEntry{
        {Name: "required-module", Type: types.StrongDependency},
        {Name: "optional-module", Type: types.WeakDependency},
    }
}
```

### 依赖解析

系统自动：

- 检测并防止循环依赖
- 计算最优初始化顺序
- 优雅处理缺失的弱依赖
- 为解析失败提供详细错误信息

有弱依赖的扩展应处理缺失的服务：

```go
func (m *MyExtension) PostInit() error {
    if userService, err := m.manager.GetServiceByName("user"); err == nil {
        m.userService = userService
    } else {
        log.Warn("用户服务不可用，功能受限")
    }
    return nil
}
```

## 服务通信

### 服务调用策略

```go
// 默认本地优先策略
result, err := manager.CallService(ctx, "user-service", "GetUser", userID)

// 显式策略
result, err := manager.CallServiceWithOptions(ctx, "user-service", "GetUser", userID, 
    &types.CallOptions{
        Strategy: types.LocalFirst,  // LocalFirst, RemoteFirst, LocalOnly, RemoteOnly
        Timeout:  30 * time.Second,
    })
```

**策略行为**：

- `LocalFirst`: 尝试本地服务，回退到 gRPC
- `RemoteFirst`: 尝试 gRPC 服务，回退到本地
- `LocalOnly`: 仅本地服务，不可用时失败
- `RemoteOnly`: 仅 gRPC 服务，不可用时失败

### 跨服务访问

```go
// 直接服务访问
userService, err := manager.GetServiceByName("user")

// 跨服务字段访问
authService, err := manager.GetCrossService("auth", "TokenManager")
```

## 事件系统

### 事件目标

事件系统支持多种传输机制：

- `EventTargetMemory`: 内存，单实例，高性能
- `EventTargetQueue`: 消息队列 (RabbitMQ/Kafka)，分布式，持久化
- `EventTargetAll`: 同时使用两种目标

### 事件操作

```go
// 订阅（自动选择最佳传输）
manager.SubscribeEvent("user.created", func(data any) {
    eventData := data.(types.EventData)
    // 处理事件
})

// 订阅到特定传输
manager.SubscribeEvent("user.created", handler, types.EventTargetMemory)

// 发布（自动选择最佳传输）
manager.PublishEvent("user.created", userData)

// 关键事件带重试发布
manager.PublishEventWithRetry("payment.failed", paymentData, 3)
```

### 事件数据结构

```go
type EventData struct {
    Time      time.Time `json:"time"`
    Source    string    `json:"source"`
    EventType string    `json:"event_type"`
    Data      any       `json:"data"`
}
```

## 安全与性能功能

### 安全沙盒

系统提供全面的安全控制：

```go
// 插件配置管理
config := map[string]any{
    "cache_ttl": "1h",
    "max_connections": 100,
}
manager.SetPluginConfig("my-plugin", config)

// 获取插件配置
if cfg, exists := manager.GetPluginConfig("my-plugin"); exists {
    // 使用配置
}
```

### 资源监控

```go
// 获取资源使用指标
metrics := manager.GetResourceMetrics()
for pluginName, metric := range metrics {
    fmt.Printf("插件 %s: 内存=%fMB, CPU=%f%%, 加载时间=%v\n", 
        pluginName, metric.MemoryUsageMB, metric.CPUUsagePercent, metric.LoadTime)
}

// 获取安全状态
securityStatus := manager.GetSecurityStatus()
fmt.Printf("安全状态：%+v\n", securityStatus)
```

### 综合指标

```go
// 获取包含安全、性能、系统信息的完整指标
enhancedMetrics := manager.GetEnhancedMetrics()
```

## 配置

```yaml
extension:
  path: "./plugins"          # 插件目录
  mode: "file"              # "file" 或 "c2hlbgo"（内置）
  includes: ["auth", "user"] # 包含特定插件
  excludes: ["debug"]       # 排除插件
  hot_reload: true          # 热重载支持
  
  # 高级配置
  max_plugins: 50           # 最大插件数量
  
  # 安全配置
  security:
    enable_sandbox: true    # 启用安全沙盒
    allowed_paths:          # 允许的插件路径
      - "/opt/plugins"
      - "/usr/local/plugins"
    blocked_extensions:     # 阻止的文件扩展名
      - ".exe"
      - ".bat"
    trusted_sources:        # 信任的插件源
      - "company.com"
      - "verified.org"
    require_signature: true # 要求插件签名
  
  # 性能配置
  performance:
    max_memory_mb: 512      # 最大内存使用 (MB)
    max_cpu_percent: 80     # 最大 CPU 使用率 (%)
    enable_metrics: true    # 启用性能指标
    metrics_interval: "30s" # 指标收集间隔
    enable_profiling: false # 启用性能分析
    gc_interval: "5m"       # 垃圾回收间隔
  
  # 插件特定配置
  plugin_config:
    auth_plugin:
      oauth_providers: ["google", "github"]
    user_plugin:
      cache_ttl: "1h"

consul:
  address: "localhost:8500"  # Consul 服务器
  scheme: "http"
  discovery:
    health_check: true       # 启用健康检查
    check_interval: "10s"    # 健康检查间隔
    timeout: "3s"           # 健康检查超时

grpc:
  enabled: true             # 启用 gRPC 支持
  host: "localhost"
  port: 9090
```

## 高级特性

### gRPC 集成

扩展可以提供 gRPC 服务：

```go
func (m *MyExtension) RegisterGRPCServices(server *grpc.Server) {
    pb.RegisterMyServiceServer(server, m.grpcService)
}
```

### 服务发现

扩展可以注册到服务发现：

```go
func (m *MyExtension) NeedServiceDiscovery() bool { return true }

func (m *MyExtension) GetServiceInfo() *types.ServiceInfo {
    return &types.ServiceInfo{
        Address: "localhost:8080",
        Tags:    []string{"api", "v1"},
        Meta:    map[string]string{"version": "1.0"},
    }
}
```

### 熔断器

防护服务故障：

```go
result, err := manager.ExecuteWithCircuitBreaker("external-service", func() (any, error) {
    return callExternalAPI()
})
```

### 插件加载模式

**文件模式**: 从文件系统加载插件

- 支持 Linux 上的 `.so` 文件，Windows 上的 `.dll`
- 热重载能力
- 包含/排除过滤
- 安全沙盒保护

**内置模式**: 使用静态编译的扩展

- 更好的性能和安全性
- 无文件系统依赖
- 编译时依赖解析

## 管理 API

运行时管理的 REST 端点：

- `GET /exts` - 列出所有扩展及元数据
- `GET /exts/status` - 获取扩展状态和健康状况
- `POST /exts/load?name=plugin` - 加载特定插件
- `POST /exts/unload?name=plugin` - 卸载插件
- `POST /exts/reload?name=plugin` - 重载插件
- `GET /exts/metrics` - 系统指标和性能数据
- `GET /exts/metrics/security` - 安全状态指标
- `GET /exts/metrics/performance` - 性能监控指标

## 性能考量

- **服务发现**: 使用适当的缓存 TTL（默认 30 秒）
- **事件传输**: 高频率选择内存，可靠性选择队列
- **熔断器**: 监控失败率并调整阈值
- **插件加载**: 生产环境优选内置模式
- **资源监控**: 根据需要启用性能指标收集
- **安全检查**: 平衡安全性和性能需求

## 故障排除

### 常见问题

**循环依赖**

```
错误: cyclic dependency detected in extensions: [module-a, module-b]
```

*解决方案*: 将其中一个依赖转换为弱依赖类型

**服务未找到**

```
错误: extension 'user-service' not found
```  

*解决方案*: 检查扩展注册和初始化顺序

**gRPC 连接失败**

```
错误: failed to get gRPC connection for service-name  
```

*解决方案*: 验证服务发现配置和网络连接

**安全验证失败**

```
错误: security validation failed: path /tmp/plugin.so is not in allowed paths
```

*解决方案*: 检查安全配置中的允许路径设置

**资源限制超出**

```
错误: resource limit check failed: insufficient memory: would exceed limit of 512MB
```

*解决方案*: 调整性能配置中的资源限制或优化插件内存使用

**插件签名验证失败**

```
错误: signature validation failed: plugin signature not found
```

*解决方案*: 确保插件文件有对应的 .sig 签名文件或禁用签名验证
