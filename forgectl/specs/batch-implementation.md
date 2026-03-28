# Batch Implementation

## Topic of Concern
> The scaffold implements plan items in dependency-ordered batches through iterative evaluation.

## Context

The implementing phase delivers plan items one at a time within dependency-ordered batches. After each batch is fully implemented, an evaluation sub-agent verifies against acceptance criteria through iterative rounds. Items progress through `pending` → `done` → `passed`/`failed` states tracked in plan.json.

Layers enforce a coarse ordering (all layer N items must be terminal before layer N+1), while `depends_on` within a layer provides fine-grained ordering. Batches are groups of up to `batch_size` unblocked items drawn from the current layer.

## Depends On
- **phase-transitions** — the planning→implementing phase shift validates plan.json, adds tracking fields, and triggers ORIENT.
- **session-init** — alternatively, `init --phase implementing` starts the implementing phase directly.
- **state-persistence** — reads and writes the state file.

## Integration Points

| Spec | Relationship |
|------|-------------|
| plan.json | Planning produces it; implementing consumes and mutates it (adding `passes` and `rounds` fields) |
| Implementation evaluator prompt (`evaluators/impl-eval.md`) | Full instructions for the implementation evaluation sub-agent: what to check, report format, verdict rules |

---

## Interface

### Inputs

#### `advance` flags — Implementing Phase

| State | Flags |
|-------|-------|
| IMPLEMENT | `--message <text>` (required, first round only) |
| EVALUATE | `--verdict PASS\|FAIL`, `--eval-report <path>` (both required) |
| COMMIT | `--message <text>` (required) |

#### `eval` command

| Command | Flags | Description |
|---------|-------|-------------|
| `eval` | none | Output full evaluation context for the sub-agent. Only valid in implementing EVALUATE. |

### Outputs

#### `advance` output

**Entering ORIENT** (after PHASE_SHIFT or init with `--phase implementing`):

```
State:   ORIENT
Phase:   implementing
Plan:    Service Configuration
Domain:  launcher
File:    launcher/.workspace/implementation_plan/plan.json
Config:  batch_size=2, rounds=1-3

Initialized plan.json for implementation:
  Items:  5 (passes: pending, rounds: 0)
  Layers: 2 (L0 Foundation: 3 items, L1 Core: 2 items)

Action:  Stop and review and discuss with user before continuing.
         Selecting first batch. Run: forgectl advance
```

**Entering IMPLEMENT** (first item in batch, first round — no prior eval):

```
State:   IMPLEMENT
Phase:   implementing
Layer:   L0 Foundation
Batch:   1/2
Item:    [config.types] ServiceEndpoint and ServicesConfig structs
         Go structs for validated service endpoint configuration.
         (1 of 2 in batch)
Steps:
  1. Define ServiceEndpoint struct with Host (string) and Port (int) fields
  2. Define ServicesConfig struct with three named ServiceEndpoint fields
  3. Add YAML struct tags for deserialization
Files:   internal/config/types.go
Spec:    service-configuration.md#interface-outputs
Ref:     notes/config.md#types
Tests:   1 functional
Action:  Implement this item.
         When complete, run: forgectl advance --message <commit msg>
```

**Entering IMPLEMENT** (next item in same batch, first round):

```
State:   IMPLEMENT
Phase:   implementing
Layer:   L0 Foundation
Batch:   1/2
Item:    [config.load] Load YAML, apply defaults, validate strictly
         Parse spectacular.yml, apply default host/port values.
         (2 of 2 in batch)
Steps:
  1. Implement LoadConfig() using goccy/go-yaml strict mode
  2. Add default port logic (portal=8080, api=8081, optimizer=8082)
  3. Add post-unmarshal validation for port range and empty host
  4. Write table-driven tests for valid, rejection, and edge cases
Files:   internal/config/load.go, internal/config/load_test.go
Spec:    service-configuration.md#behavior-loading
Ref:     notes/config.md#load
Tests:   2 functional, 2 rejection, 2 edge_case
Action:  Implement this item.
         When complete, run: forgectl advance --message <commit msg>
```

**Entering IMPLEMENT** (first item in batch, after eval — round 2+):

