package main

import (
	"bufio"
	"cmp"
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// shownInSummary is how many of the slowest packages the summary table lists.
const shownInSummary = 10

// A package counts as having got slower when it takes at least slowdownFactor times its baseline
// and at least slowdownFloor longer, so tiny packages jittering between 0.1s and 0.3s don't warn.
const (
	slowdownFactor = 2.0
	slowdownFloor  = 5 * time.Second
)

// resultLine matches the line go test prints for each package that ran, such as
// "ok  	example.com/pkg	9.341s" or "FAIL	example.com/pkg	1.2s".
var resultLine = regexp.MustCompile(`^(?:ok|FAIL)\s+(\S+)\s+([0-9.]+)s\b`)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "testtimes:", err)
		os.Exit(1)
	}
}

// run parses the flags, reads the log and the baseline, and writes the report. Annotations go to
// stdout, where GitHub Actions picks them up.
func run(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("testtimes", flag.ContinueOnError)
	budget := flags.Duration("budget", 30*time.Second, "warn about any package whose tests take longer")
	baselinePath := flags.String("baseline", "", "times an earlier run saved, to compare with (may not exist)")
	savePath := flags.String("save", "", "file to save this run's times to")
	summaryPath := flags.String("summary", "", "file to append the Markdown summary to")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		return errors.New("want exactly one go test log")
	}
	logFile, err := os.Open(flags.Arg(0))
	if err != nil {
		return err
	}
	defer func() { _ = logFile.Close() }()
	current, err := parseLog(logFile)
	if err != nil {
		return err
	}
	baseline := times{}
	if *baselinePath != "" {
		baseline, err = loadTimes(*baselinePath)
		if err != nil {
			return err
		}
	}
	rows := compare(current, baseline, *budget)
	if _, err := io.WriteString(stdout, annotations(rows, *budget)); err != nil {
		return err
	}
	if *summaryPath != "" {
		if err := appendFile(*summaryPath, summary(rows, *budget, len(baseline) > 0, modulePath())); err != nil {
			return err
		}
	}
	if *savePath != "" {
		return os.WriteFile(*savePath, []byte(current.String()), 0o644)
	}
	return nil
}

// times maps an import path to how long its tests took.
type times map[string]time.Duration

// String writes the times one package per line, "path seconds", sorted by path: the format
// loadTimes reads.
func (all times) String() string {
	var out strings.Builder
	for _, pkg := range slices.Sorted(maps.Keys(all)) {
		out.WriteString(pkg + " " + strconv.FormatFloat(all[pkg].Seconds(), 'f', 3, 64) + "\n")
	}
	return out.String()
}

// parseLog reads go test's per-package result lines.
func parseLog(log io.Reader) (times, error) {
	found := times{}
	scanner := bufio.NewScanner(log)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		match := resultLine.FindStringSubmatch(scanner.Text())
		if match == nil {
			continue
		}
		seconds, err := strconv.ParseFloat(match[2], 64)
		if err != nil {
			continue
		}
		found[match[1]] = time.Duration(seconds * float64(time.Second))
	}
	return found, scanner.Err()
}

// loadTimes reads a file written by times.String. A file that doesn't exist yet is an empty
// baseline.
func loadTimes(path string) (times, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return times{}, nil
	}
	if err != nil {
		return nil, err
	}
	loaded := times{}
	for line := range strings.SplitSeq(string(data), "\n") {
		pkg, value, ok := strings.Cut(strings.TrimSpace(line), " ")
		if !ok {
			continue
		}
		seconds, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: bad time in %q: %w", path, line, err)
		}
		loaded[pkg] = time.Duration(seconds * float64(time.Second))
	}
	return loaded, nil
}

// row is one package's time, with its baseline and what's wrong with it, if anything.
type row struct {
	pkg  string
	took time.Duration
	// baseline is zero when there is no baseline for the package.
	baseline   time.Duration
	overBudget bool
	slower     bool
}

// compare pairs each package with its baseline and flags the slow ones, slowest first.
func compare(current, baseline times, budget time.Duration) []row {
	rows := make([]row, 0, len(current))
	for pkg, took := range current {
		before := baseline[pkg]
		rows = append(rows, row{
			pkg:        pkg,
			took:       took,
			baseline:   before,
			overBudget: took > budget,
			slower: before > 0 && float64(took) >= slowdownFactor*float64(before) &&
				took-before >= slowdownFloor,
		})
	}
	slices.SortFunc(rows, func(a, b row) int {
		return cmp.Or(cmp.Compare(b.took, a.took), strings.Compare(a.pkg, b.pkg))
	})
	return rows
}

// annotations returns a GitHub warning command for each slow package.
func annotations(rows []row, budget time.Duration) string {
	var out strings.Builder
	for _, slow := range rows {
		if slow.overBudget {
			out.WriteString("::warning title=Slow tests::" + slow.pkg + " took " + seconds(slow.took) +
				", over the " + seconds(budget) + " budget\n")
		}
		if slow.slower {
			out.WriteString("::warning title=Tests got slower::" + slow.pkg + " took " + seconds(slow.took) +
				", up from " + seconds(slow.baseline) + " on master\n")
		}
	}
	return out.String()
}

// summary is the Markdown table for GitHub's step summary. module, when known, is trimmed from the
// package paths to keep the table narrow.
func summary(rows []row, budget time.Duration, compared bool, module string) string {
	var out strings.Builder
	out.WriteString("### Test times\n\n")
	if len(rows) == 0 {
		out.WriteString("No package reported a test time (none ran, or every result was cached).\n")
		return out.String()
	}
	note := "Slowest packages in this run. Budget: " + seconds(budget) + " per package."
	if compared {
		note += " Compared with the last run on master."
	}
	out.WriteString(note + "\n\n")
	out.WriteString("| Package | Time | Master | Change | |\n| --- | ---: | ---: | ---: | --- |\n")
	for _, slow := range rows[:min(len(rows), shownInSummary)] {
		pkg := strings.TrimPrefix(slow.pkg, module+"/")
		master, change := "–", "–"
		if slow.baseline > 0 {
			master = seconds(slow.baseline)
			change = signed(slow.took - slow.baseline)
		}
		var flags []string
		if slow.overBudget {
			flags = append(flags, "⚠️ over budget")
		}
		if slow.slower {
			flags = append(flags, "⚠️ much slower")
		}
		out.WriteString("| " + pkg + " | " + seconds(slow.took) + " | " + master + " | " + change + " | " +
			strings.Join(flags, ", ") + " |\n")
	}
	return out.String()
}

// seconds formats a duration as seconds with one decimal, such as "9.3s".
func seconds(d time.Duration) string {
	return strconv.FormatFloat(d.Seconds(), 'f', 1, 64) + "s"
}

// signed formats a difference with its sign, such as "+4.0s" or "-0.3s".
func signed(d time.Duration) string {
	if d >= 0 {
		return "+" + seconds(d)
	}
	return "-" + seconds(-d)
}

// modulePath reads the module path from go.mod in the current directory, or returns "" if it can't.
func modulePath() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return ""
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		if path, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.TrimSpace(path)
		}
	}
	return ""
}

// appendFile appends text to the file at path, creating it if needed.
func appendFile(path, text string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.WriteString(text); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
