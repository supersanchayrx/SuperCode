//go:build !windows

package tools

import (
	"context"
	"os/exec"
)

const shellName = "sh"

func shellCommand(ctx context.Context, command string) *exec.Cmd {
	return exec.CommandContext(ctx, "sh", "-c", command)
}
