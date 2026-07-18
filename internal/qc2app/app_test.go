package qc2app

import (
	"bytes"
	"context"
	"strings"
	"testing"
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
