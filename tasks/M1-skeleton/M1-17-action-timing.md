---
id: M1-17
title: "timeline: part-level action timing"
status: todo
milestone: M1
kind: seam
depends_on: [M1-15]
discovered_by: null
size: L
---

# M1-17 · timeline: part-level action timing

## Goal

There are no turns any more, only actions that take time. Every part that can act has its own clock: a step keeps the legs busy, a swing keeps one arm busy, and a body gets to decide whenever one of its usable parts is free. Faster legs mean quicker steps; more arms mean more swings in the air at once. This task reshapes M1-15's timeline into that model and renames the package to match. M1-18 adds attacks that land later.

## Context

- Game design: "Combat → Fighting on the map", "Combat → Timing".
- Tech design: "Designing for pivots" (Time), "Simulation loop and determinism → Tick order".
- `internal/core/turns/`: M1-15's code. Keep the heap, the controllers, the Ark resource for state, and the clock handling; replace the turn budget.
- Follow-up `M1-15.F1` (rebuild the timeline after a load) applies to the renamed package too.

## Deliverables

- Rename `internal/core/turns` to `internal/core/timeline` (`gopls rename` or by hand), and every import and reference with it.
- `internal/core/timeline/`: the model below, with tests rewritten to match.
- `internal/core/comp/comp.go`: `Turn` becomes `Actor`; add `Busy`.
- `internal/core/api/`: `turns.go` becomes `timeline.go`, with the commands, events and view below.
- `internal/content/types.go` and `content/data/turns.cue` (renamed `timeline.cue`): the config below; regenerate the schema.
- `AGENTS.md`, `.golangci.yml` (if it names the package) and `docs/go-technical-design.md`: `turns` becomes `timeline`.

## Spec

**Channels.** A channel is something that can be busy: the body itself, or one of its parts. Actions (M1-20) say which channels they occupy by tag, so a move can need one arm, two arms, or legs and an arm, and an effect can lock one channel or all of them.

```go
type Channel struct {
	Entity ecs.Entity // the body, or one of its parts
	Tags   []string   // "body" for the body itself; M2-02 gives parts their slot and grants, e.g. "arm", "grant:action"
}

// comp.Busy marks a channel busy until a tick. No component means free.
type Busy struct{ Until int64 `json:"until"` }

// comp.Actor puts a body on the timeline.
type Actor struct {
	Controller string `json:"controller"` // "player", "follow", "idle", ...
	WaitUntil  int64  `json:"wait_until"` // no decision before this tick
}
```

**Hooks,** set by later packages; keep them in the timeline's Ark resource:

```go
func SetChannels(w *sim.World, fn func(w *sim.World, body ecs.Entity) []Channel) // default: one channel, the body itself, tagged "body"; M2-02 adds one per attached part
func SetUsable(w *sim.World, fn func(w *sim.World, body ecs.Entity, c Channel) bool) // default: every channel is usable; M1-20 marks a channel usable when some action could start with it
func SetSpeed(w *sim.World, fn func(w *sim.World, body ecs.Entity) int)             // unchanged from M1-15
func SetVisible(w *sim.World, fn func(w *sim.World, body ecs.Entity) bool)           // which bodies the player may see on the timeline; default: party members only
```

**When a body decides.** Its next decision tick is the later of `WaitUntil` and the earliest `Busy.Until` among its usable channels (now, if one is free). If none is usable, it uses all its channels, so a body that can do nothing still gets to wait. The timeline orders bodies by `(decision tick, UID)`. M1-18 adds landings to the same order.

**Acting:**

```go
// Occupy makes a channel busy for ticks from now and reschedules its body.
func Occupy(w *sim.World, c Channel, ticks int64) error
func Free(w *sim.World, c Channel) bool
func Extend(w *sim.World, c Channel, ticks int64) error // adds to a channel's busy time, counting from now if it's free
func Channels(w *sim.World, body ecs.Entity) []Channel // sorted: the body first, then parts by UID
```

