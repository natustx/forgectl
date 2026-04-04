package cmd

import (
	"fmt"

	"forgectl/state"

	"github.com/spf13/cobra"
)

var (
	advanceVerdict    string
	advanceEvalReport string
	advanceMessage    string
	advanceFile       string
	advanceFrom       string
	advanceGuided     bool
	advanceNoGuided   bool
)

var advanceCmd = &cobra.Command{
	Use:   "advance",
	Short: "Transition from current state to next",
	RunE:  runAdvance,
}

func init() {
	advanceCmd.Flags().StringVar(&advanceVerdict, "verdict", "", "PASS or FAIL")
	advanceCmd.Flags().StringVar(&advanceEvalReport, "eval-report", "", "Path to evaluation report")
	advanceCmd.Flags().StringVar(&advanceMessage, "message", "", "Commit message or acceptance message")
	advanceCmd.Flags().StringVar(&advanceFile, "file", "", "Override file path (specifying DRAFT)")
	advanceCmd.Flags().StringVar(&advanceFrom, "from", "", "Path to queue input file (phase shift)")
	advanceCmd.Flags().BoolVar(&advanceGuided, "guided", false, "Enable guided mode")
	advanceCmd.Flags().BoolVar(&advanceNoGuided, "no-guided", false, "Disable guided mode")
	rootCmd.AddCommand(advanceCmd)
}

func runAdvance(cmd *cobra.Command, args []string) error {
	s, err := state.Load(stateDir)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()

	// Validate context-dependent flag constraints and print warnings.
	if err2 := validateAdvanceFlags(s); err2 != nil {
		return err2
	}
	printAdvanceWarnings(out, s)

	// Build input.
	var guided *bool
	if cmd.Flags().Changed("guided") {
		g := true
		guided = &g
	}
	if cmd.Flags().Changed("no-guided") {
		g := false
		guided = &g
	}

	in := state.AdvanceInput{
		Verdict:    advanceVerdict,
		EvalReport: advanceEvalReport,
		Message:    advanceMessage,
		File:       advanceFile,
		From:       advanceFrom,
		Guided:     guided,
	}

	err = state.Advance(s, in, stateDir)
	if err != nil {
		// Check if it's a validation error — still save state if VALIDATE was entered.
		if ve, ok := err.(*state.ValidationError); ok {
			if err2 := state.Save(stateDir, s); err2 != nil {
				return fmt.Errorf("saving state: %w", err2)
			}
			fmt.Fprintln(out)
			state.PrintAdvanceOutput(out, s, stateDir)
			fmt.Fprintln(out)
			fmt.Fprintf(out, "FAIL: %d errors in plan.json\n\n", len(ve.Errors))
			for _, e := range ve.Errors {
				fmt.Fprintf(out, "  %s\n", e)
			}
			return fmt.Errorf("validation failed")
		}
		return err
	}

	if err := state.Save(stateDir, s); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	state.PrintAdvanceOutput(out, s, stateDir)

	return nil
}

func validateAdvanceFlags(s *state.ForgeState) error {
	// --file only valid in specifying DRAFT.
	if advanceFile != "" && !(s.Phase == state.PhaseSpecifying && s.State == state.StateDraft) {
		return fmt.Errorf("--file is only valid in specifying DRAFT state (current: %s %s)", s.Phase, s.State)
	}

	// --verdict only valid in EVALUATE, RECONCILE_EVAL, RECONCILE_REVIEW.
	if advanceVerdict != "" {
		validStates := map[state.StateName]bool{
			state.StateEvaluate:        true,
			state.StateReconcileEval:   true,
			state.StateReconcileReview: true,
		}
		if !validStates[s.State] {
			return fmt.Errorf("--verdict is only valid in EVALUATE, RECONCILE_EVAL, or RECONCILE_REVIEW state (current: %s)", s.State)
		}
	}

	return nil
}

// printAdvanceWarnings prints warnings about flags that will be ignored due to config settings.
func printAdvanceWarnings(w interface{ Write([]byte) (int, error) }, s *state.ForgeState) {
	evalStates := map[state.StateName]bool{
		state.StateEvaluate:      true,
		state.StateReconcileEval: true,
	}

	// Warn if --eval-report provided but eval output is disabled.
	if advanceEvalReport != "" && evalStates[s.State] {
		var enabled bool
		switch s.Phase {
		case state.PhaseSpecifying:
			enabled = s.Config.Specifying.Eval.EnableEvalOutput
		case state.PhasePlanning:
			enabled = s.Config.Planning.Eval.EnableEvalOutput
		case state.PhaseImplementing:
			enabled = s.Config.Implementing.Eval.EnableEvalOutput
		}
		if !enabled {
			fmt.Fprintf(w, "warning: ignoring --eval-report: eval output is not enabled\n")
		}
	}

	// Warn if --message provided but commits are disabled.
	if advanceMessage != "" && !s.Config.General.EnableCommits {
		commitStates := map[state.StateName]bool{
			state.StateComplete:  true,
			state.StateAccept:    true,
			state.StateImplement: true,
			state.StateCommit:    true,
		}
		if commitStates[s.State] {
			fmt.Fprintf(w, "warning: ignoring --message: commits are not enabled\n")
		}
	}
}
