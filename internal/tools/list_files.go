package tools

import (
	"fmt"
	"os"
	"strings"
)

func ListFilesTool() Tool {
	return Tool{
		Name:        "list_files",
		Description: "List all files and directories at the given path. Use this to explore repository/codebase structure",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Directory path to list files and folders at",
				},
			},
			"required": []string{"path"},
		},
		Execute: func(args map[string]interface{}) (string, error) {
			path, ok := args["path"].(string) //typecasting to string
			if !ok {
				return "", fmt.Errorf("'path' argument is required and it must be a string")
			}

			entries, err := os.ReadDir(path)
			if err != nil {
				return "", fmt.Errorf("failed to list directory ERR: %w", err)
			}

			var result strings.Builder
			for _, entry := range entries {
				if entry.IsDir() {
					result.WriteString(fmt.Sprintf("[DIR] %s\n", entry.Name()))
				} else {
					result.WriteString(fmt.Sprintf("[FILE] %s\n", entry.Name()))
				}
			}

			return result.String(), nil
		},
	}
}
