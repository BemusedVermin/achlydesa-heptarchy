---
paths:
  - "content/**"
  - "internal/content/**"
---

# Working on content

## Data (`content/data`, in CUE)

- Anything a designer would tune lives here: numbers, names, text, glyphs, formulas, and the rules that tie records together (AGENTS.md rule 4).
- Player-facing strings are entries in a text: block. A record's name sits beside the record and is text.<record id> (material.clay is named "text.material.clay"); the loader rejects any other name, except for parts, which may share one ("text.part.arm"). An engine message is text.<package>.<event> and goes in that package's file under content/data/text/ (text/combat.cue, text/actions.cue). content/data/text.cue holds only the schema. {{.name}} placeholders are filled from the arguments of an api.Text. Records refer to text by ID, never by a literal string. A missing ID renders as [missing text: <id>], so look for that in output.
- Conditions used in more than one place go in `rules.cue` as named formulas. Formulas use them as `rule.<name>`, and the compiler inlines them.
- After editing data, run `go tool task content` (`cue vet` against the schemas). A formula that names an unknown variable or rule only fails when content loads, so also run the loader's tests: `go tool task test -- ./internal/content`.

## Schemas and the loader (`internal/content`)

- The Go structs are the source of truth. Never edit a `*_go_gen.cue` file by hand. After changing a content struct, run `go tool task schema` and commit the regenerated files with your change.
- Compile every formula through `(*Registry).Compile`, the one place that inlines named rules. forbidigo rejects `expr.Compile` and `expr.Eval` everywhere else.
- A loaded registry is immutable. Hot reload builds a whole new registry and swaps it in between ticks, so never change a record after load.
- A field that holds player-facing text is a text ID (`api.ContentID`), commented `// a text ID`.
