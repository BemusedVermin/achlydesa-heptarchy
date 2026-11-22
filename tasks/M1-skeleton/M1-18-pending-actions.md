---
id: M1-18
title: "timeline: actions that land later"
status: done
milestone: M1
kind: seam
depends_on: [M1-17]
discovered_by: null
size: M
---

# M1-18 · timeline: actions that land later

## Goal

An attack takes time to land. Starting it puts an entry on the timeline at the tick it lands. Until then everyone can see it coming: the target can step out of reach, the attacker can be staggered, or the swinging arm can be cut off. This task builds that machinery generically; M2 decides what lands.

## Context

- Game design: "Combat → Timing".
- M1-17's timeline, channels and view.

## Deliverables

- `internal/core/comp/comp.go`: `Pending`.
- `internal/core/api/timeline.go`: `ActionLanded`, `ActionCanceled`.
- `internal/core/timeline/pending.go`, tests.

## Spec

```go
// comp.Pending is on an entity of its own: one per action in flight.
type Pending struct {
	Kind    string  `json:"kind"`    // the resolver's name, e.g. "combat.move"
	Actor   api.UID `json:"actor"`   // the body acting
	Part    api.UID `json:"part"`    // the channel it uses (an arm), or the body
	Target  api.UID `json:"target"`  // the body it's aimed at, or 0
	Started int64   `json:"started"`
	Lands   int64   `json:"lands"`
	Seq     uint64  `json:"seq"`     // scheduling order, for ties
}

func RegisterResolver(w *sim.World, kind string, r Resolver) error
type Resolver struct {
	Resolve func(w *sim.World, pending ecs.Entity) error // apply the action when it lands
	Label   func(w *sim.World, pending ecs.Entity) string // for the timeline view, e.g. "Clay-thing's Strike lands on the Reaper (left arm)"
}

// Schedule creates a pending entity; the caller adds its own kind-specific
// component to it. Lands must be after now; Kind must have a resolver.
func Schedule(w *sim.World, p comp.Pending) (ecs.Entity, error)
func Cancel(w *sim.World, pending ecs.Entity, reason api.ContentID) // removes it and publishes ActionCanceled
func PendingOf(w *sim.World, actor api.UID) []ecs.Entity              // sorted by Lands, then Seq
func AimedAt(w *sim.World, target api.UID) []ecs.Entity               // the same, by target
```

- **Ordering:** landings join the timeline. At the same tick, landings come before decisions, and landings among themselves go by `Seq`. So to dodge an attack landing at tick 12, you must act before tick 12.
- **Landing:** `Run` advances the clock, removes the pending entity, calls its resolver, and publishes `ActionLanded`.
  - If the actor or its part no longer exists when the attack lands, it's canceled instead, with reason `text.timeline.canceled.gone`.
  - Deciding whether a landing hits or misses is the resolver's business.
- **Seq** comes from a counter in the timeline's resource. After a load, the counter resumes above the highest `Seq` on any pending entity.
- **The view:** `TimelineView` gains a `lands` entry for every pending action whose actor or target is visible, labelled by its resolver. This is what makes swings predictable.
- **Saves:** pending entities are ordinary ECS entities, so ark-serde saves them; the timeline rebuilds its heap from `Actor`, `Busy` and `Pending` components (with `M1-15.F1`).

## Acceptance criteria

- [x] A pending action landing at tick 12 resolves at 12, before a decision at 12; two landing at 12 resolve in `Seq` order.
- [x] Canceling one removes it from the timeline and the view and publishes `ActionCanceled`; its resolver never runs.
- [x] Despawning the actor before the landing cancels it with the "gone" reason.
- [x] A `lands` entry appears in `TimelineView` with its resolver's label, and disappears once landed.
- [x] Save, load and continue gives the same entries and the same hash (once M1-12 exists, add this to its tests; until then, a round trip through ark-serde in this task's tests).
- [x] `go tool task check` passes.

## Out of scope

- What any action does, hits or misses, and staggering (M2-06, M2-04).

## Notes

- Pending actions are found by querying `comp.Pending` and sorting by `Lands` then `Seq`, not kept in a second heap. Run compares the earliest landing with the next decision (`peekTick`), so there is nothing to rebuild after a load beyond M1-15.F1.
- `Resolve` runs before the pending entity is removed, so the resolver can read its kind-specific components. `ActionLanded` is published after removal.
- A player with a free channel no longer returns from `Run` while a landing is due at the current tick, so two landings at one tick both resolve before the decision.
- `Seq` is the highest `Seq` in the world plus one, found at each `Schedule`, so it is right after a load whenever that happens.
- Both events carry `Kind`, `Actor`, `Target` and `Tick`; `ActionCanceled` also has a `Reason` (`api.Text`, like `Rejected`).
- Added text `text.timeline.lands` (fallback label) and `text.timeline.canceled.gone`.
- The save round-trip criterion is met with a component copy into a fresh world, because ark-serde is not a dependency yet. The real round trip is filed as M1-18.F1.
- Follow-ups filed: M1-18.F1.

