---
id: C1-20
title: "Actions area: duplication, tests, comments"
status: done
milestone: C1
kind: debt
depends_on: [C1-18]
discovered_by: null
size: L
---

# C1-20 · Actions area: duplication, tests, comments

## Goal

actions, effects and timeline are tidy, and their tests are fast.

## Context

- Milestone brief: [README.md](README.md). Findings come from the audit that opened C1.
- `docs/glossary.md` for settled terms.
- This is one pull request by design: the work touches the same packages and is done in one pass.

## Deliverables

- The changes described in the Spec, in one branch and one pull request.

## Spec

Do the parts in this order, one commit each, because tests follow the code they test and comments come last once code and tests have settled: duplication and libraries, then tests, then comments.

### actions, effects, timeline: shared channel helpers and one start path

Channel helpers and the start and miss paths each exist once.

Findings:

- `!c.Body.IsZero() && c.Entity != c.Body` at `actions/actions.go:271`, `effects/effects.go:501`, `effects/env.go:54`, `combat/controller.go:72`; `effects/env.go:52-59` `firstPart` is reimplemented at `actions/actions.go:269-275`; the channels-to-UIDs loop is at `effects/timing.go:38-44`, `actions/options.go:208-214`, `actions/actions.go:448-451`. Add `Channel.IsPart`, `FirstPart([]Channel)`, `UIDs(w, cs)` in `timeline`.
- `actions/actions.go:409-433` `Start` and `StartAutonomous` differ by one bool; `ActionMissed` and its line are built twice (`:520-528`, `:584-595`); the label's args map repeats (`:601-607`); `actions/install.go:118-121` `handleDo` calls `Validate` then `Start`, evaluating the intent twice. One private `startIntent(w, in, autonomous)`, one `ev.missed(...)`, one `ev.landArgs(...)`.
- `effects/effects_test.go:160` `do` and `triggers_test.go:29` `ref` are identical (covered by the fixture task if it lands there first; otherwise do it here).
- `actions/options.go:368-373` map copy becomes `maps.Clone` keeping the `len > 0` guard; `effects/effects.go:244-248` and `effects/timing.go:146-158` become `slices.Sorted(maps.Keys(...))`.
- Conventions: `effects/timing.go:65-67` `WindingUp` and `timeline/timeline.go:482-484` `CurrentPlayer` only forward to an unexported twin (`windingUp`, `currentPlayer` at `install.go:67`): keep one. Export `ModifierFunc` and use it in `AddModifier` and `state` (`actions/actions.go:43-46`, `install.go:41`). Move `effects.Hostile` (`effects.go:534`) next to `SetHostile` (`:182`).

### Actions, effects and timeline tests: load once, shrink the long ones

The package tests run in a second or two.

Findings:

- `TestOrderMatchesSort` (`timeline_test.go:579`, 3.77s): load the registry once with `sync.OnceValues`, `stopAfter` 500 → about 100; with `-rapid.checks=10` it takes 0.33s.
- `TestEveryOptionValidatesAndABusyChannelIsRejected` (`actions_test.go:448`, 1.50s): builds a full game per enabled option; build the registry once, or check each option with `actions.Validate` on the probe world and apply one.
- `TestSameSeedSameOrder` (`timeline_test.go:559`, 0.20s, 1000 decisions): shrink to about 50 or delete, since `TestOrderMatchesSort` pins the order. `TestReplayOfRandomOptionsGivesTheSameHash` (`actions_test.go:531`, 0.45s): 200 picks → about 50.
- Exact-wording asserts: `actions_test.go:525` `"Step: northwest (lands in 0, busy 10)"`, `:493` `"Not right now."`, `:497` `"busy 7"`: assert the option ID and text ID or args, and keep `TestLabelNamesTheChosenRecord`.

### Comments: actions, effects, timeline

About 15 comments are cut or fixed.

Findings:

- Task IDs: `effects/effects.go:67,181,185`, `timeline/controllers.go:49,70`, `timeline/timeline.go:24,30` (replace by what the code does, e.g. "anatomy adds one per attached part").
- `actions/export_test.go:15` history ("the per-channel test the batch hook replaced"); `actions/options.go:249-252` doc comment broken mid-reflow; `actions/actions.go:20,577` restate the name; `actions/actions.go:42` and `install.go:31` repeat "Combat uses it" (keep on `AddValidator`); `effects/doc.go:1` and `timeline/doc.go:3` "Foundation layer."; the effects package comment omits `Context`, selectors and the timing primitives; `timeline/timeline.go:367-374` `Run` doc is 8 lines, tighten to about 4.

The policy is the comments lens: doc comments say what a thing is for and what a caller must know, and every sentence carries information (cut name-restatements, filler, history, task IDs). Other comments are terse and say only why. CUE follows the same rules. The examples below set the standard; sweep the whole area.

## Acceptance criteria

