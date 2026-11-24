---
id: M3-14
title: "Rest in town: heal integrity and set the rest place"
status: todo
milestone: M3
kind: feature
depends_on: [M3-06]
discovered_by: null
size: M
---

# M3-14 · Rest in town: heal integrity and set the rest place

## Goal

The party can rest in a town. Resting moves time on, restores damaged parts' integrity at a content rate (a severed part never comes back), and makes that town the place the party wakes after a wipe.

## Context

- `docs/game-design.md` → "What to do in town" (Rest and dream; Heal and graft) and its "Healing" paragraph (the owner's decision: integrity only), "Three tiers of people" (waking after a wipe).
- `docs/glossary.md` → rest.
- `internal/core/explore/wake.go` (`SetRestPlace`, `Wake`), `comp.RestPlace`; M3-06's gates; the effects registry.
- Folds in [M2-07.F1.F1.F1](../followups/M2-07.F1.F1.F1-rest-sets-rest-place.md): resting in a town sets the party's rest place.

## Deliverables

- `internal/content/types.go`: `MapDef.Tags`; `ActionEnv`/`BodyEnv` gain `map_tags`; schemas regenerated.
- `internal/core/combat/effects.go`: a `heal` primitive (parts' integrity up to their max; severed parts untouched).
- `internal/core/explore/wake.go`: a `set_rest_place` primitive that records the source's party and cell.
- `content/data/actions.cue`: `action.rest`, required to be on a map tagged `town`, lasting until the next dawn (content formula), healing by a content rate per tick rested, setting the rest place.
- The test town and the playground's town get the `town` tag.
- Tests: rest heals, rest sets the rest place, a wipe wakes there.

## Spec

**Firm:**

- `heal` takes `parts` (a selector) and `amount` (an Expr). It never re-adds a severed part and never raises integrity above the part's max.
- `set_rest_place` takes no parameters; it sets the party of the source body's rest place to the source's map and cell.
- `action.rest` is all content: its `requires` reads `"town" in source.map_tags`, its wind-up runs to the next dawn, and its `on_land` heals and sets the rest place. A party member resting is busy for that long; resting is offered only when no hostile is in sight.
- Dreams are M6; rest emits no dream.

## Acceptance criteria

- [ ] M2-07.F1.F1.F1's criterion: a party that rests in a town and is then wiped wakes in that town.
- [ ] A member with a damaged arm and a severed hand rests one night: the arm's integrity rises by the content amount, and the hand is still gone.
- [ ] Rest is not offered outside a town map, nor with a hostile in sight.
- [ ] `go tool task check` passes.

## Out of scope

- Dreams (M6). Healing over several days of rest is just resting again; no multi-day command.

## Notes


