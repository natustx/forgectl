# Phase Transitions

## Topic of Concern
> The scaffold enforces explicit context-refresh checkpoints between lifecycle phases.

## Context

The forgectl scaffold is a Go CLI tool (built with Cobra) that manages the full software development lifecycle through three sequential phases — specifying, planning, and implementing — backed by a single JSON state file (`forgectl-state.json`). State names (ORIENT, EVALUATE, etc.) are reused across phases with phase-specific behavior; the `phase` field determines which states are valid and how they behave.

Between phases, a PHASE_SHIFT state acts as a hard stop — the user is told to refresh their context before proceeding. This prevents stale context from carrying over between fundamentally different activities (specifying → planning → implementing).

Each phase shift has distinct mechanics: specifying→planning requires a new input file (`--from`), while planning→implementing validates and mutates the existing plan.json. The scaffold can be initialized at any phase, allowing users to skip earlier phases when inputs already exist.

## Depends On
- **spec-reconciliation** — COMPLETE triggers the specifying→planning phase shift.
- **plan-production** — ACCEPT triggers the planning→implementing phase shift.
- **state-persistence** — reads and writes the state file.

## Integration Points

| Spec | Relationship |
|------|-------------|
| spec-reconciliation | COMPLETE → PHASE_SHIFT (specifying → planning) |
| plan-production | Receives the plans queue when specifying→planning advances; ACCEPT → PHASE_SHIFT (planning → implementing) |
| batch-implementation | Receives validated plan.json when planning→implementing advances |
| session-init | Plan queue schema (same validation) reused at specifying→planning shift |

---

## Interface

### Inputs

#### `advance` flags at PHASE_SHIFT

| Phase Shift | Flags |
|-------------|-------|
| specifying → planning | `--from <path>` (required, plan queue JSON) |
| planning → implementing | (no additional flags) |

The `--guided` / `--no-guided` flags are accepted at phase shifts and update `user_guided` before the transition proceeds.

### Outputs

#### `advance` output

**Entering PHASE_SHIFT** (specifying → planning):

```
State:   PHASE_SHIFT
From:    specifying → planning

Stop and refresh your context, please.
When ready, run: forgectl advance --from <plans-queue.json>
```

**Entering PHASE_SHIFT** (planning → implementing):

```
State:   PHASE_SHIFT
From:    planning → implementing
Plan:    Service Configuration
Domain:  launcher
File:    launcher/.workspace/implementation_plan/plan.json

Stop and refresh your context, please.
When ready, run: forgectl advance
```

### Rejection

| Condition | Signal | Rationale |
|-----------|--------|-----------|
| `advance` at PHASE_SHIFT (specifying→planning) without `--from` | Error: "--from <plans-queue.json> is required at this phase shift." Exit code 1. | Planning queue must be provided |

---

## Behavior

### Specifying → Planning

#### Entering PHASE_SHIFT
When the architect advances from COMPLETE (spec-reconciliation), the scaffold transitions to PHASE_SHIFT. No work is done — the scaffold prints the phase shift message and waits.

#### Advancing from PHASE_SHIFT
The user must provide `--from <plans-queue.json>` with a plan queue input file.

1. Read and validate the plans queue at `--from`.
2. If validation fails: print errors. State remains PHASE_SHIFT.
3. If validation passes:
   - Populate the plans queue in state.
   - Set `phase` to `"planning"`.
   - Set `state` to `ORIENT`.
   - Pull first plan from queue.
   - Print the ORIENT action description.

### Planning → Implementing

#### Entering PHASE_SHIFT
When the architect advances from planning ACCEPT (plan-production), the scaffold transitions to PHASE_SHIFT.

#### Advancing from PHASE_SHIFT
No `--from` needed — the plan.json path is already known from `current_plan.file`.

1. Read plan.json at `current_plan.file`.
2. Validate the plan structure (same checks as the planning validation gate in plan-production).
3. If validation fails: print errors. State remains PHASE_SHIFT.
4. If validation passes:
   - Add `passes: "pending"` and `rounds: 0` to every item.
   - Write the updated plan.json.
   - Set `phase` to `"implementing"`.
   - Set `state` to `ORIENT`.
   - Print the ORIENT action description with initialization summary.

---

## Invariants

1. **Phase shifts are explicit.** The scaffold always stops at PHASE_SHIFT between phases. It never transitions directly across phase boundaries.
2. **specifying→planning requires input.** `--from` with a plans queue is required at this phase shift.
3. **planning→implementing validates.** plan.json is validated and mutated on advance out of PHASE_SHIFT.
4. **Guided setting is mutable.** `--guided` / `--no-guided` can change `user_guided` on any advance call, including at phase shifts.

