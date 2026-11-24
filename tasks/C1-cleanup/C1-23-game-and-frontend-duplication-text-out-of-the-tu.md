---
id: C1-23
title: "Game and frontend: duplication, text out of the tui, tests, comments"
status: done
milestone: C1
kind: debt
depends_on: [C1-18, C1-12]
discovered_by: null
size: L
---

# C1-23 · Game and frontend: duplication, text out of the tui, tests, comments

## Goal

game, save, harness and the tui are tidy, the tui draws no game text of its own, and the slow tests are fast.

## Context

- Milestone brief: [README.md](README.md). Findings come from the audit that opened C1.
- `docs/glossary.md` for settled terms.
- This is one pull request by design: the work touches the same packages and is done in one pass.

## Deliverables

- The changes described in the Spec, in one branch and one pull request.

## Spec

Do the parts in this order, one commit each, because tests follow the code they test and comments come last once code and tests have settled: duplication and libraries, then tests, then comments.

### game, harness and tui: one step loop, one option lookup

The choose, apply and view loop and the option lookup each exist once.

Findings:

- The loop is hand-written 7 times: `harness/run.go:96-118`, `harness/combat.go:133-147`, `harness/harness_test.go:72-86`, `cmd/harness/main_test.go:29-41`, `harness/bench_test.go:26-41`, `save/fight_test.go:44-58`, `save/save_test.go:100-117` (`pick`). Export `harness.Step(g, p) (events []api.Event, ok bool, err error)` and `harness.Play(g, p, steps) error` (suggestions); `runSeed` and `PlayFight` keep only their tracking.
- The tui looks up an option by ID five times: `app.go:71-86`, `keys.go:66` (`hasOption`), `keys.go:107-117`, `mapscreen.go:52-57`, `mapscreen.go:140-145`: one `optionByID(v api.View, id string) (api.Option, bool)`.
- Inside `game`: log append-and-trim (`game.go:61-66`, `log.go:20-23`) → `g.appendLog`; the `effects`/`actions` `ValidateContent` pair (`game.go:46-53`, `:139-144`) → `validateContent(w)`; the end of `Apply` (`game.go:168-171`, `:189-192`); `PositionOf` and `PartOfUID` (`inspect.go:12-37`) → a generic `component[T any]`; `applyEdits` (`save.go:138-142`) copies the `terrain` closure of `explore/source.go:27-33`: export `explore.Terrain(reg, id)`. (Content test fixtures are the fixture task's.)
- Libraries: `wg.Go` (`harness/lethality_test.go:25`), `for b.Loop()` (`harness/bench_test.go:24-31`, and drop the `loadRegistry` wrapper at `:49`), `for i := range n` (`game_test.go:100`, `readiness_test.go:23`), `strings.SplitSeq` (`menu.go:202,220`), `slices.Clone` (`log.go:24`, `save.go:73`).
- Conventions: `SwapContent` and `Note` come before the `Game` type and `New` (`game.go:37-67`); the same for `newMapScreen` (`mapscreen.go:88`); `Load` checks for nil content (`save.go:83-85`) though `build` does, with an `fmt.Errorf` without verbs (lint: `perfsprint`).

### The tui draws no game text of its own

Every word of game state the player reads comes from the core's view.

Findings:

- Hard-coded in the tui: `mapscreen.go:204` "ALERT: hostile in sight", `:207` "Control: ", `:243` "Tick %d"; `look.go:49,51` "Looking"; `actorpanel.go:107-115` "graft", "locked", "free in %d", "ready"; `actorpanel.go:128` "lands at tick %d"; `render.go:278` "Day %d · %s"; `menu.go:216` "Actions"; `menu.go:114-128` `segmentLabel` capitalizes the end of a content ID.
- `ContentMsg` (`app.go:91`) carries a `*content.Registry`, which `frontend.md` doesn't allow; `tui/doc.go:8` says otherwise.

A timeline line never shows a wrong number (C1-17.F2): in the actor panel a long target name used to cut the line's last digit (`lands at tick 8` for tick 86). Wrap or abbreviate the name, or widen the panel, whichever is least invasive, and check `actor_windup.golden` shows `lands at tick 86`.

Per the decision: the view supplies the text (a channel state label, option group labels); hot reload sends an opaque `func(*game.Game) error`; update `tui/doc.go` and `frontend.md` to say the same. Goldens change.

### Game, save, harness and tui tests: shrink the slow ones

These tests run in a few seconds, and the menu goldens show only the menu.

Findings:

- `TestLoadingAfterRandomPlayContinuesIdentically` (`save/save_test.go:201-221`) takes 24.35s: `newBundle(t)` reloads CUE on each iteration, up to 100 commands per case, and the helpers (`newRig`, `pick`, `hash`) get the outer `t` not `rt`, so rapid cannot minimize a failure. Build the bundle once, make helpers take `testing.TB` and pass `rt`, cut the slices to `before` 0..20 and `after` 0..10, turn the recorded failure (seed 0, before [491 8 720], after [222 823 592]) into a table case, delete the `.fail` file under `testdata/rapid/`. Target under 2s.
- Harness CSV tests: `TestRunWritesOneRowPerStep` (`harness_test.go:46`, 4.87s) 3×200 → 2×10; `TestRunIsByteIdentical` (`:61`, 4.17s) 2×100 → 2×15; `TestRunFillsTheCombatColumns` (`combat_test.go:14`, 2.22s) → 1 seed. `TestLethalityTargets` (`lethality_test.go:58`, 7.92s): keep, skip under `-short`.
- Five replay-hash tests: merge `TestSameSeedAndCommandsGiveTheSameHash` (`game_test.go:139`) into `TestReplayGivesTheSameHash` (`:123`); delete `TestReplayPrintsTheGamesHash` (`harness_test.go:68`); keep `TestReplayExpect` and the tui one. Delete `TestSwapContentReplacesContentAndWarns` (`game_test.go:151`, asserts almost nothing).
- The four menu goldens (`menu_actions`, `menu_types`, `menu_parts`, `menu_arms`, written by `tui/combat_test.go:135-156`) capture the whole 180×30 screen: golden-test only the menu pane and keep `TestViewDrawsEveryPane`.

### Comments: game, save, harness, tui

Every comment `go tool task comments -- ./internal/core/game ./internal/core/save ./internal/harness ./internal/frontend ./cmd` lists is cut or fixed (32 when the rule tightened), along with the findings below.

Findings:

- `game/install.go:15-16` "(M1-10 adds explore)"; `tui/mapscreen.go:14-15` "M2-10's action panel..."; `harness/doc.go:1-3` stale ("a simple policy" → pluggable policy, scripted fights); `tui/doc.go:8` conflicts with `frontend.md` (settled by the tui text task); `game/game.go:21-22` `Config.Seed` and `Config.Content` have no doc (add "Seed is ignored by Load"); `save/migrate.go:12-14` second sentence goes stale.

The policy is `.claude/rules/comments.md`: doc comments say what a caller needs, and every sentence carries information; other comments say only why; no end-of-line comments, no comments in tests other than doc comments, no task IDs. CUE follows the same rules. The examples below set the standard; sweep the whole area.

## Acceptance criteria

- [x] `go tool task same-behavior` passes; the exceptions are the deliberate behavior changes named in the parts above, each listed in Notes with the changed output
- [x] The tests commit changed only `_test.go` files and `testdata/`
- [x] `go tool task check` passes (on the owner's machine too, which the Windows fix in the tools task needs)
- [x] `go tool task comments -- ./internal/core/game ./internal/core/save ./internal/harness ./internal/frontend ./cmd` reports nothing
- [x] Each finding above is resolved, one per line in Notes, or marked dropped with a reason
- [x] `go tool task same-behavior` shows only the wording differences named in Notes.
- [x] A timeline line in the actor panel never cuts its tick, even with a long target name (C1-17.F2).
- [x] Only `_test.go` files and `testdata/` changed.

## Out of scope

- Other areas' findings: they have their own task in this milestone.
- Any new behavior beyond what the Spec names.

## Notes


- Triaged 2026-11-23: todo in C1, as the owner decided; it is part of the first cleanup milestone.

Commits: duplication (c93d338), tui text (ab6128c, plus 59507fe for C1-17.F2), tests (4e265d2), comments (0e20800), review fixes (21e45eb).

**Duplication and libraries**

- The step loop: `harness.Step(g, p)` and `harness.Play(g, p, steps)` replace all 7 loops. `PlayFight` uses an unexported `stepFrom(g, p, view)`, so it still builds one view per step.
- `optionByID(v, id)` replaces the 5 lookups. `hasOption` stays as a one-line wrapper over it.
- Also replaced: `g.appendLog`, `validateContent(w)`, `g.record(cmd)` (the end of `Apply`), `component[T]` (behind `PositionOf` and `PartOfUID`), and `explore.Terrain(reg, id)` (used by `NewSource` and `applyEdits`).
- Libraries: `wg.Go`, `b.Loop` (there was no `loadRegistry` left), `range n` at both sites, and `strings.SplitSeq`. `slices.Clone` is in `log.go`.
- **Dropped:** `slices.Clone` at `game/save.go:73`. `slices.Clone(nil)` writes `null` where `append([]api.LogLine{}, ...)` writes `[]`, which would change the save bytes and every hash.
- Conventions:
  - `SwapContent` and `Note` now follow `build`, and `newMapScreen` follows its type.
  - `Load` no longer checks for nil content: `build` does, so the `perfsprint` hit is gone.

**The tui draws no game text**

The view now carries every word the tui drew itself, rendered from new content text:

| Field | Text | Was in the tui |
| --- | --- | --- |
| `ClockView.Label` | `text.clock.label` | "Day %d · %s" |
| `PartyView.Status` | `text.explore.status.*` | "ALERT: hostile in sight", "Control: …" |
| `ChannelView.State` | `text.combat.channel.*` | "ready", "free in %d", "locked", "graft" |
| `WindupView.When` | `text.combat.lands_at` | "lands at tick %d" |
| `TimelineEntry.TickLabel` | `text.timeline.tick` | "Tick %d" |
| `View.Labels` (new `api.LabelsView`) | `text.game.actions`, `text.game.look`, `text.game.look_empty` | "Actions", "Looking" |
| `Option.Path` | the rendered name of each ID segment, built in `actions` | `segmentLabel` |

- Where the tui used to work this out:
  - `PartyView.Status` shows the control mode under the same condition the control options are offered (a new `explore.playerDecides`).
  - `ChannelView.State` puts graft before the busy state.
  - `Option.Path` names each ID segment: the action's name, each choice's record name, the direction, and the body or part names. Parts use their part template's name.
- `App.Mode()` is removed; its test now reads `State.Party.Control`. `RenderBody` lost its `now` argument.
- `ContentMsg` is now `{Reload func(*game.Game) error}`, so the tui imports only `game` and `api`. The reload-failure wording moved to `cmd/heptarchy`.
- `tui/doc.go` now says the same. `.claude/rules/frontend.md` and `docs/go-technical-design.md:233` already did, so I didn't change them.
- The goldens did not change: the core renders the same words the tui did.
- C1-17.F2: `RenderBody` wraps a long part or swing line with `lipgloss.Wrap` instead of cutting it. `TestALongTargetNameWrapsTheSwingInsteadOfCuttingItsTick` covers it, and `actor_windup.golden` shows `lands at tick 86`.

**Tests**

- `TestLoadingAfterRandomPlayContinuesIdentically`:
  - The registry is built once, and the helpers take `rapid.TB`. `*rapid.T` is not a `testing.TB`, so this is a deviation from the spec's wording.
  - The slices are cut to 0..20 and 0..10. The recorded failure is a table case, and the `.fail` file is deleted.
  - Hashes are compared after the load and at the end, not after every command. A split can't rejoin, because the hash covers the history and the RNG state.
  - It went from 24s to 9-12s on a shared machine. **The 2s target is missed**: what remains is the cost of play and hashing per command. Deferred to C1-23.F2.
- Harness CSV tests: 2×10, 2×15 and 1 seed. `TestLethalityTargets` skips under `-short`.
- `TestSameSeedAndCommandsGiveTheSameHash` is merged into `TestReplayGivesTheSameHash`. `TestReplayPrintsTheGamesHash` and `TestSwapContentReplacesContentAndWarns` are deleted.
- Menu goldens show only the menu pane, through `export_test.go` (`MenuPane`). `TestViewDrawsEveryPane` is kept.
- `save/fight_test.go`: each game now has its own bound policy and chooses from its own view. The hashes are still compared after every step.
- Added after review:
  - `actions/path_test.go`: every option's `Path` matches its ID segments.
  - A tui test for `TickLabel` and `LookEmpty`.
- The last criterion ("Only `_test.go` files and `testdata/` changed") can only apply to the tests commit, and 4e265d2 holds only those.

**Comments**

All 28 reports are fixed, including `install.go` and `mapscreen.go` (task IDs), `harness/doc.go`, `Config.Seed` and `Config.Content`, and `migrate.go`. I also tidied a few doc comments in the area, such as `View`, `Load` and `DefaultKeyMap`.

**Behavior and wording changes** (same-behavior: CSV identical; differences listed)

- Final harness hashes change, because the new text records change the content hash in the save header.
- The four menu goldens are now just the menu pane.
- The `.fail` file is removed.
- Hot-reload failure line in the log: still "content reload failed, keeping the old content: <err>", now built in `cmd/heptarchy`.
- The `applyEdits` error reads "saved edit: unknown terrain …" (was "saved edit uses unknown terrain …").
- `runSeed` wraps view errors with "step N:".

**Outside the four areas:** `actions/options.go`, `api/*.go` and their regenerated `*_go_gen.cue`, `clock/install.go` (and its test), `combat/view.go`, `explore/{control,project,source}.go`, `timeline/install.go`, and `content/data/text/*.cue`.

**Follow-ups:**

- C1-23.F1: the dev reload log lines are still English in Go.
- C1-23.F2: bring the save property test under 2s.
