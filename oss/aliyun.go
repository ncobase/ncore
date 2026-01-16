package oss

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

type AliyunAdapter struct {
	client *oss.Client
	bucket string
	region string
}

func NewAliyunAdapter(accessKeyID, secretAccessKey, region, bucket, endpoint string) (*AliyunAdapter, error) {
	provider := credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey)

	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(provider).
		WithRegion(region)

	if endpoint != "" {
		cfg = cfg.WithEndpoint(endpoint)
	}

	client := oss.NewClient(cfg)

	return &AliyunAdapter{
		client: client,
		bucket: bucket,
		region: region,
	}, nil
}

func (a *AliyunAdapter) Get(path string) (*os.File, error) {
	reader, err := a.GetStream(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	ext := filepath.Ext(path)
	pattern := fmt.Sprintf("aliyun-*%s", ext)
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

func (a *AliyunAdapter) GetStream(path string) (io.ReadCloser, error) {
	ctx := context.Background()

	result, err := a.client.GetObject(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr(a.bucket),
		Key:    oss.Ptr(path),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return result.Body, nil
}

func (a *AliyunAdapter) Put(path string, reader io.Reader) (*Object, error) {
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

	_, err := a.client.PutObject(ctx, &oss.PutObjectRequest{
		Bucket: oss.Ptr(a.bucket),
		Key:    oss.Ptr(path),
		Body:   reader,
		Metadata: map[string]string{
			"Content-Type": contentType,
		},
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

func (a *AliyunAdapter) Delete(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	ctx := context.Background()

	_, err := a.client.DeleteObject(ctx, &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(a.bucket),
		Key:    oss.Ptr(path),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

func (a *AliyunAdapter) List(path string) ([]*Object, error) {
	ctx := context.Background()

	paginator := a.client.NewListObjectsV2Paginator(&oss.ListObjectsV2Request{
		Bucket: oss.Ptr(a.bucket),
		Prefix: oss.Ptr(path),
	})

	var objects []*Object
	for paginator.HasNext() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			lastMod := oss.ToTime(obj.LastModified)
			objects = append(objects, &Object{
				Path:             oss.ToString(obj.Key),
				Name:             filepath.Base(oss.ToString(obj.Key)),
				LastModified:     &lastMod,
				Size:             obj.Size,
				StorageInterface: a,
			})
		}
	}

	return objects, nil
}

func (a *AliyunAdapter) GetURL(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	ctx := context.Background()

	presignResult, err := a.client.Presign(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr(a.bucket),
		Key:    oss.Ptr(path),
	}, func(po *oss.PresignOptions) {
		po.Expires = 1 * time.Hour
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignResult.URL, nil
}

func (a *AliyunAdapter) GetEndpoint() string {
	return fmt.Sprintf("https://%s.oss-%s.aliyuncs.com", a.bucket, a.region)
}
