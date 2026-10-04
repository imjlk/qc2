package cli

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
)

const minHelpWidth = 8

// App dispatches registered commands and provides the shared built-in commands.
type App struct {
	name     string
	tagline  string
	version  string
	stdout   io.Writer
	commands map[string]Command
}

type builtin struct {
	name            string
	dispatchAliases []string
	helpAliases     []string
	summary         func(appName string) string
	usage           func(appName string) string
	run             func(a *App, ctx context.Context, args []string) error
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
	if isReservedCommand(command.Name) {
		return fmt.Errorf("command %q is reserved", command.Name)
	}

	a.commands[command.Name] = command
	return nil
}

func (a *App) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return a.writeHelp()
	}

	if spec, ok := lookupBuiltin(args[0], func(spec builtin) []string {
		return spec.dispatchAliases
	}); ok {
		return spec.run(a, ctx, args[1:])
	}

	command, ok := a.commands[args[0]]
	if !ok {
		return fmt.Errorf("unknown command %q; run `%s list`", args[0], a.name)
	}

	return command.Run(ctx, args[1:])
}

func (a *App) runHelp(args []string) error {
	if len(args) == 0 || args[0] == "help" {
		if len(args) > 1 {
			return fmt.Errorf("help accepts at most one command name")
		}
		return a.writeHelp()
	}
	if len(args) > 1 {
		return fmt.Errorf("help accepts at most one command name")
	}

	spec, ok := lookupBuiltin(args[0], func(spec builtin) []string {
		return append([]string{spec.name}, spec.helpAliases...)
	})
	if ok {
		return writeString(a.stdout, spec.usage(a.name))
	}

	command, ok := a.commands[args[0]]
	if !ok {
		return fmt.Errorf("unknown command %q; run `%s list`", args[0], a.name)
	}

	return writeString(a.stdout, command.Usage)
}

func (a *App) writeHelp() error {
	width := a.helpWidth()
	var output strings.Builder
	fmt.Fprintf(&output, "Usage:\n  %s <command> [flags]\n\n", a.name)
	fmt.Fprintf(&output, "%s\n\n", a.tagline)
	output.WriteString("Commands:\n")

	for _, command := range a.sortedCommands() {
		fmt.Fprintf(&output, "  %-*s %s\n", width, command.Name, command.Summary)
	}
	for _, name := range []string{"list", "version", "help"} {
		spec, ok := builtinByName(name)
		if !ok {
			return fmt.Errorf("missing built-in command %q", name)
		}
		fmt.Fprintf(&output, "  %-*s %s\n", width, spec.name, spec.summary(a.name))
	}

	return writeString(a.stdout, output.String())
}

func (a *App) writeList() error {
	var output strings.Builder
	for _, command := range a.sortedCommands() {
		fmt.Fprintf(&output, "%s\t%s\n", command.Name, command.Summary)
	}
	for _, name := range []string{"help", "list", "version"} {
		spec, ok := builtinByName(name)
		if !ok {
			return fmt.Errorf("missing built-in command %q", name)
		}
		fmt.Fprintf(&output, "%s\t%s\n", spec.name, spec.summary(a.name))
	}

	return writeString(a.stdout, output.String())
}

func (a *App) helpWidth() int {
	width := minHelpWidth
	for _, command := range a.commands {
		if len(command.Name) > width {
			width = len(command.Name)
		}
	}
	for _, spec := range builtins() {
		if len(spec.name) > width {
			width = len(spec.name)
		}
	}
	return width
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

func builtins() []builtin {
	return []builtin{
		{
			name:            "help",
			dispatchAliases: []string{"-h", "--help"},
			summary: func(appName string) string {
				return fmt.Sprintf("Show help for %s or a subcommand.", appName)
			},
			run: func(a *App, _ context.Context, args []string) error {
				return a.runHelp(args)
			},
		},
		{
			name: "list",
			summary: func(string) string {
				return "List the bundled commands."
			},
			usage: func(appName string) string {
				return fmt.Sprintf("Usage:\n  %s list\n\nList the bundled commands.\n", appName)
			},
			run: func(a *App, _ context.Context, args []string) error {
				if len(args) > 0 {
					return fmt.Errorf("list does not accept arguments")
				}
				return a.writeList()
			},
		},
		{
			name:            "version",
			dispatchAliases: []string{"-v", "--version"},
			helpAliases:     []string{"-v", "--version"},
			summary: func(appName string) string {
				return fmt.Sprintf("Show the %s version.", appName)
			},
			usage: func(appName string) string {
				return fmt.Sprintf("Usage:\n  %s version\n\nShow the %s version.\n", appName, appName)
			},
			run: func(a *App, _ context.Context, args []string) error {
				if len(args) > 0 {
					return fmt.Errorf("version does not accept arguments")
				}
				return writeString(a.stdout, fmt.Sprintf("%s %s\n", a.name, a.version))
			},
		},
	}
}

func lookupBuiltin(token string, aliases func(builtin) []string) (builtin, bool) {
	for _, spec := range builtins() {
		if spec.name == token || sliceContains(aliases(spec), token) {
			return spec, true
		}
	}
	return builtin{}, false
}

func builtinByName(name string) (builtin, bool) {
	for _, spec := range builtins() {
		if spec.name == name {
			return spec, true
		}
	}
	return builtin{}, false
}

func isReservedCommand(name string) bool {
	for _, spec := range builtins() {
		if spec.name == name || sliceContains(spec.dispatchAliases, name) || sliceContains(spec.helpAliases, name) {
			return true
		}
	}
	return false
}

func sliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func writeString(w io.Writer, value string) error {
	_, err := io.WriteString(w, value)
	return err
}
