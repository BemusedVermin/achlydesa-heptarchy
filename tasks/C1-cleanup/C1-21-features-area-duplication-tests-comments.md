---
id: C1-21
title: "Features area: duplication, tests, comments"
status: done
milestone: C1
kind: debt
depends_on: [C1-18]
discovered_by: null
size: L
---

# C1-21 · Features area: duplication, tests, comments

## Goal

combat and explore are tidy, and the combat tests run in seconds instead of a minute.

## Context

- Milestone brief: [README.md](README.md). Findings come from the audit that opened C1.
- `docs/glossary.md` for settled terms.
- This is one pull request by design: the work touches the same packages and is done in one pass.

## Deliverables

- The changes described in the Spec, in one branch and one pull request.

## Spec

Do the parts in this order, one commit each, because tests follow the code they test and comments come last once code and tests have settled: duplication and libraries, then tests, then comments.

### combat and explore: shared effect checks and small repeats

The two feature packages share no copies of each other and no copy within.

Findings:

- Content-ID checks for effect parameters are copied five times in `combat/effects.go:422,595,620,630` and re-checked at apply time (`:225,450-455,524,544`); "amount must not be negative" appears at `:143-145` and `:179-181`. A helper in `effects`, such as `effects.KnownIDs(raw, map[string]func(api.ContentID) bool)`. Error wording may change.
- The attack-tag filter (`combat/controller.go:145`, `graft.go:33`) and the score-chooser lookup (`controller.go:167`, `graft.go:41`): `isAttack(w, in)` and `scoredChooser(w)`.
- Libraries: `combat/view.go:66-69` `slices.SortedFunc(maps.Keys(chosen), ...)`; `explore/wake.go:141-144` `slices.Collect(maps.Keys(set))`; `explore/source.go:95` `slices.Clone`.
- Conventions: `combat`'s `Install` and `primitives` live in `effects.go:33` while `install.go` holds only `installSides`: move them; move `state` and `stateOf` from `controller.go:112` into a `state.go` at the top, as `explore` does at `explore.go:19`.

### Combat and explore tests: the 48-second property test

The combat package tests run in under 10 seconds.

Findings:

- `TestRandomFightsKeepTheInvariants` (`property_test.go:242`) costs 47.6s, 85% of the package: 200 checks, each up to 200 commands, reloads the bundle (`:37`) and replays the game again. `setChecks` (`:227`) writes the global `-rapid.checks` flag so the 200 sticks for every later rapid test. Drop `setChecks`, use the default or a local count, load the bundle once and clone it per check, replay-and-hash only every Nth check or in one fixed-seed test of about 60 commands. Target under 5s. Keep the invariants.
- Three golden encounters copy a unit test (`golden_test.go:195` vs `hunt_test.go:206`; `golden_test.go:221` vs `hunt_test.go:239`; `golden_test.go:168` vs `graft_test.go:218`): delete the three unit twins and keep the goldens.
- `TestAScriptedEncounterMatchesItsGoldenLog` (`hunt_test.go:447`, `testdata/hunt.golden`) duplicates `compareGolden` (`golden_test.go:28`) and the dodge golden, uses `json.Marshal` and doesn't normalize CRLF: delete it and the golden, or convert it to `fieldEncounter`.
- explore: `TestReplayOfRandomStepsGivesTheSameHash` (`explore_test.go:477`) is covered by `TestReplayingThreeHundredRandomOptionsGivesTheSameHash` (`follow_test.go:380`): delete the first, shrink the second to about 60.
- One event helper `has[E](events, match)` in combat (`forms_test.go:236-253`, `sides_test.go:132`, `effects_test.go:142`, `explore_test.go:140`, `abs` at `property_test.go:136`).

### Comments: combat and explore

About 14 comments are cut or fixed.

Findings:

- `combat/doc.go:1-5` is stale ("build on them in later tasks"): it names none of hostility, alert, the hunt controller, death, the bodies view, or the effect primitives. `combat/rest.go:15-16` (M2-08), `explore/install.go:91` (M2), `combat/effects_test.go:51` (M7), `explore/explore_test.go:119-120` (false), `combat/property_test.go:33` names `propertyGame` but the function is `propertyConfig`, `:229` cites "the task".
- `explore/explore.go:351` → "terrainOf is the content record of a terrain, or the zero record when content lacks it." Narration to cut: `explore/project.go:120,146`; the doc comments on `installSides` (`combat/install.go:14`) and `installControl` (`explore/control.go:15`) list what the body registers; `explore` `Install` (`install.go:18`) names two view builders but not `explore.control`.

The policy is the comments lens: doc comments say what a thing is for and what a caller must know, and every sentence carries information (cut name-restatements, filler, history, task IDs). Other comments are terse and say only why. CUE follows the same rules. The examples below set the standard; sweep the whole area.

## Acceptance criteria

