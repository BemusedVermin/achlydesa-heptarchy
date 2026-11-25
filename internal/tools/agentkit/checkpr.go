package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// cleanupID matches the ID of a task in a cleanup milestone, such as C1-03 or C1-03.F1.
var cleanupID = regexp.MustCompile(`^C\d+-`)

// prCheck is what check-pr found: problems fail the check, warnings and comments don't.
type prCheck struct {
	problems []string
	warnings []string
	comments []fileFinding
	notes    []string
}

// checkPR checks a branch against base:
//   - the comments on the lines it changed aren't filler;
//   - tasks/INDEX.md agrees with every task file;
//   - on a task branch, the task is done or blocked, with its unticked criteria reported;
//   - on a cleanup tests task's branch, it changed only tests and task files.
func checkPR(root, branch, base string) (prCheck, error) {
	var result prCheck
	tasks, err := loadTasks(root)
	if err != nil {
		return result, err
	}
	index, err := os.ReadFile(filepath.Join(root, "tasks", "INDEX.md"))
	if err != nil {
		return result, err
	}
	if result.comments, err = commentsSince(root, base); err != nil {
		return result, err
	}
	statuses := loadIndexStatuses(string(index))
	byID := map[string]taskFile{}
	ids := make([]string, 0, len(tasks))
	for _, task := range tasks {
		byID[task.id] = task
		ids = append(ids, task.id)
		if indexed, found := statuses[task.id]; found && indexed != task.status {
			result.problems = append(result.problems, fmt.Sprintf(
				"%s: the task file says %s, tasks/INDEX.md says %s", task.id, task.status, indexed))
		}
	}
	if !strings.HasPrefix(branch, taskBranchPrefix) {
		return result, nil
	}
	id := claimFor(branch, ids)
	if id == "" {
		result.problems = append(result.problems, "branch "+branch+" names no task in tasks/")
		return result, nil
	}
	task := byID[id]
	if task.status != "done" && task.status != "blocked" {
		result.problems = append(result.problems, fmt.Sprintf(
			"%s is %s: a task's pull request sets it to done, or to blocked with the reason in Notes", id, task.status))
	}
	body, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(task.path)))
	if err != nil {
		return result, err
	}
	if unticked := untickedCriteria(string(body)); unticked > 0 {
		result.warnings = append(result.warnings, fmt.Sprintf(
			"%s has %d unticked acceptance criteria; its Notes should say why", id, unticked))
	}
	if cleanupID.MatchString(id) && task.kind == "test" {
		return result, checkOnlyTests(root, base, id, &result)
	}
	return result, nil
}

// commentsSince reviews the comments on the Go lines changed since the branch left base.
func commentsSince(root, base string) ([]fileFinding, error) {
	from, err := gitOutput(root, "merge-base", base, "HEAD")
	if err != nil {
		return nil, err
	}
	changed, err := changedLines(root, from, "*.go")
	if err != nil {
		return nil, err
	}
	return commentsOnChangedLines(root, changed), nil
}

// checkOnlyTests adds a problem for each file a tests task changed that isn't a test, golden file
// or task file.
func checkOnlyTests(root, base, id string, result *prCheck) error {
	changed, err := gitOutput(root, "diff", "--name-only", base+"...HEAD")
	if err != nil {
		return err
	}
	for path := range strings.SplitSeq(changed, "\n") {
		testOnly := strings.HasSuffix(path, "_test.go") || strings.Contains("/"+path, "/testdata/") ||
			strings.HasPrefix(path, "tasks/")
		if path != "" && !testOnly {
			result.problems = append(result.problems, fmt.Sprintf(
				"%s is a tests task but changes %s: only _test.go files, testdata/ and tasks/ may change", id, path))
		}
	}
	return nil
}

// untickedCriteria counts the "- [ ]" items in a task file's "## Acceptance criteria" section.
func untickedCriteria(body string) int {
	count, inSection := 0, false
	for line := range strings.SplitSeq(strings.ReplaceAll(body, "\r\n", "\n"), "\n") {
		if strings.HasPrefix(line, "## ") {
			inSection = strings.HasPrefix(line, "## Acceptance criteria")
			continue
		}
		if inSection && strings.HasPrefix(strings.TrimSpace(line), "- [ ]") {
			count++
		}
	}
	return count
}

// runCheckPR is the check-pr command: it checks the current checkout and prints what it found.
// In GitHub Actions, problems and warnings become annotations.
func runCheckPR(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("check-pr", flag.ContinueOnError)
	flags.SetOutput(stderr)
	branch := flags.String("branch", "", "the branch to check (default: the current branch)")
	base := flags.String("base", "origin/master", "the branch the pull request merges into")
	if err := flags.Parse(args); err != nil {
		return exitFailed
	}
	dir, err := os.Getwd()
	if err != nil {
		warn(stderr, "agentkit check-pr:", err)
		return exitFailed
	}
	top, err := gitOutput(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		warn(stderr, "agentkit check-pr:", err)
		return exitFailed
	}
	root := filepath.FromSlash(top)
	if *branch == "" {
		*branch = gitRepo{}.branch(root)
	}
	result, err := checkPR(root, *branch, *base)
	if err != nil {
		warn(stderr, "agentkit check-pr:", err)
		return exitFailed
	}
	if _, err := io.WriteString(stdout, result.report(os.Getenv("GITHUB_ACTIONS") == "true")); err != nil {
		warn(stderr, "agentkit check-pr:", err)
		return exitFailed
	}
	if len(result.problems) > 0 {
		return exitFailed
	}
	return exitOK
}

// report lists the findings, as GitHub annotations when annotate is set.
func (result prCheck) report(annotate bool) string {
	var out strings.Builder
	line := func(level, title, text string) {
		if annotate {
			// An annotation is one line; GitHub decodes %0A back into new lines.
			out.WriteString("::" + level + " title=" + title + "::" + strings.ReplaceAll(text, "\n", "%0A") + "\n")
			return
		}
		out.WriteString(level + ": " + text + "\n")
	}
	for _, problem := range result.problems {
		line("error", "Task check", problem)
	}
	for _, warning := range result.warnings {
		line("warning", "Task check", warning)
	}
	for _, comment := range result.comments {
		if annotate {
			out.WriteString("::warning file=" + comment.file + ",line=" + strconv.Itoa(comment.first) +
				",title=Comment (see .claude/rules/comments.md)::" + comment.why + "\n")
			continue
		}
		out.WriteString("warning: " + comment.String() + "\n")
	}
	for _, note := range result.notes {
		out.WriteString(note + "\n")
	}
	if len(result.problems) == 0 {
		out.WriteString("Task bookkeeping is consistent.\n")
	}
	return out.String()
}
