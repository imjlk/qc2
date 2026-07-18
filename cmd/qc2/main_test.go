package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunListsRegisteredCommands(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	if err := run(context.Background(), []string{"list"}, &stdout); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	if !strings.Contains(stdout.String(), "cpwd\t") {
		t.Fatalf("list output = %q, want cpwd command", stdout.String())
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	err := run(context.Background(), []string{"missing"}, &stdout)
	if err == nil {
		t.Fatal("run returned nil error, want unknown command error")
	}
	if !strings.Contains(err.Error(), "qc2 list") {
		t.Fatalf("run error = %q, want qc2 list guidance", err)
	}
}
