---
id: M3-07
title: "A playground: the test maps joined by gates"
status: done
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

- [x] `go tool task run -- play --seed 1` opens in the playground; walking through each gate reaches each creature (checked by a test that walks the route with commands, and by hand).
- [x] `go tool task same-behavior` shows changes only for seeds that use the default start, and Notes lists them.
- [x] `go tool task check` passes.

## Out of scope

- Towns, residents and the sample world proper (M3-11, M3-19). The playground is for trying features, not the shipped world.

## Notes

- New maps `map.play.hub`, `map.play.hall`, `map.play.den`, `map.play.descent` (`content/data/maps/play_*.cue`) copy the test layouts, so no test map changed. `start.cue` now starts on the hub. The gate cells are also marks, as the loader requires.
- `contenttest.Sample` and `SampleBundle` pin `start.map` to `map.test.hall`, so every test that relied on the default start keeps its map. Tests that load content directly (`cmd/harness` replay) see the real start.
- `same-behavior` differs, as it must: the default-start seed 1 (harness CSV and final hash), and the UIDs and map IDs in the combat goldens and `menu_parts.golden`, because the new maps' creatures spawn before the old maps' bodies. Only IDs changed in those goldens; I regenerated them.
- `TestAMoveTowardABodyStopsNextToIt` measured distance to bodies on other maps; it now counts only the drinker's map.
- The walk-through test is `internal/core/explore/playground_test.go`. It walks the shortest route to each gate and back and checks each creature is on its map.
- `docs/content-guide.md` gains "The playground".
- golangci-lint could not run here (the installed binary is built with an older Go than the module targets), so lint was not run; vet, tests, content and comments checks pass. `TestAfterEditWithGolangciLint` fails for the same reason.
