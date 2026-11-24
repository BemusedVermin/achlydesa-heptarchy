---
id: M3-17
title: Graftable and exotic recruits
status: todo
milestone: M3
kind: content
depends_on: [M3-16, M3-18]
discovered_by: null
size: S
---

# M3-17 · Graftable and exotic recruits

## Goal

Recruits come in the design's kinds: graftable humans, the cheapest way to learn what a piece does, and rarer exotic recruits from the deep, who can't be grafted, each with a strange want and a special ability.

## Context

- `docs/game-design.md` → "Recruits" (graftable recruits) and "Exotic recruits" (the sketches table).
- M3-15 and M3-16; M3-18's families; the existing primitives `attach_part`, `spawn_from_part`; triggers.

## Deliverables

- `content/data/templates.cue`: a tamed moving clay (wants the dead; takes a severed enemy part as its own through a trigger and `attach_part`) and a leashed soul-feeder (wants to feed; stronger near the dying through a status and a rule), both hirable somewhere in the sample world.
- Tests for each ability and for grafting a human recruit.
- `docs/content-guide.md`: a short example of an exotic recruit.

## Spec

- Firm: everything is content. If an ability needs a primitive that doesn't exist, file a follow-up and leave that ability out.
- The walking corpse from the still town needs the still town, which is not in M3; leave it out.

## Acceptance criteria

- [ ] A hired tamed moving clay that sees an enemy part severed takes it and gains its action.
- [ ] A graftable recruit can receive a graft with `action.graft`, and an exotic recruit cannot.
- [ ] `go tool task check` passes.

## Out of scope

- The walking corpse (the still town arrives later).

## Notes


