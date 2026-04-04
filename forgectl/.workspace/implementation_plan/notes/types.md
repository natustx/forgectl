# Types Notes

## New Phase Constant

```go
PhaseGeneratePlanningQueue PhaseName = "generate_planning_queue"
```

## New State Constants

```go
StateSelfReview StateName = "SELF_REVIEW"
```

## ForgeConfig Struct Hierarchy

All configuration locked into state at init time. Loaded from `.forgectl/config` TOML.

```go
type AgentConfig struct {
    Model string `json:"model"` // "opus", "haiku", "sonnet"
    Type  string `json:"type"`  // "eval", "explore", "refine"
    Count int    `json:"count"`
}

type EvalConfig struct {
    MinRounds       int         `json:"min_rounds"`
    MaxRounds       int         `json:"max_rounds"`
    AgentConfig     AgentConfig // embedded (model, type, count)
    EnableEvalOutput bool       `json:"enable_eval_output"`
}

type CrossRefConfig struct {
    MinRounds  int         `json:"min_rounds"`
    MaxRounds  int         `json:"max_rounds"`
    AgentConfig AgentConfig // embedded (model, type, count)
    UserReview bool        `json:"user_review"`
    Eval       AgentConfig `json:"eval"` // separate eval agent config
}

type ReconciliationConfig struct {
    MinRounds  int         `json:"min_rounds"`
    MaxRounds  int         `json:"max_rounds"`
    AgentConfig AgentConfig // embedded
    UserReview bool        `json:"user_review"`
}

type SpecifyingConfig struct {
    Batch          int                  `json:"batch"`
    CommitStrategy string               `json:"commit_strategy"` // default "all-specs"
    Eval           EvalConfig           `json:"eval"`
    CrossReference CrossRefConfig       `json:"cross_reference"`
    Reconciliation ReconciliationConfig `json:"reconciliation"`
}

type StudyCodeConfig struct {
    AgentConfig // embedded (model, type, count)
}

type RefineConfig struct {
    AgentConfig // embedded
}

type PlanningConfig struct {
    Batch                    int             `json:"batch"`
    CommitStrategy           string          `json:"commit_strategy"` // default "strict"
    SelfReview               bool            `json:"self_review"`
    PlanAllBeforeImplementing bool           `json:"plan_all_before_implementing"`
    StudyCode                StudyCodeConfig `json:"study_code"`
    Eval                     EvalConfig      `json:"eval"`
    Refine                   RefineConfig    `json:"refine"`
}

type ImplementingConfig struct {
    Batch          int        `json:"batch"`
    CommitStrategy string     `json:"commit_strategy"` // default "scoped"
    Eval           EvalConfig `json:"eval"`
}

type DomainConfig struct {
    Name string `json:"name"`
    Path string `json:"path"`
}

type PathsConfig struct {
    StateDir     string `json:"state_dir"`     // default ".forgectl/state"
    WorkspaceDir string `json:"workspace_dir"` // default ".forge_workspace"
}

type LogsConfig struct {
    Enabled       bool `json:"enabled"`        // default true
    RetentionDays int  `json:"retention_days"` // default 90
    MaxFiles      int  `json:"max_files"`      // default 50
}

type GeneralConfig struct {
    EnableCommits bool `json:"enable_commits"` // default false
    UserGuided    bool `json:"user_guided"`    // default false
}

type ForgeConfig struct {
    General        GeneralConfig      `json:"general"`
    Domains        []DomainConfig     `json:"domains"`
    Specifying     SpecifyingConfig   `json:"specifying"`
    Planning       PlanningConfig     `json:"planning"`
    Implementing   ImplementingConfig `json:"implementing"`
    Paths          PathsConfig        `json:"paths"`
    Logs           LogsConfig         `json:"logs"`
}
```

### JSON Schema in State File

Per state-persistence.md, the config is stored flat in the state JSON using the same nesting as the TOML. The `agent_type` and `agent_count` from old code are replaced by the new AgentConfig sub-objects with `model`, `type`, `count`.

## ForgeState Changes

Remove from ForgeState:
- `BatchSize int` (now in Config.Specifying.Batch, Config.Planning.Batch, Config.Implementing.Batch)
- `MinRounds int` (now in Config.*.Eval.MinRounds)
- `MaxRounds int` (now in Config.*.Eval.MaxRounds)
- `UserGuided bool` (now in Config.General.UserGuided)

Add to ForgeState:
- `Config ForgeConfig` — full config locked at init
- `SessionID string` — UUID v4 generated at init

Add ForgeState field for generate_planning_queue:
- `GeneratePlanningQueue *GeneratePlanningQueueState`

## GeneratePlanningQueueState Struct

```go
type GeneratePlanningQueueState struct {
    PlanQueueFile string       `json:"plan_queue_file"` // path to generated plan-queue.json
    Evals         []EvalRecord `json:"evals,omitempty"` // not used currently, reserved
}
```

## PlanningState Changes

Change Completed field:
```go
// Old:
Completed []interface{} `json:"completed"`

// New:
Completed []CompletedPlan `json:"completed"`
```

Add new struct:
```go
type CompletedPlan struct {
    ID     int    `json:"id"`
    Name   string `json:"name"`
    Domain string `json:"domain"`
    File   string `json:"file"`
}
```

## ImplementingState Changes

Add plan_queue for all-planning-first mode:
```go
// Add to ImplementingState:
PlanQueue []PlanQueueEntry `json:"plan_queue,omitempty"`
```

## PlanItem Schema Changes

```go
// Old:
Spec string `json:"spec,omitempty"`
Ref  string `json:"ref,omitempty"`

// New:
Specs []string `json:"specs,omitempty"` // spec refs, display only, #anchors OK, not validated on disk
Refs  []string `json:"refs,omitempty"`  // notes refs, validated on disk, relative to plan.json dir
```

Validation in validate.go:
- `items[].specs`: no file-existence check (display references, #anchors permitted)
- `items[].refs`: each path validated on disk, relative to plan.json directory
- `items[].files`: paths relative to project root, currently not validated on disk
- `refs[].path`: unchanged — relative to plan.json directory, validated on disk

## PhaseShiftInfo Changes

```go
// Existing (no change needed to struct itself):
type PhaseShiftInfo struct {
    From PhaseName `json:"from"`
    To   PhaseName `json:"to"`
}
// But To can now be "generate_planning_queue" in addition to existing values.
```

## PlanQueueEntry Changes

The spec (phase-transitions.md) references `spec_commits` in the generated plan queue:

```go
// Add to PlanQueueEntry:
SpecCommits []string `json:"spec_commits,omitempty"` // deduplicated commit hashes from completed specs
```
