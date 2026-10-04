package releaseprep

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	categoryOrder  = []string{"Added", "Changed", "Deprecated", "Removed", "Fixed", "Security"}
	frontMatterRE  = regexp.MustCompile(`(?s)\A---\r?\n(.*?)\r?\n---\r?\n(.*)\z`)
	packageEntryRE = regexp.MustCompile(`^qc2:\s+(patch|minor|major)(?:\s+\(([^)]+)\))?$`)
	tagsLineRE     = regexp.MustCompile(`(?m)^tags\s*=\s*\[(.*)\]\s*$`)
)

// Entry is one user-facing note from a changeset file.
type Entry struct {
	Bump    string
	Tag     string
	Summary string
	Source  string
}

// TagsFromConfig reads the changeset tags from .sampo/config.toml.
func TagsFromConfig(text string) ([]string, error) {
	match := tagsLineRE.FindStringSubmatch(text)
	if match == nil {
		return nil, fmt.Errorf("sampo config: changeset tags are missing")
	}
	var tags []string
	for _, part := range strings.Split(match[1], ",") {
		tag := strings.Trim(strings.TrimSpace(part), `"`)
		if tag == "" {
			continue
		}
		tags = append(tags, tag)
	}
	if len(tags) == 0 {
		return nil, fmt.Errorf("sampo config: changeset tags are empty")
	}
	return tags, nil
}

// ParseChangeset reads a Sampo changeset that names the qc2 CLI.
func ParseChangeset(name, text string, allowed []string) (Entry, error) {
	match := frontMatterRE.FindStringSubmatch(strings.TrimSpace(text))
	if match == nil {
		return Entry{}, fmt.Errorf("changeset %s: front matter is missing", name)
	}

	var entry Entry
	for _, line := range strings.Split(match[1], "\n") {
		line = strings.TrimSpace(strings.TrimRight(line, "\r"))
		if line == "" {
			continue
		}
		if entry.Bump != "" {
			return Entry{}, fmt.Errorf("changeset %s: only qc2 can be listed", name)
		}
		parts := packageEntryRE.FindStringSubmatch(line)
		if parts == nil {
			return Entry{}, fmt.Errorf("changeset %s: want `qc2: patch|minor|major (Tag)`", name)
		}
		entry.Bump = parts[1]
		entry.Tag = parts[2]
	}
	if entry.Bump == "" {
		return Entry{}, fmt.Errorf("changeset %s: package qc2 is missing", name)
	}
	if entry.Tag != "" && !contains(allowed, entry.Tag) {
		return Entry{}, fmt.Errorf("changeset %s: tag %q is not configured", name, entry.Tag)
	}

	summary := strings.TrimSpace(match[2])
	if summary == "" {
		return Entry{}, fmt.Errorf("changeset %s: summary is empty", name)
	}
	entry.Summary = summary
	entry.Source = name
	return entry, nil
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
