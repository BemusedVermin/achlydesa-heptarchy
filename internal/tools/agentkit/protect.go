package main

import (
	"os"
	"path/filepath"
	"strings"
)

// taskBranchPrefix starts the name of every branch an agent works a task on.
const taskBranchPrefix = "task/"

// protectedPaths are the files and folders, relative to the repository root, that a task branch
// may not change: the rules agents work under, and the checks and guardrails themselves. The
// design docs in docs/ aren't here: agents keep them in step with what they build. A trailing slash
// marks a folder.
var protectedPaths = []string{
	"AGENTS.md",
	"CLAUDE.md",
	".claude/",
	".github/",
	".golangci.yml",
	"Taskfile.yml",
	"internal/tools/agentkit/",
}

// protectedPath returns the entry of protectedPaths that covers relative, a slash-separated path
// from the repository root, or "" when none does. It ignores case, since Windows does.
func protectedPath(relative string) string {
	relative = strings.ToLower(strings.TrimPrefix(relative, "./"))
	for _, entry := range protectedPaths {
		lower := strings.ToLower(entry)
		folder := strings.HasSuffix(lower, "/")
		if relative == lower || (folder && (relative+"/" == lower || strings.HasPrefix(relative, lower))) {
			return entry
		}
	}
	return ""
}

// checkEdit decides whether an edit to path may go ahead. It refuses edits to protected paths on
// task branches and allows everything else, including edits outside any git checkout.
func checkEdit(path string, reader branchReader) string {
	if path == "" {
		return ""
	}
	path = filepath.Clean(path)
	dir := nearestDir(filepath.Dir(path))
	top, err := gitOutput(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return ""
	}
	branch := reader.branch(top)
	if !strings.HasPrefix(branch, taskBranchPrefix) {
		return ""
	}
	relative, err := filepath.Rel(filepath.FromSlash(top), path)
	if err != nil {
		return ""
	}
	entry := protectedPath(filepath.ToSlash(relative))
	if entry == "" {
		return ""
	}
	return filepath.ToSlash(relative) + " is off limits on task branches (" + entry + "): agents don't change " +
		"the rules they work under, the design docs, or the checks. If your task needs it changed, file a " +
		"follow-up in tasks/followups that says what should change and why, then carry on without it. " +
		"If the task's spec lists this file as a deliverable, stop and tell the owner."
}

// nearestDir returns dir, or its nearest ancestor that exists: a new file's folder may not
// exist yet.
func nearestDir(dir string) string {
	for {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}
		dir = parent
	}
}
