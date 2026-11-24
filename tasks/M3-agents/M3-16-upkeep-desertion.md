---
id: M3-16
title: Upkeep, grumbling, desertion and flight
status: todo
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

- [ ] Exit criterion: a test hires a recruit, fights, then lets days pass with no upkeep in the inventory; the recruit grumbles, then deserts, and is still an entity with `Individual` after the party leaves the map.
- [ ] A grafted recruit who deserts still carries the graft.
- [ ] A recruit who sees an ally's arm severed and passes the terror threshold flees, and the player cannot command them until terror drops.
- [ ] `go tool task check` passes.

## Out of scope

- Grudges and gratitude toward the party (M6).
- Deserters taking seats (M4's cast model).

## Notes


