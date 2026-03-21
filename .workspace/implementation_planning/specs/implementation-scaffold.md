# Implementation Plan Scaffold

## Topic of Concern
> The implementation plan scaffold manages plan generation lifecycle state through a JSON-backed state machine with validated input, deterministic transitions, and structured study phases that precede drafting.

## Context

The implementation planning process involves studying specs, codebase, and packages before drafting a plan. Without persistent state, an architect loses track of study progress and plan quality across sessions. The scaffold is a Go CLI tool (built with Cobra) that reads and writes a single JSON state file, enforcing valid transitions and providing the architect with unambiguous next-step guidance.

The scaffold extends the spec generation scaffold's evaluate/refine loop with three structured study phases (STUDY_SPECS, STUDY_CODE, STUDY_PACKAGES) that build context before drafting. It tracks study artifacts, evaluation history, deficiencies, and fixes — creating an audit trail from orientation to acceptance.

The implementation plan prompt (`IMPLEMENTATION_PLAN_PROMPT.md`) defines the methodology. The scaffold enforces the sequencing and state transitions that drive it.

## Depends On
- None. The scaffold is a standalone tool with no runtime dependencies on other project components.

## Integration Points

| Component | Relationship |
|-----------|-------------|
| Implementation plan prompt | The prompt document describes the methodology; the scaffold enforces the state machine that sequences it |
| Spec generation scaffold | The spec scaffold produces completed specs; the implementation scaffold consumes those specs as input context |
| SPEC_MANIFEST.md | STUDY_SPECS reads this manifest to locate spec files relevant to the plan |
| Evaluation sub-agent | The EVALUATE state is where the architect spawns a sub-agent; the scaffold tracks round count, verdict, deficiencies, and fixes |
| Queue input file | The architect generates a JSON file conforming to the queue schema; the scaffold validates and ingests it during init |

---

## Interface

### Inputs

#### Queue Input File (provided via `--from` on `init`)

A JSON file conforming to this schema:

```json
{
  "plans": [
    {
      "name": "Protocol Implementation",
      "domain": "protocols",
      "topic": "Implementation plan for WS1 and WS2 message contract specs",
      "file": "protocols/.workspace/implementation_plan/IMPLEMENTATION_PLAN.md",
      "specs": [
        "protocols/ws1/specs/ws1-message-contract.md",
        "protocols/ws2/specs/ws2-message-contract.md"
      ],
      "code_search_roots": ["api/", "optimizer/", "portal/"]
    }
  ]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `plans` | array | yes | Ordered list of plans to generate |
| `plans[].name` | string | yes | Display name for the plan (used in status output) |
| `plans[].domain` | string | yes | Domain grouping (e.g., "protocols", "optimizer", "api", "portal") |
| `plans[].topic` | string | yes | One-sentence topic of concern |
| `plans[].file` | string | yes | Target file path relative to project root |
| `plans[].specs` | string[] | yes | Spec file paths to study; may be empty array |
| `plans[].code_search_roots` | string[] | yes | Directory roots for codebase exploration; may be empty array |
No additional fields are permitted.

#### CLI Arguments

| Command | Flags | Description |
|---------|-------|-------------|
| `init` | `--min-rounds N` (default 1), `--max-rounds N` (required), `--from <path>` (required), `--sub-agents N` (default 3), `--user-guided` (optional, default false) | Initialize state file from a validated queue. Prints initial state after creation. |
| `advance` | `--file <path>` (optional, DRAFT only), `--verdict PASS\|FAIL` (EVALUATE), `--message <text>` (required with PASS in EVALUATE), `--deficiencies <csv>` (with FAIL in EVALUATE), `--fixed <text>` (in REFINE), `--notes <text>` (optional, STUDY_SPECS/STUDY_CODE/STUDY_PACKAGES) | Transition from current state to next. Prints the new state after transitioning. |
| `status` | none | Print full session state: current plan, study progress, eval history, queue, completed |

### Outputs

All output is to stdout. The scaffold writes state changes to `impl-scaffold-state.json`.

#### `advance` output

After transitioning, `advance` prints a structured block showing the new state and what to do next:

```
State:   STUDY_CODE
ID:      1
Plan:    Protocol Implementation
Domain:  protocols
File:    protocols/.workspace/implementation_plan/IMPLEMENTATION_PLAN.md
Specs:   protocols/ws1/specs/ws1-message-contract.md, protocols/ws2/specs/ws2-message-contract.md
Roots:   api/, optimizer/, portal/
Action:  Explore the codebase using sub-agents. Focus on code related to the specs under study.
         Sub-agents: 3. Search roots: api/, optimizer/, portal/.
         Advance with --notes <summary of findings>.
