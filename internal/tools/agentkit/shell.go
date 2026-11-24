package main

import (
	"strings"
)

// dialect says how a command line quotes and escapes characters.
type dialect int

const (
	// bash escapes with a backslash.
	bash dialect = iota
	// powershell escapes with a backtick; a backslash is an ordinary character, as in C:\Users.
	powershell
)

// command is one simple command from a command line: its words, with quotes and escapes removed.
type command []string

// wrappers are words that run the command after them: shell keywords and programs such as env.
var wrappers = map[string]bool{
	"!": true, "builtin": true, "command": true, "do": true, "elif": true, "else": true, "env": true,
	"exec": true, "if": true, "nice": true, "nohup": true, "sudo": true, "then": true, "time": true,
	"until": true, "while": true, "xargs": true,
}

// program returns the program a command runs, as a lower-case base name without an .exe suffix,
// and its arguments. It skips environment assignments (FOO=bar) and wrappers (env, sudo, if...).
func (c command) program() (string, []string) {
	for index := 0; index < len(c); index++ {
		word := c[index]
		if isAssignment(word) {
			continue
		}
		name := baseName(word)
		if wrappers[name] {
			for index+1 < len(c) && strings.HasPrefix(c[index+1], "-") {
				index++
			}
			continue
		}
		return name, c[index+1:]
	}
	return "", nil
}

