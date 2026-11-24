---
id: C1-21
title: "Features area: duplication, tests, comments"
status: todo
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

Every comment `go tool task comments -- ./internal/core/combat ./internal/core/explore` lists is cut or fixed (69 when the rule tightened), along with the findings below.

Findings:

- `combat/doc.go:1-5` is stale ("build on them in later tasks"): it names none of hostility, alert, the hunt controller, death, the bodies view, or the effect primitives. `combat/rest.go:15-16` (M2-08), `explore/install.go:91` (M2), `combat/effects_test.go:51` (M7), `explore/explore_test.go:119-120` (false), `combat/property_test.go:33` names `propertyGame` but the function is `propertyConfig`, `:229` cites "the task".
- `explore/explore.go:351` → "terrainOf is the content record of a terrain, or the zero record when content lacks it." Narration to cut: `explore/project.go:120,146`; the doc comments on `installSides` (`combat/install.go:14`) and `installControl` (`explore/control.go:15`) list what the body registers; `explore` `Install` (`install.go:18`) names two view builders but not `explore.control`.

The policy is `.claude/rules/comments.md`: doc comments say what a caller needs, and every sentence carries information; other comments say only why; no end-of-line comments, no comments in tests other than doc comments, no task IDs. CUE follows the same rules. The examples below set the standard; sweep the whole area.

## Acceptance criteria

- [ ] `go tool task same-behavior` passes; the exceptions are the deliberate behavior changes named in the parts above, each listed in Notes with the changed output
- [ ] The tests commit changed only `_test.go` files and `testdata/`
- [ ] `go tool task check` passes (on the owner's machine too, which the Windows fix in the tools task needs)
- [ ] `go tool task comments -- ./internal/core/combat ./internal/core/explore` reports nothing
- [ ] Each finding above is resolved, one per line in Notes, or marked dropped with a reason
- [ ] Only `_test.go` files and `testdata/` changed.
- [ ] `go tool task test -- ./internal/core/combat` runs in under 10s.

## Out of scope

- Other areas' findings: they have their own task in this milestone.
- Any new behavior beyond what the Spec names.

## Notes


- Triaged 2026-11-23: todo in C1, as the owner decided; it is part of the first cleanup milestone.
