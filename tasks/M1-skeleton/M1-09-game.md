---
id: M1-09
title: "game: the public API"
status: todo
milestone: M1
kind: seam
depends_on: [M1-15]
discovered_by: null
size: M
---

# M1-09 · game: the public API

## Goal

`core/game` is the one door into the engine: build a game from a seed and content, apply commands, read the view, replay a log. It holds no rules of its own; it wires packages together through their `Install` functions.

## Context

- Tech design: "Repository layout" (One API boundary), "Core packages" (`game`), "Simulation loop and determinism → Saves and replays", "Designing for pivots" (the pivot test).
- M1-04 `sim`, M1-06 `content`, M1-08 `clock`.

## Deliverables

- `internal/core/game/game.go`, `install.go`, `log.go`, `game_test.go`

## Spec

```go
type Config struct {
	Seed    uint64
	Content *content.Registry
}

func New(cfg Config) (*Game, error)
func (g *Game) Apply(cmd api.Command) ([]api.Event, error)
func (g *Game) View() (api.View, error)
func (g *Game) Hash() ([32]byte, error)
func (g *Game) History() []api.Command // every command applied, in order
func (g *Game) Codec() *api.Codec
func Replay(cfg Config, cmds []api.Command) (*Game, error)
```

**`New`:**

1. Creates a `sim.World` with the seed and sets `Content`.
2. Calls each package's `Install` in the fixed order listed in `install.go`. That list is the only place packages are named; adding a feature package adds one line there.
3. Freezes the scheduler and the view builders.
4. Drains the bus once, so setup events reach the log.

For now the list is `clock` and `turns`. M1-10 adds `explore`.

**`Apply`:**

1. Dispatches the command and gets its cost. Commands act on the current actor's turn (M1-15); a command for anyone else is rejected by its handler.
2. If the command wasn't rejected, adds the cost to the current actor's `Extra` with `turns.AddExtra`.
3. Calls `turns.Run`, which returns at once if the player's turn is still open, or plays the timeline forward (other actors' turns, clock boundaries) until a player-controlled actor's turn begins.
4. Drains the bus and appends the command to the history.
5. Returns the events in delivery order.

`New` also calls `turns.Run` once at the end, so a new game opens on the first player turn. `game.Current()` returns the current actor's UID.

Errors from handlers or systems are returned and leave the game unusable; say so in the doc comment.

**Log:** `game` subscribes to every event (add `sim.SubscribeAll` if needed) and, for each `api.Loggable` event, appends `LogLine{Tick, Render(text)}` to a ring of the last 200 lines. A view builder named `game.log` copies the ring into `View.Log`; another, `game.tick`, sets `View.Tick`.

**`Hash`:** SHA-256 of `json.Marshal(view)`. M1-12 replaces this with the save bytes; keep the function so callers don't change.

## Acceptance criteria

Until M1-10 adds the start scenario, tests use a test-only package in the install list that spawns one player-controlled actor.

- [ ] `New` gives a view at day 1, dawn, with the test actor current.
- [ ] `Apply(turns.Wait{Actor, Ticks: 60})` returns a `PhaseStarted` event, the view's log shows its rendered line, and the actor is current again afterwards.
- [ ] `Replay` of a 50-command history (random waits from a test stream) gives the same `Hash` as the original game.
- [ ] Two games with the same seed and commands give the same hash; a different seed doesn't need to differ yet.
- [ ] An unknown command kind returns an error.
- [ ] `go tool task check` passes.

## Out of scope

- Maps, the party and movement (M1-10).
- Saving and loading (M1-12).
- The frontend still uses `internal/prototype` until M1-11.

## Notes
