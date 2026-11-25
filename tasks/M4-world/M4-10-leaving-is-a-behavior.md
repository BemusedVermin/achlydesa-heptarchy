---
id: M4-10
title: Leaving is a behavior
status: todo
milestone: M4
kind: feature
depends_on: [M4-08, M4-09]
discovered_by: null
size: M
---

# M4-10 · Leaving is a behavior

## Goal

An agent whose needs aren't met where it is picks a region where they could be, ranked by content formulas, and walks there, or travels the coarse graph when far from the party. Migration and a town emptying fall out of it.

## Context

- `docs/game-design.md` → "People with needs" (Towns can empty), "Monsters with needs" (remnants retreat toward undecayed ground).
- `docs/content-guide.md` → "Behaviors" (`rank`).
- The brief's "The one-world model" → "Leaving".
- What exists: `content.BehaviorDef.Rank` and `RankEnv` (`source`, `region`, `distance`), the deleted gate-based `leave` (in git history before M4-07), M4-08's regions and M4-09's far travel.

## Deliverables

- `internal/core/agents/leave.go`: the `leave` primitive. It ranks every region (in region ID order) other than the one it stands in, walks toward the chosen region's nearest cell (`move_toward` near, the coarse graph far), and is done once it stands in that region.
- `internal/content`: `RankEnv` gains what a destination needs to be judged from its region record alone; `distance` is hexes to the region's nearest box. Regenerate the schemas if it changes.
- `content/data/behaviors.cue` and `templates.cue`: residents leave when thirst stays high, ranking regions by what their records say (see Spec); remnants leave for lower decay; the M4-03 deserter walks to a region tagged as its kind's home.
- Tests (below).
- `docs/content-guide.md` → "Behaviors": `rank` ranks regions.

## Spec

- **A rank formula reads a region record**, never a count of what is in it: no "how much water is there" from Go (that would be the abstract stock back). A resident leaves when its own need is high and ranks regions by their props and tags, such as `lists`/`tags` saying a region is a crawlspace with water. Provisional: what it knows of distant water is the region's data until rumors (M7) carry real news.
- **The choice holds.** Once chosen, the destination is kept in `comp.Doing` (its `Cell` the target cell) until reached or the behavior loses; it doesn't re-rank every decision.
- Ties: the highest rank, then the nearest, then the lower region ID. Firm.
- A rank of 0 or less is never chosen; with none left the behavior is done and scores as its considerations say.

## Acceptance criteria

- [ ] A test: a resident of a town whose cisterns and well are dry walks out and arrives in another crawlspace region, with no event but its steps.
- [ ] A test: a remnant in a high-decay region leaves for the lowest-decay region it ranks highest, near and far.
- [ ] A test: a deserter walks away from the party's region.
- [ ] `go tool task check` passes.

## Out of scope

- Digging and settling a new crawlspace, and leaving the world at emigration points: M7.
- Measuring outcomes: M4-13.

## Notes

Filled in by the implementer.
