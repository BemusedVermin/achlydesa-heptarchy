---
id: M6-04
title: A capped memory per agent, learned by sight
status: todo
milestone: M6
kind: seam
depends_on: []
discovered_by: null
size: M
---

# M6-04 · A capped memory per agent, learned by sight

## Goal

An agent remembers what it has seen, up to a cap, forgetting the oldest and least certain first. Nothing else it knows comes from the truth.

## Context

- `docs/game-design.md` → "How minds work" → "Beliefs"; "Keeping it affordable".
- `internal/core/knowledge`: the fact store (subject, claim, origin, confidence) per knower; today only the party writes to it.
- Sight: `hexmap` sight and the `combat` alert's visible-cells query. Sight per agent is the hot spot at 1,000 agents.

## Deliverables

- `internal/core/knowledge/`: `Learn` keeps at most `knowledge.cap` facts per knower (content), dropping the lowest by (confidence, tick); `Forget`.
- `internal/core/agents/`: `sense(w, body)`, which learns what a near agent sees within `agents.sight_radius`: bodies, landmarks (M6-07) and places worth remembering, with a content list of which claims to record.
- `content/data/knowledge.cue`: `claim.at` (a thing is at a cell), `claim.dead`, `origin.seen`, `origin.told`; first values for the cap and radius.
- Tests: the cap holds and the right fact is dropped; a body sees only inside its radius and line of sight; the party's own facts are not capped.

## Spec

- **Firm:** `sense` runs only on an agent's own timeline entries, capped by radius, never on every entry of the world.
- **Firm:** a belief about a moving thing is a claim with a tick; a contradicting sighting replaces it (same subject, claim, origin).
- **Firm:** the cap and radius are content. First values: 64 facts, 8 hexes, for tuning in M9.
- A fact about a thing gone (a well that was a place's last drop) is not corrected until seen again; that is the point.

## Acceptance criteria

- [ ] The tests above pass.
- [ ] A benchmark in the test file stays within a stated budget for 1,000 near agents sensing once (recorded in Notes).
- [ ] `go tool task check` passes.

## Out of scope

- Telling facts to other agents: M6-18.
- Planning over beliefs: M6-10.

## Notes

