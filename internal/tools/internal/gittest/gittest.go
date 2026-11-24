package gittest

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/tools/internal/gitcmd"
)

// Repo makes a repository in a temporary directory whose branch holds files (slash-separated paths
// to contents) in one commit, and returns its root as git reports it. It skips the test when git
// isn't installed.
//
// The repository has its own identity, so later commits work without the user's configuration,
// and core.autocrlf off, so a Windows checkout has the same line endings as the files written.
func Repo(t *testing.T, branch string, files map[string]string) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := t.TempDir()
	tree := fstest.MapFS{}
	for name, content := range files {
		tree[name] = &fstest.MapFile{Data: []byte(content)}
	}
	if err := os.CopyFS(dir, tree); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"init", "-q", "-b", branch},
		{"config", "user.name", "test"},
		{"config", "user.email", "test@example.com"},
		{"config", "commit.gpgsign", "false"},
		{"config", "core.autocrlf", "false"},
		{"add", "."},
		{"commit", "-q", "--allow-empty", "-m", "start"},
	} {
		Git(t, dir, args...)
	}
	return filepath.FromSlash(Git(t, dir, "rev-parse", "--show-toplevel"))
}

// Git runs git in dir and returns its output, failing the test if git fails.
func Git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := gitcmd.Output(dir, args...)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Write writes each file (a slash-separated path under root) with its contents, creating folders
// and replacing what is there. An empty content removes the file.
func Write(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if content == "" {
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
