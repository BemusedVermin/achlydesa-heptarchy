// Command samebehavior checks that the working tree plays exactly as a base commit does, so a
// refactor can show it changed no behavior.
//
//	samebehavior [--base origin/master] [--seeds 3] [--steps 200] [--rename old=new]...
//
// It checks out the base in a temporary worktree and plays the same headless harness games there
// and in the working tree. It compares:
//
//   - the harness CSV, byte for byte;
//   - each seed's final game hash;
//   - every golden file under a testdata folder.
//
// A rename changes IDs, and so the hashes. With --rename (repeatable, applied in order), each old
// string in the base's CSV, golden files and golden file paths is replaced by new before comparing,
// and the hashes are skipped. A rename task can then show that its diff is the rename and nothing
// else.
//
// It exits 0 when everything matches. Otherwise it exits 1 and names the first difference in each
// thing that differs.
package main
