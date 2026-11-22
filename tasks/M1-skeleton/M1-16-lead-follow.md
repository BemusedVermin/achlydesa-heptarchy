---
id: M1-16
title: Lead and follow, and switching control
status: done
milestone: M1
kind: feature
depends_on: [M1-10]
discovered_by: null
size: M
---

# M1-16 · Lead and follow, and switching control

## Goal

Exploring with four people feels like moving one. In follow mode, the player moves the leader and the rest trail behind in single file, which also works in one-cell soops. In individual mode, every member's decisions come to the player. The player can switch between the two at any time. An alert also switches the party to individual control automatically, and back when it ends, unless the player chose the mode. M2 sets the alert; this task makes both modes and the switching work.

## Context

- Game design: "Combat → Fighting on the map" (lead and follow), "Party".
- M1-10 stepping and the plan view, M1-17 controllers, channels and the `PerTurn` cadence.

## Deliverables

- `internal/core/comp/comp.go`: `Trail`; `Party` gains `Alert`.
- `internal/core/api/explore.go`: `Lead`, `LeaderChanged`.
- `internal/core/explore/follow.go`, `control.go`, tests.
- `content/data/templates.cue`: the party's followers start with controller `follow`.

## Spec

- **Party state:** `Party{ Slots int; Leader api.UID; Alert bool; Control string; ControlBy string; AlertSeen bool }`.
  - `Control` is `follow` or `individual`, and starts as `follow`.
  - `ControlBy` records who set it last: `player` or `alert`.
  - `Alert` is set by M2-07's danger check; until then it's always false.
  - `AlertSeen` is the last `Alert` value `party_control` acted on.
- **Party order** is the member order from the party template. The leader is first. Each follower follows the member just before it in that order.
- **Trail:** `comp.Trail{ Last api.Cell }` on every member: the cell it left on its most recent step.
- **The `follow` controller:** if the member ahead of it is within one step, step into that member's `Trail.Last` (with `actions.Start` and `action.step`). Otherwise path toward it with `hexmap.Path` (at most 200 nodes) and take the first step. Never step into an occupied cell. A follower that can't move waits until the next entry (`wait/event`).
- **Occupancy:** one body per cell. In the `step` primitive, stepping into another party member's cell swaps the two (the swapped member's `Trail` is updated too). Stepping into any other occupied cell is refused with `text.explore.occupied`; M2 turns that into an attack option.
- **Switching control by hand:** `Control{ Party api.UID; Mode string }`, kind `explore.control`, sets `Control` to `follow` or `individual` and `ControlBy` to `player`. It's allowed at any time, alert or not, so a player can pull the party back into single file to retreat. It publishes `ControlChanged` (loggable). While a party member is deciding, one option is offered: `control/individual` ("Control each member") in follow mode, or `control/follow` ("Lead and follow") in individual mode.
- **Switching control by alert:** content `party.individual_on_alert` (a bool, default true) turns the automatic switch on or off.
  - When `Alert` turns on and `Control` is `follow`: switch to `individual`, with `ControlBy` set to `alert`.
  - When `Alert` turns off and `ControlBy` is `alert`: switch back to `follow`.
  - A mode the player chose is never overridden when an alert ends. While an alert is on, the player can still switch back to follow, and it stays that way until the next alert starts.
- **`explore.party_control`,** a `PerTurn` system (it runs after every timeline entry). It applies the alert rules above whenever `Alert` differs from `AlertSeen`, then sets controllers from `Control`. In follow mode the leader's controller is `player` and the others' are `follow`; in individual mode all members' are `player`. Publish nothing when nothing changed.
- **One keypress, one step:** in follow mode, a keypress steps the leader, its stepping channels go busy, and the followers trail on their own decisions as their legs free up. In individual mode, every member's decisions come to the player, in timeline order.
- **Changing leader:** `Lead{ Party, Member api.UID }`, kind `explore.lead`, allowed at any time; it matters in follow mode. The chosen member moves to the front of the party order. It publishes `LeaderChanged` (loggable). Options `lead/<uid>`, labelled "Lead with <name>", for every other member.
- **Start placement:** the leader spawns on the start mark. Each other member takes the nearest free standable cell in breadth-first order over `Step`, with ties by cell order.

