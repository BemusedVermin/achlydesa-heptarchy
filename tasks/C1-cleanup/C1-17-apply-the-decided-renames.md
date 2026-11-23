---
id: C1-17
title: "Apply the decided renames"
status: in-progress
milestone: C1
kind: debt
depends_on: [C1-01, C1-02, C1-03, C1-04, C1-05, C1-06, C1-07, C1-08, C1-09, C1-10, C1-11]
discovered_by: null
size: L
---

# C1-17 · Apply the decided renames

## Goal

Every contested name from the decided questions is renamed in one pass, so the rest of the cleanup starts from settled words.

## Context

- Milestone brief: [README.md](README.md). Findings come from the audit that opened C1.
- `docs/glossary.md` for settled terms.
- This is one pull request by design: the work touches the same packages and is done in one pass.

## Deliverables

- The changes described in the Spec, in one branch and one pull request.

## Spec

Do the renames **one after another, in the order below**, each with `gopls rename` for Go identifiers (never by hand) and a search for every content ID, text ID, golden and task file that quotes it. Run `go tool task same-behavior -- --rename old=new` after each. One PR, one commit per rename. If an answer makes a rename unnecessary (for example "leave as is"), skip it and say so in Notes.

### Rename the clay creature

The decided name is used everywhere, so a reader meets one word for one thing.

Findings:

- Vocabulary 1: three meanings of "crawler"; delete or relocate `body.crawler` and `part.crawler_*` (tests use them at `combat/forms_test.go:76`, `combat/effects_test.go:61`, `anatomy/anatomy_test.go:200`, `combat/rules_test.go:59,79`); regenerate `combat/testdata/hunt.golden` and the tui goldens.

Rename with `gopls rename` for Go identifiers, never by hand, and everywhere a content ID appears (CUE, text IDs, goldens, tests, docs and task files that quote them). The final names come from the decided question; the pairs below are the recommended ones.

- `template.creature.crawler` → `template.creature.moving_clay`
- `body.clay_thing` → `body.moving_clay`
- `part.clay_thing_*` → `part.moving_clay_*`

### Separate a graft from a reattached limb

The decided name is used everywhere, so a reader meets one word for one thing.

Findings:

- Vocabulary 2: `combat/effects.go:371-378`, `content/env.go:33,59`, `choices.cue:9`, `types.cue:39`.

Rename with `gopls rename` for Go identifiers, never by hand, and everywhere a content ID appears (CUE, text IDs, goldens, tests, docs and task files that quote them). The final names come from the decided question; the pairs below are the recommended ones.

- `comp.Graft (empty Piece)` → `comp.Reattached`

If the owner chooses "both are grafts", this task becomes obsolete and the design doc is amended instead.

### Fall means a drop; wipe means a recoverable loss

The decided name is used everywhere, so a reader meets one word for one thing.

Findings:

- Vocabulary 3 and 4: `combat/hostility.go:33-34` and its 6 callers, `explore/control.go:224`, `text.combat.body_downed`, `text.explore.fallen`, `party.cue:19`, `api/combat.go:127`, `game.go:159,188`.

Rename with `gopls rename` for Go identifiers, never by hand, and everywhere a content ID appears (CUE, text IDs, goldens, tests, docs and task files that quote them). The final names come from the decided question; the pairs below are the recommended ones.

- `fallen` → `outOfFight`
- `api.PartyLost` → `api.PartyWiped`
- `combat.party_lost` → `combat.party_wiped`
- `text.combat.party_lost` → `text.combat.party_wiped`

Show `text.game.lost` only when the party cannot wake.

### Name actors, targets and origins by their role

The decided name is used everywhere, so a reader meets one word for one thing.

Findings:

- Vocabulary 5: 9 `Actor` fields in `api/actions.go`, `timeline.go`, `explore.go`; 4 `Source` fields in `api/combat.go`; `content/env.go:29,44,70,79,85`; `forms.cue:36`.

