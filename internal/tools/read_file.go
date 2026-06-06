package tools

import (
	"fmt"
	"os"
)

func ReadFileTool() Tool {
	return Tool{
		Name:        "read_file",
		Description: "Read the file at path specified. Use this to build knowledge about existing code before making changes",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The Path of the file to read",
				},
			},
			"required": []string{"path"},
		},
		Execute: func(args map[string]interface{}) (string, error) {
			path, ok := args["path"].(string)
			if !ok {
				return "", fmt.Errorf("'path' argument is required and must be a string")
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return "", fmt.Errorf("failed to read file ERR: %w", err)
			}

			return string(content), nil
		},
	}
}
