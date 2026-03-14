package pathutil

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func CurrentAbs() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get current directory: %w", err)
	}

	abs, err := filepath.Abs(wd)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path: %w", err)
	}

	return filepath.Clean(abs), nil
}

func FileURL(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path: %w", err)
	}

	normalized := filepath.ToSlash(filepath.Clean(abs))
	if runtime.GOOS == "windows" && !strings.HasPrefix(normalized, "/") {
		normalized = "/" + normalized
	}

	u := url.URL{
		Scheme: "file",
		Path:   normalized,
	}

	return u.String(), nil
}
