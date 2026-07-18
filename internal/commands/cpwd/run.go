package cpwd

import (
	"context"
	"fmt"

	"github.com/imjlk/qc2/internal/pathutil"
)

type Options struct {
	FileURL bool
	Print   bool
	Quiet   bool
}

type Result struct {
	Value  string
	Copied bool
}

func Run(ctx context.Context, opts Options, deps Dependencies) (Result, error) {
	deps = deps.withDefaults()

	wd, err := deps.WorkingDir()
	if err != nil {
		return Result{}, err
	}

	value := wd
	if opts.FileURL {
		value, err = pathutil.FileURL(wd)
		if err != nil {
			return Result{}, fmt.Errorf("build file URL: %w", err)
		}
	}

	copied := false
	if !opts.Print {
		if err := deps.Clipboard.Copy(ctx, value); err != nil {
			return Result{}, err
		}
		copied = true
	}

	if opts.Print || !opts.Quiet {
		if _, err := fmt.Fprintln(deps.Stdout, value); err != nil {
			return Result{Value: value, Copied: copied}, fmt.Errorf("write output: %w", err)
		}
	}

	return Result{Value: value, Copied: copied}, nil
}
