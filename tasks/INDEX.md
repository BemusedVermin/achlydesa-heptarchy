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
| [M2-01](M2-combat/M2-01-knowledge.md) | knowledge: the fact store | done | seam | S | M1-11, M1-13 |
| [M2-02](M2-combat/M2-02-bodies-parts.md) | Bodies and parts as entities, as channels | done | feature | M | M1-11, M1-13 |
| [M2-03](M2-combat/M2-03-materials-types-items.md) | Materials, types, items and the party's pack | done | feature | S | M2-02 |
| [M2-04](M2-combat/M2-04-effects.md) | Combat effect primitives | done | feature | M | M2-01, M2-03 |
| [M2-05](M2-combat/M2-05-triggers.md) | Triggers: content reacting to events | done | seam | S | M2-04 |
| [M2-12](M2-combat/M2-12-hot-reload.md) | Content hot reload in dev builds | done | feature | S | M1-11, M1-13 |
| [M2-13](M2-combat/M2-13-rules-as-content.md) | Rules as content: records, rules and choices in CUE | done | seam | L | M2-05, M2-12 |
| [M2-06](M2-combat/M2-06-forms-moves.md) | The seven forms as actions, with types and timing | done | content | M | M2-13 |
| [M2-07](M2-combat/M2-07-hostility-alert-death.md) | Hostility, alert and death | done | feature | M | M2-06 |
| [M2-08](M2-combat/M2-08-act-options-controller.md) | The bodies view and the creature controller | done | feature | M | M2-07 |
| [M2-09](M2-combat/M2-09-grafts-crawlers.md) | Graft autonomy and crawling parts | done | feature | S | M2-05, M2-08 |
| [M2-10](M2-combat/M2-10-action-menu.md) | Action menu and look mode on the map screen | done | feature | M | M2-08 |
| [M2-11](M2-combat/M2-11-sample-content-tests.md) | Sample content, lethality targets and golden encounters | done | test | M | M2-09 |
| [M2-14](M2-combat/M2-14-step-speed.md) | Make a game step fast again after the seven forms | done | bug | M | M2-06 |
| [M2-15](M2-combat/M2-15-triage-followups.md) | Triage the open follow-ups into milestones | done | debt | S | none |

## C1 · Cleanup

Brief: [C1-cleanup/README.md](C1-cleanup/README.md)

