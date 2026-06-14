// Package version exposes release and build metadata for logging, CLI output, and ldflags injection.
package version

import (
	"runtime/debug"
)

// LDFLAGS: set via -X github.com/Olian04/go-app-template/cmd/echo/version.<Var>=<value> at link time.
var (
	Version   = "unknown"
	Revision  = "unknown"
	BuildTime = "unknown"
)

// VersionInfo holds resolved version strings (ldflags first, then runtime/debug build settings).
type VersionInfo struct {
	Version   string
	Revision  string
	BuildTime string
}

// Info returns version metadata for display and logging.
func Info() VersionInfo {
	return VersionInfo{
		Version:   resolveVersion(),
		Revision:  resolveRevision(),
		BuildTime: resolveBuildTime(),
	}
}

func resolveVersion() string {
	if Version != "unknown" {
		return Version
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, s := range bi.Settings {
			if s.Key == "vcs.tag" && s.Value != "" {
				return s.Value
			}
		}
		if bi.Main.Version != "" {
			return bi.Main.Version
		}
	}
	return "unknown"
}

func resolveRevision() string {
	if Revision != "unknown" {
		return Revision
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, s := range bi.Settings {
			if s.Key == "vcs.revision" && s.Value != "" {
				return s.Value
			}
		}
	}
	return "unknown"
}

func resolveBuildTime() string {
	if BuildTime != "unknown" {
		return BuildTime
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, s := range bi.Settings {
			if s.Key == "vcs.time" && s.Value != "" {
				return s.Value
			}
		}
	}
	return "unknown"
}
