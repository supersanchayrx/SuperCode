package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type editFileArgs struct {
	Path    string `json:"path"`
	OldText string `json:"old_text"`
	NewText string `json:"new_text"`
}

type plannedEdit struct {
	abs    string
	path   string
	before string
	after  string
}

func EditFile() Tool {
	return Tool{
		Name:        "edit_file",
		Description: "Change part of an existing file by replacing one exact piece of text with new text. old_text must match the file exactly (including indentation) and appear exactly once; copy it from read_file output and include a few surrounding lines if needed to make it unique. Use write_file for new files. The user sees a diff and must approve every edit.",
		Parameters: json.RawMessage(`{
    "type": "object",
    "properties": {
        "path": {"type": "string", "description": "File path relative to the project root"},
        "old_text": {"type": "string", "description": "The exact text to replace; must appear exactly once in the file"},
        "new_text": {"type": "string", "description": "The text to put in its place; may be empty to delete old_text"}
    },
    "required": ["path", "old_text", "new_text"]
}`),
		Run:           editFile,
		NeedsApproval: true,
		Preview:       previewEditFile,
	}
}

func editFile(args string) (string, error) {
	e, err := planEdit(args)

	if err != nil {
		return "", err
	}

	if err := os.WriteFile(e.abs, []byte(e.after), 0o644); err != nil {
		return "", fmt.Errorf("couldn't write %q: %w", e.path, err)
	}

	return fmt.Sprintf("edited %q (%d lines before, %d lines after)", e.path, len(splitLines(e.before)), len(splitLines(e.after))), nil
}

func previewEditFile(args string) (string, error) {
	e, err := planEdit(args)

	if err != nil {
		return "", err
	}

	return diffPreview(e.before, e.after), nil
}

func planEdit(args string) (plannedEdit, error) {
	var a editFileArgs

	if err := json.Unmarshal([]byte(args), &a); err != nil {
		return plannedEdit{}, fmt.Errorf("invalid arguments for edit_file: %w", err)
	}

	if a.Path == "" {
		return plannedEdit{}, errors.New("path is required")
	}

	if a.OldText == "" {
		return plannedEdit{}, errors.New("old_text is required; to create a file or replace all of it, use write_file")
	}

	abs, err := safePath(a.Path)

	if err != nil {
		return plannedEdit{}, err
	}

	info, err := os.Stat(abs)

	if err != nil {
		return plannedEdit{}, fmt.Errorf("couldn't read %q: %w", a.Path, err)
	}

	if info.IsDir() {
		return plannedEdit{}, fmt.Errorf("%q is a directory, not a file", a.Path)
	}

	data, err := os.ReadFile(abs)

	if err != nil {
		return plannedEdit{}, fmt.Errorf("couldn't read %q: %w", a.Path, err)
	}

	before := string(data)
	oldText, newText := a.OldText, a.NewText

	if strings.Contains(before, "\r\n") {
		oldText = toCRLF(oldText)
		newText = toCRLF(newText)
	}

	switch n := strings.Count(before, oldText); {
	case n == 0:
		return plannedEdit{}, fmt.Errorf("old_text was not found in %q; read the file again and copy the exact text, including indentation", a.Path)
	case n > 1:
		return plannedEdit{}, fmt.Errorf("old_text appears %d times in %q; include more surrounding lines so it matches exactly once", n, a.Path)
	}

	return plannedEdit{
		abs:    abs,
		path:   a.Path,
		before: before,
		after:  strings.Replace(before, oldText, newText, 1),
	}, nil
}

func toCRLF(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\n", "\r\n")
}
