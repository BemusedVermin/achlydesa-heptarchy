package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// worktreesDir is where Claude Code makes its worktrees, relative to the main checkout. The cleanup
// touches no worktree outside it.
const worktreesDir = ".claude/worktrees"

// leftoverBranches are the patterns, for git for-each-ref, of the local branches agents leave
// behind: Claude Code's own worktree branches, and the task branches the do-task skill makes.
var leftoverBranches = []string{"refs/heads/worktree-*", "refs/heads/task/"}

// cleanOptions are the command-line flags of clean-worktrees.
type cleanOptions struct {
	dryRun  bool // report what would happen, and change nothing
	noFetch bool // judge against what was fetched before, without fetching origin first
}

// parseCleanFlags reads the command-line flags of clean-worktrees.
func parseCleanFlags(flags []string) (cleanOptions, error) {
	var opts cleanOptions
	for _, flag := range flags {
		switch flag {
		case "--dry-run", "-n":
			opts.dryRun = true
		case "--no-fetch":
			opts.noFetch = true
		default:
			return cleanOptions{}, fmt.Errorf("unknown flag %q: clean-worktrees takes --dry-run and --no-fetch", flag)
		}
	}
	return opts, nil
}

// worktree is one entry of "git worktree list --porcelain".
type worktree struct {
	path     string // the worktree's root, as git reports it, in this system's path form
	head     string // the commit it has checked out
	branch   string // the branch it has checked out, without refs/heads/; empty when HEAD is detached
	locked   bool   // Claude Code locks a worktree while a session or agent is using it
	lockNote string // the reason given when it was locked, if any
	prunable bool   // its folder is gone, so "git worktree prune" drops it
}

// parseWorktrees reads the output of "git worktree list --porcelain". The main checkout comes first.
func parseWorktrees(porcelain string) []worktree {
	var list []worktree
	for line := range strings.SplitSeq(porcelain, "\n") {
		key, value, _ := strings.Cut(strings.TrimRight(line, "\r"), " ")
		if key == "worktree" {
			list = append(list, worktree{path: filepath.FromSlash(value)})
			continue
		}
		if len(list) == 0 {
			continue
		}
		current := &list[len(list)-1]
		switch key {
		case "HEAD":
			current.head = value
		case "branch":
			current.branch = strings.TrimPrefix(value, "refs/heads/")
		case "locked":
			current.locked = true
			current.lockNote = value
		case "prunable":
			current.prunable = true
		}
	}
	return list
}

// judgement is what the cleanup decided about one worktree.
type judgement struct {
	worktree
	name   string // the worktree's path relative to .claude/worktrees
	remove bool   // whether removing it loses nothing
	reason string // why it goes or stays
}

// cleanup is what clean-worktrees did or, in a dry run, would do.
type cleanup struct {
	dryRun    bool
	base      string      // the branch work counts as merged into, such as origin/master
	notes     []string    // problems that didn't stop the cleanup, such as a failed fetch
	worktrees []judgement // every worktree in .claude/worktrees
	branches  []string    // leftover branches whose work is all in base, deleted or to delete
	failures  []string    // removals and deletions git refused
	strays    []string    // entries in .claude/worktrees that aren't worktrees, left alone
}

// cleanWorktrees removes the worktrees in .claude/worktrees whose work is all in the default branch
// or pushed, then the leftover agent branches whose work is all in the default branch. It keeps
// every worktree that is locked, has uncommitted or untracked files, or has commits that are
// neither pushed nor merged, and says why. dir is any directory in the repository.
func cleanWorktrees(dir string, opts cleanOptions) (cleanup, error) {
	result := cleanup{dryRun: opts.dryRun}
	listing, err := gitOutput(dir, "worktree", "list", "--porcelain")
	if err != nil {
		return result, err
	}
	all := parseWorktrees(listing)
	if len(all) == 0 {
		return result, errors.New("git listed no worktrees")
	}
	mainCheckout := all[0].path
	if !opts.noFetch {
		if _, err := gitOutput(mainCheckout, "fetch", "--prune", "--quiet", "origin"); err != nil {
			result.notes = append(result.notes,
				"Couldn't fetch origin, so work was judged against what was fetched before: "+err.Error())
		}
	}
	if result.base, err = defaultBranch(mainCheckout); err != nil {
		return result, err
	}
	current, err := gitOutput(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return result, err
	}
	root := filepath.Join(mainCheckout, filepath.FromSlash(worktreesDir))
	for _, tree := range all[1:] {
		name, ok := relativeTo(root, tree.path)
		if !ok || name == "." {
			continue
		}
		judged := judgement{worktree: tree, name: name}
		judged.remove, judged.reason = judge(tree, result.base, filepath.FromSlash(current))
		result.worktrees = append(result.worktrees, judged)
	}
	if !opts.dryRun {
		result.removeWorktrees(mainCheckout)
	}
	if err := result.findBranches(mainCheckout, all); err != nil {
		return result, err
	}
	if !opts.dryRun {
		result.deleteBranches(mainCheckout)
	}
	result.strays = strays(root, all)
	return result, nil
}

