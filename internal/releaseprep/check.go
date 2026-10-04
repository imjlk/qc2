package releaseprep

import "strings"

// RequiresChangeset reports whether a pull request diff needs a new changeset.
// Tests, release tooling, and workflow files can merge without one.
func RequiresChangeset(paths []string) bool {
	for _, path := range paths {
		if userFacing(path) {
			return true
		}
	}
	return false
}

// AddsChangeset reports whether the diff adds a changeset file.
func AddsChangeset(paths []string) bool {
	for _, path := range paths {
		normalized := strings.ReplaceAll(path, "\\", "/")
		if strings.HasPrefix(normalized, ".sampo/changesets/") && strings.HasSuffix(normalized, ".md") {
			return true
		}
	}
	return false
}

func userFacing(path string) bool {
	normalized := strings.ReplaceAll(path, "\\", "/")
	if strings.HasSuffix(normalized, "_test.go") {
		return false
	}
	switch normalized {
	case "README.md", "scripts/install.sh", "scripts/install.ps1":
		return true
	}
	prefixes := []string{
		"cmd/",
		"internal/commands/",
		"internal/clip/",
		"internal/cli/",
		"internal/qc2app/",
		"internal/version/",
		"internal/pathutil/",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(normalized, prefix) {
			return true
		}
	}
	return false
}
