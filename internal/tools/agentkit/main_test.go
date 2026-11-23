package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	cwd := t.TempDir()
	input := func(tool, field, value string) string {
		data, err := json.Marshal(map[string]any{
			"cwd":        cwd,
			"tool_name":  tool,
			"tool_input": map[string]string{field: value},
		})
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	tests := []struct {
		name       string
		args       []string
		stdin      string
		wantStatus int
		wantDenial string // a fragment of the reason in a deny decision; empty when nothing is denied
		wantStderr string
	}{
		{"guard allows", []string{"guard"}, input("Bash", "command", "git status"), exitOK, "", ""},
		{"guard refuses", []string{"guard"}, input("Bash", "command", "git push --force"), exitOK, "force-push", ""},
		{"guard reads powershell", []string{"guard"}, input("PowerShell", "command", "gh pr merge 3"), exitOK,
			"gh pr merge", ""},
		{"protect outside a checkout", []string{"protect"}, input("Write", "file_path", "AGENTS.md"), exitOK, "", ""},
		{"bad hook input fails open", []string{"guard"}, "not json", exitFailed, "", "reading the hook input"},
		{"unknown subcommand", []string{"nope"}, "", exitFailed, "", "usage"},
		{"no subcommand", nil, "", exitFailed, "", "usage"},
		{"only clean-worktrees takes flags", []string{"ready", "--dry-run"}, "", exitFailed, "", "usage"},
		{"unknown clean-worktrees flag", []string{"clean-worktrees", "--force"}, "", exitFailed, "", "unknown flag"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			status := run(test.args, strings.NewReader(test.stdin), &stdout, &stderr)
			if status != test.wantStatus || !strings.Contains(stderr.String(), test.wantStderr) {
				t.Errorf("run(%q) = %d, stderr %q; want %d, stderr containing %q",
					test.args, status, stderr.String(), test.wantStatus, test.wantStderr)
			}
			if test.wantDenial == "" {
				if stdout.Len() != 0 {
					t.Errorf("run(%q) printed %q, want nothing", test.args, stdout.String())
				}
				return
			}
			var output hookOutput
			if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
				t.Fatalf("run(%q) printed %q: %v", test.args, stdout.String(), err)
			}
			decision := output.HookSpecificOutput
			if decision.HookEventName != "PreToolUse" || decision.PermissionDecision != "deny" ||
				!strings.Contains(decision.PermissionDecisionReason, test.wantDenial) {
				t.Errorf("run(%q) decided %+v, want a PreToolUse deny containing %q",
					test.args, decision, test.wantDenial)
			}
		})
	}
}
