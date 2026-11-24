---
paths:
  - "**/*.go"
  - "**/*.cue"
---

# Comments

Write the fewest words that tell a reader something the code and its names don't. No comment at all is often the right answer.

- **Doc comments say what a caller needs.** They may explain, but every sentence carries something the name and signature don't.
- **Other comments only say why** something non-obvious is so: a constraint, a determinism requirement, a workaround. Never what the next lines do.
- **No end-of-line comments.** A field that needs a note gets a doc comment above it, and only when its name and type don't already say it.
- **No comments in tests,** other than doc comments on test functions and helpers. Say it in the test or subtest name, or in the failure message.
- **Never:**
  - restate the name or the signature;
  - say that errors are returned when something fails;
  - narrate what the code does;
  - mention history or task IDs (except in `TODO(<ID>)`).

| Instead of | Write |
| --- | --- |
| `// Load loads the registry. It returns an error if loading fails.` | `// Load reads and compiles the content under dir.` |
| `// Registry is a registry that holds the content records.` | `// Registry holds compiled content and never changes after Load.` |
| `// Apply applies the command and returns any error that occurs.` | `// Apply runs one command; an invalid command emits api.Rejected instead of failing.` |
| `Claim ContentID // a claim ID` | `Claim ContentID` |
| `Lands int64 // ticks until it lands` | `// Lands is how many ticks until the action lands.` above `Lands int64`, or a name that says it |
| `// The arm frees first, at tick 4.` above an assertion | `t.Errorf("arm freed at tick %d, want 4", tick)` |
| `// increment the count` above `count++` | nothing |
| `// TestParse tests Parse.` | nothing |
| `// Sort the bodies by UID.` above `slices.SortFunc(...)` | `// By UID, so map order can't decide the outcome.`, or nothing |
| `// installAll calls each Install in order (M1-10 adds explore).` | `// installAll calls each package's Install in a fixed order.` |

A sentinel error says which call returns it and when: `// ErrNotFound is returned by Get when no record has the ID.`

Lint wants a doc comment on every exported name. When the name already says it all, the plain Go line is enough, and better than padding: `// Material returns the material with the given ID.`

`go tool task comments -- <paths>` lists every comment under the paths that breaks these rules. The after-edit hook shows the ones on lines you change, and CI warns about them on a pull request's changed lines.
