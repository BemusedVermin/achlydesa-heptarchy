package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestGatePush(t *testing.T) {
	top := newRepo(t)
	runs := 0
	passing := func(string) (string, error) { runs++; return "ok", nil }
	failing := func(string) (string, error) { runs++; return "lint: unused variable\n", errors.New("exit 1") }

	if refusal := gatePush(top, failing); !strings.Contains(refusal, "unused variable") {
		t.Errorf("failing check: refusal = %q, want the check's output", refusal)
	}
	if refusal := gatePush(top, passing); refusal != "" {
		t.Errorf("passing check: refusal = %q", refusal)
	}
	if refusal := gatePush(top, failing); refusal != "" || runs != 2 {
		t.Errorf("after a pass on this commit: refusal = %q, runs = %d; want no refusal and no new run", refusal, runs)
	}

	if err := os.WriteFile(filepath.Join(top, "scratch.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if refusal := gatePush(top, passing); !strings.Contains(refusal, "scratch.txt") {
		t.Errorf("dirty tree: refusal = %q, want it to name the uncommitted file", refusal)
	}
	if refusal := gatePush(t.TempDir(), failing); refusal != "" {
		t.Errorf("outside a checkout: refusal = %q, want none", refusal)
	}
}

func TestMarkChecked(t *testing.T) {
	top := newRepo(t)
	if err := markChecked(top); err != nil {
		t.Fatal(err)
	}
	repo, err := openCheckout(top)
	if err != nil {
		t.Fatal(err)
	}
	if !repo.checked() {
		t.Error("clean tree: HEAD not recorded as checked")
	}
	if err := os.Remove(repo.marker); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(top, "scratch.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := markChecked(top); err != nil {
		t.Fatal(err)
	}
	if repo.checked() {
		t.Error("dirty tree: HEAD recorded as checked")
	}
	if err := markChecked(t.TempDir()); err != nil {
		t.Errorf("outside a checkout: %v", err)
	}
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
