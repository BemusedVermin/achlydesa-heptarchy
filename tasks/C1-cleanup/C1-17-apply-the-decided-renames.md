---
id: C1-17
title: "Apply the decided renames"
status: done
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
- [x] Each finding above is resolved, one per line in Notes, or marked dropped with a reason
- [x] `go tool task same-behavior -- --rename template.creature.crawler=template.creature.moving_clay` passes
- [x] `go tool task same-behavior -- --rename body.clay_thing=body.moving_clay` passes
- [x] `go tool task same-behavior -- --rename part.clay_thing_*=part.moving_clay_*` passes
- [x] `go tool task same-behavior` shows only the behavior change named in the Spec: Relic is not offered after a plain reattach, and Rot does not eat a reattached limb. Update goldens
- [x] `go tool task same-behavior` shows only the behavior change named in the Spec: player text and the wipe event name change; the rest is identical
- [x] `go tool task same-behavior -- --rename Source (knowledge)=Origin` passes
- [x] `go tool task same-behavior -- --rename source.sweep=origin.sweep` passes
- [x] `go tool task same-behavior -- --rename content.Side=content.BodyEnv` passes
- [x] `go tool task same-behavior -- --rename api.Fell.Actor=api.Fell.Body` passes
- [x] `go tool task same-behavior -- --rename target_integrity=target.integrity` passes
- [x] `go tool task same-behavior -- --rename immune_known=material_known` passes
- [x] `go tool task same-behavior -- --rename api.Chosen=api.Params` passes
- [x] `go tool task same-behavior -- --rename _forms=_actions` passes
- [x] `go tool task same-behavior -- --rename content.DamageType=content.Type` passes
- [x] `go tool task same-behavior -- --rename rest (controller)=still` passes
- [x] `go tool task same-behavior -- --rename combat.rest_ticks=combat.still_ticks` passes
- [x] `go tool task same-behavior -- --rename restController=stillController` passes
- [x] `go tool task same-behavior -- --rename rester=stiller` passes
- [ ] `go tool task same-behavior -- --rename comp.Inventory=comp.Pack` passes
- [ ] `go tool task same-behavior -- --rename content.Template.Inventory=content.Template.Pack` passes
- [ ] `go tool task same-behavior -- --rename inventory: (template field)=pack:` passes
- [x] `go tool task same-behavior -- --rename timeline.Saved.Pending=timeline.Saved.SystemsDue` passes
- [x] `go tool task same-behavior -- --rename sim.PerTurn=sim.PerEntry` passes
- [x] `go tool task same-behavior -- --rename comp.Pending.Part=comp.Pending.Channel` passes
- [x] `go tool task same-behavior -- --rename faction.deep=faction.wild` passes
- [x] `go tool task same-behavior -- --rename text.map.test_hall=text.map.test.hall` passes
- [ ] `go tool task same-behavior -- --rename text.actions.*=text.action.<event>` passes

## Out of scope

- Other areas' findings: they have their own task in this milestone.
- Any new behavior beyond what the Spec names.

## Notes


- Triaged 2026-11-23: todo in C1, as the owner decided; it is part of the first cleanup milestone.

### What was done (six commits for the eleven groups: moving clay; graft; down; actor; formulas; then still, form, timeline, faction, text IDs and inventory together)

