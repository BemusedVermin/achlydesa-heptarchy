---
id: M6-01
title: Hostility in the foundation, and a two-way controller hand-off
status: done
milestone: M6
kind: seam
depends_on: []
discovered_by: null
size: S
---

# M6-01 · Hostility in the foundation, and a two-way controller hand-off

## Goal

`agents` can ask whether two bodies are hostile and can hand a fight to the `hunt` controller and take the body back, without importing `combat`.

## Context

- `docs/go-technical-design.md` → "Repository layout" (feature packages never import each other), "Designing for pivots" → "Combat sides".
- Today `combat.Hostile` (`internal/core/combat/hostility.go`) and the `hunt` controller live in `combat`; `effects.SetHostile` is the hook the foundation already offers.
- `internal/core/timeline/controllers.go`: `RegisterController` only; there is no way for one controller to delegate to another.
- Combat AI stays separate from the mind (owner decision, `tasks/M6-minds/README.md` → "Decisions").

## Deliverables

- `internal/core/anatomy/`: `Hostile(w, a, b)` moved here from `combat`, with its test.
- `internal/core/timeline/controllers.go`: a hand-off, and `comp` gains the component it needs.
- `internal/core/combat/`: uses the moved function; no behavior change.
- Combat goldens under `internal/core/combat/testdata/` regenerated only if the change makes them differ, and said so in Notes.

## Spec

- **Firm:** a body has a base controller (its template's `actor.controller`) and optionally a delegate. `timeline.DelegateTo(w, body, name)` makes the named controller decide for the body until `timeline.Resume(w, body)` or the delegate calls it. A body with no delegate behaves exactly as today.
- **Firm:** the delegate is stored in a component, so a save keeps it; bump the save version.
- **Suggestion:** `comp.Delegate{Controller string}`. The `hunt` controller calls `Resume` when it finds no prey, so the mind decides again.
- No import from `agents` to `combat`; `agents` will call `timeline.DelegateTo(w, body, "hunt")`.

## Acceptance criteria

- [x] `go tool task test -- ./internal/core/timeline ./internal/core/anatomy ./internal/core/combat` passes, with a test that a delegated body is decided by the delegate and returns to its base controller on `Resume`.
- [x] A test that a delegate survives a save and load.
- [x] `go tool task check` passes.

## Out of scope

- Any mind or planner: M6-10 and M6-11.
- Making creatures flee: M6-19.

## Notes

- `combat.Hostile` is now `anatomy.Hostile`; its test moved to `anatomy_test.go`. `combat/hostility.go` is gone and its `nameOf` helper moved to `combat/view.go`.
- `timeline.DelegateTo`, `timeline.Resume` and `timeline.ControllerOf`, with `comp.Delegate{Controller}` holding the delegate. The save version is now 12 (master was at 11 when this merged); `save_test.go` gained the no-op migration step for 11, as earlier bumps did.
- Decisions the spec left open, each the easier one to change later:
  - A controller that hands the body off (DelegateTo or Resume) without acting lets the next controller decide at once, in the same decision, so a hand-off never costs a wait. At most two hand-offs per decision (`maxHandOffs`), so two controllers handing back and forth end in a wait instead of a loop.
  - Only a registered controller can be a delegate; `player` cannot. A delegate overrides a player base too, so a player body with a delegate is decided by the delegate.
  - `explore`'s control-mode switch still changes the base controller only and leaves a delegate alone.
  - The `hunt` controller calls `Resume` when it sees no prey. On a body whose own controller is `hunt`, that does nothing, so creatures behave as before.
- Combat goldens are unchanged; no regeneration was needed.
- Files outside the Deliverables: `internal/core/combat/hunt_test.go` (a delegated hunter hands the body back), `internal/core/save/save.go` and `save_test.go` (the version bump), `internal/core/comp/comp_go_gen.cue` (regenerated), `docs/go-technical-design.md` (the timeline row names the hand-off).
- No follow-ups filed.
