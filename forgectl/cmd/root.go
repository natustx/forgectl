package cmd

import (
	"fmt"
	"os"

	"forgectl/state"

	"github.com/spf13/cobra"
)

var version = "v0.0.1"

var rootCmd = &cobra.Command{
	Use:     "forgectl",
	Short:   "Software development lifecycle scaffold",
	Long:    "Manages the full software development lifecycle — specifying, planning, implementing — through a JSON-backed state machine.",
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// resolveSession discovers the project root, loads config, and returns
// (projectRoot, stateDir). projectRoot is used to resolve relative paths in state;
// stateDir is used for Load/Save/Exists operations.
func resolveSession() (projectRoot, stateDir string, err error) {
	projectRoot, err = state.FindProjectRoot()
	if err != nil {
		return
	}
	cfg, err := state.LoadConfig(projectRoot)
	if err != nil {
		return
	}
	stateDir = state.StateDir(projectRoot, cfg)
	return
}
