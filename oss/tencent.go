package oss

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// TencentAdapter implements the Interface for Tencent Cloud Object Storage (COS).
type TencentAdapter struct {
	client *cos.Client
	bucket string
	region string
	appID  string
}

// NewTencentAdapter creates a new Tencent COS storage adapter.
func NewTencentAdapter(secretID, secretKey, region, bucket, appID string) (*TencentAdapter, error) {
	bucketURL := fmt.Sprintf("https://%s-%s.cos.%s.myqcloud.com", bucket, appID, region)
	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bucket URL: %w", err)
	}

	baseURL := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})

	return &TencentAdapter{
		client: client,
		bucket: bucket,
		region: region,
		appID:  appID,
	}, nil
}

// Get downloads a file from Tencent COS to a temporary local file.
func (a *TencentAdapter) Get(path string) (*os.File, error) {
	reader, err := a.GetStream(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	ext := filepath.Ext(path)
	pattern := fmt.Sprintf("tencent-*%s", ext)
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

// GetStream returns a readable stream for the Tencent COS object.
func (a *TencentAdapter) GetStream(path string) (io.ReadCloser, error) {
	ctx := context.Background()

	resp, err := a.client.Object.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return resp.Body, nil
}

// Put uploads a file to Tencent COS from the given reader.
func (a *TencentAdapter) Put(path string, reader io.Reader) (*Object, error) {
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

	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	}

	_, err := a.client.Object.Put(ctx, path, reader, opt)
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

// Delete removes an object from the Tencent COS bucket.
func (a *TencentAdapter) Delete(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	ctx := context.Background()

	_, err := a.client.Object.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// List returns all objects under the specified prefix.
func (a *TencentAdapter) List(path string) ([]*Object, error) {
	ctx := context.Background()

	opt := &cos.BucketGetOptions{
		Prefix:  path,
		MaxKeys: 1000,
	}

	var objects []*Object
	isTruncated := true

	for isTruncated {
		result, _, err := a.client.Bucket.Get(ctx, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range result.Contents {
			lastModified, _ := time.Parse(time.RFC3339, obj.LastModified)
			objects = append(objects, &Object{
				Path:             obj.Key,
				Name:             filepath.Base(obj.Key),
				LastModified:     &lastModified,
				Size:             int64(obj.Size),
				StorageInterface: a,
			})
		}

		isTruncated = result.IsTruncated
		opt.Marker = result.NextMarker
	}

	return objects, nil
}

// GetURL generates a presigned URL valid for 1 hour.
func (a *TencentAdapter) GetURL(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	ctx := context.Background()

	presignedURL, err := a.client.Object.GetPresignedURL(ctx, http.MethodGet, path, a.client.GetCredential().SecretID, a.client.GetCredential().SecretKey, time.Hour, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

// GetEndpoint returns the Tencent COS endpoint URL.
func (a *TencentAdapter) GetEndpoint() string {
	return fmt.Sprintf("https://%s-%s.cos.%s.myqcloud.com", a.bucket, a.appID, a.region)
}

// Exists checks if an object exists in the Tencent COS bucket.
func (a *TencentAdapter) Exists(path string) (bool, error) {
	if path == "" {
		return false, fmt.Errorf("path cannot be empty")
	}

	ctx := context.Background()
	ok, err := a.client.Object.IsExist(ctx, path)
	if err != nil {
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}
	return ok, nil
}

// Stat retrieves object metadata without downloading content.
func (a *TencentAdapter) Stat(path string) (*Object, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	ctx := context.Background()
	resp, err := a.client.Object.Head(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}

	var size int64
	if resp.ContentLength > 0 {
		size = resp.ContentLength
	}

	lastModified, _ := time.Parse(time.RFC1123, resp.Header.Get("Last-Modified"))
	return &Object{
		Path:             path,
		Name:             filepath.Base(path),
		LastModified:     &lastModified,
		Size:             size,
		StorageInterface: a,
	}, nil
}

// tencentDriver implements the Driver interface for Tencent COS.
type tencentDriver struct{}

// Name returns the driver name.
func (d *tencentDriver) Name() string {
	return "tencent"
}

// Connect establishes a connection to Tencent COS.
func (d *tencentDriver) Connect(ctx context.Context, cfg *Config) (Interface, error) {
	return NewTencentAdapter(cfg.ID, cfg.Secret, cfg.Region, cfg.Bucket, cfg.AppID)
}

// Close closes the Tencent COS connection.
func (d *tencentDriver) Close(conn Interface) error {
	return nil
}

func init() {
	RegisterDriver(&tencentDriver{})
}
