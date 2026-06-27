package utils

import (
	"runtime"
	"testing"
)

func TestPlatformHelpers(t *testing.T) {
	ext := GetPlatformExt()
	platform := GetPlatform()

	if runtime.GOOS == "windows" {
		if ext == "" || platform == "" {
			t.Fatalf("expected non-empty platform metadata on windows, got ext=%q platform=%q", ext, platform)
		}
		return
	}

	if runtime.GOOS == "darwin" {
		if ext == "" || platform == "" {
			t.Fatalf("expected non-empty platform metadata on darwin, got ext=%q platform=%q", ext, platform)
		}
		return
	}

	if ext == "" || platform == "" {
		t.Fatalf("expected non-empty platform metadata on linux-like system, got ext=%q platform=%q", ext, platform)
	}
}
