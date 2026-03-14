package cli

import (
	"context"
	"fmt"
	"io"
	"sort"
)

type Command struct {
	Name    string
	Summary string
	Help    func(io.Writer)
	Run     func(context.Context, []string) error
}

type App struct {
	Name     string
	Tagline  string
	Version  string
	Stdout   io.Writer
	commands map[string]Command
}

func NewApp(name, tagline, version string, stdout io.Writer) *App {
	if stdout == nil {
		stdout = io.Discard
	}

	return &App{
		Name:     name,
		Tagline:  tagline,
		Version:  version,
		Stdout:   stdout,
		commands: make(map[string]Command),
	}
}

func (a *App) Register(command Command) error {
	if command.Name == "" {
		return fmt.Errorf("command name is required")
	}
	if command.Run == nil {
		return fmt.Errorf("command %q is missing a run handler", command.Name)
	}
	if _, exists := a.commands[command.Name]; exists {
		return fmt.Errorf("command %q is already registered", command.Name)
	}

	a.commands[command.Name] = command
	return nil
}

func (a *App) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		a.writeHelp()
		return nil
	}

	switch args[0] {
	case "help", "-h", "--help":
		return a.runHelp(args[1:])
	case "list":
		if len(args) > 1 {
			return fmt.Errorf("list does not accept arguments")
		}
		a.writeList()
		return nil
	case "version", "-v", "--version":
		if len(args) > 1 {
			return fmt.Errorf("version does not accept arguments")
		}
		_, err := fmt.Fprintf(a.Stdout, "%s %s\n", a.Name, a.Version)
		return err
	}

	command, ok := a.commands[args[0]]
	if !ok {
		return fmt.Errorf("unknown command %q; run `%s list`", args[0], a.Name)
	}

	return command.Run(ctx, args[1:])
}

func (a *App) runHelp(args []string) error {
	if len(args) == 0 {
		a.writeHelp()
		return nil
	}
	if len(args) > 1 {
		return fmt.Errorf("help accepts at most one command name")
	}

	switch args[0] {
	case "help":
		a.writeHelp()
		return nil
	case "list":
		_, err := fmt.Fprintf(a.Stdout, "Usage:\n  %s list\n\nList the bundled commands.\n", a.Name)
		return err
	case "version", "-v", "--version":
		_, err := fmt.Fprintf(a.Stdout, "Usage:\n  %s version\n\nShow the %s version.\n", a.Name, a.Name)
		return err
	}

	command, ok := a.commands[args[0]]
	if !ok {
		return fmt.Errorf("unknown command %q; run `%s list`", args[0], a.Name)
	}

	if command.Help == nil {
		return fmt.Errorf("command %q does not provide help output", command.Name)
	}

	command.Help(a.Stdout)
	return nil
}

func (a *App) writeHelp() {
	_, _ = fmt.Fprintf(a.Stdout, "Usage:\n  %s <command> [flags]\n\n", a.Name)
	_, _ = fmt.Fprintf(a.Stdout, "%s\n\n", a.Tagline)
	_, _ = io.WriteString(a.Stdout, "Commands:\n")

	for _, name := range a.sortedCommandNames() {
		command := a.commands[name]
		_, _ = fmt.Fprintf(a.Stdout, "  %-8s %s\n", command.Name, command.Summary)
	}

	_, _ = io.WriteString(a.Stdout, "  list     List the bundled commands.\n")
	_, _ = fmt.Fprintf(a.Stdout, "  version  Show the %s version.\n", a.Name)
	_, _ = fmt.Fprintf(a.Stdout, "  help     Show help for %s or a subcommand.\n", a.Name)
}

func (a *App) writeList() {
	for _, name := range a.sortedCommandNames() {
		command := a.commands[name]
		_, _ = fmt.Fprintf(a.Stdout, "%s\t%s\n", command.Name, command.Summary)
	}

	_, _ = fmt.Fprintf(a.Stdout, "help\tShow help for %s or a subcommand.\n", a.Name)
	_, _ = io.WriteString(a.Stdout, "list\tList the bundled commands.\n")
	_, _ = fmt.Fprintf(a.Stdout, "version\tShow the %s version.\n", a.Name)
}

func (a *App) sortedCommandNames() []string {
	names := make([]string, 0, len(a.commands))
	for name := range a.commands {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
