package cli

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
)

// App dispatches registered commands and provides the shared built-in commands.
type App struct {
	name     string
	tagline  string
	version  string
	stdout   io.Writer
	commands map[string]Command
}

func NewApp(name, tagline, version string, stdout io.Writer) *App {
	if stdout == nil {
		stdout = io.Discard
	}

	return &App{
		name:     name,
		tagline:  tagline,
		version:  version,
		stdout:   stdout,
		commands: make(map[string]Command),
	}
}

func (a *App) Register(command Command) error {
	if command.Name == "" {
		return fmt.Errorf("command name is required")
	}
	if command.Usage == "" {
		return fmt.Errorf("command %q is missing usage text", command.Name)
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
		return a.writeHelp()
	}

	switch args[0] {
	case "help", "-h", "--help":
		return a.runHelp(args[1:])
	case "list":
		if len(args) > 1 {
			return fmt.Errorf("list does not accept arguments")
		}
		return a.writeList()
	case "version", "-v", "--version":
		if len(args) > 1 {
			return fmt.Errorf("version does not accept arguments")
		}
		return writeString(a.stdout, fmt.Sprintf("%s %s\n", a.name, a.version))
	}

	command, ok := a.commands[args[0]]
	if !ok {
		return fmt.Errorf("unknown command %q; run `%s list`", args[0], a.name)
	}

	return command.Run(ctx, args[1:])
}

func (a *App) runHelp(args []string) error {
	if len(args) == 0 {
		return a.writeHelp()
	}
	if len(args) > 1 {
		return fmt.Errorf("help accepts at most one command name")
	}

	switch args[0] {
	case "help":
		return a.writeHelp()
	case "list":
		return writeString(a.stdout, fmt.Sprintf("Usage:\n  %s list\n\nList the bundled commands.\n", a.name))
	case "version", "-v", "--version":
		return writeString(a.stdout, fmt.Sprintf("Usage:\n  %s version\n\nShow the %s version.\n", a.name, a.name))
	}

	command, ok := a.commands[args[0]]
	if !ok {
		return fmt.Errorf("unknown command %q; run `%s list`", args[0], a.name)
	}

	return writeString(a.stdout, command.Usage)
}

func (a *App) writeHelp() error {
	var output strings.Builder
	fmt.Fprintf(&output, "Usage:\n  %s <command> [flags]\n\n", a.name)
	fmt.Fprintf(&output, "%s\n\n", a.tagline)
	output.WriteString("Commands:\n")

	for _, command := range a.sortedCommands() {
		fmt.Fprintf(&output, "  %-8s %s\n", command.Name, command.Summary)
	}

	output.WriteString("  list     List the bundled commands.\n")
	fmt.Fprintf(&output, "  version  Show the %s version.\n", a.name)
	fmt.Fprintf(&output, "  help     Show help for %s or a subcommand.\n", a.name)

	return writeString(a.stdout, output.String())
}

func (a *App) writeList() error {
	var output strings.Builder
	for _, command := range a.sortedCommands() {
		fmt.Fprintf(&output, "%s\t%s\n", command.Name, command.Summary)
	}

	fmt.Fprintf(&output, "help\tShow help for %s or a subcommand.\n", a.name)
	output.WriteString("list\tList the bundled commands.\n")
	fmt.Fprintf(&output, "version\tShow the %s version.\n", a.name)

	return writeString(a.stdout, output.String())
}

func (a *App) sortedCommands() []Command {
	commands := make([]Command, 0, len(a.commands))
	for _, command := range a.commands {
		commands = append(commands, command)
	}
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})
	return commands
}

func writeString(w io.Writer, value string) error {
	_, err := io.WriteString(w, value)
	return err
}
