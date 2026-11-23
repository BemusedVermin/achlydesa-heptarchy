---
id: C1-22
title: "Foundation area: duplication, tests, comments, speed errors"
status: todo
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

- [ ] `go tool task same-behavior` passes; the exceptions are the deliberate behavior changes named in the parts above, each listed in Notes with the changed output
- [ ] The tests commit changed only `_test.go` files and `testdata/`
- [ ] `go tool task check` passes (on the owner's machine too, which the Windows fix in the tools task needs)
- [ ] Each finding above is resolved, one per line in Notes, or marked dropped with a reason
- [ ] `go tool task same-behavior` passes: a game with working formulas is unchanged.
- [ ] Only `_test.go` files and `testdata/` changed.

## Out of scope

- Other areas' findings: they have their own task in this milestone.
- Any new behavior beyond what the Spec names.

## Notes


- Triaged 2026-11-23: todo in C1, as the owner decided; it is part of the first cleanup milestone.
