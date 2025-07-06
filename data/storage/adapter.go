package storage

import (
	"fmt"
	"io"
	"os"

	"github.com/casdoor/oss"
)

// OSSAdapter adapts casdoor oss.StorageInterface to our Interface
type OSSAdapter struct {
	client oss.StorageInterface
}

// NewOSSAdapter creates a new OSS adapter
func NewOSSAdapter(client oss.StorageInterface) Interface {
	return &OSSAdapter{client: client}
}

// Get receives file with given path
func (a *OSSAdapter) Get(path string) (*os.File, error) {
	return a.client.Get(path)
}

// GetStream gets file as stream
func (a *OSSAdapter) GetStream(path string) (io.ReadCloser, error) {
	return a.client.GetStream(path)
}

// Put stores reader into given path
func (a *OSSAdapter) Put(path string, reader io.Reader) (*Object, error) {
	// Validate inputs
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}
	if reader == nil {
		return nil, fmt.Errorf("reader cannot be nil")
	}

	// Store the file
	ossObj, err := a.client.Put(path, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to store file to OSS: %w", err)
	}

	// Convert oss.Object to our Object
	return &Object{
		Path:             ossObj.Path,
		Name:             ossObj.Name,
		LastModified:     ossObj.LastModified,
		Size:             ossObj.Size,
		StorageInterface: a,
	}, nil
}

// Delete deletes file
func (a *OSSAdapter) Delete(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	err := a.client.Delete(path)
	if err != nil {
		return fmt.Errorf("failed to delete file from OSS: %w", err)
	}
	return nil
}

// List lists all objects under current path
func (a *OSSAdapter) List(path string) ([]*Object, error) {
	ossObjects, err := a.client.List(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects from OSS: %w", err)
	}

	// Convert []*oss.Object to []*Object
	objects := make([]*Object, len(ossObjects))
	for i, ossObj := range ossObjects {
		objects[i] = &Object{
			Path:             ossObj.Path,
			Name:             ossObj.Name,
			LastModified:     ossObj.LastModified,
			Size:             ossObj.Size,
			StorageInterface: a,
		}
	}

	return objects, nil
}

// GetURL gets public accessible URL
func (a *OSSAdapter) GetURL(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	url, err := a.client.GetURL(path)
	if err != nil {
		return "", fmt.Errorf("failed to get URL from OSS: %w", err)
	}
	return url, nil
}

// GetEndpoint gets endpoint
func (a *OSSAdapter) GetEndpoint() string {
	return a.client.GetEndpoint()
}