```
State:   IMPLEMENT
Phase:   implementing
Layer:   L0 Foundation
Batch:   1/2
Round:   1/3
Eval:    launcher/.workspace/implementation_plan/evals/batch-1-round-1.md
Note:    PASS recorded for round 1. Minimum rounds not yet met (1/2).
Item:    [config.types] ServiceEndpoint and ServicesConfig structs
         Go structs for validated service endpoint configuration.
         (1 of 2 in batch)
Steps:
  1. Define ServiceEndpoint struct with Host (string) and Port (int) fields
  2. Define ServicesConfig struct with three named ServiceEndpoint fields
  3. Add YAML struct tags for deserialization
Files:   internal/config/types.go
Spec:    service-configuration.md#interface-outputs
Ref:     notes/config.md#types
Tests:   1 functional
Action:  Study the eval file "launcher/.workspace/implementation_plan/evals/batch-1-round-1.md"
         and implement any corrections as needed. If none found during the eval,
         please verify and look for corrections. Apply them.
         When complete, run: forgectl advance
```

**Entering EVALUATE** (implementing phase):

```
State:    EVALUATE
Phase:    implementing
Layer:    L0 Foundation
Batch:    1/2
Round:    1/3
Items:
  - [config.types] ServiceEndpoint and ServicesConfig structs
  - [config.load] Load YAML, apply defaults, validate strictly
Action:   Ask the evaluation sub-agent to verify batch items against their tests.
          The sub-agent should run: forgectl eval
          After reviewing the eval report, run:
            forgectl advance --eval-report <path> --verdict PASS|FAIL
Sub-agent: forgectl eval
```

**Entering COMMIT** (after EVALUATE, batch terminal):

```
State:   COMMIT
Phase:   implementing
Layer:   L0 Foundation
Batch:   1/2
Items:
  - [config.types] passed
  - [config.load] passed
Action:  Commit your changes before continuing.
         When ready, run: forgectl advance --message <commit msg>
```

**Entering COMMIT** (after force-accept):

```
State:   COMMIT
Phase:   implementing
Layer:   L1 Core
Batch:   3/3
Items:
  - [daemon.types] failed (force-accept, 3/3 rounds)
  - [daemon.io] failed (force-accept, 3/3 rounds)
Action:  Commit your changes before continuing.
         When ready, run: forgectl advance --message <commit msg>
```

**Entering ORIENT** (after COMMIT, more items in layer):

```
State:    ORIENT
Phase:    implementing
Layer:    L0 Foundation
Progress: 2/3 items passed
Action:   Stop and review and discuss with user before continuing.
          Selecting next batch. Run: forgectl advance
```

**Entering ORIENT** (after COMMIT, layer complete):

```
State:    ORIENT
Phase:    implementing
Layer:    L0 Foundation
Progress: 3/3 items passed — layer complete
Action:   Stop and review and discuss with user before continuing.
          Advancing to next layer. Run: forgectl advance
```

**Entering ORIENT** (force-accept):

```
State:    ORIENT
Phase:    implementing
Layer:    L1 Core
          FORCE ACCEPT: 2 items marked failed (max rounds 3/3 reached)
          - [daemon.types] Daemon state types and PID file struct
          - [daemon.io] PID file I/O operations
Progress: 2/2 items terminal (0 passed, 2 failed) — layer complete
Action:   Advancing to next layer. Run: forgectl advance
```

**DONE** (all items complete, terminal state):

```
State:   DONE
Phase:   implementing
Summary:
  L0 Foundation:  3/3 passed
  L1 Core:        2/2 passed
  Total:          5/5 items passed
  Eval rounds:    7 across 3 batches
Action:  All items complete. Session done.
```

#### `eval` output

