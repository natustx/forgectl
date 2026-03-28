# Spec Reconciliation

## Topic of Concern
> The scaffold cross-validates completed specs for dependency and integration consistency.

## Context

After all individual specs are completed (DONE state from spec-lifecycle), the scaffold enters a reconciliation phase. This phase cross-validates dependencies and integration points across all specs to ensure they are symmetric, consistent, and free of circular references. A sub-agent evaluates the reconciliation changes, with a review loop for handling failures.

## Depends On
- **spec-lifecycle** — provides the completed specs list; DONE state triggers reconciliation.
- **state-persistence** — reads and writes the state file.

## Integration Points

| Spec | Relationship |
|------|-------------|
| phase-transitions | COMPLETE state transitions to PHASE_SHIFT (specifying → planning) |
| Reconciliation eval sub-agent | Runs `git diff --staged` and evaluates consistency across all specs |

---

## Interface

### Inputs

#### `advance` flags — Reconciliation

| State | Flags |
|-------|-------|
| RECONCILE_EVAL | `--verdict PASS\|FAIL`, `--message <text>` (required with PASS) |
| RECONCILE_REVIEW | `--verdict FAIL` (optional; no verdict = accept) |

### Outputs

#### `advance` output

**Entering RECONCILE** (after DONE):

```
State:   RECONCILE
Phase:   specifying
Domain:  optimizer
Specs:   5 completed
Action:  Cross-validate all specs: verify Depends On entries, Integration Points
         symmetry, naming consistency. Stage changes with git add.
         Advance when ready.
```

**Entering RECONCILE_EVAL** (after RECONCILE):

```
State:   RECONCILE_EVAL
Phase:   specifying
Round:   1
Action:  Tell the sub-agent to run git diff --staged and evaluate
         consistency across all specs.
         Advance with --verdict PASS --message <commit msg>
           or --verdict FAIL.
```

**Entering RECONCILE_REVIEW** (after RECONCILE_EVAL FAIL):

```
State:   RECONCILE_REVIEW
Phase:   specifying
Round:   1
Action:  Reconciliation eval found issues.
         Accept: advance (or --verdict PASS)
         Fix and re-evaluate: advance --verdict FAIL
```

**Entering COMPLETE** (after RECONCILE_EVAL PASS or RECONCILE_REVIEW accept):

```
State:   COMPLETE
Phase:   specifying
Specs:   5 completed, reconciled
Action:  Specifying phase complete. Advance to continue.
```

### Rejection

| Condition | Signal | Rationale |
|-----------|--------|-----------|
| `advance --verdict` outside of RECONCILE_EVAL or RECONCILE_REVIEW | Error naming the current state. Exit code 1. | Verdict is only valid in these states |

---

## Behavior

### State Machine

```
DONE → RECONCILE → RECONCILE_EVAL
                        │
              ┌─────────┼──────────┐
              │                    │
            PASS                 FAIL
              │                    │
              ▼                    ▼
           COMPLETE         RECONCILE_REVIEW
                              │           │
                           accept        FAIL
                              │           │
                           COMPLETE    RECONCILE
```

### Transition Table

| From State | Condition | To State | Side Effects |
|------------|-----------|----------|-------------|
| DONE | always | RECONCILE | Initialize reconcile state with round 0. |
| RECONCILE | always | RECONCILE_EVAL | Increment reconcile round. |
| RECONCILE_EVAL | `--verdict PASS` | COMPLETE | Record eval. |
| RECONCILE_EVAL | `--verdict FAIL` | RECONCILE_REVIEW | Record eval (FAIL). |
| RECONCILE_REVIEW | no verdict or `--verdict PASS` | COMPLETE | Accept. |
| RECONCILE_REVIEW | `--verdict FAIL` | RECONCILE | Grant another pass. |
| COMPLETE | always | PHASE_SHIFT | Set phase shift from specifying → planning. |

### Reconcile Evaluation Sub-Agent

The reconciliation eval differs from per-spec evals. The sub-agent:
1. Runs `git diff --staged` to see all changes
2. Reads all completed spec files
3. Checks:
   - Every `Depends On` reference points to a spec that exists
   - Every dependency has a corresponding `Integration Points` entry in the target spec
   - Integration Points are symmetric (if A lists B, B lists A)
   - Spec names are consistent across all references
   - No circular dependencies exist

---

## Invariants

1. **Reconciliation follows completion.** DONE always transitions to RECONCILE. Individual specs must be complete before cross-validation.

---

## Edge Cases

- **Scenario:** Reconciliation eval finds no issues on first pass.
  - **Expected:** RECONCILE → RECONCILE_EVAL with `--verdict PASS` → COMPLETE.
  - **Rationale:** No iteration needed when specs are already consistent. The happy path is a single round.

- **Scenario:** RECONCILE_REVIEW where the architect accepts despite FAIL verdict.
  - **Expected:** Advancing without `--verdict` (or with `--verdict PASS`) transitions to COMPLETE.
  - **Rationale:** The architect has final authority to accept reconciliation state, even if the sub-agent flagged issues.

- **Scenario:** Multiple RECONCILE → RECONCILE_EVAL → RECONCILE_REVIEW → RECONCILE cycles.
  - **Expected:** Round counter increments on each RECONCILE → RECONCILE_EVAL transition. No maximum enforced.
  - **Rationale:** Reconciliation has no max_rounds limit — unlike per-spec or plan evaluation, the architect decides when to stop.

---

## Testing Criteria

### DONE transitions to RECONCILE
- **Verifies:** Reconciliation is mandatory after spec completion.
- **Given:** All specs accepted, state is DONE.
- **When:** `advance`
- **Then:** State is RECONCILE.

### Reconcile flow PASS
- **Verifies:** Direct acceptance on PASS verdict.
- **Given:** RECONCILE_EVAL.
- **When:** `advance --verdict PASS --message "Reconcile specs"`
- **Then:** State is COMPLETE.

### Reconcile flow FAIL then fix
- **Verifies:** FAIL verdict allows re-evaluation via RECONCILE_REVIEW.
- **Given:** RECONCILE_EVAL FAIL → RECONCILE_REVIEW.
- **When:** `advance --verdict FAIL`
- **Then:** State is RECONCILE.

### COMPLETE transitions to PHASE_SHIFT
- **Verifies:** Reconciliation completion triggers phase transition.
- **Given:** COMPLETE.
- **When:** `advance`
- **Then:** State is PHASE_SHIFT.

---

## Implements
- Cross-reference reconciliation after all individual specs are complete
- Reconciliation eval/review loop with PASS/FAIL verdicts
