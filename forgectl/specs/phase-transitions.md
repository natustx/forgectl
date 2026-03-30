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
| specifying → planning | `--from <path>` (optional, plan queue JSON override). If omitted, forgectl auto-generates the plan queue from completed specs. |
| planning → implementing | (no additional flags) |

The `--guided` / `--no-guided` flags are accepted at phase shifts and update `config.general.user_guided` before the transition proceeds.

### Outputs

#### `advance` output

**Entering PHASE_SHIFT** (specifying → planning):

```
State:   PHASE_SHIFT
From:    specifying → planning

Domains:  2 (optimizer, portal)
Specs:    5 completed
Roots:    optimizer → optimizer/, lib/shared/
          portal → portal/ (default)

Stop and refresh your context, please.
When ready, run:
  forgectl advance                          # auto-generate plan queue from completed specs
  forgectl advance --from <plan-queue.json> # OR provide a custom plan queue
```

**Entering PHASE_SHIFT** (planning → implementing):

```
State:   PHASE_SHIFT
From:    planning → implementing
Plan:    Service Configuration
Domain:  launcher
File:    launcher/.forge_workspace/implementation_plan/plan.json

Stop and refresh your context, please.
When ready, run: forgectl advance
```

### Rejection

| Condition | Signal | Rationale |
|-----------|--------|-----------|
| `advance` at PHASE_SHIFT (specifying→planning) with `--from` pointing to invalid plan queue | Validation errors printed. State remains PHASE_SHIFT. Exit code 1. | Override file must be valid |

---

## Behavior

### Specifying → Planning

#### Entering PHASE_SHIFT
When the architect advances from COMPLETE (spec-reconciliation), the scaffold transitions to PHASE_SHIFT. No work is done — the scaffold prints the phase shift message and waits.

#### Advancing from PHASE_SHIFT

If `--from` is provided, use the external plan queue file (override mode). If `--from` is omitted, auto-generate the plan queue from the specifying phase data.

**Auto-generation (no `--from`):**

1. Group completed specs by domain.
2. For each domain, produce a plan entry:
   - `name`: `"<Domain> Implementation Plan"` (domain name capitalized).
   - `domain`: the domain name.
   - `file`: `<domain>/.forge_workspace/implementation_plan/plan.json`.
   - `specs`: all completed spec file paths for this domain.
   - `spec_commits`: deduplicated list of all `commit_hashes` from the domain's completed specs.
   - `code_search_roots`: from `specifying.domains[<domain>].code_search_roots` if set via `set-roots`, otherwise `["<domain>/"]`.
3. Validate the generated plan queue (same checks as external input).
4. If validation passes:
   - Populate the plans queue in state.
   - Set `phase` to `"planning"`.
   - Set `state` to `ORIENT`.
   - Pull first plan from queue.
   - Print the ORIENT action description.

**Override mode (`--from` provided):**

1. Read and validate the plans queue at `--from`.
2. If validation fails: print errors. State remains PHASE_SHIFT.
3. If validation passes: same steps as auto-generation step 4.

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
2. **specifying→planning has input.** The plan queue is either auto-generated from completed specs or provided via `--from` override.
3. **planning→implementing validates.** plan.json is validated and mutated on advance out of PHASE_SHIFT.
4. **Guided setting is mutable.** `--guided` / `--no-guided` can change `config.general.user_guided` on any advance call, including at phase shifts.

---

## Edge Cases

- **Scenario:** Advance from specifying→planning PHASE_SHIFT without `--from`.
  - **Expected:** Plan queue auto-generated from completed specs. Phase transitions to planning.
  - **Rationale:** Completed specs, commit hashes, and code search roots provide all data needed to generate the plan queue.

- **Scenario:** Advance from specifying→planning PHASE_SHIFT without `--from`, domain has no `set-roots`.
  - **Expected:** `code_search_roots` defaults to `["<domain>/"]` for that domain.
  - **Rationale:** The domain directory itself is the most common search root.

- **Scenario:** Advance from specifying→planning PHASE_SHIFT with `--from` override.
  - **Expected:** External file used instead of auto-generation. Validated and consumed.
  - **Rationale:** Override allows custom plan queue when auto-generation doesn't match needs.

- **Scenario:** Advance from specifying→planning PHASE_SHIFT with invalid plans queue (`--from`).
  - **Expected:** Validation errors printed. State remains PHASE_SHIFT.
  - **Rationale:** Invalid input is rejected at the boundary rather than allowing a corrupt queue into the planning phase.

- **Scenario:** Advance from planning→implementing PHASE_SHIFT with invalid plan.json.
  - **Expected:** Validation errors printed. State remains PHASE_SHIFT.
  - **Rationale:** The implementing phase requires a structurally valid plan; errors caught here prevent failures during batch selection.

- **Scenario:** `--guided` provided at PHASE_SHIFT advance.
  - **Expected:** `config.general.user_guided` updated before the phase transition proceeds.
  - **Rationale:** The guided setting is mutable on any advance, including phase shifts, so users can change behavior at natural boundaries.

---

## Testing Criteria

### specifying→planning auto-generates plan queue
- **Verifies:** Plan queue auto-generation from completed specs.
- **Given:** PHASE_SHIFT (specifying→planning). 2 domains (optimizer with 3 specs, portal with 2 specs). Optimizer has `set-roots` configured. Portal does not.
- **When:** `advance` (no `--from`)
- **Then:** `phase: "planning"`, `state: "ORIENT"`. Plan queue has 2 entries. Optimizer entry has custom code_search_roots. Portal entry has `["portal/"]` default.

### specifying→planning auto-generation includes spec_commits
- **Verifies:** Commit hashes flow from completed specs to plan queue.
- **Given:** PHASE_SHIFT (specifying→planning). Completed specs have commit_hashes `["7cede10", "8743b1d"]` across the domain.
- **When:** `advance` (no `--from`)
- **Then:** Generated plan entry has `spec_commits: ["7cede10", "8743b1d"]` (deduplicated).

### specifying→planning with --from override
- **Verifies:** External plan queue overrides auto-generation.
- **Given:** PHASE_SHIFT (specifying→planning).
- **When:** `advance --from custom-queue.json`
- **Then:** `phase: "planning"`, `state: "ORIENT"`. Plan queue matches the external file, not auto-generated.

### specifying→planning with invalid --from override
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
- **Given:** PHASE_SHIFT, `config.general.user_guided` is true.
- **When:** `advance --no-guided --from plans-queue.json`
- **Then:** `config.general.user_guided: false` after transition.

### Full Lifecycle: Specifying only
- **Verifies:** Complete specifying phase ending at phase shift boundary.
- **Given:** Init with 2 specs, `specifying.batch: 2, specifying.eval.min_rounds: 1, specifying.eval.max_rounds: 3`.
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
- Auto-generated plan queue at specifying→planning boundary from completed specs, commit hashes, and code search roots
- Optional `--from` override for custom plan queue at specifying→planning boundary
- Plan validation and mutation at planning→implementing boundary
- Full lifecycle integration across phase boundaries
