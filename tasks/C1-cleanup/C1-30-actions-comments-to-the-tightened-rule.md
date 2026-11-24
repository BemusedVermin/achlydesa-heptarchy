---
id: C1-30
title: "Actions comments to the tightened rule"
status: todo
milestone: C1
kind: debt
depends_on: []
discovered_by: null
size: M
---

# C1-30 · Actions comments to the tightened rule

## Goal

The actions packages' comments meet `.claude/rules/comments.md`.

## Context

- C1-20 trimmed these comments before the owner tightened the rule: no end-of-line comments, no comments in tests other than doc comments, no task IDs in comments.
- `go tool task comments -- <paths>` lists every comment that breaks the rule.

## Deliverables

- One pull request that changes only comments in `internal/core/actions`, `internal/core/effects`, `internal/core/timeline`.

## Spec

Fix every comment that `go tool task comments -- ./internal/core/actions ./internal/core/effects ./internal/core/timeline` lists (92 when the rule tightened):

- **End-of-line comments:** cut them. Where a field's note says something its name and type don't, make it a doc comment above the field.
- **Comments in tests:** cut them. Where one says something the test doesn't, move it into the test or subtest name, or the failure message.
- **Task IDs:** say what the code does instead.

Leave doc comments alone otherwise: the owner is happy with them.

## Acceptance criteria

- [ ] The `go tool task comments` command above reports nothing
- [ ] `go tool task same-behavior` passes
- [ ] `go tool task check` passes
- [ ] Only comments changed, apart from test or subtest names and failure messages that took over a comment's job

## Out of scope

- Comments in other packages: each area's cleanup task covers its own.

## Notes

- Filed 2026-11-23 when the owner tightened the comment rule.
