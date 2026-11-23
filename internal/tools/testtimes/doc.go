// Command testtimes reports how long each package's tests took, from "go test" output.
//
//	testtimes [--budget 30s] [--baseline FILE] [--save FILE] [--summary FILE] LOG
//
// It reads the "ok" and "FAIL" lines that go test prints, one per package, from LOG; cached results
// carry no time and are skipped. With --summary it appends a Markdown table of the slowest packages
// to FILE, which in CI is GitHub's step summary. With --baseline it compares each package with the
// times an earlier run saved with --save: CI saves master's times and compares pull requests with
// them.
//
// A package over the budget, or at least twice as slow as its baseline and at least five seconds
// slower, gets a GitHub warning annotation. Slow tests warn; they never fail the build, so testtimes
// exits 0 whenever it could read its input.
package main
