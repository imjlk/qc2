package pathutil

import (
	"fmt"
	"os"
	"os/exec"
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

func TestFileURLEncodesSpaceAndKorean(t *testing.T) {
	var path, want string
	if runtime.GOOS == "windows" {
		path = `C:\tmp\qc2 dir\한글`
		want = "file:///C:/tmp/qc2%20dir/%ED%95%9C%EA%B8%80"
	} else {
		path = "/tmp/qc2 dir/한글"
		want = "file:///tmp/qc2%20dir/%ED%95%9C%EA%B8%80"
	}

	got, err := FileURL(path)
	if err != nil {
		t.Fatalf("FileURL returned error: %v", err)
	}
	if got != want {
		t.Fatalf("FileURL() = %q, want %q", got, want)
	}
}

func TestCurrentAbsUsesLogicalPWD(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("logical PWD follows the Unix shell convention")
	}

	tmp := t.TempDir()
	real := filepath.Join(tmp, "real")
	link := filepath.Join(tmp, "link")
	if err := os.Mkdir(real, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Symlink(real, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestCurrentAbsUsesLogicalPWDHelper$")
	cmd.Dir = link
	cmd.Env = append(os.Environ(), "QC2_PWD_CHECK=1", "PWD="+link)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helper failed: %v\n%s", err, out)
	}
	if strings.TrimSpace(string(out)) != link {
		t.Fatalf("CurrentAbs() = %q, want %q", strings.TrimSpace(string(out)), link)
	}
}

func TestCurrentAbsUsesLogicalPWDHelper(t *testing.T) {
	if os.Getenv("QC2_PWD_CHECK") != "1" {
		return
	}

	wd, err := CurrentAbs()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(wd)
	os.Exit(0)
}
