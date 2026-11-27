---
id: M6-27
title: "Graftable and exotic recruits"
status: todo
milestone: M6
kind: content
depends_on: [M6-21, M6-26]
discovered_by: null
size: S
---

# M6-27 · Graftable and exotic recruits

## Goal

Recruits can be grafted like the party, and the rarer recruits from the deep join with an unusual ability and unusual wants.

## Context

- `docs/game-design.md` → "Recruits" (Recruits can be grafted; a grafted deserter takes the piece); "Exotic recruits" (tamed moving clay, a leashed soul-feeder, a walking corpse).
- Added back from the reverted M3: `tasks/M3-agents/M3-17-exotic-recruits.md`.
- M6-21's moving clay family, M6-26's recruits.

## Deliverables

- `content/data/templates.cue`: recruit templates tagged `graftable`; at least one exotic recruit (tamed moving clay: wants the dead, takes a severed enemy part as its own).
- The exotic ability is composed from existing effect primitives and triggers, with no new Go code unless a gap is found; any gap is filed as a follow-up.
- Tests: a grafted recruit carries its piece when it quits; the tamed clay consumes a corpse for its drive and gains the part's action.

## Spec

- **Firm:** a new exotic recruit is data. Only soul-feeders and the still town's walking corpse are left for the milestones that add them.

## Acceptance criteria

- [ ] The tests above pass.
- [ ] `go tool task check` passes.

## Out of scope

- Soul-feeders and the walking corpse: M7 and M8.

## Notes

