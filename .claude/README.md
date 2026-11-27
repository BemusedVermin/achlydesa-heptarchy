# Running agents on tasks

This folder sets up Claude Code so that agents can take tasks from `tasks/` and turn them into pull requests, one at a time or several at once, on your machine or in the cloud. Agents go as far as opening a PR. You review it and merge it.

## What's here

| Path | What it does |
| --- | --- |
| `settings.json` | Permissions: which commands run without asking and which are denied. It also wires up the guardrail hooks. |
| `skills/do-task/` | `/do-task <ID>`: the whole workflow, from a fresh branch to an open PR. |
| `skills/next-tasks/` | `/next-tasks`: what can start now, and what can run side by side. |
| `skills/run-tasks/` | `/run-tasks [IDs or N]`: starts one implementer agent per task, in parallel. |
| `skills/decide/` | `/decide`: asks you the open design questions as multiple choice, then records your answers in the task files and the design doc, in one PR. |
| `skills/plan-milestone/` | `/plan-milestone [M3]`: turns a milestone's outline brief into task files. It asks you the brief's open questions, shows you the breakdown before writing specs, and wires the follow-ups aimed at the milestone to its real tasks. |
| `skills/triage/` | `/triage`: marks the proposed follow-ups that no longer apply obsolete, asks you about the rest as multiple choice, and records it all in one PR. |
| `skills/shared/` | Guidance the skills share, such as how to ask you questions. Not a skill itself. |
| `skills/audit/` | `/audit`: between milestones, read-only auditors look for comment bloat, duplication, low-value tests, unclear names, library opportunities and convention slips, and file the cleanup as a `C<n>` milestone. `/audit lock-in` turns the decisions into lint rules. Its checklists are in `lenses/`. |
| `skills/optimize/` | `/optimize [step] [procgen] [harness]`: runs the benchmarks with the profiler, one target at a time, has a `perf-auditor` per target rank the bottlenecks, and files the fixes as a `P<n>` speed milestone of proposed tasks. |
| `skills/file-followup/` | How agents file follow-ups (used by do-task). |
| `agents/task-implementer.md` | An agent that does one task in its own worktree: Sonnet, or Opus for `size: L` and `kind: seam` tasks when `/run-tasks` starts it. |
| `agents/code-auditor.md` | A read-only Opus agent that audits one area against the lens checklists. Started by `/audit`. |
| `agents/task-reviewer.md` | A read-only Opus agent that reviews a task's diff against its spec and `AGENTS.md`. |
| `agents/perf-auditor.md` | A read-only Opus agent that reads one target's benchmark results and profiles and ranks its bottlenecks. Started by `/optimize`. |
| `rules/` | Detail on top of `AGENTS.md` that loads only when Claude reads or edits matching files: `core.md` (`internal/core`), `content.md` (`content/`, `internal/content`), `frontend.md` (`internal/frontend`) and `tasks.md` (`tasks/`). |
| `../internal/tools/agentkit/` | The hooks' code (Go, with tests), plus the `ready` and `clean-worktrees` helpers. |

`settings.local.json` (your own approvals) and `worktrees/` are gitignored. Everything else here is committed, because cloud sessions and worktrees only see what's in git.

The rules are guidance; `AGENTS.md` stays the full rulebook, because other coding tools read it too. A session started in a worktree (`claude --worktree ...`) loads the rules. A `/run-tasks` agent may not: a subagent in its own worktree takes its instructions from the session that started it, so the path-scoped rules may not load there. That's why `/do-task` tells agents to read the rule files for the areas they touch.

## The guardrails

| Agents can't... | Enforced by |
| --- | --- |
| commit to master, push to master, or rebase it | guard hook |
| force-push, delete branches, merge PRs, or change repo settings, releases or secrets | guard hook, and deny rules in `settings.json` |
| throw away uncommitted work (`reset --hard`, `clean -f`, `checkout -- .`, `restore .`, `stash clear`) | guard hook |
| edit `AGENTS.md`, `CLAUDE.md`, `.claude/`, `.github/`, `Taskfile.yml`, `.golangci.yml` or the hooks on a `task/` branch | protect hook (agents file a follow-up instead). The design docs in `docs/` stay editable |
| touch your main checkout from a worktree | Claude Code's own worktree isolation |

