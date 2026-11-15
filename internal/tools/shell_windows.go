package tools

import (
	"context"
	"os/exec"
	"syscall"
)

const shellName = "cmd.exe on Windows"

func shellCommand(ctx context.Context, command string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "cmd")
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: "cmd /C " + command}
	return cmd
}
