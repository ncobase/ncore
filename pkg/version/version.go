package version

// These variables are set during build time
var (
	// Version is the current version
	Version = "0.0.0"

	// GitCommit is the git commit hash
	GitCommit = "unknown"

	// BuildTime is the build timestamp
	BuildTime = "unknown"

	// GoVersion is the go version used to build
	GoVersion = "unknown"
)

// Info contains version information
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildTime string `json:"buildTime"`
	GoVersion string `json:"goVersion"`
}

// GetVersionInfo returns version information
func GetVersionInfo() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
		GoVersion: GoVersion,
	}
}
