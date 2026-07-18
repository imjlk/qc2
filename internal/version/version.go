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
