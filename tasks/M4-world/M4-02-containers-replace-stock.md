---
id: M4-02
title: "Containers replace the town stock; delete generics and emptying"
status: todo
milestone: M4
kind: feature
depends_on: [M4-01]
discovered_by: null
size: M
---

# M4-02 · Containers replace the town stock; delete generics and emptying

## Goal

Water and food are items in containers, things with a place and an inventory, that residents fill and empty through their behaviors. The abstract town stock, the generic replacement timers and the town-emptied report are deleted.

## Context

- `docs/game-design.md` → "The living deep" (Outcomes emerge: resources are things in places, and containers), "People with needs" (Towns can empty), "Three tiers of people" (nobody respawns; new people come from births and arrivals).
- The brief's "Decisions" (containers) and "The one-world model" → "Things in places".
- What exists: `internal/core/agents/stock.go` (`take_stock`, `stockInReach`, `stockOfMap`), `generics.go` (`tickGenerics`, replacements), `emptied.go` (`publishIfEmptied`), `comp.Town`, `comp.Replacement`, `comp.Generic`, `content.TownDef` (`stock`, `stock_marks`, `residents`, `use`, `refill`, `exodus`, `generics`), `stepCarry` in `primitives.go`, and `template.town.well` in `templates.cue`.

## Deliverables

- `internal/core/agents/stock.go`, `generics.go`, `emptied.go` and their tests: deleted, with the `agents.generics` system, the `agents.emptied` subscription and the `take_stock` primitive.
- `internal/core/comp/comp.go`: delete `Town`, `Replacement` and `Generic`.
- `internal/core/agents/recruits.go`: a body is hirable when its template has `recruit`, it is alive and in the fight, and it is in no party. It no longer needs `comp.Generic`.
- `internal/core/api`: delete `TownEmptied` and `StockLine`; `RegionView` loses `Town` and `Stock`.
- `internal/content/types.go`: delete `TownDef` and `RegionDef.Town`; regenerate the schemas.
- `internal/core/agents/primitives.go`: `carry` puts what it takes into a **container** body (see Spec) rather than a stock mark. `consume` can target a container. A test for each.
- `content/data/templates.cue`: container templates, such as `template.container.cistern` and `template.container.larder`, with an `inventory`, an appearance, a `container` tag, and no actor and no needs.
- Content: each town's stock becomes containers placed in its layers with that stock in their inventories. Haulers carry water from the well to a cistern. Residents drink and eat from containers. `refill` is gone, and a town whose water comes from nowhere runs dry. Generics become ordinary placed bodies of their templates.
- `internal/frontend/tui`: the region panel stops drawing a stock; update `region_town.golden`.
- Text: delete `text.agents.no_stock` and other text only the deleted code used.
- `internal/core/save`: bump `CurrentVersion`.
- `docs/content-guide.md`: replace "Regions and towns" with a short section on containers. `docs/glossary.md`: add **container**.

## Spec

- **A container** is a body spawned from a template whose `inventory` it holds, and whose tags let a behavior's target filter find it (`kind: "body"`, `tags: ["container"]`). It has no `Actor`, so it never decides. That it needs no new component is firm: containers reuse `comp.Inventory`, `comp.Pos` and the body target filter. If target filters can't reach a body that has no `Body` plan, give containers the smallest body plan (one part) and say so in Notes rather than add a component.
- **carry** (suggestion for the params): `item`, `count`, and a `to` that is a target filter (tags) for the container to put into, replacing the stock mark. It takes from the target body's inventory as today, walks to the nearest matching container, and adds the items to that container's inventory when it stands next to it.
- **Nothing refills and nothing replaces.** A dead generic is gone. A town whose people leave is just a place with nobody in it.
- The well stays a body from `template.town.well`, whose inventory is what haulers take from. Whether the well's water itself runs out is content: give it a large count now, and leave a refilling source to a later need (file a follow-up if a behavior is missing).

## Acceptance criteria

- [ ] `git grep -w -E "TownEmptied|take_stock|stock_marks|StockLine|Replacement|tickGenerics" -- internal content docs` finds nothing.
- [ ] A test: a hauler carries water from the well to a cistern, and a thirsty resident drinks from the cistern, all as behaviors, in the sample town.
- [ ] A test: with the cistern and the well empty, residents' thirst rises and nothing adds water.
- [ ] A test: a dead generic is not replaced after any number of days.
- [ ] Changed goldens and harness output are regenerated and named in Notes.
- [ ] `go tool task check` passes.

## Out of scope

- People leaving a dry town for another: M4-10 (until then, `leave` works through gates as M3 left it).
- Placing containers in one world at human scale: M4-11.
- Recruits and hiring: M4-03.

## Notes

Filled in by the implementer.
