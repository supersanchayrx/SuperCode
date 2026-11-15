package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type writeFileArgs struct {
	Path    string  `json:"path"`
	Content *string `json:"content"`
}

func WriteFile() Tool {
	return Tool{
		Name:        "write_file",
		Description: "Create a new file, or replace an existing file's entire contents. The path is relative to the project root; missing folders are created. Always send the complete file contents. To change part of an existing file, prefer edit_file. The user sees a diff and must approve every write.",
		Parameters: json.RawMessage(`{
    "type": "object",
    "properties": {
        "path": {"type": "string", "description": "File path relative to the project root, e.g. internal/tools/hello.go"},
        "content": {"type": "string", "description": "The complete new contents of the file"}
    },
    "required": ["path", "content"]
}`),
		Run:           writeFile,
		NeedsApproval: true,
		Preview:       previewWriteFile,
	}
}

func decodeWriteFileArgs(args string) (writeFileArgs, error) {
	var a writeFileArgs

	if err := json.Unmarshal([]byte(args), &a); err != nil {
		return a, fmt.Errorf("invalid arguments for write_file: %w", err)
	}

	if a.Path == "" {
		return a, errors.New("path is required")
	}

	if a.Content == nil {
		return a, errors.New("content is required")
	}

	return a, nil
}

func writeFile(args string) (string, error) {
	a, err := decodeWriteFileArgs(args)

	if err != nil {
		return "", err
	}

	return SaveFile(a.Path, *a.Content)
}

func previewWriteFile(args string) (string, error) {
	a, err := decodeWriteFileArgs(args)

	if err != nil {
		return "", err
	}

	before, existed, err := currentContents(a.Path)

	if err != nil {
		return "", err
	}

	if !existed {
		return "(new file)\n" + diffPreview("", *a.Content), nil
	}

	return diffPreview(before, *a.Content), nil
}

func SaveFile(path, content string) (string, error) {
	_, existed, err := currentContents(path)

	if err != nil {
		return "", err
	}

	abs, err := safePath(path)

	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", fmt.Errorf("couldn't create folders for %q: %w", path, err)
	}

	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("couldn't write %q: %w", path, err)
	}

	action := "created"
	if existed {
		action = "overwrote"
	}

	return fmt.Sprintf("%s %q (%d bytes)", action, path, len(content)), nil
}

func FileExists(path string) bool {
	_, existed, err := currentContents(path)
	return err == nil && existed
}

func currentContents(path string) (string, bool, error) {
	abs, err := safePath(path)

	if err != nil {
		return "", false, err
	}

	info, err := os.Stat(abs)

	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}

	if err != nil {
		return "", false, fmt.Errorf("couldn't check %q: %w", path, err)
	}

	if info.IsDir() {
		return "", false, fmt.Errorf("%q is a directory, not a file", path)
	}

	data, err := os.ReadFile(abs)

	if err != nil {
		return "", false, fmt.Errorf("couldn't read %q: %w", path, err)
	}

	return string(data), true, nil
}
