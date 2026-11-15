package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const maxListEntries = 500

type listFilesArgs struct {
	Path string `json:"path"`
}

func ListFiles() Tool {
	return Tool{
		Name:        "list_files",
		Description: "List the files and folders in a project directory. Folders end with /. Omit path to list the project root.",
		Parameters: json.RawMessage(`{
    "type": "object",
    "properties": {
        "path": {"type": "string", "description": "Directory relative to the project root, e.g. internal/llm. Defaults to the root."}
    }
}`),
		Run: listFiles,
	}
}

func listFiles(args string) (string, error) {
	var a listFilesArgs

	if strings.TrimSpace(args) != "" {
		if err := json.Unmarshal([]byte(args), &a); err != nil {
			return "", fmt.Errorf("invalid arguments for list_files: %w", err)
		}
	}

	if a.Path == "" {
		a.Path = "."
	}

	abs, err := safePath(a.Path)

	if err != nil {
		return "", err
	}

	info, err := os.Stat(abs)

	if err != nil {
		return "", fmt.Errorf("couldn't list %q: %w", a.Path, err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("%q is a file, not a directory; use read_file to read it", a.Path)
	}

	entries, err := os.ReadDir(abs)

	if err != nil {
		return "", fmt.Errorf("couldn't list %q: %w", a.Path, err)
	}

	var b strings.Builder
	count := 0
	skipped := 0

	for _, e := range entries {
		if e.Name() == ".git" {
			continue
		}

		if count == maxListEntries {
			skipped++
			continue
		}

		b.WriteString(e.Name())

		if e.IsDir() {
			b.WriteString("/")
		}

		b.WriteString("\n")
		count++
	}

	if count == 0 {
		return "(empty directory)", nil
	}

	if skipped > 0 {
		fmt.Fprintf(&b, "... and %d more\n", skipped)
	}

	return b.String(), nil
}
