package clip

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

var ErrUnavailable = errors.New("clipboard backend unavailable")

type Copier interface {
	Copy(ctx context.Context, text string) error
}

type SystemClipboard struct {
	LookupPath     func(string) (string, error)
	CommandContext func(context.Context, string, ...string) *exec.Cmd
}

type UnavailableError struct {
	GOOS       string
	Candidates []string
}

func (e *UnavailableError) Error() string {
	if len(e.Candidates) == 0 {
		return fmt.Sprintf("clipboard backend not found on %s; use --print instead", e.GOOS)
	}

	return fmt.Sprintf(
		"clipboard backend not found on %s; install %s, or use --print instead",
		e.GOOS,
		joinCandidates(e.Candidates),
	)
}

func (e *UnavailableError) Is(target error) bool {
	return target == ErrUnavailable
}

type backend struct {
	name string
	args []string
}

func (c SystemClipboard) Copy(ctx context.Context, text string) error {
	lookupPath := c.LookupPath
	if lookupPath == nil {
		lookupPath = exec.LookPath
	}

	commandContext := c.CommandContext
	if commandContext == nil {
		commandContext = exec.CommandContext
	}

	selected, err := resolveBackend(runtime.GOOS, lookupPath)
	if err != nil {
		return err
	}

	cmd := commandContext(ctx, selected.name, selected.args...)
	cmd.Stdin = strings.NewReader(text)

	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message != "" {
			return fmt.Errorf("copy to clipboard: %s", message)
		}
		return fmt.Errorf("copy to clipboard: %w", err)
	}

	return nil
}

func resolveBackend(goos string, lookupPath func(string) (string, error)) (backend, error) {
	candidates := backendsFor(goos)
	if len(candidates) == 0 {
		return backend{}, &UnavailableError{GOOS: goos}
	}

	for _, candidate := range candidates {
		if _, err := lookupPath(candidate.name); err == nil {
			return candidate, nil
		}
	}

	names := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		names = append(names, candidate.name)
	}

	return backend{}, &UnavailableError{
		GOOS:       goos,
		Candidates: names,
	}
}

func backendsFor(goos string) []backend {
	switch goos {
	case "darwin":
		return []backend{
			{name: "pbcopy"},
		}
	case "windows":
		return []backend{
			{name: "clip"},
		}
	case "linux":
		return []backend{
			{name: "wl-copy"},
			{name: "xclip", args: []string{"-selection", "clipboard"}},
			{name: "xsel", args: []string{"--clipboard", "--input"}},
		}
	default:
		return nil
	}
}

func joinCandidates(values []string) string {
	switch len(values) {
	case 0:
		return ""
	case 1:
		return values[0]
	case 2:
		return values[0] + " or " + values[1]
	default:
		return strings.Join(values[:len(values)-1], ", ") + ", or " + values[len(values)-1]
	}
}
