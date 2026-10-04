package releaseprep

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Plan is a changelog update that consumes changeset files.
type Plan struct {
	Version   string
	Changelog string
	Consumed  []string
}

// Prepare moves changeset notes into the next changelog version.
// Notes already listed under Unreleased are moved, not repeated.
func Prepare(changelog string, entries []Entry, currentVersion, date string) (Plan, error) {
	if len(entries) == 0 {
		return Plan{}, fmt.Errorf("no changesets")
	}
	version, err := nextVersion(currentVersion, entries)
	if err != nil {
		return Plan{}, err
	}

	prefix, unreleased, history, links, err := splitChangelog(changelog)
	if err != nil {
		return Plan{}, err
	}
	sections, err := parseSections(unreleased)
	if err != nil {
		return Plan{}, err
	}

	orderOf := map[string]int{}
	position := 0
	for _, category := range categoryOrder {
		for _, bullet := range sections[category] {
			orderOf[bullet] = position
			position++
		}
	}

	released := map[string][]placedNote{}
	var consumed []string
	for i, entry := range entries {
		category := entry.Tag
		order := 1000 + i
		if matched, _, ok := takeMatching(&sections, entry.Summary); ok {
			order = orderOf[entry.Summary]
			if category == "" {
				category = matched
			}
		}
		if category == "" {
			category = "Changed"
		}
		if !contains(categoryOrder, category) {
			return Plan{}, fmt.Errorf("changeset %s: category %q is unknown", entry.Source, category)
		}
		released[category] = append(released[category], placedNote{summary: entry.Summary, order: order})
		consumed = append(consumed, entry.Source)
	}
	releasedSections := newSections()
	for category, notes := range released {
		sort.SliceStable(notes, func(i, j int) bool {
			return notes[i].order < notes[j].order
		})
		for _, note := range notes {
			releasedSections[category] = append(releasedSections[category], note.summary)
		}
	}

	var b strings.Builder
	b.WriteString(prefix)
	b.WriteString("## [Unreleased]\n\n")
	b.WriteString(writeSections(sections))
	fmt.Fprintf(&b, "## [%s] - %s\n\n", version, date)
	b.WriteString(writeSections(releasedSections))
	b.WriteString(history)
	updatedLinks, err := updateLinks(links, currentVersion, version)
	if err != nil {
		return Plan{}, err
	}
	b.WriteString(updatedLinks)
	return Plan{Version: version, Changelog: b.String(), Consumed: consumed}, nil
}

// LatestVersion returns the newest released heading in the changelog.
func LatestVersion(changelog string) (string, error) {
	for _, line := range strings.Split(changelog, "\n") {
		version, ok := headingVersion(line)
		if ok && version != "Unreleased" {
			return version, nil
		}
	}
	return "", fmt.Errorf("changelog: released version is missing")
}

// Notes returns the changelog section for one version, without the heading or links.
func Notes(changelog, version string) (string, error) {
	lines := strings.Split(strings.ReplaceAll(changelog, "\r\n", "\n"), "\n")
	start := -1
	for i, line := range lines {
		got, ok := headingVersion(line)
		if ok && got == version {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return "", fmt.Errorf("changelog: version %s is missing", version)
	}
	end := len(lines)
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## [") || strings.HasPrefix(lines[i], "[") && strings.Contains(lines[i], "]: ") {
			end = i
			break
		}
	}
	body := strings.TrimSpace(strings.Join(lines[start:end], "\n"))
	if body == "" {
		return "", fmt.Errorf("changelog: version %s has no notes", version)
	}
	return body + "\n", nil
}

func nextVersion(current string, entries []Entry) (string, error) {
	parts, err := parseVersion(current)
	if err != nil {
		return "", err
	}
	bump := "patch"
	for _, entry := range entries {
		switch entry.Bump {
		case "major":
			bump = "major"
		case "minor":
			if bump != "major" {
				bump = "minor"
			}
		case "patch":
		default:
			return "", fmt.Errorf("changeset %s: bump %q is unknown", entry.Source, entry.Bump)
		}
	}
	switch bump {
	case "major":
		parts = [3]int{parts[0] + 1, 0, 0}
	case "minor":
		parts = [3]int{parts[0], parts[1] + 1, 0}
	default:
		parts[2]++
	}
	return fmt.Sprintf("%d.%d.%d", parts[0], parts[1], parts[2]), nil
}

func parseVersion(version string) ([3]int, error) {
	version = strings.TrimPrefix(version, "v")
	fields := strings.Split(version, ".")
	if len(fields) != 3 {
		return [3]int{}, fmt.Errorf("version %q is not major.minor.patch", version)
	}
	var parts [3]int
	for i, field := range fields {
		number, err := strconv.Atoi(field)
		if err != nil || number < 0 {
			return [3]int{}, fmt.Errorf("version %q is not major.minor.patch", version)
		}
		parts[i] = number
	}
	return parts, nil
}

