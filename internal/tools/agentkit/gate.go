package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// markerName is the file, inside a checkout's git directory, that holds the commit the check last
// passed on. Each worktree has its own git directory, so each has its own marker.
const markerName = "agentkit-checked"

// gitRepo answers questions about git checkouts by running git.
type gitRepo struct{}

// branch implements branchReader.
func (gitRepo) branch(dir string) string {
	name, err := gitOutput(dir, "symbolic-ref", "--short", "-q", "HEAD")
	if err != nil {
		return ""
	}
	return name
}

// gitOutput runs git in dir and returns its standard output without the trailing new line.
func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimRight(string(out), "\r\n"), nil
}

// checkout is a git checkout that the gate can verify.
type checkout struct {
	top    string // the checkout's root directory
	head   string // the commit HEAD points at
	marker string // the marker file's path
}

// openCheckout finds the checkout dir is in. It fails when dir isn't in a git checkout of this
// repository, which it recognizes by a Taskfile.yml at the root.
func openCheckout(dir string) (checkout, error) {
	top, err := gitOutput(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return checkout{}, err
	}
	top = filepath.FromSlash(top)
	if _, err := os.Stat(filepath.Join(top, "Taskfile.yml")); err != nil {
		return checkout{}, fmt.Errorf("%s has no Taskfile.yml: %w", top, err)
	}
	head, err := gitOutput(top, "rev-parse", "HEAD")
	if err != nil {
		return checkout{}, err
	}
	marker, err := gitOutput(top, "rev-parse", "--git-path", markerName)
	if err != nil {
		return checkout{}, err
	}
	marker = filepath.FromSlash(marker)
	if !filepath.IsAbs(marker) {
		marker = filepath.Join(top, marker)
	}
	return checkout{top: top, head: head, marker: marker}, nil
}

// uncommitted returns git's short status of the checkout's uncommitted changes, untracked files
// included; empty when it is clean.
func (repo checkout) uncommitted() (string, error) {
	return gitOutput(repo.top, "status", "--porcelain")
}

// checked reports whether the check has passed on the checkout's HEAD.
func (repo checkout) checked() bool {
	data, err := os.ReadFile(repo.marker)
	return err == nil && strings.TrimSpace(string(data)) == repo.head
}

// mark records that the check passed on the checkout's HEAD.
func (repo checkout) mark() error {
	return os.WriteFile(repo.marker, []byte(repo.head+"\n"), 0o644)
}

// checkRunner runs the full check in a checkout's root, returning its output.
type checkRunner func(top string) (string, error)

// runTaskCheck runs "go tool task check".
func runTaskCheck(top string) (string, error) {
	cmd := exec.Command("go", "tool", "task", "check")
	cmd.Dir = top
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// gatePush decides whether a push from dir may go ahead: the check must have passed on exactly the
// commit being pushed, so the tree has to be clean. When the check hasn't passed on HEAD yet, the
// gate runs it now. It returns why the push must wait, or "" when it may go.
func gatePush(dir string, run checkRunner) string {
	repo, err := openCheckout(dir)
	if err != nil {
		return "" // not a checkout of this repository; nothing to gate
	}
	status, err := repo.uncommitted()
	if err != nil {
		return "couldn't read git status before pushing: " + err.Error()
	}
	if status != "" {
		return "commit your changes before pushing: the check has to pass on exactly what you push. " +
			"Uncommitted in " + repo.top + ":\n" + firstLines(status, 15)
	}
	if repo.checked() {
		return ""
	}
	out, err := run(repo.top)
	if err != nil {
		return "go tool task check failed on " + short(repo.head) + ", so the push waits. " +
			"Fix it, commit, and push again. The end of the check's output:\n" + lastLines(out, 60)
	}
	if err := repo.mark(); err != nil {
		return "the check passed but couldn't be recorded: " + err.Error()
	}
	return ""
}

// markChecked records that the check passed on the HEAD of the checkout in dir, if the tree is
// clean. "task check" runs it as its last step, so a push right after a passing check doesn't
// run it again.
func markChecked(dir string) error {
	repo, err := openCheckout(dir)
	if err != nil {
		// Outside a git checkout, such as in a source archive, there is nothing to record.
		return nil
	}
	status, err := repo.uncommitted()
	if err != nil {
		return err
	}
	if status != "" {
		return nil
	}
	return repo.mark()
}

// short returns the first 12 characters of a commit hash.
func short(hash string) string {
	if len(hash) > 12 {
		return hash[:12]
	}
	return hash
}

// firstLines returns at most count lines from the start of text.
func firstLines(text string, count int) string {
	lines := strings.Split(text, "\n")
	if len(lines) > count {
		lines = append(lines[:count], "...")
	}
	return strings.Join(lines, "\n")
}

// lastLines returns at most count lines from the end of text.
func lastLines(text string, count int) string {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	if len(lines) > count {
		lines = append([]string{"..."}, lines[len(lines)-count:]...)
	}
	return strings.Join(lines, "\n")
}
