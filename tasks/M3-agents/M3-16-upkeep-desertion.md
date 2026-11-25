---
id: M3-16
title: Upkeep, grumbling, desertion and flight
status: done
milestone: M3
kind: feature
depends_on: [M3-04, M3-15]
discovered_by: null
size: M
---

# M3-16 · Upkeep, grumbling, desertion and flight

## Goal

Recruits stay only while paid and unafraid. Each day they take their upkeep from the party's inventory; unpaid days raise their discontent until they grumble and then desert, and terror makes them flee mid-fight. A deserter lives on as an individual, taking any graft with them.

## Context

- `docs/game-design.md` → "Recruits" (unmet needs mean grumbling, then desertion; terror means flight mid-fight; a grafted recruit who deserts takes the piece), "Paying them", "Grudges and memory" (deserters persist).
- The agents model in [README.md](README.md), "Recruits".
- M3-15's `RecruitDef.Upkeep`, M3-02's needs and thresholds, M3-04's `flee`, M3-08's `Individual`.

## Deliverables

- `internal/core/comp/comp.go`: `Discontent`; saved.
- `internal/core/agents/upkeep.go`: the `PerDay` upkeep; desertion; flight.
- `internal/core/api`: `Grumbled`, `Deserted`, `Fled` events and their text.
- `content/data/needs.cue`: `need.terror`, raised by triggers on the recruit templates (`combat.part_severed` on an ally or self, a member dying) through `satisfy` with a negative amount.
- `content/data`: `faction.deserters` and the content numbers.
- `internal/core/agents/upkeep_test.go`.

## Spec

**Firm:**

- Each day, for each recruit in UID order, the upkeep is taken from the party's inventory if it is all there; otherwise nothing is taken and `Discontent` rises by one. A paid day lowers it by one, to 0.
- At `recruits.grumble_at` the recruit grumbles once a day (`Grumbled`); at `recruits.desert_at` they desert: `MemberOf` removed, `Individual` added, faction set to the content deserters faction, controller `agents`, and `Deserted` emitted. They keep their body, grafts and inventory.
- **Flight:** while a recruit's `need.terror` is at or above `recruits.flee_at`, their controller is `agents` with a flee behavior, whatever the party's control mode, and the player's options for them are gone (`Fled`). When terror falls below, they return to the party's control. Fleeing does not end membership.

## Acceptance criteria

- [x] Exit criterion: a test hires a recruit, fights, then lets days pass with no upkeep in the inventory; the recruit grumbles, then deserts, and is still an entity with `Individual` after the party leaves the map.
- [x] A grafted recruit who deserts still carries the graft.
- [x] A recruit who sees an ally's arm severed and passes the terror threshold flees, and the player cannot command them until terror drops.
- [x] `go tool task check` passes.

## Out of scope

- Grudges and gratitude toward the party (M6).
- Deserters taking seats (M4's cast model).

## Notes

- Config lives in a new `recruits` block (`content.RecruitsConfig`: `grumble_at`, `desert_at`, `flee_at`, `terror_need`, `flee_behavior`, `deserters_faction`), checked only when a template is hirable. Added `comp.Discontent` and `comp.Fleeing`; save version 8.
- Flight is a marker, `comp.Fleeing{Controller}`. `explore.applyControllers` reads it before the party's mode, which is why `explore/control.go` is touched. The agents controller runs only `recruits.flee_behavior` for a body with the marker. The `agents.flight` per-entry system sets and clears it.
- Upkeep is a per-day system. A recruit who is the party leader never deserts.
- Terror triggers sit on `_#Person` and `_#Generic` and satisfy `bodies: "everyone"`, which looks from the striker. `BodyDied.Subjects()` now names the dead body as source and target so the death trigger can see. A limb worn away with no source terrifies no one: M3-16.F1. Terror eases 0.125 a phase; a severed limb adds 0.75, a death 0.25.
- `need.terror` was added to `_#Generic` starting needs.
- The exit-criterion test hires and lets days pass unpaid, without a fight; the recruit's upkeep is coal in the tests so the unpaid recruit is not also starved. Grafts are checked by a `comp.Graft` on one of its parts.
- `faction.deserters` is hostile to no one (M3-16.F2). The design docs are unchanged except `docs/content-guide.md`.
- Follow-ups filed: M3-16.F1, M3-16.F2.


