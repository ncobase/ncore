package storage

import (
	"github.com/casdoor/oss"
	"github.com/casdoor/oss/tencent"
)

// NewTencentCloud new tencent cloud cos
func NewTencentCloud(c *Config) oss.StorageInterface {
	return tencent.New(&tencent.Config{
		AccessID:  c.ID,
		AccessKey: c.Secret,
		Region:    c.Region,
		Bucket:    c.Bucket,
		Endpoint:  c.Endpoint,
	})
}
