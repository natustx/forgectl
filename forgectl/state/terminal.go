package state

// IsTerminal reports whether the session has reached its terminal state
// with no work remaining in the current phase.
//
// Terminal states by phase:
//   - specifying:   PHASE_SHIFT with phase_shift.from == "specifying"
//   - planning:     PHASE_SHIFT with phase_shift.from == "planning"
//   - implementing: DONE
func IsTerminal(s *ForgeState) bool {
	if s.State == StatePhaseShift && s.PhaseShift != nil {
		return s.PhaseShift.From == PhaseSpecifying || s.PhaseShift.From == PhasePlanning
	}
	if s.Phase == PhaseImplementing && s.State == StateDone {
		return true
	}
	return false
}
