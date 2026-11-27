---
id: C2-01
title: Test content apart from production content
status: todo
milestone: C2
kind: debt
depends_on: [M6-27, M6-28, M1-11.F2]
discovered_by: null
size: L
---

# C2-01 · Test content apart from production content

## Goal

Tests and goldens run on test content of their own, in `content/test`. The production content in `content/data` is what the game ships. Tuning a production value then breaks no test and changes no golden. Today every test plays the real content through `contenttest.Fixture`, and that has been a big source of churn.

## Context

- `internal/content/contenttest/contenttest.go`. `Sample` and `SampleBundle` load `content/data`, and `Fixture` shrinks its world and moves the three test places side by side. `Full` loads the real world, and only `internal/core/game/large_world_test.go` calls it.
- The test places live in the production content today: `content/data/places/test_hall.cue`, `test_descent.cue` and `test_den.cue`.
- The goldens live in `internal/core/combat/testdata`, `internal/core/game/testdata`, `internal/core/worldgen/testdata`, `internal/frontend/tui/testdata` and `internal/harness/testdata`.
- `docs/content-guide.md`, and `docs/go-technical-design.md` under "Content pipeline" and under "Harness and testing", "Tests".
- AGENTS.md, Style: the cached sample registry and `SampleBundle`.
- `go tool task content` runs `cue vet ./content/...` and `cue fmt --check ./content/...`, so it already covers a new folder under `content/`.

## Deliverables

- `content/test/`: the test content, a complete bundle that `content.Load` accepts.
- `content/data/places/`: the three test places move out.
- `internal/content/contenttest/`: `Sample` and `SampleBundle` load `content/test`. `Fixture` goes away.
- One test that loads and validates `content/data`.
- Every test that plays production content today moves to the test content.
- Every golden, regenerated on the test content.
- `docs/content-guide.md`, `docs/go-technical-design.md` and `docs/glossary.md`.

## Spec

1. **Two content sets, one schema** (firm). Production content is `content/data`, and test content is `content/test`. They share the generated schema and the content kinds, and each is a complete bundle on its own.
2. **The test content stands alone** (firm). It doesn't import production files or take values from them. It holds only what the tests need: a small world, the test places, and the materials, bodies, forms, terrains, items, actions, text and so on that the tests use, with values chosen for the tests. IDs may match production IDs where a test reads better that way (suggestion), but the values belong to the test set.
3. **The test world is the fixture** (firm). The test content defines the small world and places the test places directly, so `Fixture` is no longer needed. `Sample` and `SampleBundle` keep their contract: `Sample` returns one cached, read-only registry, and `SampleBundle` returns a copy a test may edit.
4. **Tests only load and validate the production content** (firm, owner decision 2026-11-27). Exactly one test loads `content/data`, with `content.Load`, which compiles formulas and checks references. It fails only on a load or validation error.
5. **Production-content tests move** (firm). Every test that plays the production content today moves onto the test content, for example `large_world_test.go` and the step budget from M6-28. If what it checks is production tuning, it becomes a benchmark instead, or is dropped. Notes list each one and where it went.
6. **Goldens** (firm). Regenerate every golden on the test content. This pull request changes goldens, and that is expected. After it, changing a production value changes no golden.
7. **Go that names content IDs** (firm). If engine code names a content ID that the test content must therefore also define, list each one in Notes, with `file:line`. C2-04 decides where those rules go.
8. **Docs** (firm):
   - The content guide gets a section on test content and production content: what goes where; that tuning `content/data` never breaks a test; that a test needing a value sets it in `content/test`; and that a new content kind used by tests needs entries in both sets.
   - The tech design's "Tests" section says goldens run on the test content.
   - The glossary gets **production content** and **test content**, and retires "sample content".
   - AGENTS.md is protected on a task branch. Put the wording its Style line on the sample registry should take in Notes, and the C2 lock-in applies it.

## Acceptance criteria

- [ ] `content/test` loads with `content.Load`, and `go tool task content` passes on both sets.
- [ ] `content/data` has no `place.test.*` places.
- [ ] Only the load-and-validate test reads `content/data` (search the tests for the path and for the production loader).
- [ ] Change a few production values, then run `go tool task test`. Every test passes, and no golden changes. Use `plan_range` in `content/data/party.cue`, one body's numbers in `bodies.cue` and one terrain's numbers. Revert them after, and record what you changed in Notes.
- [ ] `go tool task run -- play --seed 1` still starts a game on the production content.
- [ ] The content guide, tech design and glossary are updated.
- [ ] `go tool task check-pr` passes.
- [ ] `go tool task check` passes.

## Out of scope

- Embedding the content in the binary: C2-02.
- New game-level tests and benchmarks: C2-03.
- Moving rules from Go into content: the C2 implementation tasks, filed by C2-04.

## Notes
