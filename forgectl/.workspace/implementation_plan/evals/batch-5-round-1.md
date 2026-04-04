# Evaluation Report

**Round:** 1
**Batch:** 5
**Layer:** L2 State Machine Changes

VERDICT: PASS

## Items Evaluated

### [advance.flags] enable_eval_output, enable_commits flag gating, and --guided handling

**Files reviewed:**
- `forgectl/state/advance.go`
- `forgectl/cmd/advance.go`

#### Test Results

- [PASS] advance in EVALUATE with EnableEvalOutput=false succeeds without --eval-report
  - `advancePlanning` (and specifying/implementing equivalents): when `planEvalEnabled` is false, `--eval-report` is not required. No error is returned if `in.EvalReport == ""`.

- [PASS] advance in EVALUATE with EnableEvalOutput=true fails without --eval-report
  - All three phase EVALUATE handlers gate on `EnableEvalOutput`: if true and `in.EvalReport == ""`, returns error `"--eval-report is required in EVALUATE state"`.

- [PASS] advance in EVALUATE with EnableEvalOutput=false and --eval-report provided prints warning and proceeds
  - `printAdvanceWarnings` in `cmd/advance.go` checks: if `advanceEvalReport != ""` and the current state is an eval state and the phase's eval output is disabled, it prints `"warning: ignoring --eval-report: eval output is not enabled"`. The command then proceeds. Note: the exact warning text differs from the spec (`"--eval-report is ignored, eval output is not enabled"` vs the implementation's `"warning: ignoring --eval-report: eval output is not enabled"`), but the functional behavior (warn and proceed) is correct.

- [PASS] advance at COMPLETE with EnableCommits=true and no --message fails
  - `advanceSpecifying` at `StateComplete`: `if s.Config.General.EnableCommits && in.Message == ""` returns error.

- [PASS] advance at COMPLETE with EnableCommits=false and --message provided prints warning and proceeds
  - `printAdvanceWarnings` checks `advanceMessage != ""` and `!s.Config.General.EnableCommits` and `StateComplete` is in the `commitStates` map. Warning `"warning: ignoring --message: commits are not enabled"` is printed. Command proceeds.

- [PASS] advance at implementing IMPLEMENT (first round, EvalRound==0) with EnableCommits=true and no --message fails
  - `advanceImplFromImplement`: `if batch.EvalRound == 0 && s.Config.General.EnableCommits && in.Message == ""` returns error `"--message is required for first-round implementation when enable_commits is true"`.

- [PASS] advance at implementing IMPLEMENT with EnableCommits=false and --message provided prints warning and proceeds
  - `printAdvanceWarnings`: `StateImplement` is in `commitStates`, `advanceMessage != ""`, `!s.Config.General.EnableCommits` → prints warning and proceeds.

- [PASS] advance with --guided=true at any state (including PHASE_SHIFT) sets Config.General.UserGuided=true in persisted state
  - At the top of `Advance()` (before phase dispatch and before PHASE_SHIFT handling): `if in.Guided != nil { s.Config.General.UserGuided = *in.Guided }`. The `cmd/advance.go` sets `guided = &g` (true) when `--guided` flag is changed. This fires on every call including PHASE_SHIFT.

- [PASS] advance with --no-guided at PHASE_SHIFT sets Config.General.UserGuided=false before the transition proceeds
  - Same mechanism: `--no-guided` sets `guided = &g` (false). Applied at the top of `Advance()` before `advancePhaseShift()` is called. The guided update precedes the transition.

#### Notes

The warning message text in the implementation differs slightly from the spec:
- Spec: `--eval-report is ignored, eval output is not enabled`
- Code: `warning: ignoring --eval-report: eval output is not enabled`
- Spec: `--message is ignored, commits are not enabled`
- Code: `warning: ignoring --message: commits are not enabled`

The functional behavior (warn and proceed without error) matches the spec. The exact string difference is minor and does not affect any test criteria which only check that the command proceeds.

---

### [advance.selfreview] SELF_REVIEW state in planning phase

**Files reviewed:**
- `forgectl/state/advance.go`
- `forgectl/state/output.go`

#### Test Results

- [PASS] Planning with SelfReview=true transitions VALIDATE→SELF_REVIEW→EVALUATE
  - `advancePlanningFromValidate` (and `advancePlanningFromDraftOrRefine`): when `s.Config.Planning.SelfReview` is true, sets `s.State = StateSelfReview`. Advancing from `StateSelfReview` in `advancePlanning` re-validates plan.json; if valid, sets `s.State = StateEvaluate`.

- [PASS] Planning with SelfReview=false transitions VALIDATE→EVALUATE (skips SELF_REVIEW)
  - `advancePlanningFromValidate`: when `SelfReview` is false, sets `s.State = StateEvaluate` directly. SELF_REVIEW is skipped.

#### Notes

The `output.go` includes a `StateSelfReview` case in `printPlanningOutput` (lines 270-289) with correct action description. The SELF_REVIEW state also re-validates on advance, which aligns with the spec invariant that "The validation gate also runs on advance from SELF_REVIEW in case the agent revised plan.json."

