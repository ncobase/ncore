package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileSystem represents the interface for file system storage
type FileSystem interface {
	GetFullPath(p string) string
	Get(p string) (*os.File, error)
	GetStream(p string) (io.ReadCloser, error)
	Put(p string, r io.Reader) (*Object, error)
	Delete(p string) error
	List(p string) ([]*Object, error)
	GetEndpoint() string
	GetURL(p string) (string, error)
}

// LocalFileSystem implements the FileSystem interface for local file system storage
type LocalFileSystem struct {
	Folder string
}

// NewFileSystem creates a new local file system storage
func NewFileSystem(folder string) *LocalFileSystem {
	if folder == "" {
		folder = "./uploads" // Default folder
	}

	abs, err := filepath.Abs(folder)
	if err != nil {
		panic(fmt.Sprintf("failed to get absolute path for folder %s: %v", folder, err))
	}

	if err := os.MkdirAll(abs, 0755); err != nil {
		panic(fmt.Sprintf("failed to create folder %s: %v", abs, err))
	}

	return &LocalFileSystem{Folder: abs}
}

// GetFullPath returns the full path from absolute/relative path
func (fs *LocalFileSystem) GetFullPath(p string) string {
	if p == "" {
		return fs.Folder
	}

	// Clean the path to prevent directory traversal
	p = filepath.Clean(p)

	// Prevent directory traversal attacks
	if strings.Contains(p, "..") {
		p = strings.ReplaceAll(p, "..", "")
	}

	if filepath.IsAbs(p) && strings.HasPrefix(p, fs.Folder) {
		return p
	}

	return filepath.Join(fs.Folder, p)
}

// Get receives a file with the given path
func (fs *LocalFileSystem) Get(p string) (*os.File, error) {
	if p == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	fullPath := fs.GetFullPath(p)

	// Check if file exists and is not a directory
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", p)
		}
		return nil, fmt.Errorf("failed to stat file %s: %w", p, err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file: %s", p)
	}

	return os.Open(fullPath)
}

// GetStream gets a file as a stream
func (fs *LocalFileSystem) GetStream(p string) (io.ReadCloser, error) {
	return fs.Get(p)
}

// Put stores the reader into the given path
func (fs *LocalFileSystem) Put(p string, r io.Reader) (*Object, error) {
	if p == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}
	if r == nil {
		return nil, fmt.Errorf("reader cannot be nil")
	}

	fullPath := fs.GetFullPath(p)

	// Ensure the path is a file, not a directory
	if strings.HasSuffix(p, "/") || strings.HasSuffix(fullPath, "/") {
		return nil, fmt.Errorf("path appears to be a directory, not a file: %s", p)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Check if fullPath is actually a directory (this prevents the original error)
	if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
		return nil, fmt.Errorf("target path is a directory, cannot create file: %s", fullPath)
	}

	// Create or truncate the file
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %s: %w", fullPath, err)
	}
	defer dst.Close()

	// Reset reader position if possible
	if seeker, ok := r.(io.ReadSeeker); ok {
		if _, err := seeker.Seek(0, 0); err != nil {
			return nil, fmt.Errorf("failed to seek reader: %w", err)
		}
	}

	// Copy data
	size, err := io.Copy(dst, r)
	if err != nil {
		// Clean up partial file on error
		os.Remove(fullPath)
		return nil, fmt.Errorf("failed to copy data to file: %w", err)
	}

	// Get file info for metadata
	info, err := dst.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	modTime := info.ModTime()
	return &Object{
		Path:             p,
		Name:             filepath.Base(p),
		LastModified:     &modTime,
		Size:             size,
		StorageInterface: fs,
	}, nil
}

// Delete deletes a file
func (fs *LocalFileSystem) Delete(p string) error {
	if p == "" {
		return fmt.Errorf("path cannot be empty")
	}

	fullPath := fs.GetFullPath(p)

	// Check if file exists
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // File already doesn't exist, consider it success
		}
		return fmt.Errorf("failed to stat file %s: %w", p, err)
	}

	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", p, err)
	}

	return nil
}

// List lists files
func (fs *LocalFileSystem) List(p string) ([]*Object, error) {
	var objects []*Object
	fullPath := fs.GetFullPath(p)

	// Check if path exists
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return objects, nil // Return empty list for non-existent paths
		}
		return nil, fmt.Errorf("failed to stat path %s: %w", p, err)
	}

	err := filepath.Walk(fullPath, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if currentPath == fullPath {
			return nil
		}

		// Only include files, not directories
		if !info.IsDir() {
			// Get relative path from base folder
			relPath, err := filepath.Rel(fs.Folder, currentPath)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}

			// Convert to forward slashes for consistency
			relPath = filepath.ToSlash(relPath)

			modTime := info.ModTime()
			objects = append(objects, &Object{
				Path:             relPath,
				Name:             info.Name(),
				LastModified:     &modTime,
				Size:             info.Size(),
				StorageInterface: fs,
			})
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", fullPath, err)
	}

	return objects, nil
}

// GetEndpoint gets the endpoint (for FileSystem, it's just the base path)
func (fs *LocalFileSystem) GetEndpoint() string {
	return fs.Folder
}

// GetURL gets the public accessible URL
func (fs *LocalFileSystem) GetURL(p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("path cannot be empty")
	}
	// For local filesystem, we just return the relative path
	// In a real application, you might want to return a full URL
	// with your web server's base URL
	return p, nil
}
