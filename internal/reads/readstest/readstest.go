// Package readstest lays down on disk what a pane read goes to: repositories,
// their worktrees, and the session files that say what the agent in a pane is
// about — Claude Code's transcripts and oh-my-pi's title slot.
package readstest

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Repo builds a repository on disk. Its trunk is always `main`, so passing that
// as the branch is how a pane on the trunk is written.
func Repo(t *testing.T, branch string) string {
	t.Helper()
	return RepoIn(t, t.TempDir(), branch)
}

// RepoIn builds one in a directory that may already exist, so a test can watch
// a repository appear under a pane that was read before it did.
func RepoIn(t *testing.T, root, branch string) string {
	t.Helper()

	gitDir := filepath.Join(root, ".git")
	remote := filepath.Join(gitDir, "refs", "remotes", "origin")

	mkdir(t, remote)
	write(t, filepath.Join(gitDir, "HEAD"), "ref: refs/heads/"+branch+"\n")
	write(t, filepath.Join(remote, "HEAD"), "ref: refs/remotes/origin/main\n")

	return root
}

// RepoWithNoTrunk builds a repository recording no default branch, which is
// what one that was never cloned looks like: every branch in it is worth naming.
func RepoWithNoTrunk(t *testing.T, branch string) string {
	t.Helper()

	root := Repo(t, branch)

	if err := os.Remove(
		filepath.Join(root, ".git", "refs", "remotes", "origin", "HEAD"),
	); err != nil {
		t.Fatal(err)
	}

	return root
}

// Worktree builds a worktree of the repository at root out of files, the way
// git records one: the branch in the worktree's own HEAD, and the refs it
// shares with the repository a commondir away.
func Worktree(t *testing.T, root, name, branch string) string {
	t.Helper()

	gitDir := filepath.Join(root, ".git", "worktrees", name)
	tree := filepath.Join(root, ".claude", "worktrees", name)

	mkdir(t, gitDir)
	mkdir(t, tree)
	write(t, filepath.Join(gitDir, "HEAD"), "ref: refs/heads/"+branch+"\n")
	write(t, filepath.Join(gitDir, "commondir"), "../..\n")
	write(t, filepath.Join(tree, ".git"), "gitdir: "+gitDir+"\n")

	return tree
}

// Transcript lays down one session's transcript under a Claude Code
// configuration home. Which project directory it lands in is the transcript
// reader's business, and its own tests cover that.
func Transcript(t *testing.T, home, sessionID string, lines ...string) {
	t.Helper()

	path := filepath.Join(home, "projects", "any-project", sessionID+".jsonl")

	mkdir(t, filepath.Dir(path))
	write(t, path, strings.Join(lines, "\n")+"\n")
}

// OMPSession lays down one oh-my-pi session file under dir and returns its
// path, which is how Herdr reports such a session: the file itself, not an id.
// The title is what the session calls itself, and a nameless session is "".
func OMPSession(t *testing.T, dir, title string) string {
	t.Helper()

	path := filepath.Join(
		dir,
		"2026-10-03T13-26-27-431Z_01a101f1-d167-75c3-b35c-132f7037bb6f.jsonl",
	)

	mkdir(t, dir)
	write(t, path, slot(t, title)+`{"type":"session","version":3,"cwd":"/work"}`+"\n")

	return path
}

// slot is a title slot as OMP serializes it: one JSON line of exactly 256 bytes
// with its newline, the pad field taking up the slack. That exactness is what
// lets OMP rewrite a title in place, and it is why a reader of it can never
// trust an offset into the file.
func slot(t *testing.T, title string) string {
	t.Helper()

	line := `{"type":"title","v":1,"title":` + strconv.Quote(title) +
		`,"source":"auto","updatedAt":"2026-10-03T13:26:58.541Z","pad":"`

	pad := 256 - len(line) - len(`"}`) - 1
	if pad < 0 {
		t.Fatalf("the title %q does not fit in a slot", title)
	}

	return line + strings.Repeat(" ", pad) + `"}` + "\n"
}

// AgentIn is a transcript line saying where the agent is working. Only the
// directory is read out of it, and a Windows path is full of escapes.
func AgentIn(dir string) string {
	return `{"type":"assistant","cwd":` + strconv.Quote(dir) + `}`
}

func mkdir(t *testing.T, dir string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
