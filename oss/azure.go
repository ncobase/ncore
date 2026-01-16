package oss

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
)

type AzureAdapter struct {
	client        *azblob.Client
	containerName string
	accountName   string
}

func NewAzureAdapter(accountName, accountKey, containerName string) (*AzureAdapter, error) {
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credentials: %w", err)
	}

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)
	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure client: %w", err)
	}

	return &AzureAdapter{
		client:        client,
		containerName: containerName,
		accountName:   accountName,
	}, nil
}

func (a *AzureAdapter) Get(path string) (*os.File, error) {
	reader, err := a.GetStream(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	ext := filepath.Ext(path)
	pattern := fmt.Sprintf("azure-*%s", ext)
	tmpFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := io.Copy(tmpFile, reader); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to copy blob to temp file: %w", err)
	}

	if _, err := tmpFile.Seek(0, 0); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to seek temp file: %w", err)
	}

	return tmpFile, nil
}

func (a *AzureAdapter) GetStream(path string) (io.ReadCloser, error) {
	ctx := context.Background()

	blobClient := a.client.ServiceClient().NewContainerClient(a.containerName).NewBlobClient(path)
	resp, err := blobClient.DownloadStream(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download blob: %w", err)
	}

	return resp.Body, nil
}

func (a *AzureAdapter) Put(path string, reader io.Reader) (*Object, error) {
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

	blobClient := a.client.ServiceClient().NewContainerClient(a.containerName).NewBlockBlobClient(path)

	_, err := blobClient.UploadStream(ctx, reader, &azblob.UploadStreamOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: &contentType,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload blob: %w", err)
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

func (a *AzureAdapter) Delete(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	ctx := context.Background()

	blobClient := a.client.ServiceClient().NewContainerClient(a.containerName).NewBlobClient(path)
	_, err := blobClient.Delete(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete blob: %w", err)
	}

	return nil
}

func (a *AzureAdapter) List(path string) ([]*Object, error) {
	ctx := context.Background()

	containerClient := a.client.ServiceClient().NewContainerClient(a.containerName)
	pager := containerClient.NewListBlobsFlatPager(&container.ListBlobsFlatOptions{
		Prefix: &path,
	})

	var objects []*Object
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list blobs: %w", err)
		}

		for _, blob := range page.Segment.BlobItems {
			if blob.Name == nil || blob.Properties == nil {
				continue
			}

			var size int64
			if blob.Properties.ContentLength != nil {
				size = *blob.Properties.ContentLength
			}

			objects = append(objects, &Object{
				Path:             *blob.Name,
				Name:             filepath.Base(*blob.Name),
				LastModified:     blob.Properties.LastModified,
				Size:             size,
				StorageInterface: a,
			})
		}
	}

	return objects, nil
}

func (a *AzureAdapter) GetURL(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	blobClient := a.client.ServiceClient().NewContainerClient(a.containerName).NewBlobClient(path)

	startsOn := time.Now().Add(-5 * time.Minute)
	expiresOn := time.Now().Add(1 * time.Hour)

	sasURL, err := blobClient.GetSASURL(sas.BlobPermissions{
		Read: true,
	}, expiresOn, &blob.GetSASURLOptions{
		StartTime: &startsOn,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate SAS URL: %w", err)
	}

	return sasURL, nil
}

func (a *AzureAdapter) GetEndpoint() string {
	return fmt.Sprintf("https://%s.blob.core.windows.net/%s", a.accountName, a.containerName)
}
