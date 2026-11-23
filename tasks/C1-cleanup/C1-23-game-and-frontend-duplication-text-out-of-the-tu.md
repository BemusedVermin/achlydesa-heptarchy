---
id: C1-23
title: "Game and frontend: duplication, text out of the tui, tests, comments"
status: todo
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

Per the decision: the view supplies the text (a channel state label, option group labels); hot reload sends an opaque `func(*game.Game) error`; update `tui/doc.go` and `frontend.md` to say the same. Goldens change.

### Game, save, harness and tui tests: shrink the slow ones

These tests run in a few seconds, and the menu goldens show only the menu.

Findings:

- `TestLoadingAfterRandomPlayContinuesIdentically` (`save/save_test.go:201-221`) takes 24.35s: `newBundle(t)` reloads CUE on each iteration, up to 100 commands per case, and the helpers (`newRig`, `pick`, `hash`) get the outer `t` not `rt`, so rapid cannot minimize a failure. Build the bundle once, make helpers take `testing.TB` and pass `rt`, cut the slices to `before` 0..20 and `after` 0..10, turn the recorded failure (seed 0, before [491 8 720], after [222 823 592]) into a table case, delete the `.fail` file under `testdata/rapid/`. Target under 2s.
- Harness CSV tests: `TestRunWritesOneRowPerStep` (`harness_test.go:46`, 4.87s) 3×200 → 2×10; `TestRunIsByteIdentical` (`:61`, 4.17s) 2×100 → 2×15; `TestRunFillsTheCombatColumns` (`combat_test.go:14`, 2.22s) → 1 seed. `TestLethalityTargets` (`lethality_test.go:58`, 7.92s): keep, skip under `-short`.
- Five replay-hash tests: merge `TestSameSeedAndCommandsGiveTheSameHash` (`game_test.go:139`) into `TestReplayGivesTheSameHash` (`:123`); delete `TestReplayPrintsTheGamesHash` (`harness_test.go:68`); keep `TestReplayExpect` and the tui one. Delete `TestSwapContentReplacesContentAndWarns` (`game_test.go:151`, asserts almost nothing).
- The four menu goldens (`menu_actions`, `menu_types`, `menu_parts`, `menu_arms`, written by `tui/combat_test.go:135-156`) capture the whole 180×30 screen: golden-test only the menu pane and keep `TestViewDrawsEveryPane`.

### Comments: game, save, harness, tui

About 6 comments are cut or fixed.

Findings:

- `game/install.go:15-16` "(M1-10 adds explore)"; `tui/mapscreen.go:14-15` "M2-10's action panel..."; `harness/doc.go:1-3` stale ("a simple policy" → pluggable policy, scripted fights); `tui/doc.go:8` conflicts with `frontend.md` (settled by the tui text task); `game/game.go:21-22` `Config.Seed` and `Config.Content` have no doc (add "Seed is ignored by Load"); `save/migrate.go:12-14` second sentence goes stale.

The policy is the comments lens: doc comments say what a thing is for and what a caller must know, and every sentence carries information (cut name-restatements, filler, history, task IDs). Other comments are terse and say only why. CUE follows the same rules. The examples below set the standard; sweep the whole area.

## Acceptance criteria

- [ ] `go tool task same-behavior` passes; the exceptions are the deliberate behavior changes named in the parts above, each listed in Notes with the changed output
- [ ] The tests commit changed only `_test.go` files and `testdata/`
- [ ] `go tool task check` passes (on the owner's machine too, which the Windows fix in the tools task needs)
- [ ] Each finding above is resolved, one per line in Notes, or marked dropped with a reason
- [ ] `go tool task same-behavior` shows only the wording differences named in Notes.
- [ ] Only `_test.go` files and `testdata/` changed.

## Out of scope

- Other areas' findings: they have their own task in this milestone.
- Any new behavior beyond what the Spec names.

## Notes


- Triaged 2026-11-23: todo in C1, as the owner decided; it is part of the first cleanup milestone.
