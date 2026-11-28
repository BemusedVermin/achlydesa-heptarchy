# C2 · Cleanup

**Goal:** after M6, reshape the core into a few components a person can hold in their head, each owning its own state and talking to the others only through an API it defines. Rules a designer would tune move out of the engine into content. Tests get content of their own, so tuning the game never breaks a test, and the game ships as one binary with its content inside.

The owner asked for this one directly, so its tasks come from his request and from C2-04's proposal, not from `/audit`. In his words: software is "a set of very simple services that interact over some API", and the engine should hold the core components "instead of allowing malleable things to seep into the engine".

**Exit criteria:**

- Every task is done or obsolete, and the lock-in is merged.
- The game-level tests C2-03 adds pass unchanged at the end: the refactor changed none of them.
- No test reads a value from the production content, and changing one changes no golden.
- The C2-03 benchmarks are no slower than they were before the refactor.

## Decisions (owner, 2026-11-27)

- C2 runs after M6. C2-01 depends on M6's last tasks and on the M6 follow-ups. If more follow-ups are triaged into M6 before C2 starts, add them to C2-01's `depends_on`.
- Production content lives in `content/data` and test content in `content/test`. Tests and goldens use only the test content. Tests may only load and validate the production content.
- The released binary embeds the production content.
- When the refactor finds a rule written in Go that a designer would tune, the same pass moves it into content. It isn't left for later.
- CUE stays for configuration. Behavior may need a language of its own, because rules written in CUE are awkward, some need steps, branches and loops, and they should be easy to organize. C2-04 sets out the engine's vocabulary as the API that content programs against, compares the options on the same three rules, and files the choice as a question for `/decide`.
- The refactor gates on a proposal. C2-04 maps the core and proposes its new shape, and the owner approves it in that pull request. C2-04 also files the implementation tasks as `proposed`.

## Order

1. **C2-01:** split the test content from the production content, and rebuild every golden on the test content.
2. **C2-02:** embed the production content in the binary.
3. **C2-03 and C2-04, side by side.** C2-03 pins every behavior with a game-level test and adds the benchmarks. C2-04 maps the core as it is, writes the proposal and files the implementation tasks.
4. **The owner reviews C2-04's pull request,** then runs `/decide` for any questions it filed, and `/triage` to accept the implementation tasks.
5. **The implementation tasks** (C2-05 onward), in the order the proposal gives.
6. **The docs task, then the lock-in,** both filed by C2-04. The lock-in runs in the owner's session, because it changes `AGENTS.md`, `.golangci.yml` and `.claude/`.

Once the lock-in is merged, run `/optimize` for the first speed pass. From then on the benchmarks measure the new shape.

## Tasks

| ID | Title | Kind | Size | Depends on |
| --- | --- | --- | --- | --- |
| [C2-01](C2-01-test-content.md) | Test content apart from production content | debt | L | M6-27, M6-28, M1-11.F2 |
| [C2-02](C2-02-embed-content.md) | One binary: embed the production content | feature | M | C2-01 |
| [C2-03](C2-03-pin-behavior-and-benchmark.md) | Pin every behavior with a game-level test, and benchmark the slow paths | debt | L | C2-02 |
| [C2-04](C2-04-map-and-propose.md) | Map the core as it is, and propose its new shape | debt | L | C2-02 |
