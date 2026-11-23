package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestParseWorktrees(t *testing.T) {
	porcelain := strings.Join([]string{
		"worktree /repo",
		"HEAD 1111",
		"branch refs/heads/master",
		"",
		"worktree /repo/.claude/worktrees/agent-a",
		"HEAD 2222",
		"branch refs/heads/task/M2-07-hostility",
		"locked claude agent agent-a (pid 42)",
		"",
		"worktree /repo/.claude/worktrees/gone",
		"HEAD 3333",
		"detached",
		"prunable gitdir file points to non-existent location",
		"",
	}, "\n")
	got := parseWorktrees(porcelain)
	want := []worktree{
		{path: filepath.FromSlash("/repo"), head: "1111", branch: "master"},
		{
			path: filepath.FromSlash("/repo/.claude/worktrees/agent-a"), head: "2222",
			branch: "task/M2-07-hostility", locked: true, lockNote: "claude agent agent-a (pid 42)",
		},
		{path: filepath.FromSlash("/repo/.claude/worktrees/gone"), head: "3333", prunable: true},
	}
	if !slices.Equal(got, want) {
		t.Errorf("parseWorktrees =\n%+v\nwant\n%+v", got, want)
	}
}

func TestParseCleanFlags(t *testing.T) {
	tests := []struct {
		flags   []string
		want    cleanOptions
		wantErr bool
	}{
		{flags: nil, want: cleanOptions{}},
		{flags: []string{"--dry-run"}, want: cleanOptions{dryRun: true}},
		{flags: []string{"-n", "--no-fetch"}, want: cleanOptions{dryRun: true, noFetch: true}},
		{flags: []string{"--force"}, wantErr: true},
	}
	for _, test := range tests {
		got, err := parseCleanFlags(test.flags)
		if (err != nil) != test.wantErr || got != test.want {
			t.Errorf("parseCleanFlags(%q) = %+v, %v; want %+v, error %v", test.flags, got, err, test.want, test.wantErr)
		}
	}
}

func TestPlural(t *testing.T) {
	tests := []struct {
		count int
		noun  string
		want  string
	}{
		{1, "worktree", "1 worktree"},
		{0, "worktree", "0 worktrees"},
		{3, "branch", "3 branches"},
	}
	for _, test := range tests {
		if got := plural(test.count, test.noun); got != test.want {
			t.Errorf("plural(%d, %q) = %q, want %q", test.count, test.noun, got, test.want)
		}
	}
}

// clone is a checkout cloned from a local bare repository standing in for GitHub.
type clone struct {
	t      *testing.T
	origin string // the bare repository
	top    string // the clone's root, as git reports it
}

// newClone makes a bare repository with one commit on master and clones it. It skips the test when
// git isn't installed.
func newClone(t *testing.T) clone {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	base := t.TempDir()
	origin := filepath.Join(base, "origin.git")
	seed := filepath.Join(base, "seed")
	repo := clone{t: t, origin: origin, top: seed}
	repo.git(base, "init", "-q", "--bare", "-b", "master", origin)
	repo.git(base, "init", "-q", "-b", "master", seed)
	repo.configure(seed)
	repo.commit(seed, "notes.txt", "first line\n", "start")
	repo.git(seed, "push", "-q", origin, "master")
	repo.git(base, "clone", "-q", origin, filepath.Join(base, "main"))
	repo.top = filepath.FromSlash(repo.git(filepath.Join(base, "main"), "rev-parse", "--show-toplevel"))
	repo.configure(repo.top)
	return repo
}

// git runs git in dir and fails the test if it fails.
func (repo clone) git(dir string, args ...string) string {
	repo.t.Helper()
	out, err := gitOutput(dir, args...)
	if err != nil {
		repo.t.Fatal(err)
	}
	return out
}

// configure sets the identity commits need, and turns off commit signing.
func (repo clone) configure(dir string) {
	repo.t.Helper()
	repo.git(dir, "config", "user.name", "test")
	repo.git(dir, "config", "user.email", "test@example.com")
	repo.git(dir, "config", "commit.gpgsign", "false")
}

// commit writes a file in dir and commits it.
func (repo clone) commit(dir, file, content, message string) {
	repo.t.Helper()
	if err := os.WriteFile(filepath.Join(dir, file), []byte(content), 0o644); err != nil {
		repo.t.Fatal(err)
	}
	repo.git(dir, "add", file)
	repo.git(dir, "commit", "-q", "-m", message)
}

// worktree adds a worktree under .claude/worktrees on a new branch made from origin/master, and
// returns its path.
func (repo clone) worktree(name, branch string) string {
	repo.t.Helper()
	path := filepath.Join(repo.top, ".claude", "worktrees", name)
	repo.git(repo.top, "worktree", "add", "-q", "-b", branch, path, "origin/master")
	return path
}

