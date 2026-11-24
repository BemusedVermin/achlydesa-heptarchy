package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"maps"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

// maxOtherLines caps a comment that isn't a doc comment. Doc comments may explain at any length.
const maxOtherLines = 3

var (
	// errorTautology matches saying that a failure returns an error.
	errorTautology = regexp.MustCompile(`(?i)\b(returns?|reports?|gives)\b[^.]*\berrors?\b[^.]*\b(if|when|on)\b[^.]*` +
		`\b(fail(s|ed|ure)?|goes wrong|an error|errors? occurs?|unsuccessful)\b` +
		`|\b(returns?|propagates?) (any |the |all )?errors?( that)?( it)? ` +
		`(occur(s|red)?|happens?|arises?|encounters|encountered|gets|got)\b`)
	// taskID matches a task or follow-up ID.
	taskID = regexp.MustCompile(`\b[MC]\d+-\d+(\.F\d+)*\b`)
	// todo matches a TODO naming its follow-up, the one place a task ID belongs.
	todo = regexp.MustCompile(`TODO\([^)]*\)`)
	// selfReference matches a comment that opens by describing itself as code.
	selfReference = regexp.MustCompile(
		`(?i)^(this|the) (function|method|type|struct|interface|package|test|helper|constant|variable|field)\b`,
	)
)

// fillerWords carry no information in a doc comment's first sentence once the name is said.
var fillerWords = map[string]bool{
	"a": true, "an": true, "the": true, "of": true, "for": true, "to": true, "and": true, "or": true,
	"in": true, "on": true, "with": true, "from": true, "by": true, "given": true, "its": true, "it": true,
	"this": true, "that": true, "which": true, "is": true, "are": true, "be": true, "new": true,
	"return": true, "create": true, "make": true, "get": true, "set": true, "hold": true, "represent": true,
	"contain": true, "do": true, "test": true, "check": true, "verify": true, "list": true, "store": true,
	"keep": true, "report": true, "whether": true, "describe": true,
	"value": true, "instance": true, "object": true, "struct": true, "type": true, "function": true, "method": true,
}

// commentFinding is a comment, on lines first to last, that says too little or too much.
type commentFinding struct {
	first, last int
	why         string
}

// reviewComments flags the commonest filler in a Go file's comments. A file that doesn't parse,
// such as one mid-edit, yields nothing.
func reviewComments(name string, src []byte) []commentFinding {
	files := token.NewFileSet()
	file, err := parser.ParseFile(files, name, src, parser.ParseComments)
	if err != nil || ast.IsGenerated(file) {
		return nil
	}
	inTest := strings.HasSuffix(name, "_test.go")
	docs := map[*ast.CommentGroup]documented{}
	if file.Doc != nil {
		docs[file.Doc] = documented{}
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch decl := node.(type) {
		case *ast.FuncDecl:
			addDoc(docs, decl.Doc, []*ast.Ident{decl.Name}, !inTest, paramNames(decl.Type)...)
		case *ast.GenDecl:
			addDoc(docs, decl.Doc, specNames(decl.Specs), !inTest)
		case *ast.TypeSpec:
			addDoc(docs, decl.Doc, []*ast.Ident{decl.Name}, !inTest)
		case *ast.ValueSpec:
			addDoc(docs, decl.Doc, decl.Names, !inTest)
		case *ast.Field:
			addDoc(docs, decl.Doc, decl.Names, false)
			if decl.Doc != nil {
				subject := docs[decl.Doc]
				subject.field = true
				docs[decl.Doc] = subject
			}
		}
		return true
	})

	var found []commentFinding
	for _, group := range file.Comments {
		text := group.Text()
		if strings.TrimSpace(text) == "" {
			continue
		}
		start := files.Position(group.Pos())
		lineStart := strings.LastIndexByte(string(src[:start.Offset]), '\n') + 1
		place := commentPlace{
			isPackage: group == file.Doc,
			endOfLine: strings.TrimSpace(string(src[lineStart:start.Offset])) != "",
			inTest:    inTest,
		}
		place.subject, place.isDoc = docs[group]
		for _, why := range reviewGroup(text, place) {
			found = append(found, commentFinding{start.Line, files.Position(group.End()).Line, why})
		}
	}
	return found
}

