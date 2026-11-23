---
id: C1-24
title: "Tools: fix the Windows test, shared git code, tests, comments"
status: proposed
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

About 7 comments are cut or fixed.

Findings:

- `cuefix/doc.go:1-3` and `main.go:15-18`: the package comment names only the module path rewrite, not the embedded-field rewrite; `affected/main.go:42` repeats `doc.go:5`; `agentkit/guard.go:131`, `shell.go:132`, `shell.go:86` restate the name; `affected/doc.go:5` sits in the middle of usage and the default.

The policy is the comments lens: doc comments say what a thing is for and what a caller must know, and every sentence carries information (cut name-restatements, filler, history, task IDs). Other comments are terse and say only why. CUE follows the same rules. The examples below set the standard; sweep the whole area.

## Acceptance criteria

- [ ] `go tool task same-behavior` passes; the exceptions are the deliberate behavior changes named in the parts above, each listed in Notes with the changed output
- [ ] The tests commit changed only `_test.go` files and `testdata/`
- [ ] `go tool task check` passes (on the owner's machine too, which the Windows fix in the tools task needs)
- [ ] Each finding above is resolved, one per line in Notes, or marked dropped with a reason
- [ ] `go tool task test -- ./internal/tools/samebehavior` passes on Windows and Linux.
- [ ] Only `_test.go` files changed.
- [ ] Only `_test.go` files and `testdata/` changed (cuefix's extraction is in the tools duplication task).

## Out of scope

- Other areas' findings: they have their own task in this milestone.
- Any new behavior beyond what the Spec names.

## Notes

