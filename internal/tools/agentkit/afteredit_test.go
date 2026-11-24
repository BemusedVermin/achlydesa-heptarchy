package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// fakeTools records what the after-edit hook asked for and answers with canned lint output.
type fakeTools struct {
	missing   bool
	output    string
	formatted []string
	linted    []string
}

func (fake *fakeTools) available() bool { return !fake.missing }

func (fake *fakeTools) format(_, file string) error {
	fake.formatted = append(fake.formatted, file)
	return nil
}

func (fake *fakeTools) lint(_, dir string) (string, error) {
	fake.linted = append(fake.linted, dir)
	return fake.output, nil
}

// editInput is a PostToolUse input for an edit to file.
func editInput(cwd, file string) hookInput {
	var input hookInput
	input.Cwd = cwd
	input.ToolName = "Edit"
	input.ToolInput.FilePath = file
	return input
}

func TestFindingsIn(t *testing.T) {
	output := strings.Join([]string{
		"internal/core/actions/options.go:12:3: Error return value is not checked (errcheck)",
		"internal/core/actions/choices.go:4:1: exported function Foo should have comment (revive)",
		`internal\core\actions\options.go:30:1: unused variable (unused)`,
		"",
	}, "\n")
	got := findingsIn(output, "internal/core/actions/options.go")
	want := []string{
		"internal/core/actions/options.go:12:3: Error return value is not checked (errcheck)",
		"internal/core/actions/options.go:30:1: unused variable (unused)",
	}
	if !slices.Equal(got, want) {
		t.Errorf("findingsIn = %q, want %q", got, want)
	}

	var many []string
	for line := range maxFindings + 5 {
		many = append(many, "a.go:"+strconv.Itoa(line)+":1: finding (errcheck)")
	}
	got = findingsIn(strings.Join(many, "\n"), "a.go")
	if len(got) != maxFindings+1 || got[maxFindings] != "(more not shown)" {
		t.Errorf("findingsIn kept %d lines, want %d and a note", len(got), maxFindings)
	}
}

func TestAfterEdit(t *testing.T) {
	top := newRepo(t)
	dir := filepath.Join(top, "internal", "thing")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "thing.go")
	if err := os.WriteFile(file, []byte("package thing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	findings := "internal/thing/thing.go:3:2: Error return value is not checked (errcheck)\n" +
		"internal/thing/other.go:9:1: something else (revive)\n"

	t.Run("a Go file", func(t *testing.T) {
		tools := &fakeTools{output: findings}
		note := afterEdit(editInput(top, "internal/thing/thing.go"), tools)
		if !slices.Equal(tools.formatted, []string{"internal/thing/thing.go"}) ||
			!slices.Equal(tools.linted, []string{"internal/thing"}) {
			t.Errorf("formatted %q and linted %q; want the file and its package", tools.formatted, tools.linted)
		}
		if !strings.Contains(note, "thing.go:3:2: Error return value") || strings.Contains(note, "other.go") {
			t.Errorf("note = %q, want only the edited file's finding", note)
		}
	})
	t.Run("a clean Go file", func(t *testing.T) {
		if note := afterEdit(editInput(top, file), &fakeTools{}); note != "" {
			t.Errorf("note = %q, want none", note)
		}
	})
	t.Run("not Go", func(t *testing.T) {
		tools := &fakeTools{output: findings}
		if note := afterEdit(editInput(top, "Taskfile.yml"), tools); note != "" || len(tools.formatted) > 0 {
			t.Errorf("note = %q, formatted %q; want nothing done", note, tools.formatted)
		}
	})
	t.Run("golangci-lint missing", func(t *testing.T) {
		tools := &fakeTools{missing: true, output: findings}
		if note := afterEdit(editInput(top, file), tools); note != "" || len(tools.formatted) > 0 {
			t.Errorf("note = %q, formatted %q; want nothing done", note, tools.formatted)
		}
	})
	t.Run("a filler comment", func(t *testing.T) {
		filler := filepath.Join(dir, "filler.go")
		src := "package thing\n\n// load loads. It returns an error if loading fails.\nfunc load() {}\n"
		if err := os.WriteFile(filler, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		tools := &fakeTools{missing: true}
		note := afterEdit(editInput(top, filler), tools)
		if !strings.Contains(note, "internal/thing/filler.go:3: says errors are returned") {
			t.Errorf("note = %q, want the comment flagged even without golangci-lint", note)
		}
	})
	t.Run("a deleted file", func(t *testing.T) {
		tools := &fakeTools{output: findings}
		if note := afterEdit(editInput(top, filepath.Join(dir, "gone.go")), tools); note != "" {
			t.Errorf("note = %q, want none", note)
		}
	})
}

func TestRunAfterEditOutput(t *testing.T) {
	top := newRepo(t)
	file := filepath.Join(top, "main.go")
	if err := os.WriteFile(file, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	input, err := json.Marshal(map[string]any{
		"cwd": top, "tool_name": "Write", "tool_input": map[string]string{"file_path": file},
	})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr strings.Builder
	tools := &fakeTools{output: "main.go:1:1: a finding (errcheck)\n"}
	if status := runAfterEdit(strings.NewReader(string(input)), &stdout, &stderr, tools); status != exitOK {
		t.Fatalf("status = %d, stderr %q", status, stderr.String())
	}
	var output postToolUseOutput
	if err := json.Unmarshal([]byte(stdout.String()), &output); err != nil {
		t.Fatalf("output %q isn't JSON: %v", stdout.String(), err)
	}
	if output.HookSpecificOutput.HookEventName != "PostToolUse" ||
		!strings.Contains(output.HookSpecificOutput.AdditionalContext, "main.go:1:1: a finding") {
		t.Errorf("output = %+v", output)
	}
}

// TestAfterEditWithGolangciLint runs the real golangci-lint on a small module, when it's installed.
func TestAfterEditWithGolangciLint(t *testing.T) {
	if !(golangciLint{}).available() {
		t.Skip("golangci-lint is not installed")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go is not installed")
	}
	top := newRepo(t)
	files := map[string]string{
		"go.mod":         "module example.com/thing\n\ngo 1.24\n",
		"thing/thing.go": "package thing\nimport \"os\"\n// Remove removes.\nfunc Remove( ) { os.Remove( \"x\" ) }\n",
	}
	for name, content := range files {
		path := filepath.Join(top, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	file := filepath.Join(top, "thing", "thing.go")
	note := afterEdit(editInput(top, file), golangciLint{})
	if !strings.Contains(note, "thing/thing.go:") || !strings.Contains(note, "errcheck") {
		t.Errorf("note = %q, want the errcheck finding", note)
	}
	formatted, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(formatted), "func Remove() { os.Remove(\"x\") }") {
		t.Errorf("the file wasn't formatted:\n%s", formatted)
	}
}
