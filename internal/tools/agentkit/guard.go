package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// branchReader tells the guard which branch a directory's checkout is on.
type branchReader interface {
	// branch returns the checked-out branch's short name, or "" when HEAD is detached or dir
	// isn't in a git checkout.
	branch(dir string) string
}

// verdict is the guard's answer for one command line.
type verdict struct {
	// refusal says why the line may not run; empty when it may.
	refusal string
	// pushes lists the directories the line runs git push in. Each push waits for the check.
	pushes []string
}

// guardState follows a command line from one simple command to the next: cd changes the
// directory later commands run in, and git switch -c changes the branch they commit to.
type guardState struct {
	dir      string
	branches map[string]string // directory to the branch a command earlier in the line switched to
	reader   branchReader
}

// branch returns the branch dir is on at this point of the line.
func (state *guardState) branch(dir string) string {
	if name, found := state.branches[filepath.Clean(dir)]; found {
		return name
	}
	return state.reader.branch(dir)
}

// evaluate decides whether a command line may run. cwd is the directory it starts in.
func evaluate(line string, dia dialect, cwd string, reader branchReader) verdict {
	state := &guardState{dir: cwd, branches: map[string]string{}, reader: reader}
	var result verdict
	state.walk(line, dia, &result)
	return result
}

// walk evaluates each simple command of line in turn, stopping at the first refusal.
func (state *guardState) walk(line string, dia dialect, result *verdict) {
	for _, cmd := range parseCommandLine(line, dia) {
		if result.refusal != "" {
			return
		}
		name, args := cmd.program()
		if script, nested, found := nestedScript(name, args); found {
			state.walk(script, nested, result)
			continue
		}
		switch name {
		case "cd", "chdir", "pushd", "set-location", "sl":
			if target := firstPositional(args); target != "" {
				state.dir = resolve(state.dir, target)
			}
		case "git":
			state.git(args, result)
		case "gh":
			result.refusal = checkGh(args)
		}
	}
}

// nestedScript finds the script a command hands to another shell, such as bash -c "git push" or
// pwsh -Command "git push", and the dialect to read it in.
func nestedScript(name string, args []string) (string, dialect, bool) {
	switch name {
	case "bash", "sh", "zsh", "dash", "ksh":
		for index, arg := range args {
			if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && strings.Contains(arg, "c") {
				if index+1 < len(args) {
					return args[index+1], bash, true
				}
			}
		}
	case "pwsh", "powershell":
		for index, arg := range args {
			lower := strings.ToLower(arg)
			if lower == "-c" || lower == "-command" || lower == "-commandwithargs" {
				return strings.Join(args[index+1:], " "), powershell, true
			}
		}
	case "cmd":
		for index, arg := range args {
			if strings.EqualFold(arg, "/c") || strings.EqualFold(arg, "/k") {
				return strings.Join(args[index+1:], " "), powershell, true
			}
		}
	}
	return "", bash, false
}

// firstPositional returns the first argument that isn't a flag, or "".
func firstPositional(args []string) string {
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			return arg
		}
	}
	return ""
}

