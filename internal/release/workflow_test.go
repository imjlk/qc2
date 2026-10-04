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
	text := string(workflow)
	if strings.Contains(text, "for BIN_NAME in") {
		t.Fatal("release workflow hardcodes binary names")
	}
	if !strings.Contains(text, "for bin_path in cmd/*") {
		t.Fatal("release workflow should ship every directory in cmd/")
	}
}
