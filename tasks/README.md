# Tasks

Every piece of work is a markdown file with YAML front matter. `INDEX.md` is the at-a-glance tracker; each task file is the spec an agent implements from.

## Layout

```
tasks/
├── INDEX.md              # every task, its status and dependencies
├── TEMPLATE.md           # copy this for new tasks and follow-ups
├── M1-skeleton/          # README.md (milestone brief) + one file per task
├── M2-combat/
├── M3-agents/ ... M9-tuning/   # briefs only until the milestone is broken down
├── post-mvp/             # post-MVP milestones, one folder each
└── followups/            # tasks agents filed while working; triaged by the owner
```

## Front matter

```yaml
---
id: M1-03
title: "sim: world, event bus and system scheduler"   # quote titles that contain a colon
status: todo            # proposed | todo | in-progress | blocked | done | obsolete
milestone: M1           # M1..M9, a cleanup milestone (C1, C2, ...), a post-MVP milestone (PM-STORY), or unplanned
kind: feature           # feature | seam | bug | debt | test | content | question
depends_on: [M1-01]
discovered_by: null     # the task ID that filed it, for follow-ups
size: M                 # S (< 200 lines), M (200-600), L (> 600: split it)
---
```

The task file is the source of truth for status. `INDEX.md` mirrors it, so update both in the same change.

## Statuses

| Status | Meaning | Who sets it |
| --- | --- | --- |
| `proposed` | Filed as a follow-up, not yet triaged | The agent that filed it |
| `todo` | Ready to start once its dependencies are done | The owner, or a triage task the owner assigned |
| `in-progress` | Someone is on it | The implementer |
| `blocked` | Can't continue; Notes say why and link the blocker | The implementer |
| `done` | Acceptance criteria met and the PR is open; merging the PR completes the task | The implementer |
| `obsolete` | Another task made it unnecessary; Notes say which | The owner, or a triage task the owner assigned |

## Picking a task

Take the lowest-numbered `todo` task whose `depends_on` are all `done`. Tasks in the same milestone with no dependency between them can run in parallel.

`go run ./internal/tools/agentkit ready` lists the tasks that can start now. It reads the task files, not this index, and it also shows tasks that already have a `task/` branch, so two agents don't take the same one.

## Running agents

Claude Code is set up to run agents on tasks: `/do-task <ID>` takes one task to an open pull request, and `/run-tasks` runs several in parallel worktrees. Guardrails stop agents from committing to master, force-pushing, merging, or throwing work away. How to use it, and what each guardrail does: [../.claude/README.md](../.claude/README.md).

## Definition of done

- Every acceptance criterion is met and ticked in the task file.
- `go tool task check` passes: formatting, `go vet`, tests, lint and content. Until M1-01 lands, run `go build ./...`, `go test ./...` and `golangci-lint run` by hand.
- New packages have a `doc.go`; new exported identifiers have doc comments.
- If the work changed what `docs/game-design.md` or `docs/go-technical-design.md` says, the doc is updated in the same pull request.
- The Notes section lists assumptions, deviations, files touched outside the Deliverables, and follow-ups filed.

## Follow-ups

Agents file a follow-up for anything they notice outside their task's scope. The rules are in `AGENTS.md`. In brief:

- **File name:** `followups/<parent ID>.F<n>-<slug>.md`, numbered from F1 within the parent task, so parallel agents never collide.
- **Front matter:** `status: proposed`, `discovered_by: <parent ID>`, plus the best-guess `milestone` and `kind`.
- **Index:** add a row to the Follow-ups table in `INDEX.md`.
- **Questions:** use `kind: question` for design decisions. Give the options, what each costs, and a recommendation. Never settle a game-design question in code.

The owner triages follow-ups: each becomes `todo` in a milestone, gets merged into another task, or is closed with a reason. He can do it by hand, with `/triage` (it marks what no longer applies and asks him about the rest), or by handing an agent a triage task such as M2-15. Either way, the pull request proposes the decisions and merging it approves them. Design questions go through `/decide` instead.

### How a follow-up is triaged

- **Obsolete:** another change already fixed it, it no longer applies, or it duplicates another task. Set `status: obsolete`, and say which change or task in Notes.
- **`todo` in a milestone:** pick the earliest milestone whose goal, exit criteria or tasks need it; a bug that hurts current work goes in the current milestone. Set `status: todo` and `milestone`, then set `depends_on` so `agentkit ready` offers it at the right time. For a milestone that is still an outline, depend on the milestone itself, such as `[M3]`: `agentkit ready` shows it as waiting on M3, and `/plan-milestone` replaces that with the real tasks it needs.
- **Folded into another task:** if it fits a task that isn't done yet, add its point to that task's Spec, Deliverables and acceptance criteria. Then set the follow-up to `obsolete`, naming the task.
- **Not now:** leave it `proposed`.

Record each decision as one line in the follow-up's Notes, such as `Triaged 2026-11-23: todo in M2, because it blocks the M2 exit criteria.`, and change its `INDEX.md` row in the same commit.

## Cleanup milestones

Between milestones, the owner runs a cleanup: `/audit` has read-only agents look for comment bloat, duplication, tests that cost more than they protect, unclear names, hand-written code a library already does, and convention slips, then files the work as a milestone `C1`, `C2` and so on, in `tasks/C<n>-cleanup/`. The first audit also files a task to write `docs/content-guide.md`, the guide to making content in CUE, and later audits check it for drift. The owner settles the names and new libraries with `/decide` and the tasks with `/triage`; agents do the tasks; `/audit lock-in` turns what was decided into lint rules and agent rules. The steps are in `.claude/skills/audit/SKILL.md`.

Developer tools (`internal/tools/` and what they run) are excluded from cleanups: `/audit` does not sweep them, and a cleanup task does not touch them.

A cleanup task changes no behavior on purpose; its Notes name any change it makes anyway. A tests task (`kind: test`) changes only tests, which `go tool task check-pr` checks, and CI runs it on every pull request, along with the task bookkeeping checks.

## Writing a task spec

For whoever breaks a milestone into tasks, human or agent:

- **One reviewable change.** Aim for S or M. Split anything that would be L.
- **Name the files.** List what is created and modified.
- **Give the shapes.** Write the key types and function signatures, and say which are firm and which are suggestions.
- **Make acceptance testable.** Each criterion is a command to run or an observable behavior, not "works well".
- **Fence the scope.** Out of scope lists the tempting neighbours and which task owns them.
- **Link the design.** Point to the exact doc sections rather than restating them.
