package storage

import (
	"fmt"

	aws3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/casdoor/oss/s3"
)

// NewMinio creates new minio client
func NewMinio(c *Config) (Interface, error) {
	// Validate config
	if c.Endpoint == "" {
		return nil, fmt.Errorf("endpoint is required for Minio")
	}
	if c.ID == "" || c.Secret == "" {
		return nil, fmt.Errorf("access ID and secret are required for Minio")
	}
	if c.Bucket == "" {
		return nil, fmt.Errorf("bucket is required for Minio")
	}

	// Default region if not provided
	region := c.Region
	if region == "" {
		region = "us-east-1"
	}

	client := s3.New(&s3.Config{
		AccessID:         c.ID,
		AccessKey:        c.Secret,
		Region:           region,
		Bucket:           c.Bucket,
		Endpoint:         c.Endpoint,
		S3Endpoint:       c.Endpoint,
		ACL:              aws3.BucketCannedACLPublicRead,
		S3ForcePathStyle: true,
	})

	return NewOSSAdapter(client), nil
}
