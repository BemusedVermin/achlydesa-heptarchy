// Command affected runs "go test" on only the packages a change can affect.
//
//	affected [--base REF] [--list] [FILE...]
//
// Files named on the command line stand in for the git diff, to see what a change would run.
//
// It takes the files changed since the merge base with REF (git diff REF...HEAD) and tests:
//
//   - each package that holds a changed file, and every package that imports it, directly or not;
//   - for a change under content/, each package whose tests read ../content/data, and nothing that
//     merely imports them, since their Go code has not changed;
//   - everything, if the change touches go.mod, go.sum, go.work, Taskfile.yml, .golangci.yml or
//     .github, or if no REF is given.
//
// Files that belong to no package (docs, tasks) select nothing. --list prints the selected packages
// instead of testing them. Only CI uses this; "task check" on a developer machine tests everything.
package main
