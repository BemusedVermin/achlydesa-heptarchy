---
id: M5-14
title: A wipe screen that waits for the player before the party wakes
status: done
milestone: M5
kind: feature
depends_on: [M5-13]
discovered_by: null
size: M
---

# M5-14 · A wipe screen that waits for the player before the party wakes

## Goal

When the party is wiped out, the game stops and says so on its own screen. The party wakes only when the player chooses to go on. Today the wipe, the time skip and the wake all happen inside the one command that lost the fight, so the player sees only a new day and "The party wakes, bruised, where it last rested."

## Context

- Filed by the owner on 2026-11-26 after playtesting (see M5-13's Context).
- `Game.Apply` in `internal/core/game/game.go` calls `wakeIfWiped` at the end of every command. That advances the clock by `party.wake_ticks`, calls `explore.Wake` and plays on to the next decision, all in the same `Apply`.
- A wiped party with nobody downed can't wake. `Apply` then rejects every command with `text.game.lost`, and the player gets no screen for it either.
- `docs/game-design.md` → "A party wipe is not game over" (line ~198): the party wakes where it last rested, after `party.wake_ticks`.
- The TUI has a screen stack (`internal/frontend/tui/app.go`, `Screen`, `Push`). M5-13 adds the play-by-play overlay, which should run first so the player sees the blows that wiped the party.

## Deliverables

- `internal/core/api/`: a wake command and a wipe section of the view, with regenerated schemas.
- `internal/core/game/game.go`: stop waking inside `Apply`; wake on the wake command instead.
- `content/data/text/game.cue`: the wipe screen's title and lines, and the wake option's label.
- `internal/frontend/tui/`: a wipe screen.
- Tests: game (wipe, then the wake command, then play goes on; a lost party has no wake option; save and load while wiped) and TUI (the screen opens after the play-by-play and closes on the wake option).
- `docs/game-design.md`: say the party wakes when the player goes on.

## Spec

**Core (suggestion).**

- `api.Wake` is a command, kind `game.wake`, with no fields. It is recorded in the history like any command, so replays still match.
- While `combat.PartyWiped` holds, `Apply` no longer wakes the party. The view's `Options` hold just the wake option when `explore.CanWake`, and nothing when the party is lost. Any other command is rejected, as today.
- Applying `Wake` does what `wakeIfWiped` does now: advance `party.wake_ticks`, `explore.Wake`, run the timeline, drain the bus.
- `View.Wipe *WipeView` is present while the party is wiped. It holds the rendered title and body lines, and whether the party can wake. It is set from world state, so a game saved while wiped loads onto the same screen.

**TUI.** When the view has `Wipe`, after any play-by-play has finished, push a full-screen wipe screen. It shows the title and the lines from the view, and the core's options as a menu: one "wake" option, or none for a lost party, where only quitting is left. Choosing the wake option sends its command and pops the screen.

## Acceptance criteria

- [x] After a command that wipes the party, the party has not woken: the tick hasn't jumped by `party.wake_ticks`, and the view has `Wipe` with one wake option.
- [x] Applying the wake option's command wakes the party where it last rested, after `party.wake_ticks`, and play goes on.
- [x] A wiped party with nobody downed gets a `Wipe` view with no wake option.
- [x] A game saved while wiped loads with `Wipe` set, and the same seed and commands give the same `Game.Hash` (replay test).
- [x] In the TUI, the wipe screen opens after the fight's play-by-play, and choosing the wake option closes it (a TUI test).
- [x] `docs/game-design.md` says the party wakes when the player goes on.
- [x] `go tool task check` passes.

## Out of scope

- Waking at a town, and what the enemies did meanwhile: the owner will tie the restart to towns later. Waking beside the monster that won is acceptable until then.
- The play-by-play itself: M5-13.

## Notes

- `Wake` and the `game.wipe` view builder live in `internal/core/game/wipe.go`; the wipe view clears `Options` down to the wake option (none for a lost party).
- A `Wake` sent when the party can't wake is rejected with `text.game.cannot_wake`; any other command while wiped is still `text.game.lost`.
- The wipe screen sits under the play-by-play on the stack, so the blows show first. `App.Update` now deletes the popping screen by its own index, so a screen can pop while an overlay sits above it.
- The TUI test starts from an already wiped game; the push-after-play-by-play order is by construction in `App.Choose`, not tested.
- Existing wake tests now send the wake command. No follow-ups.