// branches returns the clone's local branches.
func (repo clone) branches() []string {
	repo.t.Helper()
	return strings.Split(repo.git(repo.top, "for-each-ref", "--format=%(refname:short)", "refs/heads/"), "\n")
}

// setUpLeftovers builds one worktree, or branch, for each case the cleanup must tell apart.
func setUpLeftovers(repo clone) {
	repo.t.Helper()

	// A task whose pull request was squash-merged, and its branch deleted on GitHub.
	merged := repo.worktree("agent-merged", "task/M9-01-merged")
	repo.commit(merged, "merged.txt", "merged work\n", "M9-01 part one")
	repo.commit(merged, "notes.txt", "first line\nmerged edit\n", "M9-01 part two")
	repo.git(merged, "push", "-q", "origin", "task/M9-01-merged")
	squash := filepath.Join(filepath.Dir(repo.top), "squash")
	repo.git(repo.top, "clone", "-q", repo.origin, squash)
	repo.configure(squash)
	repo.git(squash, "merge", "-q", "--squash", "origin/task/M9-01-merged")
	repo.git(squash, "commit", "-q", "-m", "M9-01 (#1)")
	repo.git(squash, "push", "-q", "origin", "master", ":task/M9-01-merged")

	// A task with its pull request still open: pushed, not merged.
	open := repo.worktree("agent-open", "task/M9-02-open")
	repo.commit(open, "open.txt", "open work\n", "M9-02")
	repo.git(open, "push", "-q", "origin", "task/M9-02-open")

	// An agent's worktree with nothing in it: Claude Code's own branch, at master.
	repo.worktree("agent-empty", "worktree-agent-empty")

	// Uncommitted work.
	dirty := repo.worktree("agent-dirty", "task/M9-04-dirty")
	if err := os.WriteFile(filepath.Join(dirty, "draft.txt"), []byte("half done\n"), 0o644); err != nil {
		repo.t.Fatal(err)
	}

	// Committed work that was never pushed.
	local := repo.worktree("agent-local", "task/M9-05-local")
	repo.commit(local, "local.txt", "local work\n", "M9-05")

	// A worktree an agent is still using: Claude Code locks it.
	running := repo.worktree("agent-running", "worktree-agent-running")
	repo.git(repo.top, "worktree", "lock", "--reason", "claude agent agent-running", running)

	// A worktree whose folder was deleted by hand.
	gone := repo.worktree("agent-gone", "worktree-agent-gone")
	if err := os.RemoveAll(gone); err != nil {
		repo.t.Fatal(err)
	}

	// A worktree outside .claude/worktrees: not the cleanup's business.
	repo.git(repo.top, "worktree", "add", "-q", "-b", "task/M9-08-outside",
		filepath.Join(filepath.Dir(repo.top), "outside"), "origin/master")

	// Leftover branches with no worktree: one at an old master, one with unmerged work.
	repo.git(repo.top, "branch", "worktree-agent-old", "origin/master")
	repo.git(repo.top, "branch", "task/M9-09-unmerged", "task/M9-05-local")

	// A file an agent left among the worktrees.
	stray := filepath.Join(repo.top, ".claude", "worktrees", "pr-body.md")
	if err := os.WriteFile(stray, []byte("Summary\n"), 0o644); err != nil {
		repo.t.Fatal(err)
	}
}