```
=== IMPLEMENTATION EVALUATION ROUND 1/3 ===
Layer: L0 Foundation
Batch: 1/2

--- EVALUATOR INSTRUCTIONS ---

<contents of evaluators/impl-eval.md>

--- ITEMS TO EVALUATE ---

[1] config.types — ServiceEndpoint and ServicesConfig structs
    Description: Go structs for validated service endpoint configuration.
    Spec:        service-configuration.md#interface-outputs
    Ref:         notes/config.md#types
    Files:       internal/config/types.go
    Steps:
      1. Define ServiceEndpoint struct with Host (string) and Port (int) fields
      2. Define ServicesConfig struct with three named ServiceEndpoint fields
      3. Add YAML struct tags for deserialization
    Tests:
      [functional] Three named fields, not a map

[2] config.load — Load YAML, apply defaults, validate strictly
    Description: Parse spectacular.yml, apply default host/port values.
    Spec:        service-configuration.md#behavior-loading
    Ref:         notes/config.md#load
    Files:       internal/config/load.go, internal/config/load_test.go
    Steps:
      1. Implement LoadConfig() using goccy/go-yaml strict mode
      2. Add default port logic (portal=8080, api=8081, optimizer=8082)
      3. Add post-unmarshal validation for port range and empty host
      4. Write table-driven tests for valid, rejection, and edge cases
    Tests:
      [functional] Default ports applied when services are empty objects
      [functional] Default host applied when only port specified
      [rejection]  Missing services section rejected
      [rejection]  Port out of range rejected
      [rejection]  Unknown keys rejected
      [edge_case]  Empty file rejected
      [edge_case]  Duplicate ports allowed

--- REPORT OUTPUT ---

Write your evaluation report to:
  launcher/.workspace/implementation_plan/evals/batch-1-round-1.md
```

Subsequent rounds include previous evaluations:

```
=== IMPLEMENTATION EVALUATION ROUND 2/3 ===
...

--- PREVIOUS EVALUATIONS ---

Round 1: PASS — launcher/.workspace/implementation_plan/evals/batch-1-round-1.md

--- REPORT OUTPUT ---

Write your evaluation report to:
  launcher/.workspace/implementation_plan/evals/batch-1-round-2.md
```

#### Eval Report Locations

Implementation eval reports:
```
<domain>/.workspace/implementation_plan/evals/batch-N-round-M.md
```

#### `status` output — Implementing section

```
--- Implementing ---

  Layer L0 (Foundation): complete
    [bootstrap]     passed  (1 round)
    [config.types]  passed  (1 round)
    [config.load]   passed  (2 rounds)

  Layer L1 (Core): in progress
    [daemon.types]  done    (0 rounds)
    [daemon.io]     pending (0 rounds)
```

### Rejection

| Condition | Signal | Rationale |
|-----------|--------|-----------|
| `advance` in implementing IMPLEMENT (first round) without `--message` | Error. Exit code 1. | First-round items need a commit message |
| `advance` in implementing COMMIT without `--message` | Error. Exit code 1. | Batch completion needs a commit message |
| `advance` in implementing EVALUATE without `--verdict` | Error. Exit code 1. | Verdict determines the transition |
| `advance` in implementing EVALUATE without `--eval-report` | Error. Exit code 1. | Every evaluation must reference its report |
| `advance --eval-report` pointing to non-existent file | Error naming the path. Exit code 1. | Report must exist to be recorded |
| `eval` outside of implementing EVALUATE | Error naming current state and phase. Exit code 1. | Eval context only available in EVALUATE |

---

## Behavior

### Batch Calculation

Batches are groups of items drawn from the current layer. The scaffold selects up to `batch_size` unblocked items.

An item is **unblocked** when:
1. All items in prior layers have a terminal `passes` value (`passed` or `failed`)
2. All items in its `depends_on` list have a terminal `passes` value

Items are selected in the order they appear in the layer's `items` array.

### State Machine

```
ORIENT → IMPLEMENT(1) → IMPLEMENT(2) → ... → EVALUATE
                                                  │
                                    ┌──────────────┼──────────────┐
                                    │              │              │
                              PASS + rounds    FAIL + rounds   PASS/FAIL
                              >= min_rounds    < max_rounds    at boundary
                                    │              │              │
                                    ▼              ▼              │
                                 COMMIT      IMPLEMENT(1)→...    │
                                    │        (re-implement)      │
                                    ▼                            │
                              ORIENT/DONE ◄──────────────────────┘
                                                FAIL + rounds
                                                >= max_rounds
                                                      │
                                                      ▼
                                                   COMMIT → ORIENT/DONE
```

### Transition Table

