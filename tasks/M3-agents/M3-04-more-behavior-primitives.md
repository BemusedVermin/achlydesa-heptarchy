---
id: M3-04
title: "Behavior primitives: consume, carry, flee, guard and engage"
status: done
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

- [x] A test creature with `consume` walks to a dead body and starts the named action on it.
- [x] A test hauler with `carry` moves 2 water from a well body's inventory to a cistern body's inventory, and the counts change by exactly 2.
- [x] A test creature whose `flee` scores highest moves away from a hostile until it is out of sight.
- [x] A `guard` creature stays within its radius while no hostile comes near, and attacks one that does.
- [x] An `engage` creature fights as a `hunt` creature does: a test with the same seed gives the same landings for both.
- [x] `go tool task check` passes.

## Out of scope

- The agents chooser and dodging (M3-05).
- Working, trading and worship for people beyond `carry` (M3-11 composes `carry` into a water hauler; trade between residents is not in M3).

## Notes

- `consume` is done when the action starts; the action's landing follows, and the behavior is not kept to see it land (provisional).
- `carry` picks up and delivers at once when the body is next to the source or the owner, with no action time. A body standing on the `to` mark is the owner of the stock; a body on the map with no owner there waits.
- Primitives gained `Optional` params (`agents.Primitive.Optional`), since `engage`'s `controller` has a default. `engage` refuses to hand the decision to `agents` itself.
- `flee` and `guard` take sight from `combat.sight_radius` and `sight_depth`, and `guard` engages through `hunt`, which picks its own nearest hostile and may chase out of the radius.
- Outside the file list: `internal/core/actions/actions.go` now lets an explicit intent target a corpse (dead bodies have no `comp.Actor`), which `consume` needs; content also gained test actions, templates and behaviors.
- Params such as `count` and `radius` are read as whole numbers at run time; a bad one is an error then, not at load.
- The tests rest on the party's stationary leader; tests wait only a few hundred ticks, since a very long wait with only the party crashed the timeline before M3-03.F1.
- Follow-ups filed: M3-04.F1 (path around bodies), M3-04.F2 (corpse targeting opt-in).
