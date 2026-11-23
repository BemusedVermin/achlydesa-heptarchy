---
id: C1-08
title: "Is it a pack or an inventory?"
status: todo
milestone: C1
kind: question
depends_on: []
discovered_by: null
size: S
---

# C1-08 · Is it a pack or an inventory?

## Goal

The owner picks one option, and the answer is recorded here and where the glossary or design doc needs it.

## Context

- Vocabulary audit finding 11.

## Deliverables

- The answer, written in this task's Notes and in `docs/glossary.md` (and `docs/` where the design says otherwise).

## Spec

"Pack" is used in formulas, effect parameters (`from: "pack"`, `api.AttachedFromPack`), text and 72 places in Go. "Inventory" is used for the template field (`inventory:`), `comp.Inventory` and the package name.

- **pack**: Rename the template field to `pack:` and `comp.Inventory` to `comp.Pack`. The package stays `inventory`.
- **inventory**: Rename the formulas, parameters and text to `inventory`. Much larger.

**Recommendation:** `pack`.

**Decided by the owner:** `inventory`. The formulas, effect parameters (`from: "pack"`, `api.AttachedFromPack`), text and Go that say "pack" are renamed to `inventory`; the template field, `comp.Inventory` and the package already are. This is the larger rename: about 72 places in Go plus content. Applied by C1-17.

## Acceptance criteria

- [x] The owner has answered, through `/decide`.
- [x] The answer is in Notes and in the glossary.

## Out of scope

- Doing the work: the cleanup task named in the milestone brief does it.

## Notes
Decided 2026-11-23: `inventory` (not the recommended `pack`). In the glossary and the tech design.
