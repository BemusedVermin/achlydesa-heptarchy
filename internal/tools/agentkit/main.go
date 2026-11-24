package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
)

// Exit statuses. A hook refuses by printing a decision and exiting 0, not 2: "go run" turns any
// non-zero status into 1, which Claude Code reports as a hook error before letting the call go ahead.
const (
	exitOK     = 0
	exitFailed = 1
)

// errUsage reports a command line agentkit doesn't understand.
var errUsage = errors.New(
	"usage: agentkit guard | protect | after-edit | ready | check-pr [--branch B] [--base origin/master] | " +
		"clean-worktrees [--dry-run] [--no-fetch] | comments [path...]")

// hookInput is the part of Claude Code's PreToolUse and PostToolUse hook input that the hooks read.
type hookInput struct {
	Cwd       string `json:"cwd"`
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		Command      string `json:"command"`
		FilePath     string `json:"file_path"`
		NotebookPath string `json:"notebook_path"`
	} `json:"tool_input"`
}

// main runs the subcommand named by the first argument.
func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// run runs a subcommand and returns the exit status.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	takesArgs := len(args) > 0 && slices.Contains([]string{"clean-worktrees", "check-pr", "comments"}, args[0])
	if len(args) == 0 || len(args) > 1 && !takesArgs {
		warn(stderr, errUsage)
		return exitFailed
	}
	switch args[0] {
	case "guard":
		return hook(stdin, stdout, stderr, guardHook)
	case "protect":
		return hook(stdin, stdout, stderr, protectHook)
	case "after-edit":
		return runAfterEdit(stdin, stdout, stderr, golangciLint{})
	case "ready":
		dir, err := os.Getwd()
		report := ""
		if err == nil {
			report, err = runReady(dir)
		}
		if err != nil {
			warn(stderr, "agentkit ready:", err)
			return exitFailed
		}
		if _, err := io.WriteString(stdout, report); err != nil {
			warn(stderr, "agentkit ready:", err)
			return exitFailed
		}
		return exitOK
	case "check-pr":
		return runCheckPR(args[1:], stdout, stderr)
	case "clean-worktrees":
		return runCleanWorktrees(args[1:], stdout, stderr)
	case "comments":
		return runComments(args[1:], stdout, stderr)
	default:
		warn(stderr, errUsage)
		return exitFailed
	}
}

// runCleanWorktrees runs clean-worktrees from the current directory and prints its report. It fails
// when it can't run at all, or when git refused a removal or deletion it meant to make.
func runCleanWorktrees(flags []string, stdout, stderr io.Writer) int {
	opts, err := parseCleanFlags(flags)
	if err != nil {
		warn(stderr, "agentkit clean-worktrees:", err)
		return exitFailed
	}
	dir, err := os.Getwd()
	if err != nil {
		warn(stderr, "agentkit clean-worktrees:", err)
		return exitFailed
	}
	result, err := cleanWorktrees(dir, opts)
	if err != nil {
		warn(stderr, "agentkit clean-worktrees:", err)
		return exitFailed
	}
	if _, err := io.WriteString(stdout, result.report()); err != nil {
		warn(stderr, "agentkit clean-worktrees:", err)
		return exitFailed
	}
	if len(result.failures) > 0 {
		return exitFailed
	}
	return exitOK
}

// warn writes a diagnostic line to standard error. If that write fails there is nowhere left to
// report it, so the error is dropped on purpose.
func warn(stderr io.Writer, parts ...any) {
	_, _ = fmt.Fprintln(stderr, parts...)
}

// hookOutput is a PreToolUse hook's decision, in the JSON shape Claude Code reads from standard
// output.
type hookOutput struct {
	HookSpecificOutput struct {
		HookEventName            string `json:"hookEventName"`
		PermissionDecision       string `json:"permissionDecision"`
		PermissionDecisionReason string `json:"permissionDecisionReason"`
	} `json:"hookSpecificOutput"`
}

// hook reads a PreToolUse hook's input and runs decide on it. A non-empty answer from decide
// refuses the tool call: Claude Code blocks it and shows the agent the answer as the reason.
func hook(stdin io.Reader, stdout, stderr io.Writer, decide func(hookInput) string) int {
	var input hookInput
	if err := json.NewDecoder(stdin).Decode(&input); err != nil {
		warn(stderr, "agentkit: reading the hook input:", err)
		return exitFailed
	}
	refusal := decide(input)
	if refusal == "" {
		return exitOK
	}
	var output hookOutput
	output.HookSpecificOutput.HookEventName = "PreToolUse"
	output.HookSpecificOutput.PermissionDecision = "deny"
	output.HookSpecificOutput.PermissionDecisionReason = "Refused by the agent guardrails (internal/tools/agentkit): " +
		refusal
	if err := json.NewEncoder(stdout).Encode(output); err != nil {
		warn(stderr, "agentkit: writing the hook decision:", err)
		return exitFailed
	}
	return exitOK
}

// guardHook checks a Bash or PowerShell command line.
func guardHook(input hookInput) string {
	dia := bash
	if input.ToolName == "PowerShell" {
		dia = powershell
	}
	return evaluate(input.ToolInput.Command, dia, input.Cwd, gitRepo{}).refusal
}

// protectHook checks a file edit.
func protectHook(input hookInput) string {
	path := input.ToolInput.FilePath
	if path == "" {
		path = input.ToolInput.NotebookPath
	}
	if path != "" {
		path = resolve(input.Cwd, path)
	}
	return checkEdit(path, gitRepo{})
}
