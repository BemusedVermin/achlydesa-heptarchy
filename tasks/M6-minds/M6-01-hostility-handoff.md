---
id: M6-01
title: Hostility in the foundation, and a two-way controller hand-off
status: todo
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

- [ ] `go tool task test -- ./internal/core/timeline ./internal/core/anatomy ./internal/core/combat` passes, with a test that a delegated body is decided by the delegate and returns to its base controller on `Resume`.
- [ ] A test that a delegate survives a save and load.
- [ ] `go tool task check` passes.

## Out of scope

- Any mind or planner: M6-10 and M6-11.
- Making creatures flee: M6-19.

## Notes

