package releaseprep

import (
	"strings"
	"testing"
)

const sampleConfig = `
[changesets]
tags = ["Added", "Changed", "Deprecated", "Removed", "Fixed", "Security"]
`

const sampleChangelog = `# Changelog

Intro.

## [Unreleased]

### Changed

- ` + "`qc2 version` prints the commit and build time for release binaries." + `

### Fixed

- Copy non-ASCII text on Windows through the Unicode clipboard instead of ` + "`clip.exe`." + `
- Try the next Linux clipboard program when an earlier one fails to copy.

## [0.1.1] - 2026-07-19

### Fixed

- Report the module version from binaries installed with ` + "`go install`." + `

[Unreleased]: https://github.com/imjlk/qc2/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/imjlk/qc2/compare/v0.1.0...v0.1.1
`

func TestPrepareMovesExistingUnreleasedNotes(t *testing.T) {
	t.Parallel()

	tags, err := TagsFromConfig(sampleConfig)
	if err != nil {
		t.Fatalf("TagsFromConfig: %v", err)
	}
	entries := []Entry{}
	for _, source := range []struct {
		name string
		body string
	}{
		{"version.md", "---\nqc2: patch\n---\n\n`qc2 version` prints the commit and build time for release binaries.\n"},
		{"windows.md", "---\nqc2: patch\n---\n\nCopy non-ASCII text on Windows through the Unicode clipboard instead of `clip.exe`.\n"},
		{"linux.md", "---\nqc2: patch (Fixed)\n---\n\nTry the next Linux clipboard program when an earlier one fails to copy.\n"},
	} {
		entry, err := ParseChangeset(source.name, source.body, tags)
		if err != nil {
			t.Fatalf("ParseChangeset(%s): %v", source.name, err)
		}
		entries = append(entries, entry)
	}

	plan, err := Prepare(sampleChangelog, entries, "0.1.1", "2026-10-04")
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if plan.Version != "0.1.2" {
		t.Fatalf("version = %s, want 0.1.2", plan.Version)
	}
	if strings.Count(plan.Changelog, "Try the next Linux clipboard program") != 1 {
		t.Fatalf("linux note was duplicated:\n%s", plan.Changelog)
	}
	if strings.Contains(plan.Changelog, "## [Unreleased]\n\n### ") {
		t.Fatalf("unreleased notes were left behind:\n%s", plan.Changelog)
	}
	notes, err := Notes(plan.Changelog, "0.1.2")
	if err != nil {
		t.Fatalf("Notes: %v", err)
	}
	if !strings.Contains(notes, "### Changed\n\n- `qc2 version`") || !strings.Contains(notes, "### Fixed\n\n- Copy non-ASCII") {
		t.Fatalf("notes = %q", notes)
	}
	if !strings.Contains(plan.Changelog, "[Unreleased]: https://github.com/imjlk/qc2/compare/v0.1.2...HEAD") {
		t.Fatalf("unreleased link was not moved:\n%s", plan.Changelog)
	}
	if !strings.Contains(plan.Changelog, "[0.1.2]: https://github.com/imjlk/qc2/compare/v0.1.1...v0.1.2") {
		t.Fatalf("version link is missing:\n%s", plan.Changelog)
	}
	latest, err := LatestVersion(plan.Changelog)
	if err != nil || latest != "0.1.2" {
		t.Fatalf("latest = %s, %v", latest, err)
	}
}

func TestNextVersionUsesTheHighestBump(t *testing.T) {
	t.Parallel()

	version, err := nextVersion("0.1.1", []Entry{{Bump: "patch", Source: "a"}, {Bump: "minor", Source: "b"}})
	if err != nil {
		t.Fatalf("nextVersion: %v", err)
	}
	if version != "0.2.0" {
		t.Fatalf("version = %s, want 0.2.0", version)
	}
}

func TestParseChangesetRejectsAnotherPackage(t *testing.T) {
	t.Parallel()

	_, err := ParseChangeset("bad.md", "---\nnpm/qc2: patch\n---\n\nNope.\n", categoryOrder)
	if err == nil {
		t.Fatal("ParseChangeset accepted a package other than qc2")
	}
}

func TestRequiresChangesetIgnoresTestsAndReleaseTooling(t *testing.T) {
	t.Parallel()

	if RequiresChangeset([]string{"internal/clip/clip_test.go", ".github/workflows/release.yml", "internal/releaseprep/check.go"}) {
		t.Fatal("tests and release tooling required a changeset")
	}
	if !RequiresChangeset([]string{"internal/clip/clip.go"}) {
		t.Fatal("clipboard change did not require a changeset")
	}
	if !AddsChangeset([]string{".sampo/changesets/windows-unicode-clipboard.md"}) {
		t.Fatal("added changeset was not detected")
	}
}
