# Evaluation Report

**Round:** 1
**Batch:** 6/7
**Layer:** L2 State Machine Changes

VERDICT: FAIL

## Items Evaluated

### [advance.phaseshift] Multi-plan phase transitions (plan_all_before_implementing)

**Files reviewed:** `forgectl/state/advance.go`

#### Test Results

- [PASS] PlanAllBeforeImplementing=false: planning ACCEPT → PHASE_SHIFT(planning→implementing) regardless of queue
  - Lines 322–325: `!s.Config.Planning.PlanAllBeforeImplementing` → sets PHASE_SHIFT(planning→implementing) directly.
- [PASS] PlanAllBeforeImplementing=true with plans in queue: planning ACCEPT → PHASE_SHIFT(planning→planning)
  - Lines 326–348: when queue non-empty, moves CurrentPlan to Completed, pulls next plan, sets PHASE_SHIFT(planning→planning).
- [PASS] PlanAllBeforeImplementing=true with empty queue: planning ACCEPT → PHASE_SHIFT(planning→implementing), implementing PlanQueue populated
  - Lines 349–361: last plan moved to Completed, sets PHASE_SHIFT(planning→implementing). Phase shift handler (lines 762–796) populates `Implementing.PlanQueue` from `Planning.Completed[1:]`.
- [PASS] PHASE_SHIFT(planning→planning) advance: new CurrentPlan set, state=ORIENT
  - Lines 754–760: resets Round/Evals, sets phase=PhasePlanning, state=StateOrient. CurrentPlan was already updated in the ACCEPT handler before PHASE_SHIFT was set.
- [PASS] Implementing DONE with Planning.Queue non-empty and PlanAllBeforeImplementing=false: → PHASE_SHIFT(implementing→planning)
  - Lines 483–490: interleaved mode checks `s.Planning.Queue` and sets PHASE_SHIFT(implementing→planning).
- [PASS] PHASE_SHIFT(implementing→planning) advance: phase=planning, new plan from queue, state=ORIENT
  - Lines 798–818: pops from Planning.Queue, sets CurrentPlan, resets Round/Evals, phase=PhasePlanning, state=StateOrient.
- [PASS] Implementing DONE with Implementing.PlanQueue non-empty and PlanAllBeforeImplementing=true: → PHASE_SHIFT(implementing→implementing)
  - Lines 491–497: all-first mode checks `impl.PlanQueue` and sets PHASE_SHIFT(implementing→implementing).
- [PASS] PHASE_SHIFT(implementing→implementing) advance: reads, validates, and mutates next plan.json; state=ORIENT
  - Lines 820–842: calls `mutatePlanForImplementing` (which reads, validates, mutates) then resets implementing state and sets state=StateOrient.
- [PASS] Implementing DONE with no plans remaining: terminal DONE state with session summary
  - Line 499: `return fmt.Errorf("session complete.")` when no plans remain in either queue.
- [PASS] Single domain: implementing DONE with empty Planning.Queue goes to terminal DONE (not phase shift)
  - Lines 483–499: with `PlanAllBeforeImplementing=false` and empty `Planning.Queue`, falls through to `return fmt.Errorf("session complete.")`.
- [PASS] advance at planning DONE with --verdict flag returns 'DONE is a pass-through state. No flags accepted.' and exits 1
  - Lines 363–368: the StateDone case checks for any flag and returns the exact string `"DONE is a pass-through state. No flags accepted."`.
- [PASS] advance with --guided at PHASE_SHIFT updates Config.General.UserGuided before the phase transition fires
  - Lines 13–15: `in.Guided` is applied to `s.Config.General.UserGuided` unconditionally at the top of `Advance`, before the PHASE_SHIFT dispatch at line 19.

#### Notes

All twelve tests pass. The implementation correctly handles all four PHASE_SHIFT variants and the PlanAllBeforeImplementing branching logic. One observation: when `PlanAllBeforeImplementing=true`, the transition to PHASE_SHIFT(planning→implementing) is triggered directly from ACCEPT (not from DONE), bypassing the DONE state entirely. The spec describes "PHASE_SHIFT entered after planning DONE" but the functional tests only verify ACCEPT→PHASE_SHIFT behavior, so no test is broken by this ordering choice.

---

### [git.autocommit] Auto-commit: staging strategies and commit execution

**Files reviewed:** `forgectl/state/git.go`, `forgectl/state/advance.go`

#### Test Results

- [PASS] AutoCommit with 'all' strategy runs 'git add -A' then 'git commit -m <message>' and returns a non-empty hash
  - `git.go` lines 18–19: `case "all": addArgs = []string{"-C", projectRoot, "add", "-A"}`. Followed by `git commit` and `git rev-parse HEAD` returning the trimmed hash.
- [PASS] AutoCommit with 'tracked' strategy runs 'git add -u'
  - `git.go` lines 16–17: `case "tracked": addArgs = []string{"-C", projectRoot, "add", "-u"}`.
- [PASS] AutoCommit with 'strict' strategy runs 'git add' with specific file paths
  - `git.go` lines 13–16: `case "strict", "all-specs", "scoped"`: `addArgs = append([]string{"-C", projectRoot, "add"}, stageTargets...)`.
- [PASS] AutoCommit returns error when git commit fails (non-zero exit), state does not advance
  - `git.go` lines 30–33: `commitCmd.CombinedOutput()` failure returns an error. All call sites in `advance.go` propagate this error (lines 200, 318–320, 465–467), halting state advancement.
- [PASS] specifying COMPLETE with EnableCommits=true commits and registers hash on all CompletedSpecs.CommitHashes
  - `advance.go` lines 198–207: `AutoCommit` is called with spec file paths, and on success the returned hash is appended to `CommitHashes` for every entry in `s.Specifying.Completed`.
- [FAIL] AutoCommit when no files are staged (nothing to commit) skips commit execution and prints notice; state still advances
  - `AutoCommit` does not detect the "nothing to commit" condition. When `git commit` exits non-zero (e.g., with "nothing to commit, working tree clean"), the function returns an error. No notice is printed and `AutoCommit` does not distinguish this case from a real commit failure. The calling code in `advance.go` propagates the error, blocking state advancement.
- [FAIL] planning ACCEPT with EnableCommits=true and no staged files skips commit and prints notice
  - Same root cause: `advance.go` line 318 passes the error from `AutoCommit` directly to the caller, so a "nothing to commit" git exit causes ACCEPT to return an error instead of printing a notice and advancing.

#### Notes

The two edge case failures share a single root cause: `AutoCommit` in `git.go` does not inspect the `git commit` output for the "nothing to commit" string before treating a non-zero exit as a hard error. The fix requires detecting that condition in `AutoCommit`, skipping the commit, printing a notice, and returning `("", nil)` so callers treat it as a success.

---

## Deficiencies

- In `forgectl/state/git.go`, `AutoCommit` must detect the "nothing to commit" condition from `git commit`'s output (output contains "nothing to commit" or exit code 1 with that message), skip the commit, print a notice (e.g., `"notice: nothing to commit, skipping auto-commit"`), and return `("", nil)` instead of an error. This unblocks both the `AutoCommit` edge case test and the `planning ACCEPT` edge case test.

## Summary

Item `advance.phaseshift` passes all twelve tests — all four PHASE_SHIFT variants, the PlanAllBeforeImplementing branching, guided-flag ordering, and the DONE rejection are correctly implemented.

Item `git.autocommit` passes five of seven tests. Both failures are caused by `AutoCommit` not handling the "nothing to commit" case gracefully. All strategies are wired correctly and error propagation works; only the empty-staging edge case is missing.
