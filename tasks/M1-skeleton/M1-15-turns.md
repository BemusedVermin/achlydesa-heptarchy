---
id: M1-15
title: "turns: one shared timeline for every body"
status: todo
milestone: M1
kind: seam
depends_on: [M1-04, M1-07, M1-08]
discovered_by: null
size: M
---

# M1-15 · turns: one shared timeline for every body

## Goal

Exploration and fights happen on the same turns, roguelike-style. Every acting entity, in the party or not, sits on one timeline of ticks, and the clock jumps to whoever acts next. A turn is one step plus one use of each working arm. Controllers (the player, a follower, an AI) decide what happens in it. Nothing in the engine knows whether a fight is going on.

## Context

- Game design: "Combat → Fighting on the map", "Combat → Bodies and turns" (legs decide how often a body acts).
- Tech design: "Designing for pivots" (Time), "Simulation loop and determinism → Tick order".
- M1-04 scheduler (the `PerTurn` cadence), M1-07 spawning, M1-08 `clock.Advance`.
- Ark resources, for state that belongs to the whole world: the user guide's resources chapter and `go doc github.com/mlange-42/ark/ecs Resource`.

## Deliverables

- `internal/content/types.go`: `TurnsConfig`, `TurnEnv`, `Template.Turn`; regenerate the schema.
- `content/data/turns.cue`
- `internal/core/comp/comp.go`: `Turn`.
- `internal/core/api/turns.go`: commands, events, the view section.
- `internal/core/turns/turns.go`, `controllers.go`, `install.go`, tests.

## Spec

**Content:**

```go
type TurnsConfig struct {
	TurnTicks    string `json:"turn_ticks"`    // Expr with TurnEnv: ticks from one turn to the next, before extra costs
	DefaultSpeed int    `json:"default_speed"` // for entities whose speed nothing else provides
}
type TurnEnv struct {
	Speed int `expr:"speed"`
}
type TurnDef struct {
	Controller string `json:"controller"` // "player", "follow", "idle", ...
}
```

The sample `turn_ticks` is `speed >= 2 ? 10 : (speed == 1 ? 15 : 25)`, with `default_speed` 2. The numbers are placeholders for M7.

**Component:**

```go
// Turn puts an entity on the timeline.
type Turn struct {
	Next        int64  `json:"next"`         // tick of its next turn
	Controller  string `json:"controller"`   // a registered controller name
	Stepped     bool   `json:"stepped"`      // used its step this turn
	ActionsUsed int    `json:"actions_used"` // arm uses this turn
	Extra       int64  `json:"extra"`        // ticks added to the delay after this turn (terrain, climbing, waiting)
	Open        bool   `json:"open"`         // the turn has started and not ended
}
```

**Hooks** that later packages fill. Keep them, the config and the controllers in an Ark resource owned by `turns`, so `sim` never has to know about turns:

```go
func SetSpeed(w *sim.World, fn func(w *sim.World, e ecs.Entity) int)              // default: DefaultSpeed; M2-02 sums the legs
func SetBudget(w *sim.World, fn func(w *sim.World, e ecs.Entity) (steps, actions int)) // default: 1 step, 0 actions; M2-02 counts working arms
func RegisterController(w *sim.World, name string, c Controller) error

// Controller plays a turn for anything the player doesn't control.
type Controller interface {
	TakeTurn(w *sim.World, actor ecs.Entity) error // may step and act; the loop ends the turn if the controller didn't
}
```

`player` is reserved: the loop stops and waits for commands. `idle` is built in and just ends its turn.

**The loop:**

```go
// Run continues the timeline. If the current turn is still open and player-controlled,
// it returns at once. Otherwise it repeatedly takes the entity with the smallest Next
// (ties by UID), advances the clock to that tick if it's in the future (which runs
// per-phase and per-day systems), opens its turn and publishes TurnStarted. A player
// actor stops the loop; any other runs its controller and ends its turn. It errors
// after MaxAITurns (10,000) AI turns in one call, so a world with no player can't hang.
func Run(w *sim.World, clk *clock.Clock) (current ecs.Entity, err error)
func Current(w *sim.World) (ecs.Entity, bool)

// Spend records use of the current turn's budget; over budget is an error (a bug,
// since commands are validated first).
func Spend(w *sim.World, actor ecs.Entity, step bool, actions int) error
func Remaining(w *sim.World, actor ecs.Entity) (steps, actions int)
func AddExtra(w *sim.World, actor ecs.Entity, ticks int64)

// EndTurn closes the turn: Next = Now + turn_ticks(speed) + Extra, the budget
// resets, TurnEnded is published and PerTurn systems run.
func EndTurn(w *sim.World, actor ecs.Entity) error
```

Keep a min-heap keyed by `(Next, UID)` so picking the next actor stays cheap with many entities. It must give exactly the same order as a sort.

**Spawning:** templates gain `Turn *TurnDef`. Its spawner adds `comp.Turn` with that controller and `Next = Now + turn_ticks(speed)`. `M1-07`'s party members get `turn: controller: "player"` in content for now; M1-16 switches followers to `follow`.

**api:**

- `End{ Actor api.UID }`, kind `turns.end`.
- `Wait{ Actor api.UID; Ticks int64 }`, kind `turns.wait`: adds `Ticks` to `Extra` and ends the turn. `Ticks <= 0` is rejected.
- Both are rejected unless `Actor` is the current, player-controlled entity.
- `TurnStarted{ Actor api.UID; Tick int64 }` and `TurnEnded{ Actor api.UID }`, kinds `turns.turn_started` and `turns.turn_ended`. Neither is loggable.
- `TurnsView{ Current api.UID; Upcoming []api.UID }`: the next eight actors in timeline order. It's added to `View` as `Turns *TurnsView`.

**Options** for the current player-controlled actor:

- `end`: "End turn".
- `wait`: wait until the next phase, which is `Wait` with the ticks left in the phase. Labels come from `text.turns.*`.

## Acceptance criteria

- [ ] Two idle actors with speeds 2 and 1 (10 and 15 ticks) over 60 ticks: the faster takes 6 turns and the slower 4, interleaved by `(Next, UID)`.
- [ ] `Extra` from one turn delays only the next turn.
- [ ] `Run` stops at a player actor and returns it; AI controllers run without stopping; a second `Run` with the player's turn still open returns the same actor.
- [ ] Waiting across a phase boundary runs PerPhase systems once, with `w.Now` at the boundary tick when they run.
- [ ] `turns.end` from an actor that isn't current is rejected.
- [ ] With no player actors, `Run` returns an error after `MaxAITurns`.
- [ ] Two worlds with the same seed and spawns produce the same turn order over 1,000 turns.
- [ ] `go tool task check` passes.

## Out of scope

- Stepping (M1-10) and lead and follow (M1-16).
- Arms and legs (M2-02 sets the budget and speed hooks).
- Creature behavior (M2-08, M3).

## Notes
