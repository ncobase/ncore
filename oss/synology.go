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

type SynologyAdapter struct {
	client *minio.Client
	bucket string
}

func NewSynologyAdapter(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*SynologyAdapter, error) {
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Synology client: %w", err)
	}

	return &SynologyAdapter{
		client: client,
		bucket: bucket,
	}, nil
}

func (a *SynologyAdapter) Get(path string) (*os.File, error) {
	reader, err := a.GetStream(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	ext := filepath.Ext(path)
	pattern := fmt.Sprintf("synology-*%s", ext)
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

func (a *SynologyAdapter) GetStream(path string) (io.ReadCloser, error) {
	ctx := context.Background()
	object, err := a.client.GetObject(ctx, a.bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	return object, nil
}

func (a *SynologyAdapter) Put(path string, reader io.Reader) (*Object, error) {
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

func (a *SynologyAdapter) Delete(path string) error {
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

func (a *SynologyAdapter) List(path string) ([]*Object, error) {
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

func (a *SynologyAdapter) GetURL(path string) (string, error) {
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

func (a *SynologyAdapter) GetEndpoint() string {
	return a.client.EndpointURL().String()
}