Rename with `gopls rename` for Go identifiers, never by hand, and everywhere a content ID appears (CUE, text IDs, goldens, tests, docs and task files that quote them). The final names come from the decided question; the pairs below are the recommended ones.

- `Source (knowledge)` → `Origin`
- `source.sweep` → `origin.sweep`
- `content.Side` → `content.BodyEnv`
- `api.Fell.Actor` → `api.Fell.Body`

### Regular variable names in formula environments

The decided name is used everywhere, so a reader meets one word for one thing.

Findings:

- Vocabulary 6: `content/env.go` (EffectEnv :54, ScoreEnv :95, ActionEnv :69,:72, :90), `api/explore.go:122`, `rules.cue`, `combat.cue`, `types.cue`, `templates.cue`.

Rename with `gopls rename` for Go identifiers, never by hand, and everywhere a content ID appears (CUE, text IDs, goldens, tests, docs and task files that quote them). The final names come from the decided question; the pairs below are the recommended ones.

- `target_integrity` → `target.integrity`
- `immune_known` → `material_known`
- `api.Chosen` → `api.Params`

### Form, action, move and type in code and design

The decided name is used everywhere, so a reader meets one word for one thing.

Findings:

- Vocabulary 7: `templates.cue:17-22` lists eight IDs as "the seven forms"; tag the seven `form` in `forms.cue`; `comp.go:77` example `combat.move`; fix `docs/go-technical-design.md` lines 121, 177 and 188; Bundle comment "damage types".

Rename with `gopls rename` for Go identifiers, never by hand, and everywhere a content ID appears (CUE, text IDs, goldens, tests, docs and task files that quote them). The final names come from the decided question; the pairs below are the recommended ones.

- `_forms` → `_actions`
- `content.DamageType` → `content.Type`

### Rename the keep-still controller

The decided name is used everywhere, so a reader meets one word for one thing.

Findings:

- Vocabulary 8: `combat/rest.go:11`, `CombatConfig.RestTicks` (`content/types.go:250`), `combat.cue:11`.

Rename with `gopls rename` for Go identifiers, never by hand, and everywhere a content ID appears (CUE, text IDs, goldens, tests, docs and task files that quote them). The final names come from the decided question; the pairs below are the recommended ones.

- `rest (controller)` → `still`
- `combat.rest_ticks` → `combat.still_ticks`
- `restController` → `stillController`
- `rester` → `stiller`

### Call the party's carried items a pack everywhere

The decided name is used everywhere, so a reader meets one word for one thing.

Findings:

- Vocabulary 11: `templates.cue:117`, `comp/inventory.go:13`; the package `inventory` keeps its name.

Rename with `gopls rename` for Go identifiers, never by hand, and everywhere a content ID appears (CUE, text IDs, goldens, tests, docs and task files that quote them). The final names come from the decided question; the pairs below are the recommended ones.

- `comp.Inventory` → `comp.Pack`
- `content.Template.Inventory` → `content.Template.Pack`
- `inventory: (template field)` → `pack:`

### Turn, entry and pending in the timeline

The decided name is used everywhere, so a reader meets one word for one thing.

Findings:

- Vocabulary 9: `comp.go:76-79`, `timeline/save.go:21`, `timeline.go:82`, `sim/scheduler.go:14` (7 "per-turn" mentions), "turn queue" in `game/save.go:35` and `save/save.go:44`.

Rename with `gopls rename` for Go identifiers, never by hand, and everywhere a content ID appears (CUE, text IDs, goldens, tests, docs and task files that quote them). The final names come from the decided question; the pairs below are the recommended ones.

- `timeline.Saved.Pending` → `timeline.Saved.SystemsDue`
- `sim.PerTurn` → `sim.PerEntry`
- `comp.Pending.Part` → `comp.Pending.Channel`

### Rename the hostile faction

The decided name is used everywhere, so a reader meets one word for one thing.

Findings:

- Vocabulary 12: `factions.cue:19`, every template and test that names it.

Rename with `gopls rename` for Go identifiers, never by hand, and everywhere a content ID appears (CUE, text IDs, goldens, tests, docs and task files that quote them). The final names come from the decided question; the pairs below are the recommended ones.