One more hook helps rather than refuses: after every edit to a Go file, the after-edit hook formats the file with the repo's formatters and lints its package. The agent sees any lint findings in that file straight away, instead of at the check. If golangci-lint isn't installed, it does nothing.

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

For a `size: L` or `kind: seam` task, add `--model opus`: those are where a weaker first draft costs the most.

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

Then type `/run-tasks M2-06 M2-12`, or `/run-tasks 3` for the first three ready tasks. Each task gets a `task-implementer` agent in its own worktree, running in the background, and the session reports each PR as it lands. It won't start two tasks that change the same files. Large tasks and seams run on Opus, the rest on Sonnet.

### Plan the next milestone

When a milestone is done, type `/plan-milestone` in an Opus session (`claude --model opus`). It:

1. reads the next outline brief, the design docs, the last milestone's Notes and the code as it is now;
2. asks you the brief's open questions and records your answers in the design docs;
3. shows you the task breakdown and waits for your go-ahead;
4. writes the task specs, the brief's task table and the index rows, in one PR.

Follow-ups aimed at the milestone are folded into its tasks, or given real dependencies on them. If a cleanup (`C<n>`) is still running, the new milestone's first tasks wait for its lock-in.

### Clean up between milestones

Once a milestone's tasks are merged, and nothing else is running:

1. `/audit` starts a `code-auditor` per area of the code, plus one for vocabulary, and files what they find as a cleanup milestone (`tasks/C1-cleanup/`, then `C2`, ...) in one PR. Names it isn't sure of, and new libraries, become questions.
2. `/decide` settles those questions. Names go into `docs/glossary.md`.
3. `/triage` accepts, merges or drops the cleanup tasks.
4. `/run-tasks` does them. Renames run one at a time, then duplication and libraries per area, then tests per area, then comments per area, then the content guide (written on the first audit, checked for drift after). Each task proves it changed nothing with `go tool task check-pr`, which CI runs again on the PR.
5. `/audit lock-in`, in your session, turns what was decided into lint rules, AGENTS.md style and rule files, so the next audit finds less.

The auditors run on Opus, seven at a time, so an audit uses a fair share of your plan.

### Speed the game up

Once the benchmarks exist (C2-03 adds them), type `/optimize` in an Opus session, or name targets: `/optimize step harness`. It:

1. runs each target's benchmarks six times, then once more with the CPU and memory profilers, one target at a time so the timings don't skew each other;
2. starts a `perf-auditor` per target, which reads the profiles and the code behind the hot spots and ranks the bottlenecks by expected gain;
3. files the fixes as a `P<n>` milestone of proposed tasks, with the baseline numbers, in one PR.

Then `/triage` and `/run-tasks` as usual. Each task proves its gain with `benchstat` before and after on the same machine, and changes no golden.

Close anything heavy while it measures. To run benchmarks yourself:

```powershell
go tool task bench BENCH=Step -- -count 6 ./internal/core/game
```

### Triage follow-ups

Type `/triage`. It checks each `proposed` follow-up against the code as it is now, and marks the ones another change already handled as obsolete. For each of the rest, it asks you what to do: `todo` in a milestone, fold it into a task that isn't done, won't do, or not now. Everything lands in one PR: the obsolete ones come first, each with the commit that made it so, so you can push back in review. Design questions are left to `/decide`.

### Answer design questions

Type `/decide` in any session, including one in the Code tab on your phone. It asks you each open design question as multiple choice, with its recommendation first and "Not now" last. Your answers go into the task files and `docs/game-design.md` in one PR, and the questions with work left become `todo` for an agent.

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

1. Read the PR on GitHub. CI runs `go tool task check` again. CI also posts a comment on the PR, edited in place on each push, listing the slowest test packages next to master's times. It flags, without failing, any package over 30 seconds or at least twice as slow as on master.
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
