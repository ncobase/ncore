package storage

import (
	"fmt"
	"mime/multipart"
	"path"
	"strings"
	"time"

	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"
)

// FileType represents file type enum
type FileType string

// Supported file types
const (
	FileTypeFile    FileType = "file"
	FileTypeImage   FileType = "image"
	FileTypeVideo   FileType = "video"
	FileTypeAudio   FileType = "audio"
	FileTypeDoc     FileType = "doc"
	FileTypeArchive FileType = "archive"
	FileTypeOther   FileType = "other"
)

// FileHeader contains file metadata with improved path handling
type FileHeader struct {
	Name         string                `json:"name"`
	OriginalName string                `json:"original_name"`
	Size         int                   `json:"size"`
	Path         string                `json:"path"`
	Type         string                `json:"type"`
	Ext          string                `json:"ext"`
	Raw          *multipart.FileHeader `json:"raw"`
	Metadata     any                   `json:"metadata,omitempty"`
}

// GetFileHeader processes file header and generates unique file name with improved logic
func GetFileHeader(f *multipart.FileHeader, pathPrefix ...string) *FileHeader {
	if f == nil {
		return &FileHeader{
			Name:         "unknown",
			OriginalName: "unknown",
			Path:         "unknown",
			Type:         "application/octet-stream",
		}
	}

	file := &FileHeader{}
	fullName := path.Base(f.Filename)
	file.Ext = strings.ToLower(path.Ext(fullName))
	file.Size = int(f.Size)
	file.Type = f.Header.Get("Content-Type")
	file.Raw = f
	file.OriginalName = fullName

	// Clean original name without extension
	originalName := strings.TrimSuffix(fullName, file.Ext)
	if originalName == "" {
		originalName = "file"
	}

	// Clean filename for storage
	cleanName := cleanFileName(originalName)

	// Generate unique filename without path - path will be handled by service
	timestamp := time.Now().Unix()
	randomID := nanoid.Number(8)

	file.Name = fmt.Sprintf("%s_%d_%s", cleanName, timestamp, randomID)
	file.Path = fmt.Sprintf("%s%s", file.Name, file.Ext)

	// Add path prefix if provided (but this should be handled by service layer)
	if len(pathPrefix) > 0 && pathPrefix[0] != "" {
		file.Path = path.Join(pathPrefix[0], file.Path)
	}

	// Set metadata for images
	if validator.IsImage(file.Ext) {
		file.Metadata = map[string]any{
			"is_image": true,
			"ext":      file.Ext,
		}
	}

	return file
}

// cleanFileName cleans filename for safe storage
func cleanFileName(name string) string {
	// Replace unsafe characters
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")

	// Remove multiple underscores
	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}

	// Trim underscores from start and end
	name = strings.Trim(name, "_")

	if name == "" {
		name = "file"
	}

	return name
}

// RestoreOriginalFileName extracts original filename from unique filename
func RestoreOriginalFileName(uniqueName string, withExt ...bool) string {
	if strings.Contains(uniqueName, "/") {
		_, uniqueName = path.Split(uniqueName)
	}

	ext := path.Ext(uniqueName)
	nameWithoutExt := strings.TrimSuffix(uniqueName, ext)

	// Find the first underscore which separates original name from unique suffix
	parts := strings.Split(nameWithoutExt, "_")
	if len(parts) < 3 {
		// No unique suffix found, return as is
		if len(withExt) > 0 && withExt[0] {
			return uniqueName
		}
		return nameWithoutExt
	}

	// Take all parts except the last two (timestamp and random ID)
	originalParts := parts[:len(parts)-2]
	originalName := strings.Join(originalParts, "_")

	if len(withExt) > 0 && withExt[0] {
		return originalName + ext
	}
	return originalName
}

// GetFileType determines file type based on extension
func GetFileType(ext string) FileType {
	ext = strings.ToLower(ext)

	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".ico"}
	videoExts := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".mkv"}
	audioExts := []string{".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma"}
	docExts := []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt"}
	archiveExts := []string{".zip", ".rar", ".7z", ".tar", ".gz", ".bz2"}

	for _, e := range imageExts {
		if ext == e {
			return FileTypeImage
		}
	}
	for _, e := range videoExts {
		if ext == e {
			return FileTypeVideo
		}
	}
	for _, e := range audioExts {
		if ext == e {
			return FileTypeAudio
		}
	}
	for _, e := range docExts {
		if ext == e {
			return FileTypeDoc
		}
	}
	for _, e := range archiveExts {
		if ext == e {
			return FileTypeArchive
		}
	}

	return FileTypeOther
}

// FileConfig file upload configuration
type FileConfig struct {
	Path      string     `json:"path"`
	MaxSize   int64      `json:"max_size"`
	AllowType []FileType `json:"allow_type"`
}

// ValidateFile validates uploaded file against configuration
func (fc *FileConfig) ValidateFile(file *FileHeader) error {
	if fc.MaxSize > 0 && int64(file.Size) > fc.MaxSize {
		return fmt.Errorf("file size %d exceeds maximum %d", file.Size, fc.MaxSize)
	}

	if len(fc.AllowType) > 0 {
		fileType := GetFileType(file.Ext)
		allowed := false
		for _, allowedType := range fc.AllowType {
			if fileType == allowedType {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("file type %s is not allowed", fileType)
		}
	}

	return nil
}

// ImageConfig image-specific configuration
type ImageConfig struct {
	Path      string          `json:"path"`
	MaxSize   int64           `json:"max_size"`
	Thumbnail ThumbnailConfig `json:"thumbnail"`
}

// ThumbnailConfig thumbnail generation configuration
type ThumbnailConfig struct {
	Path      string `json:"path"`
	MaxWidth  int64  `json:"max_width"`
	MaxHeight int64  `json:"max_height"`
}
