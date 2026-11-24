package main

import (
	"slices"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/tools/internal/gittest"
)

func TestChangedFilesIncludesUncommittedWork(t *testing.T) {
	root := gittest.Repo(t, "master", map[string]string{
		"internal/a/a.go": "package a\n",
		"internal/b/b.go": "package b\n",
	})
	t.Chdir(root)
	gittest.Git(t, root, "switch", "-q", "-c", "task/x")
	gittest.Write(t, root, map[string]string{"internal/c/c.go": "package c\n"})
	gittest.Git(t, root, "add", ".")
	gittest.Git(t, root, "commit", "-q", "-m", "committed on the branch")
	gittest.Write(t, root, map[string]string{
		"internal/a/a.go": "package a\n\n// edited\n",
		"internal/d/d.go": "package d\n",
	})

	got, err := changedFiles("master")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"internal/a/a.go", "internal/c/c.go", "internal/d/d.go"}
	if !slices.Equal(got, want) {
		t.Errorf("changedFiles = %q, want committed, modified and untracked files %q", got, want)
	}
}
