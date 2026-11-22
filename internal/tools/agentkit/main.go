package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
)

// Exit statuses. A hook refuses a tool call by printing a decision on standard output and exiting
// with status 0, not by exiting with status 2: the hooks run through "go run", which turns any
// non-zero status from the program into status 1. Claude Code reports a status 1 to the user as a
// hook error and lets the tool call go ahead.
const (
	exitOK     = 0
	exitFailed = 1
)

// errUsage reports a command line agentkit doesn't understand.
var errUsage = errors.New("usage: agentkit guard | protect | mark-checked | ready")

// hookInput is the part of Claude Code's PreToolUse hook input that the hooks read.
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
	if len(args) != 1 {
		warn(stderr, errUsage)
		return exitFailed
	}
	switch args[0] {
	case "guard":
		return hook(stdin, stdout, stderr, guardHook)
	case "protect":
		return hook(stdin, stdout, stderr, protectHook)
	case "mark-checked":
		dir, err := os.Getwd()
		if err == nil {
			err = markChecked(dir)
		}
		if err != nil {
			// The check itself passed; failing to record it only means the next push runs it again.
			warn(stderr, "agentkit mark-checked: couldn't record the passing commit:", err)
		}
		return exitOK
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
	default:
		warn(stderr, errUsage)
		return exitFailed
	}
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

// guardHook checks a Bash or PowerShell command line, and runs the push gate for each push in it.
func guardHook(input hookInput) string {
	dia := bash
	if input.ToolName == "PowerShell" {
		dia = powershell
	}
	result := evaluate(input.ToolInput.Command, dia, input.Cwd, gitRepo{})
	if result.refusal != "" {
		return result.refusal
	}
	var gated []string
	for _, dir := range result.pushes {
		if slices.Contains(gated, dir) {
			continue
		}
		gated = append(gated, dir)
		if refusal := gatePush(dir, runTaskCheck); refusal != "" {
			return refusal
		}
	}
	return ""
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
