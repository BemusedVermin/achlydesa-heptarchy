---
id: M6-03
title: "Drives: needs and desires with a weight per individual"
status: done
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

- [x] The tests above pass.
- [x] `go tool task check` passes, and the save version is bumped.

## Out of scope

- Choosing a goal from drives: M6-14.
- Eating and drinking by the party: M6-24.

## Notes

- **No template carries drives yet.** The drives exist in `content/data/drives.cue`, but M6-21 and M6-24 attach them to creatures and the party. Attaching them now would starve the party in normal play, with no way to drink until M6-24. The tests add drives to the beast template on a copied bundle.
- **Weights:** `agents.weight` (new `agents` config, in `drives.cue`) is a formula over `typical` and `roll`; the first value draws 0.5 to 1.5 times the typical weight. One roll per drive, in drive ID order. A negative weight is a content error.
- **Stream per spawn:** `rng.Derived("agents", salt)`, where the salt hashes the template and the spawn cell. A body spawned with no position (a party member) uses its UID instead, which game start hands out in a fixed order. Two bodies of one template spawned on the same cell would draw the same weights; M6-20 may pass its own per-chunk salt if that ever matters.
- **Levels** start at 0, never go below 0, and have no ceiling. The tick (`agents.drives`, per phase) skips dead bodies, raises every drive, then runs the thresholds each level has reached, with the body as source and target. It stops early when a threshold kills the body.
- **Desires** may have a rise but no thresholds, since a desire never kills (glossary). `drive.pay` has no rise yet; M6-25 adds one.
- **Urge** must come out between 0 and 1; anything else is a content error at run time. Nothing calls it yet (M6-14 will).
- **`satisfy`** (`bodies`, `drive`, `amount`) lowers the level and stops at 0; a negative amount raises it. A body that doesn't carry the drive is skipped.
- **Thresholds** apply only damage for now. The design wants "slowed first", which needs a status that the speed formula reads. That is M6-24's job.
- Files outside the Deliverables: `internal/content/{types,load,records,registry,actions}.go` (the wiring), `internal/core/game/install.go` (installs `agents`), `internal/core/save/save.go` and `save_test.go` (save version 9), and `docs/content-guide.md` (the new Drives section, `drives.cue` in the layout, `satisfy` in the primitives).
- Follow-up filed: M6-03.F1 (question: what raises `drive.safety`; it ships with `rise: "0"`).

