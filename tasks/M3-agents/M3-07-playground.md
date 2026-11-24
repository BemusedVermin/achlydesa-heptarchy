---
id: M3-07
title: "A playground: the test maps joined by gates"
status: in-progress
milestone: M3
kind: content
depends_on: [M3-06]
discovered_by: null
size: M
---

# M3-07 · A playground: the test maps joined by gates

## Goal

The owner can play every M3 feature by hand. The test hall, the den and the descent are joined by gates into one playground that `go tool task run -- play --seed 1` opens in, and each later M3 task adds what it builds to it.

## Context

- The owner asked for this while M3 was being planned, so features can be tried as they land.
- `content/data/maps/test_hall.cue`, `test_den.cue`, `test_descent.cue`, `content/data/start.cue`; M3-06's gates.
- `docs/content-guide.md` (maps, marks and spawns).

## Deliverables

- `content/data/maps/play_*.cue` (new) or gates added to the existing test maps (see Spec).
- `content/data/start.cue`: start in the playground.
- `content/data/maps/README` comment or `docs/content-guide.md`: a short section, "The playground", listing what is where.
- Tests or goldens updated only where the start map changed, each named in Notes.

## Spec

- Firm: `go tool task run -- play --seed 1` opens in the playground with the main party, and every one of the three sample creatures (moving clay, remnant, beast) can be reached by walking through gates.
- Firm: the existing tests that start on a named map keep their map and their results. A test that relied on the default start sets `start.map` explicitly, as the den tests already do.
- Suggestion: copy the three layouts into new `map.play.*` maps with gates, rather than adding gates to the test maps, so no test map changes. Add a small hub room the party starts in, with one gate to each.
- The playground grows: later M3 tasks add a line to "The playground" in the content guide for what they put there (a town, residents, recruits, families).

## Acceptance criteria

- [ ] `go tool task run -- play --seed 1` opens in the playground; walking through each gate reaches each creature (checked by a test that walks the route with commands, and by hand).
- [ ] `go tool task same-behavior` shows changes only for seeds that use the default start, and Notes lists them.
- [ ] `go tool task check` passes.

## Out of scope

- Towns, residents and the sample world proper (M3-11, M3-19). The playground is for trying features, not the shipped world.

## Notes