// documented is what a doc comment documents.
type documented struct {
	// name is the one identifier the comment documents; empty for a group or the package.
	name string
	// signature is a function's parameter names, which the comment needn't repeat either.
	signature []string
	// required is set when lint requires the comment: exported, and not a field or in a test.
	required bool
	field    bool
}

// commentPlace is where a comment group sits.
type commentPlace struct {
	subject   documented
	isDoc     bool
	isPackage bool
	endOfLine bool
	inTest    bool
}

// reviewGroup says what is wrong with one comment group's text.
func reviewGroup(text string, place commentPlace) []string {
	if strings.HasPrefix(strings.TrimSpace(text), "TODO(") {
		return nil
	}
	isDoc, isPackage, subject := place.isDoc, place.isPackage, place.subject
	switch {
	case place.inTest && !isPackage && (!isDoc || subject.field):
		return []string{"comment in a test: say it in the test or subtest name or the failure message, or cut it"}
	case place.endOfLine:
		return []string{"end-of-line comment: cut it; a field that needs a note gets a doc comment above it"}
	}

	var found []string
	flag := func(why string) { found = append(found, why) }
	if id := taskID.FindString(todo.ReplaceAllString(text, "")); id != "" {
		flag("names task " + id + ": say what the code does, not which task did it")
	}
	if lines := proseLines(text); !isDoc && !isPackage && lines > maxOtherLines {
		flag(strconv.Itoa(lines) + "-line comment: keep only the why (at most " + strconv.Itoa(maxOtherLines) +
			" lines)")
	}
	if errorTautology.MatchString(text) {
		flag("says errors are returned on failure: cut it, or name the error a caller should handle")
	}
	if (isDoc || isPackage) && selfReference.MatchString(strings.TrimSpace(text)) {
		flag(`opens with "this function" or the like: start with the name, or cut it`)
	}
	if !isDoc || subject.name == "" {
		return found
	}
	sentence := firstSentence(text)
	switch {
	case definesAsItself(sentence, subject.name):
		flag("defines " + subject.name + " as itself: say what it holds or guarantees instead")
	case !subject.required && restatesName(sentence, subject):
		// An exported name needs its one line even when the name says it all; anything else doesn't.
		flag("the first sentence only restates " + subject.name + "'s name and signature: cut it")
	}
	return found
}

// addDoc records a doc comment and what it documents. required is false for a test file or a field.
func addDoc(
	docs map[*ast.CommentGroup]documented, doc *ast.CommentGroup, idents []*ast.Ident, required bool,
	signature ...string,
) {
	if doc == nil {
		return
	}
	subject := documented{signature: signature}
	if len(idents) == 1 {
		subject.name = idents[0].Name
		subject.required = required && idents[0].IsExported()
	}
	docs[doc] = subject
}

// paramNames returns a function type's parameter names.
func paramNames(function *ast.FuncType) []string {
	var names []string
	if function.Params == nil {
		return nil
	}
	for _, field := range function.Params.List {
		for _, name := range field.Names {
			names = append(names, name.Name)
		}
	}
	return names
}

// specNames returns the identifiers a declaration's specs declare.
func specNames(specs []ast.Spec) []*ast.Ident {
	var idents []*ast.Ident
	for _, spec := range specs {
		switch spec := spec.(type) {
		case *ast.TypeSpec:
			idents = append(idents, spec.Name)
		case *ast.ValueSpec:
			idents = append(idents, spec.Names...)
		}
	}
	return idents
}

// proseLines counts the non-blank lines of a comment that aren't an indented code block.
func proseLines(text string) int {
	count := 0
	for line := range strings.SplitSeq(text, "\n") {
		if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			count++
		}
	}
	return count
}

