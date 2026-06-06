package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

func WriteFileTool() Tool {
	return Tool{
		Name:        "write_file",
		Description: "Write content to a file. Creates the file if it doesn't exist, overwrites if it does. Also creates any missing parent directories.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The file path to write to",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "The content to write to the file",
				},
			},
			"required": []string{"path", "content"},
		},
		Execute: func(args map[string]interface{}) (string, error) {
			path, ok := args["path"].(string)
			if !ok {
				return "", fmt.Errorf("'path' argument is required and must be a string")
			}
			content, ok := args["content"].(string)
			if !ok {
				return "", fmt.Errorf("'content' argument is required and must be a string")
			}

			dir := filepath.Dir(path)
			err := os.MkdirAll(dir, 0755) //unix permission code

			if err != nil {
				return "", fmt.Errorf("Failed to create directories: %w", err)
			}

			err = os.WriteFile(path, []byte(content), 0644)
			if err != nil {
				return "", fmt.Errorf("Failed to write file: %w", err)
			}

			return fmt.Sprintf("Successfully wrote to %s", path), nil
		},
	}
}
