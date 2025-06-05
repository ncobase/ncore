package storage

import (
	"github.com/casdoor/oss/tencent"
)

// NewTencentCloud creates new tencent cloud cos client
func NewTencentCloud(c *Config) Interface {
	client := tencent.New(&tencent.Config{
		AccessID:  c.ID,
		AccessKey: c.Secret,
		Region:    c.Region,
		Bucket:    c.Bucket,
		Endpoint:  c.Endpoint,
	})
	return NewOSSAdapter(client)
}
