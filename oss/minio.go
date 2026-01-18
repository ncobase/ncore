package oss

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioAdapter implements the Interface for MinIO object storage.
// Compatible with any S3-compatible storage service.
type MinioAdapter struct {
	client *minio.Client
	bucket string
}

// NewMinioAdapter creates a new MinIO storage adapter.
func NewMinioAdapter(endpoint, accessKeyID, secretAccessKey, bucket string, useSSL bool) (*MinioAdapter, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	return &MinioAdapter{
		client: client,
		bucket: bucket,
	}, nil
}

// Get downloads a file from MinIO to a temporary local file.
func (a *MinioAdapter) Get(path string) (*os.File, error) {
	reader, err := a.GetStream(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	ext := filepath.Ext(path)
	pattern := fmt.Sprintf("minio-*%s", ext)
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

// GetStream returns a readable stream for the MinIO object.
func (a *MinioAdapter) GetStream(path string) (io.ReadCloser, error) {
	ctx := context.Background()
	object, err := a.client.GetObject(ctx, a.bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	return object, nil
}

// Put uploads a file to MinIO from the given reader.
func (a *MinioAdapter) Put(path string, reader io.Reader) (*Object, error) {
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

	info, err := a.client.PutObject(ctx, a.bucket, path, reader, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to put object: %w", err)
	}

	now := time.Now()
	return &Object{
		Path:             path,
		Name:             filepath.Base(path),
		LastModified:     &now,
		Size:             info.Size,
		StorageInterface: a,
	}, nil
}

// Delete removes an object from the MinIO bucket.
func (a *MinioAdapter) Delete(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	ctx := context.Background()
	err := a.client.RemoveObject(ctx, a.bucket, path, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// List returns all objects under the specified prefix.
func (a *MinioAdapter) List(path string) ([]*Object, error) {
	ctx := context.Background()

	opts := minio.ListObjectsOptions{
		Prefix:    path,
		Recursive: true,
	}

	var objects []*Object
	for object := range a.client.ListObjects(ctx, a.bucket, opts) {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}

		objects = append(objects, &Object{
			Path:             object.Key,
			Name:             filepath.Base(object.Key),
			LastModified:     &object.LastModified,
			Size:             object.Size,
			StorageInterface: a,
		})
	}

	return objects, nil
}

// GetURL generates a presigned URL valid for 1 hour.
func (a *MinioAdapter) GetURL(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	ctx := context.Background()
	presignedURL, err := a.client.PresignedGetObject(ctx, a.bucket, path, 1*time.Hour, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

// GetEndpoint returns the MinIO endpoint URL.
func (a *MinioAdapter) GetEndpoint() string {
	return a.client.EndpointURL().String()
}

// Exists checks if an object exists in the MinIO bucket.
func (a *MinioAdapter) Exists(path string) (bool, error) {
	if path == "" {
		return false, fmt.Errorf("path cannot be empty")
	}

	ctx := context.Background()
	_, err := a.client.StatObject(ctx, a.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}
	return true, nil
}

// Stat retrieves object metadata without downloading content.
func (a *MinioAdapter) Stat(path string) (*Object, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	ctx := context.Background()
	info, err := a.client.StatObject(ctx, a.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}

	return &Object{
		Path:             path,
		Name:             filepath.Base(path),
		LastModified:     &info.LastModified,
		Size:             info.Size,
		StorageInterface: a,
	}, nil
}

// minioDriver implements the Driver interface for MinIO.
type minioDriver struct{}

// Name returns the driver name.
func (d *minioDriver) Name() string {
	return "minio"
}

// Connect establishes a connection to MinIO storage.
func (d *minioDriver) Connect(ctx context.Context, cfg *Config) (Interface, error) {
	endpoint := cfg.Endpoint
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")
	useSSL := strings.HasPrefix(cfg.Endpoint, "https://")

	return NewMinioAdapter(endpoint, cfg.ID, cfg.Secret, cfg.Bucket, useSSL)
}

// Close closes the MinIO connection.
func (d *minioDriver) Close(conn Interface) error {
	return nil
}

func init() {
	RegisterDriver(&minioDriver{})
}