```

#### `status` output

Prints session config, current plan with study notes and eval history, queue grouped by domain, and completed plans with eval trail.

#### `init` validation output (on failure)

When the input file fails validation, the scaffold prints:
1. Each validation error (missing field, extra field, wrong type) with the path to the offending location.
2. The complete valid schema as a reference.

The scaffold exits with a non-zero code on validation failure.

### Rejection

| Condition | Signal | Rationale |
|-----------|--------|-----------|
| `init` called when `impl-scaffold-state.json` already exists | Error: "State file already exists. Delete it to reinitialize." Exit code 1. | Prevents accidental loss of in-progress state |
| `--from` file fails schema validation | Error listing violations. Prints full valid schema. Exit code 1. | Architect needs to see what's wrong |
| `--min-rounds` exceeds `--max-rounds` | Error: "--min-rounds cannot exceed --max-rounds." Exit code 1. | Invalid configuration |
| `--sub-agents` is less than 1 | Error: "--sub-agents must be at least 1." Exit code 1. | At least one agent is required |
| `advance` called with `--file` outside of DRAFT state | Error naming the current state. Exit code 1. | Flag is meaningless outside DRAFT |
| `advance` called with `--verdict` outside of EVALUATE | Error naming the current state. Exit code 1. | Verdict is only valid in EVALUATE |
| `advance` called in EVALUATE without `--verdict` | Error. Exit code 1. | Verdict determines the transition |
| `advance` called with `--verdict PASS` in EVALUATE without `--message` | Error. Exit code 1. | Accepted plans must be committed |
| `advance` or `status` called before `init` | Error. Exit code 1. | State file must exist |

---

## Behavior

### Initializing a Session

#### Preconditions
- No `impl-scaffold-state.json` exists.
- `--from`, `--max-rounds` are provided.
- `--min-rounds` <= `--max-rounds`.
- `--sub-agents` >= 1.

#### Steps
1. Read and parse the file at `--from`.
2. Validate against the queue schema.
3. If validation fails: print errors and schema, exit code 1.
4. If validation passes: create `impl-scaffold-state.json` with state ORIENT.

#### Postconditions
- State file exists with `min_rounds`, `max_rounds`, `sub_agents`, `user_guided` set, queue populated, completed empty.

#### Error Handling
- File not found: error with path. Exit code 1.
- Invalid JSON: error with parse details. Exit code 1.
- Schema failure: error listing violations, print valid schema. Exit code 1.

---

### Advancing State

#### Preconditions
- `impl-scaffold-state.json` exists.

#### Steps

| From State | Condition | To State | Side Effects |
|------------|-----------|----------|-------------|
| ORIENT | always | STUDY_SPECS | Pull next from queue into `current_plan` |
| STUDY_SPECS | always | STUDY_CODE | Record `--notes` on study record if provided |
| STUDY_CODE | always | STUDY_PACKAGES | Record `--notes` on study record if provided |
| STUDY_PACKAGES | always | SELECT | Record `--notes` on study record if provided |
| SELECT | always | DRAFT | — |
| DRAFT | always | EVALUATE | If `--file` provided, override file path. Set round to 1. |
| EVALUATE | `--verdict PASS` | ACCEPT | Record eval (PASS). Auto-commit with `--message`. |
| EVALUATE | `--verdict FAIL`, round < `min_rounds` | REFINE | Record eval (FAIL + deficiencies). Auto-refine. |
| EVALUATE | `--verdict FAIL`, round >= `min_rounds` | REFINE | Record eval (FAIL + deficiencies). |
| REFINE | always | EVALUATE | Record `--fixed` on last eval. Increment round. |
| ACCEPT | queue non-empty | ORIENT | Move plan to completed (with study notes, eval history, commit hash). |
| ACCEPT | queue empty | DONE | Move plan to completed. |
| DONE | — | Error: nothing to advance | Terminal |

#### Postconditions
- State file reflects the new state.
- Study notes accumulate on `current_plan.study`.
- Eval records accumulate on `current_plan.evals` and carry to `completed[].evals`.

#### Error Handling
- Invalid flags for state: specific error per state.
- Invalid verdict value: error.

---

### Study Phases

The three study phases build context before drafting. Each phase has a specific focus aligned with the implementation plan prompt methodology.

#### STUDY_SPECS

The architect studies the specs listed in `current_plan.specs` and the SPEC_MANIFEST.md. This includes:
- Reading the full spec files
- Reviewing git diffs associated with the specs (commits that introduced or modified them)
- Understanding dependencies, integration points, and cross-references

The `--notes` flag captures a summary of findings. The scaffold records these notes but does not enforce their content.

#### STUDY_CODE

The architect explores the codebase using sub-agents. The number of sub-agents is configured at init via `--sub-agents` (default 3).

<!-- TODO: The default sub-agent count of 3 is provisional. Adjust based on experience with codebase size and complexity. The optimal number may vary per domain. -->

The sub-agents search within the directories listed in `current_plan.code_search_roots`. The scaffold does not launch sub-agents itself — it provides the configuration and the architect orchestrates them. The scaffold records the architect's summary via `--notes`.

#### STUDY_PACKAGES

The architect studies the project's technical stack:
- Package manifest files in the codebase (go.mod, pyproject.toml, package.json)
- Library documentation via Context7 or other sources as referenced in CLAUDE.md
- Any additional package references in project configuration

The `--notes` flag captures a summary of findings.

---

## Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `--min-rounds` | integer | 1 | Minimum evaluation rounds before acceptance is allowed on FAIL |
| `--max-rounds` | integer | none (required) | Maximum evaluation rounds per plan |
| `--sub-agents` | integer | 3 | Number of sub-agents for STUDY_CODE phase |
| `--user-guided` | boolean | false | When set, SELECT state (after study phases) pauses for user discussion before drafting |
| `--from` | string | none (required on init) | Path to queue input JSON file |

<!-- TODO: --sub-agents default of 3 is provisional. See STUDY_CODE behavior section. -->

---

## State File Schema

```json
{
  "min_rounds": 1,
  "max_rounds": 3,
  "sub_agents": 3,
  "user_guided": false,
  "state": "STUDY_CODE",
  "current_plan": {
    "id": 1,
    "name": "Protocol Implementation",
    "domain": "protocols",
    "topic": "Implementation plan for WS1 and WS2 message contract specs",
    "file": "protocols/.workspace/implementation_plan/IMPLEMENTATION_PLAN.md",
    "specs": [
      "protocols/ws1/specs/ws1-message-contract.md",
      "protocols/ws2/specs/ws2-message-contract.md"
    ],
    "code_search_roots": ["api/", "optimizer/", "portal/"],
    "study": {
      "specs_notes": "WS1 and WS2 define typed JSON messages over WebSocket. WS1 is portal-API, WS2 is API-optimizer. Both use type discriminator.",
      "code_notes": "",
      "packages_notes": ""
    },
    "round": 0,
    "evals": []
  },
  "queue": [],
  "completed": [
    {
      "id": 0,
      "name": "Previous Plan",
      "domain": "optimizer",
      "file": "optimizer/.workspace/implementation_plan/IMPLEMENTATION_PLAN.md",
      "rounds_taken": 2,
      "commit_hash": "a1b2c3d",
      "study": {
        "specs_notes": "...",
        "code_notes": "...",
        "packages_notes": "..."
      },
      "evals": [
        { "round": 1, "verdict": "FAIL", "deficiencies": ["Completeness", "Traceability"], "fixed": "Added missing acceptance criteria rows" },
        { "round": 2, "verdict": "PASS", "deficiencies": null, "fixed": "" }
      ]
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `min_rounds` | integer | Minimum eval rounds before acceptance is allowed on FAIL |
| `max_rounds` | integer | Maximum eval rounds per plan |
| `sub_agents` | integer | Number of sub-agents for STUDY_CODE |
| `user_guided` | boolean | Whether SELECT (after study phases) pauses for discussion before drafting |
| `state` | string | ORIENT, STUDY_SPECS, STUDY_CODE, STUDY_PACKAGES, SELECT, DRAFT, EVALUATE, REFINE, ACCEPT, DONE |
| `current_plan.study` | object | Study phase notes |
| `current_plan.study.specs_notes` | string | Summary from STUDY_SPECS phase |
| `current_plan.study.code_notes` | string | Summary from STUDY_CODE phase |
| `current_plan.study.packages_notes` | string | Summary from STUDY_PACKAGES phase |
| `current_plan.round` | integer | Current evaluation round (0 before EVALUATE) |
| `current_plan.evals` | array | Evaluation history for the current plan |
| `current_plan.evals[].round` | integer | Which round this eval was |
| `current_plan.evals[].verdict` | string | PASS or FAIL |
| `current_plan.evals[].deficiencies` | string[] | Failed dimension names (on FAIL) |
| `current_plan.evals[].fixed` | string | What was fixed after this eval (populated during REFINE) |
| `completed[].study` | object | Study notes carried from current_plan |
| `completed[].evals` | array | Full eval history carried from current_plan |

---

## Invariants

1. **Single active plan.** At most one plan is in `current_plan` at any time.
2. **Round monotonicity.** The round counter only increments.
3. **Min/max round bounds.** FAIL below `min_rounds` auto-refines. FAIL at or above `min_rounds` goes to REFINE (architect decides whether to continue or accept).
4. **Queue order preserved.** Plans are pulled from the front of the queue.
5. **State file is the only mutable artifact.** The queue input file is read once at init.
6. **No implicit state.** All information for transitions is in the state file.
7. **Eval history is append-only.** Eval records accumulate and are never deleted or modified (except `fixed` is set during REFINE).
8. **Study phases precede SELECT.** STUDY_SPECS must complete before STUDY_CODE, which must complete before STUDY_PACKAGES. All three must complete before SELECT. No phase is skipped.
9. **Study notes persist.** Notes from each study phase are preserved through DRAFT, EVALUATE, ACCEPT, and into completed records.

---

## Edge Cases

- **Scenario:** `advance` from STUDY_SPECS without `--notes`.
  - **Expected behavior:** Transition proceeds. `specs_notes` remains empty string.
  - **Rationale:** Notes are optional. The architect may capture findings externally.

- **Scenario:** `advance --verdict FAIL` when round < `min_rounds`.
  - **Expected behavior:** Transition to REFINE (auto-refine).
  - **Rationale:** Below minimum rounds, the architect must try to fix before stopping.

- **Scenario:** `advance --verdict FAIL` when round >= `min_rounds`.
  - **Expected behavior:** Transition to REFINE.
  - **Rationale:** Past minimum, the architect refines and re-evaluates. The architect can choose to accept on the next EVALUATE pass.

- **Scenario:** Queue has a single plan item.
  - **Expected behavior:** ACCEPT transitions directly to DONE.
  - **Rationale:** No more items in queue.

- **Scenario:** `--sub-agents` set to 1.
  - **Expected behavior:** Scaffold reports 1 sub-agent in `advance` output.
  - **Rationale:** Minimum valid configuration.

- **Scenario:** `current_plan.specs` is an empty array.
  - **Expected behavior:** STUDY_SPECS still occurs. The architect studies the SPEC_MANIFEST.md for general context.
  - **Rationale:** Even without targeted specs, the manifest provides orientation.

- **Scenario:** No relevant package manifests found in codebase.
  - **Expected behavior:** STUDY_PACKAGES still occurs. The architect notes the absence and advances.
  - **Rationale:** The phase is not skipped — the architect may still find package references in CLAUDE.md or other configuration.

---

## Testing Criteria

### Init with sub-agents configuration
- **Verifies:** Initializing a Session behavior
- **Given:** `--min-rounds 1 --max-rounds 3 --sub-agents 5`
- **When:** `impl-scaffold init`
- **Then:** State file has `min_rounds: 1`, `max_rounds: 3`, `sub_agents: 5`.

### Min exceeds max rejected
- **Verifies:** Rejection table
- **Given:** `--min-rounds 5 --max-rounds 2`
- **When:** `impl-scaffold init`
- **Then:** Exit code 1.

### Sub-agents below 1 rejected
- **Verifies:** Rejection table
- **Given:** `--sub-agents 0`
- **When:** `impl-scaffold init`
- **Then:** Exit code 1.

### Study phases advance sequentially
- **Verifies:** Invariant 8
- **Given:** State is ORIENT, next plan pulled from queue
- **When:** advance through STUDY_SPECS → STUDY_CODE → STUDY_PACKAGES → SELECT → DRAFT
- **Then:** Each state transitions to the next in order. SELECT follows all study phases.

### Study notes recorded
- **Verifies:** Invariant 9, STUDY_SPECS behavior
- **Given:** State is STUDY_SPECS
- **When:** `advance --notes "Found 2 specs with shared message framing contract"`
- **Then:** `current_plan.study.specs_notes` contains the provided text.

### Study notes empty when omitted
- **Verifies:** Edge case: advance without notes
- **Given:** State is STUDY_CODE
- **When:** `advance` (no --notes)
- **Then:** `current_plan.study.code_notes` is empty string. State is STUDY_PACKAGES.

### FAIL below min_rounds auto-refines
- **Verifies:** Invariant 3, Edge case
- **Given:** `min_rounds: 2`, round 1
- **When:** `advance --verdict FAIL`
- **Then:** State is REFINE.

### FAIL at min_rounds goes to REFINE
- **Verifies:** Invariant 3
- **Given:** `min_rounds: 1`, round 1
- **When:** `advance --verdict FAIL`
- **Then:** State is REFINE.

### Deficiencies recorded on FAIL
- **Verifies:** Eval history, Invariant 7
- **Given:** State is EVALUATE
- **When:** `advance --verdict FAIL --deficiencies "Completeness,Traceability"`
- **Then:** `current_plan.evals` has an entry with `deficiencies: ["Completeness", "Traceability"]`.

### Fixed recorded on REFINE
- **Verifies:** Eval history
- **Given:** State is REFINE
- **When:** `advance --fixed "Added acceptance criteria for all testing criteria"`
- **Then:** Last eval record has `fixed: "Added acceptance criteria for all testing criteria"`.

### Eval history carried to completed
- **Verifies:** Invariant 7
- **Given:** A plan with 2 eval rounds (FAIL then PASS)
- **When:** The plan reaches ACCEPT and is moved to completed
- **Then:** `completed[].evals` has both records.

### Study notes carried to completed
- **Verifies:** Invariant 9
- **Given:** A plan with study notes populated in all three phases
- **When:** The plan reaches ACCEPT and is moved to completed
- **Then:** `completed[].study` has all three notes fields preserved.

### PASS requires message
- **Verifies:** Rejection table
- **Given:** State is EVALUATE
- **When:** `advance --verdict PASS` without `--message`
- **Then:** Exit code 1.

### Advance output includes sub-agent count and search roots
- **Verifies:** STUDY_CODE behavior
- **Given:** State is STUDY_SPECS, `sub_agents: 3`
- **When:** `impl-scaffold advance` (transitions to STUDY_CODE)
- **Then:** Output includes "Sub-agents: 3" and the code search roots.

### Full lifecycle
- **Verifies:** All states
- **Given:** Init with 1 plan, `--min-rounds 1 --max-rounds 2`
- **When:** ORIENT → STUDY_SPECS → STUDY_CODE → STUDY_PACKAGES → SELECT → DRAFT → EVALUATE(PASS) → ACCEPT → DONE
- **Then:** Completed has 1 entry with study notes and eval history. State is DONE.

### Full lifecycle with FAIL and REFINE
- **Verifies:** All states including REFINE
- **Given:** Init with 1 plan, `--min-rounds 1 --max-rounds 2`
- **When:** ORIENT → STUDY_SPECS → STUDY_CODE → STUDY_PACKAGES → SELECT → DRAFT → EVALUATE(FAIL) → REFINE → EVALUATE(PASS) → ACCEPT → DONE
- **Then:** Completed has 1 entry. Evals show FAIL then PASS.

### Queue validation rejects missing specs field
- **Verifies:** Input validation
- **Given:** Queue JSON with a plan entry missing `specs` field
- **When:** `impl-scaffold init --from queue.json --max-rounds 3`
- **Then:** Exit code 1. Error identifies the missing field.

---

## Session Archiving

Completed session state files are archived to a permanent directory:

```
<domain>/.workspace/implementation_plan/sessions/
├── protocols-2026-03-18.json
└── ...
```

- The active `impl-scaffold-state.json` is gitignored (ephemeral working state).
- Archived sessions are committed to git (permanent audit trail).
- Naming convention: `<domain>-<date>.json`.
- Archive before starting a new session. The active state file must be deleted (or the scaffold will reject `init`).

---

## Implements
- Implementation plan generation methodology from `IMPLEMENTATION_PLAN_PROMPT.md`
- TODO item: "generate the implementation planning agent's scaffold"
