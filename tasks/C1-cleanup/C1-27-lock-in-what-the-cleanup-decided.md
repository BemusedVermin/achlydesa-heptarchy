---
id: C1-27
title: "Lock in what the cleanup decided"
status: proposed
milestone: C1
kind: debt
depends_on: [C1-17, C1-18, C1-19, C1-20, C1-21, C1-22, C1-23, C1-24, C1-25, C1-26]
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

- [ ] `/audit lock-in` merged.
- [ ] `go tool task check` passes.

## Out of scope

- Anything the cleanup tasks do.

## Notes

