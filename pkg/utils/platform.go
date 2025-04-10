package utils

import (
	"runtime"

	"github.com/ncobase/ncore/pkg/types"
)

// GetPlatformExt returns the platform-specific extension
func GetPlatformExt() string {
	switch runtime.GOOS {
	case "windows":
		return types.ExtWindows
	case "darwin":
		return types.ExtDarwin
	default:
		return types.ExtLinux
	}
}

func GetPlatform() string {
	switch runtime.GOOS {
	case "windows":
		return types.PlatformWindows
	case "darwin":
		return types.PlatformDarwin
	default:
		return types.PlatformLinux
	}
}
