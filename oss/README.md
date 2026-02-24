# OSS - Object Storage Service

English | [简体中文](README_zh-CN.md)

A unified object storage abstraction layer supporting multiple cloud storage providers and local filesystem.

## Features

- **Multi-Cloud Support**: AWS S3, Azure Blob, Aliyun OSS, Tencent COS, Google Cloud Storage, MinIO, Qiniu Kodo, Synology NAS, and Local Filesystem
- **Unified Interface**: All storage providers implement the same consistent API
- **Auto-Registration**: Drivers are automatically registered via `init()` functions
- **Official SDKs**: Leverages official SDKs from each cloud provider
- **Local Storage**: Built-in support for local filesystem storage
- **Lightweight**: Standalone module with minimal dependencies

## Installation

```bash
go get github.com/ncobase/ncore/oss
```

## Quick Start

```go
package main

import (
    "fmt"
    "strings"

    "github.com/ncobase/ncore/oss"
)

func main() {
    // Create storage configuration
    cfg := &oss.Config{
        Provider: "minio",
        ID:       "minioadmin",
        Secret:   "minioadmin",
        Bucket:   "mybucket",
        Endpoint: "http://localhost:9000",
    }

    // Create storage instance
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

    // Check if file exists
    exists, err := storage.Exists("test.txt")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Exists: %v\n", exists)

    // Get file metadata
    stat, err := storage.Stat("test.txt")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Size: %d bytes\n", stat.Size)

    // Get file
    file, err := storage.Get("test.txt")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    // Get presigned download URL
    url, err := storage.GetURL("test.txt")
    if err != nil {
        panic(err)
    }
    fmt.Printf("URL: %s\n", url)

    // Delete file
    if err := storage.Delete("test.txt"); err != nil {
        panic(err)
    }
}
```

## Supported Providers

| Provider             | Provider Value         | Required Config                                                      |
| -------------------- | ---------------------- | -------------------------------------------------------------------- |
| AWS S3               | `s3`                   | ID, Secret, Bucket, Region                                           |
| Cloudflare R2        | `r2`                   | ID, Secret, Bucket, Endpoint (Region defaults to `auto`)             |
| Backblaze B2 (S3)    | `b2`                   | ID, Secret, Bucket, Endpoint, Region                                 |
| Azure Blob           | `azure`                | ID (account), Secret (key), Bucket (Endpoint optional)               |
| Aliyun OSS           | `aliyun` / `oss`       | ID, Secret, Bucket, Region                                           |
| Tencent COS          | `tencent` / `cos`      | ID, Secret, Bucket, Region, AppID (or bucket as `<bucket>-<app_id>`) |
| Google Cloud Storage | `gcs`                  | Bucket, ServiceAccountJSON or Secret                                 |
| MinIO                | `minio`                | ID, Secret, Bucket, Endpoint                                         |
| Qiniu Kodo           | `qiniu`                | ID, Secret, Bucket, Region, Endpoint                                 |
| Synology NAS         | `synology`             | ID, Secret, Bucket, Endpoint                                         |
| Local Filesystem     | `filesystem` / `local` | Bucket (path, defaults to `./uploads`)                               |

## Configuration Examples

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

### Aliyun OSS

```go
cfg := &oss.Config{
    Provider: "aliyun",
    ID:       "your-access-key-id",
    Secret:   "your-access-key-secret",
    Bucket:   "my-bucket",
    Region:   "cn-hangzhou",
}
```

### Tencent COS

```go
cfg := &oss.Config{
    Provider: "tencent",
    ID:       "your-secret-id",
    Secret:   "your-secret-key",
    Bucket:   "my-bucket", // or "my-bucket-1250000000"
    Region:   "ap-guangzhou",
    AppID:    "1234567890",
}
```

### Azure Blob Storage

```go
cfg := &oss.Config{
    Provider: "azure",
    ID:       "your-account-name",
    Secret:   "your-account-key",
    Bucket:   "my-container",
    // Optional for sovereign clouds or custom DNS zones:
    // Endpoint: "https://your-account.blob.core.windows.net",
}
```

## Provider Endpoint References (Official)

