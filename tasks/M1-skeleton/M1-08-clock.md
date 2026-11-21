---
id: M1-08
title: "clock: ticks, phases and days"
status: todo
milestone: M1
kind: feature
depends_on: [M1-04, M1-06]
discovered_by: null
size: S
---

# M1-08 · clock: ticks, phases and days

## Goal

Time is one tick counter. Phases per day and ticks per phase come from content, the turn timeline (M1-15) moves the counter, and per-phase and per-day systems run at the right boundaries.

## Context

- Game design: "Core loop and time management" (Days pass whether or not they act).
- Tech design: "Designing for pivots" (Time, System scheduler), "Simulation loop and determinism → Tick order".
- M1-04 `sim.Scheduler` cadences.

## Deliverables

- `internal/content/types.go`: add `Calendar`; regenerate the schema.
- `content/data/calendar.cue`
- `internal/core/api/clock.go`: the clock's events and view section.
- `internal/core/clock/clock.go`, `install.go`, `clock_test.go`

## Spec

**Content:**

```go
type Calendar struct {
	TicksPerPhase int64   `json:"ticks_per_phase"`
	Phases        []Phase `json:"phases"` // in order; at least one
}
type Phase struct {
	ID   api.ContentID `json:"id"`   // "phase.dawn"
	Name api.ContentID `json:"name"` // a text ID
}
```

Sample: four phases (dawn, day, dusk, night), 60 ticks each. These numbers are placeholders for tuning.

**api:**

- `PhaseStarted{ Day int; Phase api.ContentID }`, kind `clock.phase_started`, loggable.
- `DayStarted{ Day int }`, kind `clock.day_started`, loggable.
- `ClockView{ Day int; Phase string; TickInPhase, TicksPerPhase int64 }`, added to `View` as `Clock *ClockView`.

Days count from 1 for display.

**clock:**

```go
type Clock struct{ /* calendar */ }
func New(cal content.Calendar) *Clock
func (c *Clock) Day(now int64) int
func (c *Clock) Phase(now int64) int

// Advance moves time forward by ticks. For each phase boundary crossed, in
// order, it sets w.Now to the boundary tick, publishes PhaseStarted and runs
// PerPhase systems; when that boundary also starts a day, it then publishes
// DayStarted and runs PerDay systems. Finally it sets w.Now to the end tick.
// PerTurn systems are run by turns (M1-15), not here.
func (c *Clock) Advance(w *sim.World, ticks int64) error

func Install(w *sim.World, c *Clock) error // registers the Clock view builder
```

- Add `text.clock.*` entries for the log lines.
- Waiting is a turn command now: `turns.wait` in M1-15.

## Acceptance criteria

- [ ] Advancing 59 ticks from 0 crosses nothing. Advancing 1 more crosses one phase and runs PerPhase systems once.
- [ ] Advancing 240 ticks from 0 crosses four phases and one day, with PerDay systems running once, after the fourth PerPhase run.
- [ ] Advancing 1,000 ticks at once runs exactly the boundaries that 1,000 single-tick advances run, in the same order (recorded with test systems).
- [ ] The Clock view shows day 1 and dawn at tick 0, and day 2 and dawn at tick 240.
- [ ] `go tool task check` passes.

## Out of scope

- Seats, exports and the flock's schedules (later milestones register PerDay systems).

## Notes