| ID | Title | Status | Kind | Size | Depends on |
| --- | --- | --- | --- | --- | --- |
| [C1-01](C1-cleanup/C1-01-clay.md) | What do we call the clay creature that moves, and the body Stand freezes? | done | question | S | none |
| [C1-02](C1-cleanup/C1-02-graft.md) | Is an own severed part that is put back a graft? | done | question | S | none |
| [C1-03](C1-cleanup/C1-03-down.md) | Which words for being out of the fight, for dropping, and for a wipe? | done | question | S | none |
| [C1-04](C1-cleanup/C1-04-actor.md) | How are the acting body, the body acted on, a fact's origin and a formula's subject named? | done | question | M | none |
| [C1-05](C1-cleanup/C1-05-formula.md) | How are the variables in formula environments named? | done | question | S | none |
| [C1-06](C1-cleanup/C1-06-form.md) | What are "form", "action", "move" and "type"? | done | question | S | none |
| [C1-07](C1-cleanup/C1-07-still.md) | What is the controller for a creature that keeps still called? | done | question | S | none |
| [C1-08](C1-cleanup/C1-08-pack.md) | Is it a pack or an inventory? | done | question | S | none |
| [C1-09](C1-cleanup/C1-09-turn.md) | What do "turn", "per-turn" and "pending" mean? | done | question | S | none |
| [C1-10](C1-cleanup/C1-10-faction.md) | What is the hostile faction called? | done | question | S | none |
| [C1-11](C1-cleanup/C1-11-text.md) | Where does player-facing text live, and how are its IDs shaped? | done | question | S | none |
| [C1-12](C1-cleanup/C1-12-frontend.md) | Where does frontend chrome end and game text begin, and which core packages may the tui import? | done | question | S | none |
| [C1-13](C1-cleanup/C1-13-cache.md) | May tests share one cached sample registry? | done | question | S | none |
| [C1-14](C1-cleanup/C1-14-yaml.md) | May `go.yaml.in/yaml/v3` become a direct dependency? | done | question | S | none |
| [C1-15](C1-cleanup/C1-15-gitcmd.md) | Where does shared git code for the tools go, given agentkit is protected? | done | question | S | none |
| [C1-16](C1-cleanup/C1-16-speed.md) | What should a speed formula that fails at run time do? | done | question | S | none |
| [C1-17](C1-cleanup/C1-17-apply-the-decided-renames.md) | Apply the decided renames | done | debt | L | C1-01, C1-02, C1-03, C1-04, C1-05, C1-06, C1-07, C1-08, C1-09, C1-10, C1-11 |
| [C1-18](C1-cleanup/C1-18-shared-helpers-in-the-lowest-layer-and-one-test.md) | Shared helpers in the lowest layer, and one test fixture | done | debt | L | C1-17, C1-13 |
| [C1-19](C1-cleanup/C1-19-content-area-duplication-tests-comments.md) | Content area: duplication, tests, comments | done | debt | L | C1-18 |
| [C1-20](C1-cleanup/C1-20-actions-area-duplication-tests-comments.md) | Actions area: duplication, tests, comments | done | debt | L | C1-18 |
| [C1-21](C1-cleanup/C1-21-features-area-duplication-tests-comments.md) | Features area: duplication, tests, comments | done | debt | L | C1-18 |
| [C1-22](C1-cleanup/C1-22-foundation-area-duplication-tests-comments-speed.md) | Foundation area: duplication, tests, comments, speed errors | done | debt | L | C1-18, C1-16 |
| [C1-23](C1-cleanup/C1-23-game-and-frontend-duplication-text-out-of-the-tu.md) | Game and frontend: duplication, text out of the tui, tests, comments | done | debt | L | C1-18, C1-12 |
| [C1-24](C1-cleanup/C1-24-tools-fix-the-windows-test-shared-git-code-tests.md) | Tools: fix the Windows test, shared git code, tests, comments | done | debt | L | C1-18, C1-15 |
| [C1-25](C1-cleanup/C1-25-agentkit-cleanup-owner-does-this-one.md) | agentkit cleanup (owner does this one) | obsolete | debt | L | C1-24, C1-14 |
| [C1-26](C1-cleanup/C1-26-design-doc-fixes-and-the-content-guide.md) | Design-doc fixes and the content guide | done | content | L | C1-17, C1-19 |
| [C1-27](C1-cleanup/C1-27-lock-in-what-the-cleanup-decided.md) | Lock in what the cleanup decided | todo | debt | S | C1-17, C1-18, C1-19, C1-20, C1-21, C1-22, C1-23, C1-24, C1-26, C1-28, C1-29, C1-30 |
| [C1-28](C1-cleanup/C1-28-foundation-comments-to-the-tightened-rule.md) | Foundation comments to the tightened rule | done | debt | M | none |
| [C1-29](C1-cleanup/C1-29-content-comments-to-the-tightened-rule.md) | Content comments to the tightened rule | todo | debt | M | none |
| [C1-30](C1-cleanup/C1-30-actions-comments-to-the-tightened-rule.md) | Actions comments to the tightened rule | todo | debt | M | none |

## Later milestones (outlines)

Each becomes task files when the milestone before it is done: run `/plan-milestone`.

| Milestone | Brief |
| --- | --- |
| M3 · Agents | [M3-agents/README.md](M3-agents/README.md) |
| M4 · Seats | [M4-seats/README.md](M4-seats/README.md) |
| M5 · Procgen | [M5-procgen/README.md](M5-procgen/README.md) |
| M6 · Emergence | [M6-emergence/README.md](M6-emergence/README.md) |
| M7 · Tuning | [M7-tuning/README.md](M7-tuning/README.md) |

## Post-MVP milestones

Milestones for after the MVP (M1 to M7). None starts before the MVP is done, and nothing in the MVP depends on them. See [post-mvp/README.md](post-mvp/README.md).

| Milestone | Brief | Design |
| --- | --- | --- |
| PM-STORY · Story | [post-mvp/story/README.md](post-mvp/story/README.md) | `docs/story.md` |

## Follow-ups

Filed by agents while working (see the Follow-ups section of [README.md](README.md)). The owner triages each into a milestone, merges it into another task, or closes it, either directly or through a triage task such as M2-15.

