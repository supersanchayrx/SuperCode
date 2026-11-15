package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	commandTimeout   = 60 * time.Second
	maxCommandOutput = 20000
)

type runCommandArgs struct {
	Command string `json:"command"`
}

func RunCommand() Tool {
	return Tool{
		Name:        "run_command",
		Description: fmt.Sprintf("Run a shell command in the project root and return its combined output and exit code. It runs with %s, has no stdin, and is stopped after %s. Use it for builds, tests and other non-interactive commands. The user must approve every command.", shellName, commandTimeout),
		Parameters: json.RawMessage(`{
    "type": "object",
    "properties": {
        "command": {"type": "string", "description": "The command line to run, e.g. go test ./..."}
    },
    "required": ["command"]
}`),
		Run:           runCommand,
		NeedsApproval: true,
		Preview:       previewRunCommand,
	}
}

func previewRunCommand(args string) (string, error) {
	var a runCommandArgs

	if err := json.Unmarshal([]byte(args), &a); err != nil {
		return "", fmt.Errorf("invalid arguments for run_command: %w", err)
	}

	if strings.TrimSpace(a.Command) == "" {
		return "", errors.New("command is required")
	}

	return "$ " + a.Command, nil
}

func runCommand(args string) (string, error) {
	var a runCommandArgs

	if err := json.Unmarshal([]byte(args), &a); err != nil {
		return "", fmt.Errorf("invalid arguments for run_command: %w", err)
	}

	if strings.TrimSpace(a.Command) == "" {
		return "", errors.New("command is required")
	}

	root, err := os.Getwd()

	if err != nil {
		return "", fmt.Errorf("couldn't read the project directory: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	cmd := shellCommand(ctx, a.Command)
	cmd.Dir = root
	cmd.WaitDelay = 2 * time.Second

	out, err := cmd.CombinedOutput()

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "", fmt.Errorf("command timed out after %s; output so far:\n%s", commandTimeout, truncateOutput(out))
	}

	exitCode := 0

	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return "", fmt.Errorf("couldn't run command: %w", err)
		}
		exitCode = exitErr.ExitCode()
	}

	result := truncateOutput(out)
	if result == "" {
		result = "(no output)"
	}

	return fmt.Sprintf("%s\n[exit code %d]", result, exitCode), nil
}

func truncateOutput(out []byte) string {
	if len(out) > maxCommandOutput {
		return string(out[:maxCommandOutput]) + fmt.Sprintf("\n[truncated: output is %d bytes, showing the first %d]", len(out), maxCommandOutput)
	}
	return string(out)
}
