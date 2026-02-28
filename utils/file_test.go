package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDirAndWriteTemplateFile(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "nested")
	targetFile := filepath.Join(targetDir, "hello.txt")

	if err := EnsureDir(targetDir); err != nil {
		t.Fatalf("EnsureDir() error = %v", err)
	}

	if err := WriteTemplateFile(targetFile, "hello {{.Name}}", map[string]string{"Name": "ncore"}); err != nil {
		t.Fatalf("WriteTemplateFile() error = %v", err)
	}

	content, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "hello ncore" {
		t.Fatalf("unexpected rendered content: %q", string(content))
	}
}

func TestPathFileDirExists(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "file.txt")

	if ok, err := PathExists(targetFile); err != nil || ok {
		t.Fatalf("PathExists() expected false,nil for missing file, got %v,%v", ok, err)
	}
	if FileExists(targetFile) {
		t.Fatal("FileExists() expected false for missing file")
	}
	if !DirExists(tmpDir) {
		t.Fatal("DirExists() expected true for temp directory")
	}

	if err := os.WriteFile(targetFile, []byte("ok"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if ok, err := PathExists(targetFile); err != nil || !ok {
		t.Fatalf("PathExists() expected true,nil for existing file, got %v,%v", ok, err)
	}
	if !FileExists(targetFile) {
		t.Fatal("FileExists() expected true for existing file")
	}
	if DirExists(targetFile) {
		t.Fatal("DirExists() expected false for file path")
	}
}
