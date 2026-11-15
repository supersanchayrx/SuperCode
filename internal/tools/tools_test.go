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

func TestReadFile(t *testing.T) {
	got, err := readFile(`{"path":"tools.go"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "package tools") {
		t.Error("read_file did not return the file contents")
	}

	for _, args := range []string{`{}`, `{oops`, `{"path":"missing.go"}`, `{"path":"../../main.go"}`, `{"path":"."}`} {
		if _, err := readFile(args); err == nil {
			t.Errorf("readFile(%s) returned no error", args)
		}
	}
}

func TestListFiles(t *testing.T) {
	for _, args := range []string{``, `{}`, `{"path":"."}`} {
		got, err := listFiles(args)
		if err != nil {
			t.Fatalf("listFiles(%q) error: %v", args, err)
		}
		if !strings.Contains(got, "tools.go\n") {
			t.Errorf("listFiles(%q) = %q, want it to include tools.go", args, got)
		}
	}

	for _, args := range []string{`{"path":"../.."}`, `{"path":"nope"}`, `{"path":"tools.go"}`} {
		if _, err := listFiles(args); err == nil {
			t.Errorf("listFiles(%s) returned no error", args)
		}
	}
}

func TestWriteFile(t *testing.T) {
	t.Chdir(t.TempDir())

	got, err := writeFile(`{"path":"sub/hello.txt","content":"hi"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "created") {
		t.Errorf("first write = %q, want it to start with created", got)
	}

	if got, _ := writeFile(`{"path":"sub/hello.txt","content":""}`); !strings.HasPrefix(got, "overwrote") {
		t.Errorf("second write = %q, want it to start with overwrote", got)
	}

	if got, err := readFile(`{"path":"sub/hello.txt"}`); err != nil || got != "" {
		t.Errorf("after writing empty content, read = %q, %v", got, err)
	}

	for _, args := range []string{`{"path":"x.txt"}`, `{"content":"hi"}`, `{"path":"../x.txt","content":"hi"}`, `{"path":"sub","content":"hi"}`} {
		if _, err := writeFile(args); err == nil {
			t.Errorf("writeFile(%s) returned no error", args)
		}
	}
}

func TestRunCommand(t *testing.T) {
	got, err := runCommand(`{"command":"echo hello"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "hello") || !strings.Contains(got, "[exit code 0]") {
		t.Errorf("runCommand(echo hello) = %q", got)
	}

	got, err = runCommand(`{"command":"exit 3"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "[exit code 3]") {
		t.Errorf("runCommand(exit 3) = %q, want exit code 3", got)
	}

	if _, err := runCommand(`{"command":""}`); err == nil {
		t.Error("runCommand with an empty command returned no error")
	}
}

func TestApprovalFlags(t *testing.T) {
	r := NewRegistry(ReadFile(), ListFiles(), EditFile(), WriteFile(), RunCommand())
	want := map[string]bool{"read_file": false, "list_files": false, "edit_file": true, "write_file": true, "run_command": true}

	for name, approval := range want {
		tool, ok := r.Get(name)
		if !ok {
			t.Fatalf("tool %s is not registered", name)
		}
		if tool.NeedsApproval != approval {
			t.Errorf("%s NeedsApproval = %v, want %v", name, tool.NeedsApproval, approval)
		}
	}
}

func TestEditFile(t *testing.T) {
	t.Chdir(t.TempDir())

	if _, err := SaveFile("a.go", "package a\n\nfunc One() int {\n\treturn 1\n}\n\nfunc Two() int {\n\treturn 1\n}\n"); err != nil {
		t.Fatal(err)
	}

	preview, err := previewEditFile(`{"path":"a.go","old_text":"func Two() int {\n\treturn 1","new_text":"func Two() int {\n\treturn 2"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(preview, "- \treturn 1") || !strings.Contains(preview, "+ \treturn 2") {
		t.Errorf("preview does not show the change:\n%s", preview)
	}

	if got, _ := readFile(`{"path":"a.go"}`); strings.Contains(got, "return 2") {
		t.Fatal("preview changed the file")
	}

	if _, err := editFile(`{"path":"a.go","old_text":"func Two() int {\n\treturn 1","new_text":"func Two() int {\n\treturn 2"}`); err != nil {
		t.Fatal(err)
	}

	got, _ := readFile(`{"path":"a.go"}`)
	if !strings.Contains(got, "func One() int {\n\treturn 1") || !strings.Contains(got, "func Two() int {\n\treturn 2") {
		t.Errorf("file after edit:\n%s", got)
	}

	for _, args := range []string{
		`{"path":"a.go","old_text":"return 3","new_text":"x"}`,
		`{"path":"a.go","old_text":"int {","new_text":"x"}`,
		`{"path":"a.go","old_text":"","new_text":"x"}`,
		`{"path":"missing.go","old_text":"a","new_text":"b"}`,
		`{"path":"../a.go","old_text":"a","new_text":"b"}`,
	} {
		if _, err := editFile(args); err == nil {
			t.Errorf("editFile(%s) returned no error", args)
		}
	}
}

func TestEditFileCRLF(t *testing.T) {
	t.Chdir(t.TempDir())

	if _, err := SaveFile("w.txt", "one\r\ntwo\r\nthree\r\n"); err != nil {
		t.Fatal(err)
	}

	if _, err := editFile(`{"path":"w.txt","old_text":"one\ntwo","new_text":"one\n2"}`); err != nil {
		t.Fatal(err)
	}

	if got, _ := readFile(`{"path":"w.txt"}`); got != "one\r\n2\r\nthree\r\n" {
		t.Errorf("CRLF file after edit = %q", got)
	}
}

func TestDiffPreview(t *testing.T) {
	if got := diffPreview("a\nb\n", "a\nb\n"); got != "(no changes)" {
		t.Errorf("identical text = %q", got)
	}

	got := diffPreview("1\n2\n3\n4\n5\n6\n", "1\n2\n3\nfour\n5\n6\n")
	want := "@@ line 4: -1 +1 @@\n  2\n  3\n- 4\n+ four\n  5\n  6"
	if got != want {
		t.Errorf("diffPreview =\n%s\nwant\n%s", got, want)
	}

	if got := diffPreview("", "x\ny\n"); !strings.Contains(got, "+ x\n+ y") {
		t.Errorf("new file preview = %q", got)
	}

	var long strings.Builder
	for i := 0; i < 100; i++ {
		long.WriteString("line\n")
	}
	if got := diffPreview("", long.String()); !strings.Contains(got, "more lines") {
		t.Error("long preview was not shortened")
	}
}

func TestSaveFileAndFileExists(t *testing.T) {
	t.Chdir(t.TempDir())

	if FileExists("PLAN.md") {
		t.Fatal("FileExists is true before the file was written")
	}
	if _, err := SaveFile("PLAN.md", "# plan\n"); err != nil {
		t.Fatal(err)
	}
	if !FileExists("PLAN.md") {
		t.Fatal("FileExists is false after the file was written")
	}
	if FileExists("../PLAN.md") {
		t.Error("FileExists is true for a path outside the project")
	}

	preview, err := previewWriteFile(`{"path":"PLAN.md","content":"# plan\nstep 1\n"}`)
	if err != nil || !strings.Contains(preview, "+ step 1") {
		t.Errorf("write_file preview = %q, %v", preview, err)
	}

	if preview, _ := previewWriteFile(`{"path":"new.txt","content":"hi"}`); !strings.HasPrefix(preview, "(new file)") {
		t.Errorf("new file preview = %q", preview)
	}
}
