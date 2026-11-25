---
id: M4-01
title: "Everyone is an individual: delete level of detail and counts"
status: done
milestone: M4
kind: debt
depends_on: []
discovered_by: null
size: M
---

# M4-01 · Everyone is an individual: delete level of detail and counts

## Goal

Every creature and person is an entity from the first tick to its death. Off-screen populations, their drift and attrition, and promotion and demotion are deleted, not wrapped, and content places actual bodies where it used to place counts.

## Context

- `docs/game-design.md` → "The living deep" (Outcomes emerge), "Keeping it affordable" (Everyone is always an individual).
- `docs/go-technical-design.md` → "Level of detail"; "Designing for pivots" → "What stays concrete" (Outcomes emerge). AGENTS.md rule 10.
- The brief's "The one-world model" → "Everyone is an individual".
- What exists: `internal/core/agents/lod.go` (promotion, demotion, `StartLevelOfDetail`, the placer), `offscreen.go` (the per-phase drift, attrition and stock tick), `internal/content/families.go`, `content/data/families.cue`, `comp.Populations`, `comp.Family`, `comp.Individual`, and region `populations:` in the maps.

## Deliverables

- Delete `internal/core/agents/lod.go`, `offscreen.go` and their tests, and the `agents.lod` subscription and `agents.offscreen` system in `install.go`.
- Delete `internal/content/families.go`, `FamilyDef`, `DriftEnv`, `AttritionEnv`, `RegionDef.Populations`, `content/data/families.cue` and the `family.` prefix, with their schema (`go tool task schema`).
- `internal/core/comp/comp.go`: delete `Population`, `Populations`, `Family`, `Individual`.
- `internal/core/api/agents.go`: delete `FamilyCount`, `Promoted`, `Demoted`, `PopulationDied` and their kinds; `Migrated` loses `Short`, since only the off-screen tick set it (it goes whole in M4-07).
- `internal/core/agents/leave.go`: a body crossing a gate to a map the party is not on moves there as a body (placed at the gate's mark, or the nearest free cell) rather than being counted; `familyOf` and `familyListing` go.
- `internal/core/agents/recruits.go`: `desert` no longer adds `Individual`.
- `internal/core/game`: drop the `StartLevelOfDetail` call.
- Content: every region `populations:` in `content/data/maps/` becomes spawns of the same templates (marks in the layers). A count of N becomes N spawn cells. Any `generics` that relied on a family get their template directly (M4-02 deletes generics).
- Text: delete `text.agents.demoted` and any text only these events used.
- `internal/core/save`: bump `CurrentVersion`.
- Tests that covered promotion, demotion and drift are deleted; tests that used populations to put bodies on a map spawn them instead.
- `docs/glossary.md`: delete **family**, **aggregate**, **promote / demote** and **individual**. `docs/content-guide.md`: delete "Families", and update "The sample world".

## Spec

- **Bodies everywhere act.** With level of detail gone, every body with an `Actor` on every map is on the timeline. Nothing suspends a body for being on a map the party is not on. This is the interim until M4-07 makes one world, and M4-09 makes far agents cheaper.
- **No placer.** Content says exactly where each body starts. Nothing spawns a body at a random cell near the party.
- `timeline.Suspend`, `Resume` and `comp.Dormant` stay: M4-09 may use them, and they are not outcome systems. If nothing uses them once this task is done, say so in Notes.
- `api.Migrated` stays only for a body crossing a gate, until M4-07 deletes gates.
- Determinism: content spawns run in the order explore already uses (layers, rows, columns). Firm.

## Acceptance criteria

- [x] `git grep -w -E "Populations|FamilyCount|Promoted|Demoted|PopulationDied|StartLevelOfDetail|family\.[a-z_]+" -- internal content docs` finds nothing (`tasks/` aside).
- [x] `go run ./cmd/harness run --start start.sample --steps 40` runs, and the sample world's bodies exist from tick 0: a test counts the bodies on each sample map at the start and finds the numbers the old populations held.
- [x] A test shows a body on a map the party is not on still takes decisions (its `Doing` changes, or its position does) as time passes.
- [x] Every golden and harness output that changes is regenerated and named in Notes.
- [x] `go tool task check` passes.

## Out of scope

- Town stock, generics and `TownEmptied`: M4-02.
- Upkeep, desertion and flight: M4-03.
- Deleting maps, gates and `Migrated`: M4-07.
- Making far bodies cheaper: M4-09.

## Notes

- Deleted: `lod.go`, `offscreen.go`, `families.go`, `families.cue` and their tests, with `FamilyDef`, `DriftEnv`, `AttritionEnv`, `RegionEnv`, `comp.Population(s)`, `Individual`, `Family`, `PromoteRadius`.
- Deviation, `api.Migrated`: it is now `{Body, From, To}`. `Family` and `Count` meant nothing once a crossing body stays a body, so they went with `Short`. M4-07 deletes it.
- Deviation, `api.TownEmptied.Family` is now `Template`, and `TownDef.Residents` is a template ID (its living bodies on the map are the residents) since families are gone. Still reported only while the party is in the town. M4-02 deletes it.
- Beyond the list: `TownDef.Use`, `Refill`, `Exodus` went too (only the off-screen tick read them), so `map.sample.lake` no longer refills. `agents.promote_radius` became `agents.arrival_radius` (how far from a gate's mark a crossing or arriving body may be set down). Generics are marked from the bodies content spawns (`markGenerics` in `SpawnRegions`); `arrive` always sets a body down, never counts it.
- Sample world bodies are spawn cells (marks `p`, `H`, `G`, `F`, `T`, `L`, `m`, `r`, `f`, `w`); `map.play.hall` has four `c` cells (the old one plus the three the count made). Counts match the old populations: shallow 2/1/2/1, deep_a 3 clay and 2 remnants, deep_b 4/2/2/1, lake 3/1/1.
- `timeline.Suspend`, `Resume` and `comp.Dormant` have no user now; kept as the task says.
- Test fixtures: `contenttest.WithoutSpawns` strips a map's spawn cells (spawn marks may repeat, so setting `Spawns = nil` no longer validates). The agents, explore and combat fixtures use it. Offscreen drift, attrition, stock and promotion tests were deleted; new tests: `TestTheSampleWorldsBodiesExistFromTickZero`, `TestABodyOnAMapThePartyIsNotOnStillDecides`.
- Goldens regenerated: `internal/core/combat/testdata/encounters/*.jsonl` (all seven; UIDs shifted, and the outpaced migration line is the new shape), `internal/frontend/tui/testdata/{actor_sweep,actor_windup,look_after,look_before,menu_parts}.golden` (UIDs and ticks shifted by the extra bodies), `internal/harness/testdata/water_scenario.golden` (the whole sample world now acts, so moves and deaths rise). Save version is 9.
- Could not run golangci-lint here (the installed binaries are built with an older Go than go.mod's 1.27.4); `go vet`, the tests and `go tool task content` pass. `TestAfterEditWithGolangciLint` in agentkit fails for the same reason.
- Follow-ups filed: none.
