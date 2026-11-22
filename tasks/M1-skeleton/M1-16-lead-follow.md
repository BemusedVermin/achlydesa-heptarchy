---
id: M1-16
title: Lead and follow
status: todo
milestone: M1
kind: feature
depends_on: [M1-10]
discovered_by: null
size: M
---

# M1-16 · Lead and follow

## Goal

Exploring with four people feels like moving one. While no hostile is in sight, the player moves the leader and the rest trail behind in single file, which also works in one-cell soops. When the party is on alert, every member's turn comes to the player. M2 sets the alert; this task makes both modes work.

## Context

- Game design: "Combat → Fighting on the map" (lead and follow), "Party".
- M1-10 stepping and the plan view, M1-17 controllers, channels and the `PerTurn` cadence.

## Deliverables

- `internal/core/comp/comp.go`: `Trail`; `Party` gains `Alert`.
- `internal/core/api/explore.go`: `Lead`, `LeaderChanged`.
- `internal/core/explore/follow.go`, `control.go`, tests.
- `content/data/templates.cue`: the party's followers start with controller `follow`.

## Spec

- **Party state:** `Party{ Slots int; Leader api.UID; Alert bool }`. `Alert` is set by M2-07's danger check; until then it's always false.
- **Party order** is the member order from the party template. The leader is first. Each follower follows the member just before it in that order.
- **Trail:** `comp.Trail{ Last api.Cell }` on every member: the cell it left on its most recent step.
- **The `follow` controller:** if the member ahead of it is within one step, step into that member's `Trail.Last` (with `actions.Start` and `action.step`). Otherwise path toward it with `hexmap.Path` (at most 200 nodes) and take the first step. Never step into an occupied cell. A follower that can't move waits until the next entry (`wait/event`).
- **Occupancy:** one body per cell. In the `step` primitive, stepping into another party member's cell swaps the two (the swapped member's `Trail` is updated too). Stepping into any other occupied cell is refused with `text.explore.occupied`; M2 turns that into an attack option.
- **`explore.party_control`,** a `PerTurn` system (it runs after every timeline entry): when `Alert` is false, the leader's controller is `player` and the others' are `follow`; when true, all members' are `player`. Publish nothing when nothing changed.
- **One keypress, one step:** while `Alert` is false, a keypress steps the leader, its stepping channels go busy, and the followers trail on their own decisions as their legs free up. While it's true, every member's decisions come to the player, in timeline order.
- **Changing leader:** `Lead{ Party, Member api.UID }`, kind `explore.lead`, allowed only when `Alert` is false. The chosen member moves to the front of the party order. It publishes `LeaderChanged` (loggable). Options `lead/<uid>`, labelled "Lead with <name>", for every other member.
- **Start placement:** the leader spawns on the start mark. Each other member takes the nearest free standable cell in breadth-first order over `Step`, with ties by cell order.

## Acceptance criteria

- [ ] In the hall, one leader step is one keypress, and the follower steps into the cell the leader left, in the same `Apply`.
- [ ] Down the descent's one-cell soop, the follower trails single file and is never in an occupied cell.
- [ ] Stepping into the follower swaps them.
- [ ] With `Alert` forced on in a test, each member's decisions stop for the player, in timeline order.
- [ ] `explore.lead` changes the leader, the party order and the controllers; it's rejected while `Alert` is on.
- [ ] Replaying 300 random options gives the same hash as playing them.
- [ ] `go tool task check` passes.

## Out of scope

- What sets `Alert` (M2-07). Formations wider than single file: file a follow-up if open chambers make the conga line look silly.

## Notes
