package storage

import (
	"github.com/casdoor/oss"
	"github.com/casdoor/oss/azureblob"
)

// NewAzure new azure blob storage
func NewAzure(c *Config) oss.StorageInterface {
	return azureblob.New(&azureblob.Config{
		AccessId:  c.ID,
		AccessKey: c.Secret,
		Region:    c.Region,
		Bucket:    c.Bucket,
		Endpoint:  c.Endpoint,
	})
}
