# OSS - 对象存储服务

[English](README.md) | 简体中文

统一的对象存储抽象层，支持多云存储提供商和本地文件系统。

## 特性

- **多云支持**：AWS S3、Azure Blob、阿里云 OSS、腾讯云 COS、Google Cloud Storage、MinIO、七牛 Kodo、Synology NAS 和本地文件系统
- **统一接口**：所有存储提供商实现相同的一致性 API
- **自动注册**：驱动通过 `init()` 函数自动注册
- **官方 SDK**：使用各云服务商的官方 SDK
- **本地存储**：内置本地文件系统存储支持
- **轻量级**：独立模块，依赖最小化

## 安装

```bash
go get github.com/ncobase/ncore/oss
```

## 快速开始

```go
package main

import (
    "fmt"
    "strings"

    "github.com/ncobase/ncore/oss"
)

func main() {
    // 创建存储配置
    cfg := &oss.Config{
        Provider: "minio",
        ID:       "minioadmin",
        Secret:   "minioadmin",
        Bucket:   "mybucket",
        Endpoint: "http://localhost:9000",
    }

    // 创建存储实例
    storage, err := oss.NewStorage(cfg)
    if err != nil {
        panic(err)
    }

    // 上传文件
    content := strings.NewReader("你好，世界！")
    obj, err := storage.Put("test.txt", content)
    if err != nil {
        panic(err)
    }
    fmt.Printf("已上传: %s\n", obj.Path)

    // 检查文件是否存在
    exists, err := storage.Exists("test.txt")
    if err != nil {
        panic(err)
    }
    fmt.Printf("存在: %v\n", exists)

    // 获取文件元数据
    stat, err := storage.Stat("test.txt")
    if err != nil {
        panic(err)
    }
    fmt.Printf("大小: %d 字节\n", stat.Size)

    // 获取文件
    file, err := storage.Get("test.txt")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    // 获取预签名下载 URL
    url, err := storage.GetURL("test.txt")
    if err != nil {
        panic(err)
    }
    fmt.Printf("URL: %s\n", url)

    // 删除文件
    if err := storage.Delete("test.txt"); err != nil {
        panic(err)
    }
}
```

## 支持的提供商

| 提供商               | Provider 值  | 所需配置                                |
| -------------------- | ------------ | --------------------------------------- |
| AWS S3               | `s3`         | ID, Secret, Bucket, Region              |
| Azure Blob           | `azure`      | ID (账户名), Secret (密钥), Bucket      |
| 阿里云 OSS           | `aliyun`     | ID, Secret, Bucket, Region              |
| 腾讯云 COS           | `tencent`    | ID, Secret, Bucket, Region, AppID       |
| Google Cloud Storage | `gcs`        | Bucket, ServiceAccountJSON 或 Secret   |
| MinIO                | `minio`      | ID, Secret, Bucket, Endpoint            |
| 七牛 Kodo            | `qiniu`      | ID, Secret, Bucket, Region, Endpoint    |
| Synology NAS         | `synology`   | ID, Secret, Bucket, Endpoint            |
| 本地文件系统         | `filesystem` | Bucket (路径，默认 `./uploads`)         |

## 配置示例

### AWS S3

```go
cfg := &oss.Config{
    Provider: "s3",
    ID:       "your-access-key-id",
    Secret:   "your-secret-access-key",
    Bucket:   "my-bucket",
    Region:   "us-east-1",
}
```

### MinIO

```go
cfg := &oss.Config{
    Provider: "minio",
    ID:       "minioadmin",
    Secret:   "minioadmin",
    Bucket:   "mybucket",
    Endpoint: "http://localhost:9000",
}
```

### 阿里云 OSS

```go
cfg := &oss.Config{
    Provider: "aliyun",
    ID:       "your-access-key-id",
    Secret:   "your-access-key-secret",
    Bucket:   "my-bucket",
    Region:   "cn-hangzhou",
}
```

### 腾讯云 COS

```go
cfg := &oss.Config{
    Provider: "tencent",
    ID:       "your-secret-id",
    Secret:   "your-secret-key",
    Bucket:   "my-bucket",
    Region:   "ap-guangzhou",
    AppID:    "1234567890",
}
```

### Azure Blob 存储

```go
cfg := &oss.Config{
    Provider: "azure",
    ID:       "your-account-name",
    Secret:   "your-account-key",
    Bucket:   "my-container",
}
```

### Google Cloud Storage

```go
cfg := &oss.Config{
    Provider:           "gcs",
    Bucket:             "my-bucket",
    ServiceAccountJSON: "/path/to/service-account.json",
}
```

### 七牛 Kodo

