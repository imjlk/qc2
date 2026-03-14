package cpwd

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/imjlk/qc2/internal/cli"
	"github.com/imjlk/qc2/internal/clip"
	"github.com/imjlk/qc2/internal/pathutil"
)

const (
	Name    = "cpwd"
	Summary = "Copy the current working directory to the clipboard."
)

type Options struct {
	FileURL bool
	Print   bool
	Quiet   bool
}

type Dependencies struct {
	Stdout     io.Writer
	Clipboard  clip.Copier
	WorkingDir func() (string, error)
}

type Result struct {
	Value  string
	Copied bool
}

func DefaultDependencies() Dependencies {
	return Dependencies{
		Stdout:     os.Stdout,
		Clipboard:  clip.SystemClipboard{},
		WorkingDir: pathutil.CurrentAbs,
	}
}

func NewCommand(deps Dependencies) cli.Command {
	deps = deps.withDefaults()

	return cli.Command{
		Name:    Name,
		Summary: Summary,
		Help:    Help,
		Run: func(ctx context.Context, args []string) error {
			return Execute(ctx, args, deps)
		},
	}
}

func Execute(ctx context.Context, args []string, deps Dependencies) error {
	deps = deps.withDefaults()

	opts, err := parseArgs(args)
	if err != nil {
		if err == flag.ErrHelp {
			Help(deps.Stdout)
			return nil
		}
		return err
	}

	_, err = Run(ctx, opts, deps)
	return err
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

	return Result{
		Value:  value,
		Copied: copied,
	}, nil
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

func Help(w io.Writer) {
	_, _ = io.WriteString(w, `Usage:
  cpwd [flags]

Copy the current working directory to the clipboard.

Flags:
  -f, --file-url  Use a file:// URL instead of a plain absolute path.
  -p, --print     Print to stdout only and skip the clipboard.
  -q, --quiet     Suppress stdout on successful clipboard copy.
  -h, --help      Show this help text.
`)
}

func normalizeParseError(err error) error {
	if err == nil {
		return nil
	}

	message := err.Error()
	message = strings.TrimPrefix(message, "flag provided but not defined: ")
	message = strings.TrimSpace(message)
	if strings.HasPrefix(message, "-") {
		return fmt.Errorf("unknown flag: %s", message)
	}

	return err
}
