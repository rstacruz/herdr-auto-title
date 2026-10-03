// Package omp reads what an oh-my-pi session is about from the session file
// Herdr pointed at. OMP keeps the session's title in a fixed-width slot at the
// head of that file and rewrites it there, so nothing is remembered between
// reads: the answer is in the first bytes every time.
package omp

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
)

// Agent is the label Herdr gives oh-my-pi, and the only agent whose session
// file this package knows how to read.
const Agent = "omp"

// slotBytes is the size of the title slot. OMP serializes it to exactly 256
// UTF-8 bytes with the newline, so a title can be rewritten in place without
// the session file behind it being rewritten.
const slotBytes = 256

// titleSlot is the name a session has given itself, as OMP writes it in the
// slot at the head of the file. A session too old to carry one opens with its
// header instead, which has another type and yields no title.
type titleSlot struct {
	Type  string `json:"type"`
	Title string `json:"title"`
}

// Title reports the title the session gives itself, and "" when nothing can be
// read: no file there, no slot at its head, or a slot not named yet. The title
// is truncated already: OMP cuts it to fit the slot.
func Title(path string) string {
	// The path arrives over the socket and becomes the file this opens, so a
	// value that could not be one is refused rather than cleaned; only the slot
	// at its head is read, so anything else there answers nothing.
	if !filepath.IsAbs(path) {
		return ""
	}

	file, err := os.Open(path) //nolint:gosec // the path is Herdr's, as reported
	if err != nil {
		return ""
	}
	// A read-only file has nothing to report on close.
	defer func() { _ = file.Close() }()

	head := make([]byte, slotBytes)

	size, err := io.ReadFull(file, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return ""
	}

	line, _, complete := bytes.Cut(head[:size], []byte("\n"))
	if !complete {
		return ""
	}

	var read titleSlot
	if json.Unmarshal(line, &read) != nil || read.Type != "title" {
		return ""
	}

	return read.Title
}
