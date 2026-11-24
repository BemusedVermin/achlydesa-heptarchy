package main

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

const sampleLog = "task: [test-affected] go run ./internal/tools/affected --base origin/master\n" +
	"ok  \texample.com/game/internal/core/save\t9.341s\n" +
	"ok  \texample.com/game/internal/harness\t72.300s\tcoverage: 81.2% of statements\n" +
	"ok  \texample.com/game/internal/core/rng\t(cached)\n" +
	"FAIL\texample.com/game/internal/core/combat\t1.500s\n" +
	"?   \texample.com/game/cmd/tool\t[no test files]\n" +
	"--- FAIL: TestSomething (0.00s)\n"

func TestParseLog(t *testing.T) {
	got, err := parseLog(strings.NewReader(sampleLog))
	if err != nil {
		t.Fatal(err)
	}
	want := times{
		"example.com/game/internal/core/save":   9341 * time.Millisecond,
		"example.com/game/internal/harness":     72300 * time.Millisecond,
		"example.com/game/internal/core/combat": 1500 * time.Millisecond,
	}
	if !maps.Equal(got, want) {
		t.Errorf("parseLog = %v, want %v", got, want)
	}
}

func TestSaveAndLoadTimes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "times.txt")
	saved := times{"b": 1500 * time.Millisecond, "a": 9341 * time.Millisecond}
	if err := os.WriteFile(path, []byte(saved.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := saved.String(); got != "a 9.341\nb 1.500\n" {
		t.Errorf("String = %q, want sorted by package", got)
	}
	loaded, err := loadTimes(path)
	if err != nil {
		t.Fatal(err)
	}
	if !maps.Equal(loaded, saved) {
		t.Errorf("loadTimes = %v, want %v", loaded, saved)
	}
	missing, err := loadTimes(filepath.Join(t.TempDir(), "none.txt"))
	if err != nil || len(missing) != 0 {
		t.Errorf("loadTimes of a missing file = %v, %v; want empty and no error", missing, err)
	}
}

func TestCompare(t *testing.T) {
	current := times{
		"harness": 72 * time.Second,
		"save":    20 * time.Second,
		"rng":     300 * time.Millisecond,
		"new":     2 * time.Second,
	}
	baseline := times{
		"harness": 5 * time.Second,
		"save":    9 * time.Second,
		"rng":     100 * time.Millisecond,
	}
	rows := compare(current, baseline, 30*time.Second)
	want := []row{
		{pkg: "harness", took: 72 * time.Second, baseline: 5 * time.Second, overBudget: true, slower: true},
		{pkg: "save", took: 20 * time.Second, baseline: 9 * time.Second, slower: true},
		{pkg: "new", took: 2 * time.Second},
		{pkg: "rng", took: 300 * time.Millisecond, baseline: 100 * time.Millisecond},
	}
	if !slices.Equal(rows, want) {
		t.Errorf("compare = %+v, want %+v", rows, want)
	}
}

func TestReport(t *testing.T) {
	rows := compare(
		times{"example.com/game/internal/harness": 72 * time.Second, "example.com/game/internal/rng": time.Second},
		times{"example.com/game/internal/harness": 5 * time.Second},
		30*time.Second,
	)
	got := annotations(rows, 30*time.Second)
	want := "::warning title=Slow tests::example.com/game/internal/harness took 72.0s, over the 30.0s budget\n" +
		"::warning title=Tests got slower::example.com/game/internal/harness took 72.0s, up from 5.0s on master\n"
	if got != want {
		t.Errorf("annotations =\n%s\nwant\n%s", got, want)
	}
	table := summary(rows, 30*time.Second, true, "example.com/game")
	for _, part := range []string{
		"Compared with the last run on master.",
		"| internal/harness | 72.0s | 5.0s | +67.0s | ⚠️ over budget, ⚠️ much slower |",
		"| internal/rng | 1.0s | – | – |  |",
	} {
		if !strings.Contains(table, part) {
			t.Errorf("summary is missing %q:\n%s", part, table)
		}
	}
	if empty := summary(nil, 30*time.Second, false, ""); !strings.Contains(empty, "No package reported") {
		t.Errorf("summary with no rows = %q", empty)
	}
}

func TestRun(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "check.log")
	baselinePath := filepath.Join(dir, "master.txt")
	savePath := filepath.Join(dir, "times.txt")
	summaryPath := filepath.Join(dir, "summary.md")
	if err := os.WriteFile(logPath, []byte(sampleLog), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(baselinePath, []byte("example.com/game/internal/harness 5.300\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(summaryPath, []byte("earlier step\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout strings.Builder
	args := []string{"--baseline", baselinePath, "--save", savePath, "--summary", summaryPath, logPath}
	if err := run(args, &stdout); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "title=Tests got slower::example.com/game/internal/harness") {
		t.Errorf("stdout = %q, want a warning for the harness", stdout.String())
	}
	saved, err := os.ReadFile(savePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(saved), "example.com/game/internal/core/save 9.341\n") {
		t.Errorf("saved times = %q", saved)
	}
	summaryText, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(summaryText), "earlier step\n### Test times") {
		t.Errorf("summary wasn't appended:\n%s", summaryText)
	}
	if err := run([]string{}, &stdout); err == nil {
		t.Error("run with no log succeeded, want an error")
	}
}
