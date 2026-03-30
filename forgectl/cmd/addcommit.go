package cmd

import (
	"fmt"
	"path/filepath"

	"forgectl/state"

	"github.com/spf13/cobra"
)

var (
	addCommitID   int
	addCommitHash string
)

var addCommitCmd = &cobra.Command{
	Use:   "add-commit",
	Short: "Register a commit hash to a completed spec",
	RunE:  runAddCommit,
}

func init() {
	addCommitCmd.Flags().IntVar(&addCommitID, "id", 0, "Spec ID (required)")
	addCommitCmd.Flags().StringVar(&addCommitHash, "hash", "", "Commit hash (required)")
	_ = addCommitCmd.MarkFlagRequired("id")
	_ = addCommitCmd.MarkFlagRequired("hash")
	rootCmd.AddCommand(addCommitCmd)
}

func runAddCommit(cmd *cobra.Command, args []string) error {
	_, stateDir, err := resolveSession()
	if err != nil {
		return err
	}
	s, err := state.Load(stateDir)
	if err != nil {
		return err
	}

	absDir, _ := filepath.Abs(stateDir)
	repoRoot, err := state.GitRepoRoot(absDir)
	if err != nil {
		return fmt.Errorf("cannot find git repo: %w", err)
	}
	if err := state.GitHashExists(repoRoot, addCommitHash); err != nil {
		return fmt.Errorf("commit does not exist in the repository")
	}

	if err := state.AddCommitToSpec(s, addCommitID, addCommitHash); err != nil {
		return err
	}

	if err := state.Save(stateDir, s); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Added %s to spec %d.\n", addCommitHash, addCommitID)
	return nil
}
