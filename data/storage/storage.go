package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Interface represents the common interface for storage
type Interface interface {
	Get(path string) (*os.File, error)
	GetStream(path string) (io.ReadCloser, error)
	Put(path string, reader io.Reader) (*Object, error)
	Delete(path string) error
	List(path string) ([]*Object, error)
	GetURL(path string) (string, error)
	GetEndpoint() string
}

// Object represents a storage object
type Object struct {
	Path             string
	Name             string
	LastModified     *time.Time
	Size             int64
	StorageInterface Interface
}

// Get retrieves object's content
func (object Object) Get() (*os.File, error) {
	return object.StorageInterface.Get(object.Path)
}

// Config storage configuration
type Config struct {
	Provider string `json:"provider" yaml:"provider"`
	ID       string `json:"id" yaml:"id"`
	Secret   string `json:"secret" yaml:"secret"`
	Region   string `json:"region" yaml:"region"`
	Bucket   string `json:"bucket" yaml:"bucket"`
	Endpoint string `json:"endpoint" yaml:"endpoint"`
	// Extended fields for specific providers
	ServiceAccountJSON string `json:"service_account_json,omitempty" yaml:"service_account_json,omitempty"` // Google Cloud
	SharedFolder       string `json:"shared_folder,omitempty" yaml:"shared_folder,omitempty"`               // Synology
	OtpCode            string `json:"otp_code,omitempty" yaml:"otp_code,omitempty"`                         // Synology 2FA
	Debug              bool   `json:"debug,omitempty" yaml:"debug,omitempty"`                               // Synology
}

// Validate validates the storage configuration
func (c *Config) Validate() error {
	if c.Provider == "" {
		return errors.New("storage provider is required")
	}

	switch c.Provider {
	case "filesystem":
		if c.Bucket == "" {
			return errors.New("bucket (local path) is required for filesystem storage")
		}
	case "aliyun-oss", "minio", "aws-s3", "azure", "tencent-cos", "qiniu", "googlecloud", "synology":
		if c.ID == "" || c.Secret == "" || c.Bucket == "" {
			return errors.New("id, secret, and bucket are required for cloud storage")
		}
		if c.Region == "" && c.Provider != "minio" && c.Provider != "synology" {
			return errors.New("region is required for most cloud storage providers")
		}
		if c.Endpoint == "" && (c.Provider == "qiniu" || c.Provider == "synology") {
			return errors.New("endpoint is required for qiniu and synology storage")
		}
	default:
		return fmt.Errorf("unsupported storage provider: %s", c.Provider)
	}

	return nil
}

// NewStorage creates a new storage instance
func NewStorage(c *Config) (Interface, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("invalid storage config: %w", err)
	}

	switch c.Provider {
	case "aliyun-oss":
		return NewAliyun(c), nil
	case "minio":
		return NewMinio(c), nil
	case "aws-s3":
		return NewS3(c), nil
	case "azure":
		return NewAzure(c), nil
	case "filesystem":
		return NewFileSystem(c.Bucket), nil
	case "tencent-cos":
		return NewTencentCloud(c), nil
	case "qiniu":
		return NewQiniu(c)
	case "googlecloud":
		return NewGoogleCloud(c)
	case "synology":
		return NewSynology(c), nil
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", c.Provider)
	}
}

// GetConfig gets storage config from viper
func GetConfig(v *viper.Viper) *Config {
	return &Config{
		Provider:           v.GetString("storage.provider"),
		ID:                 v.GetString("storage.id"),
		Secret:             v.GetString("storage.secret"),
		Region:             v.GetString("storage.region"),
		Bucket:             v.GetString("storage.bucket"),
		Endpoint:           v.GetString("storage.endpoint"),
		ServiceAccountJSON: v.GetString("storage.service_account_json"),
		SharedFolder:       v.GetString("storage.shared_folder"),
		OtpCode:            v.GetString("storage.otp_code"),
		Debug:              v.GetBool("storage.debug"),
	}
}