- `faction.deep` → `faction.wild`

### One shape for text IDs

The decided name is used everywhere, so a reader meets one word for one thing.

Findings:

- Content V2 and vocabulary 12: move names beside their records, one `text/<package>.cue` per Go package, merge the two `text:` blocks in `text.cue`, update `.claude/rules/content.md`, and add a loader check that a record's `name` is `text.<id>` (parts exempt) with a test.

Rename with `gopls rename` for Go identifiers, never by hand, and everywhere a content ID appears (CUE, text IDs, goldens, tests, docs and task files that quote them). The final names come from the decided question; the pairs below are the recommended ones.

- `text.map.test_hall` → `text.map.test.hall`
- `text.actions.*` → `text.action.<event>`

## Acceptance criteria

- [ ] `go tool task same-behavior` passes (for each rename, with `-- --rename old=new`)
- [ ] `go tool task check` passes (on the owner's machine too, which the Windows fix in the tools task needs)
- [ ] Each finding above is resolved, one per line in Notes, or marked dropped with a reason
- [ ] `go tool task same-behavior -- --rename template.creature.crawler=template.creature.moving_clay` passes
- [ ] `go tool task same-behavior -- --rename body.clay_thing=body.moving_clay` passes
- [ ] `go tool task same-behavior -- --rename part.clay_thing_*=part.moving_clay_*` passes
- [ ] `go tool task same-behavior` shows only the behavior change named in the Spec: Relic is not offered after a plain reattach, and Rot does not eat a reattached limb. Update goldens
- [ ] `go tool task same-behavior` shows only the behavior change named in the Spec: player text and the wipe event name change; the rest is identical
- [ ] `go tool task same-behavior -- --rename Source (knowledge)=Origin` passes
- [ ] `go tool task same-behavior -- --rename source.sweep=origin.sweep` passes
- [ ] `go tool task same-behavior -- --rename content.Side=content.BodyEnv` passes
- [ ] `go tool task same-behavior -- --rename api.Fell.Actor=api.Fell.Body` passes
- [ ] `go tool task same-behavior -- --rename target_integrity=target.integrity` passes
- [ ] `go tool task same-behavior -- --rename immune_known=material_known` passes
- [ ] `go tool task same-behavior -- --rename api.Chosen=api.Params` passes
- [ ] `go tool task same-behavior -- --rename _forms=_actions` passes
- [ ] `go tool task same-behavior -- --rename content.DamageType=content.Type` passes
- [ ] `go tool task same-behavior -- --rename rest (controller)=still` passes
- [ ] `go tool task same-behavior -- --rename combat.rest_ticks=combat.still_ticks` passes
- [ ] `go tool task same-behavior -- --rename restController=stillController` passes
- [ ] `go tool task same-behavior -- --rename rester=stiller` passes
- [ ] `go tool task same-behavior -- --rename comp.Inventory=comp.Pack` passes
- [ ] `go tool task same-behavior -- --rename content.Template.Inventory=content.Template.Pack` passes
- [ ] `go tool task same-behavior -- --rename inventory: (template field)=pack:` passes
- [ ] `go tool task same-behavior -- --rename timeline.Saved.Pending=timeline.Saved.SystemsDue` passes
- [ ] `go tool task same-behavior -- --rename sim.PerTurn=sim.PerEntry` passes
- [ ] `go tool task same-behavior -- --rename comp.Pending.Part=comp.Pending.Channel` passes
- [ ] `go tool task same-behavior -- --rename faction.deep=faction.wild` passes
- [ ] `go tool task same-behavior -- --rename text.map.test_hall=text.map.test.hall` passes
- [ ] `go tool task same-behavior -- --rename text.actions.*=text.action.<event>` passes

## Out of scope

- Other areas' findings: they have their own task in this milestone.
- Any new behavior beyond what the Spec names.

## Notes


- Triaged 2026-11-23: todo in C1, as the owner decided; it is part of the first cleanup milestone.
