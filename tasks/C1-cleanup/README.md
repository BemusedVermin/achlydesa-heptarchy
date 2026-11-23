# C1 · Cleanup

**Goal:** the first cleanup, after M2. Settle the names that drifted, remove the code and tests that were copied between packages, speed up the slow tests, and tighten the comments, so M3 starts from a smaller and more consistent codebase.

**Exit criteria:** every task is done or obsolete, and the lock-in is merged.

## How it is organized

The auditors reported well over a hundred findings in 7 areas. They are grouped into **10 tasks, one pull request each**, plus a lock-in. Each task is one agent's pass over a group of packages, doing duplication, then tests, then comments, in that order, because tests follow the code they test and comments follow both. The questions are not pull requests: `/decide` answers all of them in one.

Findings that several auditors reported are merged: UID ordering and liveness checks (actions, foundation, features), the shared test fixture (content, game, foundation, features), the timeline entry kinds and option IDs (actions, game), the template name text ID (actions, features), and the party and leader queries (foundation, features).

## Findings by area and lens

| Area | Findings |
| --- | --- |
| Content | Duplication 5 (collections table, reference checks, formula runner, CUE shapes, stdlib), tests 5, comments about 21, conventions 4 (including a duplicate-error bug). |
| Actions | Duplication 5, tests 5, libraries 1, conventions 4, comments about 15. |
| Features | Duplication 7, tests 5, libraries 3, conventions 3, comments about 14. |
| Foundation | Duplication 5, tests 4, libraries 2, conventions 3, comments about 8. |
| Game and frontend | Duplication 4, tests 6, libraries 1, conventions 5, comments about 6. |
| Tools | Duplication 3, tests 4 (one a failing test on Windows), libraries 3, conventions 3, comments about 7. |
| Vocabulary (whole codebase) | 12 findings: 11 contested names (16 questions in all) and design-doc contradictions. |

**Dropped:** goconst hits on test literals (the auditors found no real ones); `TestLethalityTargets` is kept (it is worth its 8 seconds); harness policy numbers in Go (game audit V5), because the harness is outside the core and the owner hasn't asked for them to be data; `Fprintln` error-check consistency in tools (errcheck is configured to allow it); the modernize analyzer could not be run by the auditors, so the library findings come from reading, and `modernize` is tried at lock-in.

## Order

1. `/decide` the questions (C1-01 to C1-16).
2. Renames first, because every later task reads the settled names.
3. Shared helpers and the shared test fixture, because the area tasks use them.
4. The five area tasks and the tools task, which can run side by side: they touch different packages.
5. The agentkit task, which the owner does, and the docs task.
6. The lock-in, last, run by the owner.

The Windows test fix in the tools task is the one finding that matters now: `go tool task check` fails on the owner's machine until it lands. If that is a problem sooner, the owner can lift that part into its own task.

## Tasks

| ID | Title | Kind | Size | Depends on |
| --- | --- | --- | --- | --- |
| [C1-01](C1-01-clay.md) | What do we call the clay creature that moves, and the body Stand freezes? | question | S | none |
| [C1-02](C1-02-graft.md) | Is an own severed part that is put back a graft? | question | S | none |
| [C1-03](C1-03-down.md) | Which words for being out of the fight, for dropping, and for a wipe? | question | S | none |
| [C1-04](C1-04-actor.md) | How are the acting body, the body acted on, a fact's origin and a formula's subject named? | question | M | none |
| [C1-05](C1-05-formula.md) | How are the variables in formula environments named? | question | S | none |
| [C1-06](C1-06-form.md) | What are "form", "action", "move" and "type"? | question | S | none |
| [C1-07](C1-07-still.md) | What is the controller for a creature that keeps still called? | question | S | none |
| [C1-08](C1-08-pack.md) | Is it a pack or an inventory? | question | S | none |
| [C1-09](C1-09-turn.md) | What do "turn", "per-turn" and "pending" mean? | question | S | none |
| [C1-10](C1-10-faction.md) | What is the hostile faction called? | question | S | none |
| [C1-11](C1-11-text.md) | Where does player-facing text live, and how are its IDs shaped? | question | S | none |
| [C1-12](C1-12-frontend.md) | Where does frontend chrome end and game text begin, and which core packages may the tui import? | question | S | none |
| [C1-13](C1-13-cache.md) | May tests share one cached sample registry? | question | S | none |
| [C1-14](C1-14-yaml.md) | May `go.yaml.in/yaml/v3` become a direct dependency? | question | S | none |
| [C1-15](C1-15-gitcmd.md) | Where does shared git code for the tools go, given agentkit is protected? | question | S | none |
| [C1-16](C1-16-speed.md) | What should a speed formula that fails at run time do? | question | S | none |
| [C1-17](C1-17-apply-the-decided-renames.md) | Apply the decided renames | debt | L | C1-01, C1-02, C1-03, C1-04, C1-05, C1-06, C1-07, C1-08, C1-09, C1-10, C1-11 |
| [C1-18](C1-18-shared-helpers-in-the-lowest-layer-and-one-test.md) | Shared helpers in the lowest layer, and one test fixture | debt | L | C1-17, C1-13 |
| [C1-19](C1-19-content-area-duplication-tests-comments.md) | Content area: duplication, tests, comments | debt | L | C1-18 |
| [C1-20](C1-20-actions-area-duplication-tests-comments.md) | Actions area: duplication, tests, comments | debt | L | C1-18 |
| [C1-21](C1-21-features-area-duplication-tests-comments.md) | Features area: duplication, tests, comments | debt | L | C1-18 |
| [C1-22](C1-22-foundation-area-duplication-tests-comments-speed.md) | Foundation area: duplication, tests, comments, speed errors | debt | L | C1-18, C1-16 |
| [C1-23](C1-23-game-and-frontend-duplication-text-out-of-the-tu.md) | Game and frontend: duplication, text out of the tui, tests, comments | debt | L | C1-18, C1-12 |
| [C1-24](C1-24-tools-fix-the-windows-test-shared-git-code-tests.md) | Tools: fix the Windows test, shared git code, tests, comments | debt | L | C1-18, C1-15 |
| [C1-25](C1-25-agentkit-cleanup-owner-does-this-one.md) | agentkit cleanup (owner does this one) | debt | L | C1-24, C1-14 |
| [C1-26](C1-26-design-doc-fixes-and-the-content-guide.md) | Design-doc fixes and the content guide | content | L | C1-17, C1-19 |
| [C1-27](C1-27-lock-in-what-the-cleanup-decided.md) | Lock in what the cleanup decided | debt | S | C1-17, C1-18, C1-19, C1-20, C1-21, C1-22, C1-23, C1-24, C1-25, C1-26 |
