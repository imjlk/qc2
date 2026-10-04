package clip

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
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
	if err := ctx.Err(); err != nil {
		return err
	}
	return c.copy(ctx, text)
}

func copyExec(ctx context.Context, goos string, text string, lookupPath func(string) (string, error), commandContext func(context.Context, string, ...string) *exec.Cmd) error {
	available, err := availableBackends(goos, lookupPath)
	if err != nil {
		return err
	}

	var last error
	for _, selected := range available {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := runClipboardCommand(ctx, commandContext, selected, text)
		if err == nil {
			return nil
		}
		last = err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return last
}

func runClipboardCommand(ctx context.Context, commandContext func(context.Context, string, ...string) *exec.Cmd, selected backend, text string) error {
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
	available, err := availableBackends(goos, lookupPath)
	if err != nil {
		return backend{}, err
	}
	return available[0], nil
}

func availableBackends(goos string, lookupPath func(string) (string, error)) ([]backend, error) {
	candidates := backendsFor(goos)
	if len(candidates) == 0 {
		return nil, &UnavailableError{GOOS: goos}
	}

	available := make([]backend, 0, len(candidates))
	names := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		names = append(names, candidate.name)
		if _, err := lookupPath(candidate.name); err == nil {
			available = append(available, candidate)
		}
	}
	if len(available) == 0 {
		return nil, &UnavailableError{GOOS: goos, Candidates: names}
	}
	return available, nil
}

func backendsFor(goos string) []backend {
	switch goos {
	case "darwin":
		return []backend{
			{name: "pbcopy"},
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
