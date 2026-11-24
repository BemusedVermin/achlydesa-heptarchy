---
id: C1-22
title: "Foundation area: duplication, tests, comments, speed errors"
status: done
milestone: C1
kind: debt
depends_on: [C1-18, C1-16]
discovered_by: null
size: L
---

# C1-22 · Foundation area: duplication, tests, comments, speed errors

## Goal

The foundation packages are tidy, independent of content tuning, and a failing speed formula is an error.

## Context

- Milestone brief: [README.md](README.md). Findings come from the audit that opened C1.
- `docs/glossary.md` for settled terms.
- This is one pull request by design: the work touches the same packages and is done in one pass.

## Deliverables

- The changes described in the Spec, in one branch and one pull request.

## Spec

Do the parts in this order, one commit each, because tests follow the code they test and comments come last once code and tests have settled: duplication and libraries, then tests, then comments.

### foundation: stdlib forms, test-only API out of the package, honest Install signatures

The foundation packages use the standard library where it does the job.

Findings:

- `sort.Strings` → `slices.Sort` (`sim/order.go:42,60`); `reflect.DeepEqual` on string and command slices → `slices.Equal` (`clock/clock_test.go:52,64,80`, `api/codec_test.go:156`).
- `hexmap/step.go:12-13` `DefaultStepRules` is used only by hexmap tests and hard-codes 1/1/8 in Go (rule 4): move it into a `_test.go` file. `clock/install.go:11` takes `reg *content.Registry` while every other `Install` reads `w.Content`: drop the parameter.
- Test helpers missing `t.Helper()`: `hexmap/step_test.go:57,63,67`, `api/codec_test.go:36`.

Lint: `thelper`, and `predeclared` or `revive` `redefines-builtin-id` (see the hexmap task for `clear`). Tried at lock-in.

### A speed formula that fails returns an error

A content bug in a speed formula shows up as an error instead of a body that never steps.

Findings:

- `anatomy/install.go:85-93` `speedOf` turns the error into speed 0.

Per the decision: `timeline.SetSpeed` hooks return `(int, error)` and the timeline propagates it, or the loader proves the formula cannot fail. A test shows the failure surfaces.

### Foundation tests: independent of content tuning, one save round trip

Tuning content doesn't break foundation tests.

Findings:

- `clock_test.go:46,60,104-110` pin the live calendar (tick 19/20, 240, the 20/100/30/90 phases): build a `content.Calendar` literal and skip `content.Load`. `anatomy_test.go:78,83,108,235,262-268` pin the sample's members, slots, channels, grants: take expectations from the loaded plan or install a small test plan. `inventory_test.go:105,192` pin "11 types" and six starting items: drop.
- `inventory_test.go:100,128,151` (`TestAllElevenTypesLoad...`, `TestOneSidedHostilityWarnsButLoads`, `TestBadTypeContentIsRejected`) test the content loader: move them to `internal/content`.
- Three save round trips (`anatomy_test.go:239`, `knowledge_test.go:181`, `inventory_test.go:177`): merge into the `save` test next to `TestSaveThenLoadWithASpawnedActorHasTheSameHash` (`save_test.go:133`), spawning the party, calling `knowledge.Learn` once and comparing `Game.Hash`.
- Repeats: `rng_test.go:19,25` covered by the property at `:84`; `for range 100` repeat loops (`hexmap/map_test.go:126`, `path_test.go:106`, `sight_test.go:97`) → one second call; `clock_test.go:75` 1000 steps → one day plus one tick.

### Comments: foundation

About 8 comments are cut or fixed.

Findings:

- `api/knowledge.go:51-53` garbled doc (line 53 repeats 52); `api/view.go:3-6` task-ID history and omits `Bodies` ("Each package fills its own section as a pointer field; nil means the section is absent."); `clock/clock.go:62` "(M1-17)"; `hexmap/chunk.go:81` "procgen in M5"; `anatomy/install.go:10` omits `grantTag`; `hexmap/step.go:12` restates the code.

The policy is the comments lens: doc comments say what a thing is for and what a caller must know, and every sentence carries information (cut name-restatements, filler, history, task IDs). Other comments are terse and say only why. CUE follows the same rules. The examples below set the standard; sweep the whole area.

## Acceptance criteria

