package state

import "testing"

func TestAddCommitToSpec_Success(t *testing.T) {
	s := NewState(1, 1, false, twoSpecQueue())

	// Complete first spec.
	assertAdvance(t, s, adv("", ""), PhaseSelect)
	assertAdvance(t, s, adv("", ""), PhaseDraft)
	assertAdvance(t, s, adv("", ""), PhaseEvaluate)
	assertAdvance(t, s, adv("", "PASS"), PhaseAccept)
	assertAdvance(t, s, adv("", ""), PhaseOrient) // queue not empty

	if len(s.Completed) != 1 {
		t.Fatalf("completed: got %d, want 1", len(s.Completed))
	}

	// Add a commit.
	err := AddCommitToSpec(s, 1, "abc1234")
	if err != nil {
		t.Fatalf("add commit: %v", err)
	}
	if len(s.Completed[0].CommitHashes) != 1 {
		t.Fatalf("hashes: got %d, want 1", len(s.Completed[0].CommitHashes))
	}
	if s.Completed[0].CommitHashes[0] != "abc1234" {
		t.Errorf("hash: got %q, want abc1234", s.Completed[0].CommitHashes[0])
	}

	// Add another commit.
	err = AddCommitToSpec(s, 1, "def5678")
	if err != nil {
		t.Fatalf("add second commit: %v", err)
	}
	if len(s.Completed[0].CommitHashes) != 2 {
		t.Fatalf("hashes: got %d, want 2", len(s.Completed[0].CommitHashes))
	}
}

func TestAddCommitToSpec_NotFound(t *testing.T) {
	s := NewState(1, 1, false, twoSpecQueue())
	err := AddCommitToSpec(s, 99, "abc1234")
	if err == nil {
		t.Fatal("expected error for non-existent spec ID")
	}
}

func TestAddCommitToSpec_EmptyHash(t *testing.T) {
	s := NewState(1, 1, false, twoSpecQueue())
	err := AddCommitToSpec(s, 1, "")
	if err == nil {
		t.Fatal("expected error for empty hash")
	}
}

func TestAddCommitToSpec_ActiveSpec(t *testing.T) {
	s := NewState(1, 1, false, twoSpecQueue())
	assertAdvance(t, s, adv("", ""), PhaseSelect) // spec 1 is now active

	err := AddCommitToSpec(s, 1, "abc1234")
	if err == nil {
		t.Fatal("expected error for active spec")
	}
}
