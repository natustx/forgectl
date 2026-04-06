# Is-Done Command

## Topic of Concern
> The scaffold reports via exit code whether the active session has reached its terminal state with no work remaining.

## Context

Forgectl sessions progress through a state machine that ends at a phase-specific terminal state. Agents and scripts driving the session in a loop need a machine-readable signal to know when to stop. Currently the only way to determine this is to parse the human-readable output of `forgectl status` for a state name — a brittle approach that couples callers to output formatting.

The `is-done` command provides a clean, scriptable check: exit code 0 means done, non-zero means work remains.

## Depends On
- **state-persistence** — reads the state file to determine current phase and state.

## Integration Points

| Spec | Relationship |
|------|-------------|
| state-persistence | `is-done` loads the state file; it does not modify it |
| phase-transitions | Terminal states are defined by the phase transition rules; `is-done` checks for them |

---

## Interface

### Inputs

```
forgectl is-done
```

No flags. No arguments. The command reads the active session state file.

### Outputs

**Session is terminal (done):**

```
done
```

Exit code: `0`

**Session has work remaining (not done):**

```
not done: DRAFT (specifying)
```

Exit code: `1`

The format for the not-done case is: `not done: <STATE> (<PHASE>)`

**No state file found:**

```
Error: no state file found. Run 'forgectl init' first
```

Exit code: `1`

### Rejection

| Condition | Signal | Rationale |
|-----------|--------|-----------|
| No `.forgectl/` directory found | Error: "No .forgectl directory found." Exit code 1. | No project context to evaluate |
| No state file exists | Error: "no state file found. Run 'forgectl init' first". Exit code 1. | No session to check |

---

## Behavior

### Terminal State Detection

The command loads the state file and evaluates whether the session has reached a terminal state. Terminal states vary by phase:

| Phase | Terminal state | Condition |
|-------|--------------|-----------|
| specifying | `PHASE_SHIFT` | `phase_shift.from == "specifying"` |
| planning | `PHASE_SHIFT` | `phase_shift.from == "planning"` |
| implementing | `DONE` | Always terminal when reached |

#### Preconditions
- A `.forgectl/` directory exists in the directory tree (project root discovery succeeds).
- A state file exists and is valid JSON.

#### Steps
1. Discover the project root via `.forgectl/` directory walk.
2. Load the configuration from `.forgectl/config`.
3. Load the state file from the configured state directory. Crash recovery runs as part of load.
4. Evaluate the terminal state condition against the current phase and state.
5. Print the result line and exit with the appropriate code.

#### Postconditions
- The state file is not modified.
- The exit code reflects the terminal state check: 0 if terminal, 1 otherwise.

#### Error Handling

| Failure | Response |
|---------|----------|
| Project root not found | Print error. Exit 1. |
| State file missing | Print error with init hint. Exit 1. |
| State file corrupt, backup exists | Crash recovery restores from backup (standard load behavior). Evaluation proceeds. |
| State file corrupt, no backup | Print error. Exit 1. |

### Read-Only Operation

The command does not write to the state file, create log entries, or produce any side effects. It is a pure read operation.

---

## Invariants

1. **No state mutation.** The command never modifies the state file, writes logs, or produces side effects.
2. **Exit code is deterministic.** The same state file always produces the same exit code.
3. **Terminal detection is phase-complete.** All three phases (specifying, planning, implementing) have a defined terminal condition. No phase is uncovered.

---

## Edge Cases

- **Scenario:** State file exists but session was initialized directly into implementing phase (via `forgectl init --phase implementing`).
  - **Expected:** Terminal detection checks `state == DONE` for implementing. Works identically regardless of `started_at_phase`.
  - **Rationale:** Terminal detection is based on current phase and state, not session origin.

- **Scenario:** State file is in `PHASE_SHIFT` from planning to implementing.
  - **Expected:** `is-done` reports `done`. The planning phase has no more work.
  - **Rationale:** `PHASE_SHIFT` signals the current phase is complete. The next phase has not started.

- **Scenario:** State file was recovered from backup during load.
  - **Expected:** Terminal detection proceeds normally on the recovered state.
  - **Rationale:** Recovery is transparent to commands that only read state.

- **Scenario:** State file is at `DONE` in the specifying phase (queue empty, before reconciliation).
  - **Expected:** `is-done` reports `not done: DONE (specifying)`. Work remains (reconciliation).
  - **Rationale:** `DONE` in specifying is not terminal — reconciliation, cross-reference review, and `COMPLETE` must still run before `PHASE_SHIFT`.

---

## Testing Criteria

### Terminal — specifying complete
- **Verifies:** Exit code 0 when specifying phase reaches PHASE_SHIFT.
- **Given:** State file with `state: "PHASE_SHIFT"`, `phase_shift.from: "specifying"`, `phase_shift.to: "planning"`.
- **When:** `forgectl is-done`
- **Then:** Stdout: `done`. Exit code: 0.

### Terminal — planning complete
- **Verifies:** Exit code 0 when planning phase reaches PHASE_SHIFT.
- **Given:** State file with `state: "PHASE_SHIFT"`, `phase_shift.from: "planning"`, `phase_shift.to: "implementing"`.
- **When:** `forgectl is-done`
- **Then:** Stdout: `done`. Exit code: 0.

### Terminal — implementing complete
- **Verifies:** Exit code 0 when implementing phase reaches DONE.
- **Given:** State file with `phase: "implementing"`, `state: "DONE"`.
- **When:** `forgectl is-done`
- **Then:** Stdout: `done`. Exit code: 0.

### Not done — mid-specifying
- **Verifies:** Exit code 1 when work remains.
- **Given:** State file with `phase: "specifying"`, `state: "DRAFT"`.
- **When:** `forgectl is-done`
- **Then:** Stdout: `not done: DRAFT (specifying)`. Exit code: 1.

### Not done — specifying DONE is not terminal
- **Verifies:** DONE in specifying is not terminal (reconciliation pending).
- **Given:** State file with `phase: "specifying"`, `state: "DONE"`.
- **When:** `forgectl is-done`
- **Then:** Stdout: `not done: DONE (specifying)`. Exit code: 1.

### Not done — mid-implementing
- **Verifies:** Exit code 1 during active implementation.
- **Given:** State file with `phase: "implementing"`, `state: "IMPLEMENT"`.
- **When:** `forgectl is-done`
- **Then:** Stdout: `not done: IMPLEMENT (implementing)`. Exit code: 1.

### No state file
- **Verifies:** Graceful error when no session exists.
- **Given:** `.forgectl/` directory exists but no state file.
- **When:** `forgectl is-done`
- **Then:** Error message with init hint. Exit code: 1.

### No project root
- **Verifies:** Graceful error when no .forgectl/ directory.
- **Given:** Working directory with no `.forgectl/` in the directory tree.
- **When:** `forgectl is-done`
- **Then:** Error: "No .forgectl directory found." Exit code: 1.

---

## Implements
- Scriptable terminal state detection for agent loops and automation
- Machine-readable exit code contract for session completion checks
