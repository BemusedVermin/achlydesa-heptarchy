package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// contentMarker is the relative path a test file holds when it loads the content data from disk.
const contentMarker = "../content/data"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "affected:", err)
		os.Exit(1)
	}
}

// run parses the flags, selects the packages and tests them (or lists them).
func run(args []string) error {
	flags := flag.NewFlagSet("affected", flag.ContinueOnError)
	base := flags.String("base", "", "git ref to compare with; empty tests everything")
	list := flags.Bool("list", false, "print the selected packages instead of testing them")
	if err := flags.Parse(args); err != nil {
		return err
	}
	pkgs, err := listPackages()
	if err != nil {
		return err
	}
	var selected []string
	switch {
	case flags.NArg() > 0:
		// Files named on the command line stand in for the git diff, to see what a change would run.
		selected = selectPackages(flags.Args(), pkgs)
	case *base == "":
		selected = importPaths(pkgs)
	default:
		changed, diffErr := changedFiles(*base)
		if diffErr != nil {
			return diffErr
		}
		selected = selectPackages(changed, pkgs)
	}
	if *list {
		fmt.Println(strings.Join(selected, "\n"))
		return nil
	}
	if len(selected) == 0 {
		fmt.Println("affected: no package is affected by this change, so no tests run")
		return nil
	}
	fmt.Printf("affected: testing %d of %d packages\n", len(selected), len(pkgs))
	test := exec.Command("go", append([]string{"test"}, selected...)...)
	test.Stdout, test.Stderr = os.Stdout, os.Stderr
	return test.Run()
}

// changedFiles lists the files changed since the merge base with base, slash-separated, together
// with uncommitted and untracked files, so a run before committing tests the change too. In CI the
// tree is clean, so only the first list counts.
func changedFiles(base string) ([]string, error) {
	var files []string
	for _, args := range [][]string{
		{"diff", "--name-only", base + "...HEAD"},
		{"diff", "--name-only", "HEAD"},
		{"ls-files", "--others", "--exclude-standard"},
	} {
		out, err := exec.Command("git", args...).Output()
		if err != nil {
			return nil, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
		}
		files = append(files, strings.Fields(string(out))...)
	}
	slices.Sort(files)
	return slices.Compact(files), nil
}

// listedPackage is the part of "go list -json" output used here.
type listedPackage struct {
	ImportPath   string
	Dir          string
	Imports      []string
	TestImports  []string
	XTestImports []string
	TestGoFiles  []string
	XTestGoFiles []string
}

// listPackages describes every package in the module.
func listPackages() ([]pkg, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	out, err := exec.Command("go", "list", "-json", "./...").Output()
	if err != nil {
		return nil, fmt.Errorf("go list: %w", err)
	}
	var pkgs []pkg
	decoder := json.NewDecoder(bytes.NewReader(out))
	for {
		var listed listedPackage
		if err := decoder.Decode(&listed); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("go list: %w", err)
		}
		rel, err := filepath.Rel(root, listed.Dir)
		if err != nil {
			return nil, err
		}
		imports := append(append(listed.Imports, listed.TestImports...), listed.XTestImports...)
		pkgs = append(pkgs, pkg{
			ImportPath:   listed.ImportPath,
			Dir:          filepath.ToSlash(rel),
			Imports:      imports,
			ReadsContent: readsContent(listed.Dir, append(listed.TestGoFiles, listed.XTestGoFiles...)),
		})
	}
	return pkgs, nil
}

// readsContent reports whether any of the test files mentions the content data folder.
func readsContent(dir string, files []string) bool {
	for _, name := range files {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err == nil && bytes.Contains(data, []byte(contentMarker)) {
			return true
		}
	}
	return false
}
