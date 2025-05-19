package version

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"
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

// GetVersionInfo returns version information with runtime git info
func GetVersionInfo() Info {
	version := Version
	branch := Branch
	revision := Revision
	builtAt := BuiltAt

	// Only attempt to get runtime git info if we're using default values
	if version == "0.0.0" || branch == "unknown" || revision == "unknown" {
		if isGitAvailable() && isGitRepository() {
			runtimeBranch, runtimeRevision, runtimeVersion := getRuntimeGitInfo()

			// Only override defaults
			if version == "0.0.0" || version == "unknown" {
				version = runtimeVersion
			}
			if branch == "unknown" {
				branch = runtimeBranch
			}
			if revision == "unknown" {
				revision = runtimeRevision
			}
		} else {
			log.Println("WARNING: Unable to get git information at runtime - git may not be installed or this may not be a git repository")
		}
	}

	// Use current time if built time wasn't set
	if builtAt == "unknown" {
		builtAt = time.Now().Format(time.RFC3339)
	}

	return Info{
		Version:   version,
		Branch:    branch,
		Revision:  revision,
		BuiltAt:   builtAt,
		GoVersion: runtime.Version(),
	}
}

// isGitAvailable checks if git command is available
func isGitAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "--version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// isGitRepository checks if current directory is a git repository
func isGitRepository() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--is-inside-work-tree")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// getRuntimeGitInfo returns current git branch, revision and version tag with timeouts
func getRuntimeGitInfo() (branch, revision, version string) {
	branch = Branch
	revision = Revision
	version = Version

	// Use 2 second timeout for each command
	timeout := 2 * time.Second

	// Get current branch
	branchCtx, cancelBranch := context.WithTimeout(context.Background(), timeout)
	defer cancelBranch()
	branchCmd := exec.CommandContext(branchCtx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	branchOutput, err := branchCmd.Output()
	if err == nil {
		branch = strings.TrimSpace(string(branchOutput))
	} else {
		log.Printf("WARNING: Failed to get git branch: %v", err)
	}

	// Get current commit hash
	revCtx, cancelRev := context.WithTimeout(context.Background(), timeout)
	defer cancelRev()
	revCmd := exec.CommandContext(revCtx, "git", "rev-parse", "--short", "HEAD")
	revOutput, err := revCmd.Output()
	if err == nil {
		revision = strings.TrimSpace(string(revOutput))
	} else {
		log.Printf("WARNING: Failed to get git revision: %v", err)
	}

	// Get version using the suggested command
	versionCtx, cancelVersion := context.WithTimeout(context.Background(), timeout)
	defer cancelVersion()

	// Use bash to execute the complex command with pipe and sed
	versionCmd := exec.CommandContext(versionCtx, "bash", "-c",
		`git describe --tags --match "v*" --always | sed 's/-g[a-z0-9]\{7\}//'`)
	versionOutput, err := versionCmd.Output()
	if err == nil {
		version = strings.TrimSpace(string(versionOutput))
	} else {
		log.Printf("WARNING: Failed to get git version: %v", err)
		// Try a fallback for Windows or systems without bash/sed
		fallbackVersionCmd := exec.CommandContext(versionCtx, "git", "describe", "--tags", "--match", "v*", "--always")
		fallbackOutput, fallbackErr := fallbackVersionCmd.Output()
		if fallbackErr == nil {
			// Simple string parsing to remove the -gXXXXXXX part
			versionStr := strings.TrimSpace(string(fallbackOutput))
			parts := strings.Split(versionStr, "-g")
			if len(parts) > 1 {
				// Remove the hash part
				version = strings.Join(parts[:len(parts)-1], "-")
			} else {
				version = versionStr
			}
		} else {
			log.Printf("WARNING: Failed to get git version with fallback method: %v", fallbackErr)
		}
	}

	return
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
