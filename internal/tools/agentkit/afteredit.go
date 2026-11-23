package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// maxFindings caps how many lint findings the after-edit hook passes on, so one broken edit can't
// flood the agent's context.
const maxFindings = 20

// goTools formats and lints Go code. The after-edit hook runs golangci-lint through it; tests swap in
// a fake.
type goTools interface {
	// available reports whether the tools can run at all.
	available() bool
	// format formats one file in place. root is the checkout's root and file is relative to it,
	// with forward slashes.
	format(root, file string) error
	// lint lints one package directory, relative to root with forward slashes, and returns the
	// findings, one per line, each starting with the file's path relative to root.
	lint(root, dir string) (string, error)
}

// golangciLint runs golangci-lint with the checkout's own .golangci.yml.
type golangciLint struct{}

// available implements goTools.
func (golangciLint) available() bool {
	_, err := exec.LookPath("golangci-lint")
	return err == nil
}

// format implements goTools.
func (golangciLint) format(root, file string) error {
	cmd := exec.Command("golangci-lint", "fmt", file)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(strings.TrimSpace(string(out)))
	}
	return nil
}

// lint implements goTools. Several agents may lint at once, so it doesn't wait for, or refuse to
// run beside, another golangci-lint.
func (golangciLint) lint(root, dir string) (string, error) {
	pattern := "./" + strings.TrimPrefix(dir+"/", "./")
	cmd := exec.Command("golangci-lint", "run", "--allow-parallel-runners", "--show-stats=false",
		"--output.text.print-issued-lines=false", "--output.text.colors=false", pattern)
	cmd.Dir = root
	out, err := cmd.Output()
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 {
		// Status 1 means it found issues, which are on standard output.
		err = nil
	}
	return string(out), err
}

// postToolUseOutput is a PostToolUse hook's answer, in the JSON shape Claude Code reads from
// standard output. Claude Code shows additionalContext to the agent next to the tool's result.
type postToolUseOutput struct {
	HookSpecificOutput struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext"`
	} `json:"hookSpecificOutput"`
}

// runAfterEdit reads a PostToolUse hook's input, formats and lints the edited Go file, and passes
// any findings in that file on to the agent. It never blocks anything: when it can't do its job, it
// stays quiet and the check catches the problem later.
func runAfterEdit(stdin io.Reader, stdout, stderr io.Writer, tools goTools) int {
	var input hookInput
	if err := json.NewDecoder(stdin).Decode(&input); err != nil {
		warn(stderr, "agentkit after-edit: reading the hook input:", err)
		return exitFailed
	}
	note := afterEdit(input, tools)
	if note == "" {
		return exitOK
	}
	var output postToolUseOutput
	output.HookSpecificOutput.HookEventName = "PostToolUse"
	output.HookSpecificOutput.AdditionalContext = note
	if err := json.NewEncoder(stdout).Encode(output); err != nil {
		warn(stderr, "agentkit after-edit: writing the hook output:", err)
		return exitFailed
	}
	return exitOK
}

// afterEdit formats the Go file an edit touched and lints its package. It returns what the agent
// should hear: the lint findings in that file, or nothing.
func afterEdit(input hookInput, tools goTools) string {
	file := input.ToolInput.FilePath
	if !strings.HasSuffix(file, ".go") || !tools.available() {
		return ""
	}
	if !filepath.IsAbs(file) {
		file = filepath.Join(input.Cwd, file)
	}
	if _, err := os.Stat(file); err != nil {
		return ""
	}
	top, err := gitOutput(filepath.Dir(file), "rev-parse", "--show-toplevel")
	if err != nil {
		return ""
	}
	root := filepath.FromSlash(top)
	rel, ok := relativeTo(root, file)
	if !ok || rel == "." {
		return ""
	}
	// A file mid-change may not parse yet; the lint below reports that, so a failed format is
	// left to it.
	_ = tools.format(root, rel)
	out, err := tools.lint(root, path.Dir(rel))
	if err != nil {
		return ""
	}
	findings := findingsIn(out, rel)
	if len(findings) == 0 {
		return ""
	}
	return "golangci-lint found these in " + rel + " after your edit. Fix them before you finish; " +
		"if one comes from a change you're still in the middle of, carry on and check again after:\n" +
		strings.Join(findings, "\n")
}

// findingsIn picks the lines of golangci-lint's output that are about file, at most maxFindings of
// them. golangci-lint prints paths relative to the root it ran in, with backslashes on Windows.
func findingsIn(output, file string) []string {
	var found []string
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimRight(line, "\r")
		if !strings.HasPrefix(strings.ReplaceAll(line, `\`, "/"), file+":") {
			continue
		}
		if len(found) == maxFindings {
			found = append(found, "(more not shown)")
			break
		}
		found = append(found, line)
	}
	return found
}
