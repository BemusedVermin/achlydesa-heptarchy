package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProtectedPath(t *testing.T) {
	tests := []struct {
		relative string
		want     string
	}{
		{"AGENTS.md", "AGENTS.md"},
		{"agents.md", "AGENTS.md"},
		{"./CLAUDE.md", "CLAUDE.md"},
		{"docs/go-technical-design.md", ""},
		{"docs", ""},
		{".claude/settings.json", ".claude/"},
		{".github/workflows/ci.yml", ".github/"},
		{".golangci.yml", ".golangci.yml"},
		{"Taskfile.yml", "Taskfile.yml"},
		{"internal/tools/agentkit/guard.go", "internal/tools/agentkit/"},
		{"internal/core/combat/combat.go", ""},
		{"tasks/INDEX.md", ""},
		{"tasks/followups/M2-07.F1-x.md", ""},
		{"content/data/forms.cue", ""},
		{"docsify/x.md", ""},
		{"internal/tools/agentkitten/x.go", ""},
	}
	for _, test := range tests {
		if got := protectedPath(test.relative); got != test.want {
			t.Errorf("protectedPath(%q) = %q, want %q", test.relative, got, test.want)
		}
	}
}

func TestCheckEdit(t *testing.T) {
	top := newRepo(t)
	tests := []struct {
		name    string
		branch  string
		path    string
		refused bool
	}{
		{"task branch, code", "task/M2-07-x", "internal/core/combat/combat.go", false},
		{"task branch, design docs", "task/M2-07-x", "docs/game-design.md", false},
		{"task branch, agents rules", "task/M2-07-x", "AGENTS.md", true},
		{"task branch, new folder under a protected one", "task/M2-07-x", ".claude/new/deep/file.md", true},
		{"owner's branch, agents rules", "spec/M3", "AGENTS.md", false},
		{"master, agents rules", "master", "AGENTS.md", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := fakeBranches{filepath.Clean(top): test.branch}
			refusal := checkEdit(filepath.Join(top, filepath.FromSlash(test.path)), reader)
			if (refusal != "") != test.refused {
				t.Errorf("checkEdit(%s) on %s = %q, want refused %v", test.path, test.branch, refusal, test.refused)
			}
		})
	}
	if refusal := checkEdit(filepath.Join(t.TempDir(), "docs", "x.md"), fakeBranches{}); refusal != "" {
		t.Errorf("checkEdit outside a checkout = %q, want no refusal", refusal)
	}
}

func TestNearestDir(t *testing.T) {
	root := t.TempDir()
	if got := nearestDir(filepath.Join(root, "a", "b")); got != root {
		t.Errorf("nearestDir = %q, want %q", got, root)
	}
	if err := os.Mkdir(filepath.Join(root, "a"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := nearestDir(filepath.Join(root, "a", "b")); got != filepath.Join(root, "a") {
		t.Errorf("nearestDir = %q, want %q", got, filepath.Join(root, "a"))
	}
}
