package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// taskFile is what ready needs from a task file's front matter.
type taskFile struct {
	id        string
	title     string
	status    string
	dependsOn []string
	path      string // slash-separated, relative to the repository root
}

// parseFrontMatter reads the YAML front matter at the start of a task file: flat "key: value"
// lines between two "---" lines. Values may be quoted, may be flow lists like [M1-01, M1-02], and
// may end in a # comment. It reports false when the text has no front matter.
func parseFrontMatter(text string) (map[string]string, bool) {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return nil, false
	}
	body, _, found := strings.Cut(text[len("---\n"):], "\n---")
	if !found {
		return nil, false
	}
	fields := map[string]string{}
	for line := range strings.SplitSeq(body, "\n") {
		key, value, found := strings.Cut(line, ":")
		if !found || strings.HasPrefix(key, " ") || strings.HasPrefix(key, "#") {
			continue
		}
		fields[strings.TrimSpace(key)] = unquote(stripComment(strings.TrimSpace(value)))
	}
	return fields, true
}

// stripComment removes a trailing # comment that isn't inside quotes.
func stripComment(value string) string {
	var quote rune
	for index, char := range value {
		switch {
		case quote != 0 && char == quote:
			quote = 0
		case quote == 0 && (char == '"' || char == '\''):
			quote = char
		case quote == 0 && char == '#' && (index == 0 || value[index-1] == ' ' || value[index-1] == '\t'):
			return strings.TrimSpace(value[:index])
		}
	}
	return value
}

// unquote removes one pair of matching quotes around a value.
func unquote(value string) string {
	if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
		return value[1 : len(value)-1]
	}
	return value
}

// parseList reads a flow list such as [M1-01, M1-02]. An empty list, null or "none" has no items.
func parseList(value string) []string {
	value = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(value, "["), "]"))
	var items []string
	for item := range strings.SplitSeq(value, ",") {
		item = unquote(strings.TrimSpace(item))
		if item != "" && item != "null" && item != "none" {
			items = append(items, item)
		}
	}
	return items
}

