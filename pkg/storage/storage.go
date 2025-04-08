package storage

import (
	"errors"

	"github.com/casdoor/oss"
	"github.com/spf13/viper"
)

// Config config
type Config struct {
	Provider string
	ID       string
	Secret   string
	Region   string
	Bucket   string
	Endpoint string
}

// NewStorage new storage
func NewStorage(c *Config) (oss.StorageInterface, error) {
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
	default:
		return nil, errors.New("unsupported storage type")
	}
}

// GetConfig get storage config
func GetConfig(v *viper.Viper) *Config {
	return &Config{
		Provider: v.GetString("storage.provider"),
		ID:       v.GetString("storage.id"),
		Secret:   v.GetString("storage.secret"),
		Region:   v.GetString("storage.region"),
		Bucket:   v.GetString("storage.bucket"),
		Endpoint: v.GetString("storage.endpoint"),
	}
}
