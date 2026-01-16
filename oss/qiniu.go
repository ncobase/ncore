package oss

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/storage"
	"net/http"
)

type QiniuAdapter struct {
	mac          *auth.Credentials
	bucket       string
	region       string
	domain       string
	bucketMgr    *storage.BucketManager
	uploadConfig *storage.Config
}

func NewQiniuAdapter(accessKey, secretKey, bucket, region, domain string) (*QiniuAdapter, error) {
	mac := auth.New(accessKey, secretKey)

	cfg := &storage.Config{
		UseHTTPS:      true,
		UseCdnDomains: false,
	}

	regionMap := map[string]*storage.Region{
		"cn-east-1":      &storage.ZoneHuadong,
		"cn-north-1":     &storage.ZoneHuabei,
		"cn-south-1":     &storage.ZoneHuanan,
		"us-west-1":      &storage.ZoneBeimei,
		"ap-southeast-1": &storage.ZoneXinjiapo,
	}

	if r, ok := regionMap[region]; ok {
		cfg.Region = r
	} else {
		cfg.Region = &storage.ZoneHuadong
	}

	bucketMgr := storage.NewBucketManager(mac, cfg)

	return &QiniuAdapter{
		mac:          mac,
		bucket:       bucket,
		region:       region,
		domain:       domain,
		bucketMgr:    bucketMgr,
		uploadConfig: cfg,
	}, nil
}

func (a *QiniuAdapter) Get(path string) (*os.File, error) {
	reader, err := a.GetStream(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	ext := filepath.Ext(path)
	pattern := fmt.Sprintf("qiniu-*%s", ext)
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

func (a *QiniuAdapter) GetStream(path string) (io.ReadCloser, error) {
	publicURL := storage.MakePublicURL(a.domain, path)
	privateURL := storage.MakePrivateURL(a.mac, publicURL, a.domain, 3600)

	resp, err := http.Get(privateURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to get object: status code %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func (a *QiniuAdapter) Put(path string, reader io.Reader) (*Object, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}
	if reader == nil {
		return nil, fmt.Errorf("reader cannot be nil")
	}

	putPolicy := storage.PutPolicy{
		Scope: fmt.Sprintf("%s:%s", a.bucket, path),
	}
	upToken := putPolicy.UploadToken(a.mac)

	formUploader := storage.NewFormUploader(a.uploadConfig)
	ret := storage.PutRet{}

	putExtra := storage.PutExtra{}
	if ext := filepath.Ext(path); ext != "" {
		if ct := getContentType(ext); ct != "" {
			putExtra.MimeType = ct
		}
	}

	err := formUploader.Put(context.Background(), &ret, upToken, path, reader, -1, &putExtra)
	if err != nil {
		return nil, fmt.Errorf("failed to upload object: %w", err)
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

func (a *QiniuAdapter) Delete(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	err := a.bucketMgr.Delete(a.bucket, path)
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

func (a *QiniuAdapter) List(path string) ([]*Object, error) {
	limit := 1000
	delimiter := ""
	marker := ""
	var objects []*Object

	for {
		entries, _, nextMarker, hasMore, err := a.bucketMgr.ListFiles(a.bucket, path, delimiter, marker, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, entry := range entries {
			putTime := time.Unix(0, entry.PutTime*100)
			objects = append(objects, &Object{
				Path:             entry.Key,
				Name:             filepath.Base(entry.Key),
				LastModified:     &putTime,
				Size:             entry.Fsize,
				StorageInterface: a,
			})
		}

		if !hasMore {
			break
		}
		marker = nextMarker
	}

	return objects, nil
}

func (a *QiniuAdapter) GetURL(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	publicURL := storage.MakePublicURL(a.domain, path)
	privateURL := storage.MakePrivateURL(a.mac, publicURL, a.domain, 3600)

	return privateURL, nil
}

func (a *QiniuAdapter) GetEndpoint() string {
	return a.domain
}
