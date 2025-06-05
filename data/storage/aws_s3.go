package storage

import (
	aws3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/casdoor/oss/s3"
)

// NewS3 creates new aws s3 client
func NewS3(c *Config) Interface {
	client := s3.New(&s3.Config{
		AccessID:   c.ID,
		AccessKey:  c.Secret,
		Region:     c.Region,
		Bucket:     c.Bucket,
		Endpoint:   c.Endpoint,
		S3Endpoint: c.Endpoint,
		ACL:        aws3.BucketCannedACLPublicRead,
	})
	return NewOSSAdapter(client)
}
