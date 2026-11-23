package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// taskRepo makes a git repository with master holding a task index and the given files, then
// switches to branch and writes changes there, committed. It returns the repository's root.
func taskRepo(t *testing.T, files, changes map[string]string, branch string) string {
	t.Helper()
	top := newRepo(t)
	writeAll := func(set map[string]string) {
		for name, content := range set {
			path := filepath.Join(top, filepath.FromSlash(name))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	commit := func(message string) {
		for _, args := range [][]string{
			{"add", "-A"},
			{"-c", "user.name=test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false",
				"commit", "-q", "--allow-empty", "-m", message},
		} {
			if _, err := gitOutput(top, args...); err != nil {
				t.Fatal(err)
			}
		}
	}
	writeAll(files)
	commit("base")
	if _, err := gitOutput(top, "branch", "-q", "-f", "base"); err != nil {
		t.Fatal(err)
	}
	if _, err := gitOutput(top, "switch", "-q", "-c", branch); err != nil {
		t.Fatal(err)
	}
	writeAll(changes)
	commit("change")
	return top
}

// task renders a task file.
func task(id, status, kind, extra, criteria string) string {
	return "---\nid: " + id + "\ntitle: \"" + id + "\"\nstatus: " + status + "\nkind: " + kind + "\n" + extra +
		"---\n\n# " + id + "\n\n## Acceptance criteria\n\n" + criteria + "\n## Notes\n\n- [ ] not a criterion\n"
}

// renameLine declares the rename C1-02 makes.
const renameLine = "renames: [map.clay_plain=map.clay_flat]\n"

const indexHeader = "| ID | Title | Status |\n| --- | --- | --- |\n"

// markDone sets the status of the index row that starts with row to done.
func markDone(index, row string) string {
	var lines []string
	for line := range strings.SplitSeq(index, "\n") {
		if strings.HasPrefix(line, "| "+row) {
			line = strings.Replace(line, "| todo |", "| done |", 1)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func TestCheckPR(t *testing.T) {
	base := map[string]string{
		"tasks/INDEX.md": indexHeader +
			"| [M3-01](M3/M3-01.md) | a | todo |\n" +
			"| [C1-02](C1/C1-02.md) | b | todo |\n" +
			"| [C1-03](C1/C1-03.md) | c | todo |\n",
		"tasks/M3/M3-01.md":           task("M3-01", "todo", "feature", "", "- [ ] one\n"),
		"tasks/C1/C1-02.md":           task("C1-02", "todo", "debt", renameLine, "- [ ] one\n"),
		"tasks/C1/C1-03.md":           task("C1-03", "todo", "test", "", "- [ ] one\n"),
		"internal/core/rules.go":      "package core\n",
		"internal/core/rules_test.go": "package core\n",
	}
	paths := map[string]string{"M3-01": "tasks/M3/M3-01.md", "C1-02": "tasks/C1/C1-02.md", "C1-03": "tasks/C1/C1-03.md"}
	// done sets a task to done in its file and in the index.
	done := func(id, kind, extra, criteria string) map[string]string {
		row := "[" + id + "](" + strings.TrimPrefix(paths[id], "tasks/") + ")"
		index := markDone(base["tasks/INDEX.md"], row)
		return map[string]string{paths[id]: task(id, "done", kind, extra, criteria), "tasks/INDEX.md": index}
	}
	with := func(changes map[string]string, more map[string]string) map[string]string {
		for name, content := range more {
			changes[name] = content
		}
		return changes
	}

	tests := []struct {
		name         string
		branch       string
		changes      map[string]string
		sameBehavior bool
		problems     []string // fragments, one per expected problem
		warnings     []string
		wantRenames  []string // renames passed to the same-behavior check; nil when it shouldn't run
	}{
		{
			name:   "a finished feature task",
			branch: "task/M3-01-needs",
			changes: with(
				done("M3-01", "feature", "", "- [x] one\n"),
				map[string]string{"internal/core/rules.go": "package core\n\n// x\n"},
			),
		},
		{
			name:     "a task left in todo",
			branch:   "task/M3-01-needs",
			changes:  map[string]string{"internal/core/rules.go": "package core\n\n// x\n"},
			problems: []string{"M3-01 is todo"},
			warnings: []string{"M3-01 has 1 unticked acceptance criteria"},
		},
		{
			name:   "the index disagrees",
			branch: "task/M3-01-needs",
			changes: map[string]string{
				"tasks/M3/M3-01.md": task("M3-01", "done", "feature", "", "- [x] one\n"),
			},
			problems: []string{"M3-01: the task file says done, tasks/INDEX.md says todo"},
		},
		{
			name:     "unticked criteria",
			branch:   "task/M3-01-needs",
			changes:  done("M3-01", "feature", "", "- [x] one\n- [ ] two\n"),
			warnings: []string{"M3-01 has 1 unticked acceptance criteria"},
		},
		{
			name:     "a branch for no task",
			branch:   "task/M9-99-nothing",
			problems: []string{"names no task"},
		},
		{
			name:         "a rename that keeps behavior",
			branch:       "task/C1-02-rename",
			changes:      done("C1-02", "debt", renameLine, "- [x] one\n"),
			sameBehavior: true,
			wantRenames:  []string{"map.clay_plain=map.clay_flat"},
		},
		{
			name:        "a cleanup that changes behavior",
			branch:      "task/C1-02-rename",
			changes:     done("C1-02", "debt", renameLine, "- [x] one\n"),
			problems:    []string{"C1-02 changed behavior"},
			wantRenames: []string{"map.clay_plain=map.clay_flat"},
		},
		{
			name:   "a tests task that changes only tests",
			branch: "task/C1-03-tests",
			changes: with(
				done("C1-03", "test", "", "- [x] one\n"),
				map[string]string{"internal/core/rules_test.go": "package core\n\n// shorter\n"},
			),
		},
		{
			name:   "a tests task that changes code",
			branch: "task/C1-03-tests",
			changes: with(
				done("C1-03", "test", "", "- [x] one\n"),
				map[string]string{"internal/core/rules.go": "package core\n\n// x\n"},
			),
			problems: []string{"C1-03 is a tests task but changes internal/core/rules.go"},
		},
		{
			name:   "not a task branch",
			branch: "dependabot/go_modules/all",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := taskRepo(t, base, test.changes, test.branch)
			var gotRenames []string
			ran := false
			checker := func(_, _ string, renames []string) (string, bool) {
				ran, gotRenames = true, renames
				return "same-behavior report", test.sameBehavior
			}
			result, err := checkPR(root, test.branch, "base", checker)
			if err != nil {
				t.Fatal(err)
			}
			matches := func(kind string, got, want []string) {
				t.Helper()
				if len(got) != len(want) {
					t.Fatalf("%s = %q, want %d like %q", kind, got, len(want), want)
				}
				for index := range want {
					if !strings.Contains(got[index], want[index]) {
						t.Errorf("%s[%d] = %q, want it to contain %q", kind, index, got[index], want[index])
					}
				}
			}
			matches("problems", result.problems, test.problems)
			matches("warnings", result.warnings, test.warnings)
			if ran != (test.wantRenames != nil) || !slices.Equal(gotRenames, test.wantRenames) {
				t.Errorf("same-behavior ran = %v with renames %q; want ran = %v with %q",
					ran, gotRenames, test.wantRenames != nil, test.wantRenames)
			}
		})
	}
}

func TestUntickedCriteria(t *testing.T) {
	body := "## Goal\n\n- [ ] not here\n\n## Acceptance criteria\n\n- [x] done\n- [ ] open\n  - [ ] nested\n\n## Notes\n\n- [ ] not here\n"
	if got := untickedCriteria(body); got != 2 {
		t.Errorf("untickedCriteria = %d, want 2", got)
	}
}

func TestPRCheckReport(t *testing.T) {
	result := prCheck{
		problems: []string{"C1-02 changed behavior:\nline 3 differs"},
		warnings: []string{"M3-01 has 1 unticked"},
	}
	if got := result.report(true); got != "::error title=Task check::C1-02 changed behavior:%0Aline 3 differs\n"+
		"::warning title=Task check::M3-01 has 1 unticked\n" {
		t.Errorf("annotated report = %q", got)
	}
	if got := (prCheck{}).report(false); got != "Task bookkeeping is consistent.\n" {
		t.Errorf("clean report = %q", got)
	}
}
