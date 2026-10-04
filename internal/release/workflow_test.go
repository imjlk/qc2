package release

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseWorkflowShipsCmdDirectories(t *testing.T) {
	t.Parallel()

	workflowPath := filepath.Join("..", "..", ".github", "workflows", "release.yml")
	workflow, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read release workflow: %v", err)
	}
	text := strings.ReplaceAll(string(workflow), "\r\n", "\n")
	if strings.Contains(text, "for BIN_NAME in") {
		t.Fatal("release workflow hardcodes binary names")
	}
	if !strings.Contains(text, "for bin_path in cmd/*") {
		t.Fatal("release workflow should ship every directory in cmd/")
	}
	if !strings.Contains(text, "body_path: release-notes.md") {
		t.Fatal("release workflow should publish the changelog section")
	}
	if !strings.Contains(text, "workflow_dispatch:") {
		t.Fatal("release workflow should accept a dispatched tag from release prepare")
	}

	preparePath := filepath.Join("..", "..", ".github", "workflows", "release-prepare.yml")
	prepare, err := os.ReadFile(preparePath)
	if err != nil {
		t.Fatalf("read release prepare workflow: %v", err)
	}
	prepareText := strings.ReplaceAll(string(prepare), "\r\n", "\n")
	if !strings.Contains(prepareText, "workflows:\n      - ci") {
		t.Fatal("release prepare should run after ci on main")
	}
	if !strings.Contains(prepareText, "go run ./scripts/prepare-release.go apply") {
		t.Fatal("release prepare should consume changesets")
	}
	if !strings.Contains(prepareText, "gh workflow run release.yml --ref") {
		t.Fatal("release prepare should start the release workflow for the new tag")
	}
	if !strings.Contains(prepareText, "secrets.RELEASE_TOKEN || secrets.GITHUB_TOKEN") {
		t.Fatal("release prepare should prefer a maintainer token when one is configured")
	}
	if !strings.Contains(prepareText, "if [ -z \"${RELEASE_TOKEN:-}\" ]; then") {
		t.Fatal("release prepare should dispatch the release workflow only for the default token")
	}
}