// defaultBranch returns origin's default branch as a remote-tracking ref, such as origin/master.
func defaultBranch(dir string) (string, error) {
	ref, err := gitOutput(dir, "symbolic-ref", "--short", "-q", "refs/remotes/origin/HEAD")
	if err == nil && ref != "" {
		return ref, nil
	}
	for _, ref := range []string{"origin/master", "origin/main"} {
		if _, err := gitOutput(dir, "rev-parse", "--verify", "-q", ref+"^{commit}"); err == nil {
			return ref, nil
		}
	}
	return "", errors.New("can't find origin's default branch (no origin/HEAD, origin/master or origin/main): " +
		"run git fetch origin, then try again")
}

// judge decides whether removing a worktree loses nothing, and says why. current is the root of the
// checkout the cleanup runs from.
func judge(tree worktree, base, current string) (remove bool, reason string) {
	switch {
	case tree.locked:
		why := "locked, so a Claude session or agent may still be using it. If none is, run " +
			"git worktree unlock \"" + tree.path + "\" and clean again"
		if tree.lockNote != "" {
			why += " (locked because: " + tree.lockNote + ")"
		}
		return false, why
	case tree.prunable:
		return true, "its folder is already gone"
	case tree.head == "":
		return false, "it has no commit checked out"
	}
	if name, ok := relativeTo(tree.path, current); ok && name == "." {
		return false, "you ran the cleanup from inside it"
	}
	status, err := gitOutput(tree.path, "status", "--porcelain")
	if err != nil {
		return false, "couldn't read its git status: " + err.Error()
	}
	if status != "" {
		return false, plural(len(strings.Split(status, "\n")), "uncommitted or untracked file")
	}
	if alreadyIn(tree.path, tree.head, base) {
		return true, "everything in it is in " + base
	}
	if pushed(tree.path, tree.head) {
		return true, "its commits are pushed"
	}
	count, err := gitOutput(tree.path, "rev-list", "--count", tree.head, "--not", "--remotes")
	if number, convErr := strconv.Atoi(count); err == nil && convErr == nil {
		return false, plural(number, "commit") + " neither pushed nor in " + base
	}
	return false, "it has commits that are neither pushed nor in " + base
}

// alreadyIn reports whether everything rev holds is already in base. A squash merge puts a branch's
// changes on base as one new commit, so the branch's own commits never become part of base; merging
// the branch into base then changes nothing, which is what the second test looks for.
func alreadyIn(dir, rev, base string) bool {
	if _, err := gitOutput(dir, "merge-base", "--is-ancestor", rev, base); err == nil {
		return true
	}
	// merge-tree fails on a conflict, and on a git older than 2.38; either way, nothing is assumed.
	merged, err := gitOutput(dir, "merge-tree", "--write-tree", base, rev)
	if err != nil {
		return false
	}
	baseTree, err := gitOutput(dir, "rev-parse", base+"^{tree}")
	if err != nil {
		return false
	}
	mergedTree, _, _ := strings.Cut(merged, "\n")
	return mergedTree == baseTree
}

// pushed reports whether a remote-tracking branch holds rev, so its commits are on the remote.
func pushed(dir, rev string) bool {
	refs, err := gitOutput(dir, "for-each-ref", "--contains", rev, "--format=%(refname)", "refs/remotes/")
	return err == nil && refs != ""
}

// removeWorktrees removes the worktrees judged safe to remove, then prunes git's records of
// worktrees whose folders are gone. A removal git refuses keeps its worktree, with the reason.
func (result *cleanup) removeWorktrees(mainCheckout string) {
	for index := range result.worktrees {
		judged := &result.worktrees[index]
		if !judged.remove || judged.prunable {
			continue
		}
		if _, err := gitOutput(mainCheckout, "worktree", "remove", judged.path); err != nil {
			judged.remove = false
			judged.reason = "git couldn't remove it (is a program using its files?): " + err.Error()
		}
	}
	if _, err := gitOutput(mainCheckout, "worktree", "prune"); err != nil {
		result.failures = append(result.failures, "git worktree prune: "+err.Error())
	}
}

