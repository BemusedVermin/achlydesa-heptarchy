package main

import (
	"maps"
	"path"
	"slices"
	"strings"
)

// pkg is what the selection needs to know about one Go package.
type pkg struct {
	ImportPath string
	// Dir is the package's directory, relative to the repository root, with forward slashes.
	Dir string
	// Imports holds the import paths of everything its code and tests import.
	Imports []string
	// ReadsContent says whether its tests load the content data from disk.
	ReadsContent bool
}

// globalFiles are files whose change can affect every package.
var globalFiles = []string{"go.mod", "go.sum", "go.work", "go.work.sum", "Taskfile.yml", ".golangci.yml"}

// selectPackages returns the import paths to test, sorted, for the changed files (slash-separated,
// relative to the repository root).
func selectPackages(changed []string, pkgs []pkg) []string {
	if slices.ContainsFunc(changed, isGlobal) {
		return importPaths(pkgs)
	}
	byDir := make(map[string]pkg, len(pkgs))
	for _, candidate := range pkgs {
		byDir[candidate.Dir] = candidate
	}
	selected := map[string]bool{}
	var seeds []string
	for _, file := range changed {
		if strings.HasPrefix(file, "content/") {
			for _, candidate := range pkgs {
				if candidate.ReadsContent {
					selected[candidate.ImportPath] = true
				}
			}
			continue
		}
		// A file in testdata or another subfolder belongs to the nearest package above it.
		for dir := path.Dir(file); dir != "."; dir = path.Dir(dir) {
			if owner, ok := byDir[dir]; ok {
				seeds = append(seeds, owner.ImportPath)
				break
			}
		}
	}
	for _, importPath := range dependents(seeds, pkgs) {
		selected[importPath] = true
	}
	return slices.Sorted(maps.Keys(selected))
}

// isGlobal reports whether a change to file can affect every package.
func isGlobal(file string) bool {
	return slices.Contains(globalFiles, file) || strings.HasPrefix(file, ".github/")
}

// dependents returns the seeds and every package that imports one of them, directly or not.
func dependents(seeds []string, pkgs []pkg) []string {
	importedBy := map[string][]string{}
	for _, candidate := range pkgs {
		for _, imported := range candidate.Imports {
			importedBy[imported] = append(importedBy[imported], candidate.ImportPath)
		}
	}
	seen := map[string]bool{}
	queue := slices.Clone(seeds)
	for len(queue) > 0 {
		next := queue[0]
		queue = queue[1:]
		if seen[next] {
			continue
		}
		seen[next] = true
		queue = append(queue, importedBy[next]...)
	}
	return slices.Sorted(maps.Keys(seen))
}

// importPaths returns the import path of every package, sorted.
func importPaths(pkgs []pkg) []string {
	result := make([]string, len(pkgs))
	for index, candidate := range pkgs {
		result[index] = candidate.ImportPath
	}
	slices.Sort(result)
	return result
}
