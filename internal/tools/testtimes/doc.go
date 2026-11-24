// Command testtimes reports how long each package's tests took, from "go test" output.
//
//	testtimes [--budget 30s] [--baseline FILE] [--save FILE] [--summary FILE] LOG
//
// --summary appends a Markdown table of the slowest packages to FILE; --baseline compares with the
// times an earlier run saved with --save. A package over budget, or twice its baseline and at least
// five seconds slower, gets a warning. It exits 0 whenever it can read LOG: slow tests never fail CI.
package main
