package storage

import (
	"github.com/casdoor/oss"
	"github.com/casdoor/oss/aliyun"
)

// NewAliyun new aliyun oss
func NewAliyun(c *Config) oss.StorageInterface {
	return aliyun.New(&aliyun.Config{
		AccessID:  c.ID,
		AccessKey: c.Secret,
		Bucket:    c.Bucket,
		Endpoint:  c.Endpoint,
	})
}
