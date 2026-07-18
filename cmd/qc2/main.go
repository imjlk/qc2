package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/imjlk/qc2/internal/qc2app"
)

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "qc2:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout io.Writer) error {
	app, err := qc2app.New(stdout)
	if err != nil {
		return err
	}
	return app.Run(ctx, args)
}
