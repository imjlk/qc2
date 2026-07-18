package main

import (
	"context"
	"fmt"
	"os"

	"github.com/imjlk/qc2/internal/commands/cpwd"
)

func main() {
	deps := cpwd.DefaultDependencies(os.Stdout)
	if err := cpwd.Execute(context.Background(), os.Args[1:], deps); err != nil {
		fmt.Fprintln(os.Stderr, "cpwd:", err)
		os.Exit(1)
	}
}
