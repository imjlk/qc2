package clip

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveBackendPrefersFirstAvailableLinuxBackend(t *testing.T) {
	t.Parallel()

	lookupPath := func(name string) (string, error) {
		if name == "xclip" {
			return "/usr/bin/xclip", nil
		}
		return "", os.ErrNotExist
	}

	got, err := resolveBackend("linux", lookupPath)
	if err != nil {
		t.Fatalf("resolveBackend returned error: %v", err)
	}

	if got.name != "xclip" {
		t.Fatalf("resolveBackend() selected %q, want %q", got.name, "xclip")
	}
}

func TestResolveBackendReturnsUnavailableError(t *testing.T) {
	t.Parallel()

	_, err := resolveBackend("linux", func(string) (string, error) {
		return "", os.ErrNotExist
	})
	if err == nil {
		t.Fatal("resolveBackend() returned nil error, want unavailable error")
	}

	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("resolveBackend() error = %v, want ErrUnavailable", err)
	}
}

func TestCopyReturnsCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := (SystemClipboard{}).Copy(ctx, "qc2")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Copy() error = %v, want context.Canceled", err)
	}
}

func TestCopyExecWritesStdin(t *testing.T) {
	t.Parallel()

	outputPath := filepath.Join(t.TempDir(), "stdin.txt")
	lookupPath := func(name string) (string, error) {
		if name == "pbcopy" {
			return "pbcopy", nil
		}
		return "", os.ErrNotExist
	}
	var gotName string
	commandContext := func(ctx context.Context, name string, args ...string) *exec.Cmd {
		gotName = name
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestClipboardHelper$")
		cmd.Env = append(os.Environ(), "QC2_CLIP_HELPER=ok", "QC2_CLIP_OUT="+outputPath)
		return cmd
	}

	const text = "한글"
	err := copyExec(context.Background(), "darwin", text, lookupPath, commandContext)
	if err != nil {
		t.Fatalf("copyExec() error = %v", err)
	}
	if gotName != "pbcopy" {
		t.Fatalf("command = %q, want pbcopy", gotName)
	}
	got, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read helper output: %v", err)
	}
	if string(got) != text {
		t.Fatalf("clipboard stdin = %q, want %q", got, text)
	}
}

func TestCopyExecReturnsCommandOutput(t *testing.T) {
	t.Parallel()

	lookupPath := func(name string) (string, error) {
		if name == "pbcopy" {
			return "pbcopy", nil
		}
		return "", os.ErrNotExist
	}
	commandContext := func(ctx context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestClipboardHelper$")
		cmd.Env = append(os.Environ(), "QC2_CLIP_HELPER=fail")
		return cmd
	}

	err := copyExec(context.Background(), "darwin", "hello", lookupPath, commandContext)
	if err == nil || !strings.Contains(err.Error(), "backend failed") {
		t.Fatalf("copyExec() error = %v, want backend failed", err)
	}
}

func TestClipboardHelper(t *testing.T) {
	mode := os.Getenv("QC2_CLIP_HELPER")
	if mode == "" {
		return
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
	if mode == "fail" {
		os.Stderr.WriteString("backend failed")
		os.Exit(1)
	}
	if path := os.Getenv("QC2_CLIP_OUT"); path != "" {
		if err := os.WriteFile(path, data, 0o600); err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(1)
		}
	}
	os.Stdout.Write(data)
	os.Exit(0)
}