---

## Edge Cases

- **Scenario:** Advance from specifying→planning PHASE_SHIFT without `--from`.
  - **Expected:** Error. State remains PHASE_SHIFT.
  - **Rationale:** The planning phase requires a plans queue; the scaffold cannot proceed without input.

- **Scenario:** Advance from specifying→planning PHASE_SHIFT with invalid plans queue.
  - **Expected:** Validation errors printed. State remains PHASE_SHIFT.
  - **Rationale:** Invalid input is rejected at the boundary rather than allowing a corrupt queue into the planning phase.

- **Scenario:** Advance from planning→implementing PHASE_SHIFT with invalid plan.json.
  - **Expected:** Validation errors printed. State remains PHASE_SHIFT.
  - **Rationale:** The implementing phase requires a structurally valid plan; errors caught here prevent failures during batch selection.

- **Scenario:** `--guided` provided at PHASE_SHIFT advance.
  - **Expected:** `user_guided` updated before the phase transition proceeds.
  - **Rationale:** The guided setting is mutable on any advance, including phase shifts, so users can change behavior at natural boundaries.

---

## Testing Criteria

### specifying→planning requires --from
- **Verifies:** Missing input rejection at specifying→planning boundary.
- **Given:** PHASE_SHIFT (specifying→planning).
- **When:** `advance` without `--from`
- **Then:** Exit code 1.

### specifying→planning with valid queue
- **Verifies:** Successful phase transition with valid input.
- **Given:** PHASE_SHIFT (specifying→planning).
- **When:** `advance --from plans-queue.json`
- **Then:** `phase: "planning"`, `state: "ORIENT"`.

### specifying→planning with invalid queue
- **Verifies:** Validation failure preserves state.
- **Given:** PHASE_SHIFT (specifying→planning).
- **When:** `advance --from invalid.json`
- **Then:** Errors printed. State remains PHASE_SHIFT.

### planning→implementing with valid plan
- **Verifies:** Successful phase transition with plan.json mutation.
- **Given:** PHASE_SHIFT (planning→implementing).
- **When:** `advance`
- **Then:** `phase: "implementing"`, `state: "ORIENT"`. plan.json items mutated.

### planning→implementing with invalid plan
- **Verifies:** Validation failure preserves state.
- **Given:** PHASE_SHIFT (planning→implementing), plan.json has errors.
- **When:** `advance`
- **Then:** Errors printed. State remains PHASE_SHIFT.

### --guided at phase shift
- **Verifies:** Guided setting mutation at phase boundaries.
- **Given:** PHASE_SHIFT, user_guided is true.
- **When:** `advance --no-guided --from plans-queue.json`
- **Then:** `user_guided: false` after transition.

### Full Lifecycle: Specifying only
- **Verifies:** Complete specifying phase ending at phase shift boundary.
- **Given:** Init with 2 specs, `--batch-size 2 --min-rounds 1 --max-rounds 3`.
- **When:** Complete both specs → DONE → RECONCILE → RECONCILE_EVAL(PASS) → COMPLETE → PHASE_SHIFT.
- **Then:** State is PHASE_SHIFT. Specifying completed with 2 specs.

### Full Lifecycle: Three-phase
- **Verifies:** End-to-end lifecycle across all three phases.
- **Given:** Init with specs queue.
- **When:** Specifying: draft + accept 2 specs, reconcile → PHASE_SHIFT → Planning: study, draft plan, evaluate → PHASE_SHIFT → Implementing: batch items, evaluate → DONE.
- **Then:** State is DONE. All three phase sections populated in state file.

### Full Lifecycle: Start at planning
- **Verifies:** Mid-entry lifecycle starting at planning phase.
- **Given:** `init --phase planning --from plans-queue.json`.
- **When:** Planning → PHASE_SHIFT → Implementing → DONE.
- **Then:** Specifying section is null. Planning and implementing complete.

### Full Lifecycle: Start at implementing
- **Verifies:** Single-phase lifecycle starting at implementing.
- **Given:** `init --phase implementing --from plan.json`.
- **When:** Implementing → DONE.
- **Then:** Specifying and planning sections are null. Implementing complete.

---

## Implements
- Explicit PHASE_SHIFT checkpoints between phases with context refresh
- Phase shift input injection (`--from` at specifying→planning boundary)
- Plan validation and mutation at planning→implementing boundary
- Full lifecycle integration across phase boundaries
