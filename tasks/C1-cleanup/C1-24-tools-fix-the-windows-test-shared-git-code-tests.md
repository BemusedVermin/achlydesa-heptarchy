---
id: C1-24
title: "Tools: fix the Windows test, shared git code, tests, comments"
status: done
milestone: C1
kind: debt
depends_on: [C1-18, C1-15]
discovered_by: null
size: L
---

# C1-24 · Tools: fix the Windows test, shared git code, tests, comments

## Goal

The tools' tests pass on the owner's machine and share their git fixtures.

## Context

- Milestone brief: [README.md](README.md). Findings come from the audit that opened C1.
- `docs/glossary.md` for settled terms.
- This is one pull request by design: the work touches the same packages and is done in one pass.

## Deliverables

- The changes described in the Spec, in one branch and one pull request.

## Spec

Do the parts in this order, one commit each, because tests follow the code they test and comments come last once code and tests have settled: duplication and libraries, then tests, then comments.

### Fix samebehavior's TestRun on Windows

`go tool task check` passes on the owner's machine.

Findings:

- `samebehavior/main_test.go:33` `newRepo` has no `.gitattributes`. Git for Windows sets `core.autocrlf=true`, so the base worktree checks out `play.csv` with CRLF while the working copy has LF: `line 1 is "seed,step,map", was "seed,step,map\r"` in the cases at `:75` ("a refactor that changes no behavior", "a different step", "a rename, declared"). CI runs on Linux, so it passes there. The real tool is unaffected.

In `newRepo`, run `git config core.autocrlf false` after `init`, or put `* -text` in `.gitattributes` in `baseFiles`. The shared `gittest.Repo` in the tools duplication task carries the same guard.

### tools: one git test fixture, shared git code, one entry-point shape

The tools' tests build repos one way, and the tools share what they can.

Findings:

- Git test fixtures are copied four times: `agentkit/git_test.go:10-38`, `agentkit/worktrees_test.go:86-134`, `affected/main_test.go:17-35`, `samebehavior/main_test.go:33-73`; the `MkdirAll`+`WriteFile` loop also at `agentkit/afteredit_test.go:158-166`, `ready_test.go:151-159`. Add `internal/tools/internal/gittest` with `Repo(t, branch, files map[string]string) string`, setting the identity, `commit.gpgsign=false` and `core.autocrlf=false`, writing files with `os.CopyFS(dir, fstest.MapFS{...})`.
- `agentkit/git.go:23-32` `gitOutput` and `samebehavior/main.go:298-308` `git` are identical; `affected/main.go:77-80` shells out on its own and drops stderr. Per the decision, a shared `gitcmd.Output` for samebehavior and affected only (affected's errors gain git's stderr).
- Five tools, four entry-point shapes: affected (`run(args) error` writing to `os.Stdout`, `main.go:54,58,61`) and cuefix (`os.Exit` inside the walk loop, `main.go:44-47`) get `run(args []string, stdout io.Writer) error` and a two-line `main`, as `testtimes` has. `cuefix` and `affected` use bare `fmt.Fprintln(os.Stderr, ...)` at `affected/main.go:22`, `cuefix/main.go:45`; the others write `_, _ =`.
- Libraries: `affected/select.go:56-61` and `:88-92` → `slices.Sorted(maps.Keys(...))`; `testtimes/main_test.go:29-36,52` → `maps.Equal`; `:80-87` → `slices.Equal`.

**Not in this task:** anything under `internal/tools/agentkit/`, which agents cannot edit on a `task/` branch (see the agentkit task).

### Tools tests: fewer repos, test the pure parts directly

Tool tests build fewer git repos and cuefix has a test.

Findings:

- `samebehavior` `TestRun` builds a repo and worktree for each of 6 cases (about 2.7s warm, 12s cold, `main_test.go:75-158`): move "a different step" and "a different final hash" into table tests of `firstDifference` and `compareHashes`; keep 4 git-backed cases.
- `TestRunCleanWorktrees` and `TestCleanReportWhenEmpty` (`agentkit/worktrees_test.go:331,343`) each build a bare repo and clone: merge into one that runs `clean-worktrees --no-fetch` on a fresh clone. Keep `TestCleanWorktrees` (guards against deleting unsaved work).
- `cuefix` has no test: with `fix(text string) string` extracted, one table test covering a module path, an embedded field and a same-type field that isn't embedded.

Edits under `internal/tools/agentkit/` are owner-only.

### Comments: tools

Every comment `go tool task comments -- ./internal/tools/affected ./internal/tools/cuefix ./internal/tools/samebehavior ./internal/tools/testtimes` lists is cut or fixed (8 when the rule tightened), along with the findings below.

Findings:

- `cuefix/doc.go:1-3` and `main.go:15-18`: the package comment names only the module path rewrite, not the embedded-field rewrite; `affected/main.go:42` repeats `doc.go:5`; `agentkit/guard.go:131`, `shell.go:132`, `shell.go:86` restate the name; `affected/doc.go:5` sits in the middle of usage and the default.

The policy is `.claude/rules/comments.md`: doc comments say what a caller needs, and every sentence carries information; other comments say only why; no end-of-line comments, no comments in tests other than doc comments, no task IDs. CUE follows the same rules. The examples below set the standard; sweep the whole area.

## Acceptance criteria

- [x] `go tool task same-behavior` passes; the exceptions are the deliberate behavior changes named in the parts above, each listed in Notes with the changed output
- [x] The tests commit changed only `_test.go` files and `testdata/`
- [x] `go tool task check` passes (on the owner's machine too, which the Windows fix in the tools task needs)
- [x] `go tool task comments -- ./internal/tools/affected ./internal/tools/cuefix ./internal/tools/samebehavior ./internal/tools/testtimes` reports nothing
- [x] Each finding above is resolved, one per line in Notes, or marked dropped with a reason
- [x] `go tool task test -- ./internal/tools/samebehavior` passes on Windows and Linux.
- [x] Only `_test.go` files changed.
- [x] Only `_test.go` files and `testdata/` changed (cuefix's extraction is in the tools duplication task).

## Out of scope

- Other areas' findings: they have their own task in this milestone.
- Any new behavior beyond what the Spec names.

## Notes

- Triaged 2026-11-23: todo in C1, as the owner decided; it is part of the first cleanup milestone.

Findings, one per line:

- Windows TestRun: master's `newRepo` already set `core.autocrlf false`, so the test already passed on Windows; the guard now lives in `gittest.Repo` (`internal/tools/internal/gittest`), with the identity and `commit.gpgsign=false`, writing files with `os.CopyFS`.
- Git fixtures: `affected/main_test.go` and `samebehavior/main_test.go` use `gittest.Repo`. agentkit's (`git_test.go`, `worktrees_test.go`, `afteredit_test.go`, `ready_test.go`): dropped, owner-only; filed as C1-24.F1, since C1-25 doesn't list them.
- `gitOutput`/`git`: samebehavior and affected use `internal/tools/internal/gitcmd.Output`; affected's git errors now carry git's stderr. agentkit keeps `gitOutput`, as C1-15 decided.
- Entry points: affected has `run(args, stdout io.Writer) error` and a two-line `main`; cuefix has `run(args []string) error` without `stdout`, because it prints nothing and an unused writer would be dead weight. Both write `_, _ = fmt.Fprintln(os.Stderr, ...)`.
- Libraries: `affected/select.go` uses `slices.Sorted(maps.Keys(...))` twice; testtimes' tests use `maps.Equal` and `slices.Equal`.
- samebehavior TestRun: "a different step" and "a different final hash" moved to `TestFirstDifference` and `TestCompareHashes`; 4 git-backed cases remain.
- `TestRunCleanWorktrees`/`TestCleanReportWhenEmpty` merge: dropped, owner-only; in C1-24.F1.
- cuefix: `fix(text string) string` extracted, with `TestFix` covering the module path, an embedded field, a same-type field that isn't embedded and a top-level field.
- Comments: the 8 the checker listed are gone (the two in affected's test went with the fixture). cuefix's package comment names both rewrites; `affected/main.go:42` cut; `affected/doc.go`'s command-line-files sentence moved after the default; `pkg.ImportPath`'s restating doc cut. `agentkit/guard.go:131`, `shell.go:86`, `shell.go:132`: dropped, owner-only (C1-25's comments criterion covers them).

Behavior changes that `same-behavior` (which passes) can't show:

- affected's git errors read `git <args>: <err>: <stderr>` instead of `git <args>: <err>`.
- affected's `--list` and test output go through the writer `run` is given (`os.Stdout` from `main`); `dependents` now returns a sorted list, which `selectPackages` sorted anyway.

Other:

- New packages beyond the Spec's `Repo`: `gittest.Git` (runs git, failing the test) and `gittest.Write` (writes or, for empty content, removes files), which the tests needed for changes after the first commit.
- The last three criteria come from merged sub-tasks and contradict the Spec's code changes; I read them as applying to the tests commits, which touch only `_test.go` files. The code changes, cuefix's `fix` extraction included, are in the duplication commit.
- Commits: duplication, tests, comments, but two gofmt fix-ups of the samebehavior tests landed around the comments commit, and a review fix-up after it.
- `go tool task schema` regenerates the CUE schemas with no diff, so cuefix behaves as before.
- The "a seed missing" case of `compareHashes` was left out so as not to pin its message `seed 2 ends at , was 00bb`; filed as C1-24.F2.
- Follow-ups: C1-24.F1, C1-24.F2.