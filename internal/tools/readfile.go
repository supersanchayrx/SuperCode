package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

const maxReadBytes = 20000

type readFileArgs struct {
	Path string `json:"path"`
}

func ReadFile() Tool {

	tool := Tool{
		Name:        "read_file",
		Description: "Read a text file in the project and return its contents. The path is relative to the project root. Large files are truncated.",
		Parameters: json.RawMessage(`{
    "type": "object",
    "properties": {
        "path": {"type": "string", "description": "File path relative to the project root, e.g. internal/llm/types.go"}
    },
    "required": ["path"]
}`),

		Run: readFile,
	}
	return tool
}

func readFile(args string) (string, error) {

	var a readFileArgs

	if err := json.Unmarshal([]byte(args), &a); err != nil {
		return "", fmt.Errorf("invalid arguments for read_file: %w", err)
	}

	if a.Path == "" {
		return "", errors.New("path is required")
	}

	abs, err := safePath(a.Path)

	if err != nil {
		return "", err
	}

	info, err := os.Stat(abs)

	if err != nil {
		return "", fmt.Errorf("couldn't read %q: %w", a.Path, err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("%q is a directory; use list_files to see what's in it", a.Path)
	}

	data, err := os.ReadFile(abs)

	if err != nil {
		return "", fmt.Errorf("couldn't read %q: %w", a.Path, err)
	}

	if len(data) > maxReadBytes {
		note := fmt.Sprintf("\n\n[truncated: file is %d bytes, showing the first %d]", len(data), maxReadBytes)
		return string(data[:maxReadBytes]) + note, nil
	}

	return string(data), nil
}
