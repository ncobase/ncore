package storage

import (
	"github.com/casdoor/oss/synology"
)

// NewSynology creates new synology NAS storage client
func NewSynology(c *Config) Interface {
	sharedFolder := c.SharedFolder
	if sharedFolder == "" {
		sharedFolder = c.Bucket // Fallback to Bucket field
	}

	client := synology.New(&synology.Config{
		Endpoint:      c.Endpoint,
		AccessID:      c.ID,
		AccessKey:     c.Secret,
		SharedFolder:  sharedFolder,
		SessionExpire: false,
		Verify:        true,
		Debug:         c.Debug,
		OtpCode:       c.OtpCode,
	})
	return NewOSSAdapter(client)
}
