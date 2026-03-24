# Spec Lifecycle

## Topic of Concern
> The scaffold sequences individual spec drafting through iterative evaluation rounds.

## Context

The specifying phase guides an architect through drafting, evaluating, refining, and accepting specs from a queue. Each spec is processed one at a time through a cycle of drafting, sub-agent evaluation, and refinement until accepted (or force-accepted at max rounds). Accepted specs move to the completed list and the next spec is pulled from the queue.

This spec covers the individual spec lifecycle from ORIENT through DONE. Cross-reference reconciliation after all specs are complete is covered by spec-reconciliation.

## Depends On
- **state-persistence** — reads and writes the state file.
- **session-init** — populates the spec queue during `init --phase specifying`.

## Integration Points

| Spec | Relationship |
|------|-------------|
| Spec generation skill | The skill document describes the spec authoring process; the scaffold enforces the state machine that sequences it |
| Spec generation sub-agent | The specifying EVALUATE state is where the architect spawns a sub-agent; the scaffold tracks round count, verdict, and eval report path |
| spec-reconciliation | Receives the completed specs list when the queue is exhausted (DONE state) |
| commit-tracking | Completed specs receive commit hashes via add-commit and reconcile-commit |
| Eval output directory | The specifying eval sub-agent writes output to `<project>/specs/.eval/`; the scaffold does not read these files but the convention is documented |

---

## Interface

### Inputs

#### `advance` flags — Specifying Phase

| State | Flags |
|-------|-------|
| DRAFT | `--file <path>` (optional, override file path) |
| EVALUATE | `--verdict PASS\|FAIL`, `--eval-report <path>` (required), `--message <text>` (required with PASS) |
| REFINE | (no flags) |

The `--guided` / `--no-guided` flags are accepted on any `advance` call regardless of state.

### Outputs

#### `advance` output

**Entering SELECT** (after ORIENT):

```
State:   SELECT
Phase:   specifying
ID:      1
Spec:    Repository Loading
Domain:  optimizer
File:    optimizer/specs/repository-loading.md
Topic:   The optimizer clones or locates a repository and provides its path for downstream modules
Sources: .workspace/planning/optimizer/repo-snapshot-loading.md
Action:  Review topic and planning sources.
         Stop and review and discuss with user before continuing.
         Advance to begin drafting.
```

**Entering DRAFT** (after SELECT):

```
State:   DRAFT
Phase:   specifying
ID:      1
Spec:    Repository Loading
Domain:  optimizer
File:    optimizer/specs/repository-loading.md
Action:  Draft the spec. Advance when ready.
         Use --file <path> if the file path changed.
```

**Entering EVALUATE** (after DRAFT or REFINE):

```
State:   EVALUATE
Phase:   specifying
ID:      1
Spec:    Repository Loading
Domain:  optimizer
File:    optimizer/specs/repository-loading.md
Round:   1/3
Action:  Spawn evaluation sub-agent against the spec.
         Eval output: optimizer/specs/.eval/repository-loading-r1.md
         Advance with --verdict PASS --eval-report <path> --message <commit msg>
           or --verdict FAIL --eval-report <path>
```

**Entering REFINE** (after EVALUATE FAIL or PASS below min_rounds):

```
State:   REFINE
Phase:   specifying
ID:      1
Spec:    Repository Loading
Domain:  optimizer
File:    optimizer/specs/repository-loading.md
Round:   1/3
Action:  Read the eval report and address any findings in the spec file.
         Eval report: optimizer/specs/.eval/repository-loading-r1.md
         When changes are complete, run: forgectl advance
```

**Entering ACCEPT** (after EVALUATE PASS, round >= min_rounds):

```
State:   ACCEPT
Phase:   specifying
ID:      1
Spec:    Repository Loading
Domain:  optimizer
File:    optimizer/specs/repository-loading.md
Round:   2/3
Commit:  a1b2c3d
Action:  Spec accepted. Advance to continue.
```

**Entering DONE** (after ACCEPT, queue empty — all individual specs complete):

```
State:   DONE
Phase:   specifying
Specs:   5 completed
Action:  All individual specs complete. Advance to begin reconciliation.
```

#### Eval Report Locations

Specifying eval reports follow this convention:
```
<project>/specs/.eval/<spec-name>-rN.md
```

#### `status` output — Specifying section

```
--- Current ---

State:   REFINE
ID:      3
Spec:    Repository Loading (optimizer)
File:    optimizer/specs/repository-loading.md
Round:   1/3
Action:  Read the eval report and address any findings in the spec file.
         Eval report: optimizer/specs/.eval/repository-loading-r1.md
         When changes are complete, run: forgectl advance

--- Queue ---

  [4] Snapshot Diffing (optimizer)
  [5] Portal Rendering (portal)

--- Completed ---

  [1] Configuration Models (optimizer)  — 2 rounds, commit a1b2c3d
       Round 1: FAIL — optimizer/specs/.eval/configuration-models-r1.md
       Round 2: PASS — optimizer/specs/.eval/configuration-models-r2.md
  [2] API Gateway (api)                — 1 round, commit e4f5a6b
       Round 1: PASS — api/specs/.eval/api-gateway-r1.md
```

### Rejection

| Condition | Signal | Rationale |
|-----------|--------|-----------|
| `advance --file` outside of specifying DRAFT | Error naming the current state. Exit code 1. | Flag is meaningless outside DRAFT |
| `advance --verdict` outside of EVALUATE | Error naming the current state. Exit code 1. | Verdict is only valid in evaluation states |
| `advance` in specifying EVALUATE without `--verdict` | Error. Exit code 1. | Verdict determines the transition |
| `advance` in specifying EVALUATE without `--eval-report` | Error. Exit code 1. | Every evaluation must reference its report |
| `advance --verdict PASS` in specifying EVALUATE without `--message` | Error. Exit code 1. | Accepted specs need a commit message |
| `advance --eval-report` pointing to non-existent file | Error naming the path. Exit code 1. | Report must exist to be recorded |

