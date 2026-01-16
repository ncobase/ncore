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

type GCSAdapter struct {
	client       *storage.Client
	bucket       string
	bucketHandle *storage.BucketHandle
}

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

func (a *GCSAdapter) GetStream(path string) (io.ReadCloser, error) {
	ctx := context.Background()

	reader, err := a.bucketHandle.Object(path).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader: %w", err)
	}

	return reader, nil
}

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

func (a *GCSAdapter) GetEndpoint() string {
	return fmt.Sprintf("https://storage.googleapis.com/%s", a.bucket)
}