**The loop:**

```go
// Run plays the timeline forward. While the current body is player-controlled and
// still has a free usable channel at this tick, it returns at once: the player can act
// again, for example with a second arm. Otherwise it takes the earliest entry, advances
// the clock to its tick (running per-phase and per-day systems), publishes Ready for a
// decision and either stops (player) or runs the body's controller. A controller must
// occupy a channel or wait; if it does neither, the loop makes it wait until the next
// entry. PerTurn systems run after every entry. MaxAITurns still guards against hangs.
func Run(w *sim.World, clk *clock.Clock) (current ecs.Entity, err error)
```

- Keep `sim.PerTurn`'s name, and update its doc comment: it now runs after every timeline entry.
- `EndTurn`, `Spend`, `Remaining`, `AddExtra`, the budget hook and `Turn.Stepped`, `ActionsUsed` and `Extra` all go.

**Content:** `TurnsConfig` becomes `TimelineConfig`:

- `StepTicks string`: Expr over `SpeedEnv{speed}`, the base ticks a step keeps its channels busy (the step action in M1-20 and M1-10 reads it); it replaces `turn_ticks`, same placeholder numbers.
- `DefaultSpeed int`.
- `WindowTicks int`: how far ahead the timeline view looks; start at 30.
- `Template.Turn` becomes `Template.Actor`. Explore adds terrain and climbing on top of `StepTicks` (M1-10).

**api:**

- `Wait{ Actor api.UID; Until string; Ticks int64 }`, kind `timeline.wait`.
  - `Until` is `"ticks"` (with `Ticks > 0`), `"event"` (the next entry on the timeline that isn't this body's) or `"phase"` (the next phase boundary).
  - It sets `WaitUntil`. Rejected unless `Actor` is the current player body.
- `End` is retired: there's no turn left to end.
- `Ready{ Actor api.UID; Tick int64 }`, kind `timeline.ready`, replaces `TurnStarted`; `TurnEnded` goes.
- The view section replaces `TurnsView`. It lists every decision of a visible body from `Now` to `Now + WindowTicks`, in timeline order. M1-18 adds landings to it.

```go
type TimelineView struct {
	Now     int64
	Current api.UID // the body deciding now, or 0
	Entries []TimelineEntry
}
type TimelineEntry struct {
	Tick   int64
	Kind   string // "ready" now; "lands" from M1-18
	Actor  api.UID
	Glyph  string
	Color  string
	Target api.UID // for "lands"
	Label  string  // rendered, e.g. "The Reaper is ready"
}
```

**Options** for the current player body: `wait/event`, labelled "Wait until something happens", and `wait/phase`, labelled "Wait until <next phase>".

## Acceptance criteria

- [ ] Occupying a body's only channel for 6 ticks puts its next decision 6 ticks later; `Extend` by 2 more makes it 8.
- [ ] With two test channels, occupying one leaves the body deciding again at the same tick with the other, and `Run` returns the same body.
- [ ] A channel that the usable hook rejects doesn't make the body decide when it frees.
- [ ] `wait/event` jumps to the next entry belonging to someone else; `wait/phase` to the next phase boundary, running PerPhase systems once.
- [ ] A controller that neither occupies nor waits is made to wait, and the loop doesn't hang.
- [ ] `TimelineView` lists only visible bodies, within `WindowTicks`, in `(tick, UID)` order.
- [ ] Two worlds with the same seed and actions produce the same 1,000 entries.
- [ ] No package named `turns` remains.
- [ ] `go tool task check` passes.

## Out of scope

- Attacks that land later and canceling them (M1-18).
- Stepping (M1-10) and lead and follow (M1-16).
- Parts as channels (M2-02). Deciding what occupies which channels: actions (M1-20).

## Notes

This is L because the rename touches every caller. If it runs long, land the rename alone first and say so in Notes.
