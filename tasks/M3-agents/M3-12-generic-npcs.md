---
id: M3-12
title: Generic NPCs and replacement timers
status: todo
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

- [ ] A test kills a generic guard in the test town; a new guard appears after the content number of ticks.
- [ ] In a town that has emptied, killing the last generic brings no replacement.
- [ ] `go tool task check` passes.

## Out of scope

- Hiring generics as recruits (M3-15).

## Notes


