# Multi-Plan Phase Transitions Notes

## Config Flag

`config.planning.plan_all_before_implementing` (bool, default `false`)

## Mode: Interleaved (false, default)

Each domain is planned then implemented before the next domain begins.

### Flow
```
specifying → generate_planning_queue → planning ORIENT
  → [plan domain A] → ACCEPT
  → PHASE_SHIFT (planning → implementing) ← fires on each ACCEPT
  → implementing ORIENT → [implement domain A] → DONE
  → PHASE_SHIFT (implementing → planning)  ← fires when Planning.Queue non-empty
  → planning ORIENT → [plan domain B] → ACCEPT
  → PHASE_SHIFT (planning → implementing)
  → implementing ORIENT → [implement domain B] → DONE
  → final DONE (Planning.Queue empty, implementing.PlanQueue empty)
```

### Planning ACCEPT logic (interleaved)
- Always → PHASE_SHIFT (planning → implementing)
- `s.PhaseShift = {From: "planning", To: "implementing"}`
- No planning DONE state needed in this mode.

### Implementing DONE logic (interleaved)
- If `s.Planning.Queue` is non-empty:
  - → PHASE_SHIFT (implementing → planning)
  - `s.PhaseShift = {From: "implementing", To: "planning"}`
- If `s.Planning.Queue` is empty:
  - → final DONE (terminal state)

### PHASE_SHIFT (implementing → planning) advance
1. Pop first entry from `s.Planning.Queue` → `s.Planning.CurrentPlan`.
2. Reset `s.Planning.Round = 0`, `s.Planning.Evals = []`.
3. Set `phase = "planning"`, `state = ORIENT`.
4. Print planning ORIENT output for the new plan.

## Mode: All Planning First (true)

All domains planned, then all domains implemented.

### Flow
```
specifying → generate_planning_queue → planning ORIENT
  → [plan domain A] → ACCEPT
  → PHASE_SHIFT (planning → planning)  ← domain boundary, not last plan
  → planning ORIENT → [plan domain B] → ACCEPT
  → PHASE_SHIFT (planning → implementing)  ← last plan
  → implementing ORIENT → [implement domain A] → DONE
  → PHASE_SHIFT (implementing → implementing)  ← domain boundary, plans remain in impl queue
  → implementing ORIENT → [implement domain B] → DONE
  → final DONE
```

### Planning ACCEPT logic (all-first)
- If `s.Planning.Queue` is non-empty:
  - → PHASE_SHIFT (planning → planning) [domain boundary]
  - Move CurrentPlan to Completed. Pull next plan from Queue → CurrentPlan.
  - `s.PhaseShift = {From: "planning", To: "planning"}`
- If `s.Planning.Queue` is empty:
  - → PHASE_SHIFT (planning → implementing)
  - `s.PhaseShift = {From: "planning", To: "implementing"}`
  - Populate `s.Implementing.PlanQueue` from `s.Planning.Completed` (the remaining plans to implement sequentially).

### PHASE_SHIFT (planning → planning) advance
1. Reset Planning state for new plan (Round=0, Evals=[], CurrentPlan already set).
2. Set `state = ORIENT`.
3. Print planning ORIENT output.

### PHASE_SHIFT (planning → implementing) advance (all-first)
1. Validate and mutate plan.json (add passes/rounds).
2. Set `phase = "implementing"`, `state = ORIENT`.
3. Populate `s.Implementing.PlanQueue` from remaining completed plans (all except the first, which is now being implemented).
4. Print implementing ORIENT output.

### Implementing DONE logic (all-first)
- If `s.Implementing.PlanQueue` is non-empty:
  - → PHASE_SHIFT (implementing → implementing) [domain boundary]
  - `s.PhaseShift = {From: "implementing", To: "implementing"}`
- If `s.Implementing.PlanQueue` is empty:
  - → final DONE

### PHASE_SHIFT (implementing → implementing) advance
1. Pop first entry from `s.Implementing.PlanQueue` → current plan.
2. Read, validate, and mutate plan.json for the new plan.
3. Set `state = ORIENT`.
4. Print implementing ORIENT output.

## PHASE_SHIFT Output Messages

### planning → planning (domain boundary)
```
State:   PHASE_SHIFT
From:    planning → planning (next domain)
Completed: <domain> — <plan name> (<N> rounds)
Next:      <domain> — <plan name>

Stop and refresh your context, please.
When ready, run: forgectl advance
```

### implementing → planning
```
State:   PHASE_SHIFT
From:    implementing → planning
Completed: <domain> — N/N items passed (N batches)
Next:      <domain> — <plan name>

Stop and refresh your context, please.
When ready, run: forgectl advance
```

### implementing → implementing (domain boundary)
```
State:   PHASE_SHIFT
From:    implementing → implementing (next domain)
Completed: <domain> — N/N items passed (N batches)
Next:      <domain> — <plan name>

Stop and refresh your context, please.
When ready, run: forgectl advance
```

## State Tracking

PlanningState.Completed tracks completed plans:
```go
type CompletedPlan struct {
    ID     int    `json:"id"`
    Name   string `json:"name"`
    Domain string `json:"domain"`
    File   string `json:"file"`
}
```

ImplementingState.PlanQueue tracks remaining plans in all-first mode:
```go
PlanQueue []PlanQueueEntry `json:"plan_queue,omitempty"`
```

## Current Plan Context in Implementing

When multi-plan support is active, the implementing phase needs to know the current plan's file to reference it (for PHASE_SHIFT validation). Add to ImplementingState:
```go
CurrentPlanFile   string `json:"current_plan_file,omitempty"`
CurrentPlanDomain string `json:"current_plan_domain,omitempty"`
```
These are set when implementing begins for a new plan.
