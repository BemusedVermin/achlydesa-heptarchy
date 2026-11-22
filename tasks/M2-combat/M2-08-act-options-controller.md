---
id: M2-08
title: The bodies view and the creature controller
status: todo
milestone: M2
kind: feature
depends_on: [M2-07]
discovered_by: null
size: M
---

# M2-08 · The bodies view and the creature controller

## Goal

The player can see enough to plan: every visible body's parts, what's known about them, which of their parts are busy, and what's about to land on whom. Hostile creatures act through a controller that uses the same actions and checks as the player, scored by a content formula. There's no attack command of its own: the player attacks with M1-20's `actions.do` and its options.

## Context

- Tech design: "Designing for pivots" (Choices come from the core).
- Game design: "Combat → Timing", "Combat → Lethality", "Combat → Grafts" (grafts quietly shape what's offered: keep option order in M1-20's one function).
- M1-17 channels and controllers, M1-18 pending actions, M1-20 actions and options, M2-06 forms, M2-07 hostility and alert, M2-13 records and rules in formulas.

## Deliverables

- `internal/core/api/combat.go`: the `BodiesView` section.
- `internal/content/types.go`: `CombatConfig.AIScore` (Expr over `ScoreEnv`), `CombatConfig.Aims` (Expr over M2-13's `GrantEnv`), `CombatConfig.AttackTag` and `CombatConfig.StepAction`; regenerate the schema.
- `internal/core/combat/view.go`, `controller.go`, tests and a golden file.

## Spec

**`BodiesView`,** added to `View` as `Bodies *BodiesView`:

- `Current`: the deciding body's UID.
- `Bodies`: every visible body plus all party members. Each has its UID, name, faction name, whether it's hostile to the party, statuses, and its channels. For each part channel:
  - name, slot, integrity and max, and whether it's a graft;
  - material, or "?" when the party doesn't know it (via `knowledge.Best`); members' own materials are always known;
  - ticks until it's free (0 if free), and whether a lock is the reason;
  - what it's winding up, if anything: the action's name, target, and ticks until it lands.
- `Threats`: every pending action aimed at a party member, with attacker, target, target part and ticks until landing, sorted by landing tick. The map uses this to highlight threatened members, and the timeline strip already lists the landings (M1-18).

**The `hunt` controller,** registered with the timeline:

1. Pick the nearest visible hostile, by hex distance plus depth difference, then UID.
2. If any startable action (`actions.Startable`) with the tag `combat.attack_tag` names (`attack` to start) reaches it, choose one with the chooser and start it.
3. Otherwise, if the action `combat.step_action` names (`action.step` to start) is startable, step along `hexmap.Path` toward it (at most 200 nodes).
4. Otherwise wait until the next entry.

A body for which `combat.aims` is false picks its target at random among valid intents (stream `combat`). `aims` starts as `grant.targeting > 0`, so a body that has lost its head swings wild; code never names the grant kind. Hunting creatures don't dodge yet; reading wind-ups is M3 behavior.

**Chooser:**

```go
type Chooser interface {
	Choose(w *sim.World, actor ecs.Entity, intents []actions.Intent) (actions.Intent, bool) // false = none
}
// ScoreChooser scores each intent with content's ai_score and takes the best, breaking ties with the combat stream.
type ScoreChooser struct{}
```

`ScoreEnv` exposes at least M2-13's `source`, `target`, `chosen` and `pack`, so scores can use `rule.weak` and read the loaded type's `lists.consumes`, plus `immune_known`, `target_is_party`, `target_integrity`, `windup` and `height`. Choosers are registered by name, so `agents` can supply a smarter one in M3.

## Acceptance criteria

- [ ] On alert, the deciding member's options list only legal intents. A `Do` built from an option always starts; a hand-built `Do` with an out-of-reach target or a busy arm is rejected with a reason.
- [ ] Unknown materials show "?" until a Sweep lands, then show the material.
- [ ] `Threats` lists a creature's swing at the Reaper with the right landing tick, and it's gone once the swing lands or misses.
- [ ] A hunting creature closes in over several steps and attacks once in reach.
- [ ] The party can dodge: stepping out of reach before a hunter's swing lands makes it miss.
- [ ] A creature with its legs severed falls behind a party walking away, and the alert eventually ends.
- [ ] Two creatures of mutually hostile factions fight each other with the party out of sight.
- [ ] A scripted encounter's event log matches a golden file.
- [ ] `go tool task check` passes.

## Out of scope

- Drawing any of this (M2-10). Needs-driven behavior and dodging AI (M3).

## Notes
