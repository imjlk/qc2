package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestAppRunListIncludesRegisteredAndBuiltinCommands(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	app := NewApp("qc2", "Small CLI utilities for everyday workflows.", "dev", &stdout)
	if err := app.Register(Command{
		Name:    "cpwd",
		Summary: "Copy the current working directory to the clipboard.",
		Usage:   "cpwd usage\n",
		Run: func(context.Context, []string) error {
			return nil
		},
	}); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if err := app.Run(context.Background(), []string{"list"}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	output := stdout.String()
	for _, want := range []string{"cpwd\t", "help\t", "list\t", "version\t"} {
		if !strings.Contains(output, want) {
			t.Fatalf("list output = %q, want to contain %q", output, want)
		}
	}
}

func TestAppRunDispatchesRegisteredCommand(t *testing.T) {
	t.Parallel()

	app := NewApp("qc2", "Small CLI utilities for everyday workflows.", "dev", io.Discard)
	var gotArgs []string
	if err := app.Register(Command{
		Name:    "echo",
		Summary: "Echo test args.",
		Usage:   "echo usage\n",
		Run: func(_ context.Context, args []string) error {
			gotArgs = append([]string(nil), args...)
			return nil
		},
	}); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if err := app.Run(context.Background(), []string{"echo", "one", "two"}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if strings.Join(gotArgs, ",") != "one,two" {
		t.Fatalf("dispatched args = %v, want [one two]", gotArgs)
	}
}

func TestAppRunVersion(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	app := NewApp("qc2", "Small CLI utilities for everyday workflows.", "dev", &stdout)

	if err := app.Run(context.Background(), []string{"version"}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if stdout.String() != "qc2 dev\n" {
		t.Fatalf("version output = %q, want %q", stdout.String(), "qc2 dev\n")
	}
}

func TestAppRunReturnsUnknownCommandError(t *testing.T) {
	t.Parallel()

	app := NewApp("qc2", "Small CLI utilities for everyday workflows.", "dev", io.Discard)

	err := app.Run(context.Background(), []string{"missing"})
	if err == nil {
		t.Fatal("Run returned nil error, want unknown command error")
	}
}

func TestAppRegisterRejectsDuplicates(t *testing.T) {
	t.Parallel()

	app := NewApp("qc2", "Small CLI utilities for everyday workflows.", "dev", io.Discard)
	command := Command{
		Name:    "cpwd",
		Summary: "Copy the current working directory to the clipboard.",
		Usage:   "cpwd usage\n",
		Run: func(context.Context, []string) error {
			return nil
		},
	}

	if err := app.Register(command); err != nil {
		t.Fatalf("first Register returned error: %v", err)
	}

	if err := app.Register(command); err == nil {
		t.Fatal("second Register returned nil error, want duplicate registration error")
	}
}

func TestAppRunHelpForCommand(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	app := NewApp("qc2", "Small CLI utilities for everyday workflows.", "dev", &stdout)
	if err := app.Register(Command{
		Name:    "cpwd",
		Summary: "Copy the current working directory to the clipboard.",
		Usage:   "cpwd help\n",
		Run: func(context.Context, []string) error {
			return errors.New("should not run")
		},
	}); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if err := app.Run(context.Background(), []string{"help", "cpwd"}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if stdout.String() != "cpwd help\n" {
		t.Fatalf("help output = %q, want %q", stdout.String(), "cpwd help\n")
	}
}

func TestAppRunReturnsOutputError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("write failed")
	app := NewApp("qc2", "Small CLI utilities for everyday workflows.", "dev", errorWriter{err: wantErr})

	err := app.Run(context.Background(), []string{"version"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run error = %v, want %v", err, wantErr)
	}
}

type errorWriter struct {
	err error
}

func (w errorWriter) Write([]byte) (int, error) {
	return 0, w.err
}
