# Scaffold CLI

## Topic of Concern
> The scaffold CLI manages spec generation lifecycle state through a JSON-backed state machine with validated input and deterministic transitions.

## Context

The spec generation process involves multiple states (orient, select, draft, evaluate, refine, accept) applied to a queue of specifications. Without persistent state, an architect loses track of progress across sessions. The scaffold is a Go CLI tool (built with Cobra) that reads and writes a single JSON state file, enforcing valid transitions and providing the architect with unambiguous next-step guidance.

## Depends On
- None. The scaffold is a standalone tool with no runtime dependencies on other project components.

## Integration Points

| Component | Relationship |
|-----------|-------------|
| Spec generation skill | The skill document describes the process; the scaffold enforces the state machine that drives it |
| Evaluation sub-agent | The EVALUATE state is where the architect spawns a sub-agent; the scaffold tracks round count and verdict but does not invoke the sub-agent itself |
| Queue input file | The architect generates a JSON file conforming to the queue schema; the scaffold validates and ingests it during init |

---

## Interface

### Inputs

#### Queue Input File (provided via `--from` on `init`)

A JSON file conforming to this schema:

```json
{
  "specs": [
    {
      "name": "Repository Loading",
      "domain": "optimizer",
      "topic": "The optimizer clones or locates a repository and provides its path for downstream modules",
      "file": "optimizer/specs/repository-loading.md",
      "planning_sources": [
        ".workspace/planning/optimizer/repo-snapshot-loading.md"
      ],
      "depends_on": []
    }
  ]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `specs` | array | yes | Ordered list of specs to generate |
| `specs[].name` | string | yes | Display name for the spec (used in status output) |
| `specs[].domain` | string | yes | Domain grouping (e.g., "optimizer", "api", "web-portal", "protocols") |
| `specs[].topic` | string | yes | One-sentence topic of concern |
| `specs[].file` | string | yes | Target file path relative to project root |
| `specs[].planning_sources` | string[] | yes | Planning document paths the spec is derived from; may be empty array |
| `specs[].depends_on` | string[] | yes | Names of specs this one depends on; may be empty array |

No additional fields are permitted.

#### CLI Arguments

| Command | Flags | Description |
|---------|-------|-------------|
| `init` | `--rounds N` (required), `--from <path>` (required), `--user-guided` (optional, default false) | Initialize state file from a validated queue |
| `next` | none | Print the current state and what the architect does now |
| `advance` | `--file <path>` (DRAFT only), `--verdict PASS\|FAIL` (EVALUATE only) | Transition from current state to next |
| `status` | none | Print full session state: current spec, round, queue, completed |

### Outputs

All output is to stdout. The scaffold writes state changes to `scaffold-state.json`.

#### `next` output

Prints a structured block:

```
State:   EVALUATE
Spec:    Repository Loading
Domain:  optimizer
File:    optimizer/specs/repository-loading.md
Round:   2/3
Action:  Spawn Opus evaluation sub-agent for this spec.
```

#### `status` output

Prints current spec, queue contents grouped by domain, and completed specs with rounds taken.

#### `init` validation output (on failure)

When the input file fails validation, the scaffold prints:
1. Each validation error (missing field, extra field, wrong type) with the path to the offending location.
2. The complete valid schema as a reference, so the architect can see the expected structure.

The scaffold exits with a non-zero code on validation failure.

### Rejection

| Condition | Signal | Rationale |
|-----------|--------|-----------|
| `init` called when `scaffold-state.json` already exists | Error message: "State file already exists. Delete it to reinitialize." Exit code 1. | Prevents accidental loss of in-progress session state |
| `--from` file fails schema validation — missing required field | Error listing each missing field by JSON path. Prints full valid schema. Exit code 1. | Architect needs to see what's wrong and what's right |
| `--from` file fails schema validation — extra field present | Error listing each extra field by name. Prints full valid schema. Exit code 1. | Extra fields signal a schema misconception; reject rather than silently drop |
| `--from` file fails schema validation — wrong type | Error listing the field, expected type, and actual type. Prints full valid schema. Exit code 1. | Type mismatches cause downstream failures if not caught at init |
| `advance` called with `--file` outside of DRAFT state | Error: "--file is only valid in DRAFT state. Current state: <state>." Exit code 1. | Flag is meaningless outside DRAFT |
| `advance` called with `--verdict` outside of EVALUATE state | Error: "--verdict is only valid in EVALUATE state. Current state: <state>." Exit code 1. | Flag is meaningless outside EVALUATE |
| `advance` called in DRAFT without `--file` | Error: "DRAFT state requires --file <path>." Exit code 1. | The spec file path must be recorded |
| `advance` called in EVALUATE without `--verdict` | Error: "EVALUATE state requires --verdict PASS or --verdict FAIL." Exit code 1. | Verdict determines the transition |
| `advance` called when no current spec is active | Error: "No active spec. Run 'next' to see queue status." Exit code 1. | Cannot advance without a spec in progress |
| `next` or `advance` called before `init` | Error: "No state file found. Run 'scaffold init' first." Exit code 1. | State file must exist |
| Queue is empty and `advance` is called in ACCEPT | Message: "All specs complete." Exit code 0. | Normal termination |

---

## Behavior

### Initializing a Session

#### Preconditions
- No `scaffold-state.json` exists in the scaffold directory.
- A queue file path is provided via `--from`.
- `--rounds` is provided and is a positive integer.

#### Steps
1. Read the file at `--from`.
2. Parse as JSON.
3. Validate against the queue schema:
   a. Top-level must be an object with exactly one key: `specs`.
   b. `specs` must be a non-empty array.
   c. Each element must have exactly the fields: `name`, `domain`, `topic`, `file`, `planning_sources`, `depends_on`.
   d. No additional fields at any level.
   e. `name`, `domain`, `topic`, `file` must be non-empty strings.
   f. `planning_sources` and `depends_on` must be arrays of strings (may be empty).
4. If validation fails: print all errors, print the full valid schema, exit code 1.
5. If validation passes: create `scaffold-state.json` with the initialized state.

#### Postconditions
- `scaffold-state.json` exists with state `ORIENT`, `current_spec` set to null, `queue` populated from the input file, `completed` empty, `evaluation_rounds` set, `user_guided` set.

#### Error Handling
- File not found at `--from` path: "File not found: <path>." Exit code 1.
- File is not valid JSON: "Invalid JSON in <path>: <parse error>." Exit code 1.
- Schema validation failure: behavior described in step 4.

---

### Querying Next Action

#### Preconditions
- `scaffold-state.json` exists.

#### Steps
1. Read `scaffold-state.json`.
2. If `current_spec` is null and queue is non-empty: display that the architect is in ORIENT and the next spec to pick up.
3. If `current_spec` is null and queue is empty: display "All specs complete."
4. If `current_spec` is active: display the current state, spec details, round (if in EVALUATE), and the action the architect takes.

#### Postconditions
- No state mutation. `next` is read-only.

#### Error Handling
- State file missing: "No state file found. Run 'scaffold init' first." Exit code 1.
- State file is corrupt JSON: "State file is corrupt: <parse error>." Exit code 1.

---

### Advancing State

#### Preconditions
- `scaffold-state.json` exists.
- Current state allows the transition (see transition table).

#### Steps

The transition depends on the current state:

| From State | Condition | To State | Side Effects |
|------------|-----------|----------|-------------|
| ORIENT | always | SELECT | Pull next item from queue into `current_spec`, set state to SELECT |
| SELECT | `user_guided` is false | DRAFT | Set state to DRAFT |
| SELECT | `user_guided` is true | DRAFT | Set state to DRAFT (architect advances after user discussion) |
| DRAFT | `--file` provided | EVALUATE | Record file path on `current_spec`, set round to 1, set state to EVALUATE |
| EVALUATE | `--verdict PASS` | ACCEPT | Set state to ACCEPT |
| EVALUATE | `--verdict FAIL` and round < `evaluation_rounds` | REFINE | Set state to REFINE |
| EVALUATE | `--verdict FAIL` and round >= `evaluation_rounds` | ACCEPT | Set state to ACCEPT (max rounds reached; architect presents to user) |
| REFINE | always | EVALUATE | Increment round, set state to EVALUATE |
| ACCEPT | queue is non-empty | ORIENT | Move `current_spec` to `completed`, set `current_spec` to null, set state to ORIENT |
| ACCEPT | queue is empty | DONE | Move `current_spec` to `completed`, set `current_spec` to null |

Write the updated state to `scaffold-state.json`.

#### Postconditions
- State file reflects the new state.
- If transitioning from ACCEPT: `current_spec` moved to `completed` array.

#### Error Handling
- Invalid `--verdict` value (not PASS or FAIL): "Invalid verdict: <value>. Use PASS or FAIL." Exit code 1.
- Required flag missing for current state: specific error per state (see Rejection table).

---

### Querying Session Status

#### Preconditions
- `scaffold-state.json` exists.

#### Steps
1. Read `scaffold-state.json`.
2. Print session configuration: `evaluation_rounds`, `user_guided`.
3. Print current spec (if active): name, domain, state, round, file.
4. Print queue grouped by domain.
5. Print completed specs with rounds taken.

#### Postconditions
- No state mutation. `status` is read-only.

#### Error Handling
- State file missing: "No state file found. Run 'scaffold init' first." Exit code 1.

---

## Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `--rounds` | integer | none (required) | Number of evaluation rounds per spec |
| `--user-guided` | boolean | false | When set, SELECT state pauses for user discussion |
| `--from` | string | none (required on init) | Path to queue input JSON file |

---

## State File Schema

The `scaffold-state.json` file has this structure:

```json
{
  "evaluation_rounds": 3,
  "user_guided": true,
  "state": "EVALUATE",
  "current_spec": {
    "name": "Repository Loading",
    "domain": "optimizer",
    "topic": "The optimizer clones or locates a repository...",
    "file": "optimizer/specs/repository-loading.md",
    "planning_sources": [".workspace/planning/optimizer/repo-snapshot-loading.md"],
    "depends_on": [],
    "round": 2
  },
  "queue": [],
  "completed": [
    {
      "name": "Configuration Models",
      "domain": "optimizer",
      "file": "optimizer/specs/configuration-models.md",
      "rounds_taken": 1
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `evaluation_rounds` | integer | Set at init, immutable after |
| `user_guided` | boolean | Set at init, immutable after |
| `state` | string | One of: ORIENT, SELECT, DRAFT, EVALUATE, REFINE, ACCEPT, DONE |
| `current_spec` | object or null | The spec currently being worked on |
| `current_spec.round` | integer | Current evaluation round (starts at 1 in EVALUATE) |
| `queue` | array | Specs not yet started, in priority order |
| `completed` | array | Finished specs with metadata |
| `completed[].rounds_taken` | integer | How many evaluation rounds the spec went through |

---

## Invariants

1. **Single active spec.** At most one spec is in `current_spec` at any time. The queue holds pending specs; completed holds finished specs. There is no parallel processing.
2. **Round monotonicity.** The round counter on `current_spec` only increments. It starts at 1 when entering EVALUATE and increments on each REFINE → EVALUATE transition.
3. **Round bound.** The round counter never exceeds `evaluation_rounds`. When round equals `evaluation_rounds` and verdict is FAIL, the transition goes to ACCEPT, not REFINE.
4. **Queue order preserved.** Specs are pulled from the front of the queue. The order set at init is the order they are processed.
5. **State file is the only mutable artifact.** The scaffold reads the queue input file once (at init). After init, the state file is the sole source of truth. The queue input file is not referenced again.
6. **No implicit state.** Every piece of information the scaffold needs to determine the next transition is in `scaffold-state.json`. There is no in-memory state that survives across invocations.

---

## Edge Cases

- **Scenario:** Architect calls `init` when `scaffold-state.json` already exists.
  - **Expected behavior:** Error message and exit code 1. No mutation.
  - **Rationale:** Accidental reinit would destroy in-progress state. Force explicit deletion.

- **Scenario:** Queue input file has zero specs (empty `specs` array).
  - **Expected behavior:** Validation error: "specs array must be non-empty." Exit code 1.
  - **Rationale:** An empty queue means there's nothing to do. Catch this at init, not at runtime.

- **Scenario:** A spec in the queue has a `depends_on` entry that names a spec not present in the queue or completed list.
  - **Expected behavior:** Warning printed during init: "Warning: <spec name> depends on <dep name> which is not in the queue." Init still succeeds.
  - **Rationale:** Dependencies are advisory signals, not hard blockers. The architect decides whether to reorder.

- **Scenario:** Architect calls `advance` in SELECT when `user_guided` is true but hasn't discussed with the user.
  - **Expected behavior:** The scaffold advances to DRAFT. It does not enforce that discussion happened.
  - **Rationale:** The scaffold tracks state, it does not enforce process. The architect is responsible for following the process. The scaffold is a tool, not a gatekeeper.

- **Scenario:** State file contains a state value not in the valid set.
  - **Expected behavior:** Error: "Invalid state in state file: <value>. Valid states: ORIENT, SELECT, DRAFT, EVALUATE, REFINE, ACCEPT, DONE." Exit code 1.
  - **Rationale:** Corrupt or hand-edited state files are caught immediately.

- **Scenario:** `advance --verdict FAIL` when round equals `evaluation_rounds`.
  - **Expected behavior:** Transition to ACCEPT (not REFINE). The `next` command notes: "Max evaluation rounds reached. Present spec to user for final decision."
  - **Rationale:** Prevents infinite evaluation loops. After N rounds, a human decides.

---

## Testing Criteria

### Init creates valid state file
- **Verifies:** Initializing a Session behavior.
- **Given:** No `scaffold-state.json` exists. A valid queue file with 3 specs.
- **When:** `scaffold init --rounds 2 --from queue.json`
- **Then:** `scaffold-state.json` is created with state ORIENT, `current_spec` null, queue containing 3 specs in order, `evaluation_rounds` 2, `user_guided` false.

### Init rejects invalid queue — missing field
- **Verifies:** Schema validation rejection.
- **Given:** A queue file where one spec is missing the `domain` field.
- **When:** `scaffold init --rounds 2 --from queue.json`
- **Then:** Exit code 1. Output lists the missing field. Output includes the full valid schema.

### Init rejects invalid queue — extra field
- **Verifies:** Extra field rejection.
- **Given:** A queue file where one spec has an extra field `priority`.
- **When:** `scaffold init --rounds 2 --from queue.json`
- **Then:** Exit code 1. Output names the extra field `priority`. Output includes the full valid schema.

### Init rejects existing state file
- **Verifies:** Existing state file edge case.
- **Given:** `scaffold-state.json` already exists.
- **When:** `scaffold init --rounds 2 --from queue.json`
- **Then:** Exit code 1. Error message about existing state file. No mutation to the file.

### Full lifecycle single spec
- **Verifies:** All state transitions for one spec.
- **Given:** Init with 1 spec, `--rounds 1`.
- **When:** `advance` through ORIENT → SELECT → DRAFT (with `--file`) → EVALUATE (with `--verdict PASS`) → ACCEPT.
- **Then:** State is DONE. `completed` has 1 entry with `rounds_taken` 1. `queue` is empty. `current_spec` is null.

### Evaluate-refine loop respects round limit
- **Verifies:** Round bound invariant.
- **Given:** Init with 1 spec, `--rounds 2`. Advance to EVALUATE.
- **When:** `advance --verdict FAIL` (round 1) → advance (REFINE) → `advance --verdict FAIL` (round 2).
- **Then:** State is ACCEPT (not REFINE). Round is 2.

### Next is read-only
- **Verifies:** No state mutation on `next`.
- **Given:** State file in any state.
- **When:** `next` is called.
- **Then:** State file is byte-identical before and after.

### Status displays grouped output
- **Verifies:** Querying Session Status behavior.
- **Given:** Init with specs across 3 domains. 1 completed, 1 active, 2 queued.
- **When:** `scaffold status`
- **Then:** Output shows current spec details, queue grouped by domain, completed list with rounds taken.

### Dependency warning on init
- **Verifies:** Dependency edge case.
- **Given:** Queue file where spec B depends_on spec A, but spec A is not in the queue.
- **When:** `scaffold init --rounds 1 --from queue.json`
- **Then:** Init succeeds. Warning printed naming the unresolved dependency.

### Advance with wrong flag for state
- **Verifies:** Flag-state mismatch rejection.
- **Given:** State is ORIENT.
- **When:** `scaffold advance --verdict PASS`
- **Then:** Exit code 1. Error: "--verdict is only valid in EVALUATE state."

---

## Implements
- Scaffold state machine design from spec generation skill process
