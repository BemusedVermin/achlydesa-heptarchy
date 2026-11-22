package main

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// fakeBranches answers branchReader from a map of directory to branch.
type fakeBranches map[string]string

func (branches fakeBranches) branch(dir string) string {
	return branches[filepath.Clean(dir)]
}

func TestEvaluate(t *testing.T) {
	root := filepath.FromSlash("/work/repo")
	other := filepath.FromSlash("/work/other")
	branches := fakeBranches{root: "task/M2-07-hostility", other: "master"}
	onMaster := fakeBranches{root: "master"}
	tests := []struct {
		name       string
		line       string
		dia        dialect
		branches   fakeBranches
		wantRefuse string // a fragment of the refusal; empty when the line may run
		wantPushes []string
	}{
		{"status", "git status", bash, branches, "", nil},
		{"push task branch", "git push -u origin HEAD", bash, branches, "", []string{root}},
		{"push explicit task branch", "git push origin task/M2-07-hostility", bash, branches, "", []string{root}},
		{"dry run isn't gated", "git push -n origin HEAD", bash, branches, "", nil},
		{"force", "git push --force", bash, branches, "force-push", nil},
		{"force short", "git push -f origin HEAD", bash, branches, "force-push", nil},
		{"force bundled", "git push -uf origin HEAD", bash, branches, "force-push", nil},
		{"force with lease", "git push --force-with-lease=x origin HEAD", bash, branches, "force-push", nil},
		{"force refspec", "git push origin +HEAD:task/x", bash, branches, "force-push", nil},
		{"delete flag", "git push origin --delete task/x", bash, branches, "deleting", nil},
		{"delete refspec", "git push origin :task/x", bash, branches, "deleting", nil},
		{"all branches", "git push --all", bash, branches, "--all", nil},
		{"tags", "git push --tags origin", bash, branches, "--tags", nil},
		{"to master by refspec", "git push origin HEAD:master", bash, branches, "pushing to master", nil},
		{"to main by ref", "git push origin refs/heads/main", bash, branches, "pushing to main", nil},
		{"master by name", "git push origin master", bash, branches, "pushing to master", nil},
		{"on master", "git push", bash, onMaster, "pushing to master", nil},
		{"HEAD on master", "git push origin HEAD", bash, onMaster, "pushing to master", nil},
		{"push option value isn't a refspec", "git push -o master origin HEAD", bash, branches, "", []string{root}},
		{"commit on task branch", "git commit -m 'M2-07: done'", bash, branches, "", nil},
		{"commit on master", "git commit -am x", bash, onMaster, "only the owner changes master", nil},
		{"rebase on master", "git rebase origin/master", bash, onMaster, "git rebase on master", nil},
		{"switch then commit", "git switch -c task/M2-07-x origin/master && git commit -m x", bash, onMaster, "", nil},
		{"switch to master then commit", "git switch master && git commit -m x", bash, branches,
			"git commit on master", nil},
		{"cd elsewhere", "cd ../other && git commit -m x", bash, branches, "git commit on master", nil},
		{"git -C elsewhere", "git -C ../other push", bash, branches, "pushing to master", nil},
		{"git -c option", "git -c user.name=x commit -m y", bash, onMaster, "git commit on master", nil},
		{"nested bash", `bash -c "git push --force"`, bash, branches, "force-push", nil},
		{"nested login shell", `bash -lc 'git push -f'`, bash, branches, "force-push", nil},
		{"nested pwsh", `pwsh -NoProfile -Command git push --force`, bash, branches, "force-push", nil},
		{"env prefix", "GIT_TRACE=1 git push --force", bash, branches, "force-push", nil},
		{"after other commands", "go tool task check && git push --force", bash, branches, "force-push", nil},
		{"reset hard", "git reset --hard origin/master", bash, branches, "reset --hard", nil},
		{"reset soft", "git reset --soft HEAD~1", bash, branches, "", nil},
		{"clean", "git clean -fdx", bash, branches, "git clean", nil},
		{"clean dry run", "git clean -n", bash, branches, "", nil},
		{"checkout dot", "git checkout -- .", bash, branches, "checkout -- .", nil},
		{"checkout file", "git checkout -- internal/x.go", bash, branches, "", nil},
		{"checkout force", "git checkout -f task/y", bash, branches, "--force", nil},
		{"switch discard", "git switch --discard-changes task/y", bash, branches, "--force", nil},
		{"restore dot", "git restore .", bash, branches, "restore", nil},
		{"restore staged dot", "git restore --staged .", bash, branches, "", nil},
		{"stash clear", "git stash clear", bash, branches, "stash clear", nil},
		{"stash pop", "git stash pop", bash, branches, "", nil},
		{"branch force delete", "git branch -D task/old", bash, branches, "branch -D", nil},
		{"branch delete master", "git branch -d master", bash, branches, "default branch", nil},
		{"branch list", "git branch --list 'task/*'", bash, branches, "", nil},
		{"worktree force remove", "git worktree remove -f .claude/worktrees/x", bash, branches, "worktree remove", nil},
		{"filter-branch", "git filter-branch --tree-filter x", bash, branches, "rewrites history", nil},
		{"gh pr create", "gh pr create --base master --title x --body-file b.md", bash, branches, "", nil},
		{"gh pr merge", "gh pr merge 45 --squash", bash, branches, "gh pr merge", nil},
		{"gh pr merge with repo flag", "gh -R a/b pr merge 45", bash, branches, "gh pr merge", nil},
		{"gh repo delete", "gh repo delete a/b --yes", bash, branches, "gh repo delete", nil},
		{"gh api create pull", "gh api repos/a/b/pulls -f title=x -f head=task/x -f base=master", bash, branches,
			"", nil},
		{"gh api read", "gh api repos/a/b/pulls/45", bash, branches, "", nil},
		{"gh api merge", "gh api -X PUT repos/a/b/pulls/45/merge", bash, branches, "PUT repos/a/b/pulls/45/merge", nil},
		{"gh api delete", "gh api --method DELETE repos/a/b/git/refs/heads/x", bash, branches, "DELETE", nil},
		{"gh api attached method", "gh api -XDELETE repos/a/b/issues/1/labels/x", bash, branches, "DELETE", nil},
		{"gh api ref update", "gh api repos/a/b/git/refs/heads/master -f sha=abc", bash, branches, "POST", nil},
		{"gh api graphql merge", `gh api graphql -f query='mutation { mergePullRequest(input: {}) { id } }'`,
			bash, branches, "graphql", nil},
		{"gh api graphql read", `gh api graphql -f query='{ viewer { login } }'`, bash, branches, "", nil},
		{"powershell force", "Set-Location ../other; git push --force", powershell, branches, "force-push", nil},
		{"powershell cd", "Set-Location ../other; git commit -m x", powershell, branches, "git commit on master", nil},
		{"push text in a commit message", `git commit -m "never git push --force"`, bash, branches, "", nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := evaluate(test.line, test.dia, root, test.branches)
			switch {
			case test.wantRefuse == "" && got.refusal != "":
				t.Errorf("evaluate(%q) refused: %s", test.line, got.refusal)
			case test.wantRefuse != "" && !strings.Contains(got.refusal, test.wantRefuse):
				t.Errorf("evaluate(%q) = %q, want a refusal containing %q", test.line, got.refusal, test.wantRefuse)
			case !reflect.DeepEqual(got.pushes, test.wantPushes):
				t.Errorf("evaluate(%q) pushes = %q, want %q", test.line, got.pushes, test.wantPushes)
			}
		})
	}
}
