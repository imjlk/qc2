package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/imjlk/qc2/internal/releaseprep"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "apply":
		apply(os.Args[2:])
	case "notes":
		notes(os.Args[2:])
	case "latest":
		latest(os.Args[2:])
	case "check":
		check(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func apply(args []string) {
	flags := flag.NewFlagSet("apply", flag.ExitOnError)
	root := flags.String("root", ".", "repository root")
	current := flags.String("current", "", "current version, without the leading v")
	date := flags.String("date", "", "release date, YYYY-MM-DD")
	if err := flags.Parse(args); err != nil {
		exitErr(err)
	}
	if *current == "" || *date == "" {
		exitErr(fmt.Errorf("apply requires --current and --date"))
	}

	entries, err := loadChangesets(*root)
	if err != nil {
		exitErr(err)
	}
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "no changesets")
		os.Exit(2)
	}
	changelogPath := filepath.Join(*root, "CHANGELOG.md")
	changelog, err := os.ReadFile(changelogPath)
	if err != nil {
		exitErr(err)
	}
	plan, err := releaseprep.Prepare(string(changelog), entries, *current, *date)
	if err != nil {
		exitErr(err)
	}
	if err := os.WriteFile(changelogPath, []byte(plan.Changelog), 0o644); err != nil {
		exitErr(err)
	}
	for _, name := range plan.Consumed {
		if err := os.Remove(filepath.Join(*root, ".sampo", "changesets", name)); err != nil {
			exitErr(err)
		}
	}
	fmt.Println(plan.Version)
}

func notes(args []string) {
	flags := flag.NewFlagSet("notes", flag.ExitOnError)
	root := flags.String("root", ".", "repository root")
	version := flags.String("version", "", "version whose notes are printed")
	if err := flags.Parse(args); err != nil {
		exitErr(err)
	}
	if *version == "" {
		exitErr(fmt.Errorf("notes requires --version"))
	}
	changelog, err := os.ReadFile(filepath.Join(*root, "CHANGELOG.md"))
	if err != nil {
		exitErr(err)
	}
	text, err := releaseprep.Notes(string(changelog), strings.TrimPrefix(*version, "v"))
	if err != nil {
		exitErr(err)
	}
	fmt.Print(text)
}

func latest(args []string) {
	flags := flag.NewFlagSet("latest", flag.ExitOnError)
	root := flags.String("root", ".", "repository root")
	if err := flags.Parse(args); err != nil {
		exitErr(err)
	}
	changelog, err := os.ReadFile(filepath.Join(*root, "CHANGELOG.md"))
	if err != nil {
		exitErr(err)
	}
	version, err := releaseprep.LatestVersion(string(changelog))
	if err != nil {
		exitErr(err)
	}
	fmt.Println(version)
}

func check(args []string) {
	flags := flag.NewFlagSet("check", flag.ExitOnError)
	base := flags.String("base", "", "git base revision")
	if err := flags.Parse(args); err != nil {
		exitErr(err)
	}
	if *base == "" {
		exitErr(fmt.Errorf("check requires --base"))
	}
	changed, err := gitDiff(*base, "ACMRD")
	if err != nil {
		exitErr(err)
	}
	added, err := gitDiff(*base, "A")
	if err != nil {
		exitErr(err)
	}
	if releaseprep.RequiresChangeset(changed) && !releaseprep.AddsChangeset(added) {
		exitErr(fmt.Errorf("user-facing changes need a new file in .sampo/changesets"))
	}
}

func loadChangesets(root string) ([]releaseprep.Entry, error) {
	configPath := filepath.Join(root, ".sampo", "config.toml")
	configText, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	tags, err := releaseprep.TagsFromConfig(string(configText))
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(root, ".sampo", "changesets")
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var entries []releaseprep.Entry
	for _, dirEntry := range dirEntries {
		name := dirEntry.Name()
		if dirEntry.IsDir() || !strings.HasSuffix(name, ".md") || strings.HasPrefix(name, ".") {
			continue
		}
		text, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		entry, err := releaseprep.ParseChangeset(name, string(text), tags)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func gitDiff(base, filter string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter="+filter, base+"...HEAD")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git diff %s: %w\n%s", base, err, out)
	}
	text := strings.TrimSpace(string(out))
	if text == "" {
		return nil, nil
	}
	return strings.Split(text, "\n"), nil
}

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "prepare-release: %s\n", err)
	os.Exit(1)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: prepare-release apply|notes|latest|check")
}
