# NCoreOSS - Object Storage Service for Go

English | [简体中文](README_zh-CN.md)

Standalone object storage service module with support for multiple cloud storage providers and local file systems.

## Features

- 🌐 **Multi-Cloud Support**: AWS S3, Azure Blob, Aliyun OSS, Tencent COS, Google Cloud Storage, MinIO, Synology, Qiniu, and Local Filesystem (9 providers).
- 📦 **Unified Interface**: All storage providers use the same consistent API interface.
- 🔐 **Official SDKs**: Leverages official SDKs from each cloud provider to ensure stability and feature completeness.
- 💾 **Local Storage**: Built-in support for local file system storage.
- ⚡ **Lightweight & Efficient**: Standalone module that can be imported independently, minimizing dependencies.
- 🛠 **Extensible**: Simple interface design makes it easy to add new storage providers.

## Installation

```bash
go get github.com/ncobase/ncore/oss
```

## Quick Start

### 1. Direct Usage (No Registration Needed)

You can use NCoreOSS directly by importing the module. For direct use, no manual driver registration is required as the core module handles the built-in providers.

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

 // Upload file
 content := strings.NewReader("Hello, World!")
 obj, err := storage.Put("test.txt", content)
 if err != nil {
  panic(err)
 }
 fmt.Printf("Uploaded: %s\n", obj.Path)

 // Get file
 file, err := storage.Get("test.txt")
 if err != nil {
  panic(err)
 }
 defer file.Close()

 // Get download URL
 url, err := storage.GetURL("test.txt")
 if err != nil {
  panic(err)
 }
 fmt.Printf("URL: %s\n", url)

 // Delete file
 err = storage.Delete("test.txt")
 if err != nil {
  panic(err)
 }
}
```

## Supported Providers

### Cloud Storage Services

| Provider             | Provider Value | Required Config                                      |
| -------------------- | -------------- | ---------------------------------------------------- |
| AWS S3               | `s3`           | Endpoint, Bucket, ID, Secret, Region                 |
| Azure Blob           | `azure`        | Endpoint, Bucket (container), ID (account), Secret   |
| Aliyun OSS           | `aliyun`       | Endpoint, Bucket, ID, Secret                         |
| Tencent COS          | `tencent`      | Endpoint, Bucket, ID, Secret                         |
| Google Cloud Storage | `gcs`          | Endpoint, Bucket, ID (project ID), Secret (JSON key) |
| MinIO                | `minio`        | Endpoint, Bucket, ID, Secret                         |
| Qiniu Kodo           | `qiniu`        | Endpoint, Bucket, ID, Secret                         |
| Synology NAS         | `synology`     | Endpoint, Bucket, ID, Secret                         |

### Local Storage

| Provider         | Provider Value | Required Config             |
| ---------------- | -------------- | --------------------------- |
| Local Filesystem | `filesystem`   | Path (local directory path) |

## Configuration Examples

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

### Aliyun OSS

```go
cfg := &oss.Config{
 Provider: "aliyun",
 Endpoint: "oss-cn-hangzhou.aliyuncs.com",
 Bucket:   "my-bucket",
 ID:       "your-access-key-id",
 Secret:   "your-access-key-secret",
}
```

### Tencent COS

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

### Local Filesystem

```go
cfg := &oss.Config{
 Provider: "filesystem",
 Path:     "/var/data/storage", // Local storage path
}
```

## API Reference

### Core Interface

```go
type Interface interface {
 // Put uploads a file to storage
 Put(path string, reader io.Reader, opts ...WriteOption) (*Object, error)

 // Get retrieves a file from storage
 Get(path string) (io.ReadCloser, error)

 // Delete removes a file from storage
 Delete(path string) error

 // GetURL returns the public access URL for a file
 GetURL(path string) (string, error)

 // List returns all objects with the given prefix
 List(prefix string) ([]*Object, error)

 // Stat retrieves object metadata
 Stat(path string) (*Object, error)

 // Exists checks if an object exists
 Exists(path string) (bool, error)
}
```

### Object Metadata

```go
type Object struct {
 Path         string    // Object path
 Name         string    // Object name
 Size         int64     // Size in bytes
 LastModified time.Time // Last modified time
 ETag         string    // ETag
 ContentType  string    // Content type
}
```

## Advanced Usage

### Upload with Options

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

### List Files

```go
// List all objects with "images/" prefix
objects, err := storage.List("images/")
if err != nil {
 panic(err)
}

for _, obj := range objects {
 fmt.Printf("File: %s, Size: %d bytes\n", obj.Path, obj.Size)
}
```

### Check File Existence

```go
exists, err := storage.Exists("file.txt")
if err != nil {
 panic(err)
}

if exists {
 fmt.Println("File exists")
} else {
 fmt.Println("File does not exist")
}
```

### Get File Metadata

```go
obj, err := storage.Stat("document.pdf")
if err != nil {
 panic(err)
}

fmt.Printf("File: %s\n", obj.Name)
fmt.Printf("Size: %d bytes\n", obj.Size)
fmt.Printf("Last Modified: %s\n", obj.LastModified)
```

## Module Relationship

### Relationship with `ncore/data/storage`

Historically, storage providers were part of the `ncore/data` module. Starting from v0.2.0, the storage logic has been extracted into this standalone `ncore/oss` module.

- **Standalone Module**: `ncore/oss` is now an independent module and no longer resides under `data/`.
- **Decoupled**: It can be used independently of the data layer.
- **Improved Performance**: Extracting storage providers significantly reduced the binary size and dependency count of the core `data` module.

## Migration Guide

### Migrating from `ncore/data/storage`

If you are upgrading from an older version where storage was part of `data`:

```go
// Old Code
import "github.com/ncobase/ncore/data/storage"

// New Code
import "github.com/ncobase/ncore/oss"
```

The configuration object remains compatible; you only need to update the import path.

## License

See the [LICENSE](../LICENSE) file for details.

## Contributing

Pull requests and issues are welcome!

## Related Links

- [NCore Main Repository](https://github.com/ncobase/ncore)
- [Documentation](https://docs.ncobase.com)
