//go:build !windows

package clip

import (
	"context"
	"os/exec"
	"runtime"
)

func (c SystemClipboard) copy(ctx context.Context, text string) error {
	lookupPath := c.LookupPath
	if lookupPath == nil {
		lookupPath = exec.LookPath
	}

	commandContext := c.CommandContext
	if commandContext == nil {
		commandContext = exec.CommandContext
	}

	return copyExec(ctx, runtime.GOOS, text, lookupPath, commandContext)
}
