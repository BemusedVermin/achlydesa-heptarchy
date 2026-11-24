---
paths:
  - "**/*_test.go"
  - "**/contenttest/**"
  - "**/gametest/**"
---

# Writing tests

- A helper that takes `testing.TB` names it `tb` and calls `tb.Helper()` first (thelper enforces this).
- Load the sample content with `contenttest.Sample(t)`, which is cached once per test binary. The registry is immutable and never written to. A test that edits content calls `contenttest.SampleBundle(t)` for its own deep copy, edits that, and builds a registry from it.
- Build games with `gametest.New`, `Start` and `Reload` rather than copying setup between packages.
- A helper parameter that every caller passes the same value for is dropped (unparam enforces this).
- A property test or a determinism check needs a few dozen steps, not thousands. Keep a test over a second only for a stated reason.
- Comments in tests follow `.claude/rules/comments.md`: none but doc comments.
