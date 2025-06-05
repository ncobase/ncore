package storage

import (
	"github.com/casdoor/oss/qiniu"
)

// NewQiniu creates new qiniu cloud storage client
func NewQiniu(c *Config) (Interface, error) {
	client, err := qiniu.New(&qiniu.Config{
		AccessID:  c.ID,
		AccessKey: c.Secret,
		Region:    c.Region,
		Bucket:    c.Bucket,
		Endpoint:  c.Endpoint,
	})
	if err != nil {
		return nil, err
	}
	return NewOSSAdapter(client), nil
}
