package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// newRepo makes a git repository with one commit holding a Taskfile.yml, and returns its root as
// git reports it. It skips the test when git isn't installed.
func newRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Taskfile.yml"), []byte("version: \"3\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"init", "-q", "-b", "task/M0-00-test"},
		{"add", "Taskfile.yml"},
		{
			"-c", "user.name=test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false",
			"commit", "-q", "-m", "start",
		},
	} {
		if _, err := gitOutput(dir, args...); err != nil {
			t.Fatal(err)
		}
	}
	top, err := gitOutput(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		t.Fatal(err)
	}
	return filepath.FromSlash(top)
}

func TestGitRepoBranch(t *testing.T) {
	top := newRepo(t)
	if got := (gitRepo{}).branch(top); got != "task/M0-00-test" {
		t.Errorf("branch = %q, want task/M0-00-test", got)
	}
	if got := (gitRepo{}).branch(t.TempDir()); got != "" {
		t.Errorf("branch outside a checkout = %q, want empty", got)
	}
}
