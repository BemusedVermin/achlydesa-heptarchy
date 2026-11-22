# Task index

The tracker. Each task file's front matter is the source of truth; update this table in the same change. Workflow: [README.md](README.md). Agent rules: [../AGENTS.md](../AGENTS.md).

Status values: `proposed`, `todo`, `in-progress`, `blocked`, `done`. Rows are in dependency order.

## M1 · Skeleton

Brief: [M1-skeleton/README.md](M1-skeleton/README.md)

| ID | Title | Status | Kind | Size | Depends on |
| --- | --- | --- | --- | --- | --- |
| [M1-01](M1-skeleton/M1-01-tooling.md) | Tooling, lint guards and package stubs | done | seam | S | none |
| [M1-02](M1-skeleton/M1-02-rng.md) | rng: named deterministic streams | done | feature | S | M1-01 |
| [M1-03](M1-skeleton/M1-03-api.md) | api: the plain-data boundary and its codec | done | seam | S | M1-01 |
| [M1-04](M1-skeleton/M1-04-sim.md) | sim: world, event bus, scheduler, command registry | done | seam | M | M1-02, M1-03 |
| [M1-05](M1-skeleton/M1-05-comp-hexmap.md) | comp basics and hexmap: 3D cells, chunks, topology | done | seam | M | M1-03 |
| [M1-06](M1-skeleton/M1-06-content.md) | content: CUE loader, IDs, text, formulas | done | seam | M | M1-03 |
| [M1-07](M1-skeleton/M1-07-templates-spawn.md) | Templates as component bags, and spawning | done | seam | M | M1-04, M1-05, M1-06 |
| [M1-08](M1-skeleton/M1-08-clock.md) | clock: ticks, phases and days | done | feature | S | M1-04, M1-06 |
| [M1-15](M1-skeleton/M1-15-turns.md) | turns: one shared timeline for every body | todo | seam | M | M1-04, M1-07, M1-08 |
| [M1-09](M1-skeleton/M1-09-game.md) | game: the public API | todo | seam | M | M1-15 |
| [M1-14](M1-skeleton/M1-14-sight-pathing.md) | hexmap: sight lines and pathing in 3D | todo | feature | M | M1-05 |
| [M1-10](M1-skeleton/M1-10-explore.md) | explore: maps from content, movement and the seamless plan view | todo | feature | L | M1-09, M1-14 |
| [M1-16](M1-skeleton/M1-16-lead-follow.md) | Lead and follow | todo | feature | M | M1-10 |
| [M1-11](M1-skeleton/M1-11-tui.md) | Terminal frontend on the new API, with a CLI | todo | feature | M | M1-16 |
| [M1-12](M1-skeleton/M1-12-save.md) | save: versioned snapshots and migrations | todo | seam | M | M1-16 |
| [M1-13](M1-skeleton/M1-13-harness.md) | harness: seeded runs and replays | todo | feature | S | M1-12 |

## M2 · Combat

Brief: [M2-combat/README.md](M2-combat/README.md)

| ID | Title | Status | Kind | Size | Depends on |
| --- | --- | --- | --- | --- | --- |
| [M2-01](M2-combat/M2-01-knowledge.md) | knowledge: the fact store | todo | seam | S | M1-11, M1-13 |
| [M2-02](M2-combat/M2-02-bodies-parts.md) | Bodies and parts as entities, with turn budgets and speed | todo | feature | M | M1-11, M1-13 |
| [M2-03](M2-combat/M2-03-materials-types-items.md) | Materials, types, items and the party's pack | todo | feature | S | M2-02 |
| [M2-04](M2-combat/M2-04-effects.md) | effects: context, registry and the primitives | todo | seam | L | M2-01, M2-03 |
| [M2-05](M2-combat/M2-05-triggers.md) | Triggers: content reacting to events | todo | seam | S | M2-04 |
| [M2-06](M2-combat/M2-06-forms-moves.md) | Forms, reach and move resolution | todo | feature | M | M2-04 |
| [M2-07](M2-combat/M2-07-hostility-alert-death.md) | Hostility, alert and death | todo | feature | M | M2-06 |
| [M2-08](M2-combat/M2-08-act-options-controller.md) | The act command, options and the creature controller | todo | feature | M | M2-07 |
| [M2-09](M2-combat/M2-09-grafts-crawlers.md) | Graft autonomy and crawling parts | todo | feature | S | M2-05, M2-08 |
| [M2-10](M2-combat/M2-10-action-menu.md) | Action menu and look mode on the map screen | todo | feature | M | M2-08 |
| [M2-11](M2-combat/M2-11-sample-content-tests.md) | Sample content, golden encounters and properties | todo | test | M | M2-09 |
| [M2-12](M2-combat/M2-12-hot-reload.md) | Content hot reload in dev builds | todo | feature | S | M1-11, M1-13 |

## Later milestones (outlines)

Each becomes task files when the milestone before it is done.

| Milestone | Brief |
| --- | --- |
| M3 · Agents | [M3-agents/README.md](M3-agents/README.md) |
| M4 · Seats | [M4-seats/README.md](M4-seats/README.md) |
| M5 · Procgen | [M5-procgen/README.md](M5-procgen/README.md) |
| M6 · Emergence | [M6-emergence/README.md](M6-emergence/README.md) |
| M7 · Tuning | [M7-tuning/README.md](M7-tuning/README.md) |

## Follow-ups

Filed by agents while working (see the Follow-ups section of [README.md](README.md)). The owner triages each into a milestone, merges it into another task, or closes it.

| ID | Title | Status | Kind | Milestone | Discovered by |
| --- | --- | --- | --- | --- | --- |
