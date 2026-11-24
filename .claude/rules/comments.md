---
paths:
  - "**/*.go"
  - "**/*.cue"
---

# Comments

Write the fewest words that tell a reader something the code and its names don't. When in doubt, cut. No comment at all is often the right answer.

- **Doc comments: one line.** Add a second sentence only when it changes how a caller uses the thing: a precondition, an error worth handling, a unit, an ordering or determinism guarantee.
- **Package comments: one to three sentences.** What the package is for, and the one thing a newcomer must know.
- **Other comments: none,** unless they say why something non-obvious is so.
- **Never:**
  - restate the name or the signature;
  - say that errors are returned when something fails;
  - list parameters whose names already say it;
  - narrate what the next lines do;
  - mention history or task IDs (except in `TODO(<ID>)`).

| Instead of | Write |
| --- | --- |
| `// Load loads the registry. It returns an error if loading fails.` | `// Load reads and compiles the content under dir.` |
| `// NewWorld creates a new World with the given seed. The seed is used for the RNG streams. It returns the new world.` | `// NewWorld starts a world whose RNG streams all derive from seed.` |
| `// Registry is a registry that holds the content records.` | `// Registry holds compiled content and never changes after Load.` |
| `// Apply applies the command to the world and returns any error that occurs.` | `// Apply runs one command; an invalid command emits api.Rejected instead of failing.` |
| `// increment the count` above `count++` | nothing |
| `// TestParse tests Parse.` | nothing: the test name says it |
| `// Sort the bodies by UID.` above `slices.SortFunc(...)` | `// By UID, so map order can't decide the outcome.`, or nothing |

A sentinel error says which call returns it and when: `// ErrNotFound is returned by Get when no record has the ID.`

Lint wants a doc comment on every exported name. When the name already says it all, the plain Go line is enough, and better than padding: `// Material returns the material with the given ID.` Unexported names, fields and test functions need no comment, so a restating one is cut.

The after-edit hook flags the commonest filler in the lines you change, and CI warns about it on a pull request's changed lines.