- **Clay.** `template.creature.crawler`, `body.clay_thing`, `part.clay_*` are now `moving_clay` (the parts are `part.moving_clay_{maw,legs,core}`; the task text said `part.clay_thing_*` but the real IDs were `part.clay_*`). The unused flesh `body.crawler` and its parts are kept as test fixtures: `body.test.crawler`, `part.test.crawler_*`. Player text "The crawler" became "The moving clay".
- **Graft.** `comp.Graft` is now only a fitted piece (it always has a `Piece`). A reattached limb gets the new marker `comp.Reattached`. Behavior change: Relic is not offered, and Rot does not eat, after a plain reattach. No existing golden exercised it; `TestAReattachedLimbIsNotAGraftForRelicOrRot` shows that a reattached limb counts as no graft (so Relic's `source.grafts` check does not see it) and that Rot does not eat it.
- **Harness.** The harness outcome `lost` (set on `PartyWiped`, the recoverable wipe) is now `OutcomeWiped = "wiped"`; "lost" is reserved for the unrecoverable end.
- **Down.** `fallen` is `outOfFight` (combat and explore), `PartyLost` is `PartyWiped` (type, kind `combat.party_wiped`, text `text.combat.party_wiped`), `text.explore.fallen` is `text.explore.out_of_fight`. Texts reworded (plain defaults, the owner can change them): "{{.name}} is out of the fight.", "The party is wiped out.", "They cannot lead: they are out of the fight.". `text.game.lost` is shown only when `PartyWiped` holds, and `Apply` wakes a wakeable party at once, so it already shows only when the party cannot wake; I left it as it was.
- **Actor.** Nine event and command fields `Actor` are now `Source` (JSON `source`), `api.Fell.Actor` is `Body`. Knowledge `Source` is `Origin` (`api.Fact.Origin`, `content.Origin`, bundle key `origins`, `origin.sweep`, `text.origin.*`, and the `reveal` effect's `origin` parameter). `content.Side` is `content.BodyEnv` and `effects.SideOf` is `effects.BodyEnvOf`. `comp.Pending.Actor` was not in the decision and is unchanged.
- **Formulas.** Target facts are under `target.` (`target.integrity`, `.max_integrity`, `.locked`, `.graft`, `.statuses`, `.winding_up`, `.is_party`), as fields of `BodyEnv`. `immune_known` is `material_known`, `choice` is dropped (nothing used it), `api.Chosen` is `api.Params` (method `Params()`).
- **Form.** `_forms` is `_actions` (it holds the seven forms and a step); the seven forms are tagged `form`. `content.DamageType` is `content.Type`. Fixed the doc comment in `comp.go` and the lines in `docs/go-technical-design.md`.
- **Still.** `rest` controller is `still` (`still.go`, `stillController`, `stiller`, `still_ticks`, `StillTicks`).
- **Inventory (deviation from the spec text).** The owner decided `inventory`, not the spec's recommended `pack` (C1-08), so I did **not** rename `comp.Inventory` to `comp.Pack`, `Template.Inventory` or `inventory:`. Instead everything that said "pack" is now `inventory`: the formula variable (`inventory[...]`), `from: "inventory"`, `api.AttachedFromInventory`, `text.combat.inventory_short`, the `Inventory` env fields, and comments. The three `Pack` acceptance criteria are therefore left unticked: they describe the rejected option.
- **Timeline.** `timeline.Saved.Pending` is `SystemsDue` (JSON `systems_due`), the timeline's own flag is `systemsDue`, `sim.PerTurn` is `PerEntry`, `comp.Pending.Part` is `Channel` (JSON `channel`), "turn queue" and "per-turn" are gone from code and docs. Save version bumped to 2 (old saves fail cleanly).
- **Faction.** `faction.deep` is `faction.wild`, text "the wild".
- **Text IDs.** Names sit beside their records; engine messages are one file per Go package in `content/data/text/` (`actions`, `clock`, `combat`, `effects`, `explore`, `game`, `knowledge`, `timeline`); `content/data/text.cue` now holds only the schema. `text.map.test_*` is `text.map.test.*`. The spec pair `text.actions.*` to `text.action.<event>` was **skipped** on purpose: the decision (and the glossary) says an engine message is `text.<package>.<event>`, and the package is `actions`, so those IDs already have the decided shape, and `text.action.*` would have shared a namespace with the action records' names. The criterion for it is left unticked. New loader check `checkTextNames` (parts exempt) with `TestARecordsNameIsItsOwnIDAsText`. Test fixtures that cloned a record under a new ID now name it correctly.

### Deviations and things to look at

- The `same-behavior` ticks rest on **one combined run** with every declared rename in order (a cumulative list, `"actor":` handled with a placeholder so only event fields change), not on one run per rename; running each rename alone against master would fail, since later renames are in the tree too. The first criterion (a bare `same-behavior` passes) is unticked because of the differences below. Everything else in the harness CSV and every golden match, except: the harness CSV outcome column (`lost` is now `wiped`, line 43: the one intended value change), `bad_actions.json` (the `pack` rename), and three TUI goldens (`menu_parts`, `menu_arms`, `actor_windup`) where the longer name "The moving clay" shifts the layout and truncates one line (`lands at tick 86` shows as `tick 8`). That is the width, not a tick change. Game hashes are skipped under renames, as the tool documents.
- `go tool task check` cannot pass fully on this Windows machine: `internal/tools/samebehavior`'s own tests fail on a CRLF mismatch in the harness CSV (`seed,step,map`), which fails the same way on master. C1-24 owns that fix.
- Small fix outside the renames: `clock.Install`'s view rendered phase names with the registry it was installed with, so after `SwapContent` it used the old one. Once a phase name had to be `text.<phase id>` the swap test exposed it. It now uses `w.Content` (the installed registry only if there is none).
- The spec asks to update `.claude/rules/content.md`; a task branch cannot edit `.claude/`, so filed **C1-17.F1**.
- The `Graft` finding's scope: `act_alone` still requires a real graft (a piece), so a reattached limb never acts alone.

- Follow-ups filed: **C1-17.F1** (`.claude/rules/content.md`, protected on a task branch) and **C1-17.F2** (the TUI cuts a tick digit when the target name is long: the longer "The moving clay" exposes it).