// resolve returns target as seen from dir.
func resolve(dir, target string) string {
	if target == "~" || strings.HasPrefix(target, "~/") || strings.HasPrefix(target, `~\`) {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, target[1:])
		}
	}
	if filepath.IsAbs(target) {
		return filepath.Clean(target)
	}
	return filepath.Join(dir, target)
}

// isDefaultBranch reports whether name is the branch only the owner changes.
func isDefaultBranch(name string) bool {
	return strings.EqualFold(name, "master") || strings.EqualFold(name, "main")
}

// gitOptionsWithValue are git's global options that take the next word as their value.
var gitOptionsWithValue = []string{"-c", "--git-dir", "--work-tree", "--namespace", "--config-env"}

// git evaluates one git command.
func (state *guardState) git(args []string, result *verdict) {
	dir := state.dir
	index := 0
	for index < len(args) && strings.HasPrefix(args[index], "-") {
		switch {
		case args[index] == "-C" && index+1 < len(args):
			dir = resolve(dir, args[index+1])
			index++
		case slices.Contains(gitOptionsWithValue, args[index]):
			index++
		}
		index++
	}
	if index >= len(args) {
		return
	}
	sub, rest := args[index], args[index+1:]
	switch sub {
	case "push":
		result.refusal = checkPush(rest, state.branch(dir))
		if result.refusal == "" && !hasFlag(rest, "-n", "--dry-run") {
			result.pushes = append(result.pushes, dir)
		}
	case "commit", "merge", "rebase", "cherry-pick", "revert", "am":
		if branch := state.branch(dir); isDefaultBranch(branch) {
			result.refusal = "git " + sub + " on " + branch + ": only the owner changes " + branch +
				", by merging pull requests. Make a task branch first: " +
				"git switch -c task/<ID>-<slug> origin/master. If you meant to switch branches earlier in " +
				"this command, run the switch as its own command first."
		}
	case "switch", "checkout":
		result.refusal = checkSwitch(sub, rest)
		if name := switchedTo(sub, rest); name != "" {
			state.branches[filepath.Clean(dir)] = name
		}
	default:
		result.refusal = checkGitOther(sub, rest)
	}
}

// hasFlag reports whether args hold any of the flags.
func hasFlag(args []string, flags ...string) bool {
	return slices.ContainsFunc(args, func(arg string) bool { return slices.Contains(flags, arg) })
}

// hasShortFlag reports whether args hold the one-letter flag, alone or bundled like -fd.
func hasShortFlag(args []string, letter rune) bool {
	return slices.ContainsFunc(args, func(arg string) bool {
		return len(arg) > 1 && arg[0] == '-' && arg[1] != '-' && strings.ContainsRune(arg[1:], letter)
	})
}

// pushOptionsWithValue are git push options that take the next word as their value.
var pushOptionsWithValue = []string{"-o", "--push-option", "--repo", "--receive-pack", "--exec"}

// checkPush refuses force pushes, deletions, pushes of every branch or tag, and pushes to the
// default branch. branch is the branch the push runs on.
func checkPush(args []string, branch string) string {
	const force = "force-pushing rewrites history others may have built on. " +
		"Never force-push: to bring a pushed branch up to date, git merge origin/master and push again."
	var positional []string
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch {
		case arg == "--":
			positional = append(positional, args[index+1:]...)
			index = len(args)
		case arg == "--force" || strings.HasPrefix(arg, "--force-with-lease") || arg == "--force-if-includes":
			return force
		case arg == "--delete":
			return deleteRemote
		case arg == "--all" || arg == "--mirror" || arg == "--tags" || arg == "--prune":
			return "git push " + arg + " pushes more than your branch. " +
				"Push only your task branch: git push -u origin HEAD"
		case slices.Contains(pushOptionsWithValue, arg):
			index++
		case strings.HasPrefix(arg, "--"):
		case strings.HasPrefix(arg, "-") && len(arg) > 1:
			if strings.ContainsRune(arg, 'f') {
				return force
			}
			if strings.ContainsRune(arg, 'd') {
				return deleteRemote
			}
		default:
			positional = append(positional, arg)
		}
	}
	if len(positional) <= 1 {
		if isDefaultBranch(branch) {
			return pushToDefault(branch)
		}
		return ""
	}
	for _, refspec := range positional[1:] {
		if strings.HasPrefix(refspec, "+") {
			return force
		}
		source, destination, mapped := strings.Cut(refspec, ":")
		if mapped && source == "" {
			return deleteRemote
		}
		target := source
		if mapped {
			target = destination
		}
		if target == "HEAD" || target == "@" {
			target = branch
		}
		target = strings.TrimPrefix(target, "refs/heads/")
		if isDefaultBranch(target) {
			return pushToDefault(target)
		}
	}
	return ""
}

// deleteRemote is the refusal for a push that deletes a remote branch.
const deleteRemote = "deleting remote branches is the owner's call. " +
	"Leave the branch; say in your report that it can go."

// pushToDefault is the refusal for a push to the default branch.
func pushToDefault(branch string) string {
	return "pushing to " + branch + " skips review: only the owner changes it, by merging pull requests. " +
		"Push your task branch instead: git push -u origin HEAD"
}

// checkSwitch refuses switches that throw away uncommitted work.
func checkSwitch(sub string, args []string) string {
	if hasFlag(args, "--force", "--discard-changes") || hasShortFlag(args, 'f') {
		return "git " + sub + " --force throws away uncommitted changes. Commit them first, or ask the owner."
	}
	if sub == "checkout" {
		const discardAll = " throws away every uncommitted change. Restore single files by name, or ask the owner."
		if dash := slices.Index(args, "--"); dash >= 0 && slices.Contains(args[dash+1:], ".") {
			return "git checkout -- ." + discardAll
		}
		if len(args) == 1 && args[0] == "." {
			return "git checkout ." + discardAll
		}
	}
	return ""
}

// switchedTo returns the branch a git switch or git checkout moves to, when it can tell.
func switchedTo(sub string, args []string) string {
	for index, arg := range args {
		if (arg == "-c" || arg == "-C" || arg == "--create" || arg == "-b" || arg == "-B") && index+1 < len(args) {
			return args[index+1]
		}
	}
	if sub == "switch" {
		return firstPositional(args)
	}
	return ""
}

// checkGitOther refuses the remaining git commands that throw away work or rewrite history.
func checkGitOther(sub string, args []string) string {
	const discard = " throws away uncommitted work, and nothing can bring it back. Ask the owner first."
	switch sub {
	case "reset":
		if hasFlag(args, "--hard") {
			return "git reset --hard" + discard
		}
	case "clean":
		if hasFlag(args, "--force") || hasShortFlag(args, 'f') {
			return "git clean -f deletes untracked files for good. Ask the owner first."
		}
	case "restore":
		if slices.Contains(args, ".") && !hasFlag(args, "--staged", "-S") {
			return "git restore ." + discard + " Restore single files by name."
		}
	case "stash":
		if len(args) > 0 && args[0] == "clear" {
			return "git stash clear deletes every stash for good. Ask the owner first."
		}
	case "branch":
		if hasFlag(args, "-D") || (hasFlag(args, "-d", "--delete") && hasFlag(args, "-f", "--force")) {
			return "git branch -D deletes a branch with unmerged commits. Leave it; say in your report that it can go."
		}
		if hasFlag(args, "-d", "--delete", "-m", "-M", "--move") && slices.ContainsFunc(args, isDefaultBranch) {
			return "the default branch is the owner's. Leave it alone."
		}
	case "worktree":
		if len(args) > 0 && args[0] == "remove" && (hasFlag(args, "--force") || hasShortFlag(args, 'f')) {
			return "git worktree remove --force" + discard
		}
	case "filter-branch", "filter-repo", "replace", "update-ref":
		return "git " + sub + " rewrites history or refs directly. Ask the owner first."
	}
	return ""
}

// ghFlagsWithValue are gh options, global or for gh api, that take the next word as their value.
var ghFlagsWithValue = []string{
	"-R", "--repo", "--hostname", "-H", "--header", "-q", "--jq", "-t", "--template", "--cache", "-p", "--preview",
}

// checkGh refuses gh commands that merge, delete or change settings: merging is the owner's step.
func checkGh(args []string) string {
	words := make([]string, 0, 2)
	for index := 0; index < len(args) && len(words) < 2; index++ {
		switch {
		case slices.Contains(ghFlagsWithValue, args[index]):
			index++
		case strings.HasPrefix(args[index], "-"):
		default:
			words = append(words, args[index])
		}
	}
	if len(words) == 0 {
		return ""
	}
	if words[0] == "api" {
		return checkGhAPI(args[slices.Index(args, "api")+1:])
	}
	if len(words) < 2 {
		return ""
	}
	refused := map[string][]string{
		"pr":       {"merge"},
		"repo":     {"delete", "archive", "rename", "edit"},
		"release":  {"create", "delete", "edit", "upload"},
		"secret":   {"set", "delete", "remove"},
		"variable": {"set", "delete", "remove"},
		"ruleset":  {"create", "delete", "edit"},
	}
	if slices.Contains(refused[words[0]], words[1]) {
		return "gh " + words[0] + " " + words[1] + " is the owner's step. Open the pull request and stop there."
	}
	return ""
}

// checkGhAPI refuses gh api calls that delete things, merge pull requests, move branch refs or
// change branch protection.
func checkGhAPI(args []string) string {
	method, endpoint := "", ""
	var fields []string
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch {
		case (arg == "-X" || arg == "--method") && index+1 < len(args):
			method = strings.ToUpper(args[index+1])
			index++
		case strings.HasPrefix(arg, "--method="):
			method = strings.ToUpper(strings.TrimPrefix(arg, "--method="))
		case strings.HasPrefix(arg, "-X") && len(arg) > 2:
			method = strings.ToUpper(arg[2:])
		case slices.Contains([]string{"-f", "-F", "--field", "--raw-field", "--input"}, arg) && index+1 < len(args):
			fields = append(fields, args[index+1])
			index++
		case slices.Contains(ghFlagsWithValue, arg):
			index++
		case strings.HasPrefix(arg, "-"):
		case endpoint == "":
			endpoint = strings.ToLower(arg)
		}
	}
	if method == "" {
		method = "GET"
		if len(fields) > 0 {
			method = "POST"
		}
	}
	refusal := "gh api " + method + " " + endpoint + " is the owner's step: agents open pull requests and stop there."
	if endpoint == "graphql" {
		query := strings.ToLower(strings.Join(fields, " "))
		mutations := []string{"mergepullrequest", "automerge", "deleteref", "deleterepository", "updateref"}
		for _, mutation := range mutations {
			if strings.Contains(query, mutation) {
				return refusal
			}
		}
		return ""
	}
	segments := strings.Split(strings.Trim(strings.SplitN(endpoint, "?", 2)[0], "/"), "/")
	switch {
	case method == "DELETE":
		return refusal
	case slices.Contains(segments, "merge") || slices.Contains(segments, "merges"):
		return refusal
	case method != "GET" && (strings.Contains(endpoint, "git/refs") || slices.Contains(segments, "protection")):
		return refusal
	}
	return ""
}
