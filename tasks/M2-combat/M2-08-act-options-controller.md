---
id: M2-08
title: The act command, options and the creature controller
status: todo
milestone: M2
kind: feature
depends_on: [M2-07]
discovered_by: null
size: M
---

# M2-08 · The act command, options and the creature controller

## Goal

On a member's turn, the core offers every legal move as an option, grouped for menus. Hostile creatures take their own turns through a controller that closes distance and attacks using the same validity check, scored by a content formula. The frontend never works out what's legal, and fleeing is just walking away from something slower.

## Context

- Tech design: "Designing for pivots" (Choices come from the core).
- Game design: "Combat → Grafts" (grafts quietly shape what the player is offered: keep option ordering in one function so a later rule can change it).
- M1-15 controllers and budgets, M2-06 `Validate` and `Resolve`, M2-07 hostility and alert.

## Deliverables

- `internal/core/api/combat.go`: `Act` and the `BodiesView` section.
- `internal/content/types.go`: `CombatConfig.AIScore` (Expr over `ChoiceEnv`); regenerate the schema.
- `internal/core/combat/act.go`, `options.go`, `controller.go`, `view.go`, tests and a golden file.

## Spec

**Command:** `Act{ Actor, Part api.UID; Form, Type api.ContentID; Target, TargetPart api.UID }`, kind `combat.act`.

- It's rejected unless the actor is current and player-controlled and `Validate` passes.
- On success it resolves the move and spends one action (`turns.Spend`).
- When the actor has no step and no actions left, its turn ends automatically.

**Options,** while the current actor is a player-controlled member:

- One option per valid move, with IDs `act/<form>/<type>/<target uid>/<part uid>`. Leave out segments that don't apply, so Stand is `act/stand`.
- Labels are rendered text such as "Strike · Cut · Clay-thing · left arm". A part whose material the party doesn't know shows as "?" (via `knowledge.Best`).
- Order: form, then type, then target UID, then part UID, through one function, `orderOptions`.
- Cap the list at 500 options and file a follow-up if a sample encounter exceeds it.
- M1-10's `step.*` and M1-15's `end` and `wait` options stay alongside these.

**`BodiesView`,** added to `View` as `Bodies *BodiesView`:

- `Current`: the current actor's UID, with its `StepsLeft` and `ActionsLeft`.
- `Bodies`: every visible body, plus all party members. Each has its UID, name, faction name, whether it's hostile to the party, statuses, and its parts. Each part has its UID, name, slot, integrity and max, whether it's locked, whether it's a graft, and its material, or "?" when the party doesn't know it. Members' own materials are always known.

**The `hunt` controller,** registered with `turns`:

1. Pick the nearest visible hostile, by hex distance plus depth difference, then UID.
2. While actions remain and some valid move reaches it, choose a move with the chooser and resolve it.
3. If none reaches, take one step along `hexmap.Path` toward it (at most 200 nodes), then try step 2 again.
4. With no hostile in sight, end the turn.

A body without a functioning targeting part picks its target at random among valid moves (stream `combat`).

**Chooser:**

```go
type Chooser interface {
	Choose(w *sim.World, actor ecs.Entity, moves []Move) (Move, bool) // false = no move
}
// ScoreChooser scores each move with content's ai_score and takes the best, breaking ties with the combat stream.
type ScoreChooser struct{}
```

`ChoiceEnv` exposes at least `weak`, `immune_known`, `target_is_party`, `consumes_item`, `target_integrity` and `height`. Choosers are registered by name, so `agents` can supply a smarter one in M3.

## Acceptance criteria

- [ ] On alert, the current member's options list only legal moves. An `Act` built from an option always succeeds; a hand-built illegal `Act` (out of reach, arm already used, not the actor's turn) is rejected with a reason.
- [ ] Unknown materials show "?" until a Sweep reveals them.
- [ ] A hunting creature closes in over several turns and attacks when in reach.
- [ ] A creature with its legs severed falls behind a party walking away, and the alert eventually ends.
- [ ] Two creatures of mutually hostile factions fight each other with the party out of sight (no party involvement in the test).
- [ ] A scripted six-turn encounter's event log matches a golden file.
- [ ] `go tool task check` passes.

## Out of scope

- Drawing any of this (M2-10). Needs-driven behavior (M3).

## Notes