---

## Behavior

### State Machine

```
ORIENT → SELECT → DRAFT → EVALUATE
                              │
                    ┌─────────┼──────────┐
                    │         │          │
              PASS ≥ min   FAIL < max  PASS < min
                    │         │          │
                    ▼         ▼          ▼
                 ACCEPT    REFINE     REFINE
                    │         │          │
              ┌─────┘         └────┬─────┘
              │                    │
         queue empty?         EVALUATE
           yes → DONE
           no → ORIENT
                              FAIL ≥ max → ACCEPT (forced)
```

### Transition Table

| From State | Condition | To State | Side Effects |
|------------|-----------|----------|-------------|
| ORIENT | always | SELECT | Pull next from queue into `current_spec` |
| SELECT | always | DRAFT | — |
| DRAFT | always | EVALUATE | If `--file` provided, override file path. Set round to 1. |
| EVALUATE | `--verdict PASS`, round >= `min_rounds` | ACCEPT | Record eval (PASS + eval report). Auto-commit with `--message`. |
| EVALUATE | `--verdict PASS`, round < `min_rounds` | REFINE | Record eval (PASS + eval report). Min rounds not met. |
| EVALUATE | `--verdict FAIL`, round < `max_rounds` | REFINE | Record eval (FAIL + eval report). |
| EVALUATE | `--verdict FAIL`, round >= `max_rounds` | ACCEPT | Record eval (FAIL + eval report). Forced acceptance. |
| REFINE | always | EVALUATE | Increment round. |
| ACCEPT | queue non-empty | ORIENT | Move spec to completed (with eval history + commit hash). |
| ACCEPT | queue empty | DONE | Move spec to completed. |

### Eval Output Convention

The specifying evaluation sub-agent writes structured markdown to a known directory:

```
<project>/specs/.eval/
├── <spec-name>-r1.md
├── <spec-name>-r2.md
└── ...
```

The scaffold does not read or write these files. This is a convention for the architect and sub-agent.

---

## Invariants

1. **Single active spec.** At most one spec in `current_spec` at any time.
2. **Round monotonicity.** The specifying round counter only increments.
3. **Queue order preserved.** Specs are pulled from the front of the queue.
4. **Min rounds enforced.** PASS below `min_rounds` forces another cycle.
5. **Max rounds enforced.** FAIL at `max_rounds` forces acceptance.
6. **Guided pauses.** When `user_guided` is true, SELECT output includes "Stop and review and discuss with user before continuing."

---

## Edge Cases

- **Scenario:** `advance --verdict FAIL` when round < `max_rounds`.
  - **Expected:** REFINE.
  - **Rationale:** More evaluation rounds remain; the architect gets another chance to address deficiencies.

- **Scenario:** `advance --verdict FAIL` when round >= `max_rounds`.
  - **Expected:** ACCEPT (forced).
  - **Rationale:** The maximum rounds are exhausted. The spec is accepted as-is to prevent indefinite loops.

- **Scenario:** `advance --verdict PASS` when round < `min_rounds`.
  - **Expected:** REFINE (min rounds not met).
  - **Rationale:** Even with a passing verdict, minimum evaluation rounds must be completed to ensure sufficient review.

---

## Testing Criteria

### Study and draft advance sequentially
- **Verifies:** Sequential state progression through specifying states.
- **Given:** ORIENT.
- **When:** advance through SELECT → DRAFT → EVALUATE.
- **Then:** Each transitions in order.

### FAIL below max_rounds goes to REFINE
- **Verifies:** FAIL verdict with remaining rounds triggers refinement.
- **Given:** EVALUATE, max_rounds: 3, round 1.
- **When:** `advance --verdict FAIL --eval-report .eval/spec-r1.md`
- **Then:** State is REFINE.

### FAIL at max_rounds forces ACCEPT
- **Verifies:** FAIL verdict at max rounds forces acceptance.
- **Given:** EVALUATE, max_rounds: 2, round 2.
- **When:** `advance --verdict FAIL --eval-report .eval/spec-r2.md`
- **Then:** State is ACCEPT (forced).

### PASS below min_rounds goes to REFINE
- **Verifies:** PASS verdict below min rounds requires more evaluation.
- **Given:** EVALUATE, min_rounds: 2, round 1.
- **When:** `advance --verdict PASS --eval-report .eval/spec-r1.md`
- **Then:** State is REFINE.

### PASS at min_rounds goes to ACCEPT
- **Verifies:** PASS verdict at min rounds triggers acceptance with commit.
- **Given:** EVALUATE, min_rounds: 2, round 2.
- **When:** `advance --verdict PASS --eval-report .eval/spec-r2.md --message "Add spec"`
- **Then:** State is ACCEPT. Commit created.

### PASS requires message
- **Verifies:** PASS verdict rejection without commit message.
- **Given:** EVALUATE.
- **When:** `advance --verdict PASS --eval-report .eval/spec-r1.md` without `--message`
- **Then:** Exit code 1.

### DONE transitions to reconciliation
- **Verifies:** Queue exhaustion triggers reconciliation phase.
- **Given:** All specs accepted, state is DONE.
- **When:** `advance`
- **Then:** State is RECONCILE (see spec-reconciliation).

---

## Implements
- Specifying phase: queue-driven spec drafting with eval/refine loop
- Single active spec with sequential queue processing
- Eval round enforcement (min/max) with forced acceptance