| From State | Condition | To State | Side Effects |
|------------|-----------|----------|-------------|
| ORIENT | unblocked items exist in current layer | IMPLEMENT | Select batch. Present first item. |
| ORIENT | all layer items terminal, more layers | ORIENT (next layer) | Advance `current_layer`. |
| ORIENT | all layers complete | DONE | — |
| IMPLEMENT | more items in batch | IMPLEMENT | Mark current item `done`. Present next item. |
| IMPLEMENT | last item in batch | EVALUATE | Mark current item `done`. Increment `rounds` on all batch items. |
| EVALUATE | PASS, rounds >= min_rounds | COMMIT | Mark items `passed`. Record eval. |
| EVALUATE | PASS, rounds < min_rounds | IMPLEMENT | Record eval. Re-present first item with eval file. |
| EVALUATE | FAIL, rounds < max_rounds | IMPLEMENT | Record eval. Re-present first item with eval file. |
| EVALUATE | FAIL, rounds >= max_rounds | COMMIT | Mark items `failed`. Record eval. Force-accept. |
| COMMIT | more batches or layers | ORIENT | — |
| COMMIT | all layers complete | DONE | — |
| DONE | — | Error: "session complete." | Terminal state. |

### Item `passes` Transitions

| Event | `passes` change |
|-------|----------------|
| Engineer advances past item in IMPLEMENT | `pending` → `done` |
| EVALUATE PASS + rounds >= min_rounds | `done` → `passed` |
| EVALUATE FAIL + rounds >= max_rounds | `done` → `failed` |
| EVALUATE FAIL + rounds < max_rounds | stays `done` |
| EVALUATE PASS + rounds < min_rounds | stays `done` |

### IMPLEMENT Behavior

Presents **one item at a time**. Displays full context: name, description, steps, files, spec, ref, test summary.

**First round (no prior eval):** Action says "Implement this item." Advance requires `--message` — the scaffold commits after each item.

**Subsequent rounds (after eval):** Action says "Study the eval file and implement any corrections." No `--message` required — corrections are committed at the COMMIT state after the batch passes.

### EVALUATE Behavior

Two actors:

**Sub-agent** runs `forgectl eval` to receive full item details, evaluator prompt, report target path, and previous eval history.

**Engineer** reviews the report, runs `forgectl advance --eval-report <path> --verdict PASS|FAIL`.

### COMMIT State

Hard stop after a batch reaches terminal evaluation. Ensures all implementation work is committed before proceeding.

Appears after:
- EVALUATE with PASS + sufficient rounds
- EVALUATE with FAIL at max_rounds (force-accept)

The engineer commits changes, then runs `forgectl advance --message <commit msg>` to proceed.

---

## Invariants

1. **Layer ordering enforced.** All items in layer N must be terminal before layer N+1.
2. **Dependency ordering enforced.** Items only delivered when `depends_on` items are terminal.
3. **Item order preserved.** Items delivered in layer's `items` array order.
4. **One item at a time.** IMPLEMENT presents a single item per advance.
5. **plan.json is the progress record.** `passes` and `rounds` reflect current state.
6. **COMMIT precedes progression.** Every batch boundary passes through COMMIT before ORIENT/DONE.
7. **First-round commits.** IMPLEMENT advance requires `--message` and commits on the first round only. Subsequent rounds do not commit (corrections committed at COMMIT state).
8. **Two actors, two commands.** Engineer uses `advance`; sub-agent uses `eval`.
9. **Scaffold does not parse eval files.** Verdict provided via `--verdict`; file stored as path reference.
10. **Min rounds enforced.** PASS below `min_rounds` forces another implementation cycle.
11. **Max rounds enforced.** FAIL at `max_rounds` forces acceptance.
12. **Guided pauses.** When `user_guided` is true, ORIENT output includes "Stop and review and discuss with user before continuing."

---

## Edge Cases

- **Scenario:** Layer has fewer items than `batch_size`.
  - **Expected:** Single batch contains all items.
  - **Rationale:** Batches are capped at `batch_size` but may be smaller. No padding or splitting occurs.

- **Scenario:** Batch has one item.
  - **Expected:** IMPLEMENT → EVALUATE directly.
  - **Rationale:** Single-item batches skip the multi-item advance loop; the item is marked `done` and evaluation begins.

- **Scenario:** EVALUATE PASS but rounds < min_rounds.
  - **Expected:** Re-enter IMPLEMENT. No commit reminder (not first round).
  - **Rationale:** Minimum evaluation rounds must be met regardless of verdict. The engineer gets another pass through the items.

- **Scenario:** EVALUATE FAIL at max_rounds.
  - **Expected:** Items `failed`. COMMIT. ORIENT.
  - **Rationale:** The maximum rounds are exhausted. Items are force-accepted as failed to prevent indefinite loops.

- **Scenario:** Item depends on a `failed` item.
  - **Expected:** Still unblocked — `failed` is terminal.
  - **Rationale:** `failed` is a terminal state just like `passed`. Dependent items proceed regardless of whether dependencies passed or failed.

