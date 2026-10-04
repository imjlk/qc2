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
	"time"
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

func TestCopyExecFallsBackWhenEarlierBackendFails(t *testing.T) {
	t.Parallel()

	outputPath := filepath.Join(t.TempDir(), "stdin.txt")
	lookupPath := func(name string) (string, error) {
		switch name {
		case "wl-copy", "xclip", "xsel":
			return "/usr/bin/" + name, nil
		default:
			return "", os.ErrNotExist
		}
	}
	var calls []string
	commandContext := func(ctx context.Context, name string, args ...string) *exec.Cmd {
		calls = append(calls, name)
		env := append(os.Environ(), "QC2_CLIP_HELPER=fail")
		if name == "xclip" {
			env = append(os.Environ(), "QC2_CLIP_HELPER=ok", "QC2_CLIP_OUT="+outputPath)
		}
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestClipboardHelper$")
		cmd.Env = env
		if name == "xclip" && strings.Join(args, " ") != "-selection clipboard" {
			t.Errorf("xclip args = %q, want %q", args, []string{"-selection", "clipboard"})
		}
		return cmd
	}

	const text = "qc2"
	err := copyExec(context.Background(), "linux", text, lookupPath, commandContext)
	if err != nil {
		t.Fatalf("copyExec() error = %v", err)
	}
	if strings.Join(calls, ",") != "wl-copy,xclip" {
		t.Fatalf("backends tried = %q, want wl-copy then xclip", calls)
	}
	got, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read helper output: %v", err)
	}
	if string(got) != text {
		t.Fatalf("clipboard stdin = %q, want %q", got, text)
	}
}

func TestCopyExecStopsAfterCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	lookupPath := func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	}
	var calls []string
	commandContext := func(ctx context.Context, name string, args ...string) *exec.Cmd {
		calls = append(calls, name)
		cancel()
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestClipboardHelper$")
		cmd.Env = append(os.Environ(), "QC2_CLIP_HELPER=fail")
		return cmd
	}

	err := copyExec(ctx, "linux", "qc2", lookupPath, commandContext)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("copyExec() error = %v, want context.Canceled", err)
	}
	if strings.Join(calls, ",") != "wl-copy" {
		t.Fatalf("backends tried = %q, want only wl-copy", calls)
	}
}

func TestCopyExecReturnsCanceledContextFromLastBackend(t *testing.T) {
	t.Parallel()

	ready := filepath.Join(t.TempDir(), "ready")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lookupPath := func(name string) (string, error) {
		if name == "xsel" {
			return "/usr/bin/xsel", nil
		}
		return "", os.ErrNotExist
	}
	commandContext := func(ctx context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestClipboardHelper$")
		cmd.Env = append(os.Environ(), "QC2_CLIP_HELPER=block", "QC2_CLIP_READY="+ready)
		return cmd
	}

	done := make(chan error, 1)
	go func() {
		done <- copyExec(ctx, "linux", "qc2", lookupPath, commandContext)
	}()

	deadline := time.Now().Add(2 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("clipboard helper did not start")
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("copyExec() error = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("copyExec() did not return after cancel")
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
	if mode == "block" {
		os.Stderr.WriteString("backend failed")
		if path := os.Getenv("QC2_CLIP_READY"); path != "" {
			if err := os.WriteFile(path, []byte("ready"), 0o600); err != nil {
				os.Stderr.WriteString(err.Error())
				os.Exit(1)
			}
		}
		time.Sleep(time.Hour)
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
