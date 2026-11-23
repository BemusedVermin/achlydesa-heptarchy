# Lens: Go and CUE conventions

The repository's own rules in `AGENTS.md` and `docs/go-technical-design.md` come first. Where they differ from general convention, don't "fix" the code: report the conflict as a question for the owner.

**Go** (Effective Go, Go Code Review Comments, Google's Go style guide):

- Naming:
  - MixedCaps, and initialisms in one case (`ID`, `UID`, `HTTP`);
  - no stutter (`combat.CombatState`);
  - getters without `Get`;
  - receiver names short and the same across a type's methods;
  - package names short and lower case, with no `util` or `common` packages.
- Errors:
  - error strings in lower case, without final punctuation;
  - wrap with `%w`;
  - sentinel errors named `ErrX`;
  - errors returned, not logged and returned.
- Interfaces small and defined where they're used; return concrete types.
- Zero values that work, with no constructor needed where none adds anything.
- Tests:
  - table-driven;
  - `t.Helper()` in helpers;
  - `cmp`-style comparisons rather than hand-written loops;
  - no sleeping.
- File layout: the type, then its constructor, then its methods; `doc.go` holds the package comment.

**CUE** (the CUE docs and the schema in `content/`):

- Formatted as `cue fmt` would: check with `go tool cue fmt --diff ./content/...`, which shows the changes without writing them.
- Definitions (`#Name`) for every shape, constraints in the schema rather than repeated in data, and defaults (`*value | other`) instead of the same value written into every entry.
- Field names in one style across all content. Hidden fields (`_x`) for helpers that aren't data.

For each convention broken more than once, name the golangci-lint linter or rule that would catch it from now on (for example `revive`'s `receiver-naming`, `errname`, `gocritic`). The lock-in step uses these.

Report each finding with `file:line` (or a count and examples, when it's widespread) and the fix.
