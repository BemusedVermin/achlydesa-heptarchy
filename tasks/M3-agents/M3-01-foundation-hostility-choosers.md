---
id: M3-01
title: Hostility, choosers and controller lookup in the foundation
status: in-progress
milestone: M3
kind: seam
depends_on: []
discovered_by: null
size: M
---

# M3-01 · Hostility, choosers and controller lookup in the foundation

## Goal

The `agents` feature package can ask whether two bodies are hostile, use a named attack chooser, and hand a decision to another registered controller, all without importing `combat`. Today all three live in `combat`, which `agents` may not import (AGENTS.md rule 2).

## Context

- The agents model in [README.md](README.md), "Layers".
- `docs/go-technical-design.md` → "Repository layout" (the two layers) and "Designing for pivots" → "Behaviors" and "Combat sides".
- `internal/core/combat/hostility.go` (`Hostile`), `internal/core/combat/controller.go` (`Chooser`, `RegisterChooser`, `ScoreChooser`, `scoredChooser`, the `hunt` controller), `internal/core/combat/graft.go` (graft autonomy uses the scored chooser).
- `internal/core/timeline/controllers.go` (`RegisterController`, the controller registry).

## Deliverables

- `internal/core/anatomy/hostility.go` (new): `Hostile` and the faction lookup it needs, moved from `combat`.
- `internal/core/actions/choosers.go` (new): the `Chooser` interface and its registry, moved from `combat`.
- `internal/core/timeline/controllers.go`: a lookup by name.
- `internal/content/types.go`: `ActorDef` gains `Chooser`; schemas regenerated (`go tool task schema`).
- `internal/core/combat/*.go`: use the moved pieces; `ScoreChooser` stays in `combat` and registers itself under `score` as now.
- Tests for each moved piece where they live now; existing combat tests keep passing unchanged.

## Spec

**Firm:**

```go
// anatomy
func Hostile(w *sim.World, a, b ecs.Entity) bool

// actions
type Chooser interface {
    Choose(w *sim.World, actor ecs.Entity, intents []Intent) (Intent, bool)
}
func RegisterChooser(w *sim.World, name string, c Chooser) error
func ChooserFor(w *sim.World, body ecs.Entity) (Chooser, bool) // the template's actor.chooser, else "score"

// timeline
func ControllerNamed(w *sim.World, name string) (Controller, bool)
```

- `actor.chooser` in a template names the chooser the body's attacks are picked with; empty means `score`. The `hunt` controller and graft autonomy call `ChooserFor` instead of the fixed name.
- `ControllerNamed` never returns the reserved `player` controller.
- Behavior stays the same: `go tool task same-behavior` shows no change on the default seeds.

**Suggestion:** keep `combat.Hostile` as a one-line forward for one task if moving every caller at once makes the diff hard to review, and remove it before this task is done.

## Acceptance criteria

- [ ] `Hostile`, the chooser registry and the controller lookup are reachable from a foundation package: `grep -rn "core/combat" internal/core/anatomy internal/core/actions internal/core/timeline` finds nothing, and `combat` no longer defines `Hostile` or `RegisterChooser`.
- [ ] A test registers a second chooser, names it in a test template's `actor.chooser`, and the hunt controller picks through it.
- [ ] A test registers a controller that delegates to `ControllerNamed(w, "hunt")` and the body attacks as a hunter does.
- [ ] `go tool task same-behavior` reports no behavior change.
- [ ] `go tool task check` passes.

## Out of scope

- The agents chooser itself (M3-05).
- Any new hostility rule, such as sparing believers (M3-18 does it with factions).

## Notes