func TestCleanWorktrees(t *testing.T) {
	repo := newClone(t)
	setUpLeftovers(repo)

	wantRemoved := map[string]string{
		"agent-merged": "everything in it is in origin/master",
		"agent-open":   "its commits are pushed",
		"agent-empty":  "everything in it is in origin/master",
		"agent-gone":   "its folder is already gone",
	}
	wantKept := map[string]string{
		"agent-dirty":   "1 uncommitted or untracked file",
		"agent-local":   "1 commit neither pushed nor in origin/master",
		"agent-running": "locked because: claude agent agent-running",
	}
	wantBranches := []string{"task/M9-01-merged", "worktree-agent-empty", "worktree-agent-gone", "worktree-agent-old"}

	check := func(name string, result cleanup) {
		t.Helper()
		if result.base != "origin/master" {
			t.Errorf("%s: base = %q, want origin/master", name, result.base)
		}
		seen := map[string]bool{}
		for _, judged := range result.worktrees {
			seen[judged.name] = true
			want, removed := wantRemoved[judged.name]
			if !removed {
				want = wantKept[judged.name]
			}
			if judged.remove != removed || !strings.Contains(judged.reason, want) {
				t.Errorf("%s: %s: remove = %v, reason %q; want remove = %v, reason containing %q",
					name, judged.name, judged.remove, judged.reason, removed, want)
			}
		}
		if len(seen) != len(wantRemoved)+len(wantKept) {
			t.Errorf("%s: judged %d worktrees, want %d (only those in .claude/worktrees)",
				name, len(seen), len(wantRemoved)+len(wantKept))
		}
		if got := slices.Sorted(slices.Values(result.branches)); !slices.Equal(got, wantBranches) {
			t.Errorf("%s: branches = %q, want %q", name, got, wantBranches)
		}
		if !slices.Equal(result.strays, []string{"pr-body.md"}) {
			t.Errorf("%s: strays = %q, want [pr-body.md]", name, result.strays)
		}
		if len(result.failures) > 0 {
			t.Errorf("%s: failures = %q", name, result.failures)
		}
	}

	before := repo.branches()
	dry, err := cleanWorktrees(repo.top, cleanOptions{dryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	check("dry run", dry)
	if after := repo.branches(); !slices.Equal(after, before) {
		t.Errorf("dry run changed the branches: %q, then %q", before, after)
	}
	if _, err := os.Stat(filepath.Join(repo.top, ".claude", "worktrees", "agent-open")); err != nil {
		t.Errorf("dry run removed a worktree: %v", err)
	}
	if report := dry.report(); !strings.Contains(report, "Would remove 4 worktrees:") ||
		!strings.Contains(report, "Would delete 4 branches") {
		t.Errorf("dry run report doesn't say what it would do:\n%s", report)
	}

	result, err := cleanWorktrees(repo.top, cleanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	check("cleanup", result)
	for name := range wantRemoved {
		if _, err := os.Stat(filepath.Join(repo.top, ".claude", "worktrees", name)); !os.IsNotExist(err) {
			t.Errorf("%s is still on disk: %v", name, err)
		}
	}
	for name := range wantKept {
		if _, err := os.Stat(filepath.Join(repo.top, ".claude", "worktrees", name)); err != nil {
			t.Errorf("%s was removed: %v", name, err)
		}
	}
	branches := repo.branches()
	for _, branch := range wantBranches {
		if slices.Contains(branches, branch) {
			t.Errorf("branch %s wasn't deleted", branch)
		}
	}
	for _, branch := range []string{"task/M9-02-open", "task/M9-04-dirty", "task/M9-08-outside", "task/M9-09-unmerged"} {
		if !slices.Contains(branches, branch) {
			t.Errorf("branch %s was deleted", branch)
		}
	}
	if listing := repo.git(repo.top, "worktree", "list", "--porcelain"); strings.Contains(listing, "agent-gone") {
		t.Errorf("the deleted worktree wasn't pruned:\n%s", listing)
	}

	again, err := cleanWorktrees(repo.top, cleanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if report := again.report(); !strings.Contains(report, "Kept 3 worktrees") || strings.Contains(report, "Removed") {
		t.Errorf("a second cleanup should only list what it keeps:\n%s", report)
	}
}

func TestCleanWorktreesFromInsideOne(t *testing.T) {
	repo := newClone(t)
	inside := repo.worktree("agent-here", "worktree-agent-here")
	result, err := cleanWorktrees(inside, cleanOptions{dryRun: true, noFetch: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.worktrees) != 1 || result.worktrees[0].remove ||
		!strings.Contains(result.worktrees[0].reason, "from inside it") {
		t.Errorf("worktrees = %+v, want agent-here kept because the cleanup runs from inside it", result.worktrees)
	}
	if len(result.branches) != 0 {
		t.Errorf("branches = %q, want none: worktree-agent-here is checked out", result.branches)
	}
}

func TestRunCleanWorktrees(t *testing.T) {
	repo := newClone(t)
	repo.worktree("agent-done", "worktree-agent-done")
	t.Chdir(repo.top)
	var stdout, stderr strings.Builder
	status := run([]string{"clean-worktrees", "--no-fetch"}, strings.NewReader(""), &stdout, &stderr)
	if status != exitOK || stderr.Len() != 0 || !strings.Contains(stdout.String(), "Removed 1 worktree:") {
		t.Errorf("run = %d, stdout %q, stderr %q; want success and agent-done removed",
			status, stdout.String(), stderr.String())
	}
}

func TestCleanReportWhenEmpty(t *testing.T) {
	repo := newClone(t)
	result, err := cleanWorktrees(repo.top, cleanOptions{noFetch: true})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := result.report(), "Nothing to clean up in .claude/worktrees.\n"; got != want {
		t.Errorf("report = %q, want %q", got, want)
	}
}
