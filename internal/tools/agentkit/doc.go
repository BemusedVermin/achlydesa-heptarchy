// Command agentkit holds the guardrails for coding agents in this repository.
//
// Claude Code runs it as hooks (see .claude/settings.json), and the skills in .claude/skills run
// it as a helper:
//
//	guard         before a Bash or PowerShell command: refuses commits to master, force pushes,
//	              branch deletion, merging pull requests and commands that throw away work. It
//	              only reads the command line and asks git which branch is checked out, so it
//	              stays fast; CI runs the full check on every pull request
//	protect       before a file edit: on task/ branches, refuses edits to the rules agents work
//	              under and to the checks themselves (the design docs in docs/ stay editable)
//	after-edit    after a file edit: formats an edited Go file with golangci-lint fmt, lints its
//	              package, and shows the agent the findings in that file. It never blocks, and
//	              does nothing when golangci-lint isn't installed
//	ready         lists the tasks an agent can start now, from the task files' front matter
//	clean-worktrees
//	              removes the worktrees in .claude/worktrees that hold nothing unsaved (their work
//	              is pushed or already in master), then the leftover worktree- and task/ branches
//	              whose work is in master; keeps locked or unsaved ones and says why. --dry-run
//	              only reports; --no-fetch skips fetching origin first
//
// The hooks are a seatbelt, not a sandbox: they catch the usual commands an agent writes by
// mistake, and Claude Code's own permission rules and worktree isolation sit alongside them.
// A hook refuses a tool call by printing a "deny" decision as JSON, and Claude Code shows the
// agent the reason. If a hook fails instead (it can't read its input, or doesn't build), Claude
// Code reports the error and the tool call goes ahead.
package main
