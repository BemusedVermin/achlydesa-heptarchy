# Lens: comments

The owner's policy:

- **Doc comments explain, concisely.** Doc comments are the package comment and the comments on exported identifiers and on struct fields. They say what a thing is for and anything a caller must know, but every sentence has to carry information. Cut:
  - restatements of the name ("Registry is a registry");
  - filler ("This function is used to...");
  - history ("previously", "now", "after M2-06");
  - task IDs, and anything the signature already says.
- **Every other comment is terse, and only says why.** Keep a comment only where the code can't say it: a non-obvious constraint, a determinism requirement, a workaround and its reason. Remove comments that say what the next lines do, section banners, commented-out code, and restated test case names.
- **No end-of-line comments, and no comments in tests** other than doc comments on test functions and helpers. The full rule is `.claude/rules/comments.md`.
- **CUE follows the same rules.** A comment in content explains design intent only where a value would otherwise puzzle a designer.

What to report, per package:

- What `go tool task comments -- <the area's packages>` reports, as a count; the cleanup task fixes all of it.
- How many other comments to cut or shorten, roughly, and the worst few as examples, each with `file:line` and its replacement (or "cut").
- Doc comments that are missing, or that mislead because the code changed under them.
- Don't list every comment: a cleanup task sweeps the package. Your examples set the standard it follows.
