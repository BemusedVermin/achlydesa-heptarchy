---
id: C1-29
title: "Content comments to the tightened rule"
status: done
milestone: C1
kind: debt
depends_on: []
discovered_by: null
size: M
---

# C1-29 · Content comments to the tightened rule

## Goal

The content packages' comments meet `.claude/rules/comments.md`.

## Context

- C1-19 trimmed these comments before the owner tightened the rule: no end-of-line comments, no comments in tests other than doc comments, no task IDs in comments.
- `go tool task comments -- <paths>` lists every comment that breaks the rule.

## Deliverables

- One pull request that changes only comments in `internal/content`.

## Spec

Fix every comment that `go tool task comments -- ./internal/content` lists (115 when the rule tightened):

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
- Field notes became doc comments above the field, reworded where the "defines X as itself" lint objected; notes that restated the name or type were cut.
- Renamed test struct fields `want` to `wantFragments` (load_test.go) and `wantErr` (rules_test.go) so they carry what their comments said.
- Regenerated `internal/content/*_go_gen.cue` with `go tool task schema`; the new doc comments flow into them.
- `.claude/rules/content.md` still says text-ID fields are commented `// a text ID`, which the tightened rule forbids at end of line. Fields now carry `// Name is a text ID.` above them. That rule file is off-limits on a task branch; the owner may want it updated.
