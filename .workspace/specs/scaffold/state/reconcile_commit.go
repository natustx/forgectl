package state

import (
	"fmt"
	"os/exec"
	"strings"
)

// ReconcileCommit takes a commit hash, determines which spec files were
// touched, and adds the hash to each matching completed spec.
// Returns the list of spec names that were updated.
func ReconcileCommit(s *ScaffoldState, workDir string, hash string) ([]string, error) {
	if hash == "" {
		return nil, fmt.Errorf("commit hash cannot be empty")
	}

	// Get files changed in the commit.
	cmd := exec.Command("git", "show", "--name-only", "--format=", hash)
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git show failed for %s: %w", hash, err)
	}

	changedFiles := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			changedFiles[line] = true
		}
	}

	// Match against completed specs.
	var updated []string
	for i := range s.Completed {
		if changedFiles[s.Completed[i].File] {
			// Check for duplicate.
			alreadyHas := false
			for _, h := range s.Completed[i].CommitHashes {
				if h == hash {
					alreadyHas = true
					break
				}
			}
			if !alreadyHas {
				s.Completed[i].CommitHashes = append(s.Completed[i].CommitHashes, hash)
				updated = append(updated, s.Completed[i].Name)
			}
		}
	}

	// Also register to the reconcile state if it exists.
	reconcileUpdated := false
	if s.Reconcile != nil {
		alreadyHas := false
		for _, h := range s.Reconcile.CommitHashes {
			if h == hash {
				alreadyHas = true
				break
			}
		}
		if !alreadyHas {
			s.Reconcile.CommitHashes = append(s.Reconcile.CommitHashes, hash)
			reconcileUpdated = true
		}
	}

	if len(updated) == 0 && !reconcileUpdated {
		// Check if the commit touched spec files but they already had the hash.
		for i := range s.Completed {
			if changedFiles[s.Completed[i].File] {
				return nil, fmt.Errorf("commit %s already registered to all affected specs", hash)
			}
		}
		return nil, fmt.Errorf("commit %s did not touch any completed spec files", hash)
	}

	if reconcileUpdated {
		updated = append(updated, "(reconcile)")
	}

	return updated, nil
}
