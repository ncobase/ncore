package validator

import (
	"path/filepath"
	"strings"
)

var (
	imageExts = map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
		".ico":  true,
		".svg":  true,
		".tiff": true,
	}

	videoExts = map[string]bool{
		".mp4":  true,
		".avi":  true,
		".mkv":  true,
		".mov":  true,
		".webm": true,
	}

	audioExts = map[string]bool{
		".mp3":  true,
		".wav":  true,
		".ogg":  true,
		".m4a":  true,
		".flac": true,
	}

	documentExts = map[string]bool{
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".xls":  true,
		".xlsx": true,
		".ppt":  true,
		".pptx": true,
		".txt":  true,
	}
)

// IsImage verify is image
func IsImage(ext string) bool {
	if !strings.HasPrefix(ext, ".") {
		return false
	}
	return imageExts[ext]
}

// IsVideo verify is video
func IsVideo(ext string) bool {
	if !strings.HasPrefix(ext, ".") {
		return false
	}
	return videoExts[ext]
}

// IsAudio verify is audio
func IsAudio(ext string) bool {
	if !strings.HasPrefix(ext, ".") {
		return false
	}
	return audioExts[ext]
}

// IsDocument verify is document
func IsDocument(ext string) bool {
	if !strings.HasPrefix(ext, ".") {
		return false
	}
	return documentExts[ext]
}

// IsImageFile checks if a file is an image based on its name/extension
func IsImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return imageExts[ext]
}

// IsVideoFile checks if a file is a video based on its name/extension
func IsVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return videoExts[ext]
}

// IsAudioFile checks if a file is an audio based on its name/extension
func IsAudioFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return audioExts[ext]
}

// IsDocumentFile checks if a file is a document based on its name/extension
func IsDocumentFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return documentExts[ext]
}