---

### [advance.genqueue] generate_planning_queue phase state machine

**Files reviewed:**
- `forgectl/state/advance.go`
- `forgectl/state/types.go`

#### Test Results

- [PASS] specifying COMPLETE → PHASE_SHIFT with From=specifying, To=generate_planning_queue
  - `advanceSpecifying` at `StateComplete`: sets `s.State = StatePhaseShift` and `s.PhaseShift = &PhaseShiftInfo{From: PhaseSpecifying, To: PhaseGeneratePlanningQueue}`.

- [PASS] PHASE_SHIFT(specifying) advance without --from: transitions to generate_planning_queue ORIENT and writes plan-queue.json
  - `advancePhaseShift` when `From == PhaseSpecifying && To == PhaseGeneratePlanningQueue` and `in.From == ""`: calls `autoGeneratePlanQueue`, sets `s.GeneratePlanningQueue = &GeneratePlanningQueueState{PlanQueueFile: planQueueFile}`, sets `s.Phase = PhaseGeneratePlanningQueue`, `s.State = StateOrient`, clears `s.PhaseShift`.

- [PASS] Auto-generated plan-queue.json groups specs by domain and sets code_search_roots default to ['<domain>/']
  - `autoGeneratePlanQueue`: groups `s.Specifying.Completed` by `spec.Domain` in first-appearance order. When no `DomainRoots` entry exists for a domain, defaults to `[]string{domain + "/"}`. One `PlanQueueEntry` per domain is written to `plan-queue.json`.

- [PASS] Auto-generated plan-queue.json uses set-roots data when available for a domain
  - `autoGeneratePlanQueue` lines 757-765: checks `s.Specifying.DomainRoots[domain]`; if present and non-empty, uses those roots instead of the default.

- [PASS] generate_planning_queue ORIENT → REFINE on advance
  - `advanceGeneratePlanningQueue` at `StateOrient`: sets `s.State = StateRefine`.

- [PASS] generate_planning_queue REFINE advance: invalid plan-queue.json prints errors and stays REFINE
  - `advanceGeneratePlanningQueue` at `StateRefine`: reads `PlanQueueFile`, calls `ValidatePlanQueue`. If errors, returns `&ValidationError{Errors: validationErrs}`. In `runAdvance`, `ValidationError` saves state (which remains REFINE since no mutation occurred) and prints errors. State stays REFINE.

- [PASS] generate_planning_queue REFINE advance: valid plan-queue.json transitions to PHASE_SHIFT(genqueue→planning)
  - `advanceGeneratePlanningQueue` at `StateRefine`: if validation passes, sets `s.State = StatePhaseShift` and `s.PhaseShift = &PhaseShiftInfo{From: PhaseGeneratePlanningQueue, To: PhasePlanning}`.

- [PASS] generate_planning_queue PHASE_SHIFT advance: transitions to planning ORIENT with correct CurrentPlan
  - `advancePhaseShift` when `From == PhaseGeneratePlanningQueue && To == PhasePlanning`: reads plan queue from `PlanQueueFile`, calls `populatePlanningFromQueue`, sets `s.Planning`, `s.Phase = PhasePlanning`, `s.State = StateOrient`. `populatePlanningFromQueue` pulls the first entry from the queue into `CurrentPlan` with all fields populated.

- [PASS] PHASE_SHIFT(specifying) advance with --from: skips generate_planning_queue, transitions directly to planning ORIENT
  - `advancePhaseShift` when `From == PhaseSpecifying && To == PhaseGeneratePlanningQueue` and `in.From != ""`: reads and validates the --from file, calls `populatePlanningFromQueue`, sets `s.Phase = PhasePlanning`, `s.State = StateOrient`, clears `s.PhaseShift`. generate_planning_queue phase is never entered.

- [PASS] PHASE_SHIFT(specifying) advance with invalid --from: prints errors, stays PHASE_SHIFT
  - `advancePhaseShift` with `in.From` set to invalid file: `ValidatePlanQueue` returns errors, function returns `&ValidationError{Errors: validationErrs}`. In `runAdvance`, the `ValidationError` handler saves state (state not yet mutated, still PHASE_SHIFT) and prints errors. State stays PHASE_SHIFT.

#### Notes

The `GeneratePlanningQueueState` type in `types.go` includes `PlanQueueFile` and `Evals` fields, and is correctly wired into `ForgeState.GeneratePlanningQueue`. The `PhaseGeneratePlanningQueue` constant is properly defined. The `output.go` has correct `printGeneratePlanningQueueOutput` for both ORIENT and REFINE states.

---

## Summary

All 19 test cases across the three items pass. The state machine correctly implements flag gating for `enable_eval_output` and `enable_commits`, the `--guided`/`--no-guided` mutation at any state (including PHASE_SHIFT), the optional SELF_REVIEW state in planning, and the full generate_planning_queue phase state machine including auto-generation of plan-queue.json, ORIENT/REFINE/PHASE_SHIFT transitions, and both the --from override path and the default generation path. Minor cosmetic differences in warning message formatting are noted but do not affect functional correctness.
