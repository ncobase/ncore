package storage

import (
	"fmt"
	"mime/multipart"
	"path"
	"strings"
	"time"

	"github.com/ncobase/ncore/utils/nanoid"
)

// FileType file type
type FileType string

// supported file types
const (
	FileTypeFile    FileType = "file"
	FileTypeImage   FileType = "image"
	FileTypeVideo   FileType = "video"
	FileTypeAudio   FileType = "audio"
	FileTypeDoc     FileType = "doc"
	FileTypeArchive FileType = "archive"
	FileTypeOther   FileType = "other"
)

// FileHeader file header
type FileHeader struct {
	Name     string                `json:"name"`
	Size     int                   `json:"size"`
	Path     string                `json:"path"`
	Type     string                `json:"type"`
	Ext      string                `json:"ext"`
	Raw      *multipart.FileHeader `json:"raw"`
	Metadata any                   `json:"metadata,omitempty"`
}

// GetFileHeader gets file header data and handles file renaming to ensure uniqueness
func GetFileHeader(f *multipart.FileHeader, prefix ...string) *FileHeader {
	file := &FileHeader{}

	fullName := path.Base(f.Filename)
	file.Ext = strings.ToLower(path.Ext(fullName))

	file.Size = int(f.Size)
	file.Type = f.Header.Get("Content-Type")
	file.Raw = f

	// Generate unique file name
	originalName := fullName[0 : len(fullName)-len(file.Ext)]
	uniqueSuffix := fmt.Sprintf("%d-%s", time.Now().Unix(), nanoid.Number(8))
	file.Name = fmt.Sprintf("%s-%s", originalName, uniqueSuffix)
	file.Path = fmt.Sprintf("%s%s", file.Name, file.Ext)

	if len(prefix) > 0 && prefix[0] != "" {
		file.Path = path.Join(prefix[0], file.Path)
	}

	// TODO get image exif data
	if IsImage(file.Ext) {
		file.Metadata = map[string]any{}
	}

	return file
}

// RestoreOriginalFileName restores the original file name by removing the unique suffix.
// If withExt is true, it preserves the file extension.
func RestoreOriginalFileName(uniqueName string, withExt ...bool) string {
	if strings.Contains(uniqueName, "/") {
		_, uniqueName = path.Split(uniqueName)
	}
	ext := path.Ext(uniqueName)
	nameWithoutExt := strings.TrimSuffix(uniqueName, ext)

	// Find the second-to-last '-' character
	lastDashIndex := strings.LastIndex(nameWithoutExt, "-")
	if lastDashIndex == -1 {
		return uniqueName // No unique suffix found, return as is
	}

	secondLastDashIndex := strings.LastIndex(nameWithoutExt[:lastDashIndex], "-")
	if secondLastDashIndex == -1 {
		return uniqueName // Not enough dashes to remove suffix, return as is
	}

	// Extract the original name
	originalName := nameWithoutExt[:secondLastDashIndex]
	if len(withExt) > 0 && withExt[0] {
		return originalName + ext
	}
	return originalName
}

// FileConfig file config
type FileConfig struct {
	Path      string     `json:"path"`
	MaxSize   int64      `json:"max_size"`
	AllowType []FileType `json:"allow_type"`
}

// ImageConfig image config
type ImageConfig struct {
	Path      string          `json:"path"`
	MaxSize   int64           `json:"max_size"`
	Thumbnail ThumbnailConfig `json:"thumbnail"`
}

// ThumbnailConfig thumbnail config
type ThumbnailConfig struct {
	Path      string `json:"path"`
	MaxWidth  int64  `json:"max_width"`
	MaxHeight int64  `json:"max_height"`
}

// IsImage verify is image
func IsImage(ext string) bool {
	return strings.HasPrefix(ext, ".") && (ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".bmp" || ext == ".webp" || ext == ".ico" || ext == ".svg")
}

// IsVideo verify is video
func IsVideo(ext string) bool {
	return strings.HasPrefix(ext, ".") && (ext == ".mp4" || ext == ".avi" || ext == ".mkv" || ext == ".mov" || ext == ".webm")
}
