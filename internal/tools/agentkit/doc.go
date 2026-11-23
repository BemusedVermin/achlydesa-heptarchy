// Command agentkit holds the guardrails for coding agents in this repository.
//
// Claude Code runs it as hooks (see .claude/settings.json), and the skills in .claude/skills run
// it as a helper:
//
//	guard         before a Bash or PowerShell command: refuses commits to master, force pushes,
//	              branch deletion, merging pull requests and commands that throw away work, and
//	              makes every git push wait until "go tool task check" has passed on the commit
//	protect       before a file edit: on task/ branches, refuses edits to the rules, the design
//	              docs and the checks themselves
//	mark-checked  run by "task check" after it passes: records the commit it passed on, so the
//	              push that follows doesn't run the check again
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
