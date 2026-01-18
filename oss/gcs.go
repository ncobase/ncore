package oss

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GCSAdapter implements the Interface for Google Cloud Storage.
type GCSAdapter struct {
	client       *storage.Client
	bucket       string
	bucketHandle *storage.BucketHandle
}

// NewGCSAdapter creates a new Google Cloud Storage adapter.
func NewGCSAdapter(serviceAccountJSON, bucket string) (*GCSAdapter, error) {
	ctx := context.Background()

	var client *storage.Client
	var err error

	if serviceAccountJSON != "" {
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(serviceAccountJSON))
	} else {
		client, err = storage.NewClient(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	bucketHandle := client.Bucket(bucket)

	return &GCSAdapter{
		client:       client,
		bucket:       bucket,
		bucketHandle: bucketHandle,
	}, nil
}

// Get downloads an object from GCS to a temporary local file.
func (a *GCSAdapter) Get(path string) (*os.File, error) {
	reader, err := a.GetStream(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	ext := filepath.Ext(path)
	pattern := fmt.Sprintf("gcs-*%s", ext)
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

// GetStream returns a readable stream for the GCS object.
func (a *GCSAdapter) GetStream(path string) (io.ReadCloser, error) {
	ctx := context.Background()

	reader, err := a.bucketHandle.Object(path).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader: %w", err)
	}

	return reader, nil
}

// Put uploads a file to GCS from the given reader.
func (a *GCSAdapter) Put(path string, reader io.Reader) (*Object, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}
	if reader == nil {
		return nil, fmt.Errorf("reader cannot be nil")
	}

	ctx := context.Background()

	obj := a.bucketHandle.Object(path)
	writer := obj.NewWriter(ctx)

	contentType := "application/octet-stream"
	if ext := filepath.Ext(path); ext != "" {
		if ct := getContentType(ext); ct != "" {
			contentType = ct
		}
	}
	writer.ContentType = contentType

	if _, err := io.Copy(writer, reader); err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to write object: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get object attrs: %w", err)
	}

	return &Object{
		Path:             path,
		Name:             filepath.Base(path),
		LastModified:     &attrs.Updated,
		Size:             attrs.Size,
		StorageInterface: a,
	}, nil
}

// Delete removes an object from the GCS bucket.
func (a *GCSAdapter) Delete(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	ctx := context.Background()

	if err := a.bucketHandle.Object(path).Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// List returns all objects under the specified prefix.
func (a *GCSAdapter) List(path string) ([]*Object, error) {
	ctx := context.Background()

	query := &storage.Query{
		Prefix: path,
	}

	it := a.bucketHandle.Objects(ctx, query)

	var objects []*Object
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate objects: %w", err)
		}

		objects = append(objects, &Object{
			Path:             attrs.Name,
			Name:             filepath.Base(attrs.Name),
			LastModified:     &attrs.Updated,
			Size:             attrs.Size,
			StorageInterface: a,
		})
	}

	return objects, nil
}

// GetURL generates a signed URL valid for 1 hour.
func (a *GCSAdapter) GetURL(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(1 * time.Hour),
	}

	url, err := a.bucketHandle.SignedURL(path, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return url, nil
}

// GetEndpoint returns the Google Cloud Storage endpoint URL.
func (a *GCSAdapter) GetEndpoint() string {
	return fmt.Sprintf("https://storage.googleapis.com/%s", a.bucket)
}

// Exists checks if an object exists in the GCS bucket.
func (a *GCSAdapter) Exists(path string) (bool, error) {
	if path == "" {
		return false, fmt.Errorf("path cannot be empty")
	}

	ctx := context.Background()
	_, err := a.bucketHandle.Object(path).Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}
	return true, nil
}

// Stat retrieves object metadata without downloading content.
func (a *GCSAdapter) Stat(path string) (*Object, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	ctx := context.Background()
	attrs, err := a.bucketHandle.Object(path).Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}

	return &Object{
		Path:             path,
		Name:             filepath.Base(path),
		LastModified:     &attrs.Updated,
		Size:             attrs.Size,
		StorageInterface: a,
	}, nil
}

// gcsDriver implements the Driver interface for Google Cloud Storage.
type gcsDriver struct{}

// Name returns the driver name.
func (d *gcsDriver) Name() string {
	return "gcs"
}

// Connect establishes a connection to Google Cloud Storage.
func (d *gcsDriver) Connect(ctx context.Context, cfg *Config) (Interface, error) {
	serviceAccountJSON := cfg.ServiceAccountJSON
	if serviceAccountJSON == "" && cfg.Secret != "" {
		serviceAccountJSON = cfg.Secret
	}
	return NewGCSAdapter(serviceAccountJSON, cfg.Bucket)
}

// Close closes the GCS connection and releases resources.
func (d *gcsDriver) Close(conn Interface) error {
	if adapter, ok := conn.(*GCSAdapter); ok && adapter.client != nil {
		return adapter.client.Close()
	}
	return nil
}

func init() {
	RegisterDriver(&gcsDriver{})
}
