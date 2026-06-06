package tools

import (
	"fmt"
	"os/exec"
	"runtime"
)

func RunCommandTool() Tool {
	return Tool{
		Name:        "run_command",
		Description: "Run a shell command and return it's output. Use this to run tests, install dependencies or check build status",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "The shell command to execute",
				},
			},
			"required": []string{"command"},
		},
		Execute: func(args map[string]interface{}) (string, error) {
			command, ok := args["command"].(string)
			if !ok {
				return "", fmt.Errorf("'command' argument is required and must be a string")
			}

			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = exec.Command("cmd", "/c", command)
			} else {
				cmd = exec.Command("sh", "-c", command)
			}

			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Sprintf("Command failed: \n%s\nERROR: %s", string(output), err), nil
			}

			return string(output), nil
		},
	}
}