func splitChangelog(changelog string) (prefix, unreleased, history, links string, err error) {
	text := strings.ReplaceAll(changelog, "\r\n", "\n")
	lines := strings.Split(text, "\n")
	unreleasedAt := -1
	historyAt := -1
	linksAt := -1
	for i, line := range lines {
		version, ok := headingVersion(line)
		if ok && version == "Unreleased" && unreleasedAt < 0 {
			unreleasedAt = i
			continue
		}
		if ok && unreleasedAt >= 0 && historyAt < 0 {
			historyAt = i
		}
		if linksAt < 0 && strings.HasPrefix(line, "[") && strings.Contains(line, "]: ") {
			linksAt = i
		}
	}
	if unreleasedAt < 0 || historyAt < 0 || linksAt < 0 || linksAt < historyAt {
		return "", "", "", "", fmt.Errorf("changelog: unreleased section or version links are missing")
	}
	prefix = joinBlock(lines[:unreleasedAt])
	unreleased = strings.Join(lines[unreleasedAt+1:historyAt], "\n")
	history = joinBlock(lines[historyAt:linksAt])
	links = strings.TrimRight(strings.Join(lines[linksAt:], "\n"), "\n") + "\n"
	return prefix, unreleased, history, links, nil
}

func joinBlock(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	text := strings.Join(lines, "\n")
	// Join drops the blank line represented by a trailing empty element.
	if lines[len(lines)-1] == "" {
		return text + "\n"
	}
	if !strings.HasSuffix(text, "\n") {
		return text + "\n"
	}
	return text
}

func headingVersion(line string) (string, bool) {
	if !strings.HasPrefix(line, "## [") {
		return "", false
	}
	rest := strings.TrimPrefix(line, "## [")
	end := strings.IndexByte(rest, ']')
	if end < 0 {
		return "", false
	}
	return rest[:end], true
}

func parseSections(body string) (map[string][]string, error) {
	sections := newSections()
	category := ""
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case line == "":
			continue
		case strings.HasPrefix(line, "### "):
			category = strings.TrimSpace(strings.TrimPrefix(line, "### "))
			if !contains(categoryOrder, category) {
				return nil, fmt.Errorf("changelog: category %q is unknown", category)
			}
		case strings.HasPrefix(line, "- "):
			if category == "" {
				return nil, fmt.Errorf("changelog: bullet %q is outside a category", line)
			}
			sections[category] = append(sections[category], strings.TrimPrefix(line, "- "))
		default:
			return nil, fmt.Errorf("changelog: unexpected unreleased line %q", line)
		}
	}
	return sections, nil
}

type placedNote struct {
	summary string
	order   int
}

func takeMatching(sections *map[string][]string, summary string) (string, int, bool) {
	for _, category := range categoryOrder {
		bullets := (*sections)[category]
		for i, bullet := range bullets {
			if bullet == summary {
				(*sections)[category] = append(bullets[:i], bullets[i+1:]...)
				return category, i, true
			}
		}
	}
	return "", 0, false
}

func newSections() map[string][]string {
	sections := make(map[string][]string, len(categoryOrder))
	for _, category := range categoryOrder {
		sections[category] = nil
	}
	return sections
}

func writeSections(sections map[string][]string) string {
	var b strings.Builder
	for _, category := range categoryOrder {
		bullets := sections[category]
		if len(bullets) == 0 {
			continue
		}
		fmt.Fprintf(&b, "### %s\n\n", category)
		for _, bullet := range bullets {
			b.WriteString("- ")
			b.WriteString(strings.ReplaceAll(bullet, "\n", "\n  "))
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func updateLinks(links, current, next string) (string, error) {
	lines := strings.Split(strings.TrimRight(links, "\n"), "\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "[Unreleased]: ") {
		return "", fmt.Errorf("changelog: unreleased compare link is missing")
	}
	old := "[Unreleased]: "
	marker := "/compare/v" + current + "...HEAD"
	if !strings.Contains(lines[0], marker) {
		return "", fmt.Errorf("changelog: unreleased link does not compare v%s", current)
	}
	repo := strings.TrimSuffix(lines[0], marker)
	repo = strings.TrimPrefix(repo, old)
	lines[0] = old + repo + "/compare/v" + next + "...HEAD"
	insert := fmt.Sprintf("[%s]: %s/compare/v%s...v%s", next, repo, current, next)
	updated := append([]string{lines[0], insert}, lines[1:]...)
	return strings.Join(updated, "\n") + "\n", nil
}
