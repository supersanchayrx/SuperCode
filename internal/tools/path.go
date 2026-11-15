package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func safePath(p string) (string, error) {

	root, err := os.Getwd()

	if err != nil {
		return "", fmt.Errorf("couldn't read the project directory: %w", err)
	}

	abs, err := filepath.Abs(p)

	if err != nil {
		return "", fmt.Errorf("invalid path %q: %w", p, err)
	}

	rel, err := filepath.Rel(root, abs)

	if err != nil {
		return "", fmt.Errorf("path %q is outside the project directory; use a path relative to the project root", p)
	}

	if pathStart := strings.Split(rel, string(filepath.Separator)); pathStart[0] == ".." {
		return "", fmt.Errorf("path %q is outside the project directory; use a path relative to the project root", p)
	}

	return abs, nil
}
