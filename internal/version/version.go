package version

import (
	"runtime/debug"
	"strings"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func String() string {
	return resolvedVersion()
}

// Report is the text shown by `qc2 version`. Release builds add the commit
// and build time injected with -ldflags. Local and go install builds omit
// those lines because they stay at their defaults.
func Report() string {
	var output strings.Builder
	output.WriteString(resolvedVersion())
	if Commit != "" && Commit != "none" {
		output.WriteString("\ncommit: ")
		output.WriteString(Commit)
	}
	if BuildDate != "" && BuildDate != "unknown" {
		output.WriteString("\nbuilt: ")
		output.WriteString(BuildDate)
	}
	return output.String()
}

func resolvedVersion() string {
	if Version != "dev" {
		return Version
	}

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return Version
	}

	return resolve(Version, buildInfo.Main.Version)
}

func resolve(fallback, moduleVersion string) string {
	if moduleVersion == "" || moduleVersion == "(devel)" {
		return fallback
	}

	return strings.TrimPrefix(moduleVersion, "v")
}
