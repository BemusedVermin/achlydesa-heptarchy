// Command agentkit holds the guardrails and helpers for coding agents in this repository. Claude
// Code runs it as hooks (see .claude/settings.json), and the skills run it as a helper.
//
//	guard            refuses commits to master, force pushes, branch deletion, merging pull
//	                 requests and commands that throw away work
//	protect          on task/ branches, refuses edits to the rules agents work under
//	after-edit       formats and lints an edited Go file and flags filler in its new comments
//	ready            lists the tasks an agent can start now
//	check-pr         checks a branch's task bookkeeping and the comments it changed; CI runs it
//	clean-worktrees  removes agent worktrees and branches whose work is pushed or merged
//	comments         lists the comments under the given paths that break .claude/rules/comments.md
//
// The hooks catch the usual mistakes, not every one. A hook that fails, rather than refusing, lets
// the tool call go ahead.
package main
