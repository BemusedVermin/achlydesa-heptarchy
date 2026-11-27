---
id: M6-02
title: Sources and containers
status: done
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

- [x] The tests above pass.
- [x] `go tool task check` passes, and the save version is bumped.

## Out of scope

- Agents hauling: M6-13 and M6-23.
- Towns that contain them: M6-22.

## Notes

- `comp.Source.Refill` is the template ID whose `source.refill` formula is read each phase, so the component holds an ID, not a formula, and a hot reload picks up a new rate. The formula sees `stock` and `max` (`content.RefillEnv`) and returns an int; a negative result is a content error. `Registry.Refill(template, env)` runs it.
- A source's stock lives in `comp.Source.Stock`, apart from any `comp.Inventory` on the same entity. `inventory.Move` gives from and puts into a source's stock when the source is of that item (putting is refused over `Max`), and uses the inventory otherwise.
- Where the spec was silent: a source spawns full; "within reach" is the same cell or a neighbouring one (`hexmap.CellDistance` ≤ 1, which counts depth), fixed in code as the meaning of "next to" rather than a tuning number; `Move` returns a bool like `Take`, and moves only unbound stacks.
- `Move` needs a `Pos` on both entities. A party keeps its inventory on the party entity, which has none, so M6-13 must decide how a member takes into the party's inventory.
- Sample content: `item.water`, `item.ration` (no tags, so the type choice doesn't offer them), `template.source.well`, `template.container.cistern` and `template.container.larder`. No place spawns them yet. The containers start with five items because a template can't spawn an empty container: M6-02.F1.
- Save version 11, after #206 takes 10 (`save_test.go` gets the matching no-op migration step). `inventory.Install` now also registers the `source` spawner and the per-phase `inventory.refill` system.
- Files outside the deliverables: `internal/core/save/`, `internal/content/load_test.go` (a bad-source load case), `docs/content-guide.md` (templates row) and `docs/go-technical-design.md` (the `inventory` package row).
- Follow-ups: M6-02.F1.

