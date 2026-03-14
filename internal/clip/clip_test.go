package clip

import (
	"errors"
	"os"
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
