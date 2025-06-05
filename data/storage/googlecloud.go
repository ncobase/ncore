package storage

import (
	"github.com/casdoor/oss/googlecloud"
)

// GoogleCloudConfig extends Config with Google Cloud specific fields
type GoogleCloudConfig struct {
	*Config
	ServiceAccountJSON string `json:"service_account_json" yaml:"service_account_json"`
}

// NewGoogleCloud creates new google cloud storage client
func NewGoogleCloud(c *Config) (Interface, error) {
	serviceAccountJSON := c.ServiceAccountJSON
	if serviceAccountJSON == "" {
		serviceAccountJSON = c.Secret // Fallback to Secret field
	}

	client, err := googlecloud.New(&googlecloud.Config{
		ServiceAccountJson: serviceAccountJSON,
		Bucket:             c.Bucket,
		Endpoint:           c.Endpoint,
	})
	if err != nil {
		return nil, err
	}
	return NewOSSAdapter(client), nil
}
