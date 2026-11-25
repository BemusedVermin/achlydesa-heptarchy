---
id: M3-17
title: Graftable and exotic recruits
status: done
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

- [ ] A hired tamed moving clay that sees an enemy part severed takes it and gains its action. Not met: left out, see Notes and M3-17.F1.
- [x] A graftable recruit can receive a graft with `action.graft`, and an exotic recruit cannot.
- [x] `go tool task check` passes.

## Out of scope

- The walking corpse (the still town arrives later).

## Notes

- **Clay ability left out.** `attach_part` attaches to the trigger's target (the maimed body), and triggers fire only for the event's subjects, so a clay cannot take a part it only saw cut. The spec says to file a follow-up and leave it out: M3-17.F1. The tamed clay is hirable but has no ability yet, so the first acceptance criterion is open.
- **Soul-feeder.** `_feederGrows` adds `status.feeding` to a feeder that damages a body for which `dying` holds; `rule.hit` adds 2 for it. It covers the dying on either side as long as the feeder is the one striking.
- **Graftable.** The `graftable` tag moved onto the generic guard and field hand (it was on no generic); exotic recruits lack it, which is what keeps them from grafts.
- **Content added:** `item.bone` (the clay's want and upkeep; the party starts with 3), two families, and the lake town keeps one of each exotic recruit.
- Tests: `internal/core/combat/exotic_test.go`, `TestAGraftableRecruitTakesAGraftAndAnExoticOneDoesNot` in `forms_test.go` (its ring gained `recruit` and `exotic` bodies), `internal/core/agents/exotic_test.go`.
- `go tool task fmt` fails here (the installed golangci-lint is older than the module's Go version); I ran gofmt directly.
- Follow-up filed: M3-17.F1.
