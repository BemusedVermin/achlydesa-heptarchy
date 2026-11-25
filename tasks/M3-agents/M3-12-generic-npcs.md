---
id: M3-12
title: Generic NPCs and replacement timers
status: done
milestone: M3
kind: feature
depends_on: [M3-11]
discovered_by: null
size: S
---

# M3-12 · Generic NPCs and replacement timers

## Goal

Towns keep their generic people: guards, merchants and would-be recruits with very basic needs are replaced on a timer while the town is inhabited, and an emptied town stops replacing them.

## Context

- The agents model in [README.md](README.md), "Towns".
- `docs/game-design.md` → "Three tiers of people" (generic: replaced on a timer while the town is still inhabited).
- `docs/go-technical-design.md` → the `clock` row (timers for generic NPC replacement).
- M3-08's `Town`, M3-10's `TownEmptied`, M3-11's residents.

## Deliverables

- `internal/content/types.go`: `TownDef.Generics` (template, how many, replacement ticks); schemas regenerated.
- `internal/core/comp/comp.go`: `Generic{Town api.UID}` and the town's next-replacement ticks; saved.
- `internal/core/agents/generics.go`: replacement, on screen as bodies and off screen as counts.
- `content/data/templates.cue`: a generic guard and a generic field hand with very basic needs (food, safety).
- `internal/core/agents/generics_test.go`.

## Spec

**Firm:**

- Each town keeps up to `count` generics of each template. When one dies or leaves the town, a replacement is due `ticks` later; it appears at the town's gate mark if the party is there, otherwise in the town's count.
- A town with no residents left replaces nobody, ever, so emptied towns stay empty.
- Replacement uses the `lod` stream for its placement only.

## Acceptance criteria

- [x] A test kills a generic guard in the test town; a new guard appears after the content number of ticks.
- [x] In a town that has emptied, killing the last generic brings no replacement.
- [x] `go tool task check` passes.

## Out of scope

- Hiring generics as recruits (M3-15).

## Notes

- `comp.Generic.Town` is the town's `api.MapID`, not a UID: region entities have no UID, and giving them one would shift every later UID.
- Replacements are not queued by hooking deaths and departures. A per-entry system (`agents.generics`) compares what each town holds (living generics on its map plus its family count) with the wanted count, and queues a timer for each generic missing and not already queued. A loss is noticed at the next timeline entry, so the timer starts then.
- A generic's template must be the only template of exactly one family (content check); that family's count is the off-screen count. Generics need `residents` on the town. "Emptied" means the residents family has no living body on the map and no count; generics need not belong to that family. The pending queue is dropped once the town is emptied.
- Arrival on screen is one of the free cells nearest the map's first gate, picked with the `lod` stream.
- The test town gained generics (2 guards, 1 field hand) and `residents: family.people`. Save version raised to 7. No docs changed.
- Follow-ups filed: none.


