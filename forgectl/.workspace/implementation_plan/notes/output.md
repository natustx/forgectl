# Output Format Changes Notes

## generate_planning_queue Phase Outputs

### ORIENT output
```
State:   ORIENT
Phase:   generate_planning_queue

Generated: .forgectl/state/plan-queue.json

Advance to continue.
```

### REFINE output
```
State:   REFINE
Phase:   generate_planning_queue

Stop and review the generated plan queue .forgectl/state/plan-queue.json. Reorder and edit as needed.

Advance when ready.
```

### PHASE_SHIFT (generate_planning_queue → planning)
```
State:   PHASE_SHIFT
From:    generate_planning_queue → planning

Advance to continue.
```

## Updated PHASE_SHIFT Output (specifying → generate_planning_queue)

```
State:   PHASE_SHIFT
From:    specifying → generate_planning_queue

Stop and refresh your context, please.
When ready:
  forgectl advance                            # generate plan queue from completed specs
  forgectl advance --from <plan-queue.json>   # OR provide a plan queue (skips generation)
```

## SELF_REVIEW Output (planning)

```
State:   SELF_REVIEW
Phase:   planning
Plan:    <plan name>
Domain:  <domain>
File:    <plan file>
Round:   <round>/<max_rounds>
Specs:   <spec1>
         <spec2>
Notes:   <workspace>/implementation_plan/notes/
Action:  Review your plan against the specs and your study notes.
         Verify coverage, dependency ordering, and layer structure.
         Revise plan.json and notes as needed before evaluation.
         After completion of the above, advance to continue.
```

## STUDY_CODE Output Update

Add specs list:
```
State:   STUDY_CODE
Phase:   planning
Plan:    <plan name>
Domain:  <domain>
File:    <plan file>
Roots:   <root1>, <root2>
Specs:   <spec1>
         <spec2>
Action:  Explore the codebase in relation to the specs under study.
         Sub-agents: N. Search roots: <roots>.
         Focus: find code relevant to the specs listed above.
         Advance when done.
```

## IMPLEMENT Output Updates

Change `Spec:` → `Specs:` (multiple lines if multiple), `Ref:` → `Refs:` (multiple lines):

```
Specs:   service-configuration.md#interface-outputs
         config-validation.md#behavior-strict-mode
Refs:    notes/config.md
```

When only one spec/ref, still use `Specs:` / `Refs:` for consistency.

## EVALUATE (implementing) Output Updates

Same Specs/Refs change in items list within eval output.

## ORIENT (implementing) Output Updates

After COMMIT (more items in layer):
```
State:    ORIENT
Phase:    implementing
Layer:    L0 Foundation
Progress: 2/4 items passed
Next:     2 unblocked items in next batch
Action:   STOP please review and discuss with user before continuing.
          After completion of the above, advance to select next batch.
```

After COMMIT (layer complete, more layers):
```
State:    ORIENT
Phase:    implementing
Layer:    L0 Foundation
Progress: 3/3 items passed — layer complete
Next:     L1 Core — 2 items: [daemon.types], [daemon.io]
Action:   STOP please review and discuss with user before continuing.
          After completion of the above, advance to next layer.
```

After COMMIT (last layer complete):
```
State:    ORIENT
Phase:    implementing
Layer:    L1 Core
Progress: 2/2 items passed — layer complete (final layer)
Action:   STOP please review and discuss with user before continuing.
          After completion of the above, advance to continue.
```

## DONE (implementing) Variations

When plans remain (interleaved: planning queue non-empty; all-first: impl plan queue non-empty):
```
State:   DONE
Phase:   implementing
Domain:  launcher
Summary:
  L0 Foundation:  3/3 passed
  L1 Core:        2/2 passed
  Total:          5/5 items passed
  Eval rounds:    7 across 3 batches
Action:  Domain complete. Advance to continue to next domain.
```

When no plans remain (session complete):
```
State:   DONE
Phase:   implementing
Summary:
  launcher:  5/5 items passed (3 batches)
  portal:    3/3 items passed (2 batches)
  Total:     8/8 items passed
Action:  All items complete. Session done.
```

The summary when no plans remain aggregates across all completed implementing domains. This requires tracking completed implementing domains in state.

