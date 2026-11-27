---
id: M6-03
title: "Drives: needs and desires with a weight per individual"
status: in-progress
milestone: M6
kind: seam
depends_on: []
discovered_by: null
size: M
---

# M6-03 · Drives: needs and desires with a weight per individual

## Goal

Every body can carry drives: needs that rise with time and can kill, and desires that don't. Each individual has its own weight on each drive, so personality is data on the individual.

## Context

- `docs/game-design.md` → "How minds work" → "Drives"; "People with needs"; "Party" → "Food and water".
- `docs/go-technical-design.md` → "Designing for pivots" → "Behaviors", "Determinism rules" (named streams).
- The reverted M3 needs work is in git history (`tasks/M3-agents/M3-02-needs.md`); take the ideas, not the code: it ticked per phase and applied threshold effects.

## Deliverables

- `internal/content/`: `Drive` records (`drive.water`, `drive.pay`), each with `kind: need|desire`, an Expr `rise` per phase, an Expr `urge` (0..1), and thresholds that apply effects while the level is at or above them. `Template.Drives` lists typical weights.
- `internal/core/comp/`: `Drives{Levels, Weights}` keyed by drive ID; regenerate the schema.
- `internal/core/agents/`: the per-phase tick, the `satisfy` effect (a negative amount raises a level), and the spawn rule that draws an individual's weights around its template's, from the `agents` stream.
- `content/data/drives.cue`: `drive.water`, `drive.food`, `drive.rest`, `drive.safety` (needs) and `drive.pay` (desire), first values for tuning.
- Tests: levels rise by their formula; a threshold effect fires while held; two individuals of one template get different weights and the same seed gives the same weights.

## Spec

- **Firm:** needs tick per phase, not per timeline entry, to keep the timeline cheap.
- **Firm:** weights are drawn once at spawn and never redrawn; they are saved. Changing weights later (grief, grudges) is M8.
- **Firm:** the stream is derived per spawn from the world seed and the spawn's chunk or place, so load order never decides a weight (called out because M6-20 spawns the whole world).
- **Suggestion:** thresholds reuse the `triggers` effect shape, so starvation is `damage` on vital parts.

## Acceptance criteria

- [ ] The tests above pass.
- [ ] `go tool task check` passes, and the save version is bumped.

## Out of scope

- Choosing a goal from drives: M6-14.
- Eating and drinking by the party: M6-24.

## Notes

