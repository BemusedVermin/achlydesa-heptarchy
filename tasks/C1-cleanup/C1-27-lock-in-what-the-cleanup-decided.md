---
id: C1-27
title: "Lock in what the cleanup decided"
status: done
milestone: C1
kind: debt
depends_on: [C1-17, C1-18, C1-19, C1-20, C1-21, C1-22, C1-23, C1-24, C1-26, C1-28, C1-29, C1-30]
discovered_by: null
size: S
---

# C1-27 · Lock in what the cleanup decided

## Goal

What the cleanup decided becomes lint rules and rules for agents, so the next audit finds less.

## Context

- Milestone brief: [README.md](README.md). The `/audit` skill, section "Lock-in".

## Deliverables

- `.golangci.yml`, `AGENTS.md`, `.claude/rules/*`, the lens files under `.claude/skills/audit/lenses/`, `Taskfile.yml`: changed by the owner.

## Spec

**Run `/audit lock-in` in the owner's session.** It changes files that agents cannot touch on a `task/` branch. It tries each "Lint:" rule the auditors named (`gochecknoglobals` for `internal/`, `perfsprint`, `thelper`, `predeclared`, `intrange`, `modernize`, `unparam`, `usestdlibvars`, and the forbidigo-style check for task IDs in comments where it can be done), keeps what `golangci-lint run` passes, and adds the decided conventions to Style and the rule files. It also adds `cue fmt --check ./content/...` to the `content` task in `Taskfile.yml`.

## Acceptance criteria

- [x] `/audit lock-in` merged.
- [x] `go tool task check` passes.

## Out of scope

- Anything the cleanup tasks do.

## Notes


- Triaged 2026-11-23: todo in C1, as the owner decided; it is part of the first cleanup milestone.
- Run in the owner's session on branch `ccr-fcc21494-1p7qgm`, from `origin/master`.
- **Lint rules on:** `intrange`, `modernize` (without `omitzero`), `perfsprint`, `predeclared`, `thelper`, `unparam`, `usestdlibvars`. The 7 perfsprint, 3 intrange, 16 modernize, 4 thelper and 6 unparam findings are fixed in code; no behavior changed. `golangci-lint` 2.15 is needed: the 2.5 binary was built with Go 1.25 and refuses this module.
- **Left off:** `gochecknoglobals`. All 21 non-test hits are read-only tables or the immutable sample registry, which the owner allowed; converting them to functions is churn, and the linter can't tell them from mutable state. The conventions lens says to check by reading. `modernize`'s `omitzero` is disabled: following it removes `omitempty` from `Vocabulary` and `Reason`, which makes those schema fields required.
- **No task-ID check:** `forbidigo` reads code, not comments, so `go tool task comments` (agentkit) stays the check.
- **Content:** `go tool task content` now also runs `cue fmt --check`; `content/data/text/explore.cue` was misformatted and is fixed.
- **Docs:** `AGENTS.md` (read-before list, Style), `.claude/rules/` (`content`, `core`, `frontend`, new `tests` and `tools`), and the lens files.
- Follow-up: none filed.