## COMMIT Output Variations

When `enable_commits: true`:
```
State:   COMMIT
Phase:   implementing
Layer:   L0 Foundation
Batch:   1/2
Items:
  - [config.types] passed
  - [config.load] passed
Action:  Advance with --message "your commit message" to commit and continue.
```

When `enable_commits: false`:
```
State:   COMMIT
Phase:   implementing
Layer:   L0 Foundation
Batch:   1/2
Items:
  - [config.types] passed
  - [config.load] passed
Action:  Advance to continue.
```

## COMPLETE (specifying) Variations

When `enable_commits: true`:
```
State:   COMPLETE
Phase:   specifying
Specs:   5 completed, reconciled
Action:  Specifying phase complete.
         Advance with --message "your commit message" to commit and continue.
```

When `enable_commits: false`:
```
State:   COMPLETE
Phase:   specifying
Specs:   5 completed, reconciled
Action:  Specifying phase complete. Advance to continue.
```

## ACCEPT (planning) Variations

When `enable_commits: true`:
```
State:   ACCEPT
Phase:   planning
Plan:    <plan name>
Domain:  <domain>
File:    <file>
Round:   N/N
Action:  Plan accepted. Advance with --message "your commit message" to commit and continue.
```

When `enable_commits: false`:
```
State:   ACCEPT
Phase:   planning
Plan:    <plan name>
Domain:  <domain>
File:    <file>
Round:   N/N
Action:  Plan accepted. Advance to continue.
```

## EVALUATE Output Variations (enable_eval_output)

When `enable_eval_output: true`:
```
Action:  Please spawn 1 opus sub-agent to evaluate the plan.
         Sub-agent runs: forgectl eval
         After completion of the above, advance with --verdict PASS|FAIL --eval-report <path>
```

When `enable_eval_output: false`:
```
Action:  Please spawn 1 opus sub-agent to evaluate the plan.
         Sub-agent runs: forgectl eval
         After completion of the above, advance with --verdict PASS|FAIL
```

Same pattern applies in specifying EVALUATE, CROSS_REFERENCE_EVAL, RECONCILE_EVAL, implementing EVALUATE.

## REFINE Output Updates

Style change for REFINE action messages. When eval output enabled:
```
Action:  Study the eval file "<path>"
         and implement any corrections as needed.
         Apply "fresh" eyes and a tightened lens when reviewing the work,
         then apply corrections as needed.
         After completion of the above, advance to continue.
```

When eval output disabled:
```
Action:  Make corrections based off communication with the evaluator.
         Implement any corrections as needed.
         Apply "fresh" eyes and a tightened lens when reviewing the work,
         then apply corrections as needed.
         After completion of the above, advance to continue.
```

## Eval Command Output (enable_eval_output)

The `forgectl eval` command output varies by enable_eval_output:

When `enable_eval_output: true`: include `--- REPORT OUTPUT ---` section with path.
When `enable_eval_output: false`: omit `--- REPORT OUTPUT ---` section entirely.

## RECONCILE_EVAL Eval Output (new)

```
=== RECONCILIATION EVALUATION ROUND N/N ===

--- EVALUATOR INSTRUCTIONS ---

<contents of evaluators/reconcile-eval.md>

--- DOMAINS ---

optimizer: 3 specs
portal:    2 specs

--- RECONCILIATION CONTEXT ---

Run: git diff --staged

--- REPORT OUTPUT ---  (only when enable_eval_output: true)

Write your evaluation report to:
  specs/.eval/reconciliation-rN.md
```

## Status --verbose Output

```
Session: .forgectl/state/forgectl-state.json
Phase:   specifying
State:   REFINE
Config:  batch=3, rounds=1-3, guided=true
...

--- Verbose Output ---

Queue:
  [1] repository-loading.md (optimizer)
  [2] snapshot-diffing.md (optimizer)

Completed:
  [1] repository-loading.md — 2 rounds, PASS
      Evals: r1: FAIL, r2: PASS
      Commit: abc1234

Reconciliation: (none yet)
```

Per-item detail in --verbose shows: spec name, domain, rounds, eval history, commit hashes.
