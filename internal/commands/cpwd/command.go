package cpwd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/imjlk/qc2/internal/clip"
	"github.com/imjlk/qc2/internal/pathutil"
)

const (
	Name    = "cpwd"
	Summary = "Copy the current working directory to the clipboard."
	Usage   = `Usage:
  cpwd [flags]

Copy the current working directory to the clipboard.

Flags:
  -f, --file-url  Use a file:// URL instead of a plain absolute path.
  -p, --print     Print to stdout only and skip the clipboard.
  -q, --quiet     Suppress stdout on successful clipboard copy.
  -h, --help      Show this help text.
`
)

type Dependencies struct {
	Stdout     io.Writer
	Clipboard  clip.Copier
	WorkingDir func() (string, error)
}

func DefaultDependencies(stdout io.Writer) Dependencies {
	return Dependencies{
		Stdout:     stdout,
		Clipboard:  clip.SystemClipboard{},
		WorkingDir: pathutil.CurrentAbs,
	}
}

func Execute(ctx context.Context, args []string, deps Dependencies) error {
	deps = deps.withDefaults()

	opts, err := parseArgs(args)
	if errors.Is(err, flag.ErrHelp) {
		_, writeErr := io.WriteString(deps.Stdout, Usage)
		return writeErr
	}
	if err != nil {
		return err
	}

	_, err = Run(ctx, opts, deps)
	return err
}

func (d Dependencies) withDefaults() Dependencies {
	if d.Stdout == nil {
		d.Stdout = io.Discard
	}
	if d.Clipboard == nil {
		d.Clipboard = clip.SystemClipboard{}
	}
	if d.WorkingDir == nil {
		d.WorkingDir = pathutil.CurrentAbs
	}
	return d
}

func parseArgs(args []string) (Options, error) {
	var opts Options
	var help bool

	fs := flag.NewFlagSet(Name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&opts.FileURL, "file-url", false, "use a file:// URL instead of a plain absolute path")
	fs.BoolVar(&opts.FileURL, "f", false, "use a file:// URL instead of a plain absolute path")
	fs.BoolVar(&opts.Print, "print", false, "print to stdout only and skip the clipboard")
	fs.BoolVar(&opts.Print, "p", false, "print to stdout only and skip the clipboard")
	fs.BoolVar(&opts.Quiet, "quiet", false, "suppress stdout on successful clipboard copy")
	fs.BoolVar(&opts.Quiet, "q", false, "suppress stdout on successful clipboard copy")
	fs.BoolVar(&help, "help", false, "show help text")
	fs.BoolVar(&help, "h", false, "show help text")

	if err := fs.Parse(args); err != nil {
		return Options{}, normalizeParseError(err)
	}
	if help {
		return Options{}, flag.ErrHelp
	}
	if fs.NArg() > 0 {
		return Options{}, fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}

	return opts, nil
}

func normalizeParseError(err error) error {
	message := strings.TrimSpace(strings.TrimPrefix(err.Error(), "flag provided but not defined: "))
	if strings.HasPrefix(message, "-") {
		return fmt.Errorf("unknown flag: %s", message)
	}
	return err
}
