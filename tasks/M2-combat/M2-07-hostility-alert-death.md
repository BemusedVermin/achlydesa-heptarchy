---
id: M2-07
title: "Hostility, alert and death"
status: todo
milestone: M2
kind: feature
depends_on: [M2-06]
discovered_by: null
size: M
---

# M2-07 · Hostility, alert and death

## Goal

Fights emerge from the map instead of starting as a mode. This task covers three things: hostility between factions; an alert that switches the party to per-member control while a hostile is in sight; and death, which turns bodies into corpses and takes them off the timeline. There's no fight entity anywhere.

## Context

- Game design: "Combat → Fighting on the map", "The living deep → Monsters with needs" (families fight each other), "Three tiers of people" (plot armor).
- Tech design: "Designing for pivots" (Combat sides, Time).
- This milestone's README, "The combat model" (Alert, Sides, Death).
- M1-14 `Visible`, M1-16 `Party.Alert` and lead and follow, M2-02 bodies, `comp.Allegiance` (the `spawn_from_part` effect already adds it), M2-13 rules as content.

## Deliverables

- `internal/content/types.go`: `CombatConfig`; `MapDef.Spawns`; `Template.Faction`; regenerate the schema.
- `content/data/combat.cue`; factions on the person templates and the sample creature in `templates.cue`; a creature spawn on `map.test.hall` and one in the soop under the hall on `map.test.descent`.
- `internal/core/comp/comp.go`: `Calm`.
- `internal/core/api/combat.go`: `AlertStarted`, `AlertEnded`, `BodyDied`, `BodyDowned`, `PartyLost`.
- `internal/core/combat/hostility.go`, `alert.go`, `death.go`, `install.go`, tests.
- `internal/core/explore`: spawn map marks' templates at start.
- From M1-16.F2: `internal/core/api/explore.go`: `PartyView` gains `Control string` (`follow` or `individual`) and `Alert bool`; `internal/core/explore/project.go` fills them. Test: the view shows the mode before and after `explore.control`, and the alert when forced.
- `internal/core/game`: stop accepting commands after `PartyLost`.

## Spec

**Content:** `CombatConfig{ CalmTicks int64; SightRadius, SightDepth int; CorpseGlyph string; CorpseColor string }` under `combat`. `MapDef` gains `Spawns map[string]api.ContentID`: a mark character to a template, spawned at start in row order, with `Pos` and the template's `Turn`.

**Sides:** `Template.Faction *api.ContentID` (`faction`), checked at load to name a faction. A combat spawner gives each body spawned from such a template `comp.Allegiance{Faction}`.

**Hostility:**

```go
// Hostile reports whether two bodies are enemies: either one's faction lists the other's.
func Hostile(w *sim.World, a, b ecs.Entity) bool
```

**Alert:** a `PerTurn` system, `combat.alert`, for the controlled party.

- If any living hostile body is visible (`hexmap.Visible` with the combat sight radius and depth) from any living member: set `Calm.Since` to now. If `Party.Alert` was false, set it and publish `AlertStarted` (loggable, naming the first hostile seen by UID).
- Otherwise, once `CalmTicks` have passed since `Calm.Since`, clear `Alert` and publish `AlertEnded` (loggable).
- Register M1-19's hostility hook with `Hostile`, so the `allies` and `enemies` selectors work.
- M1-16's `party_control` already switches controllers on `Alert`. Register `combat.alert` before it (`After`), so the switch takes effect at once.

**Death:** a subscriber to `PartDamaged` and `PartSevered` checks the README's death rule for the affected body.

- **Dies:** publish `BodyDied`, add `Dead{Tick}`, remove its `Actor` so it leaves the timeline, cancel its pending actions, and set its `Appearance` to the corpse glyph and color.
- **Protected:** it becomes `Downed` instead; `BodyDowned` is published, and it leaves the timeline the same way until something revives it (later milestones).
- **Party wipe:** when every member of the controlled party is dead or downed, publish `PartyLost` (loggable). `game.Apply` rejects every later command with `text.game.lost`. File a `question` follow-up on what a party wipe means for the story.

## Acceptance criteria

- [ ] Walking the leader into sight of the hall's creature sets `Alert`, publishes `AlertStarted`, and switches a party in follow mode to individual control.
- [ ] On `map.test.descent`, standing in the hall directly over the creature in the soop below never sets `Alert`: there's no sight through rock.
- [ ] After `CalmTicks` with no hostile visible, `AlertEnded` fires and lead and follow resume, unless the player chose individual control.
- [ ] `Hostile` is symmetric when only one faction lists the other (table test).
- [ ] Applying damage directly (through effects) until a vital part is gone kills the body: the corpse stays on the map with its corpse glyph and never decides again.
- [ ] A protected member is downed instead of killed; downing every member publishes `PartyLost`, and the next command is rejected.
- [ ] `go tool task check` passes.

## Out of scope

- Attacking, options and creature behavior (M2-08). The action menu (M2-10).

## Notes
