---
id: M6-02
title: Sources and containers
status: in-progress
milestone: M6
kind: seam
depends_on: []
discovered_by: null
size: M
---

# M6-02 · Sources and containers

## Goal

Water and food come from somewhere. A well, lake or field is a source that refills slowly; a cistern, larder or pile on the floor is a container. Anything can be taken from or put into one by a body next to it.

## Context

- `docs/game-design.md` → "The living deep" ("Resources are things in places"), "How minds work" → "Supplies are things in places".
- `docs/go-technical-design.md` → "Designing for pivots" (Identity, Entity templates).
- What exists: `comp.Inventory` and the `inventory` package (Add, Take, Count, Usable) work on any entity; nothing but bodies carries one.

## Deliverables

- `internal/core/comp/`: `Source{Item, Stock, Max, Refill}` beside `Inventory`; regenerate the schema.
- `internal/core/inventory/`: `Move(w, from, to, item, count)` between two entities that stand within reach; a per-phase system that refills sources, registered with the scheduler.
- `internal/content/`: a template can carry `source:` (item, max, refill formula) and an inventory; `inventory:` already exists for bodies.
- `content/data/`: a sample well, cistern and larder template, and `item.water`, `item.ration` if absent (see `items.cue`).
- Tests: a source refills by its rate per phase and never above `Max`; `Move` refuses outside reach and over what is held; a save round-trips a half-empty source.

## Spec

- **Firm:** a container is any entity with `Pos` and `Inventory`. No `comp.Container`. A source adds refilling and is also taken from like a container.
- **Firm:** the refill rate is an Expr formula in content, read per phase. It is physics, not an outcome (AGENTS.md rule 10): nothing refills a cistern, only a source does.
- **Firm:** take and put are actions in `actions` (M6-13); this task provides the data and the inventory move they call.
- A source may be a terrain cell kind later (lake water); this task only needs entities.

## Acceptance criteria

- [ ] The tests above pass.
- [ ] `go tool task check` passes, and the save version is bumped.

## Out of scope

- Agents hauling: M6-13 and M6-23.
- Towns that contain them: M6-22.

## Notes