// loadTasks reads every task file under root/tasks: the markdown files that start with front
// matter holding an id, except the template.
func loadTasks(root string) ([]taskFile, error) {
	var tasks []taskFile
	err := filepath.WalkDir(filepath.Join(root, "tasks"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".md") || entry.Name() == "TEMPLATE.md" {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fields, found := parseFrontMatter(string(data))
		if !found || fields["id"] == "" {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		tasks = append(tasks, taskFile{
			id:        fields["id"],
			title:     fields["title"],
			status:    fields["status"],
			dependsOn: parseList(fields["depends_on"]),
			path:      filepath.ToSlash(relative),
		})
		return nil
	})
	slices.SortFunc(tasks, func(first, second taskFile) int { return compareIDs(first.id, second.id) })
	return tasks, err
}

// indexRow matches a row of a table in tasks/INDEX.md, capturing the task ID from its link.
var indexRow = regexp.MustCompile(`^\|\s*\[([^\]]+)\]\(`)

// loadIndexStatuses reads the status column of every table in tasks/INDEX.md, by task ID.
func loadIndexStatuses(text string) map[string]string {
	statuses := map[string]string{}
	statusColumn := -1
	for line := range strings.SplitSeq(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if !strings.HasPrefix(line, "|") {
			statusColumn = -1
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		for index := range cells {
			cells[index] = strings.TrimSpace(cells[index])
		}
		if statusColumn < 0 {
			statusColumn = slices.Index(cells, "Status")
			continue
		}
		if match := indexRow.FindStringSubmatch(line); match != nil && statusColumn < len(cells) {
			statuses[match[1]] = cells[statusColumn]
		}
	}
	return statuses
}

// claimFor returns the ID of the task a task branch is named for, or "": a branch task/<ID> or
// task/<ID>-<slug>. When several IDs fit, the longest wins, so task/M2-13.F1-x is M2-13.F1.
func claimFor(branch string, ids []string) string {
	_, name, found := strings.Cut(branch, taskBranchPrefix)
	if !found {
		return ""
	}
	best := ""
	for _, id := range ids {
		if (name == id || strings.HasPrefix(name, id+"-")) && len(id) > len(best) {
			best = id
		}
	}
	return best
}

// readyReport lists the tasks an agent may start now, and why the others can't start yet.
// branches are the local and remote task branches; indexStatuses is INDEX.md's status column.
func readyReport(tasks []taskFile, branches []string, indexStatuses map[string]string) string {
	byID := map[string]taskFile{}
	ids := make([]string, 0, len(tasks))
	for _, task := range tasks {
		byID[task.id] = task
		ids = append(ids, task.id)
	}
	claims := map[string][]string{}
	for _, branch := range branches {
		if id := claimFor(branch, ids); id != "" {
			claims[id] = append(claims[id], branch)
		}
	}
	var ready, claimed, waiting, active, mismatched []string
	for _, task := range tasks {
		line := fmt.Sprintf("  %-10s %s  (%s)", task.id, task.title, task.path)
		if indexed, found := indexStatuses[task.id]; found && indexed != task.status {
			mismatched = append(mismatched, fmt.Sprintf("  %s: the task file says %s, INDEX.md says %s",
				task.id, task.status, indexed))
		}
		switch task.status {
		case "in-progress", "blocked":
			active = append(active, fmt.Sprintf("  %-10s %s  [%s]", task.id, task.title, task.status))
		case "todo":
			var missing []string
			for _, dependency := range task.dependsOn {
				if byID[dependency].status != "done" {
					missing = append(missing, dependency)
				}
			}
			switch {
			case len(claims[task.id]) > 0:
				claimed = append(claimed, line+"  on "+strings.Join(claims[task.id], ", "))
			case len(missing) > 0:
				waiting = append(waiting, fmt.Sprintf("  %-10s needs %s", task.id, strings.Join(missing, ", ")))
			default:
				ready = append(ready, line)
			}
		}
	}
	var report strings.Builder
	section := func(heading string, lines []string, whenEmpty string) {
		if len(lines) == 0 && whenEmpty == "" {
			return
		}
		report.WriteString(heading + "\n")
		if len(lines) == 0 {
			report.WriteString("  " + whenEmpty + "\n")
		}
		for _, line := range lines {
			report.WriteString(line + "\n")
		}
		report.WriteString("\n")
	}
	section("Ready: todo, every dependency done, and no task branch yet", ready, "none")
	section("Claimed: todo, but a task branch exists (another agent may be on it)", claimed, "")
	section("In progress or blocked", active, "")
	section("Waiting on dependencies", waiting, "")
	section("INDEX.md disagrees with the task file (the task file is the source of truth)", mismatched, "")
	return strings.TrimRight(report.String(), "\n") + "\n"
}

// runReady prints the ready report for the checkout dir is in.
func runReady(dir string) (string, error) {
	top, err := gitOutput(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	top = filepath.FromSlash(top)
	tasks, err := loadTasks(top)
	if err != nil {
		return "", err
	}
	index, err := os.ReadFile(filepath.Join(top, "tasks", "INDEX.md"))
	if err != nil {
		return "", err
	}
	refs, err := gitOutput(top, "for-each-ref", "--format=%(refname:short)", "refs/heads/task/", "refs/remotes/")
	if err != nil {
		return "", err
	}
	var branches []string
	for ref := range strings.SplitSeq(refs, "\n") {
		if strings.Contains(ref, taskBranchPrefix) {
			branches = append(branches, ref)
		}
	}
	return readyReport(tasks, branches, loadIndexStatuses(string(index))), nil
}

// compareIDs orders task IDs naturally, so M1-9 sorts before M1-10 and M2-13 before M2-13.F1.
func compareIDs(first, second string) int {
	for first != "" && second != "" {
		firstChunk, firstRest := leadingChunk(first)
		secondChunk, secondRest := leadingChunk(second)
		if order := compareChunks(firstChunk, secondChunk); order != 0 {
			return order
		}
		first, second = firstRest, secondRest
	}
	return len(first) - len(second)
}

// leadingChunk splits off the leading run of digits, or of non-digits.
func leadingChunk(text string) (string, string) {
	digits := isDigit(text[0])
	end := 1
	for end < len(text) && isDigit(text[end]) == digits {
		end++
	}
	return text[:end], text[end:]
}

// compareChunks compares two runs as numbers when both are digits, otherwise as text.
func compareChunks(first, second string) int {
	if isDigit(first[0]) && isDigit(second[0]) {
		first, second = strings.TrimLeft(first, "0"), strings.TrimLeft(second, "0")
		if len(first) != len(second) {
			return len(first) - len(second)
		}
	}
	return strings.Compare(first, second)
}

// isDigit reports whether a byte is an ASCII digit.
func isDigit(char byte) bool {
	return char >= '0' && char <= '9'
}
