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

func TestAppHelpLayoutMatchesCurrentCommands(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	app := NewApp("qc2", "Small CLI utilities for everyday workflows.", "dev", &stdout)
	if err := app.Register(Command{
		Name:    "cpwd",
		Summary: "Copy the current working directory to the clipboard.",
		Usage:   "cpwd help\n",
		Run: func(context.Context, []string) error {
			return nil
		},
	}); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if err := app.Run(context.Background(), nil); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	const want = "" +
		"Usage:\n" +
		"  qc2 <command> [flags]\n" +
		"\n" +
		"Small CLI utilities for everyday workflows.\n" +
		"\n" +
		"Commands:\n" +
		"  cpwd     Copy the current working directory to the clipboard.\n" +
		"  list     List the bundled commands.\n" +
		"  version  Show the qc2 version.\n" +
		"  help     Show help for qc2 or a subcommand.\n"
	if stdout.String() != want {
		t.Fatalf("help output = %q, want %q", stdout.String(), want)
	}
}

func TestAppListOrder(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	app := NewApp("qc2", "Small CLI utilities for everyday workflows.", "dev", &stdout)
	if err := app.Register(Command{
		Name:    "cpwd",
		Summary: "Copy the current working directory to the clipboard.",
		Usage:   "cpwd help\n",
		Run: func(context.Context, []string) error {
			return nil
		},
	}); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if err := app.Run(context.Background(), []string{"list"}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	const want = "" +
		"cpwd\tCopy the current working directory to the clipboard.\n" +
		"help\tShow help for qc2 or a subcommand.\n" +
		"list\tList the bundled commands.\n" +
		"version\tShow the qc2 version.\n"
	if stdout.String() != want {
		t.Fatalf("list output = %q, want %q", stdout.String(), want)
	}
}

func TestAppHelpAlignsLongCommandName(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	app := NewApp("qc2", "tagline", "dev", &stdout)
	if err := app.Register(Command{
		Name:    "longcommand",
		Summary: "Do the long thing.",
		Usage:   "usage\n",
		Run: func(context.Context, []string) error {
			return nil
		},
	}); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if err := app.Run(context.Background(), nil); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	longColumn := summaryColumn(stdout.String(), "Do the long thing.")
	listColumn := summaryColumn(stdout.String(), "List the bundled commands.")
	if longColumn == -1 || longColumn != listColumn {
		t.Fatalf("summary columns = long %d, list %d", longColumn, listColumn)
	}
}

func TestAppHelpVersionAlias(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	app := NewApp("qc2", "tagline", "dev", &stdout)
	if err := app.Run(context.Background(), []string{"help", "--version"}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	const want = "Usage:\n  qc2 version\n\nShow the qc2 version.\n"
	if stdout.String() != want {
		t.Fatalf("help output = %q, want %q", stdout.String(), want)
	}
}

func TestAppRegisterRejectsReservedName(t *testing.T) {
	t.Parallel()

	app := NewApp("qc2", "tagline", "dev", io.Discard)
	err := app.Register(Command{
		Name:    "version",
		Summary: "Replace version.",
		Usage:   "usage\n",
		Run: func(context.Context, []string) error {
			return nil
		},
	})
	if err == nil {
		t.Fatal("Register returned nil error, want reserved command error")
	}
}

func summaryColumn(output, summary string) int {
	for _, line := range strings.Split(output, "\n") {
		column := strings.Index(line, summary)
		if column >= 0 {
			return column
		}
	}
	return -1
}
