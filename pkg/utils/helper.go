package utils

import (
	"runtime"

	"ncore/pkg/types"
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
