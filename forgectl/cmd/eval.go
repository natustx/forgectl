package cmd

import (
	"forgectl/state"

	"github.com/spf13/cobra"
)

var evalCmd = &cobra.Command{
	Use:   "eval",
	Short: "Output full evaluation context for the sub-agent",
	Long:  "Only valid in EVALUATE states (planning and implementing phases).",
	RunE:  runEval,
}

func init() {
	rootCmd.AddCommand(evalCmd)
}

func runEval(cmd *cobra.Command, args []string) error {
	projectRoot, stateDir, err := resolveSession()
	if err != nil {
		return err
	}
	s, err := state.Load(stateDir)
	if err != nil {
		return err
	}

	return state.PrintEvalOutput(cmd.OutOrStdout(), s, projectRoot)
}
