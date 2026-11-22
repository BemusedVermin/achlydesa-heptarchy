# Task index

The tracker. Each task file's front matter is the source of truth; update this table in the same change. Workflow: [README.md](README.md). Agent rules: [../AGENTS.md](../AGENTS.md).

Status values: `proposed`, `todo`, `in-progress`, `blocked`, `done`, `obsolete`. Rows are in dependency order.

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
| [M1-15](M1-skeleton/M1-15-turns.md) | turns: one shared timeline for every body | done | seam | M | M1-04, M1-07, M1-08 |
| [M1-17](M1-skeleton/M1-17-action-timing.md) | timeline: part-level action timing | done | seam | L | M1-15 |
| [M1-09](M1-skeleton/M1-09-game.md) | game: the public API | done | seam | M | M1-17 |
| [M1-14](M1-skeleton/M1-14-sight-pathing.md) | hexmap: sight lines and pathing in 3D | done | feature | M | M1-05 |
| [M1-18](M1-skeleton/M1-18-pending-actions.md) | timeline: actions that land later | done | seam | M | M1-17 |
| [M1-19](M1-skeleton/M1-19-effects-core.md) | effects: the effect core, selectors and timing primitives | done | seam | M | M1-06, M1-18 |
| [M1-20](M1-skeleton/M1-20-actions.md) | actions: content-defined actions | done | seam | L | M1-07, M1-14, M1-19 |
| [M1-10](M1-skeleton/M1-10-explore.md) | explore: maps from content, movement and the seamless plan view | done | feature | L | M1-09, M1-14, M1-20 |
| [M1-16](M1-skeleton/M1-16-lead-follow.md) | Lead and follow, and switching control | done | feature | M | M1-10 |
| [M1-11](M1-skeleton/M1-11-tui.md) | Terminal frontend on the new API, with a CLI | done | feature | M | M1-16 |
| [M1-12](M1-skeleton/M1-12-save.md) | save: versioned snapshots and migrations | done | seam | M | M1-16 |
| [M1-13](M1-skeleton/M1-13-harness.md) | harness: seeded runs and replays | done | feature | S | M1-12 |

## M2 · Combat

Brief: [M2-combat/README.md](M2-combat/README.md)

| ID | Title | Status | Kind | Size | Depends on |
| --- | --- | --- | --- | --- | --- |
| [M2-01](M2-combat/M2-01-knowledge.md) | knowledge: the fact store | todo | seam | S | M1-11, M1-13 |
| [M2-02](M2-combat/M2-02-bodies-parts.md) | Bodies and parts as entities, as channels | todo | feature | M | M1-11, M1-13 |
| [M2-03](M2-combat/M2-03-materials-types-items.md) | Materials, types, items and the party's pack | todo | feature | S | M2-02 |
| [M2-04](M2-combat/M2-04-effects.md) | Combat effect primitives | todo | feature | M | M2-01, M2-03 |
| [M2-05](M2-combat/M2-05-triggers.md) | Triggers: content reacting to events | todo | seam | S | M2-04 |
| [M2-06](M2-combat/M2-06-forms-moves.md) | The seven forms as actions, with types and timing | todo | content | M | M2-04 |
| [M2-07](M2-combat/M2-07-hostility-alert-death.md) | Hostility, alert and death | todo | feature | M | M2-06 |
| [M2-08](M2-combat/M2-08-act-options-controller.md) | The bodies view and the creature controller | todo | feature | M | M2-07 |
| [M2-09](M2-combat/M2-09-grafts-crawlers.md) | Graft autonomy and crawling parts | todo | feature | S | M2-05, M2-08 |
| [M2-10](M2-combat/M2-10-action-menu.md) | Action menu and look mode on the map screen | todo | feature | M | M2-08 |
| [M2-11](M2-combat/M2-11-sample-content-tests.md) | Sample content, lethality targets and golden encounters | todo | test | M | M2-09 |
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
| [M1-15.F1](followups/M1-15.F1-timeline-after-load.md) | Rebuild the turn timeline after loading a save | obsolete | seam | M1 | M1-15 |
| [M1-14.F1](followups/M1-14.F1-sight-across-links.md) | Sight and pathing across map links | proposed | seam | M1 | M1-14 |
| [M1-17.F1](followups/M1-17.F1-reschedule-on-hook-change.md) | timeline: reschedule bodies when a hook's answer changes | proposed | seam | M1 | M1-17 |
| [M1-18.F1](followups/M1-18.F1-pending-save-roundtrip.md) | Round-trip pending actions through a real save | proposed | test | M1 | M1-18 |
| [M1-19.F1](followups/M1-19.F1-effect-validation-end-to-end.md) | Test that game.New rejects bad effects in real content | proposed | test | M1 | M1-19 |
| [M1-20.F1](followups/M1-20.F1-check-draws-roll.md) | Effect checks must not draw from the RNG | proposed | bug | M1 | M1-20 |
| [M1-20.F2](followups/M1-20.F2-cell-targets.md) | Cell-targeted actions | proposed | seam | unplanned | M1-20 |
| [M1-10.F1](followups/M1-10.F1-slope-floor-in-plan-view.md) | Plan view: a slope's ceiling beats its floor | proposed | question | M1 | M1-10 |
| [M1-10.F2](followups/M1-10.F2-fall-injuries.md) | Injuries from falls | proposed | feature | M2 | M1-10 |
| [M1-10.F3](followups/M1-10.F3-sight-vs-prototype.md) | Sight differs from the prototype at pillar edges | proposed | debt | M1 | M1-10 |
| [M1-16.F1](followups/M1-16.F1-follow-across-links.md) | Followers and swaps across map links | proposed | seam | M1 | M1-16 |
| [M1-16.F2](followups/M1-16.F2-party-view-control.md) | Show the party's control mode and alert in the view | proposed | feature | M1 | M1-16 |
| [M1-12.F1](followups/M1-12.F1-load-with-spawning-extra.md) | Let Load run a Config.Extra that spawns | proposed | debt | M1 | M1-12 |
| [M1-11.F1](followups/M1-11.F1-wasd-alternating-diagonals.md) | WASD: up and down zigzag diagonally | proposed | feature | M1 | M1-11 |
