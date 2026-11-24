package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/tools/internal/gittest"
)

// filePlayer stands in for the harness: its CSV and log are the checkout's play.csv and play.log.
type filePlayer struct{}

func (filePlayer) play(root string, _, _ int) ([]byte, []byte, error) {
	csv, err := os.ReadFile(filepath.Join(root, "play.csv"))
	if err != nil {
		return nil, nil, err
	}
	log, err := os.ReadFile(filepath.Join(root, "play.log"))
	return csv, log, err
}

// baseFiles are the files committed on master in every test repository.
var baseFiles = map[string]string{
	"play.csv": "seed,step,map\n1,1,map.clay_plain\n1,2,map.clay_plain\n",
	"play.log": "level=INFO msg=\"seed finished\" seed=1 hash=00aa\n",
	"internal/tui/testdata/clay_plain.golden": "the clay plain\n",
	"internal/tui/view.go":                    "package tui\n",
}

// newRepo makes a repository whose master holds baseFiles, makes cwd its root, and returns the root.
func newRepo(t *testing.T) string {
	t.Helper()
	root := gittest.Repo(t, "master", baseFiles)
	t.Chdir(root)
	return root
}

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		changes map[string]string
		args    []string
		want    []string
		same    bool
	}{
		{
			name:    "a refactor that changes no behavior",
			changes: map[string]string{"internal/tui/view.go": "package tui\n\n// View draws.\n"},
			want:    []string{"Same behavior as master: 3 harness rows", "and 1 golden files"},
			same:    true,
		},
		{
			name:    "a regenerated golden file",
			changes: map[string]string{"internal/tui/testdata/clay_plain.golden": "the wet clay plain\n"},
			want:    []string{"internal/tui/testdata/clay_plain.golden: line 1"},
		},
		{
			name: "a rename, declared",
			changes: map[string]string{
				"play.csv": "seed,step,map\n1,1,map.clay_flat\n1,2,map.clay_flat\n",
				"play.log": "seed=1 hash=00cc\n",
				"internal/tui/testdata/clay_plain.golden": "",
				"internal/tui/testdata/clay_flat.golden":  "the clay plain\n",
			},
			args: []string{"--rename", "clay_plain=clay_flat"},
			want: []string{"Same behavior as master"},
			same: true,
		},
		{
			name: "a rename, undeclared",
			changes: map[string]string{
				"internal/tui/testdata/clay_plain.golden": "",
				"internal/tui/testdata/clay_flat.golden":  "the clay plain\n",
			},
			want: []string{
				"internal/tui/testdata/clay_flat.golden: new golden file",
				"internal/tui/testdata/clay_plain.golden: golden file removed",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := newRepo(t)
			gittest.Write(t, root, test.changes)
			var stdout strings.Builder
			err := run(append([]string{"--base", "master"}, test.args...), &stdout, filePlayer{})
			if test.same && err != nil {
				t.Fatalf("run: %v\n%s", err, stdout.String())
			}
			if !test.same && !errors.Is(err, errDifferent) {
				t.Fatalf("run = %v, want errDifferent\n%s", err, stdout.String())
			}
			for _, fragment := range test.want {
				if !strings.Contains(stdout.String(), fragment) {
					t.Errorf("report is missing %q:\n%s", fragment, stdout.String())
				}
			}
			if worktrees := gittest.Git(t, root, "worktree", "list"); strings.Count(worktrees, "\n") > 0 {
				t.Errorf("the base worktree wasn't removed:\n%s", worktrees)
			}
		})
	}
}

func TestParseOptions(t *testing.T) {
	opts, err := parseOptions([]string{"--seeds", "5", "--rename", "a=b", "--rename", "c=d"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.base != "origin/master" || opts.seeds != 5 || opts.steps != 200 ||
		len(opts.renames) != 2 || opts.renames[1] != (rename{"c", "d"}) {
		t.Errorf("parseOptions = %+v", opts)
	}
	for _, bad := range [][]string{{"--rename", "nothing"}, {"--rename", "=b"}, {"extra"}} {
		if _, err := parseOptions(bad); err == nil {
			t.Errorf("parseOptions(%q) succeeded, want an error", bad)
		}
	}
}

func TestApplyRenamesInOrder(t *testing.T) {
	got := applyRenames("form.cut form.cutter", []rename{{"form.cutter", "form.blade"}, {"form.cut", "form.slice"}})
	if got != "form.slice form.blade" {
		t.Errorf("applyRenames = %q", got)
	}
}

func TestFirstDifference(t *testing.T) {
	tests := []struct {
		name      string
		want, got string
		report    string
	}{
		{"the same", "a\nb\n", "a\nb\n", ""},
		{"a different step", "a\nb\n", "a\nc\n", `line 2 is "c", was "b"`},
		{"a line added", "a\n", "a\nb\n", `line 2 is "b", was ""`},
		{"a line removed", "a\nb", "a", `line 2 is "", was "b"`},
		{"CRLF against LF", "a\nb\n", "a\r\nb\r\n", `line 1 is "a\r", was "a"`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := firstDifference(test.want, test.got); got != test.report {
				t.Errorf("firstDifference(%q, %q) = %q, want %q", test.want, test.got, got, test.report)
			}
		})
	}
}

func TestCompareHashes(t *testing.T) {
	tests := []struct {
		name      string
		want, got string
		report    string
	}{
		{"the same", "seed=1 hash=00aa\nseed=2 hash=00bb\n", "seed=2 hash=00bb\nseed=1 hash=00aa\n", ""},
		{"a different final hash", "seed=1 hash=00aa\n", "seed=1 hash=00bb\n", "seed 1 ends at 00bb, was 00aa"},
		{"a seed missing", "seed=1 hash=00aa\nseed=2 hash=00bb\n", "seed=1 hash=00aa\n", "seed 2 ends at , was 00bb"},
		{"a seed added", "seed=1 hash=00aa\n", "seed=1 hash=00aa\nseed=2 hash=00bb\n", "2 seeds reported a hash, was 1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := compareHashes(seedHashes([]byte(test.want)), seedHashes([]byte(test.got)))
			if got != test.report {
				t.Errorf("compareHashes = %q, want %q", got, test.report)
			}
		})
	}
}
