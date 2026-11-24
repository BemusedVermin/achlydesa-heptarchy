---
id: M3-04
title: "Behavior primitives: consume, carry, flee, guard and engage"
status: todo
milestone: M3
kind: feature
depends_on: [M3-01, M3-03]
discovered_by: null
size: M
---

# M3-04 · Behavior primitives: consume, carry, flee, guard and engage

## Goal

The rest of the behavior vocabulary the design names exists as registered primitives, so content can make creatures eat, haul, run, stand guard and fight.

## Context

- The agents model in [README.md](README.md), "Behaviors and utility".
- `docs/game-design.md` → "Monsters with needs" (move toward, consume, carry, guard, flee) and "People with needs" (work, flee).
- M3-01's `anatomy.Hostile`, `actions.ChooserFor` and `timeline.ControllerNamed`; M3-02's `satisfy`; `internal/core/inventory` (`Add`, `Take`, `Count`).

## Deliverables

- `internal/core/agents/primitives.go`: `consume`, `carry`, `flee`, `guard`, `engage`.
- `content/data/behaviors.cue`: a test behavior for each; test templates.
- `internal/core/agents/primitives_test.go`.

## Spec

**Firm**, by primitive:

- **`consume`**: go to the target (a body, or a cell or mark holding an inventory owner), then start the content action named in `params.action` on it. Done when it lands. What is eaten and which need it meets are the action's effects (`satisfy`, item use), so the primitive names no need.
- **`carry`**: take `params.count` of `params.item` from the target's inventory into the body's own, walk to `params.to` (a mark), and put it into the inventory of whatever there owns one (a town's stock arrives in M3-08; tests use a body standing on the mark). Done when delivered.
- **`flee`**: step to the open neighbour that most increases the hex distance from the nearest hostile in sight (`anatomy.Hostile`); done when none is in sight.
- **`guard`**: stay within `params.radius` of the target mark or body; when a hostile comes within that radius, behave as `engage` toward it.
- **`engage`**: hand the decision to the controller named in `params.controller` (default `hunt`) through `timeline.ControllerNamed`. That controller picks attacks with `actions.ChooserFor`, so the template's chooser applies.

Each new primitive is registered in `agents` with the parameters it declares; a behavior with a missing or unknown parameter fails at load.

## Acceptance criteria

- [ ] A test creature with `consume` walks to a dead body and starts the named action on it.
- [ ] A test hauler with `carry` moves 2 water from a well body's inventory to a cistern body's inventory, and the counts change by exactly 2.
- [ ] A test creature whose `flee` scores highest moves away from a hostile until it is out of sight.
- [ ] A `guard` creature stays within its radius while no hostile comes near, and attacks one that does.
- [ ] An `engage` creature fights as a `hunt` creature does: a test with the same seed gives the same landings for both.
- [ ] `go tool task check` passes.

## Out of scope

- The agents chooser and dodging (M3-05).
- Working, trading and worship for people beyond `carry` (M3-11 composes `carry` into a water hauler; trade between residents is not in M3).

## Notes