```go
cfg := &oss.Config{
    Provider: "qiniu",
    ID:       "your-access-key",
    Secret:   "your-secret-key",
    Bucket:   "my-bucket",
    Region:   "cn-east-1",
    Endpoint: "https://my-bucket.qiniudn.com",
}
```

### Synology NAS

```go
cfg := &oss.Config{
    Provider: "synology",
    ID:       "your-access-key",
    Secret:   "your-secret-key",
    Bucket:   "my-bucket",
    Endpoint: "https://nas.example.com:5001",
}
```

### 本地文件系统

```go
cfg := &oss.Config{
    Provider: "filesystem",
    Bucket:   "/var/data/storage",
}
```

## API 参考

### Interface

```go
type Interface interface {
    // Get 下载文件到临时文件并返回文件句柄
    // 调用者负责关闭文件并在完成后删除
    Get(path string) (*os.File, error)

    // GetStream 返回用于流式下载大文件的可读流
    // 调用者负责在完成后关闭读取器
    GetStream(path string) (io.ReadCloser, error)

    // Put 从给定的读取器上传文件到指定路径
    // 成功时返回对象元数据
    Put(path string, reader io.Reader) (*Object, error)

    // Delete 删除指定路径的文件
    // 如果文件不存在或成功删除则返回 nil
    Delete(path string) error

    // List 返回指定路径前缀下的所有对象
    // 如果没有找到对象则返回空切片
    List(path string) ([]*Object, error)

    // GetURL 生成用于下载文件的预签名 URL
    // URL 通常有效期为 1 小时
    GetURL(path string) (string, error)

    // GetEndpoint 返回存储服务端点 URL
    GetEndpoint() string

    // Exists 检查指定路径是否存在对象
    Exists(path string) (bool, error)

    // Stat 获取对象元数据而不下载内容
    Stat(path string) (*Object, error)
}
```

### Object

```go
type Object struct {
    Path             string     // 存储中的文件路径
    Name             string     // 文件名
    LastModified     *time.Time // 最后修改时间
    Size             int64      // 文件大小（字节）
    StorageInterface Interface  // 关联的存储接口
}
```

### Config

```go
type Config struct {
    Provider           string // 存储提供商: s3, minio, aliyun, azure, tencent, qiniu, gcs, synology, filesystem
    ID                 string // Access Key ID / 账户名
    Secret             string // Secret Access Key / 账户密钥
    Region             string // 区域（云存储必需）
    Bucket             string // 存储桶名 / 容器名 / 本地路径
    Endpoint           string // 自定义端点（MinIO, Synology 必需）
    ServiceAccountJSON string // GCS 服务账户 JSON 文件路径
    AppID              string // 腾讯云 COS 应用 ID
    Debug              bool   // 启用调试模式
}
```

## 高级用法

### 列出文件

```go
objects, err := storage.List("images/")
if err != nil {
    panic(err)
}

for _, obj := range objects {
    fmt.Printf("文件: %s, 大小: %d 字节\n", obj.Path, obj.Size)
}
```

### 检查文件是否存在

```go
exists, err := storage.Exists("file.txt")
if err != nil {
    panic(err)
}

if exists {
    fmt.Println("文件存在")
} else {
    fmt.Println("文件不存在")
}
```

### 获取文件元数据

```go
obj, err := storage.Stat("document.pdf")
if err != nil {
    panic(err)
}

fmt.Printf("文件: %s\n", obj.Name)
fmt.Printf("大小: %d 字节\n", obj.Size)
fmt.Printf("最后修改: %s\n", obj.LastModified)
```

### 流式下载

```go
reader, err := storage.GetStream("large-file.zip")
if err != nil {
    panic(err)
}
defer reader.Close()

// 处理流...
```

## 自定义驱动

实现 `Driver` 接口以添加新的存储提供商支持：

```go
type Driver interface {
    // Name 返回驱动名称
    Name() string

    // Connect 建立与存储服务的连接
    Connect(ctx context.Context, cfg *Config) (Interface, error)

    // Close 关闭存储连接
    Close(conn Interface) error
}
```

在 `init()` 函数中注册驱动：

```go
func init() {
    oss.RegisterDriver(&myDriver{})
}
```

## 从 ncore/data/storage 迁移

如果您正在从旧版本升级：

```go
// 旧代码
import "github.com/ncobase/ncore/data/storage"

// 新代码
import "github.com/ncobase/ncore/oss"
```

## 许可证

详见 [LICENSE](../LICENSE) 文件。

## 贡献

欢迎提交 Pull Request 和 Issue！

## 相关链接

- [NCore 主仓库](https://github.com/ncobase/ncore)
