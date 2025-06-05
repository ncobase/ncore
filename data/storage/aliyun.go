package storage

import (
	"github.com/casdoor/oss/aliyun"
)

// NewAliyun creates new aliyun oss client
func NewAliyun(c *Config) Interface {
	client := aliyun.New(&aliyun.Config{
		AccessID:  c.ID,
		AccessKey: c.Secret,
		Bucket:    c.Bucket,
		Endpoint:  c.Endpoint,
	})
	return NewOSSAdapter(client)
}
