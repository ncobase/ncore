package storage

import (
	"github.com/casdoor/oss/azureblob"
)

// NewAzure creates new azure blob storage client
func NewAzure(c *Config) Interface {
	client := azureblob.New(&azureblob.Config{
		AccessId:  c.ID,
		AccessKey: c.Secret,
		Region:    c.Region,
		Bucket:    c.Bucket,
		Endpoint:  c.Endpoint,
	})
	return NewOSSAdapter(client)
}