- [x] `go tool task same-behavior` passes; the exceptions are the deliberate behavior changes named in the parts above, each listed in Notes with the changed output
- [x] The tests commit changed only `_test.go` files and `testdata/`
- [x] `go tool task check` passes (on the owner's machine too, which the Windows fix in the tools task needs)
- [x] Each finding above is resolved, one per line in Notes, or marked dropped with a reason
- [x] Only `_test.go` files and `testdata/` changed.

## Out of scope

- Other areas' findings: they have their own task in this milestone.
- Any new behavior beyond what the Spec names.

## Notes


- Triaged 2026-11-23: todo in C1, as the owner decided; it is part of the first cleanup milestone.

Three commits, as the Spec orders: code (`a49ecef`), tests (`2af736f`), comments; then one commit of review fixes (line length, one leftover comment, and the option test reworked; it touches code and tests together). No deliberate behavior change: `go tool task same-behavior` reports the same 601 harness rows and 27 goldens as master.

Findings, one per line:

- Channel helpers: `timeline.Channel.IsPart`, `timeline.FirstPart`, `timeline.UIDs` added; the four `IsPart` sites, both `firstPart` copies and the three channels-to-UIDs loops (`effects.uidsOf`, `actions.channelUIDs`, the loop in `start`) use them.
- `Start`/`StartAutonomous`: both call one private `startIntent(w, in, autonomous)`.
- `ActionMissed` built twice: one `ev.missed(actor, source, target, action, reason)`.
- Label args map: one `ev.landArgs(actor, action)`, used by `missed` and `labelLanding`.
- `handleDo` evaluated twice: `Validate` and `handleDo` share a private `ev.validate(in)`, and `handleDo` starts the plan it already worked out.
- `effects_test` `do` / `triggers_test` `ref`: already done by C1-18 (`contenttest.Do`); nothing left.
- `options.go` map copy: `maps.Clone`, keeping the `len > 0` guard.
- `effects.go` parameter names and `timing.go` interrupted actors: `slices.Sorted(maps.Keys(...))`.
- `WindingUp`/`windingUp`: one exported `effects.WindingUp`.
- `CurrentPlayer`/`currentPlayer`: one exported `timeline.CurrentPlayer`, in `install.go` where the unexported one was.
- `ModifierFunc` exported and used by `AddModifier` and `state.modifiers`.
- `effects.Hostile` moved next to `SetHostile`.
- `TestOrderMatchesSort`: the registry was already loaded once (`contenttest.Sample` uses `sync.OnceValues` since C1-18); `stopAfter` 500 → 100. 3.77s → 0.05s.
- `TestEveryOptionValidatesAndABusyChannelIsRejected`: checks each enabled option with `actions.Validate` on the probe world (through `actions.IntentOf`, exported in `export_test.go`) and applies one, instead of building a game per option. 1.50s → 0.13s.
- `TestSameSeedSameOrder`: 1000 → 50 decisions (kept, since it checks two runs agree, which `TestOrderMatchesSort` does not).
- `TestReplayOfRandomOptionsGivesTheSameHash`: 200 → 50 picks.
- Exact-wording asserts: the step and load labels, the validator's refusal and the gate's reason are compared with the text the core renders from their text IDs and args (new helper `optionLabel`). `TestLabelNamesTheChosenRecord` is unchanged.
- Task IDs in comments: replaced by who sets the hook (combat, explore, anatomy) or what the code does.
- `export_test.go` history, `options.go` broken reflow, `StepRules`/`landKind`/`resolveLanding` name restatements, `Validator`'s and `AddModifier`'s "Combat uses" (kept on `AddValidator`), "Foundation layer." in both doc.go files, the effects package comment (now covers `Context`, selectors, the timing primitives and triggers) and `Run`'s doc (8 lines → 4): all fixed.
- Sweep: also tightened the docs of `Context`, `Primitive`, `effects.Install` (it also subscribes the trigger runner) and `ModifierFunc`, and said why `startIntent` keeps its own check beside `validate`.

Deviations and things to check:

- `internal/core/combat/controller.go` (features area) changed by one line: `scoreOf` uses `timeline.FirstPart`, because the finding names that site.
- The last acceptance criterion ("Only `_test.go` files and `testdata/` changed.") contradicts the Deliverables, since the duplication and comments parts change code: it is a copy of the tests-commit criterion before it, made when the C1 tasks were generated. I ticked it as that criterion, which holds for `2af736f`; it cannot hold for the whole branch.
- `go tool task check` was run as `check-ci BASE=origin/master` (passes, on this Windows machine); the full `check` on Windows still needs the C1-24 tools fix.
- Package test times vary several-fold with machine load; run alone, `timeline` takes about 0.5s and `actions` and `effects` about 1.5s.
- At the owner's request, a later commit trims the comment edits so the branch shortens comments overall: in `.go` files it adds 38 comment lines (531 words) and removes 52 (720 words) against master, net -14 lines and -189 words.
- No follow-ups filed.
