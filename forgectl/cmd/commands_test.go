package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"forgectl/state"
)

// setupProjectDir creates a temp dir with .forgectl/ and an empty config,
// changes cwd into it, and registers cleanup to restore cwd.
func setupProjectDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".forgectl"), 0755); err != nil {
		t.Fatal(err)
	}
	// Empty config — all defaults apply.
	if err := os.WriteFile(filepath.Join(dir, ".forgectl", "config"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
	return dir
}

// resolvedStateDir returns the expected state dir for a project root using DefaultConfig.
func resolvedStateDir(projectRoot string) string {
	cfg := state.DefaultConfig()
	return state.StateDir(projectRoot, cfg)
}

func TestInitCommand(t *testing.T) {
	dir := setupProjectDir(t)

	// Write spec queue.
	input := state.SpecQueueInput{
		Specs: []state.SpecQueueEntry{
			{Name: "Spec A", Domain: "test", Topic: "topic A", File: "specs/a.md", PlanningSources: []string{}, DependsOn: []string{}},
			{Name: "Spec B", Domain: "test", Topic: "topic B", File: "specs/b.md", PlanningSources: []string{}, DependsOn: []string{}},
		},
	}
	data, _ := json.Marshal(input)
	queueFile := filepath.Join(dir, "specs-queue.json")
	os.WriteFile(queueFile, data, 0644)

	initFrom = queueFile
	initPhase = "specifying"

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	err := runInit(initCmd, nil)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// Load state from the resolved state dir and verify.
	sd := resolvedStateDir(dir)
	s, err := state.Load(sd)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if s.Phase != state.PhaseSpecifying {
		t.Errorf("phase = %s, want specifying", s.Phase)
	}
	if s.State != state.StateOrient {
		t.Errorf("state = %s, want ORIENT", s.State)
	}
	if len(s.Specifying.Queue) != 2 {
		t.Errorf("queue has %d specs, want 2", len(s.Specifying.Queue))
	}
	if s.SessionID == "" {
		t.Error("session_id should be set")
	}
}

func TestInitLocksConfig(t *testing.T) {
	dir := setupProjectDir(t)

	// Write a custom config to verify it gets locked in.
	customCfg := `[specifying]
batch = 5

[implementing]
batch = 7
`
	os.WriteFile(filepath.Join(dir, ".forgectl", "config"), []byte(customCfg), 0644)

	input := state.SpecQueueInput{
		Specs: []state.SpecQueueEntry{
			{Name: "Spec A", Domain: "test", Topic: "t", File: "a.md", PlanningSources: []string{}, DependsOn: []string{}},
		},
	}
	data, _ := json.Marshal(input)
	queueFile := filepath.Join(dir, "queue.json")
	os.WriteFile(queueFile, data, 0644)

	initFrom = queueFile
	initPhase = "specifying"

	err := runInit(initCmd, nil)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	sd := resolvedStateDir(dir)
	s, err := state.Load(sd)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if s.Config.Specifying.Batch != 5 {
		t.Errorf("config.specifying.batch = %d, want 5", s.Config.Specifying.Batch)
	}
	if s.Config.Implementing.Batch != 7 {
		t.Errorf("config.implementing.batch = %d, want 7", s.Config.Implementing.Batch)
	}
}

func TestInitRejectsExistingState(t *testing.T) {
	dir := setupProjectDir(t)

	// Create existing state in the resolved state dir.
	sd := resolvedStateDir(dir)
	if err := os.MkdirAll(sd, 0755); err != nil {
		t.Fatal(err)
	}
	s := &state.ForgeState{Phase: state.PhaseSpecifying, State: state.StateOrient}
	state.Save(sd, s)

	initFrom = "dummy"
	initPhase = "specifying"

	err := runInit(initCmd, nil)
	if err == nil {
		t.Error("expected error for existing state file")
	}
}

func TestInitRejectsInvalidPhase(t *testing.T) {
	setupProjectDir(t)
	initFrom = "dummy"
	initPhase = "invalid"

	err := runInit(initCmd, nil)
	if err == nil {
		t.Error("expected error for invalid phase")
	}
}

func TestInitRejectsInvalidConfig(t *testing.T) {
	dir := setupProjectDir(t)

	// Write config with constraint violation.
	badCfg := `[specifying.eval]
min_rounds = 5
max_rounds = 2
`
	os.WriteFile(filepath.Join(dir, ".forgectl", "config"), []byte(badCfg), 0644)

	initFrom = "dummy"
	initPhase = "specifying"

	err := runInit(initCmd, nil)
	if err == nil {
		t.Error("expected error for invalid config")
	}
}

func TestStatusCommand(t *testing.T) {
	dir := setupProjectDir(t)

	// Save state to the resolved state dir.
	sd := resolvedStateDir(dir)
	if err := os.MkdirAll(sd, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := state.DefaultConfig()
	s := &state.ForgeState{
		Phase:          state.PhaseSpecifying,
		State:          state.StateOrient,
		Config:         cfg,
		StartedAtPhase: state.PhaseSpecifying,
		Specifying:     state.NewSpecifyingState([]state.SpecQueueEntry{}),
	}
	state.Save(sd, s)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	err := runStatus(statusCmd, nil)
	if err != nil {
		t.Fatalf("status: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("status should produce output")
	}
}
