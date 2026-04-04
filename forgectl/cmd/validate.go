package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"forgectl/state"

	"github.com/spf13/cobra"
)

var validateType string

// osExit is a variable so tests can override it to avoid calling os.Exit.
var osExit = os.Exit

var validateCmd = &cobra.Command{
	Use:   "validate <file>",
	Short: "Validate a JSON file (spec-queue, plan-queue, or plan)",
	Args:  cobra.ExactArgs(1),
	RunE:  runValidate,
}

func init() {
	validateCmd.Flags().StringVar(&validateType, "type", "", "File type override: spec-queue, plan-queue, plan")
	rootCmd.AddCommand(validateCmd)
}

// typeToKey maps file type names to their expected top-level JSON key.
var typeToKey = map[string]string{
	"spec-queue": "specs",
	"plan-queue": "plans",
	"plan":       "context",
}

// keyToType maps top-level JSON keys to file type names.
var keyToType = map[string]string{
	"specs":   "spec-queue",
	"plans":   "plan-queue",
	"context": "plan",
}

func runValidate(cmd *cobra.Command, args []string) error {
	file := args[0]
	out := cmd.OutOrStdout()

	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", file, err)
	}

	// Parse JSON to inspect top-level keys.
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return fmt.Errorf("invalid JSON: %v", err)
	}

	// Find which recognized key is present.
	var detectedKey string
	for _, k := range []string{"specs", "plans", "context"} {
		if _, ok := top[k]; ok {
			detectedKey = k
			break
		}
	}

	fileType := validateType
	if fileType != "" {
		// --type override: validate the expected key matches.
		expectedKey, ok := typeToKey[fileType]
		if !ok {
			return fmt.Errorf("unknown type %q: must be spec-queue, plan-queue, or plan", fileType)
		}
		if detectedKey != expectedKey {
			hint := ""
			if foundType, ok := keyToType[detectedKey]; ok {
				hint = fmt.Sprintf("\n  Hint: did you mean --type %s?", foundType)
			}
			foundKey := detectedKey
			if foundKey == "" {
				// Pick the first key found for the error message.
				for k := range top {
					foundKey = k
					break
				}
			}
			return fmt.Errorf("--type %s expects top-level key %q, found %q.%s", fileType, expectedKey, foundKey, hint)
		}
	} else {
		// Auto-detect.
		if detectedKey == "" {
			foundKey := ""
			for k := range top {
				foundKey = k
				break
			}
			msg := "Error: cannot detect file type.\n"
			msg += "  Expected one of these top-level keys:\n"
			msg += "    \"specs\"    → spec-queue (used in specifying phase)\n"
			msg += "    \"plans\"    → plan-queue (used in planning phase)\n"
			msg += "    \"context\"  → plan.json  (used in planning/implementing phases)\n"
			msg += fmt.Sprintf("  Found: %q\n", foundKey)
			msg += "  Hint: use --type to specify the file type explicitly."
			fmt.Fprintln(out, msg)
			osExit(1)
			return nil
		}
		fileType = keyToType[detectedKey]
		fmt.Fprintf(out, "Detected: %s (top-level key: %q)\n\n", fileType, detectedKey)
	}

	// Run validation.
	var errs []string
	var entryCount int
	switch fileType {
	case "spec-queue":
		errs = state.ValidateSpecQueue(data)
		var sq state.SpecQueueInput
		if json.Unmarshal(data, &sq) == nil {
			entryCount = len(sq.Specs)
		}
	case "plan-queue":
		errs = state.ValidatePlanQueue(data)
		var pq state.PlanQueueInput
		if json.Unmarshal(data, &pq) == nil {
			entryCount = len(pq.Plans)
		}
	case "plan":
		baseDir := filepath.Dir(file)
		errs = state.ValidatePlanJSON(data, baseDir)
	}

	base := filepath.Base(file)
	if len(errs) == 0 {
		if fileType == "plan" {
			fmt.Fprintf(out, "Validated: %s — no errors.\n", base)
		} else {
			fmt.Fprintf(out, "Validated: %s — %d entries, no errors.\n", base, entryCount)
		}
		return nil
	}

	fmt.Fprintf(out, "Error: validation failed with %d errors:\n", len(errs))
	for i, e := range errs {
		fmt.Fprintf(out, "  %d. %s\n", i+1, e)
	}
	osExit(1)
	return nil
}
