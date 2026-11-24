---
id: C1-26
title: "Design-doc fixes and the content guide"
status: done
milestone: C1
kind: content
depends_on: [C1-17, C1-19]
discovered_by: null
size: L
---

# C1-26 · Design-doc fixes and the content guide

## Goal

The design docs agree with the code, and a designer has a content guide.

## Context

- Milestone brief: [README.md](README.md). Findings come from the audit that opened C1.
- `docs/glossary.md` for settled terms.
- This is one pull request by design: the work touches the same packages and is done in one pass.

## Deliverables

- The changes described in the Spec, in one branch and one pull request.

## Spec

### Fix the contradictions in the design docs

The design docs agree with the code and with each other.

Findings:

- `docs/game-design.md:12` says "no grid and no range bands" but the doc has a hex-grid section (`:105`) and forms have `range`.
- `game-design.md:546` lists primitives `sever_check`, `lock_part`, `freeze_others`, `transfer_action`, `drain_stat`, `freeze_trait`; the real ones are `damage`, `sever`, `lock`, `delay`, `ready`, `interrupt`, `add_status`, `attach_part`, `reveal`, `consume`, `raise_max_integrity`, `act_alone`, `spawn_from_part`, `step`. Mark the list as examples or use the real names.
- `docs/go-technical-design.md:27,153` write `Pos{Map, Cell}`; the field is `At` (`comp.go:12`). (Lines 121, 177, 188 belong to the form rename.)

Documents only: update `docs/`, change no code.

### Write the content guide

`docs/content-guide.md` tells a designer how to add and change content, with names that have settled.

Findings:

- `docs/content-guide.md` doesn't exist.

Cover: the layout of `content/data` and `maps/`; ID prefixes, with the table that the collections task made; text IDs and where they live (as decided); `rules.cue` and `rule.<name>` inlining; choices, mods and `vocabulary.cue`; triggers and the trigger roles; recipes that point at real files (a creature in `templates.cue`, a type in `types.cue` with `#Consumable`, a part in `bodies.cue`, a map in `maps/test_den.cue`); checking the work with `go tool task content`, the loader tests and `play`. Terms match `docs/glossary.md`.

## Acceptance criteria

- [x] Every file, ID and command the docs name exists
- [x] The guide's terms match `docs/glossary.md`
- [x] `go tool task content` passes
- [x] Only files under `docs/` changed
- [x] Only files under `docs/` changed.
- [x] Every identifier the edited lines name exists.
- [x] Every file, ID and command the guide names exists.
- [x] Its terms match the glossary.
- [x] `go tool task content` passes.

## Out of scope

- Other areas' findings: they have their own task in this milestone.
- Any new behavior beyond what the Spec names.

## Notes


- Triaged 2026-11-23: todo in C1, as the owner decided; it is part of the first cleanup milestone.

Docs changed:

- `docs/game-design.md`: the Genre line now says actions are picked from menus, on the hex grid, with range in hexes rather than range bands. The primitives list names the 14 real primitives. Also, outside the listed findings: the "validated on load" sentence now says effects are checked when a game starts, which is where `effects.ValidateContent` runs; the task reviewer found it.
- `docs/go-technical-design.md`: `Pos{Map, Cell}` became `Pos{Map, At}` at the two places named; the form-rename lines were left alone.
- `docs/content-guide.md`: new. The ID table follows the `collections()` table in `internal/content/records.go`, in its order, plus calendar phase IDs.

Checking: the guide is careful about when a mistake is caught. Record references and most formulas fail on load (`internal/content` tests); effect params (formulas, IDs, selectors), primitives and trigger event kinds fail only when a game starts (`internal/core/game` tests, or `play`). Every backticked ID in the guide was checked against `content/data` with a script.

The "Only files under `docs/` changed" criteria hold for the docs work; `tasks/` changed only for this task's bookkeeping and the follow-up.

Follow-ups: C1-26.F1 (three more uses of "move" for an action in `game-design.md`, outside this Spec's lines).
