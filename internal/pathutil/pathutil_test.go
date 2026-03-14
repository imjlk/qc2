package pathutil

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFileURLUsesFileScheme(t *testing.T) {
	path := "."
	got, err := FileURL(path)
	if err != nil {
		t.Fatalf("FileURL returned error: %v", err)
	}

	if !strings.HasPrefix(got, "file://") {
		t.Fatalf("FileURL() = %q, want file:// prefix", got)
	}
}

func TestFileURLFromAbsolutePathOnCurrentOS(t *testing.T) {
	var path string
	if runtime.GOOS == "windows" {
		path = `C:\tmp\qc2`
	} else {
		path = filepath.Join(string(filepath.Separator), "tmp", "qc2")
	}

	got, err := FileURL(path)
	if err != nil {
		t.Fatalf("FileURL returned error: %v", err)
	}

	if runtime.GOOS == "windows" {
		if got != "file:///C:/tmp/qc2" {
			t.Fatalf("FileURL() = %q, want %q", got, "file:///C:/tmp/qc2")
		}
		return
	}

	if got != "file:///tmp/qc2" {
		t.Fatalf("FileURL() = %q, want %q", got, "file:///tmp/qc2")
	}
}