// firstSentence returns a comment's text up to its first full stop or blank line.
func firstSentence(text string) string {
	paragraph, _, _ := strings.Cut(text, "\n\n")
	paragraph = strings.ReplaceAll(paragraph, "\n", " ")
	if sentence, _, found := strings.Cut(paragraph, ". "); found {
		return sentence
	}
	return strings.TrimSuffix(strings.TrimSpace(paragraph), ".")
}

// restatesName reports whether a sentence says nothing beyond the words of its subject's name and
// signature.
func restatesName(sentence string, subject documented) bool {
	known := map[string]bool{}
	for _, name := range append([]string{subject.name}, subject.signature...) {
		known[strings.ToLower(name)] = true
		for _, word := range splitName(name) {
			known[word] = true
		}
	}
	words := strings.FieldsFunc(strings.ToLower(sentence), func(r rune) bool { return !unicode.IsLetter(r) })
	if len(words) < 2 {
		return false
	}
	for _, word := range words {
		if !inSet(known, word) && !inSet(fillerWords, word) {
			return false
		}
	}
	return true
}

// definesAsItself reports whether a sentence opens "Registry is a registry". "Day is the day number
// at tick now" is fine: the article "the" introduces what tells it apart.
func definesAsItself(sentence, name string) bool {
	rest, found := strings.CutPrefix(sentence, name+" is ")
	if !found {
		return false
	}
	words := strings.Fields(strings.ToLower(rest))
	if len(words) < 2 || words[0] != "a" && words[0] != "an" {
		return false
	}
	words = words[1:]
	nameWords := splitName(name)
	if len(words) < len(nameWords) {
		return false
	}
	for index, nameWord := range nameWords {
		if !inSet(map[string]bool{nameWord: true}, strings.Trim(words[index], ".,;:")) {
			return false
		}
	}
	return true
}

// inSet reports whether a word, or the word without a common ending, is in set: "creates",
// "created" and "creating" all match "create".
func inSet(set map[string]bool, word string) bool {
	if set[word] {
		return true
	}
	for _, ending := range []struct{ suffix, replacement string }{
		{"s", ""}, {"es", ""}, {"ed", ""}, {"d", ""}, {"ing", ""}, {"ing", "e"},
	} {
		if trimmed, ok := strings.CutSuffix(word, ending.suffix); ok && set[trimmed+ending.replacement] {
			return true
		}
	}
	return false
}

// splitName splits an identifier into its lower-case words: "NewUIDCounter" gives new, uid, counter.
func splitName(name string) []string {
	var words []string
	runes := []rune(name)
	start := 0
	for index := 1; index <= len(runes); index++ {
		atEnd := index == len(runes)
		boundary := !atEnd && unicode.IsUpper(runes[index]) &&
			(unicode.IsLower(runes[index-1]) || index+1 < len(runes) && unicode.IsLower(runes[index+1]))
		if atEnd || boundary || runes[index] == '_' {
			if word := strings.Trim(string(runes[start:index]), "_"); word != "" {
				words = append(words, strings.ToLower(word))
			}
			start = index
		}
	}
	return words
}

// lineRange is a span of lines in a file, counted from 1.
type lineRange struct{ first, last int }

// hunkHeader matches a hunk header of "git diff -U0", capturing where the new lines start and how
// many there are.
var hunkHeader = regexp.MustCompile(`^@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@`)

// changedLines runs "git diff -U0" from a commit to the working tree and returns the lines each file
// gained or changed, by path relative to root with forward slashes.
func changedLines(root, from string, pathspecs ...string) (map[string][]lineRange, error) {
	args := append([]string{
		"diff", "-U0", "--no-color", "--no-ext-diff", "--src-prefix=a/", "--dst-prefix=b/", from, "--",
	}, pathspecs...)
	out, err := gitOutput(root, args...)
	if err != nil {
		return nil, err
	}
	return parseHunks(out), nil
}

