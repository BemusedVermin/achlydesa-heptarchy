---
id: C1-13
title: "May tests share one cached sample registry?"
status: proposed
milestone: C1
kind: question
depends_on: []
discovered_by: null
size: S
---

# C1-13 · May tests share one cached sample registry?

## Goal

The owner picks one option, and the answer is recorded here and where the glossary or design doc needs it.

## Context

- Content audit D4, game audit D1, foundation audit D4.

## Deliverables

- The answer, written in this task's Notes and in `docs/glossary.md` (and `docs/` where the design says otherwise).

## Spec

The sample content loads 37 times in 28 test files, and `sampleDir` is defined 10 times. A `sync.OnceValues`-cached `contenttest.Sample(t)` would load it once per test binary. The registry is immutable and safe for concurrent use, but AGENTS.md says no global mutable state.

- **Allow it, test-only**: A cached registry in a test-support package, and a deep-copied bundle for tests that mutate.
- **No cache**: A shared helper that still loads each time, which is slower.

**Recommendation:** Allow it for the immutable registry only.

## Acceptance criteria

- [ ] The owner has answered, through `/decide`.
- [ ] The answer is in Notes and in the glossary.

## Out of scope

- Doing the work: the cleanup task named in the milestone brief does it.

## Notes

