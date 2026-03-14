package main

import (
	"context"
	"fmt"
	"os"

	"github.com/imjlk/qc2/internal/cli"
	"github.com/imjlk/qc2/internal/commands/cpwd"
	"github.com/imjlk/qc2/internal/version"
)

func main() {
	app, err := newApp()
	if err != nil {
		fmt.Fprintln(os.Stderr, "qc2:", err)
		os.Exit(1)
	}

	if err := app.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "qc2:", err)
		os.Exit(1)
	}
}

func newApp() (*cli.App, error) {
	deps := cpwd.DefaultDependencies()
	deps.Stdout = os.Stdout

	app := cli.NewApp("qc2", "Small CLI utilities for everyday workflows.", version.String(), os.Stdout)
	if err := app.Register(cpwd.NewCommand(deps)); err != nil {
		return nil, err
	}
	return app, nil
}
