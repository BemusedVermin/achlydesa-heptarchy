---
id: C1-28
title: "Foundation comments to the tightened rule"
status: todo
milestone: C1
kind: debt
depends_on: []
discovered_by: null
size: M
---

# C1-28 · Foundation comments to the tightened rule

## Goal

The foundation packages' comments meet `.claude/rules/comments.md`.

## Context

- C1-22 trimmed the foundation's comments before the rule tightened. The owner then ruled out end-of-line comments, comments in tests other than doc comments, and task IDs in comments.
- `go tool task comments -- <paths>` lists every comment that breaks the rule.

## Deliverables

- One pull request that changes only comments in the foundation packages: `internal/core/api`, `sim`, `comp`, `anatomy`, `rng`, `clock`, `hexmap`, `knowledge` and `inventory`.

## Spec

Fix every comment that `go tool task comments -- ./internal/core/api ./internal/core/sim ./internal/core/comp ./internal/core/anatomy ./internal/core/rng ./internal/core/clock ./internal/core/hexmap ./internal/core/knowledge ./internal/core/inventory` lists (118 when the rule tightened):

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

- Comments outside the foundation packages: each area's cleanup task covers its own.

## Notes

- Filed 2026-11-23 when the owner tightened the comment rule.
