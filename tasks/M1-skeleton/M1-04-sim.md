---
id: M1-04
title: "sim: world, event bus, scheduler, command registry"
status: review
milestone: M1
kind: seam
depends_on: [M1-02, M1-03]
discovered_by: null
size: M
---

# M1-04 · sim: world, event bus, scheduler, command registry

## Goal

The engine spine: one world value holding the Ark ECS world and the registries every package plugs into, so a new system, command or view section is a registration rather than an edit to a central loop.

## Context

- Tech design: "Designing for pivots" (System scheduler, Commands and events, View sections), "Core packages" (`sim`), "Simulation loop and determinism".
- Ark: user guide at https://mlange-42.github.io/ark/ and `go doc github.com/mlange-42/ark/ecs`. Use Ark for the ECS only; this task's bus is separate from Ark's own event system, which reports ECS changes.

## Deliverables

- `internal/core/sim/world.go`, `bus.go`, `scheduler.go`, `commands.go`, `views.go`
- A `_test.go` file for each.
- `go.mod`: adds `github.com/mlange-42/ark`.

## Spec

**World** (shape is firm, field types follow Ark):

```go
type World struct {
	ECS      *ecs.World    // or whatever ecs.NewWorld returns
	RNG      *rng.Streams
	Bus      *Bus
	Systems  *Scheduler
	Commands *Commands
	Views    *Views
	Codec    *api.Codec
	Now      int64         // current tick; only clock writes it
	nextUID  api.UID
}

func NewWorld(seed uint64) *World
func (w *World) NewUID() api.UID // 1, 2, 3, ...; never reused
func (w *World) NextUID() api.UID // peek, for saves
func (w *World) SetNextUID(api.UID) // for loading saves
```

**Bus:** events are queued, then delivered first in, first out.

```go
func Subscribe[E api.Event](b *Bus, name string, fn func(w *World, e E))
func (b *Bus) Publish(e api.Event)
// Drain delivers queued events, including ones published by subscribers during
// delivery, until the queue is empty. Subscribers of one kind run in the order
// they subscribed. It returns every delivered event in delivery order, and an
// error once more than MaxCascade (10,000) events are delivered in one drain.
func (b *Bus) Drain(w *World) ([]api.Event, error)
```

Events with no subscribers are still returned.

**Scheduler:**

```go
type Cadence uint8
const (
	PerTurn Cadence = iota // after every actor's turn (turns, M1-15, runs it)
	PerPhase                 // at each day-phase boundary
	PerDay                   // at each day boundary
)

type System struct {
	Name    string   // unique, "<package>.<what>", e.g. "agents.think"
	Cadence Cadence
	After   []string // systems of the same cadence that must run first
	Run     func(w *World) error
}

func (s *Scheduler) Register(sys System) error // duplicate name -> error
func (s *Scheduler) Freeze() error              // sorts each cadence; see below
func (s *Scheduler) Order(c Cadence) []string
func (s *Scheduler) Run(w *World, c Cadence) error
```

- `Freeze` topologically sorts each cadence on `After`, breaking ties by name in ascending order (Kahn's algorithm with a sorted ready set).
- It errors on an unknown name in `After`, an `After` naming a system of another cadence, or a cycle. Name the systems involved in the error.
- `Register` after `Freeze`, and `Run` before it, return errors.
- `Run` drains the bus after each system and stops at the first error.

**Commands:**

```go
// Handler applies a command. It publishes events on w.Bus and returns the
// command's time cost in ticks. An invalid player command publishes
// api.Rejected and returns 0, nil; errors are for bugs only.
type Handler[C api.Command] func(w *World, cmd C) (cost int64, err error)

func Handle[C api.Command](w *World, h Handler[C]) error // also registers C with w.Codec; duplicate -> error
func (c *Commands) Dispatch(w *World, cmd api.Command) (cost int64, err error) // unknown kind -> error
func (c *Commands) Kinds() []string // sorted
```

**Views:**

```go
type ViewBuilder struct {
	Name  string
	After []string
	Build func(w *World, v *api.View) error
}
func (v *Views) Register(b ViewBuilder) error
func (v *Views) Build(w *World) (api.View, error) // runs builders in sorted order, like the scheduler
```

Reuse the scheduler's sorting function for views; don't write it twice.

## Acceptance criteria

- [x] Scheduler: with A, B (after A), C (after A) and D (after B and C), the order is A, B, C, D, and it's identical over 100 runs.
- [x] Scheduler: a cycle, an unknown `After` name and a cross-cadence `After` each error at `Freeze`, naming the systems.
- [x] Bus: an event published inside a subscriber is delivered after the current one, and `Drain` returns all of them in order.
- [x] Bus: a subscriber that republishes forever makes `Drain` return an error instead of hanging.
- [x] Commands: a registered test command dispatches to its handler; an unknown kind errors; a duplicate `Handle` errors.
- [x] `NewUID` returns 1, 2, 3 on a fresh world.
- [x] `go tool task check` passes.

## Out of scope

- Content loading (M1-06), the clock (M1-08), and any real systems.
- Spawning entities (M1-07).

## Notes

- Added `order.go` (not in the Deliverables list): the shared `sortByAfter` used by both the scheduler and the views, as the spec asks for one sort function.
- Ark is v0.8.3; `ecs.NewWorld()` is used with its defaults.
- `Bus.Drain` clears the queue when it hits `MaxCascade`, so the world is usable after the error. Events already delivered are still returned alongside the error.
- `Scheduler.Register` also rejects a nil `Run` and an out-of-range cadence; `Views.Register` rejects a nil `Build`.
- `Views.Build` sets `View.Tick` from `w.Now`; other sections are left to builders.
- Bus subscriber names are stored for diagnostics but not yet used in any message.
- No follow-ups filed.
