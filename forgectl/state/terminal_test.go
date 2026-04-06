package state

import "testing"

func TestIsTerminal(t *testing.T) {
	tests := []struct {
		name string
		s    *ForgeState
		want bool
	}{
		{
			name: "specifying PHASE_SHIFT is terminal",
			s: &ForgeState{
				Phase: PhaseSpecifying,
				State: StatePhaseShift,
				PhaseShift: &PhaseShiftInfo{
					From: PhaseSpecifying,
					To:   PhasePlanning,
				},
			},
			want: true,
		},
		{
			name: "planning PHASE_SHIFT is terminal",
			s: &ForgeState{
				Phase: PhasePlanning,
				State: StatePhaseShift,
				PhaseShift: &PhaseShiftInfo{
					From: PhasePlanning,
					To:   PhaseImplementing,
				},
			},
			want: true,
		},
		{
			name: "implementing DONE is terminal",
			s: &ForgeState{
				Phase: PhaseImplementing,
				State: StateDone,
			},
			want: true,
		},
		{
			name: "specifying DRAFT is not terminal",
			s: &ForgeState{
				Phase: PhaseSpecifying,
				State: StateDraft,
			},
			want: false,
		},
		{
			name: "specifying DONE is not terminal — reconciliation pending",
			s: &ForgeState{
				Phase: PhaseSpecifying,
				State: StateDone,
			},
			want: false,
		},
		{
			name: "implementing IMPLEMENT is not terminal",
			s: &ForgeState{
				Phase: PhaseImplementing,
				State: StateImplement,
			},
			want: false,
		},
		{
			name: "PHASE_SHIFT with nil PhaseShift is not terminal",
			s: &ForgeState{
				Phase: PhaseSpecifying,
				State: StatePhaseShift,
			},
			want: false,
		},
		{
			name: "planning EVALUATE is not terminal",
			s: &ForgeState{
				Phase: PhasePlanning,
				State: StateEvaluate,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTerminal(tt.s)
			if got != tt.want {
				t.Errorf("IsTerminal() = %v, want %v (phase=%s, state=%s)", got, tt.want, tt.s.Phase, tt.s.State)
			}
		})
	}
}
