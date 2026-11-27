---
id: C2-03
title: Pin every behavior with a game-level test, and benchmark the slow paths
status: todo
milestone: C2
kind: debt
depends_on: [C2-02]
discovered_by: null
size: L
---

# C2-03 · Pin every behavior with a game-level test, and benchmark the slow paths

## Goal

Before the refactor moves packages around, every behavior the game has today gets a test that drives the game from the outside: commands in through `game.Game.Apply`, events and views out. The refactor can then change everything underneath without touching those tests, and a broken behavior shows up as a failing test. This task also adds benchmarks for the three paths to speed up (a game step, building the world, and harness runs), so the refactor can show it made nothing slower and `/optimize` has something to measure.

## Context

- `docs/game-design.md`, section by section: what the game is meant to do. The code is what it does today.
- `internal/core/game`: `New`, `Apply`, `View`, `Hash`, `Replay`. `internal/core/game/gametest`: `Start` and `Reload`. `internal/harness`: `Step`, `Play`, `Run`, the random policy, and the replay goldens.
- The test content from C2-01, and the embedded production content from C2-02.
- AGENTS.md, Style, "Tests": this is a hobby project, so test the rules that can really break and add no near-duplicates.
- The two benchmarks that exist today: `internal/core/agents/sense_test.go` and `internal/core/hexmap/sight_test.go`.
- `go tool task bench`.

## Deliverables

- `tasks/C2-cleanup/coverage.md`: the coverage map.
- New game-level tests (`_test.go`) in `internal/core/game` and `internal/harness`, with goldens under `testdata/` where output is large.
- Additions to `content/test` that those tests need.
- Benchmarks (`_test.go`) in `internal/core/game` and `internal/harness`.

## Spec

1. **The coverage map** (firm). `coverage.md` has one row per behavior the code implements today, grouped by the `game-design.md` section it comes from. That covers exploration (steps, climbing, drops and the fall injury, sight and memory, the plan view, reachability), combat (forms, types, timing, lethality, grafts, crawling parts, being downed and the wipe), actions and the action menu, the pack, the timeline and turns, minds (drives, goals, both planners, far agents, hearsay, fleeing and engaging), towns and residents, hiring and recruits, procgen (chunks, regions, settlements, the populated deep), saves, replay, and the frontend's plan view. Each row has these columns:
   - the behavior;
   - its doc section;
   - the tests that pin it, as file and test name;
   - whether one of them is game-level.
2. **What "game-level" means** (firm). A test is game-level when it imports only `game`, `game/gametest`, `api`, `content`, `content/contenttest` and `harness`, and no feature or foundation package. Package-level tests stay. They move with their packages, and the refactor may rewrite them. Game-level tests must not need to change.
3. **Fill the gaps** (firm). Every behavior with no game-level test gets one. Prefer one test per behavior, table-driven where several real cases exist, with a golden for large output. Edge cases stay in the package tests. Every test runs on the test content. Add to `content/test` what a test needs, and never read `content/data`.
4. **Benchmarks** (firm). They run on the embedded production content, because they measure the game players get. A benchmark never fails, so tuning production values breaks nothing. This is the one place besides C2-01's validate test where code under test loads production content. Use `b.Loop`. Put them in `game` and `harness`, where they survive the refactor. Suggested names:
   - `BenchmarkStep/explore` and `BenchmarkStep/combat`: one player step at a fixed seed, after enough warm-up steps that the deep is populated and far agents run. The second is with an encounter under way.
   - `BenchmarkNewGame`: starting a game at a fixed seed, which builds and populates the world.
   - `BenchmarkExplore`: the party walking across several chunk boundaries, so new chunks are built.
   - `BenchmarkHarnessGame`: one headless game of a fixed number of steps. `BenchmarkHarnessRun` runs several games side by side.
5. **The baseline** (firm). Run `go tool task bench BENCH=. -- -count 6 ./internal/core/game ./internal/harness` and put the output in Notes, with the CPU line. The refactor tasks compare against runs on their own machine, so this is a reference, not a gate.
6. **Code stays as it is** (firm). This task changes only `_test.go` files, `testdata/`, `content/test/` and `tasks/`. If a behavior is broken, file a follow-up and pin the current behavior with a test named to say so.

## Acceptance criteria

- [ ] `coverage.md` lists every behavior, and each row names at least one game-level test.
- [ ] No new test imports a package outside the game-level list (check the import blocks).
- [ ] No new test reads `content/data`.
- [ ] `go tool task bench BENCH=. -- ./internal/core/game ./internal/harness` runs every new benchmark, and the baseline is in Notes.
- [ ] `git diff --name-only origin/master...HEAD` lists only `_test.go` files, `testdata/`, `content/test/` and `tasks/`.
- [ ] `go tool task check-pr` passes.
- [ ] `go tool task check` passes.

## Out of scope

- Fixing behavior the tests show is broken: file a follow-up.
- Speeding anything up: `/optimize`, after C2.
- Deleting or merging the existing package tests: the refactor tasks do that as their packages move.

## Notes
