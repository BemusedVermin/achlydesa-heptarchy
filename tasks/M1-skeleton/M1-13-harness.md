---
id: M1-13
title: "harness: seeded runs and replays"
status: todo
milestone: M1
kind: feature
depends_on: [M1-12]
discovered_by: null
size: S
---

# M1-13 · harness: seeded runs and replays

## Goal

A headless binary that plays many seeded games with a simple policy and writes CSV, and that replays any recorded log to check its hash. Later milestones add columns and policies, not a new tool.

## Context

- Tech design: "Harness and testing → The harness".
- M1-09 `game.Replay`, M1-12 `game.Hash`.

## Deliverables

- `cmd/harness/main.go`: cobra commands.
- `internal/harness/run.go`, `replay.go`, `harness_test.go`: the logic, importable by tests.

## Spec

**`harness run --seeds N [--seed-start S] [--steps M] [--content DIR] [--out FILE]`**

- For each seed from `S` to `S+N-1`, start a game and take `M` steps.
- Each step picks uniformly among the view's enabled options, using a policy generator seeded from the game seed (`rand.New(rand.NewPCG(seed, 0x6861726e657373))`; randomness outside the core is allowed).
- Write one CSV row per step with `encoding/csv` and the columns `seed, step, tick, day, phase, map, q, r, depth, events, rejected`. Read the leader's position from its `Pos` component (through a small test-support accessor on `game`), not from the view, which has no depths.
- After each seed, log one `slog` line with the seed and the final hash.

Write the policy behind a small interface so later milestones can add smarter ones:

```go
type Policy interface {
	Choose(view api.View) (api.Option, bool)
}
```

**`harness replay FILE --seed N [--content DIR] [--expect HEX]`** prints the final hash, and exits 1 if `--expect` doesn't match.

## Acceptance criteria

- [ ] `harness run --seeds 3 --steps 200` writes 600 rows.
- [ ] Two runs with the same arguments produce byte-identical CSV (test calls `internal/harness` directly).
- [ ] `harness replay` of a log written with `Codec.WriteLog` from a test game prints that game's `Hash`.
- [ ] `--expect` with a wrong hash exits 1.
- [ ] `go tool task check` passes.

## Out of scope

- Days-long runs and metrics beyond the columns above (M7).

## Notes