// isAssignment reports whether a word sets an environment variable, like GOFLAGS=-v.
func isAssignment(word string) bool {
	name, _, found := strings.Cut(word, "=")
	if !found || name == "" {
		return false
	}
	for index, char := range name {
		letter := char == '_' || (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
		if !letter && (index == 0 || char < '0' || char > '9') {
			return false
		}
	}
	return true
}

// baseName returns a program path's file name in lower case, without .exe, .cmd or .bat. It
// splits on both slashes, since a PowerShell line can name C:\Program Files\Git\cmd\git.exe.
func baseName(path string) string {
	if cut := strings.LastIndexAny(path, `/\`); cut >= 0 {
		path = path[cut+1:]
	}
	path = strings.ToLower(path)
	for _, suffix := range []string{".exe", ".cmd", ".bat"} {
		path = strings.TrimSuffix(path, suffix)
	}
	return path
}

// parseCommandLine splits a command line into its simple commands. It handles quotes, separators,
// subshells, substitutions, groups, redirections and here-documents: enough to find the git and gh
// commands an agent writes, not a full shell parser.
func parseCommandLine(line string, dia dialect) []command {
	scanner := lineScanner{runes: []rune(line), dia: dia}
	scanner.scan()
	return scanner.commands
}

// lineScanner holds parseCommandLine's state.
type lineScanner struct {
	runes    []rune
	dia      dialect
	commands []command
	words    command
	word     strings.Builder
	inWord   bool
	heredocs []string // delimiters of here-documents whose text starts at the next new line
}

// escape is the dialect's escape character.
func (scanner *lineScanner) escape() rune {
	if scanner.dia == powershell {
		return '`'
	}
	return '\\'
}

// endWord finishes the word being read, if any.
func (scanner *lineScanner) endWord() {
	if scanner.inWord {
		scanner.words = append(scanner.words, scanner.word.String())
		scanner.word.Reset()
		scanner.inWord = false
	}
}

// endCommand finishes the command being read, if it has any words.
func (scanner *lineScanner) endCommand() {
	scanner.endWord()
	if len(scanner.words) > 0 {
		scanner.commands = append(scanner.commands, scanner.words)
		scanner.words = nil
	}
}

// standalone reports whether the rune at index stands alone as a word: nothing is being read and
// a space, a separator or the end of the line follows it.
func (scanner *lineScanner) standalone(index int) bool {
	if scanner.inWord {
		return false
	}
	return index+1 >= len(scanner.runes) || strings.ContainsRune(" \t\r\n;", scanner.runes[index+1])
}

// scan reads the whole line.
func (scanner *lineScanner) scan() {
	runes := scanner.runes
	for index := 0; index < len(runes); index++ {
		char := runes[index]
		switch {
		case char == scanner.escape() && index+1 < len(runes):
			index++
			if runes[index] != '\n' {
				scanner.word.WriteRune(runes[index])
				scanner.inWord = true
			}
		case char == '\'' || char == '"':
			index = scanner.readQuoted(index, &scanner.word)
			scanner.inWord = true
		case char == ' ' || char == '\t' || char == '\r':
			scanner.endWord()
		case char == '\n':
			scanner.endCommand()
			index = scanner.skipHeredocs(index)
		case strings.ContainsRune(";|&()", char) || (char == '`' && scanner.dia == bash):
			scanner.endCommand()
		case (char == '{' || char == '}') && scanner.standalone(index):
			scanner.endCommand()
		case char == '>' || char == '<':
			index = scanner.skipRedirection(index)
		default:
			scanner.word.WriteRune(char)
			scanner.inWord = true
		}
	}
	scanner.endCommand()
}

// readQuoted reads the quoted text that starts at index into out, and returns the index of the
// closing quote. In double quotes the escape character still escapes; in single quotes nothing does.
func (scanner *lineScanner) readQuoted(index int, out *strings.Builder) int {
	runes := scanner.runes
	quote := runes[index]
	index++
	for ; index < len(runes) && runes[index] != quote; index++ {
		if quote == '"' && runes[index] == scanner.escape() && index+1 < len(runes) {
			next := runes[index+1]
			if scanner.dia == powershell || strings.ContainsRune("\"\\$`\n", next) {
				index++
			}
		}
		out.WriteRune(runes[index])
	}
	return index
}

// readWord reads the word that starts at the first non-blank rune at or after index, and returns
// it with the index of its last rune.
func (scanner *lineScanner) readWord(index int) (string, int) {
	runes := scanner.runes
	for index < len(runes) && (runes[index] == ' ' || runes[index] == '\t') {
		index++
	}
	var word strings.Builder
	for ; index < len(runes); index++ {
		char := runes[index]
		if char == '\'' || char == '"' {
			index = scanner.readQuoted(index, &word)
			continue
		}
		if strings.ContainsRune(" \t\r\n;|&()<>", char) {
			break
		}
		word.WriteRune(char)
	}
	return word.String(), index - 1
}

// skipRedirection drops the redirection that starts at index, such as 2>&1, >> log.txt or a
// here-document's <<EOF, and returns the index of its last rune. A file-descriptor number just
// before the operator belongs to it, so it is dropped too.
func (scanner *lineScanner) skipRedirection(index int) int {
	if scanner.inWord && strings.Trim(scanner.word.String(), "0123456789") == "" {
		scanner.word.Reset()
		scanner.inWord = false
	}
	runes := scanner.runes
	operator := string(runes[index])
	for index+1 < len(runes) && strings.ContainsRune("<>|", runes[index+1]) && len(operator) < 3 {
		index++
		operator += string(runes[index])
	}
	if operator == "<<" && scanner.dia == bash {
		if index+1 < len(runes) && runes[index+1] == '-' {
			index++
		}
		delimiter, end := scanner.readWord(index + 1)
		scanner.heredocs = append(scanner.heredocs, delimiter)
		return end
	}
	if index+1 < len(runes) && runes[index+1] == '&' {
		index++
		for index+1 < len(runes) && strings.ContainsRune("0123456789-", runes[index+1]) {
			index++
		}
		return index
	}
	_, end := scanner.readWord(index + 1)
	return end
}

// skipHeredocs skips the text of any here-documents that start after the new line at index, and
// returns the index of the new line that ends the last one.
func (scanner *lineScanner) skipHeredocs(index int) int {
	runes := scanner.runes
	for _, delimiter := range scanner.heredocs {
		for index < len(runes) {
			end := index + 1
			for end < len(runes) && runes[end] != '\n' {
				end++
			}
			text := strings.TrimLeft(strings.TrimRight(string(runes[index+1:end]), "\r"), "\t")
			index = end
			if text == delimiter {
				break
			}
		}
	}
	scanner.heredocs = nil
	return index
}
