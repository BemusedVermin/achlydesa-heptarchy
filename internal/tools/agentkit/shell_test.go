package main

import (
	"reflect"
	"testing"
)

func TestParseCommandLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		dia  dialect
		want []command
	}{
		{"one command", "git status", bash, []command{{"git", "status"}}},
		{"and list", "go tool task check && git push -u origin HEAD", bash,
			[]command{{"go", "tool", "task", "check"}, {"git", "push", "-u", "origin", "HEAD"}}},
		{"separators", "a; b | c || d & e", bash, []command{{"a"}, {"b"}, {"c"}, {"d"}, {"e"}}},
		{"new lines", "cd x\ngit push", bash, []command{{"cd", "x"}, {"git", "push"}}},
		{"double quotes keep separators", `git commit -m "a; b && c"`, bash,
			[]command{{"git", "commit", "-m", "a; b && c"}}},
		{"single quotes are literal", `echo 'a \" b'`, bash, []command{{"echo", `a \" b`}}},
		{"escaped quote in double quotes", `echo "say \"hi\""`, bash, []command{{"echo", `say "hi"`}}},
		{"quotes join a word", `git push origin "task/x":'task/x'`, bash,
			[]command{{"git", "push", "origin", "task/x:task/x"}}},
		{"subshell", "(cd x && git push)", bash, []command{{"cd", "x"}, {"git", "push"}}},
		{"substitution", "echo $(git push) `git status`", bash,
			[]command{{"echo", "$"}, {"git", "push"}, {"git", "status"}}},
		{"braces group", "{ git push; }", bash, []command{{"git", "push"}}},
		{"braces in words stay", "git show HEAD@{1} ${X}", bash, []command{{"git", "show", "HEAD@{1}", "${X}"}}},
		{"redirections dropped", "git push 2>&1 >out.txt < in", bash, []command{{"git", "push"}}},
		{"append redirection", "go test ./... >> log.txt", bash, []command{{"go", "test", "./..."}}},
		{"here-document skipped", "cat > body.md <<'EOF'\ngit push --force\nEOF\ngit status", bash,
			[]command{{"cat"}, {"git", "status"}}},
		{"indented here-document", "cat <<-EOF\n\tgit push -f\n\tEOF\necho done", bash,
			[]command{{"cat"}, {"echo", "done"}}},
		{"here-document in quotes is text", "gh pr create --body \"$(cat <<'EOF'\ngit push -f\nEOF\n)\"", bash,
			[]command{{"gh", "pr", "create", "--body", "$(cat <<'EOF'\ngit push -f\nEOF\n)"}}},
		{"line continuation", "git push \\\n  origin HEAD", bash, []command{{"git", "push", "origin", "HEAD"}}},
		{"powershell keeps backslashes", `cd C:\Users\me\repo; git push`, powershell,
			[]command{{"cd", `C:\Users\me\repo`}, {"git", "push"}}},
		{"powershell backtick escape", "git commit -m \"a`\"b\"", powershell,
			[]command{{"git", "commit", "-m", `a"b`}}},
		{"powershell call operator", `& "C:\Program Files\Git\cmd\git.exe" push`, powershell,
			[]command{{`C:\Program Files\Git\cmd\git.exe`, "push"}}},
		{"powershell script block", "1..2 | ForEach-Object { git push }", powershell,
			[]command{{"1..2"}, {"ForEach-Object"}, {"git", "push"}}},
		{"empty", "  ", bash, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := parseCommandLine(test.line, test.dia)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("parseCommandLine(%q) = %q, want %q", test.line, got, test.want)
			}
		})
	}
}

func TestProgram(t *testing.T) {
	tests := []struct {
		name     string
		cmd      command
		wantName string
		wantArgs []string
	}{
		{"plain", command{"git", "push"}, "git", []string{"push"}},
		{"assignment", command{"GIT_TRACE=1", "git", "push"}, "git", []string{"push"}},
		{"wrappers", command{"sudo", "-E", "env", "FOO=1", "git", "push"}, "git", []string{"push"}},
		{"keyword", command{"if", "git", "push"}, "git", []string{"push"}},
		{"windows path", command{`C:\Git\cmd\GIT.EXE`, "push"}, "git", []string{"push"}},
		{"unix path", command{"/usr/bin/gh", "pr", "merge"}, "gh", []string{"pr", "merge"}},
		{"only assignments", command{"A=1"}, "", nil},
		{"not an assignment", command{"=x", "y"}, "=x", []string{"y"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			name, args := test.cmd.program()
			if name != test.wantName || !reflect.DeepEqual(args, test.wantArgs) {
				t.Errorf("program() = %q %q, want %q %q", name, args, test.wantName, test.wantArgs)
			}
		})
	}
}
