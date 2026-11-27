---
id: M6-24
title: "The party eats, drinks and rests"
status: todo
milestone: M6
kind: feature
depends_on: [M6-03, M6-07, M6-13]
discovered_by: null
size: M
---

# M6-24 · The party eats, drinks and rests

## Goal

Party members have the same hunger, thirst and fatigue drives as anyone, met by the same take and consume actions; resting in town heals integrity and sets where the party wakes after a wipe.

## Context

- `docs/game-design.md` → "Party" → "Food and water"; "What to do in town" → "Healing"; "Three tiers" (party wipe).
- Added back from the reverted M3: `tasks/M3-agents/M3-13-party-food-water.md` and `M3-14-rest-heal.md`.
- `comp.RestPlace` and `explore.SetRestPlace` exist; the `heal` and `set_rest_place` primitives were reverted.

## Deliverables

- `content/data/templates.cue`: the party templates gain `drive.water`, `drive.food`, `drive.rest` with weights; the party's body is player-controlled, so drives only rise and apply their threshold effects (slowed first, then integrity lost from vital parts).
- The player eats and drinks from the party's inventory with M6-13's actions in the action menu.
- `internal/core/effects` or `combat`: `heal` (restores a damaged part's integrity at a content rate, never regrows a severed one) and `set_rest_place`.
- `content/data/`: a `rest` action on a bed landmark (M6-07) that moves time on, heals over days, and sets the rest place.
- Tests: a party member who goes without water is slowed and later loses integrity; drinking lake water spends the item Soak uses; resting heals a damaged part, never a severed one, and a wipe wakes the party at the bed.

## Spec

- **Firm:** the thresholds and rates are content, first values for tuning.
- **Firm:** the wipe screen (M5-14) is unchanged; only where it wakes the party can change.

## Acceptance criteria

- [ ] The tests above pass.
- [ ] `go tool task check` passes, and the save version is bumped.

## Out of scope

- Recruits: M6-25.

## Notes

