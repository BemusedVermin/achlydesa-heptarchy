---
id: C1-19
title: "Content area: duplication, tests, comments"
status: done
milestone: C1
kind: debt
depends_on: [C1-18]
discovered_by: null
size: L
---

# C1-19 · Content area: duplication, tests, comments

## Goal

The content loader and data are tidy, and its tests are fast.

## Context

- Milestone brief: [README.md](README.md). Findings come from the audit that opened C1.
- `docs/glossary.md` for settled terms.
- This is one pull request by design: the work touches the same packages and is done in one pass.

## Deliverables

- The changes described in the Spec, in one branch and one pull request.

## Spec

Do the parts in this order, one commit each, because tests follow the code they test and comments come last once code and tests have settled: duplication and libraries, then tests, then comments.

### content: one table of collections, one reference check, one formula runner

Adding a collection or a reference check touches one place, and a dangling reference reports once.

Findings:

- `checkIDs`, `recordCollections`, `dataOf`, `collectionIDs` and `RecordName` (`load.go:158-231`, `records.go:13-94,205-227`) list the 14 collections five times, with goconst hits on `"items"` and `"factions"`. One `collections` table drives all five (about 150 lines), iterating in today's order so error order holds.
- The "does this reference exist" check is written out 29 times (`actions.go:123`, `load.go:239,285,333,357,389`, `maps.go:76`, `types_check.go:60,185`, 20 `unknown <kind> %q` checks). **Bug:** material name text is checked in both `load.go:360-361` `checkTextRefs` and `types_check.go:63-64`, so one missing text reports twice (`TestLoadRejects` uses `Contains`, so it can't see it). Add a `problems` collector with `addf`, `text(path, ref)` and a generic `ref[V]`; delete `checkTextRefs`.
- Run-formula-and-check-type is copied 11 times (`registry.go:96-104,145-153,158-166,177-185,190-198`, `maps.go:37-45`, `choices.go:27-35`, `actions.go:42-66,78-98`, `types_check.go:206-212,221-228,250-257`): one `run[T any](program, env, what) (T, error)` in `compile.go`; delete `runAction` and `runMod`. The wording of the `returned %T` errors may change.
- CUE repeats: `bound: false` 8 times (`items.cue:6-14`), `grants: []` 5 times (`bodies.cue`), `immune_to: []` 5 times (`materials.cue:20-56`), Reaper and Salken (`templates.cue:25-48`), three hunters (`templates.cue:50-85`), four attack forms (`forms.cue:9-68`: `range: 1, sight: true, tags: ["attack"]`), `maps/test_*.cue` (`fill`, `open`, legend; and `maps: [string]: content.#MapDef` once like the other 13 collections). Add `omitempty` to `Item.Bound` and `PartTemplate.Grants` (then `go tool task schema`), a default for `immune_to`, hidden definitions beside `#Consumable` (`types.cue:9`).
- Libraries: `TerrainIDs` (`registry.go:78-85`), `MapIDs` (`maps.go:20-27`), `ActionIDs` (`actions.go:33-40`) become calls to `sortedIDs` (`types_check.go:45`); `sortedKeys` (`load.go:138-145`) becomes `slices.Sorted(maps.Keys(m))`; `recordOf` props loop (`records.go:103-105`) becomes `maps.Clone`; `Calendar` (`registry.go:110`) and `Warnings` (`types_check.go:43`) become `slices.Clone`; `watch.go:549-552` stop-and-drain is unneeded since Go 1.23.
- Conventions: 12 of 20 CUE files are not `cue fmt`-clean (`actions`, `combat`, `factions`, `items`, `knowledge`, `materials`, `party`, `rules`, `templates`, `text`, `timeline`, `types`): run `go tool cue fmt` once. Exported package-level slices `TargetKinds` (`types.go:337`), `TriggerRoles` (`:204`), `ValueKinds` (`:100`) become functions or unexported slices; `ruleVar` (`compile.go:14`) becomes a `const`. The `cue fmt --check` step in the `content` task changes `Taskfile.yml` and is part of the lock-in.

Lint: `gochecknoglobals` scoped to `internal/` (for the slices), modernize (for the stdlib items). Both are tried at lock-in.

### Content tests: load once, assert less

The content tests are fast and don't break when content is tuned.

Findings:

- `TestLoadRejectsBadActions` (`actions_test.go:45`, 0.95s, 12 loads) and `TestBodiesRejectBadReferences` (`bodies_test.go:177`, 0.95s, 11 loads): keep the cases, deep-copy a loaded bundle per case. Target under 0.15s each.
- `TestWatchOneCallbackPerBurst` (`watch_test.go:16`, 1.63s) and `TestWatchIgnoresOtherFiles` (`:55`, 0.75s) use `time.Sleep`: make the debounce a parameter (an unexported `watch(dir, debounce, onChange)` behind `Watch`), shrink to one burst; under 0.3s total.
- Assertions pinned to sample wording and tuning: `load_test.go:67` exact `"Packed clay. No way through."`, `:64` exactly two terrains, `:100` step ticks 25/15/10, `bodies_test.go:164` five parts, `rules_test.go:244` `type.soak` consumes exactly 1. Move formula checks to `validData`; against the sample assert presence and that text renders without `[missing text:`.
- The error-fragment loop is copied 5 times (`load_test.go:280`, `:86`, `rules_test.go:128`, `actions_test.go:133`, `bodies_test.go:245`): `requireErrContains(t, err, fragments...)`; `FuzzLoad` (`load_test.go:340`) copies `writeData`: make it take `testing.TB`.
- `TestCUEConstraintsHoldAgainstTheRealSchema` (`schema_test.go:47`, 0.82s) string-replaces an exact line of `materials.cue`: write an override file into the copied tree instead.

### Comments: content

About 21 comments in content are cut or fixed.

Findings:

- `load.go:66` doc for `newRegistry` sits on `cueFiles` (move it; `:91` has none). `content/data/actions.cue:9` says the step effect arrives with M1-10, which is false: cut. `registry.go:14` "Build one with [Load]" → "[Load] or [New]".
- History and task IDs to cut: `types.go:8`, `:192`, `:351`, `env.go:35`, `:62`, `types.cue:29`, `bodies.cue:89`, `templates.cue:88`.
- Go doc comments that repeat tuning values (the generated schema reference): `types.go:222-230,238` ("4 to start", "6 to start", "12 to start", "1, 1 and 8 to start", "True to start"), `CombatConfig` `types.go:246` "Placeholders to tune."
- Missing the required text-ID comment (`.claude/rules/content.md`): `Terrain.Name` (`types.go:150`), `Start.Map` (`types.go:299`).

The policy is the comments lens: doc comments say what a thing is for and what a caller must know, and every sentence carries information (cut name-restatements, filler, history, task IDs). Other comments are terse and say only why. CUE follows the same rules. The examples below set the standard; sweep the whole area.

## Acceptance criteria

- [x] `go tool task same-behavior` passes; the exceptions are the deliberate behavior changes named in the parts above, each listed in Notes with the changed output
- [x] The tests commit changed only `_test.go` files and `testdata/`
- [x] `go tool task check` passes (on the owner's machine too, which the Windows fix in the tools task needs)
- [x] Each finding above is resolved, one per line in Notes, or marked dropped with a reason
- [x] A test shows one dangling text reference reports exactly one error.
- [x] `go tool cue fmt --check ./content/...` reports nothing.
- [x] Only `_test.go` files and `testdata/` changed.

## Out of scope

- Other areas' findings: they have their own task in this milestone.
- Any new behavior beyond what the Spec names.

## Notes


- Triaged 2026-11-23: todo in C1, as the owner decided; it is part of the first cleanup milestone.

### Findings, one per line

Duplication and libraries (commit "one collections table, one reference check, one formula runner"):

- Collections table: done. `collections()` in `records.go` drives `checkIDs`, `dataOf`, `collectionIDs`, `RecordName`, `buildRecords` and `checkTextNames`; `recordCollections` is gone. It is a function, not a package var (gochecknoglobals), and the registry builds it once (`Registry.collections`) because `RecordName` runs per option.
- Reference checks: done. `problems.go` holds the collector (`addf`, `add`, `text`, generic `ref[V]`); every `unknown <kind> %q` and text reference check in the loader goes through it. `checkTextRefs` is deleted; its terrain-name check moved into `checkTerrains` (the renamed `compileFormulas`).
- Duplicate-error bug: fixed. A missing material name text now reports once, from `checkTypes`. `TestADanglingTextReferenceReportsOnce` shows it.
- Formula runner: done. `run[T]` in `compile.go` serves all 11 sites; `runAction` and `runMod` are deleted. The `returned %T` errors now read "returned X, not int64/bool/float64".
- CUE repeats: done. `bound: false` and `grants: []` dropped and `immune_to` defaults to `[]`. **Deviation:** `bound` and `grants` get a CUE default (`items: [string]: content.#Item & {bound: *false | bool}`, `parts: [string]: content.#PartTemplate & {grants: *[] | [...]}`) instead of `omitempty` on `Item.Bound` and `PartTemplate.Grants`. `omitempty` changed the bundle's JSON, so the content hash and every final game hash, and `check-pr` refuses a cleanup whose same-behavior check fails, with no way to declare an exception. The defaults remove the same repeats and leave the decoded bundle byte for byte as it was. Switching to `omitempty` later is one tag each plus `go tool task schema`; `_#Person` (Reaper, Salken) and `_#Hunter` (three hunters) in `templates.cue`, `_#Attack` in `forms.cue`, `_#TestMap` plus the single `maps: [string]: content.#MapDef` in the new `maps/maps.cue`. The hidden definitions sit beside the records that use them rather than in `types.cue`.
- Libraries: done. `TerrainIDs`, `MapIDs`, `ActionIDs` call `sortedIDs`; `sortedKeys` is `slices.Sorted(maps.Keys(m))` (kept as a one-line helper, since about 40 call sites use it); `recordOf` uses `maps.Clone`; `Calendar` and `Warnings` use `slices.Clone`; the watch timer no longer drains on stop.
- `cue fmt`: nothing to do. `go tool cue fmt --check ./content/...` was already clean on master (C1-17 must have formatted it), and it still is.
- Exported slices: `TargetKinds`, `TriggerRoles`, `ValueKinds` are now unexported functions (`targetKinds()` ...); nothing outside the package used them. `ruleVar` is a const, which makes `cue get go` emit `compile_go_gen.cue` with `_#ruleVar`; it is committed.
- Taskfile `cue fmt --check` step and the gochecknoglobals/modernize lint: left for the lock-in (C1-27), as the Spec says; this branch may not edit `Taskfile.yml` or `.golangci.yml`.

Tests (commit "content tests load once and assert less", `_test.go` files only):

- `TestLoadRejectsBadActions` and `TestBodiesRejectBadReferences`: the sample bundle is loaded once per binary and each case decodes a fresh copy from its JSON (`export_test.go`). About 0.04s and 0.03s.
- Watch tests: `Watch` calls an unexported `watch(dir, debounce, onChange)` (that part is in the code commit); the tests use a 40 ms debounce and one burst. About 0.15s and 0.12s. After review the debounce went from 30 to 40 ms against a slow Windows runner; if the one-callback test ever flakes, raise `testDebounce` first.
- Pinned assertions: the exact bump text, the exact terrain count, the five parts and "consumes exactly 1" now assert presence and that text renders without `[missing text:`; the step-tick numbers and default speed are checked against `validData`.
- `requireErrContains(t, err, fragments...)` replaces the 5 fragment loops; `writeData` takes `testing.TB` and `FuzzLoad` uses it.
- `TestCUEConstraintsHoldAgainstTheRealSchema` writes an `override.cue` with a new material that is weak and immune to one type, so it no longer depends on any line of `materials.cue`. It still takes about 0.7s, which is the copy of the CUE tree and two loads.

Comments (commit "content comments say what callers need"):

- `newRegistry`'s doc is on `newRegistry`; `actions.cue` M1-10 line cut; `Registry` says "[Load] or [New]".
- History and task IDs cut: `Bundle` "Later tasks add fields", `Template` "Later: Combatant, Needs", `ActionDef.Choices` "(M2)", `EffectEnv` and `ActionEnv` "deliberately", `types.cue` (M2-05), `bodies.cue` (M2-11). "for M7's tuning" became "for tuning" in `types.cue`, `templates.cue`, `items.cue` and `forms.cue`. The `templates.cue:88` line number no longer matched anything with history after C1-17; the whole file was swept.
- Tuning values cut from the `PartyConfig` docs and "Placeholders to tune." from `CombatConfig`; the generated schemas follow.
- `Terrain.Name` is commented `// a text ID`; `Start.Map` `// a map ID` (it holds a map ID, not text).

### Deliberate behavior changes

- `go tool task same-behavior` passes outright: 601 harness rows (3 seeds, 200 steps each) and 27 golden files, with equal final hashes. (A first version with `omitempty` changed the hashes; see the CUE repeats line.)
- Error order inside one joined load error: `checkTextNames` now reports in table order (materials, terrains, templates, actions, maps, claims, origins, types, items, factions, statuses), terrain-name text errors come from `checkTerrains` near the end, and material-name text errors only from `checkTypes`. `checkParty` now reports its non-negative numbers in name order; it ranged over a map before, so that order was random. `checkIDs` keeps today's order.

### Other

- The code commit also changed two strings in `schema_test.go`, so that commit's tests stay green after the `immune_to` default changed `materials.cue`; the tests commit then rewrote that test.
- `go tool task check` passes on this Windows machine, and so does `check-ci`.
- The last criterion, "Only `_test.go` files and `testdata/` changed", reads as a repeat of the tests-commit criterion: it holds for the tests commit, not for the branch, which by design changes code, CUE and schemas too.
- Follow-up filed: C1-19.F1 (`go tool task schema` on master writes an uncommitted `internal/core/api/options_go_gen.cue`, left by C1-18; this branch deleted the generated file rather than change `api`).