- [x] `go tool task same-behavior` passes; the exceptions are the deliberate behavior changes named in the parts above, each listed in Notes with the changed output
- [x] The tests commit changed only `_test.go` files and `testdata/`
- [x] `go tool task check` passes (on the owner's machine too, which the Windows fix in the tools task needs)
- [x] Each finding above is resolved, one per line in Notes, or marked dropped with a reason
- [x] `go tool task same-behavior` passes: a game with working formulas is unchanged.
- [x] Only `_test.go` files and `testdata/` changed.

## Out of scope

- Other areas' findings: they have their own task in this milestone.
- Any new behavior beyond what the Spec names.

## Notes

Commits, one per part: `00c8b44` stdlib and APIs, `9546bd1` speed errors, `bb20bfb` tests (only `_test.go` files), `193183a` comments, then `98e57f1` review fixes (tests and a comment revert).

Findings, one per line:

- `sim/order.go` `sort.Strings` → `slices.Sort`: done. `clock_test.go` and `api/codec_test.go:156` `reflect.DeepEqual` → `slices.Equal`: done.
- `hexmap.DefaultStepRules`: moved to `step_test.go` as the unexported `defaultStepRules`.
- `clock.Install(w, c, reg)` → `clock.Install(w, c)`: reads `w.Content` at view time, and returns an error when the world has no content, as `anatomy`, `timeline` and `inventory` do. Callers updated: `game/install.go`, `effects_test.go`, `timeline_test.go`.
- `t.Helper()`: `api/codec_test.go` `newTestCodec` already called it; thelper wanted the `testing.TB` parameter named `tb`, so it is. The `hexmap/step_test.go` (and `dirto_test.go`) setup closures are gone: the tables list cells to `raise` and `dig`, and a `shaft(depth)` helper builds the drop columns. `thelper` and `predeclared` report nothing in the foundation packages now. No `clear` shadowing is left in hexmap (`map.go:68` is the builtin). Turning these linters on is left to the lock-in, C1-27, since agents can't edit `.golangci.yml`.
- Speed errors: `timeline.SetSpeed` hooks return `(int, error)`. `timeline.Speed` returns `(int, error)`. `timeline.StepTicks` propagates the error. `anatomy.speedOf` wraps the content error and no longer returns 0. The callers in `actions/actions.go` (`env`) and `explore/install.go` (`stepCost`) return the error. `TestAFailingSpeedFormulaIsAnError` swaps in a speed formula that divides by zero and checks that `Speed` and `StepTicks` both fail.
- `clock_test.go`: runs on a 12-tick, three-phase `content.Calendar` literal. `setup` still sets `w.Content` to the cached `contenttest.Sample` registry, because the view renders phase names through content and `content.New` can't build a registry from a calendar alone. The view test computes the expected name with `Render`, so it pins no sample text. That is one cached load per test binary, not one per test.
- `anatomy_test.go`: expectations come from the member's loaded plan: slots, vital, integrity, channel tags, grant totals, detach and reattach. Protected comes from the template. The save round trip is removed, the determinism test no longer pins 12 entities, and the legs test finds the parts that grant speed instead of using index 3.
- `inventory_test.go`: the pin on 11 types and the six starting items is dropped. One `TestSpawningGivesTheTemplatesInventory` remains, so the spawner keeps a test; it takes its expectations from the party template.
- The three loader tests moved to `internal/content/types_test.go` (package `content`, using `sampleRegistry` and `sampleBundle`). The 11-type count is dropped, and `TestAllElevenTypes...` is renamed `TestSampleTypesLoadAndNameIsNeverRefused`.
- Save round trips: the ones in anatomy, knowledge and inventory are removed and replaced by `save_test.go` `TestSaveThenLoadWithThePartyAndAFactHasTheSameHash`. It spawns the party, calls `knowledge.Learn` once, checks that the fact changes the hash, and compares `Game.Hash` across a reload.
- Repeats: `rng_test.go` `TestSameSeedAndNameReproduce` and `TestCreationOrderIrrelevant` are removed, because `TestProperties` covers them. The hexmap `for range 100` loops (map, path, sight) are now one second call. The clock chunking test runs one day plus one tick.
- Comments:
  - `api/knowledge.go` garbled line: cut.
  - `api/view.go`: the task-ID history is replaced by the given sentence, which also covers `Bodies`.
  - `clock/clock.go` "(M1-17)": cut.
  - `hexmap/chunk.go` "procgen in M5": cut.
  - `anatomy/install.go`: the tag comment now names `grantTag`.
  - `hexmap/step.go:12`: went with `DefaultStepRules`, and `StepRules` now gives its unit.
  - Left alone: the task IDs in `timeline/controllers.go:49` and `timeline/timeline.go:24,30`, and in `effects/effects.go`. C1-20's spec names those lines, and it is running in parallel, so leaving them avoids a merge conflict. The `SetSpeed` comment did change, because its signature did.

Behavior changes (the exceptions for `same-behavior`): a speed formula that fails at run time is now an error, not speed 0; `clock.Install` errors on a world with no content. Signatures changed: `clock.Install`, `timeline.SetSpeed`, `timeline.Speed`. A game whose formulas all work behaves the same: `go tool task same-behavior` passes with no exceptions (601 harness rows, 27 golden files).

Acceptance: I read the last criterion, "Only `_test.go` files and `testdata/` changed", as applying to the tests commit, like the second criterion. The task as a whole has to change code. `go tool task check` was run as `check-ci`; on the owner's machine it still needs the Windows fix in C1-24.

Files outside the area: `internal/core/actions/actions.go`, `internal/core/explore/install.go` and `internal/core/game/install.go` (call sites of the changed signatures), `internal/core/effects/effects_test.go` (the `clock.Install` call), and `internal/core/api/options_go_gen.cue`. That last one is new: `go tool task schema` generated it for the existing `options.go`, which master had never regenerated.

No follow-ups filed: the linters belong to C1-27, and the comments left alone belong to C1-20.

- Triaged 2026-11-23: todo in C1, as the owner decided; it is part of the first cleanup milestone.