// parseHunks reads the lines each file gained or changed from "git diff -U0" output.
func parseHunks(diff string) map[string][]lineRange {
	changed := map[string][]lineRange{}
	file, previous := "", ""
	for line := range strings.SplitSeq(diff, "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case strings.HasPrefix(line, "+++ ") && strings.HasPrefix(previous, "--- "):
			file, _ = strings.CutPrefix(strings.TrimPrefix(line, "+++ "), "b/")
			if file == "/dev/null" {
				file = ""
			}
		case strings.HasPrefix(line, "@@ ") && file != "":
			match := hunkHeader.FindStringSubmatch(line)
			if match == nil {
				break
			}
			start, _ := strconv.Atoi(match[1])
			count := 1
			if match[2] != "" {
				count, _ = strconv.Atoi(match[2])
			}
			if count > 0 {
				changed[file] = append(changed[file], lineRange{start, start + count - 1})
			}
		}
		previous = line
	}
	return changed
}

// fileFinding is a comment finding in one file.
type fileFinding struct {
	file string
	commentFinding
}

// String formats a finding as file:line: why.
func (finding fileFinding) String() string {
	return finding.file + ":" + strconv.Itoa(finding.first) + ": " + finding.why
}

// commentsOnChangedLines reviews the comments of the changed Go files under root and keeps the
// findings that touch a changed line. Files under testdata/ are fixtures, and are skipped.
func commentsOnChangedLines(root string, changed map[string][]lineRange) []fileFinding {
	var found []fileFinding
	for _, file := range slices.Sorted(maps.Keys(changed)) {
		if !strings.HasSuffix(file, ".go") || strings.Contains("/"+file, "/testdata/") {
			continue
		}
		src, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
		if err != nil {
			continue
		}
		for _, finding := range reviewComments(file, src) {
			touched := slices.ContainsFunc(changed[file], func(lines lineRange) bool {
				return finding.first <= lines.last && lines.first <= finding.last
			})
			if touched {
				found = append(found, fileFinding{file, finding})
			}
		}
	}
	return found
}

// runComments is the comments command: it reviews every comment in the Go files under the given
// paths, "." by default, and exits 1 when it finds any to cut or rewrite.
func runComments(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		args = []string{"."}
	}
	files := map[string][]lineRange{}
	for _, arg := range args {
		err := filepath.WalkDir(strings.TrimSuffix(arg, "/..."), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			name := entry.Name()
			if entry.IsDir() && path != "." && (strings.HasPrefix(name, ".") || name == "testdata") {
				return filepath.SkipDir
			}
			if !entry.IsDir() && strings.HasSuffix(name, ".go") {
				files[filepath.ToSlash(path)] = []lineRange{{1, math.MaxInt}}
			}
			return nil
		})
		if err != nil {
			warn(stderr, "agentkit comments:", err)
			return exitFailed
		}
	}
	found := commentsOnChangedLines(".", files)
	var out strings.Builder
	for _, finding := range found {
		out.WriteString(finding.String() + "\n")
	}
	if len(found) > 0 {
		out.WriteString(strconv.Itoa(len(found)) + " comments to cut or rewrite, as .claude/rules/comments.md says\n")
	}
	if _, err := io.WriteString(stdout, out.String()); err != nil {
		warn(stderr, "agentkit comments:", err)
		return exitFailed
	}
	if len(found) > 0 {
		return exitFailed
	}
	return exitOK
}

// commentNote tells the agent about filler in the comments of file, relative to root, on the lines
// changed since the last commit: all of them, for a file git doesn't track yet.
func commentNote(root, file string) string {
	changed, err := changedLines(root, "HEAD", file)
	if err != nil {
		return ""
	}
	if len(changed) == 0 {
		if tracked, err := gitOutput(root, "ls-files", "--", file); err != nil || tracked != "" {
			return ""
		}
		changed = map[string][]lineRange{file: {{1, math.MaxInt}}}
	}
	var lines []string
	for _, finding := range commentsOnChangedLines(root, changed) {
		if len(lines) == maxFindings {
			lines = append(lines, "(more not shown)")
			break
		}
		lines = append(lines, finding.String())
	}
	if len(lines) == 0 {
		return ""
	}
	return "Comments to cut or rewrite, as .claude/rules/comments.md says:\n" + strings.Join(lines, "\n")
}
