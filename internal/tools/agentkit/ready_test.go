package main

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestParseFrontMatter(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		want  map[string]string
		found bool
	}{
		{"task file", "---\nid: M2-07\ntitle: \"Hostility: alert\"\nstatus: todo\ndepends_on: [M2-06]\n---\n# body",
			map[string]string{"id": "M2-07", "title": "Hostility: alert", "status": "todo", "depends_on": "[M2-06]"},
			true},
		{"comments", "---\nstatus: todo   # proposed | todo\ntitle: 'a # b'\n---\n",
			map[string]string{"status": "todo", "title": "a # b"}, true},
		{"windows line ends", "---\r\nid: M1-01\r\n---\r\n", map[string]string{"id": "M1-01"}, true},
		{"no front matter", "# Tasks\n\nid: M1-03\n", nil, false},
		{"unterminated", "---\nid: M1-01\n", nil, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, found := parseFrontMatter(test.text)
			if found != test.found || !reflect.DeepEqual(got, test.want) {
				t.Errorf("parseFrontMatter() = %v, %v; want %v, %v", got, found, test.want, test.found)
			}
		})
	}
}

func TestParseList(t *testing.T) {
	tests := []struct {
		value string
		want  []string
	}{
		{"[]", nil},
		{"[M1-01]", []string{"M1-01"}},
		{"[M1-11, M1-13]", []string{"M1-11", "M1-13"}},
		{`["M1-11", 'M1-13']`, []string{"M1-11", "M1-13"}},
		{"null", nil},
		{"none", nil},
	}
	for _, test := range tests {
		if got := parseList(test.value); !reflect.DeepEqual(got, test.want) {
			t.Errorf("parseList(%q) = %q, want %q", test.value, got, test.want)
		}
	}
}

func TestCompareIDs(t *testing.T) {
	ids := []string{"M2-13.F1", "M1-10", "PM-STORY-01", "M2-13", "M1-9", "M1-01", "M2-02.F1"}
	slices.SortFunc(ids, compareIDs)
	want := []string{"M1-01", "M1-9", "M1-10", "M2-02.F1", "M2-13", "M2-13.F1", "PM-STORY-01"}
	if !reflect.DeepEqual(ids, want) {
		t.Errorf("sorted = %q, want %q", ids, want)
	}
}

func TestClaimFor(t *testing.T) {
	ids := []string{"M2-13", "M2-13.F1", "M2-01", "PM-STORY-01"}
	tests := []struct {
		branch string
		want   string
	}{
		{"task/M2-13-rules", "M2-13"},
		{"task/M2-13.F1-schema", "M2-13.F1"},
		{"origin/task/M2-01", "M2-01"},
		{"task/PM-STORY-01-scenes", "PM-STORY-01"},
		{"task/M2-1-x", ""},
		{"worktree-M2-13", ""},
	}
	for _, test := range tests {
		if got := claimFor(test.branch, ids); got != test.want {
			t.Errorf("claimFor(%q) = %q, want %q", test.branch, got, test.want)
		}
	}
}

func TestLoadIndexStatuses(t *testing.T) {
	index := strings.Join([]string{
		"## M2",
		"",
		"| ID | Title | Status | Kind |",
		"| --- | --- | --- | --- |",
		"| [M2-06](M2-combat/M2-06.md) | Forms | todo | content |",
		"",
		"| Milestone | Brief |",
		"| --- | --- |",
		"| M3 · Agents | [M3-agents/README.md](M3-agents/README.md) |",
		"",
		"| ID | Title | Status | Kind | Milestone | Discovered by |",
		"| --- | --- | --- | --- | --- | --- |",
		"| [M2-13.F1](followups/M2-13.F1.md) | Schema tests | done | test | unplanned | M2-13 |",
	}, "\n")
	want := map[string]string{"M2-06": "todo", "M2-13.F1": "done"}
	if got := loadIndexStatuses(index); !reflect.DeepEqual(got, want) {
		t.Errorf("loadIndexStatuses() = %v, want %v", got, want)
	}
}

func TestReadyReport(t *testing.T) {
	tasks := []taskFile{
		{id: "M2-05", title: "Triggers", status: "done", path: "tasks/M2-05.md"},
		{id: "M2-06", title: "Forms", status: "todo", dependsOn: []string{"M2-05"}, path: "tasks/M2-06.md"},
		{id: "M2-07", title: "Hostility", status: "todo", dependsOn: []string{"M2-06"}, path: "tasks/M2-07.md"},
		{id: "M2-12", title: "Hot reload", status: "todo", path: "tasks/M2-12.md"},
		{id: "M2-13", title: "Rules", status: "in-progress", path: "tasks/M2-13.md"},
		{id: "M2-13.F1", title: "Schema tests", status: "proposed", path: "tasks/f.md"},
	}
	report := readyReport(tasks, []string{"origin/task/M2-12-reload"}, map[string]string{"M2-06": "done"})
	for _, want := range []string{
		"Ready: todo, every dependency done, and no task branch yet\n  M2-06      Forms  (tasks/M2-06.md)\n",
		"Claimed: todo, but a task branch exists (another agent may be on it)\n" +
			"  M2-12      Hot reload  (tasks/M2-12.md)  on origin/task/M2-12-reload\n",
		"In progress or blocked\n  M2-13      Rules  [in-progress]\n",
		"Waiting on dependencies\n  M2-07      needs M2-06\n",
		"  M2-06: the task file says todo, INDEX.md says done\n",
	} {
		if !strings.Contains(report, want) {
			t.Errorf("report lacks %q:\n%s", want, report)
		}
	}
	if strings.Contains(report, "M2-13.F1") {
		t.Errorf("report lists a proposed task:\n%s", report)
	}
}

func TestReadyReportNothingReady(t *testing.T) {
	report := readyReport([]taskFile{{id: "M1-01", status: "done"}}, nil, nil)
	if report != "Ready: todo, every dependency done, and no task branch yet\n  none\n" {
		t.Errorf("report = %q", report)
	}
}

func TestLoadTasks(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"tasks/TEMPLATE.md":           "---\nid: M0-00\nstatus: todo\n---\n",
		"tasks/README.md":             "# Tasks\n\n```yaml\n---\nid: M1-03\n---\n```\n",
		"tasks/M1/M1-10.md":           "---\nid: M1-10\ntitle: Ten\nstatus: done\ndepends_on: []\n---\n",
		"tasks/M1/M1-02.md":           "---\nid: M1-02\ntitle: Two\nstatus: todo\ndepends_on: [M1-10]\n---\n",
		"tasks/followups/M1-02.F1.md": "---\nid: M1-02.F1\nstatus: proposed\n---\n",
	}
	for name, text := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	tasks, err := loadTasks(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []taskFile{
		{id: "M1-02", title: "Two", status: "todo", dependsOn: []string{"M1-10"}, path: "tasks/M1/M1-02.md"},
		{id: "M1-02.F1", status: "proposed", path: "tasks/followups/M1-02.F1.md"},
		{id: "M1-10", title: "Ten", status: "done", path: "tasks/M1/M1-10.md"},
	}
	if !reflect.DeepEqual(tasks, want) {
		t.Errorf("loadTasks() = %+v, want %+v", tasks, want)
	}
}
