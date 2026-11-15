// This file was written by an AI assistant, not by hand.
//
// Tests for the tools package. Run them with:
//
//	go test ./internal/tools -v
//
// Note: while tests run, the working directory is the package folder
// (internal/tools), so that folder acts as the "project root" for safePath.

package tools

import (
	"fmt"
	"strings"
	"testing"
)

// TestSafePath checks which paths safePath allows and which it refuses.
// It is table-driven: each row is one case, and the loop below runs them all.
func TestSafePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		// Paths inside the project root are allowed.
		{"file in root", "tools.go", false},
		{"root itself", ".", false},
		{"dot-dot inside a path that stays in root", "sub/../tools.go", false},
		{"file name starting with two dots", "..notes.txt", false},

		// Paths that escape the project root are refused.
		{"parent directory", "..", true},
		{"file in parent directory", "../path.go", true},
		{"two levels up", "../../main.go", true},
		{"absolute path elsewhere on the drive", `C:\Windows`, true},
		{"different drive", `D:\stuff`, true},
	}

	for _, tc := range tests {
		// t.Run gives each case its own name in the -v output,
		// so a failure says exactly which row broke.
		t.Run(tc.name, func(t *testing.T) {
			got, err := safePath(tc.path)

			// (err != nil) is true when safePath refused the path.
			// It must match what the row expects.
			if (err != nil) != tc.wantErr {
				t.Fatalf("safePath(%q) error = %v, wantErr %v", tc.path, err, tc.wantErr)
			}

			// When a path is refused, nothing should be returned.
			if tc.wantErr && got != "" {
				t.Errorf("safePath(%q) returned %q along with an error, want empty string", tc.path, got)
			}

			// The refusal message should name the path, because the model reads it.
			// safePath quotes the path with %q, which escapes backslashes
			// (C:\Windows becomes "C:\\Windows"), so look for the quoted form.
			quoted := fmt.Sprintf("%q", tc.path)
			if tc.wantErr && !strings.Contains(err.Error(), quoted) {
				t.Errorf("error %q does not mention the path %s", err, quoted)
			}
		})
	}
}

// fakeTool builds a Tool whose Run just reports its own name and arguments,
// so the tests can see which tool the registry actually called.
func fakeTool(name string) Tool {
	return Tool{
		Name: name,
		Run: func(args string) (string, error) {
			return name + ":" + args, nil
		},
	}
}

// TestRegistryRun checks that Run finds a tool by name and calls it.
func TestRegistryRun(t *testing.T) {
	r := NewRegistry(fakeTool("a"), fakeTool("b"))

	got, err := r.Run("b", "hello")
	if err != nil {
		t.Fatalf("Run returned an error: %v", err)
	}
	if got != "b:hello" {
		t.Errorf("Run(\"b\", \"hello\") = %q, want %q", got, "b:hello")
	}
}

// TestRegistryUnknownTool checks that asking for a tool that isn't
// registered returns an error instead of panicking.
func TestRegistryUnknownTool(t *testing.T) {
	r := NewRegistry(fakeTool("a"))

	if _, err := r.Run("nope", "{}"); err == nil {
		t.Fatal("Run with an unknown tool name returned no error")
	}
}

// TestRegistryListOrder checks that List returns tools in the order they
// were registered, and that a duplicate name is skipped (the first one wins).
func TestRegistryListOrder(t *testing.T) {
	first := fakeTool("a")
	duplicate := Tool{Name: "a", Description: "should be skipped"}

	r := NewRegistry(first, fakeTool("b"), duplicate, fakeTool("c"))

	list := r.List()

	var names []string
	for _, tool := range list {
		names = append(names, tool.Name)
	}

	want := "a,b,c"
	if got := strings.Join(names, ","); got != want {
		t.Fatalf("List() names = %s, want %s", got, want)
	}

	// The kept "a" must be the first one, not the duplicate.
	if list[0].Description == "should be skipped" {
		t.Error("duplicate tool replaced the first one")
	}
}

// TestRegistryEmpty checks that a registry with no tools still works.
func TestRegistryEmpty(t *testing.T) {
	r := NewRegistry()

	if len(r.List()) != 0 {
		t.Errorf("List() on an empty registry has %d tools, want 0", len(r.List()))
	}
	if _, err := r.Run("anything", "{}"); err == nil {
		t.Error("Run on an empty registry returned no error")
	}
}
