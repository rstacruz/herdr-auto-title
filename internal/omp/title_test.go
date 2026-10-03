package omp

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// session lays down a session file whose slot names it, and returns its path.
// The file continues past the slot as a real one does.
func session(t *testing.T, title string) string {
	t.Helper()

	path := filepath.Join(
		t.TempDir(),
		"2026-10-03T13-26-27-431Z_01a101f1-d167-75c3-b35c-132f7037bb6f.jsonl",
	)
	write(t, path, slotLine(t, title)+`{"type":"session","version":3,"cwd":"/work/dashboard"}`+"\n")

	return path
}

// slot is one title slot as OMP serializes it: a JSON line of exactly 256 bytes
// with its newline, the pad field taking up the slack. The size is what lets
// OMP rewrite a title in place, so a test that changes one must change nothing
// else about the file.
func slotLine(tb testing.TB, title string) string {
	tb.Helper()

	line := `{"type":"title","v":1,"title":` + strconv.Quote(title) +
		`,"source":"auto","updatedAt":"2026-10-03T13:26:58.541Z","pad":"`

	pad := slotBytes - len(line) - len(`"}`) - 1
	if pad < 0 {
		tb.Fatalf("the title %q does not fit in a slot", title)
	}

	return line + strings.Repeat(" ", pad) + `"}` + "\n"
}

func write(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestTitleReadsTheSlotAtTheHeadOfASession(t *testing.T) {
	t.Parallel()

	path := session(t, "Add OMP session naming support")

	if got := Title(path); got != "Add OMP session naming support" {
		t.Errorf("title = %q, want the title in the slot", got)
	}
}

func TestTitleIsEmptyWhenNothingNamesTheSession(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name    string
		content string
	}{
		{"a slot not yet named", slotLine(t, "")},
		{"a session older than the slot", `{"type":"session","version":3,"cwd":"/work"}` + "\n"},
		{"a first line of another type", `{"type":"model_change","id":"7ac10489"}` + "\n"},
		{"a first line that is not JSON", "not a session at all\n"},
		{"a slot with no newline", `{"type":"title","v":1,"title":"Cut off"}`},
		{"an empty file", ""},
	} {
		path := filepath.Join(t.TempDir(), "session.jsonl")
		write(t, path, test.content)

		if got := Title(path); got != "" {
			t.Errorf("%s: title = %q, want nothing", test.name, got)
		}
	}
}

func TestTitleIsEmptyWhenThereIsNoFileToRead(t *testing.T) {
	t.Parallel()

	if got := Title(filepath.Join(t.TempDir(), "session.jsonl")); got != "" {
		t.Errorf("title = %q, want nothing for a file that is not there", got)
	}
}

func TestTitleRefusesAPathThatCouldNotBeOne(t *testing.T) {
	t.Parallel()

	// The path arrives over the socket and becomes the file this opens.
	if got := Title("session.jsonl"); got != "" {
		t.Errorf("title = %q, want a relative path refused", got)
	}
}

func TestTitleIsReadAgainAfterTheSlotIsRewritten(t *testing.T) {
	t.Parallel()

	path := session(t, "First name")
	if got := Title(path); got != "First name" {
		t.Fatalf("title = %q with the first slot in place, so this test proves nothing", got)
	}

	write(t, path, slotLine(t, "Second name")+`{"type":"session","version":3}`+"\n")

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if got := Title(path); got != "Second name" {
		t.Errorf("title = %q, want the rewritten one at size %d", got, info.Size())
	}
}