- **Scenario:** All layers complete.
  - **Expected:** COMMIT → DONE.
  - **Rationale:** DONE is the terminal state; no more batches or layers to process.

- **Scenario:** `eval` called outside EVALUATE.
  - **Expected:** Error.
  - **Rationale:** Evaluation context is only meaningful in EVALUATE state; the sub-agent has nothing to evaluate otherwise.

---

## Testing Criteria

### ORIENT selects first batch
- **Verifies:** Batch selection from layer items.
- **Given:** ORIENT (implementing), L0 has 4 items, batch_size 2.
- **When:** `advance`
- **Then:** State is IMPLEMENT. First item presented.

### IMPLEMENT presents items one at a time
- **Verifies:** Single-item presentation with batch progression.
- **Given:** IMPLEMENT, batch has 2 items, on item 1.
- **When:** `advance --message "Implement config types"`
- **Then:** Item 1 `done`. Item 2 presented.

### IMPLEMENT last item → EVALUATE
- **Verifies:** Last item triggers evaluation.
- **Given:** IMPLEMENT, last item in batch.
- **When:** `advance --message "Implement config load"`
- **Then:** Item `done`. Rounds incremented. State is EVALUATE.

### First-round IMPLEMENT requires --message
- **Verifies:** Commit message required on first round.
- **Given:** IMPLEMENT, first round (no prior eval).
- **When:** `advance` without `--message`
- **Then:** Exit code 1.

### First-round IMPLEMENT commits with --message
- **Verifies:** Scaffold commits after first-round items.
- **Given:** IMPLEMENT, first round (no prior eval).
- **When:** `advance --message "Implement config types"`
- **Then:** Scaffold commits. Next item presented (or EVALUATE if last).

### Subsequent-round IMPLEMENT does not require --message
- **Verifies:** No commit on subsequent rounds.
- **Given:** IMPLEMENT, entered after EVALUATE (round 2+).
- **When:** `advance`
- **Then:** Advances without committing. No error.

### EVALUATE PASS with sufficient rounds → COMMIT
- **Verifies:** PASS with sufficient rounds marks items passed.
- **Given:** EVALUATE, rounds >= min_rounds.
- **When:** `advance --eval-report ... --verdict PASS`
- **Then:** Items `passed`. State is COMMIT.

### EVALUATE FAIL at max_rounds → COMMIT
- **Verifies:** FAIL at max rounds marks items failed.
- **Given:** EVALUATE, rounds == max_rounds.
- **When:** `advance --eval-report ... --verdict FAIL`
- **Then:** Items `failed`. State is COMMIT.

### EVALUATE FAIL within max_rounds → IMPLEMENT
- **Verifies:** FAIL within max rounds triggers re-implementation.
- **Given:** EVALUATE, rounds < max_rounds.
- **When:** `advance --eval-report ... --verdict FAIL`
- **Then:** State is IMPLEMENT. First item with eval file.

### COMMIT → ORIENT (more items)
- **Verifies:** Batch completion returns to ORIENT for next batch.
- **Given:** COMMIT, more items in layer.
- **When:** `advance --message "Complete batch"`
- **Then:** State is ORIENT.

### COMMIT → DONE (all complete)
- **Verifies:** Final batch completion reaches terminal state.
- **Given:** COMMIT, all layers complete.
- **When:** `advance --message "Complete batch"`
- **Then:** State is DONE.

### Implementing eval command outputs item details
- **Verifies:** Eval command assembles full evaluation context.
- **Given:** EVALUATE (implementing), batch has 2 items.
- **When:** `forgectl eval`
- **Then:** Output includes impl-eval.md contents, item details, report target.

### Failed items don't block dependents
- **Verifies:** Failed items are terminal for dependency resolution.
- **Given:** Item A `failed`, item B depends on A.
- **Then:** B is unblocked.

### DONE cannot advance
- **Verifies:** Terminal state rejects further advancement.
- **Given:** DONE.
- **When:** `advance`
- **Then:** Error.

---

## Implements
- Implementing phase: layer-ordered batched item delivery with one-at-a-time presentation
- COMMIT state for batch boundary commits
- First-round commit reminders in IMPLEMENT output
- Dual evaluator prompts: impl-eval.md for implementation sub-agent
