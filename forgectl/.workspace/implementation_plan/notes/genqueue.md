# Generate Planning Queue Phase Notes

## Overview

A three-state phase between specifying and planning. Inserted into the lifecycle after specifying COMPLETE.

State machine: `ORIENT → REFINE → PHASE_SHIFT → (planning) ORIENT`

## Triggering

### From specifying PHASE_SHIFT

When `advance` is called without `--from`:
1. Set `phase` = `"generate_planning_queue"`.
2. Set `state` = `ORIENT`.
3. Run auto-generation (see below).
4. Print ORIENT output.

When `advance` is called with `--from <file>`:
1. Validate plan queue at `--from`.
2. If invalid: print errors. State remains PHASE_SHIFT.
3. If valid: populate plans queue, set `phase = "planning"`, `state = ORIENT`, pull first plan. Skip generate_planning_queue entirely.

## ORIENT: Auto-Generation Algorithm

1. Collect `s.Specifying.Completed` specs (all completed specs across all domains).
2. Group by domain. Preserve order: the first time each domain appears in the completed list determines domain order.
3. For each domain, look up code_search_roots:
   - Check `s.Specifying` for set-roots data per domain (stored via set-roots command).
   - If set: use configured code_search_roots.
   - If not set: default to `["<domain>/"]`.
4. For each domain, produce a PlanQueueEntry:
   - `name`: `"<DomainCapitalized> Implementation Plan"` (capitalize first letter of domain name).
   - `domain`: domain name.
   - `file`: `<domain>/<workspace_dir>/implementation_plan/plan.json` (workspace_dir from config.paths.workspace_dir, default ".forge_workspace").
   - `specs`: all `file` paths from completed specs for this domain.
   - `spec_commits`: deduplicated union of all `commit_hashes` from all completed specs for this domain.
   - `code_search_roots`: as determined above.
5. Write PlanQueueInput as JSON to `<state_dir>/plan-queue.json`.
6. Store the path in `s.GeneratePlanningQueue.PlanQueueFile`.
7. Print ORIENT output (see output.md).

Advance from ORIENT → REFINE.

## REFINE

The architect reviews and may edit `<state_dir>/plan-queue.json`.

On advance:
1. Read `s.GeneratePlanningQueue.PlanQueueFile`.
2. Run ValidatePlanQueue on the contents.
3. If invalid: print errors. State remains REFINE. Exit code 1.
4. If valid: set `state = PHASE_SHIFT`. Set `s.PhaseShift = {From: "generate_planning_queue", To: "planning"}`. Print PHASE_SHIFT output.

## PHASE_SHIFT (generate_planning_queue → planning)

On advance without `--from`:
1. Read `s.GeneratePlanningQueue.PlanQueueFile` (already validated at REFINE).
2. Populate `s.Planning.Queue` from the plan queue entries (all except first).
3. Set `s.Planning.CurrentPlan` = first entry.
4. Set `phase = "planning"`, `state = ORIENT`.
5. Print planning ORIENT output.

On advance with `--from <override>`:
1. Read and validate plan queue at `--from`.
2. If invalid: print errors. State remains PHASE_SHIFT.
3. If valid: same steps as without `--from`, using override entries.

## State Storage

GeneratePlanningQueueState stored in `s.GeneratePlanningQueue`:
```json
{
  "plan_queue_file": ".forgectl/state/plan-queue.json"
}
```

## set-roots Storage

The set-roots command stores code_search_roots per domain in `s.Specifying`. A new sub-struct is needed in SpecifyingState or as a map:

```go
// Add to SpecifyingState:
DomainRoots map[string][]string `json:"domain_roots,omitempty"`
// key: domain name, value: code search roots
```

The set-roots command updates this map. Auto-generation reads from it.

## SpecifyingState for set-roots

The existing spec for `set-roots` command stores roots on SpecifyingState. The auto-generation algorithm needs to read from this. Confirm the current code has this field or add it as part of this implementation.
