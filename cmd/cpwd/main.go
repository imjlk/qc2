package main

import (
	"context"
	"fmt"
	"os"

	"github.com/imjlk/qc2/internal/commands/cpwd"
)

func main() {
	command := cpwd.NewCommand(cpwd.DefaultDependencies())
	if err := command.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "cpwd:", err)
		os.Exit(1)
	}
}
