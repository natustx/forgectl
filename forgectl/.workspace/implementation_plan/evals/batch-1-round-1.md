# Evaluation Report

**Round:** 1
**Batch:** 1
**Layer:** L0 Foundation — Types and Config Structs

VERDICT: FAIL

## Items Evaluated

### [types.config] ForgeConfig struct hierarchy

**Files reviewed:** `forgectl/state/types.go`

#### Test Results

- [PASS] `AgentConfig` struct defined with `model`, `type`, `count` fields and correct JSON tags
- [PASS] `EvalConfig` struct defined with `MinRounds`, `MaxRounds`, embedded `AgentConfig` (flat JSON promotion), and `EnableEvalOutput`
- [PASS] `CrossRefConfig` struct defined with `MinRounds`, `MaxRounds`, embedded `AgentConfig`, `UserReview`, and nested `Eval AgentConfig`
- [PASS] `ReconciliationConfig` struct defined with `MinRounds`, `MaxRounds`, embedded `AgentConfig`, `UserReview`
- [PASS] `SpecifyingConfig` struct defined with `Batch`, `CommitStrategy`, `Eval`, `CrossReference`, `Reconciliation`
- [PASS] `StudyCodeConfig` and `RefineConfig` defined with embedded `AgentConfig`
- [PASS] `PlanningConfig` defined with `Batch`, `CommitStrategy`, `SelfReview`, `PlanAllBeforeImplementing`, `StudyCode`, `Eval`, `Refine`
- [PASS] `ImplementingConfig` defined with `Batch`, `CommitStrategy`, `Eval`
- [PASS] `DomainConfig`, `PathsConfig`, `LogsConfig`, `GeneralConfig` structs defined correctly
- [PASS] `ForgeConfig` top-level struct combines all sub-configs
- [PASS] `DefaultForgeConfig()` returns `enable_commits=false`
- [PASS] `DefaultForgeConfig()` sets commit strategy defaults: `specifying="all-specs"`, `planning="strict"`, `implementing="scoped"`
- [PASS] `DefaultForgeConfig()` sets `min_rounds=1`, `max_rounds=3` for all eval configs
- [FAIL] `ForgeConfig` is not embedded in or referenced by `ForgeState` — the state file schema requires a top-level `"config"` key, but `ForgeState` has no `Config ForgeConfig` field
  - The spec (`state-persistence.md` § State File Schema) shows `"config": { ... }` as a first-class field in `forgectl-state.json`. Without a `Config ForgeConfig \`json:"config"\`` field on `ForgeState`, the schema cannot be satisfied.
- [FAIL] `ForgeState` still retains the old flat fields `BatchSize int`, `MinRounds int`, `MaxRounds int`, `UserGuided bool` that the spec says should be replaced by the `ForgeConfig` hierarchy
  - These fields produce stale JSON keys (`batch_size`, `min_rounds`, `max_rounds`, `user_guided`) at the state root level, contradicting the spec schema which places these values under `config.*`.

#### Notes

The `ForgeConfig` struct hierarchy itself is well-designed and matches the spec schema. The `AgentConfig` embedding pattern correctly produces flat JSON (model/type/count promoted to the same JSON level as min_rounds/max_rounds). `DefaultForgeConfig()` values are accurate.

The deficiency is a wiring gap: the item's step 6 ("Add ForgeConfig top-level struct combining all sub-configs") was completed for the `ForgeConfig` type itself, but the integration into `ForgeState` — replacing the flat fields — was not done. The spec's state file schema requires `"config": { ... }` in the JSON output of `ForgeState`, which can only happen if `ForgeState` has a `Config ForgeConfig \`json:"config"\`` field.

## Deficiencies

- Add `Config ForgeConfig \`json:"config"\`` field to `ForgeState` in `forgectl/state/types.go`. The state-persistence.md schema places all configuration under a top-level `"config"` key; without this field, `ForgeState` cannot marshal to the required schema.
- Remove the deprecated flat fields `BatchSize int \`json:"batch_size"\``, `MinRounds int \`json:"min_rounds"\``, `MaxRounds int \`json:"max_rounds"\``, and `UserGuided bool \`json:"user_guided"\`` from `ForgeState` in `forgectl/state/types.go`. These are replaced by the nested `ForgeConfig` hierarchy under `Config`. (Note: removing these fields may require updating any code in `cmd/` or `state/` that currently reads them — those callsites should be updated to use `state.Config.General.UserGuided`, `state.Config.Specifying.Eval.MinRounds`, etc.)

## Summary

The `ForgeConfig` struct hierarchy is complete and correct in isolation. However, `ForgeState` was not updated to include `Config ForgeConfig` and still carries the old flat fields that this item was meant to replace. Two targeted changes to `ForgeState` are needed to satisfy the acceptance criteria.
