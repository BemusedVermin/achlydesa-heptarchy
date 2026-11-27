---
id: M5-13
title: A play-by-play of every fight in the message log
status: todo
milestone: M5
kind: feature
depends_on: []
discovered_by: null
size: M
---

# M5-13 · A play-by-play of every fight in the message log

## Goal

After the player commits a move, they can read, one beat at a time, everything that happened before their next decision: who acted, what landed, which part it hit and for how much, what came off, who went down. Today a hit that lands leaves no trace, so a fight reads as "nothing happened" and then the party wakes the next day.

## Context

- Filed by the owner on 2026-11-26 after playtesting: attacking the moving clay, committing, and seeing only the next day and the wake line, with nothing landing.
- `internal/core/game/log.go` (`installLog`) writes a log line for every event that implements `api.Loggable`. Today that is only `ActionMissed`, `BodyDowned`, `BodyDied`, `PartyWiped`, `PartyWoke`, alert, clock, control and knowledge events.
- `ActionStarted` (`internal/core/api/actions.go`), `ActionLanded` and `ActionCanceled` (`api/timeline.go`), `PartDamaged`, `PartSevered`, `PartAttached` and `StatusAdded` (`api/combat.go`), and `ActionsInterrupted`, `ChannelsLocked` and `ChannelsDelayed` (`api/effects.go`) publish no log text. So a successful hit is invisible.
- `ActionMissed` already carries its own rendered-later line (`Line api.Text`), built by the code that publishes it, which knows the names. Events are plain data and can't look names up themselves (AGENTS.md rule 6).
- One `Game.Apply` runs the timeline to the next decision (`internal/core/game/game.go`), so many beats resolve between two key presses. The TUI (`internal/frontend/tui/app.go`) only re-renders the view, and `RenderLog` in `render.go` shows the last lines that fit in the side panel.
- `docs/game-design.md` → "Combat"; `docs/go-technical-design.md` → the boundary and the view.

## Deliverables

- `internal/core/api/actions.go`, `api/combat.go`, `api/timeline.go`, `api/effects.go`: a log line on the events a player needs in order to follow a fight (see Spec), with regenerated `*_go_gen.cue` schemas.
- The publishers of those events (in `combat`, `effects`, `actions` and `timeline`): fill in the line with names, the action, the part and the amount.
- `content/data/text/combat.cue` (and others as needed): the text for every new line.
- `internal/core/api/view.go`: `LogLine.Pause` (see Spec).
- `internal/core/game/log.go`: carry `Pause` into the view's log.
- `internal/frontend/tui/`: a play-by-play overlay that steps through the lines a command added.
- Tests: a golden encounter's log, and the overlay's stepping.

## Spec

**Lines in the core (suggestion: follow `ActionMissed`).** Give each of these events a `Line api.Text` field that the publisher fills in, and make them `Loggable`:

- `ActionStarted`: "Ada begins a Strike at the clay-thing (lands at tick 12)."
- `PartDamaged`: "The clay-thing's Strike cuts Ada's left arm: 3 integrity lost, 2 left."
- `PartSevered`: "Ada's left arm is cut away."
- `StatusAdded`: "Ada is staggered until tick 20."
- `ActionCanceled` and `ActionsInterrupted`: the actor, the action and why.

`ActionLanded` needs no line of its own: what it did shows as damage, a sever, a status or a miss. The wording is content. The examples above are only for shape.

**Who sees what.** Log a line only when the party could perceive it: a party member is the actor or the target, or the body is in the party's sight. Use the sight the core already keeps. If no cheap check exists, log lines involving party members only, and file a follow-up for sight. The rest of this is a design call the docs don't make (rule 9), so this is the reversible option.

**Pause (suggestion: a field on `LogLine`).** `LogLine.Pause bool` marks a line the frontend should stop on. The core sets it for fight beats: started, damaged, severed, status, missed, canceled, downed, died, wiped, woke. Clock and phase lines don't pause, so walking around stays quick. The core decides which lines matter; the frontend only paces them (rule 7).

**The overlay (TUI).** The app counts the log's lines before `Apply`. If any new line has `Pause`, it pushes a play-by-play screen over the map. The screen shows the new lines one at a time, the earlier ones staying above, "the press A to continue" of an RPG. One key advances, another skips to the end, and the last advance closes the overlay. Key help is the only frontend text. The side panel's log is unchanged.

## Acceptance criteria

- [ ] In a golden encounter where a strike lands, the log holds the start line, the damage line (actor, action, target part, amount lost, integrity left) and, if the part comes off, the sever line.
- [ ] A body the party can't perceive leaves no line (or, without a sight check, a non-party body acting on a non-party body leaves none).
- [ ] Walking with no fight pauses nothing: no clock or phase line has `Pause`.
- [ ] The TUI opens the play-by-play after a command whose new lines include a pause line, steps one line per key, skips with the skip key, and closes after the last line (a TUI test).
- [ ] `go tool task schema` is run and its output committed.
- [ ] `go tool task check` passes.

## Out of scope

- The wipe screen and the "wake" step: M5-14.
- A readable trace of a replay, and saving the record as you play: M5-15.
- Combat numbers and balance: M9.

## Notes

Filled in by the implementer: assumptions, deviations, files touched outside Deliverables, follow-ups filed.
