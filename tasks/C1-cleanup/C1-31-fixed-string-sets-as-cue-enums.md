---
id: C1-31
title: Fixed string sets as CUE enums
status: proposed
milestone: unplanned
kind: debt
depends_on: []
discovered_by: null
size: L
---

# C1-31 · Fixed string sets as CUE enums

## Goal

Every fixed set of strings in the content schemas is a named Go string type, and the generated CUE schema accepts only its values. A typo such as `target: "slef"` then fails at `cue vet` and at load, at the line that holds it. Today it passes the schema and is caught only by hand-written Go checks.

Raised by the owner, not by a task.

**Decided by the owner:** two checks keep it enforced. One is golangci-lint's `exhaustive` linter. The other is a test that every generated enum is enforced. This task may edit `.golangci.yml` to turn `exhaustive` on: an owner-approved exception to the AGENTS.md rule that task branches leave it alone.

## Context

- `docs/go-technical-design.md` (content and schemas) and `.claude/rules/content.md`: the Go structs are the source of truth, and `go tool task schema` generates the `*_go_gen.cue` files from them.
- What `cue get go` does now:
  - **Named constant types become disjunctions that aren't enforced.** `api.Dir` (`internal/core/api/space_go_gen.cue`) is generated as `#Dir: uint8 // #enumDir` plus `#enumDir: #DirE | #DirNE | ...`. The set is only named in a comment, so any `uint8` passes.
  - **Untyped constants give plain strings.** The target kinds (`internal/content/types.go`, `TargetNone = "none"` and the rest) become `target: string`.
- **Trial result.** Rewriting `#X: string // #enumX` to `#X: #enumX` makes CUE enforce the set. `target: "slef"` fails with an "empty disjunction" error that lists the allowed values, and the constants' own `#X & "value"` definitions don't cycle.
- **Load also enforces the schema.** `content.Load` (`internal/content/load.go`) builds the data package against the schemas and runs `Validate(cue.Concrete(true))`. So an enforced enum also fails at load and on hot reload.
- **Some Go checks must stay.** Bundles built or edited in Go never pass through CUE. Examples are `contenttest.SampleBundle` copies, and records sent straight to `Registry` checks in tests.
- `internal/tools/cuefix` already post-processes every `*_go_gen.cue` after `cue get go` (module path casing, embedded definitions).

Hand-written checks of a fixed set found so far:

| Field | Constants and list | Check |
| --- | --- | --- |
| `ActionDef.Target` | `Target*`, `targetKinds()` (`types.go`) | `checkActionShape` (`actions.go:115`) |
| `Trigger.Role` | `Role*`, `triggerRoles()` (`types.go`) | `triggers.go:90` |
| `Claim.Value` (empty allowed) | `Value*`, `valueKinds()` (`types.go`) | `checkKnowledge` (`load.go:159`) |

Audit for the rest: every `string` field in `internal/content` and in the `api` and `comp` types that content uses, whose doc says "one of", or that is matched against constants. Some are **not** fixed sets and stay plain strings:
- **Vocabulary lists.** `vocabulary.grants`, `vocabulary.tags` and similar are declared by content, so Go can't know them.
- **Open IDs.** Event kinds (`Trigger.On`), effect primitives (`EffectRef.Do`) and controller or chooser names are filled by registries at run time.

## Deliverables

- `internal/content/*.go` (and `api` or `comp` where a content field uses one): a named string type with typed constants for each fixed set. Each field takes the type, and its doc comment drops "one of the X constants".
- `internal/tools/cuefix`: rewrite `#X: <base> // #enumX` to `#X: #enumX` when `#enumX` is defined in the same file, plus a test.
- Regenerated `*_go_gen.cue` (`go tool task schema`).
- Remove the `xKinds()` lists and Go checks whose only job the schema now does. Keep any check a Go-built bundle relies on, with its error message unchanged, and say which in Notes.
- Tests: a `content` test that a bundle with an out-of-set value fails `Load` with an error naming the field.
- `docs/content-guide.md`: one line on how to add a value to a fixed set.
- `.golangci.yml` (owner-approved exception): enable `exhaustive`. Fix what it finds in code.
- A schema guard test, in `internal/tools/cuefix` or `internal/content`: it fails when any `*_go_gen.cue` defines an `#enumX` that its `#X` doesn't use.

## Spec

- **Shape (suggestion).** For example `type TargetKind string` with `TargetNone TargetKind = "none"`, and `ActionDef.Target TargetKind`.
- **Empty values.** Where the empty string is a valid "none", as in `Claim.Value`, either add an empty constant to the set or keep the field optional (`omitempty`). Choose whichever keeps the schema accepting today's content.
- **Same behaviour.** `go tool task same-behavior` shows no change.
- **Same content.** No content file changes, unless one already holds an out-of-set value. In that case it is a bug: say so in Notes.

- **`exhaustive` (suggestion).** Set `default-signifies-exhaustive: true`, so a `switch` with a `default` passes. Leave the map-literal check off. The linter also covers existing named types such as `api.Dir`, so expect findings there too. Fix them, and list them in Notes. If a finding needs a behaviour change, file it as a follow-up and don't fix it here.
- **Schema guard.** It reads the committed `*_go_gen.cue` files, so it also fails when someone regenerates without running cuefix.

## Acceptance criteria

- [ ] Every fixed string set found by the audit is a named type, and its generated schema is an enforced disjunction.
- [ ] `go tool task content` rejects a scratch content file with an out-of-set value (checked by hand, not committed), and accepts the real content.
- [ ] `content.Load` rejects an out-of-set value, covered by a test.
- [ ] `cuefix` has a test for the enum rewrite.
- [ ] `exhaustive` is on in `.golangci.yml`, and `go tool task lint` passes.
- [ ] The schema guard test fails on a generated file whose `#X` doesn't use its `#enumX` (shown by a test case), and passes on the committed schemas.
- [ ] `go tool task same-behavior` shows no change.
- [ ] `go tool task check` passes.

## Out of scope

- Content-declared vocabularies and run-time registries (event kinds, effect primitives, controllers, choosers).
- Numeric enums such as `api.Dir`. The same cuefix rewrite covers them, but content doesn't set them by hand today. Note any that content does set.

## Notes

Filled in by the implementer.
