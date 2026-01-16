# NCoreOSS - Go 语言对象存储服务

[English](README.md) | 简体中文

支持多云存储提供商和本地文件系统的独立对象存储服务模块。

## 特性

- 🌐 **多云支持**：AWS S3、Azure Blob、阿里云 OSS、腾讯云 COS、Google Cloud Storage、MinIO、七牛 Kodo、Synology NAS 和本地文件系统（9 个提供商）
- 📦 **统一接口**：所有存储提供商使用相同的一致性 API 接口
- 🔐 **官方 SDK**：使用各云服务商的官方 SDK，确保稳定性和功能完整性
- 💾 **本地存储**：内置本地文件系统存储支持
- ⚡ **轻量高效**：可独立导入的独立模块，最小化依赖
- 🛠 **可扩展**：简单的接口设计，便于添加新的存储提供商

## 安装

```bash
go get github.com/ncobase/ncore/oss
```

## 快速开始

### 1. 直接使用（无需注册）

您可以直接导入模块使用 NCoreOSS。直接使用时无需手动注册驱动，核心模块会处理内置提供商。

```go
package main

import (
 "fmt"
 "strings"
 "github.com/ncobase/ncore/oss"
)

func main() {
 cfg := &oss.Config{
  Provider: "minio",
  ID:       "minioadmin",
  Secret:   "minioadmin",
  Bucket:   "mybucket",
  Endpoint: "http://localhost:9000",
 }

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
 fmt.Printf("已上传：%s\n", obj.Path)

 // 获取文件
 file, err := storage.Get("test.txt")
 if err != nil {
  panic(err)
 }
 defer file.Close()

 // 获取下载 URL
 url, err := storage.GetURL("test.txt")
 if err != nil {
  panic(err)
 }
 fmt.Printf("URL: %s\n", url)

 // 删除文件
 err = storage.Delete("test.txt")
 if err != nil {
  panic(err)
 }
}
```

## 支持的提供商

### 云存储服务

| 提供商               | Provider 值 | 所需配置                                           |
| -------------------- | ----------- | -------------------------------------------------- |
| AWS S3               | `s3`        | Endpoint, Bucket, ID, Secret, Region               |
| Azure Blob           | `azure`     | Endpoint, Bucket (容器名), ID (账户名), Secret     |
| 阿里云 OSS           | `aliyun`    | Endpoint, Bucket, ID, Secret                       |
| 腾讯云 COS           | `tencent`   | Endpoint, Bucket, ID, Secret                       |
| Google Cloud Storage | `gcs`       | Endpoint, Bucket, ID (项目 ID), Secret (JSON 密钥) |
| MinIO                | `minio`     | Endpoint, Bucket, ID, Secret                       |
| 七牛 Kodo            | `qiniu`     | Endpoint, Bucket, ID, Secret                       |
| Synology NAS         | `synology`  | Endpoint, Bucket, ID, Secret                       |

### 本地存储

| 提供商       | Provider 值  | 所需配置            |
| ------------ | ------------ | ------------------- |
| 本地文件系统 | `filesystem` | Path (本地目录路径) |

## 配置示例

### MinIO

```go
cfg := &oss.Config{
 Provider: "minio",
 Endpoint: "http://localhost:9000",
 Bucket:   "mybucket",
 ID:       "minioadmin",
 Secret:   "minioadmin",
 UseSSL:   false,
}
```

### 阿里云 OSS

```go
cfg := &oss.Config{
 Provider: "aliyun",
 Endpoint: "oss-cn-hangzhou.aliyuncs.com",
 Bucket:   "my-bucket",
 ID:       "your-access-key-id",
 Secret:   "your-access-key-secret",
}
```

### 腾讯云 COS

```go
cfg := &oss.Config{
 Provider: "tencent",
 Endpoint: "https://cos.ap-guangzhou.myqcloud.com",
 Bucket:   "my-bucket-1234567890",
 ID:       "your-secret-id",
 Secret:   "your-secret-key",
}
```

### AWS S3

```go
cfg := &oss.Config{
 Provider: "s3",
 Endpoint: "s3.amazonaws.com",
 Bucket:   "my-bucket",
 Region:   "us-east-1",
 ID:       "your-access-key-id",
 Secret:   "your-secret-access-key",
}
```

### 本地文件系统

```go
cfg := &oss.Config{
 Provider: "filesystem",
 Path:     "/var/data/storage", // 本地存储路径
}
```

## API 参考

### 核心接口

```go
type Interface interface {
 // Put 上传文件到存储
 Put(path string, reader io.Reader, opts ...WriteOption) (*Object, error)

 // Get 从存储获取文件
 Get(path string) (io.ReadCloser, error)

 // Delete 从存储删除文件
 Delete(path string) error

 // GetURL 获取文件的公开访问 URL
 GetURL(path string) (string, error)

 // List 列出指定前缀的所有对象
 List(prefix string) ([]*Object, error)

 // Stat 获取对象的元数据
 Stat(path string) (*Object, error)

 // Exists 检查对象是否存在
 Exists(path string) (bool, error)
}
```

### 对象元数据

```go
type Object struct {
 Path         string    // 对象路径
 Name         string    // 对象名称
 Size         int64     // 大小（字节）
 LastModified time.Time // 最后修改时间
 ETag         string    // ETag
 ContentType  string    // 内容类型
}
```

## 高级用法

### 带选项的上传

```go
import "github.com/ncobase/ncore/oss"

obj, err := storage.Put("image.jpg", reader,
 oss.WithContentType("image/jpeg"),
 oss.WithMetadata(map[string]string{
  "user-id": "12345",
  "version": "v1",
 }),
)
```

### 列出文件

```go
// 列出所有带 "images/" 前缀的对象
objects, err := storage.List("images/")
if err != nil {
 panic(err)
}

for _, obj := range objects {
 fmt.Printf("文件：%s，大小：%d 字节\n", obj.Path, obj.Size)
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

fmt.Printf("文件：%s\n", obj.Name)
fmt.Printf("大小：%d 字节\n", obj.Size)
fmt.Printf("最后修改：%s\n", obj.LastModified)
```

## 模块关系

### 与 `ncore/data/storage` 的关系

历史上，存储提供商是 `ncore/data` 模块的一部分。从 v0.2.0 开始，存储逻辑已被提取到这个独立的 `ncore/oss` 模块中。

- **独立模块**：`ncore/oss` 现在是独立模块，不再位于 `data/` 下
- **解耦**：可以独立于数据层使用
- **性能提升**：提取存储提供商显著减少了核心 `data` 模块的二进制大小和依赖数量

## 迁移指南

### 从 `ncore/data/storage` 迁移

如果您正在从存储是 `data` 一部分的旧版本升级：

```go
// 旧代码
import "github.com/ncobase/ncore/data/storage"

// 新代码
import "github.com/ncobase/ncore/oss"
```

配置对象保持兼容；您只需要更新导入路径。

## 许可证

详见 [LICENSE](../LICENSE) 文件。

## 贡献

欢迎提交 Pull Request 和 Issue！

## 相关链接

- [NCore 主仓库](https://github.com/ncobase/ncore)
- [文档](https://docs.ncobase.com)
