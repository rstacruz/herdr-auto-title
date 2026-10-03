package herdr

import "testing"

// A session is read as the reference kind its agent reports: an id for most,
// and the file the session is written to for pi and oh-my-pi. Reading one as
// the other is a mistake this pair of accessors exists to make impossible.

func TestIDFor(t *testing.T) {
	t.Parallel()

	id := &AgentSessionInfo{
		Agent: "claude",
		Kind:  SessionRefID,
		Value: "8852bfe0-8b24-4a23-a35e-7521d04da061",
	}

	if got, ok := id.IDFor("claude"); !ok || got != id.Value {
		t.Errorf("id = %q, %v, want the reference's value", got, ok)
	}

	if _, ok := id.IDFor("omp"); ok {
		t.Error("a Claude Code session answered for oh-my-pi")
	}

	if _, ok := id.PathFor("claude"); ok {
		t.Error("a reference by id answered for a path")
	}

	var none *AgentSessionInfo
	if _, ok := none.IDFor("claude"); ok {
		t.Error("a pane holding no session answered for one")
	}
}

func TestPathFor(t *testing.T) {
	t.Parallel()

	path := &AgentSessionInfo{
		Agent: "omp",
		Kind:  SessionRefPath,
		Value: "/home/r/sessions/-Dev/2026-10-03T13-26-27-431Z_01a101f1.jsonl",
	}

	if got, ok := path.PathFor("omp"); !ok || got != path.Value {
		t.Errorf("path = %q, %v, want the reference's value", got, ok)
	}

	if _, ok := path.PathFor("claude"); ok {
		t.Error("an oh-my-pi session answered for Claude Code")
	}

	if _, ok := path.IDFor("omp"); ok {
		t.Error("a reference by path answered for an id")
	}

	var none *AgentSessionInfo
	if _, ok := none.PathFor("omp"); ok {
		t.Error("a pane holding no session answered for one")
	}
}
