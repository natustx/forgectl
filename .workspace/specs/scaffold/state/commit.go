package state

import "fmt"

// AddCommitToSpec appends a commit hash to a completed spec by ID.
func AddCommitToSpec(s *ScaffoldState, specID int, hash string) error {
	if hash == "" {
		return fmt.Errorf("commit hash cannot be empty")
	}

	for i := range s.Completed {
		if s.Completed[i].ID == specID {
			s.Completed[i].CommitHashes = append(s.Completed[i].CommitHashes, hash)
			return nil
		}
	}

	// Check if it's the active spec.
	if s.CurrentSpec != nil && s.CurrentSpec.ID == specID {
		return fmt.Errorf("spec %d is still active (state: %s). Commits are added to completed specs", specID, s.State)
	}

	return fmt.Errorf("no completed spec with ID %d", specID)
}
