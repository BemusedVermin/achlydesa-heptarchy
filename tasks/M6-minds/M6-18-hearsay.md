---
id: M6-18
title: "Hearsay: agents tell each other what they know"
status: todo
milestone: M6
kind: feature
depends_on: [M6-04, M6-11]
discovered_by: null
size: S
---

# M6-18 · Hearsay: agents tell each other what they know

## Goal

When two near agents meet, each may pass a fact to the other, so a newcomer can learn where the water is. A told fact is lower in confidence and marked as told.

## Context

- `docs/game-design.md` → "How minds work" (Beliefs: seen and heard); "Grudges and memory"; "People with needs" (rumors are M8).
- `internal/core/knowledge` (M6-04's cap); `docs/glossary.md`: origin.

## Deliverables

- `internal/core/agents/tell.go`: an operator-free `tell` step the `mind` controller may take when another agent of a shared language is in reach: the speaker picks its fact by confidence and recency, which the hearer has not got, and copies it with `origin.told` and a content confidence factor.
- `content/data/`: the factor (`agents.told_confidence`, first value 0.6) and a `goal.chat`, weak, so it happens when nothing else is pressing.
- Tests: a told fact reaches the hearer at lower confidence; telling never raises confidence; a hearer does not accept what it already holds with more confidence; telling happens only within reach.

## Spec

- **Firm:** telling is near-only: both agents are near and in reach, so it costs no search.
- **Firm:** a fact does not drift or change when told; how stories change in the retelling is M8's rumors.
- **Firm:** a told fact is replaced by a seen one (same subject and claim) automatically.

## Acceptance criteria

- [ ] The tests above pass.
- [ ] `go tool task check` passes.

## Out of scope

- Drift, rumors, distortion: M8.

## Notes

