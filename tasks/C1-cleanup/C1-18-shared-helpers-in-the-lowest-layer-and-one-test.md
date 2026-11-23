---
id: C1-18
title: "Shared helpers in the lowest layer, and one test fixture"
status: in-progress
milestone: C1
kind: debt
depends_on: [C1-17, C1-13]
discovered_by: null
size: L
---

# C1-18 · Shared helpers in the lowest layer, and one test fixture

## Goal

Code that several areas wrote their own copy of lives once, at the lowest layer both sides may import, and tests share one fixture.

## Context

- Milestone brief: [README.md](README.md). Findings come from the audit that opened C1.
- `docs/glossary.md` for settled terms.
- This is one pull request by design: the work touches the same packages and is done in one pass.

## Deliverables

- The changes described in the Spec, in one branch and one pull request.

## Spec

Each part is independent of the others except where its text says so. One PR, one commit per part. Lowest layers first.

### sim: one UID order, one upsert, one liveness check

The core's most common idioms each have one shared form in `sim`.

Findings:

- UID sorting is written about 9 ways: three identical three-way switches (`anatomy/query.go:13` `byUID`, `actions/options.go:165` `cmpUID`, `effects/selectors.go:202` `compareUID`), `timeline/pending.go:93` `cmpInt`, `timeline/timeline.go:144-149`, `timeline/install.go:124-136` (which does `int(a.tick - b.tick)`), `effects/triggers.go:124-132`, and subtraction comparators that wrap past 2^63: `combat/alert.go:43`, `combat/effects.go:328`, `combat/view.go:69`, `explore/control.go:127`, plus the threat sort `combat/view.go:95-99` and `combat/effects.go:261-269` `compareIDs`.
- Four identical liveness helpers (`actions/actions.go:51`, `combat/hostility.go:31`, `effects/effects.go:457`, `explore/explore.go:112`) and eight inline `e.IsZero() || !w.ECS.Alive(e)` checks (`anatomy/query.go:29,42,51,74`, `inventory/inventory.go:83,103`, `combat/graft.go:20`, `combat/effects.go:439`). Ark's `Alive` is already false for the zero entity.
- Upsert a component (`if Has { *Get = v } else { Add }`) is open-coded five times: `combat/effects.go:115`, `combat/death.go:124`, `explore/wake.go:19,81,87`.
- Collect-then-sort over map keys at `sim/commands.go:66-71`, `sim/order.go:42,60`; relation-target loop at `sim/spawn.go:103` and `anatomy/query.go:33`.

Add to `internal/core/sim` (suggestions): `func (w *World) SortByUID(entities []ecs.Entity) []ecs.Entity`, `func (w *World) CompareUID(a, b ecs.Entity) int` (next to `UIDOf` in `spawn.go`), `func (w *World) Alive(e ecs.Entity) bool`, `func Put[T any](w *World, e ecs.Entity, v T)`. Replace every site above with `cmp.Compare`, `cmp.Or` and `slices.Sorted(maps.Keys(...))` (as `timeline/save.go:45` already does). Add one test in `sim` that pins `Alive(ecs.Entity{}) == false`, in case Ark changes.

### anatomy: party and leader queries in one place

Feature packages ask the party questions through `anatomy`, and a leader change goes through one function.

Findings:

- `controlledParty` is identical in `combat/alert.go:19` and `explore/explore.go:46`. `controlled` in `knowledge/knowledge.go:93`, `partyControlled` in `combat/controller.go:97` and `HolderOf` in `inventory/inventory.go:85` ask the same questions.
- `explore/explore.go:60` `membersOf` repeats `anatomy.Members` (`anatomy/query.go:78`) before its own sort. `explore/explore.go:100-110` `partyOf` scans every `MemberOf` entity on the step path; `combat/death.go:86` already uses `ecs.NewMap[comp.MemberOf](...).GetRelation(e)`.
- `fallen` (now `outOfFight`) is copied at `explore/control.go:224` and `combat/hostility.go:34`.
- Leader handover is written three ways: `combat/death.go:91-105` `handOverLead` and `explore/control.go:105-114` `handleLead` reorder `Order` and publish `LeaderChanged`; `explore/wake.go:120-131` `leadFromStanding` sets `Leader` only, so after a wake `Order` still starts with the dead member and nothing is published (a latent bug).
- `combat/alert.go:62-69` rebuilds the visible-cell set that `seenCells` (`combat/view.go:24`) returns.

Add to `anatomy` (suggestions): `ControlledParty(w) (ecs.Entity, bool)`, `PartyOf(w, body) (ecs.Entity, bool)`, `InControlledParty(w, e) bool`, `OutOfFight(w, body) bool`, `SetLeader(w, party, uid)` (moves the UID to the front of `Order`, sets `Leader`, publishes `api.LeaderChanged`). `firstHostileSeen` calls `seenCells`. **One deliberate behavior change:** after a wake, `Order` starts with the new leader and `LeaderChanged` is published. Say so in Notes and update the golden.

### content: one place builds a template's name text ID

The rule "template name text is `text.template.<id>`, with `text.template.unknown` as the fallback" lives in one place.

Findings:

- Copied at `actions/actions.go:530-539` `nameOf`, `timeline/install.go:149-154`, `explore/explore.go:198-206` `templateName`, `combat/hostility.go:39-46` `nameOf`.

Add `func (r *Registry) TemplateName(id api.ContentID) api.Text` in `internal/content` (suggestion). Callers look up the body's template ID themselves. Delete the four copies.

### api: constants for entry kinds and option IDs, one reject constructor, one codec table

Strings that both the core and a frontend must agree on are named once in `internal/core/api`.

Findings:

