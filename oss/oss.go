// Package oss provides a unified object storage abstraction layer supporting
// multiple cloud storage providers including AWS S3, Azure Blob, Aliyun OSS,
// Tencent COS, Google Cloud Storage, MinIO, Qiniu Kodo, Synology NAS,
// and local filesystem.
//
// All providers implement a common Interface for consistent operations
// across different storage backends. Drivers are auto-registered via init()
// functions, enabling transparent provider selection at runtime.
package oss

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

// Interface defines unified object storage operations.
// All storage providers implement this interface for consistent API access.
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

// Object represents metadata about a stored object.
type Object struct {
	Path             string     // File path in storage
	Name             string     // File name
	LastModified     *time.Time // Last modification time
	Size             int64      // File size in bytes
	StorageInterface Interface  // Associated storage interface
}

// Get retrieves the file for this object.
// This is a convenience method that calls the associated storage interface's Get method.
func (object Object) Get() (*os.File, error) {
	return object.StorageInterface.Get(object.Path)
}

// Config holds configuration for object storage providers.
type Config struct {
	Provider           string `json:"provider" yaml:"provider"`                                             // Storage provider: minio, s3, aliyun, azure, tencent, qiniu, gcs, synology, filesystem
	ID                 string `json:"id" yaml:"id"`                                                         // Access key ID / Account name
	Secret             string `json:"secret" yaml:"secret"`                                                 // Secret access key / Account key
	Region             string `json:"region" yaml:"region"`                                                 // Region (required for cloud storage)
	Bucket             string `json:"bucket" yaml:"bucket"`                                                 // Bucket name / Container name / Local path
	Endpoint           string `json:"endpoint" yaml:"endpoint"`                                             // Custom endpoint (required for MinIO, Synology)
	ServiceAccountJSON string `json:"service_account_json,omitempty" yaml:"service_account_json,omitempty"` // Service account JSON file path for Google Cloud Storage
	SharedFolder       string `json:"shared_folder,omitempty" yaml:"shared_folder,omitempty"`               // Synology shared folder (optional)
	OtpCode            string `json:"otp_code,omitempty" yaml:"otp_code,omitempty"`                         // Synology 2FA code (optional)
	Debug              bool   `json:"debug,omitempty" yaml:"debug,omitempty"`                               // Enable debug mode (optional)
	AppID              string `json:"app_id,omitempty" yaml:"app_id,omitempty"`                             // Tencent COS Application ID
}

// Validate checks if the configuration is valid and sets default values where applicable.
func (c *Config) Validate() error {
	if c.Provider == "" {
		return errors.New("storage provider is required")
	}

	switch c.Provider {
	case "filesystem", "local":
		if c.Bucket == "" {
			c.Bucket = "./uploads"
		}
	case "aliyun", "aliyun-oss":
		if c.ID == "" || c.Secret == "" || c.Bucket == "" {
			return errors.New("id, secret, and bucket are required for Aliyun OSS")
		}
		if c.Region == "" {
			c.Region = "cn-hangzhou"
		}
	case "s3", "aws-s3", "aws":
		if c.ID == "" || c.Secret == "" || c.Bucket == "" {
			return errors.New("id, secret, and bucket are required for AWS S3")
		}
		if c.Region == "" {
			c.Region = "us-east-1"
		}
	case "azure", "azure-blob":
		if c.ID == "" || c.Secret == "" || c.Bucket == "" {
			return errors.New("account name, account key, and container are required for Azure Blob")
		}
	case "tencent", "tencent-cos":
		if c.ID == "" || c.Secret == "" || c.Bucket == "" || c.AppID == "" {
			return errors.New("id, secret, bucket, and app_id are required for Tencent COS")
		}
		if c.Region == "" {
			c.Region = "ap-guangzhou"
		}
	case "minio":
		if c.ID == "" || c.Secret == "" || c.Bucket == "" || c.Endpoint == "" {
			return errors.New("id, secret, bucket, and endpoint are required for MinIO")
		}
	case "qiniu":
		if c.ID == "" || c.Secret == "" || c.Bucket == "" {
			return errors.New("id, secret, and bucket are required for Qiniu Kodo")
		}
		if c.Region == "" {
			c.Region = "cn-east-1"
		}
	case "gcs", "google", "google-cloud":
		if c.Bucket == "" {
			return errors.New("bucket is required for Google Cloud Storage")
		}
		if c.Secret == "" && c.ServiceAccountJSON == "" {
			return errors.New("service account JSON is required for Google Cloud Storage")
		}
	case "synology":
		if c.ID == "" || c.Secret == "" || c.Bucket == "" || c.Endpoint == "" {
			return errors.New("id, secret, bucket, and endpoint are required for Synology")
		}
	default:
		return fmt.Errorf("unsupported storage provider: %s", c.Provider)
	}

	return nil
}

// Driver defines the storage driver interface.
// Implement this interface to add support for new storage providers.
type Driver interface {
	// Name returns the driver name.
	Name() string

	// Connect establishes a connection to the storage service.
	Connect(ctx context.Context, cfg *Config) (Interface, error)

	// Close closes the storage connection.
	Close(conn Interface) error
}

var driverRegistry = make(map[string]Driver)

// RegisterDriver registers a storage driver.
// Typically called in the driver package's init function.
func RegisterDriver(driver Driver) {
	name := driver.Name()
	if _, exists := driverRegistry[name]; exists {
		panic(fmt.Sprintf("oss driver %s already registered", name))
	}
	driverRegistry[name] = driver
}

// GetDriver retrieves a driver by name.
// Returns an error if the driver is not registered.
func GetDriver(name string) (Driver, error) {
	driver, ok := driverRegistry[name]
	if !ok {
		return nil, fmt.Errorf("oss driver %s not found", name)
	}
	return driver, nil
}

// NewStorage creates a storage instance based on the provided configuration.
// Automatically selects the appropriate storage provider.
func NewStorage(c *Config) (Interface, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("invalid storage config: %w", err)
	}

	if c.Provider == "filesystem" || c.Provider == "local" {
		return NewFileSystem(c.Bucket), nil
	}

	driver, err := GetDriver(c.Provider)
	if err != nil {
		return nil, err
	}

	storage, err := driver.Connect(context.Background(), c)
	if err != nil {
		return nil, fmt.Errorf("failed to connect with %s driver: %w", c.Provider, err)
	}

	return storage, nil
}
