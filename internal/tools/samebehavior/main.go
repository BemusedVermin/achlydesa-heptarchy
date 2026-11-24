package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/tools/internal/gitcmd"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, harnessPlayer{}); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "samebehavior:", err)
		os.Exit(1)
	}
}

// errDifferent reports that the working tree doesn't behave as the base does. The details are
// already printed.
var errDifferent = errors.New("behavior differs from the base")

// options are samebehavior's command-line flags.
type options struct {
	base    string
	seeds   int
	steps   int
	renames []rename
}

// rename is one --rename old=new.
type rename struct{ old, new string }

// parseOptions reads samebehavior's flags.
func parseOptions(args []string) (options, error) {
	opts := options{}
	flags := flag.NewFlagSet("samebehavior", flag.ContinueOnError)
	flags.StringVar(&opts.base, "base", "origin/master", "the commit to compare with")
	flags.IntVar(&opts.seeds, "seeds", 3, "harness seeds to play")
	flags.IntVar(&opts.steps, "steps", 200, "steps per seed")
	flags.Func("rename", "old=new: treat old in the base as new (repeatable)", func(value string) error {
		old, replacement, ok := strings.Cut(value, "=")
		if !ok || old == "" {
			return fmt.Errorf("want old=new, got %q", value)
		}
		opts.renames = append(opts.renames, rename{old, replacement})
		return nil
	})
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	if flags.NArg() > 0 {
		return options{}, fmt.Errorf("unexpected arguments %q", flags.Args())
	}
	return opts, nil
}

// applyRenames replaces each old string with its new one, in order.
func applyRenames(text string, renames []rename) string {
	for _, each := range renames {
		text = strings.ReplaceAll(text, each.old, each.new)
	}
	return text
}

// player plays the harness in one checkout.
type player interface {
	// play runs the harness in the checkout at root and returns its CSV and its log.
	play(root string, seeds, steps int) (csv, log []byte, err error)
}

// harnessPlayer runs ./cmd/harness with go run.
type harnessPlayer struct{}

// play implements player.
func (harnessPlayer) play(root string, seeds, steps int) ([]byte, []byte, error) {
	cmd := exec.Command("go", "run", "./cmd/harness", "run",
		"--seeds", strconv.Itoa(seeds), "--steps", strconv.Itoa(steps))
	cmd.Dir = root
	var csv, log bytes.Buffer
	cmd.Stdout = &csv
	cmd.Stderr = &log
	if err := cmd.Run(); err != nil {
		return nil, nil, fmt.Errorf("harness in %s: %w\n%s", root, err, log.String())
	}
	return csv.Bytes(), log.Bytes(), nil
}

// run plays both checkouts, compares them and prints what differs.
func run(args []string, stdout io.Writer, games player) error {
	opts, err := parseOptions(args)
	if err != nil {
		return err
	}
	root, err := gitcmd.Output(".", "rev-parse", "--show-toplevel")
	if err != nil {
		return err
	}
	root = filepath.FromSlash(root)
	baseRoot, cleanup, err := checkOut(root, opts.base)
	if err != nil {
		return err
	}
	defer cleanup()

	baseCSV, baseLog, err := games.play(baseRoot, opts.seeds, opts.steps)
	if err != nil {
		return err
	}
	headCSV, headLog, err := games.play(root, opts.seeds, opts.steps)
	if err != nil {
		return err
	}
	var problems []string
	if diff := firstDifference(applyRenames(string(baseCSV), opts.renames), string(headCSV)); diff != "" {
		problems = append(problems, "harness CSV: "+diff)
	}
	if len(opts.renames) == 0 {
		if diff := compareHashes(seedHashes(baseLog), seedHashes(headLog)); diff != "" {
			problems = append(problems, "final hashes: "+diff)
		}
	}
	goldenProblems, goldens, err := compareGoldens(root, opts.base, opts.renames)
	if err != nil {
		return err
	}
	problems = append(problems, goldenProblems...)

	rows := strings.Count(string(headCSV), "\n")
	if len(problems) == 0 {
		summary := fmt.Sprintf("Same behavior as %s: %d harness rows (%d seeds, %d steps each) and %d golden files.\n",
			opts.base, rows, opts.seeds, opts.steps, goldens)
		_, err := io.WriteString(stdout, summary)
		return err
	}
	report := "Behavior differs from " + opts.base + ":\n- " + strings.Join(problems, "\n- ") + "\n"
	if _, err := io.WriteString(stdout, report); err != nil {
		return err
	}
	return errDifferent
}

