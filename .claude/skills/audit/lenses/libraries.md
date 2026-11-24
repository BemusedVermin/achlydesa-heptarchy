# Lens: libraries

Find code that a library already does better: hand-written code, or code that predates a newer standard library feature.

- **Go standard library first,** at the Go version in `go.mod`. Typical finds:
  - `slices`, `maps`, `cmp` and `iter` instead of hand-written loops and sorts;
  - `strings.Cut` and `strings.SplitSeq`;
  - `min` and `max`, `errors.Join`, `sync.OnceValue`, `slog`;
  - `testing` helpers such as `t.Chdir` and `t.Context`.

  The modernize analyzer finds many of these at once: `go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest ./...` (it only reports, unless you pass `-fix`; don't).
- **Then the dependencies already in `go.mod`, and the libraries the tech design names** (AGENTS.md rule 8: reuse first). Look for code that reimplements what one of them offers, such as a hand-rolled query the ECS library has, or flag parsing that cobra does.
- **New third-party libraries last.** Propose one only when it removes a meaningful amount of code or risk. Give the lines it replaces, what it does, its maintenance status and license, and why the standard library isn't enough. Each new dependency is the owner's decision, so it becomes a question.

Report each finding with `file:line`, the replacement, and the lines it saves.

**The modernize analyzers now run in `golangci-lint`** (all but `omitzero`), so the lint run covers the standard-library finds above. Don't run it separately; spend the time on hand-written code a dependency already in `go.mod` replaces.
