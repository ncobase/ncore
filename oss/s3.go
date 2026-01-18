package oss

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Adapter implements the Interface for AWS S3 storage.
// Supports both AWS S3 and S3-compatible services with custom endpoints.
type S3Adapter struct {
	client   *s3.Client
	presign  *s3.PresignClient
	bucket   string
	region   string
	endpoint string
}

// NewS3Adapter creates a new S3 storage adapter.
// For S3-compatible services, set the endpoint parameter.
func NewS3Adapter(accessKeyID, secretAccessKey, region, bucket, endpoint string) (*S3Adapter, error) {
	ctx := context.Background()

	var cfg aws.Config
	var err error

	if endpoint != "" {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				accessKeyID,
				secretAccessKey,
				"",
			)),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				accessKeyID,
				secretAccessKey,
				"",
			)),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		}
	})

	return &S3Adapter{
		client:   client,
		presign:  s3.NewPresignClient(client),
		bucket:   bucket,
		region:   region,
		endpoint: endpoint,
	}, nil
}

// Get downloads a file from S3 to a temporary local file.
func (a *S3Adapter) Get(path string) (*os.File, error) {
	reader, err := a.GetStream(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	ext := filepath.Ext(path)
	pattern := fmt.Sprintf("s3-*%s", ext)
	tmpFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := io.Copy(tmpFile, reader); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to copy object to temp file: %w", err)
	}

	if _, err := tmpFile.Seek(0, 0); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to seek temp file: %w", err)
	}

	return tmpFile, nil
}

// GetStream returns a readable stream for the S3 object.
func (a *S3Adapter) GetStream(path string) (io.ReadCloser, error) {
	ctx := context.Background()

	resp, err := a.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return resp.Body, nil
}

// Put uploads a file to S3 from the given reader.
func (a *S3Adapter) Put(path string, reader io.Reader) (*Object, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}
	if reader == nil {
		return nil, fmt.Errorf("reader cannot be nil")
	}

	ctx := context.Background()

	contentType := "application/octet-stream"
	if ext := filepath.Ext(path); ext != "" {
		if ct := getContentType(ext); ct != "" {
			contentType = ct
		}
	}

	_, err := a.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(a.bucket),
		Key:         aws.String(path),
		Body:        reader,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to put object: %w", err)
	}

	now := time.Now()
	return &Object{
		Path:             path,
		Name:             filepath.Base(path),
		LastModified:     &now,
		Size:             0,
		StorageInterface: a,
	}, nil
}

// Delete removes an object from the S3 bucket.
func (a *S3Adapter) Delete(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	ctx := context.Background()

	_, err := a.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// List returns all objects under the specified prefix.
func (a *S3Adapter) List(path string) ([]*Object, error) {
	ctx := context.Background()

	paginator := s3.NewListObjectsV2Paginator(a.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(a.bucket),
		Prefix: aws.String(path),
	})

	var objects []*Object
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			objects = append(objects, &Object{
				Path:             aws.ToString(obj.Key),
				Name:             filepath.Base(aws.ToString(obj.Key)),
				LastModified:     obj.LastModified,
				Size:             *obj.Size,
				StorageInterface: a,
			})
		}
	}

	return objects, nil
}

// GetURL generates a presigned URL valid for 1 hour.
func (a *S3Adapter) GetURL(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	ctx := context.Background()

	presignedReq, err := a.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(path),
	}, s3.WithPresignExpires(1*time.Hour))

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedReq.URL, nil
}

// GetEndpoint returns the S3 endpoint URL.
func (a *S3Adapter) GetEndpoint() string {
	if a.endpoint != "" {
		return a.endpoint
	}
	return fmt.Sprintf("https://s3.%s.amazonaws.com", a.region)
}

// Exists checks if an object exists in the S3 bucket.
func (a *S3Adapter) Exists(path string) (bool, error) {
	if path == "" {
		return false, fmt.Errorf("path cannot be empty")
	}

	ctx := context.Background()
	_, err := a.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		// Check if error is "not found"
		var nsk *types.NotFound
		if errors.As(err, &nsk) {
			return false, nil
		}
		// Also check for NoSuchKey
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}
	return true, nil
}

// Stat retrieves object metadata without downloading content.
func (a *S3Adapter) Stat(path string) (*Object, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	ctx := context.Background()
	resp, err := a.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}

	return &Object{
		Path:             path,
		Name:             filepath.Base(path),
		LastModified:     resp.LastModified,
		Size:             aws.ToInt64(resp.ContentLength),
		StorageInterface: a,
	}, nil
}

// s3Driver implements the Driver interface for AWS S3.
type s3Driver struct{}

// Name returns the driver name.
func (d *s3Driver) Name() string {
	return "s3"
}

// Connect establishes a connection to AWS S3.
func (d *s3Driver) Connect(ctx context.Context, cfg *Config) (Interface, error) {
	endpoint := ""
	if cfg.Endpoint != "" {
		endpoint = cfg.Endpoint
	}
	return NewS3Adapter(cfg.ID, cfg.Secret, cfg.Region, cfg.Bucket, endpoint)
}

// Close closes the S3 connection.
func (d *s3Driver) Close(conn Interface) error {
	return nil
}

func init() {
	RegisterDriver(&s3Driver{})
}