// checkOut makes a temporary worktree of base and returns its root and a function that removes it.
func checkOut(root, base string) (string, func(), error) {
	dir, err := os.MkdirTemp("", "samebehavior-")
	if err != nil {
		return "", nil, err
	}
	tree := filepath.Join(dir, "base")
	if _, err := gitcmd.Output(root, "worktree", "add", "--detach", "--quiet", tree, base); err != nil {
		_ = os.RemoveAll(dir)
		return "", nil, err
	}
	cleanup := func() {
		// Playing the harness writes nothing into the tree, so it is clean and removes normally.
		if _, err := gitcmd.Output(root, "worktree", "remove", tree); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "samebehavior: couldn't remove the base worktree:", err)
			return
		}
		_ = os.Remove(dir)
	}
	return tree, cleanup, nil
}

// firstDifference describes the first line where two texts differ, or returns "" when they match.
func firstDifference(want, got string) string {
	if want == got {
		return ""
	}
	wantLines, gotLines := strings.Split(want, "\n"), strings.Split(got, "\n")
	for index := range max(len(wantLines), len(gotLines)) {
		var wantLine, gotLine string
		if index < len(wantLines) {
			wantLine = wantLines[index]
		}
		if index < len(gotLines) {
			gotLine = gotLines[index]
		}
		if wantLine != gotLine {
			return fmt.Sprintf("line %d is %q, was %q", index+1, gotLine, wantLine)
		}
	}
	return "they differ only in line endings"
}

// hashLine matches the harness's "seed finished" log line.
var hashLine = regexp.MustCompile(`seed=(\d+) hash=([0-9a-f]+)`)

// seedHashes reads each seed's final hash from the harness's log.
func seedHashes(log []byte) map[string]string {
	hashes := map[string]string{}
	for _, match := range hashLine.FindAllStringSubmatch(string(log), -1) {
		hashes[match[1]] = match[2]
	}
	return hashes
}

// compareHashes describes the first seed whose final hash differs, or returns "" when all match.
func compareHashes(want, got map[string]string) string {
	for _, seed := range slices.Sorted(maps.Keys(want)) {
		hash, reported := got[seed]
		if !reported {
			return fmt.Sprintf("seed %s reported no hash, was %s", seed, want[seed])
		}
		if hash != want[seed] {
			return fmt.Sprintf("seed %s ends at %s, was %s", seed, got[seed], want[seed])
		}
	}
	if len(got) != len(want) {
		return fmt.Sprintf("%d seeds reported a hash, was %d", len(got), len(want))
	}
	return ""
}

// compareGoldens compares every golden file, under a testdata folder, in the working tree with the
// base's, after the renames. It returns one problem per file that differs, and how many it compared.
func compareGoldens(root, base string, renames []rename) ([]string, int, error) {
	baseFiles, err := gitcmd.Output(root, "ls-tree", "-r", "--name-only", base)
	if err != nil {
		return nil, 0, err
	}
	headFiles, err := gitcmd.Output(root, "ls-files", "--cached", "--others", "--exclude-standard")
	if err != nil {
		return nil, 0, err
	}
	baseByRenamed := map[string]string{}
	for _, path := range goldenPaths(baseFiles) {
		baseByRenamed[applyRenames(path, renames)] = path
	}
	var problems []string
	compared := 0
	for _, path := range goldenPaths(headFiles) {
		got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		// Deleted but still in the index: the loop after this one reports it if the base had it.
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, 0, err
		}
		basePath, ok := baseByRenamed[path]
		if !ok {
			problems = append(problems, path+": new golden file")
			continue
		}
		delete(baseByRenamed, path)
		compared++
		want, err := gitcmd.Output(root, "show", base+":"+basePath)
		if err != nil {
			return nil, 0, err
		}
		// git show gives the blob as committed; a Windows checkout may have CRLF line endings.
		current := strings.TrimRight(strings.ReplaceAll(string(got), "\r\n", "\n"), "\n")
		if diff := firstDifference(applyRenames(want, renames), current); diff != "" {
			problems = append(problems, path+": "+diff)
		}
	}
	for _, path := range slices.Sorted(maps.Keys(baseByRenamed)) {
		problems = append(problems, baseByRenamed[path]+": golden file removed")
	}
	return problems, compared, nil
}

// goldenPaths picks the paths under a testdata folder from a newline-separated list, sorted.
func goldenPaths(list string) []string {
	var found []string
	for path := range strings.SplitSeq(list, "\n") {
		if strings.HasPrefix(path, "testdata/") || strings.Contains(path, "/testdata/") {
			found = append(found, path)
		}
	}
	slices.Sort(found)
	return found
}
