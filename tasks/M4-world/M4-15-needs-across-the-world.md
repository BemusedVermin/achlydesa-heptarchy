---
id: M4-15
title: Needs reach across the world
status: todo
milestone: M4
kind: feature
depends_on: [M4-09, M4-10]
discovered_by: null
size: M
---

# M4-15 · Needs reach across the world

## Goal

An agent goes for whatever meets its needs, anywhere it knows of in the world, walking near the party and travelling the coarse graph far from it. Knowing of nothing, it sets off at random and learns as it goes. There is no `leave`: a town empties because its people found what they need elsewhere, and an agent lives wherever it ends up resting.

## Context

- `docs/game-design.md` → "The living deep" (Outcomes emerge), "People with needs" (Knowing where, Towns can empty), "Monsters with needs" (remnants retreat toward undecayed ground).
- The brief's "The one-world model" → "Needs across the world".
- What exists:
  - `chooser.resolve`, `nearestBody` and `nearestCell`;
  - `content.BehaviorDef.Rank` and `RankEnv`;
  - M4-09's far travel, and M4-10's knowing.

## Deliverables

- `internal/core/agents`: `resolve` finds a behavior's target among what the agent can see and what it knows of, anywhere in the world. Each candidate is scored by the behavior's `rank` formula, which sees the candidate and the cost of getting there (`distance` in hexes along the route, or along the coarse graph when far). The best wins. Near agents walk with `move_toward`; far agents travel M4-09's graph.
- **Knowing nothing.** A behavior with no known target sets off along a random path, drawn from a named RNG stream, and re-resolves as it learns. How far it wanders is the behavior's params.
- **Resting.** Rest behaviors target the best rest spot the agent knows, like any other target. Where it rests is where it lives; nothing records a home.
- `internal/content`: `RankEnv` gains what a candidate needs to be judged, such as its region's data for ground, and `distance` becomes route distance. Regenerate the schemas.
- `content/data/behaviors.cue` and `templates.cue`: residents drink, eat and rest from what they know; remnants go for less-decayed ground they know of; a recruit who has quit (M4-03) goes on meeting its needs like anyone else.
- Tests (below). `docs/content-guide.md` → "Behaviors": targets come from what the agent knows, and what `rank` sees.

## Spec

- **No region speaks for what's in it.** A destination is a thing, a body or ground the agent knows of, never a region record claiming water. Firm.
- **No leave, no migrate.** No primitive, behavior or event exists to move an agent elsewhere for its own sake. Firm (owner, 2026-11-25).
- **The choice holds.** A chosen target stays in `comp.Doing` until reached, known to be gone, or the behavior loses, so the agent doesn't re-rank every decision.
- Ties: the highest rank, then the nearest, then the lower UID or cell. A rank of 0 or less is never chosen.

## Acceptance criteria

- [ ] A test: when a crawlspace's water runs out, a resident who has seen water in another crawlspace walks there, drinks, and rests there afterwards, with no event but its own actions.
- [ ] A test: an agent that knows of no water sets off at random and goes for water once it sees some.
- [ ] A test: a remnant goes for less-decayed ground it knows of, near the party and far from it.
- [ ] A test: the same seed gives the same paths, random ones included.
- [ ] `git grep -w -E "leave|migrate" -- internal/core/agents content/data` finds no primitive or behavior by those names.
- [ ] `go tool task check` passes.

## Out of scope

- Drives beyond survival, such as fulfillment and wishes: M7.
- Hearsay: M7's rumors.
- Digging and settling a new crawlspace, and leaving the world at emigration points: M7.
- Measuring outcomes: M4-13.

## Notes

Filled in by the implementer.
