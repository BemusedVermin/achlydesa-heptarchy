---
id: M6-18
title: "Hearsay: agents tell each other what they know"
status: done
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

- [x] The tests above pass.
- [x] `go tool task check` passes.

## Out of scope

- Drift, rumors, distortion: M8.

## Notes

- Telling is a `tell` effect on a new `action.chat` (target a body within range 2 and sight, recovery one step), not an operator. The mind controller starts it when its goal is `agents.chat_goal`; `Choose` skips that goal unless a partner in reach has something to hear, so `goal.chat` (weight 0.08, above wandering) never idles an agent.
- Content gained `agents.chat_goal`, `chat_action`, `told_origin` and `told_confidence` (0.6). A told fact keeps its subject, claim, value and cell; its confidence is the speaker's times the factor, so a retold fact weakens again.
- The speaker picks by confidence, then recency, skipping facts about the hearer and any the hearer holds (`knowledge.Best`) at the told copy's confidence or more. A seen fact outranks a told one by `Best`.
- No language exists in the design or code, so any two living agents with minds can tell each other. Filed M6-18.F1 (question).
- No template lists `action.chat` yet; M6-20 gives minds their actions.

