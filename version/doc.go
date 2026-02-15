// Package version provides build-time version information and git metadata
// for applications built with ncore framework.
//
// This package exposes:
//   - Application version (from git tags)
//   - Git branch and commit information
//   - Build timestamp
//   - Go version used for compilation
//   - Version information retrieval and display
//
// # Version Variables
//
// These variables are set at build time using ldflags:
//
//	var (
//	    Version  = "0.0.0"     // Semantic version from git tags
//	    Branch   = "unknown"   // Git branch name
//	    Revision = "unknown"   // Git commit hash (short)
//	    BuiltAt  = "unknown"   // Build timestamp
//	    GoVersion = runtime.Version() // Go compiler version
//	)
//
// # Build Integration
//
// Set version information during build with ldflags:
//
//	go build -ldflags "\
//	  -X github.com/ncobase/ncore/version.Version=1.2.3 \
//	  -X github.com/ncobase/ncore/version.Branch=main \
//	  -X github.com/ncobase/ncore/version.Revision=abc123 \
//	  -X 'github.com/ncobase/ncore/version.BuiltAt=$(date)'"
//
// Or use the provided helper:
//
//	make build  // Uses Makefile with automatic version detection
//
// # Retrieving Version Info
//
// Get structured version information:
//
//	info := version.GetVersionInfo()
//	// Returns version.Info struct with all metadata
//
//	fmt.Printf("Version: %s\n", info.Version)
//	fmt.Printf("Branch: %s\n", info.Branch)
//	fmt.Printf("Revision: %s\n", info.Revision)
//	fmt.Printf("Built At: %s\n", info.BuiltAt)
//	fmt.Printf("Go Version: %s\n", info.GoVersion)
//
// # Display Formats
//
// Print version in different formats:
//
//	// Human-readable format
//	version.Print()
//	// Output:
//	// Version: 1.2.3
//	// Branch: main
//	// Revision: abc123
//	// Built At: 2024-01-15 10:30:00
//	// Go Version: go1.21.5
//
//	// String format
//	str := version.GetVersionInfo().String()
//
//	// JSON format
//	json := version.GetVersionInfo().JSON()
//	// Returns formatted JSON with all version info
//
// # Git Integration
//
// Automatic version detection from git:
//
//	// Detects version from git tags (v1.2.3)
//	// Falls back to commit hash if no tags
//	info := version.DetectFromGit()
//
// This requires git to be available in PATH and is primarily
// used during development builds.
//
// # CLI Integration
//
// Use with flag package for --version flag:
//
//	import (
//	    "flag"
//	    "github.com/ncobase/ncore/version"
//	)
//
//	var showVersion = flag.Bool("version", false, "Show version")
//
//	func main() {
//	    flag.Parse()
//	    if *showVersion {
//	        version.Print()
//	        os.Exit(0)
//	    }
//	    // Application code...
//	}
//
// # Best Practices
//
//   - Always set version info in production builds
//   - Use semantic versioning (MAJOR.MINOR.PATCH)
//   - Tag releases in git (v1.2.3 format)
//   - Include build timestamp for debugging
//   - Display version in logs on startup
//   - Expose version via API endpoint
//   - Use version for compatibility checks
package version
