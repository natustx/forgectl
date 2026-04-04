package cmd

import (
	"forgectl/state"

	"github.com/spf13/cobra"
)

var statusVerbose bool

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print current state with action guidance and session overview",
	RunE:  runStatus,
}

func init() {
	statusCmd.Flags().BoolVarP(&statusVerbose, "verbose", "v", false, "Show full session overview")
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	s, err := state.Load(stateDir)
	if err != nil {
		return err
	}

	state.PrintStatus(cmd.OutOrStdout(), s, stateDir, statusVerbose)
	return nil
}
