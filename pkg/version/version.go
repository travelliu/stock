package version

import (
	"fmt"
	"runtime"
	"strings"
)

const unknown = "unknown"

var (
	// Version is the current program version (semver).
	version = "0.0.0-dev"

	// GitCommit is the git commit hash at build time.
	gitCommit = unknown
	// GitTag is the git tag at build time.
	gitTag = unknown
	// BuildTimestamp is the UTC build time.
	buildTimestamp = unknown
)

// BuildInfo describes compile-time information.
type BuildInfo struct {
	Version   string `json:"version,omitempty"`
	GitCommit string `json:"git_commit,omitempty"`
	GitTag    string `json:"git_tag,omitempty"`
	GoVersion string `json:"go_version,omitempty"`
}

// GetVersion returns the semver string.
func GetVersion() string {
	v := version
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	return v
}

// GetShortVersion returns "vX.Y.Z-<short-sha>".
func GetShortVersion() string {
	shortSha := gitCommit
	if len(shortSha) > 7 {
		shortSha = shortSha[:7]
	}
	if shortSha == unknown {
		return GetVersion()
	}
	return fmt.Sprintf("%s-%s", GetVersion(), shortSha)
}

// GetLongVersion returns multi-line version info.
func GetLongVersion() string {
	return fmt.Sprintf(
		"Version: %s\n"+
			"Git Tag: %s\n"+
			"Git Commit: %s\n"+
			"Build Time: %s\n"+
			"Go Version: %s\n",
		GetVersion(),
		gitTag,
		gitCommit,
		buildTimestamp,
		runtime.Version(),
	)
}

// GetBuildInfo returns structured build info.
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		Version:   GetVersion(),
		GitCommit: gitCommit,
		GitTag:    gitTag,
		GoVersion: runtime.Version(),
	}
}
