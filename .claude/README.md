# Running agents on tasks

This folder sets up Claude Code so that agents can take tasks from `tasks/` and turn them into pull requests, one at a time or several at once, on your machine or in the cloud. Agents go as far as opening a PR. You review it and merge it.

## What's here

| Path | What it does |
| --- | --- |
| `settings.json` | Permissions: which commands run without asking and which are denied. It also wires up the guardrail hooks. |
| `skills/do-task/` | `/do-task <ID>`: the whole workflow, from a fresh branch to an open PR. |
| `skills/next-tasks/` | `/next-tasks`: what can start now, and what can run side by side. |
| `skills/run-tasks/` | `/run-tasks [IDs or N]`: starts one implementer agent per task, in parallel. |
| `skills/file-followup/` | How agents file follow-ups (used by do-task). |
| `agents/task-implementer.md` | A Sonnet agent that does one task in its own worktree. |
| `agents/task-reviewer.md` | A read-only Opus agent that reviews a task's diff against its spec and `AGENTS.md`. |
| `rules/` | Detail on top of `AGENTS.md` that loads only when Claude reads or edits matching files: `core.md` (`internal/core`), `content.md` (`content/`, `internal/content`), `frontend.md` (`internal/frontend`) and `tasks.md` (`tasks/`). |
| `../internal/tools/agentkit/` | The hooks' code (Go, with tests), plus the `ready` helper. |

`settings.local.json` (your own approvals) and `worktrees/` are gitignored. Everything else here is committed, because cloud sessions and worktrees only see what's in git.

The rules are guidance; `AGENTS.md` stays the full rulebook, because other coding tools read it too. A session started in a worktree (`claude --worktree ...`) loads the rules. A `/run-tasks` agent may not: a subagent in its own worktree takes its instructions from the session that started it, so the path-scoped rules may not load there. That's why `/do-task` tells agents to read the rule files for the areas they touch.

## The guardrails

| Agents can't... | Enforced by |
| --- | --- |
| commit to master, push to master, or rebase it | guard hook |
| force-push, delete branches, merge PRs, or change repo settings, releases or secrets | guard hook, and deny rules in `settings.json` |
| throw away uncommitted work (`reset --hard`, `clean -f`, `checkout -- .`, `restore .`, `stash clear`) | guard hook |
| push a commit that `go tool task check` hasn't passed on | guard hook: it runs the check before the push if it hasn't passed on that commit yet |
| edit `AGENTS.md`, `CLAUDE.md`, `docs/`, `.claude/`, `.github/`, `Taskfile.yml`, `.golangci.yml` or the hooks on a `task/` branch | protect hook (agents file a follow-up instead) |
| touch your main checkout from a worktree | Claude Code's own worktree isolation |

Edits are accepted automatically, and the usual go, git and gh commands run without asking. Anything else asks you first.

These guardrails are a seatbelt, not a sandbox. They catch the commands an agent writes by mistake; a determined workaround could get past them. The one guard nothing in a session can get around is GitHub branch protection, so turn that on too (see below).

The hooks run through `go run`, so they need Go on the PATH, which you have anyway. If a hook can't run, Claude Code shows a hook error and lets the command through, so don't ignore hook errors.

## Day to day

All commands are PowerShell, run from the repo folder:

```powershell
cd C:\Users\liamm\Documents\achlydesa-heptarchy
```

### See what's ready

```powershell
go run ./internal/tools/agentkit ready
```

Or type `/next-tasks` in a Claude session, which also tells you which ready tasks can run side by side.

### One task, in your terminal

```powershell
claude --worktree M2-06 "/do-task M2-06"
```

`--worktree M2-06` gives the session its own copy of the repo under `.claude\worktrees\M2-06`, made fresh from GitHub's master, so it can't collide with anything else. The agent branches to `task/M2-06-...`, implements, runs the check, gets a review, and opens the PR. Answer it if it asks you something.

When you exit, Claude asks whether to keep the worktree. Once the PR is open, the work is on GitHub, so remove it.

### Several tasks at once

You can open one terminal per task, each with its own worktree name:

```powershell
claude --worktree M2-06 "/do-task M2-06"
claude --worktree M2-12 "/do-task M2-12"
```

Or let one session fan the work out:

```powershell
claude
```

Then type `/run-tasks M2-06 M2-12`, or `/run-tasks 3` for the first three ready tasks. Each task gets a `task-implementer` agent in its own worktree, running in the background, and the session reports each PR as it lands. It won't start two tasks that change the same files.

### In the cloud

At [claude.ai/code](https://claude.ai/code), pick this repository and send `/do-task M2-06`. Or send a task to the cloud from your terminal:

```powershell
claude --cloud "/do-task M2-06"
```

Do this once first: in the cloud environment's settings, add a setup script that installs golangci-lint, so the check can run:

```bash
#!/bin/bash
GOTOOLCHAIN=go1.27.4 GOBIN=/usr/local/bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 || true
```

The cloud caches what the script installs, so later sessions start with it. Keep the two versions in step with `go.mod` and `.github/workflows/ci.yml`.

### Review and merge

1. Read the PR on GitHub. CI runs `go tool task check` again.
2. For a second opinion, ask in any Claude session: `Use the task-reviewer agent on PR #45 for task M2-06.`
3. To ask for changes, start a session and describe them:

   ```powershell
   claude --worktree pr45 "Update PR #45: <what to change>. Follow 'Updating an open pull request' in the do-task skill."
   ```

4. Merge on GitHub yourself.

When parallel PRs land, the later ones often conflict on `tasks/INDEX.md`. Fix each the same way: `Update PR #46 from master.`

### Clean up worktrees

From your main checkout, with no Claude sessions running in the worktrees:

```powershell
go tool task clean-worktrees -- --dry-run   # see what it would do
go tool task clean-worktrees
```

It fetches origin, then removes each worktree in `.claude\worktrees` whose work is already in master (squash merges included) or pushed. Then it deletes the leftover `worktree-*` and `task/*` branches whose work is all in master. It keeps any worktree that is locked (a session or agent may still be using it), has uncommitted or untracked files, or has commits that are neither pushed nor merged, and says why for each. Files in `.claude\worktrees` that aren't worktrees are listed and left alone.

Claude Code also removes subagent worktrees by itself once they're old and hold no unpushed work.

## Protect master on GitHub

This takes a minute, and it's the backstop for everything above. On GitHub, go to Settings, then Rules, then Rulesets, and add a branch ruleset for `master` that:

- requires a pull request before merging
- requires the `check` status check (the CI job) to pass
- blocks force pushes and deletion

After that, even a confused agent, or a hook that didn't run, can't change master.

## When a guardrail gets in your way

- The guardrails are for agents. You can always run the command yourself in your own terminal.
- To change a rule, edit `settings.json` or `internal/tools/agentkit` on a branch that isn't a `task/` branch, and keep the hook tests passing.
- To see what the guard would say about a command, pipe it a hook input:

  ```powershell
  '{"cwd":"C:/Users/liamm/Documents/achlydesa-heptarchy","tool_name":"PowerShell","tool_input":{"command":"git push --force"}}' | go run ./internal/tools/agentkit guard
  ```

## Where to read more

- [Subagents](https://code.claude.com/docs/en/sub-agents), [skills](https://code.claude.com/docs/en/skills) and [hooks](https://code.claude.com/docs/en/hooks)
- [Worktrees](https://code.claude.com/docs/en/worktrees) and [permissions](https://code.claude.com/docs/en/permissions)
- [Cloud environments](https://code.claude.com/docs/en/claude-code-on-the-web)
