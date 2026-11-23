# Lens: tests

Weigh each test by what it protects against what it costs, and report the ones that don't earn their keep.

**What it protects.** High value:

- game rules and content rules;
- the invariants the design depends on: determinism, the layers, saves that round-trip, content that validates;
- anything a player or designer would notice breaking.

Low value: infrastructure (the harness, tools, command-line wiring), which needs a thin smoke test, not exhaustive checks. Judge by what the test protects, not by where it lives. For example, a harness test that two runs of the same seed match is guarding the core's determinism, so it stays, but small.

**What it costs:**

- **Run time.** List the slowest tests in your area with `go test -count=1 -v <packages> 2>&1 | grep -E '^\s*--- (PASS|FAIL)' | sort -t'(' -k2 -rn | head -20`. Anything over a second needs a reason.
- **Upkeep.** Golden files that change with every content tweak, exact-string assertions on wording, and mocks that mirror the implementation.
- **Flakiness:** sleeps, wall-clock time, order dependence.

**What to look for:**

- **Redundancy.** The same behavior asserted at several levels (unit, integration and golden); keep the cheapest one that would catch the bug. Table cases or property tests that cover the same ground twice.
- **Trivial targets.** Getters, constructors that only assign, tests that only check a mock was called.
- **Oversized runs.** Seeds, steps, iterations or `rapid` checks far beyond what the bug they hunt needs. A determinism check needs a few dozen steps, not thousands.
- **Wrong subject.** Tests of the test harness or tooling that run longer than the code they protect deserves.

Report each finding with:

- the test name and `file:line`;
- its run time;
- what it protects, in a few words;
- the recommendation: delete, shrink to a stated size, merge into another named test, or keep with a reason;
- the size of the change.
