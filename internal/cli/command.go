package cli

import "context"

// Command describes one command that can be registered with an App.
type Command struct {
	Name    string
	Summary string
	Usage   string
	Run     func(context.Context, []string) error
}
