package version

import (
	"encoding/json"
	"fmt"
)

// These variables are set during build time
var (
	// Version is the current version
	Version = "0.0.0"

	// Branch is current branch name the code is built off.
	Branch = "unknown"

	// Revision is the short commit hash of source tree
	Revision = "unknown"

	// BuiltAt is the build time
	BuiltAt = "unknown"

	// GoVersion is the go version used to build
	GoVersion = "unknown"
)

// Info contains version information
type Info struct {
	Version   string `json:"version"`
	Branch    string `json:"branch"`
	Revision  string `json:"revision"`
	BuiltAt   string `json:"builtAt"`
	GoVersion string `json:"goVersion"`
}

// GetVersionInfo returns version information
func GetVersionInfo() Info {
	return Info{
		Version:   Version,
		Branch:    Branch,
		Revision:  Revision,
		BuiltAt:   BuiltAt,
		GoVersion: GoVersion,
	}
}

// String returns a string representation of version information
func (i Info) String() string {
	return fmt.Sprintf("Version: %s\nBranch: %s\nRevision: %s\nBuilt At: %s\nGo Version: %s",
		i.Version, i.Branch, i.Revision, i.BuiltAt, i.GoVersion)
}

// JSON returns a JSON representation of version information
func (i Info) JSON() (string, error) {
	data, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Print prints version information to stdout
func Print() {
	fmt.Println(GetVersionInfo().String())
}
