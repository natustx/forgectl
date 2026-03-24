# State Persistence

## Topic of Concern
> The scaffold persists session state through atomic file writes with crash recovery on startup.

## Context

Every forgectl command reads and writes a single JSON state file (`forgectl-state.json`). The file must survive process crashes, power loss, and interrupted writes without data loss. A backup file and write-ahead pattern provide crash safety. On startup, the scaffold detects and recovers from interrupted writes before any command logic runs.

Completed sessions are archived to a git-tracked directory for audit purposes.

All scaffold output is to stdout. The scaffold writes state changes to `forgectl-state.json` and (during implementing phase) updates `passes` and `rounds` fields in plan.json.

## Depends On
- None. State persistence is a foundational concern with no upstream dependencies.

## Integration Points

| Spec | Relationship |
|------|-------------|
| session-init | Creates the initial state file; persistence layer provides the write mechanism |
| spec-lifecycle | Reads and mutates specifying phase state through the persistence layer |
| spec-reconciliation | Reads and mutates reconciliation state through the persistence layer |
| plan-production | Reads and mutates planning phase state through the persistence layer |
| batch-implementation | Reads and mutates implementing phase state and plan.json through the persistence layer |

---

## Interface

### Inputs

The persistence layer is internal — no CLI inputs. All commands trigger reads and writes through it.

The `status` command reads and displays the state file:

| Command | Flags | Description |
|---------|-------|-------------|
| `status` | none | Print current state with action guidance + full session overview |

### Outputs

#### File Layout

```
forgectl-state.json           ← active state (gitignored)
forgectl-state.json.bak       ← previous state (gitignored)
forgectl-state.json.tmp       ← write-in-progress (transient, gitignored)
sessions/                      ← archived completed sessions (git tracked)
```

#### `status` Output

The `status` command prints the current state with action guidance at the top, followed by the full session overview. The session header includes session file, phase, and configuration. Each phase that has been reached contributes a section. The current state's action guidance matches the `advance` output for that state.

**Mid-specifying:**

```
Session: forgectl-state.json
Phase:   specifying
Config:  rounds=1-3, batch_size=2, guided=true

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

**Mid-planning:**

```
Session: forgectl-state.json
Phase:   planning
Config:  rounds=1-3, batch_size=2

--- Current ---

State:   EVALUATE
Plan:    Service Configuration (launcher)
File:    launcher/.workspace/implementation_plan/plan.json
Round:   1/3
Action:  Run evaluation sub-agent against the plan (round 1/3).
         Sub-agent: forgectl eval
         Advance with --verdict PASS|FAIL --eval-report <path>.

--- Specifying ---

  Complete (5 specs, reconciled)

--- Planning ---

  Evals: (none yet)

--- Queue ---

  empty
```

**Mid-implementing:**

```
Session: forgectl-state.json
Phase:   implementing
Config:  batch_size=2, rounds=1-3

--- Current ---

State:   IMPLEMENT
Plan:    Service Configuration (launcher)
File:    launcher/.workspace/implementation_plan/plan.json
Layer:   L1 Core (2 items)
Batch:   3/3
Item:    [daemon.io] PID file I/O operations (2 of 2)
Round:   0
Action:  Implement this item.
         When complete, run: forgectl advance --message <commit msg>

--- Specifying ---

  Complete (5 specs)

--- Planning ---

  Accepted (2 rounds)
    Round 1: FAIL — launcher/.workspace/implementation_plan/evals/round-1.md
    Round 2: PASS — launcher/.workspace/implementation_plan/evals/round-2.md

--- Implementing ---

  Layer L0 (Foundation): complete
    [bootstrap]     passed  (1 round)
    [config.types]  passed  (1 round)
    [config.load]   passed  (2 rounds)

  Layer L1 (Core): in progress
    [daemon.types]  done    (0 rounds)
    [daemon.io]     pending (0 rounds)
```

**Started at implementing directly (`--phase implementing`):**

```
Session: forgectl-state.json
Phase:   implementing (started here)
Config:  batch_size=2, rounds=1-3

--- Current ---

