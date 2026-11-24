package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProtectedPath(t *testing.T) {
	tests := []string{
		"AGENTS.md",
		"CLAUDE.md",
		".claude/settings.json",
		".github/workflows/ci.yml",
		"Taskfile.yml",
		"internal/tools/agentkit/guard.go",
		"internal/core/combat/combat.go",
		"tasks/INDEX.md",
	}
	for _, relative := range tests {
		if got := protectedPath(relative); got != "" {
			t.Errorf("protectedPath(%q) = %q, want none", relative, got)
		}
	}
}

func TestCheckEdit(t *testing.T) {
	top := newRepo(t)
	tests := []struct {
		name   string
		branch string
		path   string
	}{
		{"task branch, code", "task/M2-07-x", "internal/core/combat/combat.go"},
		{"task branch, agents rules", "task/M2-07-x", "AGENTS.md"},
		{"task branch, agentkit", "task/M2-07-x", "internal/tools/agentkit/checkpr.go"},
		{"task branch, new folder under .claude", "task/M2-07-x", ".claude/new/deep/file.md"},
		{"master, agents rules", "master", "AGENTS.md"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := fakeBranches{filepath.Clean(top): test.branch}
			if refusal := checkEdit(filepath.Join(top, filepath.FromSlash(test.path)), reader); refusal != "" {
				t.Errorf("checkEdit(%s) on %s = %q, want no refusal", test.path, test.branch, refusal)
			}
		})
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