## Acceptance criteria

- [x] In the hall, one leader step is one keypress, and the follower steps into the cell the leader left, in the same `Apply`.
- [x] Down the descent's one-cell soop, the follower trails single file and is never in an occupied cell.
- [x] Stepping into the follower swaps them.
- [x] `explore.control` to `individual` while calm gives each member's decisions to the player, in timeline order; switching back to `follow` resumes trailing.
- [x] With `Alert` forced on in follow mode, the party switches to individual control; when it's forced off again, it switches back to follow.
- [x] A player who chose individual mode before an alert stays in it after the alert ends; a player who switched to follow during an alert stays in follow until the next alert.
- [x] With `party.individual_on_alert` false, an alert changes nothing.
- [x] `explore.lead` changes the leader, the party order and the controllers, in either mode.
- [x] Replaying 300 random options gives the same hash as playing them.
- [x] `go tool task check` passes.

## Out of scope

- What sets `Alert` (M2-07). Formations wider than single file: file a follow-up if open chambers make the conga line look silly.

## Notes

- **Party order is stored.** `comp.Party` gains `Order []api.UID` (not in the spec's struct), filled at spawn from the template and rewritten by `explore.lead`. "Moves to the front" needs the order of the rest to survive more than one lead change, which deriving it from UIDs would not give with three or more members. `membersOf` sorts by it, so the party view's member order is party order.
- **Constants.** The control modes and `ControlBy` values are `api.ControlFollow`, `ControlIndividual`, `ControlByPlayer` and `ControlByAlert`. `ControlChanged` carries `By`; `LeaderChanged` carries the leader's rendered name for the log.
- **`party.individual_on_alert`** is a plain bool on `PartyConfig`, set to `true` in `party.cue`. A bundle built without it gets `false`, so "default true" lives in the sample content, not in Go.
- **The `follow` rule.** A member steps into the trail cell of the member ahead when that cell is one step away and not the ahead member's own cell. If it is not reachable but the member ahead is adjacent, it waits. Otherwise it paths toward the member ahead (200 nodes) and takes the first step. A destination with any body in it is never stepped into; it waits.
- **Swaps** happen only between members of the same party, and only for a level, up or down step on the same map. A swap across a drop or a link is refused as occupied (see M1-16.F1). The swapped member gets a `Moved` event, a reversed kind, its sight recorded and its `Trail` set.
- **Applying controllers at once.** `explore.control` and `explore.lead` set controllers straight away as well as in the `party_control` system, because `timeline.Run` returns before its per-turn systems when the current body is the player's. When that switch takes the player's current body away from the player, it goes back on the timeline with the new `timeline.Reschedule`; otherwise deciding would have left it with no entry.
- **Alert** is only seen by `party_control` after the next timeline entry, so tests force `Alert` and then apply any command.
- **Placement** is now breadth-first over `Step` (was hex rings), ties by depth, then R, then Q.
- **Files touched outside the list:** `internal/core/timeline/timeline.go` (`Reschedule`); `internal/core/sim/spawn.go` (party `Order`, `Control`); `internal/core/api/explore.go` also holds the commands and events; `internal/content/types.go` and `content/data/party.cue`, `text.cue`; regenerated `*_go_gen.cue`.
- **Existing tests changed.** `TestAStepUpCostsMoreThanTheSameStepDown` now counts busy ticks from the tick of the step, because followers now act inside the same `Apply` and time has moved on when it returns. `TestDefaultVisibilityIsTheParty` registers an idle `follow` controller, since the sample party's follower uses it.
- **Follow-ups filed:** M1-16.F1 (followers and swaps across map links), M1-16.F2 (control mode and alert in `PartyView`). I did not file a follow-up for the conga line in open chambers; the criteria do not exercise one, so there was nothing to judge by.
