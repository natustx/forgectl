package cmd

import (
	"fmt"
	"os"

	"forgectl/state"

	"github.com/spf13/cobra"
)

var isDoneCmd = &cobra.Command{
	Use:   "is-done",
	Short: "Exit 0 if the session has reached its terminal state, 1 otherwise",
	Long:  "Reports whether the active session has no work remaining. Exit code 0 means done, 1 means work remains or no session exists.",
	RunE:  runIsDone,
}

func init() {
	rootCmd.AddCommand(isDoneCmd)
}

func runIsDone(cmd *cobra.Command, args []string) error {
	_, stateDir, _, err := resolveSession()
	if err != nil {
		return err
	}
	s, err := state.Load(stateDir)
	if err != nil {
		return err
	}

	if state.IsTerminal(s) {
		fmt.Fprintln(cmd.OutOrStdout(), "done")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "not done: %s (%s)\n", s.State, s.Phase)
	os.Exit(1)
	return nil
}
