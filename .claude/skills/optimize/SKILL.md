---
name: optimize
description: Find what makes the game slow. Runs the benchmarks with the profiler, has read-only perf-auditor agents rank the bottlenecks in the game step, world building and the harness, and files the fixes as a P<n> milestone of proposed tasks. Use when the owner wants to speed the game up or find its bottlenecks.
argument-hint: "[step] [procgen] [harness]"
disable-model-invocation: true
---

# Find the bottlenecks, then file the fixes

Arguments: $ARGUMENTS

A speed pass works like a cleanup:

1. `/optimize`: measure, have read-only auditors find the bottlenecks, and file the fixes as a speed milestone `P<n>`.
2. `/triage`: the owner accepts, merges or drops the tasks.
3. `/run-tasks`: agents do them, each proving its gain with a benchmark.

## 1. Before starting

1. `git fetch origin`, and measure `origin/master`: a clean checkout or a fresh worktree, with nothing uncommitted.
2. **The benchmarks must exist.** List them with `go test -list 'Benchmark.*' ./internal/core/game ./internal/harness`. They come from C2-03. If a target below has none, say which, and stop: its benchmark is a task of its own.
3. **A running cleanup.** If a `tasks/C*-cleanup/` milestone has tasks that aren't done or obsolete, every task this pass files depends on that cleanup's lock-in. The refactor would otherwise collide with the speed work.
4. The milestone is `P<n>`, one higher than the last `tasks/P*-perf/` folder (`P1` if there is none). Read the last one's brief, if any, for what it already fixed.

## 2. The targets

The arguments pick them, and the default is all three:

| Target | What it measures | Benchmarks |
| --- | --- | --- |
| `step` | A player's action resolving, far agents included | `BenchmarkStep` and its sub-benchmarks in `internal/core/game` |
| `procgen` | Building the world: a new game, and new chunks as the party explores | `BenchmarkNewGame` and `BenchmarkExplore` in `internal/core/game` |
| `harness` | Headless games, alone and side by side | `BenchmarkHarnessGame` and `BenchmarkHarnessRun` in `internal/harness` |

If the names have changed since, map each target to its benchmarks from the `go test -list` output.

## 3. Measure, one target at a time

Benchmarks running side by side skew each other's timings, so run them one after another here and nowhere else. Make a temporary folder outside the repository, called `<tmp>` below, and for each target:

1. Timings: `go tool task bench BENCH=<pattern> -- -count 6 <package>`, saved to `<tmp>/<target>.txt`.
2. Profiles: `go tool task bench BENCH=<pattern> -- -count 1 -cpuprofile <tmp>/<target>.cpu -memprofile <tmp>/<target>.mem -o <tmp>/<target>.test <package>`. `-o` keeps the test binary out of the repository.

Note the Go version and the CPU line from the output.

## 4. Start the auditors

Start one `perf-auditor` agent per target, all at once, in the background. They only read the profiles, so running them side by side is fine. The prompt is:

`Target <target>: benchmarks <names> in <package>. Results: <tmp>/<target>.txt. CPU profile: <tmp>/<target>.cpu. Memory profile: <tmp>/<target>.mem. Test binary: <tmp>/<target>.test.`

Say what you started. Wait for every report.

## 5. Turn the findings into tasks

1. **Merge** findings that several auditors reported. A hot spot shared by the step and the harness is one task.
2. **Drop** what isn't worth a task, and say why in the brief.
3. **Questions** (`kind: question`), for `/decide`: a fix that changes how the game plays, even slightly, such as coarser far simulation or a cap on agents. That is a design decision (AGENTS.md rule 9).
4. **One task per fix,** with `kind: perf`. Fixes that touch the same code depend on each other, so they run one at a time. Each task's Spec carries the auditor's evidence and fix. Its acceptance criteria are:
   - the named benchmark is at least the expected share faster. Compare the median of `-count 6` before and after, on the same machine, with `go run golang.org/x/perf/cmd/benchstat@latest before.txt after.txt`, and put the numbers in Notes.
   - every golden and every game-level test is unchanged: the same seed plays the same game;
   - no other target's benchmark is more than 5% slower;
   - `go tool task check-pr` and `go tool task check` pass.

## 6. Write it down

On a new branch from `origin/master` named `perf/P<n>`:

- **`tasks/P<n>-perf/README.md`,** the milestone brief:
  - the goal;
  - the baseline, one row per benchmark with its median time, allocations and bytes, and the machine;
  - the findings by target, with what was dropped and why;
  - the order;
  - the exit criteria: every task done or obsolete.
- **One task file per task:** `P<n>-NN-<slug>.md`, from `tasks/TEMPLATE.md`, with `status: proposed`, `milestone: P<n>`, a kind, a size and `depends_on`.
- **`tasks/INDEX.md`:** a `P<n> · Speed` section after the current milestone, one row per task.

Commit, push, and open a pull request as in step 7 of the do-task skill. Its body has:

- the baseline;
- the biggest expected wins;
- the questions;
- what was dropped;
- the next steps: `/decide` if there are questions, then `/triage`, then `/run-tasks`.

Tell the owner in three or four lines: how many tasks, the biggest expected wins, the pull request's link, and that `/triage` comes next.
