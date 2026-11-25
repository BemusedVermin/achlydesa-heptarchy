---
id: M4-10
title: Knowing where things are
status: todo
milestone: M4
kind: feature
depends_on: [M4-08]
discovered_by: null
size: M
---

# M4-10 · Knowing where things are

## Goal

Every agent knows where things that meet its needs are because it saw them, or because content says it starts out knowing them. What it knows can be wrong: a cistern it remembers full may be empty now. Hearsay joins this in M7.

## Context

- `docs/game-design.md` → "The living deep" (Outcomes emerge), "People with needs" (Knowing where).
- The brief's "The one-world model" → "Needs across the world".
- What exists:
  - `internal/core/knowledge` (`Learn`, `About`, `Best`) and `api.Fact`, whose subject is a UID;
  - `chooser.resolve`, `nearestBody` and `nearestCell` in `internal/core/agents/behaviors.go`, which today find the nearest match on the whole map whether or not the agent has seen it;
  - sight (`hexmap.Map.Visible`, `SightRadius`, `SightDepth`);
  - M4-08's regions.

## Deliverables

- `internal/core/agents` (a file such as `knowing.go`), with three parts:
  - **Learning by sight.** As an agent sees bodies, cells and regions that a target filter or terrain in its behaviors could match, it learns where they are and when it saw them. Learning runs when the agent decides, so it costs nothing between decisions.
  - **Correction.** Seeing a known place again replaces what the agent knew of it: a container seen empty, or a body seen gone, stops counting.
  - **Starting knowledge.** A template or place in content lists what its agents know at the start, such as a resident knowing its own crawlspace's containers and beds.
- `internal/core/api` and `internal/core/knowledge`: a fact that can name a cell or a region as well as a UID, either by extending `api.Fact` or with a sibling record. Say which in Notes.
- `internal/content`: the starting-knowledge field; regenerate the schemas.
- Tests (below). `docs/content-guide.md`: starting knowledge. `docs/glossary.md`: **knowing** (what an agent knows of where things are), if the term is new.
- `internal/core/save`: bump `CurrentVersion`.

## Spec

- **Any need.** Nothing in Go names water, food, rest, pay or any other need. What gets learned is decided by the target filters and terrains in the agent's own behaviors. Firm.
- **Seen, or given, never assumed.** Nothing in the world is known to an agent that hasn't seen it or been given it by content. Hearsay is M7. Firm (owner, 2026-11-25).
- **Bounded memory.** An agent keeps at most `agents.knowing.limit` facts per kind, forgetting the oldest. The limit is content and a placeholder for M8.
- Deterministic: learning walks seen cells in a fixed order, and ties go by UID or cell order.

## Acceptance criteria

- [ ] A test: an agent that walks past a full cistern knows of it afterwards; one that never saw it doesn't.
- [ ] A test: an agent that sees a cistern it knew of, now empty, no longer counts it as water.
- [ ] A test: a resident given its crawlspace's containers as starting knowledge knows them on the first tick.
- [ ] `go tool task check` passes.

## Out of scope

- Using what is known to choose where to go, and random paths when nothing is known: M4-15.
- Hearsay and rumors: M7.

## Notes

Filled in by the implementer.
