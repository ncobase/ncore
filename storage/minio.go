package storage

import (
	aws3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/casdoor/oss"
	"github.com/casdoor/oss/s3"
)

// NewMinio new minio
func NewMinio(c *Config) oss.StorageInterface {
	return s3.New(&s3.Config{
		AccessID:         c.ID,
		AccessKey:        c.Secret,
		Region:           c.Region,
		Bucket:           c.Bucket,
		Endpoint:         c.Endpoint,
		S3Endpoint:       c.Endpoint,
		ACL:              aws3.BucketCannedACLPublicRead,
		S3ForcePathStyle: true,
	})
}