- Timeline entry kinds `"ready"` and `"lands"` are bare literals in `timeline/install.go:144,174,176`, `timeline/pending.go:174`, `frontend/tui/render.go:247,253,256,264`, `save/fight_test.go:22`; `api/timeline.go:49` documents them only in a comment. Add `api.EntryReady` and `api.EntryLands` (suggested names).
- Option ID prefixes `step/`, `lead/<uid>`, `control/`, `wait/event` are built in `explore/control.go:192-211` and `timeline/install.go:196` and parsed with `strings.Split` and UID parsing in the tui (`keys.go:50,79,126-150`, `menu.go:18,34,113,202`, `look.go:16`, `app.go:117`). Add prefix constants and a builder or parser beside `api.ControlFollow`.
- `explore/control.go:43` `rejectCommand(kind, reason string)` and `timeline/install.go:76` `reject(kind, reason api.ContentID)` differ only in a parameter type: add `api.Reject(kind string, reason api.ContentID) api.Rejected`.
- `api/codec.go:56-84` and `:110-150`: `RegisterCommand`/`RegisterEvent` and `DecodeCommand`/`DecodeEvent` are the same code twice. Use a private generic `kindTable[T any]`; keep the error strings.

The boundary stays plain data: constants and constructors only, no behavior beyond building a value.

### hexmap: one cell compare, one distance, one direction finder

Cell geometry is derived in `hexmap` only.

Findings:

- `compareCells` in `hexmap/map.go:81` (subtraction) and `explore/explore.go:41` (`cmp.Or`); export `hexmap.CompareCells`.
- `abs` copies: `hexmap/path.go:119`, `combat/controller.go:200`, `explore/explore.go:344`, and two test files. "Hex distance plus depth difference" at `combat/controller.go:192` and `hexmap/path.go:29-30`: add `hexmap.CellDistance(a, b api.Cell) int`.
- `hexmap/sight.go:119` `Distance` re-implements `hex.Point.DistanceTo`, which `path.go:29` uses. `floorDiv`, `floorMod` (`chunk.go:28,36`) and `floorDivInt` (`sight.go:110`) collapse to one.
- The six-direction "Step, skip non-local" loop appears twice at `hexmap/path.go:59-76` and `:95-112`: a private `localSteps(cell) iter.Seq[...]`.
- `explore/follow.go:84-99` `dirTo` and the loop in `combat/controller.go:222-233` `stepToward`: add `(*hexmap.Map).DirTo(from, to api.Cell, rules StepRules)`. `hexmap.StepRules{MaxStepUp: cfg.MaxStepUp, ...}` is built from content in `explore/explore.go:34` and `combat/controller.go:217`: expose one accessor both can import (`effects` or `actions`, since hexmap doesn't import content).
- Direction loops `for dir := api.Dir(0); dir < api.DirCount; dir++` at `explore/explore.go:327` and `follow.go:89` versus `range` at `project.go:201`; `shadowed builtin clear` at `hexmap/sight.go:46`.

### A shared sample-content test fixture

A test gets the sample content, a game and a mutated bundle from one support package instead of its own copy.

Findings:

- `sampleDir`/`contentDir` is defined 15 times (10 in the content audit, 5 in game audit), with `LoadBundle` + `content.New` + `t.Fatal` beside it; the sample loads 37 times in 28 test files.
- The `game.Config{Install: func(w *sim.World, _ *clock.Clock) error { world = w; ... }}` helper is copied into 13 test files (`anatomy_test.go:31`, `inventory_test.go:183`, `knowledge_test.go:193`, ...).
- "Keep the crawler standing" is copied in `save_test.go:44-48`, `harness_test.go:25-29`, `tui/combat_test.go:43-47`, `tui/render_test.go:36-40` and 3 places in `combat/hunt_test.go`; "party can only step" in `save_test.go:36-41` and `tui/render_test.go:29-34`.
- Per-package rig helpers (`rig` effects_test.go:40, `ring` forms_test.go:51, `field` sides_test.go:30, `fight` property_test.go:80, `fixture` explore_test.go:31) and the `do`/`ref` pair in `effects` (`effects_test.go:160`, `triggers_test.go:29`, `combat/effects_test.go:101`).

Add `internal/content/contenttest` and `internal/core/game/gametest` (suggestions). `Sample(t testing.TB) *content.Registry` (cached per the decision), `SampleBundle(t) content.Bundle` (a fresh deep copy), `Edit[V any](m map[string]V, key string, f func(*V))`, named tweaks `StandingCrawler`, `StepOnlyParty`, `NoStart` (and its post-rename name), `gametest.New(t, seed, mutate)` returning the game, world and config, and `Reload`. Packages inside `internal/content` that cannot import it use an `export_test.go` twin. Test-only imports pass depguard. Convert the sites above.

## Acceptance criteria

- [ ] `go tool task same-behavior` passes; the exceptions are the deliberate behavior changes named in the parts above, each listed in Notes with the changed output
- [ ] `go tool task check` passes (on the owner's machine too, which the Windows fix in the tools task needs)
- [ ] Each finding above is resolved, one per line in Notes, or marked dropped with a reason
- [ ] No hand-written three-way compare of UIDs or ticks remains: `grep` for `int(w.UIDOf(` finds nothing.
- [ ] A test shows that after a wake `Order[0]` is the new leader and `LeaderChanged` was published.
- [ ] Every `sampleDir` and `contentDir` constant is gone from tests.

## Out of scope

- Other areas' findings: they have their own task in this milestone.
- Any new behavior beyond what the Spec names.

## Notes


- Triaged 2026-11-23: todo in C1, as the owner decided; it is part of the first cleanup milestone.
