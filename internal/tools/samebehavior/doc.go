// Command samebehavior checks that the working tree plays exactly as a base commit does, so a
// refactor can show it changed no behavior.
//
//	samebehavior [--base origin/master] [--seeds 3] [--steps 200] [--rename old=new]...
//
// It plays the same harness games in a worktree of the base and compares the CSV, each seed's final
// hash and every golden file under testdata. Each --rename replaces old with new in the base's output
// first and skips the hashes, so a rename task can show its diff is the rename and nothing else.
package main
