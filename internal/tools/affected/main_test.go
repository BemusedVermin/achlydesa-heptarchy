package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
)

func TestChangedFilesIncludesUncommittedWork(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := t.TempDir()
	t.Chdir(dir)
	write := func(name, content string) {
		t.Helper()
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	git := func(args ...string) {
		t.Helper()
		args = append([]string{
			"-c", "user.name=test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false",
		}, args...)
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	write("internal/a/a.go", "package a\n")
	write("internal/b/b.go", "package b\n")
	git("init", "-q", "-b", "master")
	git("add", ".")
	git("commit", "-q", "-m", "start")
	git("switch", "-q", "-c", "task/x")
	write("internal/c/c.go", "package c\n")
	git("add", ".")
	git("commit", "-q", "-m", "committed on the branch")
	write("internal/a/a.go", "package a\n\n// edited\n") // modified, not committed
	write("internal/d/d.go", "package d\n")              // untracked

	got, err := changedFiles("master")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"internal/a/a.go", "internal/c/c.go", "internal/d/d.go"}
	if !slices.Equal(got, want) {
		t.Errorf("changedFiles = %q, want %q", got, want)
	}
}