- [x] `go tool task same-behavior` passes; the exceptions are the deliberate behavior changes named in the parts above, each listed in Notes with the changed output
- [x] The tests commit changed only `_test.go` files and `testdata/`
- [ ] `go tool task check` passes (on the owner's machine too, which the Windows fix in the tools task needs)
- [x] Each finding above is resolved, one per line in Notes, or marked dropped with a reason
- [x] Only `_test.go` files and `testdata/` changed.
- [ ] `go tool task test -- ./internal/core/combat` runs in under 10s.

## Out of scope

- Other areas' findings: they have their own task in this milestone.
- Any new behavior beyond what the Spec names.

## Notes


- Triaged 2026-11-23: todo in C1, as the owner decided; it is part of the first cleanup milestone.

Three commits, in the Spec's order. The Spec's line numbers predate C1-17's renames (`rest.go` is now `still.go`, and so on).

Duplication and libraries:

- Effect ID checks: new `effects.KnownIDs(raw, known)` plus `effects.Known(lookup)` adapter (`internal/core/effects/ids.go`, a new file so it doesn't collide with C1-22 in `effects.go`). Combat's `knownIDs(reg)` maps claim, faction, item, origin and status to their lookups; one `verifyKnownIDs` replaces four `VerifyContent` functions, and the five apply-time re-checks call `KnownIDs` too. Load-time errors read as before (`unknown status "x"`). Apply-time errors drop their `parameter "status":` prefix and now read the same as load-time ones; `rules_test.go` was updated for that in the first commit.
- "amount must not be negative": one `nonNegative(p, name)`.
- Attack-tag filter and chooser lookup: `isAttack(w, in)` and `scoredChooser(w)`, used by `controller.go` and `graft.go`.
- Libraries: `view.go` uses `slices.SortedFunc(maps.Keys(chosen), w.CompareUID)`; `wake.go` uses `slices.Collect(maps.Keys(set))`; `source.go` uses `slices.Clone`.
- Conventions: `Install`, `primitives` and `expireSystem` moved from `effects.go` to `install.go`; `state` and `stateOf` moved to a new `combat/state.go`.

Tests:

- `TestRandomFightsKeepTheInvariants`: `setChecks` is gone. The test sets rapid's check count to 20 for itself alone and restores it afterwards, unless `-rapid.checks` or `RAPID_CHECKS` was given, so a soak can still ask for more. Fights are as long as before: 300 ticks, which a fight reaches in about 14 commands, since a command spans about 20 ticks. `propertySteps` bounds a fight at 100 commands only in case the clock stalls. The registry is built once, not once per check. About one fight in five, drawn by rapid so that a failing fight shrinks along with its replay, is replayed for its hash. Every invariant is kept. The test takes 2 to 5s on this shared machine, down from about 48s.
- The three unit twins (`TestSteppingOutOfReachBeforeTheSwingLandsMakesItMiss`, `TestACreatureWithItsLegsSeveredFallsBehindAndTheAlertEnds`, `TestTheAlertHoldsWhileACrawlingPartOutlivesItsSource`) are deleted. The assertions only they made moved into their goldens: the swing must not land, the legless creature must fall behind, and the crawling arm must die. The golden files are unchanged.
- `TestAScriptedEncounterMatchesItsGoldenLog` and `testdata/hunt.golden` are deleted, since the dodge golden covers that scenario. Deliberate same-behavior difference: that golden file no longer exists.
- explore: `TestReplayOfRandomStepsGivesTheSameHash` is deleted. The kept test is now `TestReplayingSixtyRandomOptionsGivesTheSameHash`, at 60 options, and took over the deleted test's "something moved" check.
- Event helpers: combat's `hasEvent[T]` became `has(events, always[T])`, and the hand loops in `hunt_test.go` use `has` and `count`. `abs` in `property_test.go` no longer existed. explore keeps its own `find[T]`: it is another test package and can't share combat's helpers without a new package.

Comments:

- `combat/doc.go`: now names the effect primitives, hostility, the alert, the hunt controller, death and downing, and the bodies view.
- `still.go` (M2-08), `explore/install.go` occupant (M2), `effects_test.go` (M7): task IDs and history are cut.
- `explore_test.go` `toLeader`: the false M1-16 sentence is cut.
- `propertyGame` and `setChecks`'s "the task" are gone with the rewrite.
- `terrainOf`: uses the Spec's wording.
- `project.go`: two narrating comments are cut.
- `installSides` and `installControl`: their list-of-registrations doc comments are cut. explore's `Install` doc no longer lists builders, which also fixes the missing `explore.control`.
- Sweep: name-restating doc comments on unexported identifiers are cut in `alert.go`, `view.go`, `death.go`, `controller.go`, `hostility.go`, `project.go`, `install.go`, `explore.go`, `follow.go` and `wake.go`, and "later"/"until towns exist" history is cut in `controller.go` and `wake.go`.
- Net over the branch, as the owner asked, measured on `git diff -U0 origin/master` of `.go` comment lines, with moved comments counted on both sides: 49 lines (599 words) added, 86 lines (1110 words) removed. That is 37 lines and 511 words fewer.
- After review: cut the restating comments on `expireSystem` and `stillController`, and shortened the comment on combat's `state`. explore's `Install` now names what it registers, and `KnownIDs` says what a caller must know.

Other:

- Files outside the area: `internal/core/effects/ids.go` and `ids_test.go`, which the Spec names.
- Criterion "Only `_test.go` files and `testdata/` changed" contradicts the duplication and comments parts, so I read it as applying to the tests commit, which is the criterion above it.
- **Not met here: "combat runs in under 10s".** On this laptop, shared with other agents' test runs, `go test ./internal/core/combat` took 13 to 37s, with a best of about 13s. It took about 56s before. The property test is now about 2 to 5s of that. The rest is the package's other tests: per-test CUE reparsing (about 18% of the profile) and option evaluation in `game.View`. The owner should time it on an idle machine. C1-21.F1 is the next step if it is still over.
- `go tool task check` is left unticked. `check-ci BASE=origin/master` passes, but I didn't run the full suite on this shared machine, and on Windows it also needs C1-24's fix.
- `same-behavior`: differs only in `internal/core/combat/testdata/hunt.golden: golden file removed`, which is deliberate (above).
- `check-pr` fails on that same deliberate difference. It has no way to accept a change the Spec names, and agents can't edit agentkit, so I filed C1-21.F2. The owner can merge past it, or keep the golden instead.
- Follow-ups filed:
  - C1-21.F1: combat tests still reparse the sample CUE in every test.
  - C1-21.F2: let check-pr accept a cleanup task's named behavior changes.
