---
id: C1-25
title: "agentkit cleanup (owner does this one)"
status: proposed
milestone: C1
kind: debt
depends_on: [C1-24, C1-14]
discovered_by: null
size: L
---

# C1-25 · agentkit cleanup (owner does this one)

## Goal

agentkit is smaller and simpler.

## Context

- Milestone brief: [README.md](README.md). Findings come from the audit that opened C1.
- `docs/glossary.md` for settled terms.
- This is one pull request by design: the work touches the same packages and is done in one pass.

## Deliverables

- The changes described in the Spec, in one branch and one pull request.

## Spec

**Owner-only.** `internal/tools/agentkit/` is protected on `task/` branches. The owner does this on a branch that isn't a `task/` branch, or lets an agent do it with an exception.

### agentkit: shared hook plumbing, simpler types, YAML library (owner does this one)

agentkit is smaller and simpler.

Findings:

- `agentkit/main.go:122-142` `hook` and `afteredit.go:79-97` `runAfterEdit` both decode a `hookInput`, call a function, wrap and encode: one `respond(stdin, stdout, stderr, answer func(hookInput) any) int`. `ready.go:201-213` and `worktrees.go:355-364` differ only in `whenEmpty`: one `writeSection`.
- `verdict` (`guard.go:17-21`) wraps a string and is threaded as `*verdict`: have `walk` return the refusal string. `goTools.format` returns an error its only caller drops (`afteredit.go:123`), with no context at `afteredit.go:46`.
- `ready.go:22-80` hand-rolls front-matter parsing (`parseFrontMatter`, `stripComment`, `unquote`, `parseList`, with tests at `ready_test.go:12-55`): per the decision, cut the front matter and `yaml.Unmarshal` it.
- `guard.go:405` `strings.SplitN(endpoint, "?", 2)[0]` → `strings.Cut`; `worktrees_test.go:282` `os.IsNotExist` → `errors.Is(err, fs.ErrNotExist)`; `shell.go:29` receiver `c command` → `cmd`, as every other receiver in the package is a full word.
- `protect.go:64-65` says agents don't change "the design docs", but `docs/` is deliberately unprotected: remove ", the design docs," from the refusal text.

**Owner-only.** `internal/tools/agentkit/` is protected on `task/` branches. The owner does this on a branch that isn't a `task/` branch, or an agent runs it with the owner's exception.

## Acceptance criteria

- [ ] `go tool task same-behavior` passes; the exceptions are the deliberate behavior changes named in the parts above, each listed in Notes with the changed output
- [ ] `go tool task check` passes (on the owner's machine too, which the Windows fix in the tools task needs)
- [ ] `go tool task comments -- ./internal/tools/agentkit` reports nothing (41 when the rule tightened)
- [ ] Each finding above is resolved, one per line in Notes, or marked dropped with a reason

## Out of scope

- Other areas' findings: they have their own task in this milestone.
- Any new behavior beyond what the Spec names.

## Notes


- Triaged 2026-11-23: left proposed, the owner chose not now.
