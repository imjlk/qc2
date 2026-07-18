package qc2app

import (
	"context"
	"io"

	"github.com/imjlk/qc2/internal/cli"
	"github.com/imjlk/qc2/internal/commands/cpwd"
	"github.com/imjlk/qc2/internal/version"
)

const (
	Name    = "qc2"
	Tagline = "Small CLI utilities for everyday workflows."
)

func New(stdout io.Writer) (*cli.App, error) {
	app := cli.NewApp(Name, Tagline, version.String(), stdout)
	deps := cpwd.DefaultDependencies(stdout)

	err := app.Register(cli.Command{
		Name:    cpwd.Name,
		Summary: cpwd.Summary,
		Usage:   cpwd.Usage,
		Run: func(ctx context.Context, args []string) error {
			return cpwd.Execute(ctx, args, deps)
		},
	})
	if err != nil {
		return nil, err
	}

	return app, nil
}
