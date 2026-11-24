---
id: M3-03
title: Behaviors and the agents controller
status: done
milestone: M3
kind: seam
depends_on: [M3-02]
discovered_by: null
size: M
---

# M3-03 · Behaviors and the agents controller

## Goal

Creatures and people choose what to do by utility: each behavior a template lists is scored by its content considerations, the best one runs, and a registered behavior primitive turns it into actions on the timeline. The `agents` controller does this for any body whose template names it.

## Context

- The agents model in [README.md](README.md), "Behaviors and utility".
- `docs/go-technical-design.md` → "Designing for pivots" → "Behaviors", and the `agents` row of "Core packages".
- `internal/core/timeline/controllers.go`, `internal/core/actions` (`Startable`, `Start`, `StepRules`), `internal/core/hexmap/path.go` (`Path`), and `combat`'s `hunt` controller as the pattern for stepping toward something.

## Deliverables

- `internal/content/types.go`: `BehaviorDef`, `behaviors` on `Template`; `content.ConsiderEnv`; schemas regenerated.
- `internal/core/comp/comp.go`: `Doing`, registered and saved.
- `internal/core/agents/behaviors.go`: the primitive registry and scoring; `controller.go`: the `agents` controller; `primitives.go`: `idle`, `wander`, `move_toward`.
- `content/data/behaviors.cue` (new) with test behaviors; test templates in `content/data/templates.cue`.
- `internal/core/agents/behaviors_test.go`, `controller_test.go`.
- `internal/core/save/save.go`: bump `CurrentVersion`.

## Spec

**Firm:**

```go
// content
type BehaviorDef struct {
    Name     api.ContentID
    Do       string            // a registered behavior primitive
    Params   map[string]string // Expr or IDs, as the primitive declares
    Consider []string          // Expr floats over ConsiderEnv; the score is their product
    Target   *TargetDef        // what the behavior goes for, if anything
}
type TargetDef struct {
    Kind string                // "body", "mark" or "cell"
    Tags []string              // a body's template tags it must have
    Dead *bool                 // bodies only: dead or living
    Hostile *bool              // bodies only
    Mark string                // marks only: a mark name of the current map
    Terrain api.ContentID      // cells only
}

// agents
type Primitive struct {
    Name   string
    Params []string
    // Step does one decision's worth of the behavior: start an action, or report done.
    Step func(w *sim.World, body ecs.Entity, b content.BehaviorDef, target Target) (done bool, err error)
}
func RegisterPrimitive(w *sim.World, p Primitive) error

// comp
type Doing struct { Behavior api.ContentID; Target api.UID; Cell api.Cell; Since int64 }
```

- On `Decide`, the controller scores every behavior the template lists. The current behavior (in `Doing`) gets the content bonus `agents.keep_bonus` (hysteresis). The best wins; ties are broken by the `agents` stream. It then calls the primitive's `Step` once.
- The target is resolved when a behavior is chosen: the nearest body, mark or cell (by `hexmap.CellDistance`, then UID or cell order) on the body's map. A behavior with a target and none found scores 0.
- `ConsiderEnv` gives at least `source` (`BodyEnv`, with needs), `target` (`BodyEnv`, when the target is a body), `distance`, `phase`, `now` and `doing` (the current behavior's key).
- `move_toward` steps along `hexmap.Path` with the content step action (`combat.step_action`'s config value, read the way `hunt` reads it) and is done on arrival. `wander` steps to a random open neighbour (stream `agents`). `idle` waits for the next entry.
- No content record, tag or need is named in Go.

**Suggestion:** cache each body's resolved target in `Doing` and re-resolve only when the behavior changes or the target is gone, since resolving scans every body on the map.

## Acceptance criteria

- [x] A test creature with `wander` and `move_toward` (toward a mark) scored by an Expr over `source.needs.water` walks to the mark once its water need is high, and wanders before that.
- [x] Two runs with the same seed give the same positions after 200 ticks.
- [x] Keeping the current behavior: with `keep_bonus` set, a creature whose two behaviors score within the bonus does not switch back and forth every decision (a test counts switches).
- [x] `go test -bench . ./internal/harness` shows the existing step benchmark within 20% of master, with no agents on the map.
- [x] `go tool task check` passes.

## Out of scope

- `consume`, `carry`, `flee`, `guard`, `engage` (M3-04).
- Traveling through gates (M3-11's `leave`).
- Off-screen behavior (M3-10).

## Notes

- **Deviation:** `TargetDef.Kind` is a named type, `content.BehaviorTargetKind`, with constants `body`, `mark` and `cell`, so the schema rejects any other value (the content guide's rule for fixed sets). Everything else keeps the spec's names.
- **Content:** behaviors are a `behaviors` collection (`behavior.` IDs, named `text.<id>`); `agents` holds `keep_bonus` (0.1) and `cell_radius` (6, how far a cell target is looked for, which the spec left open). The test behaviors and `template.test.drinker` sit in the sample content; no map uses them. The tests put a `water` mark in their own copy of the hall.
- **Choices to check, all reversible:**
  - The keep bonus is added to the current behavior's score, and only when that score is above 0.
  - A behavior scoring 0 or less is never chosen, and one whose target isn't found isn't scored. With nothing left the body waits.
  - When `Step` reports done, `Doing` is removed, so a finished behavior loses its keep bonus.
  - `move_toward` arrives on the cell for a mark or a cell, and next to a body (distance 1). It waits when there is no path within 200 nodes, as `hunt` does.
  - `doing` in `ConsiderEnv` is the full behavior ID (`"behavior.test.wander"`), as `phase` is a full phase ID. `distance` is `hexmap.CellDistance`.
  - Body targets are any entity with a position and an identity on the agent's map, dead ones included; there's no sight check.
  - A body target is kept while it still matches, and a cell target while its terrain is unchanged (the spec's suggestion); a mark is looked up every decision, since that is a map lookup.
- **Files outside the Deliverables:** `internal/content/behaviors.go` (new: checks, `Behavior`, `BehaviorIDs`, `Agents`, `BehaviorScore`), `internal/content/{env,load,records,registry}.go`, `internal/core/hexmap/map.go` and `map_test.go` (`Map.Mark` and the `Marked` interface, read from the source explore already builds, so `agents` can find marks without importing `explore`), `internal/core/agents/{doc,install}.go`, `internal/core/game/game.go` (`agents.ValidateContent` runs with the other start-of-game checks: unknown primitives and params), `docs/content-guide.md` (a Behaviors section).
- **Benchmark:** `BenchmarkStep`, five runs each: master 2.97–3.29 ms/op, this branch 2.96–3.17 ms/op.
- **Follow-up:** M3-03.F1. A body that a per-phase system kills while the clock advances to its own decision crashes `timeline.Run` (nil `Actor`). The drinker hit it through `need.water`'s damage threshold, so the agents tests clear the thresholds.

