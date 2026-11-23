---
id: C1-26
title: "Design-doc fixes and the content guide"
status: proposed
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

- [ ] Every file, ID and command the docs name exists
- [ ] The guide's terms match `docs/glossary.md`
- [ ] `go tool task content` passes
- [ ] Only files under `docs/` changed
- [ ] Only files under `docs/` changed.
- [ ] Every identifier the edited lines name exists.
- [ ] Every file, ID and command the guide names exists.
- [ ] Its terms match the glossary.
- [ ] `go tool task content` passes.

## Out of scope

- Other areas' findings: they have their own task in this milestone.
- Any new behavior beyond what the Spec names.

## Notes

