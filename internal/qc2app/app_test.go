package qc2app

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/imjlk/qc2/internal/version"
)

func TestNewRegistersCPWD(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	app, err := New(&stdout)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if err := app.Run(context.Background(), []string{"list"}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !strings.Contains(stdout.String(), "cpwd\t") {
		t.Fatalf("list output = %q, want cpwd command", stdout.String())
	}
}

func TestVersionIncludesBuildDetails(t *testing.T) {
	originalVersion := version.Version
	originalCommit := version.Commit
	originalDate := version.BuildDate
	version.Version = "1.2.3"
	version.Commit = "abc123"
	version.BuildDate = "2026-07-19T00:00:00Z"
	t.Cleanup(func() {
		version.Version = originalVersion
		version.Commit = originalCommit
		version.BuildDate = originalDate
	})

	var stdout bytes.Buffer
	app, err := New(&stdout)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if err := app.Run(context.Background(), []string{"version"}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	const want = "qc2 1.2.3\ncommit: abc123\nbuilt: 2026-07-19T00:00:00Z\n"
	if stdout.String() != want {
		t.Fatalf("version output = %q, want %q", stdout.String(), want)
	}
}