// findBranches lists the leftover agent branches whose work is all in the default branch and which
// no remaining worktree has checked out.
func (result *cleanup) findBranches(mainCheckout string, all []worktree) error {
	args := append([]string{"for-each-ref", "--format=%(refname:short)"}, leftoverBranches...)
	refs, err := gitOutput(mainCheckout, args...)
	if err != nil {
		return err
	}
	var inUse []string
	for _, tree := range all {
		gone := slices.ContainsFunc(result.worktrees, func(judged judgement) bool {
			return judged.remove && judged.path == tree.path
		})
		if !gone && tree.branch != "" {
			inUse = append(inUse, tree.branch)
		}
	}
	for branch := range strings.SplitSeq(refs, "\n") {
		if branch != "" && !slices.Contains(inUse, branch) && alreadyIn(mainCheckout, branch, result.base) {
			result.branches = append(result.branches, branch)
		}
	}
	return nil
}

// deleteBranches deletes the branches findBranches listed. A deletion git refuses is reported as a
// failure, and the branch is left out of the deleted ones.
func (result *cleanup) deleteBranches(mainCheckout string) {
	var deleted []string
	for _, branch := range result.branches {
		// -D, because git's own merge test for -d doesn't recognise a squash merge. findBranches
		// has already checked that everything on the branch is in the default branch.
		if _, err := gitOutput(mainCheckout, "branch", "-D", branch); err != nil {
			result.failures = append(result.failures, "couldn't delete branch "+branch+": "+err.Error())
			continue
		}
		deleted = append(deleted, branch)
	}
	result.branches = deleted
}

// strays lists the entries in root that aren't worktrees, such as notes an agent left there.
func strays(root string, all []worktree) []string {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var found []string
	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())
		isWorktree := slices.ContainsFunc(all, func(tree worktree) bool {
			name, ok := relativeTo(tree.path, path)
			return ok && name == "."
		})
		if !isWorktree {
			found = append(found, entry.Name())
		}
	}
	return found
}

// relativeTo returns path relative to root, with forward slashes, and whether path is root or
// inside it. On Windows the comparison ignores case, as the file system does.
func relativeTo(root, path string) (string, bool) {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return filepath.ToSlash(rel), true
}

// plural returns a count with a noun, made plural unless the count is one.
func plural(count int, noun string) string {
	if count == 1 {
		return "1 " + noun
	}
	if strings.HasSuffix(noun, "ch") || strings.HasSuffix(noun, "s") {
		return strconv.Itoa(count) + " " + noun + "es"
	}
	return strconv.Itoa(count) + " " + noun + "s"
}

// report describes the cleanup for the person who ran it.
func (result cleanup) report() string {
	removeVerb, deleteVerb := "Removed", "Deleted"
	if result.dryRun {
		removeVerb, deleteVerb = "Would remove", "Would delete"
	}
	var removed, kept []string
	for _, judged := range result.worktrees {
		line := judged.name
		if judged.branch != "" {
			line += " (" + judged.branch + ")"
		}
		line += ": " + judged.reason
		if judged.remove {
			removed = append(removed, line)
		} else {
			kept = append(kept, line)
		}
	}
	strayPaths := make([]string, 0, len(result.strays))
	for _, stray := range result.strays {
		strayPaths = append(strayPaths, worktreesDir+"/"+stray)
	}
	var report strings.Builder
	section := func(heading string, lines []string) {
		if len(lines) == 0 {
			return
		}
		report.WriteString(heading + "\n")
		for _, line := range lines {
			report.WriteString("  " + line + "\n")
		}
		report.WriteString("\n")
	}
	section("Note:", result.notes)
	if len(removed)+len(kept)+len(result.branches)+len(result.strays)+len(result.failures) == 0 {
		report.WriteString("Nothing to clean up in " + worktreesDir + ".\n")
	}
	section(removeVerb+" "+plural(len(removed), "worktree")+":", removed)
	section("Kept "+plural(len(kept), "worktree")+", so nothing is lost:", kept)
	section(deleteVerb+" "+plural(len(result.branches), "branch")+" whose work is all in "+result.base+":",
		result.branches)
	section("Failed:", result.failures)
	section("Not worktrees, so left alone (delete them yourself if you don't need them):", strayPaths)
	return strings.TrimRight(report.String(), "\n") + "\n"
}
