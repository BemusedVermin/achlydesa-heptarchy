---
id: C1-28
title: "Foundation comments to the tightened rule"
status: done
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

- C1-22 trimmed these comments before the owner tightened the rule: no end-of-line comments, no comments in tests other than doc comments, no task IDs in comments.
- `go tool task comments -- <paths>` lists every comment that breaks the rule.

## Deliverables

- One pull request that changes only comments in `internal/core/api`, `internal/core/sim`, `internal/core/comp`, `internal/core/anatomy`, `internal/core/rng`, `internal/core/clock`, `internal/core/hexmap`, `internal/core/knowledge`, `internal/core/inventory`.

## Spec

Fix every comment that `go tool task comments -- ./internal/core/api ./internal/core/sim ./internal/core/comp ./internal/core/anatomy ./internal/core/rng ./internal/core/clock ./internal/core/hexmap ./internal/core/knowledge ./internal/core/inventory` lists (118 when the rule tightened):

- **End-of-line comments:** cut them. Where a field's note says something its name and type don't, make it a doc comment above the field.
- **Comments in tests:** cut them. Where one says something the test doesn't, move it into the test or subtest name, or the failure message.
- **Task IDs:** say what the code does instead.

Leave doc comments alone otherwise: the owner is happy with them.

## Acceptance criteria

- [x] The `go tool task comments` command above reports nothing
- [x] `go tool task same-behavior` passes
- [x] `go tool task check` passes
- [x] Only comments changed, apart from test or subtest names and failure messages that took over a comment's job

## Out of scope

- Comments in other packages: each area's cleanup task covers its own.

## Notes

- Filed 2026-11-23 when the owner tightened the comment rule.
- End-of-line notes that said something the name and type didn't became doc comments above the field; the rest were cut. Three "why" notes moved onto the line above.
- Test comments were cut; none was moved into a test name, as the tests already say it.
