---
id: C1-14
title: "May `go.yaml.in/yaml/v3` become a direct dependency?"
status: done
milestone: C1
kind: question
depends_on: []
discovered_by: null
size: S
---

# C1-14 · May `go.yaml.in/yaml/v3` become a direct dependency?

## Goal

The owner picks one option, and the answer is recorded here and where the glossary or design doc needs it.

## Context

- Tools audit L1.

## Deliverables

- The answer, written in this task's Notes and in `docs/glossary.md` (and `docs/` where the design says otherwise).

## Spec

`internal/tools/agentkit/ready.go` hand-rolls a YAML subset (about 45 lines and 40 of tests). The module is already in `go.sum` as an indirect dependency through Task. Rule 8 requires the owner's say for a new dependency.

- **Yes**: Use it for the front matter. Malformed front matter becomes an error instead of being half-read.
- **No**: Keep the parser.

**Recommendation:** Yes.

**Decided by the owner:** yes. `go.yaml.in/yaml/v3` becomes a direct dependency, used for task front matter in agentkit, so malformed front matter is an error instead of being half-read. Applied by C1-25 (the owner's agentkit task).

## Acceptance criteria

- [x] The owner has answered, through `/decide`.
- [x] The answer is in Notes and in the glossary.

## Out of scope

- Doing the work: the cleanup task named in the milestone brief does it.

## Notes
Decided 2026-11-23: yes. Why: it replaces a hand-rolled parser and is already in go.sum through Task.