- AWS S3: <https://docs.aws.amazon.com/AmazonS3/latest/userguide/UsingBucket.html>
- Cloudflare R2 (S3 API): <https://developers.cloudflare.com/r2/api/s3/api/>
- Backblaze B2 (S3 Compatible API): <https://www.backblaze.com/docs/cloud-storage-s3-compatible-api>
- Tencent COS endpoints and naming: <https://www.tencentcloud.com/document/product/436/6224>
- Aliyun OSS regions/endpoints: <https://www.alibabacloud.com/help/en/oss/user-guide/regions-and-endpoints>
- Azure Blob endpoint formats: <https://learn.microsoft.com/en-us/azure/storage/blobs/storage-blob-query-endpoint-srp>

### Google Cloud Storage

```go
cfg := &oss.Config{
    Provider:           "gcs",
    Bucket:             "my-bucket",
    ServiceAccountJSON: "/path/to/service-account.json",
}
```

### Qiniu Kodo

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

### Local Filesystem

```go
cfg := &oss.Config{
    Provider: "filesystem",
    Bucket:   "/var/data/storage",
}
```

## API Reference

### Interface

```go
type Interface interface {
    // Get downloads a file to a temporary file and returns the file handle.
    // Caller is responsible for closing the file and removing it when done.
    Get(path string) (*os.File, error)

    // GetStream returns a readable stream for streaming large file downloads.
    // Caller is responsible for closing the reader when done.
    GetStream(path string) (io.ReadCloser, error)

    // Put uploads a file from the given reader to the specified path.
    // Returns object metadata on success.
    Put(path string, reader io.Reader) (*Object, error)

    // Delete removes the file at the specified path.
    // Returns nil if file doesn't exist or was successfully deleted.
    Delete(path string) error

    // List returns all objects under the specified path prefix.
    // Returns empty slice if no objects found.
    List(path string) ([]*Object, error)

    // GetURL generates a presigned URL for downloading the file.
    // URL is typically valid for 1 hour.
    GetURL(path string) (string, error)

    // GetEndpoint returns the storage service endpoint URL.
    GetEndpoint() string

    // Exists checks if an object exists at the specified path.
    Exists(path string) (bool, error)

    // Stat retrieves object metadata without downloading content.
    Stat(path string) (*Object, error)
}
```

### Object

```go
type Object struct {
    Path             string     // File path in storage
    Name             string     // File name
    LastModified     *time.Time // Last modification time
    Size             int64      // File size in bytes
    StorageInterface Interface  // Associated storage interface
}
```

### Config

```go
type Config struct {
    Provider           string // Storage provider: s3, minio, aliyun, azure, tencent, qiniu, gcs, synology, filesystem
    ID                 string // Access key ID / Account name
    Secret             string // Secret access key / Account key
    Region             string // Region (required for cloud storage)
    Bucket             string // Bucket name / Container name / Local path
    Endpoint           string // Custom endpoint (required for MinIO, Synology)
    ServiceAccountJSON string // Service account JSON file path for GCS
    AppID              string // Tencent COS Application ID
    Debug              bool   // Enable debug mode
}
```

## Advanced Usage

### List Files

```go
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

### Stream Download

```go
reader, err := storage.GetStream("large-file.zip")
if err != nil {
    panic(err)
}
defer reader.Close()

// Process the stream...
```

## Custom Drivers

Implement the `Driver` interface to add support for new storage providers:

```go
type Driver interface {
    // Name returns the driver name.
    Name() string

    // Connect establishes a connection to the storage service.
    Connect(ctx context.Context, cfg *Config) (Interface, error)

    // Close closes the storage connection.
    Close(conn Interface) error
}
```

Register your driver in an `init()` function:

```go
func init() {
    oss.RegisterDriver(&myDriver{})
}
```

## Migration from ncore/data/storage

If you are upgrading from an older version:

```go
// Old Code
import "github.com/ncobase/ncore/data/storage"

// New Code
import "github.com/ncobase/ncore/oss"
```

## License

See the [LICENSE](../LICENSE) file for details.

## Contributing

Pull requests and issues are welcome!

## Related Links

- [NCore Main Repository](https://github.com/ncobase/ncore)
