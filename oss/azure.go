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
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
)

// AzureAdapter implements the Interface for Azure Blob Storage.
type AzureAdapter struct {
	client        *azblob.Client
	containerName string
	accountName   string
}

// NewAzureAdapter creates a new Azure Blob Storage adapter.
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

// Get downloads a blob from Azure to a temporary local file.
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

// GetStream returns a readable stream for the Azure blob.
func (a *AzureAdapter) GetStream(path string) (io.ReadCloser, error) {
	ctx := context.Background()

	blobClient := a.client.ServiceClient().NewContainerClient(a.containerName).NewBlobClient(path)
	resp, err := blobClient.DownloadStream(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download blob: %w", err)
	}

	return resp.Body, nil
}

// Put uploads a file to Azure Blob Storage from the given reader.
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

// Delete removes a blob from the Azure container.
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

// List returns all blobs under the specified prefix.
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

// GetURL generates a SAS URL valid for 1 hour.
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

// GetEndpoint returns the Azure Blob Storage endpoint URL.
func (a *AzureAdapter) GetEndpoint() string {
	return fmt.Sprintf("https://%s.blob.core.windows.net/%s", a.accountName, a.containerName)
}

// Exists checks if a blob exists in the Azure container.
func (a *AzureAdapter) Exists(path string) (bool, error) {
	if path == "" {
		return false, fmt.Errorf("path cannot be empty")
	}

	ctx := context.Background()
	blobClient := a.client.ServiceClient().NewContainerClient(a.containerName).NewBlobClient(path)
	_, err := blobClient.GetProperties(ctx, nil)
	if err != nil {
		// Check if blob not found
		if bloberror.HasCode(err, bloberror.BlobNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check blob existence: %w", err)
	}
	return true, nil
}

// Stat retrieves blob metadata without downloading content.
func (a *AzureAdapter) Stat(path string) (*Object, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	ctx := context.Background()
	blobClient := a.client.ServiceClient().NewContainerClient(a.containerName).NewBlobClient(path)
	props, err := blobClient.GetProperties(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get blob metadata: %w", err)
	}

	var size int64
	if props.ContentLength != nil {
		size = *props.ContentLength
	}

	return &Object{
		Path:             path,
		Name:             filepath.Base(path),
		LastModified:     props.LastModified,
		Size:             size,
		StorageInterface: a,
	}, nil
}

// azureDriver implements the Driver interface for Azure Blob Storage.
type azureDriver struct{}

// Name returns the driver name.
func (d *azureDriver) Name() string {
	return "azure"
}

// Connect establishes a connection to Azure Blob Storage.
func (d *azureDriver) Connect(ctx context.Context, cfg *Config) (Interface, error) {
	return NewAzureAdapter(cfg.ID, cfg.Secret, cfg.Bucket)
}

// Close closes the Azure storage connection.
func (d *azureDriver) Close(conn Interface) error {
	return nil
}

func init() {
	RegisterDriver(&azureDriver{})
}
