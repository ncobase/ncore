package storage

import (
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
	ossObj, err := a.client.Put(path, reader)
	if err != nil {
		return nil, err
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
	return a.client.Delete(path)
}

// List lists all objects under current path
func (a *OSSAdapter) List(path string) ([]*Object, error) {
	ossObjects, err := a.client.List(path)
	if err != nil {
		return nil, err
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
	return a.client.GetURL(path)
}

// GetEndpoint gets endpoint
func (a *OSSAdapter) GetEndpoint() string {
	return a.client.GetEndpoint()
}