| ID | Title | Status | Kind | Milestone | Discovered by |
| --- | --- | --- | --- | --- | --- |
| [C1-23.F2](followups/C1-23.F2-save-property-test-speed.md) | Bring the save property test under 2 seconds | done | test | C1 | C1-23 |
| [C1-23.F1](followups/C1-23.F1-reload-messages-in-content.md) | Move the hot reload's log messages into content text | obsolete | debt | C1 | C1-23 |
| [C1-26.F1](followups/C1-26.F1-move-means-movement-in-game-design.md) | Use "move" only for movement in game-design.md | todo | debt | C1 | C1-26 |
| [C1-24.F2](followups/C1-24.F2-samebehavior-missing-seed-message.md) | samebehavior says clearly when a seed reports no hash | done | bug | C1 | C1-24 |
| [C1-24.F1](followups/C1-24.F1-agentkit-tests-use-gittest.md) | agentkit's tests use gittest and build fewer repositories | obsolete | test | C1 | C1-24 |
| [C1-21.F2](followups/C1-21.F2-check-pr-allowed-behavior-changes.md) | Let check-pr accept a cleanup task's named behavior changes | todo | seam | C1 | C1-21 |
| [C1-21.F1](followups/C1-21.F1-combat-tests-load-content-each-time.md) | Stop reparsing the sample content in every combat test | todo | test | C1 | C1-21 |
| [C1-17.F2](followups/C1-17.F2-tui-line-cuts-the-tick.md) | The actor panel cuts a line's last digit when the target's name is long | obsolete | bug | C1 | C1-17 |
| [C1-17.F1](followups/C1-17.F1-content-rules-text-ids.md) | Correct .claude/rules/content.md for the text ID shape | done | debt | C1 | C1-17 |
| [M2-11.F2](followups/M2-11.F2-where-the-three-creatures-spawn.md) | Where do the sample's three creatures spawn? | obsolete | question | M2 | M2-11 |
| [M2-11.F1](followups/M2-11.F1-ready-frees-a-winding-channel.md) | Stand and Give free a channel that still has a swing winding up | done | bug | M2 | M2-11 |
| [M1-17.F2](followups/M1-17.F2-advance-to-next-real-action.md) | Advance the timeline to the player's next real action | done | bug | M2 | M1-17 |
| [M2-09.F1.F1](followups/M2-09.F1.F1-what-a-crawling-part-can-do.md) | What can a crawling part do? | done | question | M2 | M2-09.F1 |
| [M2-10.F2](followups/M2-10.F2-agentkit-afteredit-windows-path.md) | Fix TestAfterEditWithGolangciLint on Windows paths | done | bug | unplanned | M2-10 |
| [M2-10.F1](followups/M2-10.F1-fight-to-the-end-test.md) | Test a whole fight on the map screen, to the party resuming follow | done | test | M2 | M2-10 |
| [M2-09.F1](followups/M2-09.F1-crawling-parts-know-no-actions.md) | Give a crawling part the actions to fight with | done | seam | M2 | M2-09 |
| [M2-08.F2](followups/M2-08.F2-alert-started-twice.md) | AlertStarted is published twice for one alert | done | bug | M2 | M2-08 |
| [M2-08.F1](followups/M2-08.F1-sample-crawler-hunts.md) | Make the sample crawler hunt | done | content | M2 | M2-08 |
| [M2-07.F1.F1.F1](followups/M2-07.F1.F1.F1-rest-sets-rest-place.md) | Resting in a town sets the party's rest place | todo | feature | M3 | M2-07.F1.F1, M3 |
| [M2-07.F1.F1](followups/M2-07.F1.F1-wake-after-wipe.md) | Wake at the last rest town after a party wipe | done | feature | M3 | M2-07.F1 |
| [M2-06.F1.F1](followups/M2-06.F1.F1-forms-doc-strike-sever.md) | Update the forms table in the game design for Strike and Sever | done | debt | unplanned | M2-06.F1 |
| [M2-06.F1](followups/M2-06.F1-strike-sever-weakness.md) | Should Strike and Sever differ on a weakness the way these placeholders do? | done | question | unplanned | M2-06 |
| [M2-06.F4.F1](followups/M2-06.F4.F1-record-name-keepsake-decision.md) | Record in the design doc that Name and Keepsake are never refused | done | debt | M2 | M2-06.F4 |
| [M2-07.F3](followups/M2-07.F3-wipe-inside-apply.md) | Stop the timeline when the party is wiped inside an Apply | done | bug | M2 | M2-07 |
| [M2-07.F1](followups/M2-07.F1-party-wipe-story.md) | What does a party wipe mean for the story? | done | question | unplanned | M2-07 |
| [M2-07.F2](followups/M2-07.F2-idle-bodies-ping-pong.md) | Idle bodies waiting on each other move time on a tick at a time | done | bug | M2 | M2-07 |
| [M2-06.F2](followups/M2-06.F2-crawls-tag.md) | A crawls tag, so Stand freezes only the crawlers | done | feature | M2 | M2-06 |
| [M2-06.F3](followups/M2-06.F3-relic-innate-to-grafts.md) | Relic as innate to grafts | done | seam | M2 | M2-06 |
| [M2-06.F5](followups/M2-06.F5-option-labels-name-the-choice.md) | Option labels show a chosen record's ID, not its name | done | bug | M2 | M2-06 |
| [M2-06.F4](followups/M2-06.F4-disabled-option-is-information.md) | A disabled option is information: Name and Keepsake on the wrong body | done | question | unplanned | M2-06 |
| [M2-13.F1](followups/M2-13.F1-cue-schema-tests.md) | Test CUE constraints against the real schema | done | test | unplanned | M2-13 |
| [M2-01.F1](followups/M2-01.F1-render-fact-values.md) | Render a fact's value as text in the log | done | seam | unplanned | M2-01 |
| [M1-15.F1](followups/M1-15.F1-timeline-after-load.md) | Rebuild the turn timeline after loading a save | obsolete | seam | M1 | M1-15 |
| [M1-14.F1](followups/M1-14.F1-sight-across-links.md) | Sight and pathing across map links | obsolete | seam | unplanned | M1-14 |
| [M1-17.F1](followups/M1-17.F1-reschedule-on-hook-change.md) | timeline: reschedule bodies when a hook's answer changes | obsolete | seam | M1 | M1-17 |
| [M1-18.F1](followups/M1-18.F1-pending-save-roundtrip.md) | Round-trip pending actions through a real save | obsolete | test | M1 | M1-18 |
| [M1-19.F1](followups/M1-19.F1-effect-validation-end-to-end.md) | Test that game.New rejects bad effects in real content | obsolete | test | M1 | M1-19 |
| [M1-20.F1](followups/M1-20.F1-check-draws-roll.md) | Effect checks must not draw from the RNG | done | bug | M1 | M1-20 |
| [M1-20.F2](followups/M1-20.F2-cell-targets.md) | Cell-targeted actions | obsolete | seam | unplanned | M1-20 |
| [M1-10.F1](followups/M1-10.F1-slope-floor-in-plan-view.md) | Plan view: a slope's ceiling beats its floor | done | question | M1 | M1-10 |
| [M1-10.F2](followups/M1-10.F2-fall-injuries.md) | Injuries from falls | done | feature | M2 | M1-10 |
| [M1-10.F1.F1](followups/M1-10.F1.F1-plan-view-reachability.md) | Plan view: draw by reachability within a limit | done | feature | M1 | M1-10.F1 |
| [M2-02.F1](followups/M2-02.F1-despawn-body-parts.md) | Despawning a body must remove its parts | done | bug | M2 | M2-02 |
| [M2-04.F1](followups/M2-04.F1-validate-content-ids-in-effects.md) | Check content IDs named by effect parameters at load | done | seam | M2 | M2-04 |
| [M2-04.F2](followups/M2-04.F2-lock-reason-on-busy.md) | Tell a lock from ordinary busy time | done | seam | M2 | M2-04 |
| [M2-05.F1](followups/M2-05.F1-harden-primitive-and-event-sources.md) | A raise_max_integrity primitive and a source on combat events | done | seam | M2 | M2-05 |
| [M1-10.F1.F1.F1](followups/M1-10.F1.F1.F1-cache-plan-reachability.md) | Plan view: do not recompute reachability on every view build | obsolete | debt | unplanned | M1-10.F1.F1 |
| [M1-10.F3](followups/M1-10.F3-sight-vs-prototype.md) | Sight differs from the prototype at pillar edges | obsolete | debt | M1 | M1-10 |
| [M1-16.F1](followups/M1-16.F1-follow-across-links.md) | Followers and swaps across map links | obsolete | seam | unplanned | M1-16 |
| [M1-16.F2](followups/M1-16.F2-party-view-control.md) | Show the party's control mode and alert in the view | obsolete | feature | M1 | M1-16 |
| [M1-12.F1](followups/M1-12.F1-load-with-spawning-extra.md) | Let Load run a Config.Extra that spawns | done | debt | M1 | M1-12 |
| [M1-11.F1](followups/M1-11.F1-wasd-alternating-diagonals.md) | WASD: up and down zigzag diagonally | done | feature | M1 | M1-11 |
| [M2-12.F1](followups/M2-12.F1-swap-content-caches.md) | Hot reload does not refresh content cached at game start | done | seam | unplanned | M2-12 |
