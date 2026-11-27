---
id: M6-04
title: A capped memory per agent, learned by sight
status: done
milestone: M6
kind: seam
depends_on: []
discovered_by: null
size: M
---

# M6-04 · A capped memory per agent, learned by sight

## Goal

An agent remembers what it has seen, up to a cap, forgetting the oldest and least certain first. Nothing else it knows comes from the truth.

## Context

- `docs/game-design.md` → "How minds work" → "Beliefs"; "Keeping it affordable".
- `internal/core/knowledge`: the fact store (subject, claim, origin, confidence) per knower; today only the party writes to it.
- Sight: `hexmap` sight and the `combat` alert's visible-cells query. Sight per agent is the hot spot at 1,000 agents.

## Deliverables

- `internal/core/knowledge/`: `Learn` keeps at most `knowledge.cap` facts per knower (content), dropping the lowest by (confidence, tick); `Forget`.
- `internal/core/agents/`: `sense(w, body)`, which learns what a near agent sees within `agents.sight_radius`: bodies, landmarks (M6-07) and places worth remembering, with a content list of which claims to record.
- `content/data/knowledge.cue`: `claim.at` (a thing is at a cell), `claim.dead`, `origin.seen`, `origin.told`; first values for the cap and radius.
- Tests: the cap holds and the right fact is dropped; a body sees only inside its radius and line of sight; the party's own facts are not capped.

## Spec

- **Firm:** `sense` runs only on an agent's own timeline entries, capped by radius, never on every entry of the world.
- **Firm:** a belief about a moving thing is a claim with a tick; a contradicting sighting replaces it (same subject, claim, origin).
- **Firm:** the cap and radius are content. First values: 64 facts, 8 hexes, for tuning in M9.
- A fact about a thing gone (a well that was a place's last drop) is not corrected until seen again; that is the point.

## Acceptance criteria

- [x] The tests above pass.
- [x] A benchmark in the test file stays within a stated budget for 1,000 near agents sensing once (recorded in Notes).
- [x] `go tool task check` passes.

## Out of scope

- Telling facts to other agents: M6-18.
- Planning over beliefs: M6-10.

## Notes

- **Benchmark budget:** `BenchmarkThousandAgentsSenseOnce` (`internal/core/agents/sense_test.go`): 1,000 bodies on a 64×64-hex, 4-depth open area, each sensing once with a bus drain after each. Budget 150 ms per pass; measured 60–80 ms on a 4-core 2.3 GHz Xeon. Most of it is line-of-sight checks and `Learn`'s eviction churn once a knower is past its cap.
- **Content shape:** "a content list of which claims to record" is the `agents` config in `knowledge.cue`: `at_claim` and `dead_claim` name the claims, and an empty one records nothing. `seen_origin` and `seen_confidence` (first value 1) say how sure a sighting is. I added `sight_depth` (first value 6, the same as the party's and combat's) because sight needs a depth range. `knowledge.cap` is 64 and `agents.sight_radius` is 8. Both configs sit in `knowledge.cue`, as the Deliverables say.
- **What is sensed:** bodies (entities with `comp.Pos` and `comp.Body`), giving `claim.at` (value `"q,r,depth"`, written by `knowledge.CellValue` and read back by `knowledge.ParseCell`) and `claim.dead` (empty value). Landmarks and ground (the "places worth remembering") come with M6-07, which extends `sense`. No place entities exist yet.
- **Who calls `sense`:** no mind controller exists yet, so only the tests call `sense`. M6-10/M6-11's controller calls it from its own `Decide`, which keeps the firm "own timeline entries only" rule.
- **Cap:** `Learn` evicts the lowest (confidence, tick), then the lowest key, so learning order cannot change the outcome. If the new fact is the one evicted, it publishes nothing. The controlled party and its members are never capped. With no content loaded (`w.Content == nil`, as in older tests), nothing is capped.
- **Outside the Deliverables:** `hexmap.Map.Sees` (`internal/core/hexmap/sight.go`) checks one line of sight without building a whole visible field, which keeps sensing cheap. A rapid test checks that it agrees with `Visible`. The other files: `internal/content/{types,load,registry}.go` (configs, getters, checks), `internal/content/load_test.go` (minimal content gains the two configs), the regenerated schemas, `comp.Knowledge`'s doc comment, `content/data/text/knowledge.cue`, and one line each in `docs/content-guide.md` and `docs/go-technical-design.md`.
- **Save:** no component changed shape, so the save version is not bumped.
- **Follow-ups:** M6-04.F1 (each re-sighting publishes `FactLearned`, so many agents sensing before one drain pass the bus's 10,000-event cascade limit).

