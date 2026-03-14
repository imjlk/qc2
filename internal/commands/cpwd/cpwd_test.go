package cpwd

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/imjlk/qc2/internal/pathutil"
)

type fakeClipboard struct {
	values []string
	err    error
}

func (f *fakeClipboard) Copy(_ context.Context, text string) error {
	if f.err != nil {
		return f.err
	}

	f.values = append(f.values, text)
	return nil
}

func TestRunCopiesAndPrintsByDefault(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	clipboard := &fakeClipboard{}

	result, err := Run(context.Background(), Options{}, Dependencies{
		Stdout:    &stdout,
		Clipboard: clipboard,
		WorkingDir: func() (string, error) {
			return "/tmp/qc2", nil
		},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if !result.Copied {
		t.Fatalf("Run() Copied = false, want true")
	}

	if len(clipboard.values) != 1 || clipboard.values[0] != "/tmp/qc2" {
		t.Fatalf("clipboard values = %#v, want %q", clipboard.values, "/tmp/qc2")
	}

	if stdout.String() != "/tmp/qc2\n" {
		t.Fatalf("stdout = %q, want %q", stdout.String(), "/tmp/qc2\n")
	}
}

func TestRunPrintOnlySkipsClipboard(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	clipboard := &fakeClipboard{}

	result, err := Run(context.Background(), Options{Print: true}, Dependencies{
		Stdout:    &stdout,
		Clipboard: clipboard,
		WorkingDir: func() (string, error) {
			return "/tmp/qc2", nil
		},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if result.Copied {
		t.Fatalf("Run() Copied = true, want false")
	}

	if len(clipboard.values) != 0 {
		t.Fatalf("clipboard values = %#v, want no clipboard writes", clipboard.values)
	}

	if stdout.String() != "/tmp/qc2\n" {
		t.Fatalf("stdout = %q, want %q", stdout.String(), "/tmp/qc2\n")
	}
}

func TestRunBuildsFileURL(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	clipboard := &fakeClipboard{}
	want, err := pathutil.FileURL("/tmp/qc2")
	if err != nil {
		t.Fatalf("FileURL returned error: %v", err)
	}

	result, err := Run(context.Background(), Options{FileURL: true, Print: true}, Dependencies{
		Stdout:    &stdout,
		Clipboard: clipboard,
		WorkingDir: func() (string, error) {
			return "/tmp/qc2", nil
		},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if result.Value != want {
		t.Fatalf("Run() Value = %q, want %q", result.Value, want)
	}

	if stdout.String() != want+"\n" {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want+"\n")
	}
}

func TestRunQuietSuppressesOutput(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	clipboard := &fakeClipboard{}

	result, err := Run(context.Background(), Options{Quiet: true}, Dependencies{
		Stdout:    &stdout,
		Clipboard: clipboard,
		WorkingDir: func() (string, error) {
			return "/tmp/qc2", nil
		},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if !result.Copied {
		t.Fatalf("Run() Copied = false, want true")
	}

	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty output", stdout.String())
	}
}

func TestRunReturnsClipboardError(t *testing.T) {
	t.Parallel()

	clipboard := &fakeClipboard{err: errors.New("clipboard failed")}

	_, err := Run(context.Background(), Options{}, Dependencies{
		Clipboard: clipboard,
		WorkingDir: func() (string, error) {
			return "/tmp/qc2", nil
		},
	})
	if err == nil {
		t.Fatal("Run() returned nil error, want clipboard error")
	}
}

func TestExecuteHelp(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer

	err := Execute(context.Background(), []string{"--help"}, Dependencies{
		Stdout: &stdout,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if stdout.Len() == 0 {
		t.Fatal("Execute() help output is empty")
	}
}

func TestExecuteShortFlags(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	clipboard := &fakeClipboard{}
	want, err := pathutil.FileURL("/tmp/qc2")
	if err != nil {
		t.Fatalf("FileURL returned error: %v", err)
	}

	err = Execute(context.Background(), []string{"-f", "-p"}, Dependencies{
		Stdout:    &stdout,
		Clipboard: clipboard,
		WorkingDir: func() (string, error) {
			return "/tmp/qc2", nil
		},
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if len(clipboard.values) != 0 {
		t.Fatalf("clipboard values = %#v, want no clipboard writes", clipboard.values)
	}

	if stdout.String() != want+"\n" {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want+"\n")
	}
}
