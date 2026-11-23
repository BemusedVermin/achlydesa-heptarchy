package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := t.TempDir()
	for _, args := range [][]string{{"init", "-q", "-b", "master"}} {
		if _, err := git(dir, args...); err != nil {
			t.Fatal(err)
		}
	}
	for name, content := range baseFiles {
		write(t, dir, name, content)
	}
	for _, args := range [][]string{
		{"add", "."},
		{"-c", "user.name=test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false",
			"commit", "-q", "-m", "start"},
	} {
		if _, err := git(dir, args...); err != nil {
			t.Fatal(err)
		}
	}
	top, err := git(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(top)
	return filepath.FromSlash(top)
}

func write(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		changes map[string]string // files to write in the working tree; "" deletes
		args    []string
		want    []string // fragments of the report
		same    bool
	}{
		{
			name:    "a refactor that changes no behavior",
			changes: map[string]string{"internal/tui/view.go": "package tui\n\n// View draws.\n"},
			want:    []string{"Same behavior as master: 3 harness rows", "and 1 golden files"},
			same:    true,
		},
		{
			name:    "a different step",
			changes: map[string]string{"play.csv": "seed,step,map\n1,1,map.clay_plain\n1,2,map.crawlspace\n"},
			want:    []string{`harness CSV: line 3 is "1,2,map.crawlspace", was "1,2,map.clay_plain"`},
		},
		{
			name:    "a different final hash",
			changes: map[string]string{"play.log": "seed=1 hash=00bb\n"},
			want:    []string{"final hashes: seed 1 ends at 00bb, was 00aa"},
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
			for name, content := range test.changes {
				if content == "" {
					if err := os.Remove(filepath.Join(root, filepath.FromSlash(name))); err != nil {
						t.Fatal(err)
					}
					continue
				}
				write(t, root, name, content)
			}
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
			if worktrees, _ := git(root, "worktree", "list"); strings.Count(worktrees, "\n") > 0 {
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