State:   EVALUATE
Plan:    launcher/.workspace/implementation_plan/plan.json
Layer:   L0 Foundation (3 items)
Batch:   1/2
Round:   1/3
Items:   [config.types], [config.load]
Action:  Ask the evaluation sub-agent to verify batch items against their tests.
         The sub-agent should run: forgectl eval
         After reviewing the eval report, run:
           forgectl advance --eval-report <path> --verdict PASS|FAIL

--- Implementing ---

  Layer L0 (Foundation): in progress
    [bootstrap]     passed  (1 round)
    [config.types]  done    (1 round)
    [config.load]   done    (1 round)
```

### Rejection

| Condition | Signal | Rationale |
|-----------|--------|-----------|
| `advance` or `status` or `eval` called before `init` | Error. Exit code 1. | State file must exist |

---

## Behavior

### Atomic Writes

Every state mutation follows this sequence:
1. Write new state to `forgectl-state.json.tmp`.
2. Rename `forgectl-state.json` → `forgectl-state.json.bak`.
3. Rename `forgectl-state.json.tmp` → `forgectl-state.json`.

Steps 2 and 3 are filesystem renames — atomic on POSIX. If the process crashes between steps, the startup recovery logic handles it.

### Startup Recovery

On every command (before reading state), the scaffold checks:

| Condition | Action |
|-----------|--------|
| `.json` exists, `.tmp` does not | Normal. Proceed. |
| `.json` missing, `.bak` exists | Crashed between step 2 and 3. Rename `.bak` → `.json`. Warn user. |
| `.json` missing, `.tmp` exists | Crashed between step 1 and 2. Rename `.tmp` → `.json`. Warn user. |
| `.json` exists, `.tmp` exists | Crashed after step 1, before cleanup. Delete `.tmp`. Proceed with `.json`. |
| `.json` corrupt (invalid JSON) | Rename `.json` → `.json.corrupt`, rename `.bak` → `.json`. Warn user. |
| None exist | No state. Only `init` is valid. |

### Session Archiving

Completed session state files are archived to a permanent directory:

```
sessions/
├── optimizer-2026-03-15.json
├── launcher-2026-03-21.json
└── ...
```

- The active `forgectl-state.json` is gitignored (ephemeral working state).
- Archived sessions are committed to git (permanent audit trail).
- Naming convention: `<domain>-<date>.json`.
- Archive before starting a new session. The active state file must be deleted (or the scaffold will reject `init`).

### State File Schema

```json
{
  "phase": "implementing",
  "state": "IMPLEMENT",
  "batch_size": 2,
  "min_rounds": 1,
  "max_rounds": 3,
  "user_guided": true,
  "started_at_phase": "specifying",

  "specifying": {
    "current_spec": null,
    "queue": [],
    "completed": [
      {
        "id": 1,
        "name": "Configuration Models",
        "domain": "optimizer",
        "file": "optimizer/specs/configuration-models.md",
        "rounds_taken": 2,
        "commit_hash": "a1b2c3d",
        "evals": [
          { "round": 1, "verdict": "FAIL", "eval_report": "optimizer/specs/.eval/configuration-models-r1.md" },
          { "round": 2, "verdict": "PASS", "eval_report": "optimizer/specs/.eval/configuration-models-r2.md" }
        ]
      }
    ],
    "reconcile": {
      "round": 1,
      "evals": [
        { "round": 1, "verdict": "PASS" }
      ]
    }
  },

  "planning": {
    "current_plan": {
      "id": 1,
      "name": "Service Configuration",
      "domain": "launcher",
      "topic": "Implementation plan for service configuration",
      "file": "launcher/.workspace/implementation_plan/plan.json",
      "specs": ["launcher/specs/service-configuration.md"],
      "code_search_roots": ["launcher/"]
    },
    "round": 2,
    "evals": [
      { "round": 1, "verdict": "FAIL", "eval_report": "launcher/.workspace/implementation_plan/evals/round-1.md" },
      { "round": 2, "verdict": "PASS", "eval_report": "launcher/.workspace/implementation_plan/evals/round-2.md" }
    ],
    "queue": [],
    "completed": []
  },

  "implementing": {
    "current_layer": { "id": "L0", "name": "Foundation" },
    "batch_number": 2,
    "current_batch": {
      "items": ["config.types", "config.load"],
      "current_item_index": 0,
      "eval_round": 0,
      "evals": []
    },
    "layer_history": [
      {
        "layer_id": "L0",
        "batches": [
          {
            "batch_number": 1,
            "items": ["bootstrap"],
            "eval_rounds": 1,
            "evals": [
              { "round": 1, "verdict": "PASS", "eval_report": "launcher/.workspace/implementation_plan/evals/batch-1-round-1.md" }
            ]
          }
        ]
      }
    ]
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `phase` | string | `specifying`, `planning`, or `implementing` |
| `state` | string | Current state within the active phase |
| `batch_size` | integer | Max items per batch |
| `min_rounds` | integer | Minimum eval rounds per cycle |
| `max_rounds` | integer | Maximum eval rounds per cycle |
| `user_guided` | boolean | Whether guided pauses are active |
| `started_at_phase` | string | Which phase the session was initialized at (for display) |
| **Specifying** | | |
| `specifying.current_spec` | object/null | The spec being worked on |
| `specifying.queue` | array | Remaining specs |
| `specifying.completed` | array | Finished specs with eval history and commit hashes |
| `specifying.completed[].evals` | array | Full eval trail per spec |
| `specifying.reconcile` | object | Reconciliation round and eval history |
| **Planning** | | |
| `planning.current_plan` | object/null | The plan being worked on |
| `planning.round` | integer | Current planning eval round |
| `planning.evals` | array | Planning eval history |
| `planning.queue` | array | Remaining plans |
| `planning.completed` | array | Finished plans |
| **Implementing** | | |
| `implementing.current_layer` | object | Active layer |
| `implementing.batch_number` | integer | Global batch counter (1-indexed) |
| `implementing.current_batch` | object | Active batch state |
| `implementing.current_batch.items` | string[] | Item IDs in batch |
| `implementing.current_batch.current_item_index` | integer | 0-based index |
| `implementing.current_batch.eval_round` | integer | Current eval round for batch |
| `implementing.current_batch.evals` | array | Batch eval history |
| `implementing.layer_history` | array | Completed batches and layers |

Phase sections that haven't been reached yet are `null` in the state file. When starting at a later phase (`--phase planning`), earlier phase sections remain `null`.

---

## Invariants

1. **Phase is authoritative.** The `phase` field determines which states are valid and how shared state names behave.
2. **State file is durable.** Atomic writes with backup prevent corruption. Startup recovery handles interrupted writes.
3. **Eval history is append-only.** Eval records accumulate and are never deleted or modified.

---

## Edge Cases

- **Scenario:** Process crashes after writing `.tmp` but before renaming `.bak`.
  - **Expected:** On next startup, `.tmp` is renamed to `.json`. State recovered.
  - **Rationale:** The `.tmp` file contains the most recent intended state; recovering it preserves the latest mutation.

- **Scenario:** State file contains invalid JSON but `.bak` is intact.
  - **Expected:** Corrupt file renamed to `.json.corrupt`, `.bak` restored as `.json`. Warning printed.
  - **Rationale:** Rolling back to the previous known-good state is safer than failing. The corrupt file is preserved for debugging.

- **Scenario:** Both `.json` and `.tmp` exist.
  - **Expected:** `.tmp` deleted, `.json` used as-is.
  - **Rationale:** The `.json` file was not yet replaced, so it holds the last committed state. The orphaned `.tmp` is from an incomplete write.

---

## Testing Criteria

### Recovery from crash between backup and rename
- **Verifies:** Startup recovery handles missing `.json` with existing `.bak`.
- **Given:** `.json` missing, `.bak` exists.
- **When:** Any command runs.
- **Then:** `.bak` renamed to `.json`. Warning printed. Command proceeds.

### Recovery from corrupt state file
- **Verifies:** Startup recovery handles corrupt JSON with valid backup.
- **Given:** `.json` contains invalid JSON, `.bak` exists.
- **When:** Any command runs.
- **Then:** `.json` renamed to `.json.corrupt`. `.bak` renamed to `.json`. Warning printed.

---

## Implements
- Atomic state file writes with backup and startup recovery
- State file schema for all three phases
- Session archiving to git-tracked sessions directory
- Status command: session overview assembled from all phase sections
