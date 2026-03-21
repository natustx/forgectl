# Implementation Plan Scaffold

## Topic of Concern
> The implementation plan scaffold manages plan generation lifecycle state through a JSON-backed state machine with validated input, deterministic transitions, structured study phases, automatic plan validation, and guided evaluation that precede acceptance.

## Context

The implementation planning process involves studying specs, codebase, and packages before drafting a plan. Without persistent state, an architect loses track of study progress and plan quality across sessions. The scaffold is a Go CLI tool (built with Cobra) that reads and writes a single JSON state file, enforcing valid transitions and providing the architect with unambiguous next-step guidance.

The scaffold extends the spec generation scaffold's evaluate/refine loop with three structured study phases (STUDY_SPECS, STUDY_CODE, STUDY_PACKAGES) that build context before drafting. It tracks evaluation history — creating an audit trail from orientation to acceptance.

The DRAFT phase produces a structured JSON plan (`plan.json`) and accompanying notes files, following the format defined in `PLAN_FORMAT.md`. Validation runs automatically after DRAFT and after each REFINE cycle — if the plan's structure is valid, the scaffold advances directly to EVALUATE; if invalid, the scaffold enters a VALIDATE loop until the architect fixes the plan. The EVALUATE phase uses a dedicated evaluator prompt (`EVALUATOR_PROMPT.md`) to guide a sub-agent through comprehensive spec coverage assessment.

## Depends On
- None. The scaffold is a standalone tool with no runtime dependencies on other project components.

## Integration Points

| Component | Relationship |
|-----------|-------------|
| Implementation plan prompt | The prompt document describes the methodology; the scaffold enforces the state machine that sequences it |
| Spec generation scaffold | The spec scaffold produces completed specs; the implementation scaffold consumes those specs as input context |
| SPEC_MANIFEST.md | STUDY_SPECS reads this manifest to locate spec files relevant to the plan |
| Plan format definition (`planctl/PLAN_FORMAT.md`) | Defines the JSON schema for `plan.json` and conventions for notes files. Referenced during REVIEW, DRAFT, and validation. |
| Evaluator prompt (`planctl/EVALUATOR_PROMPT.md`) | Full instructions for the evaluation sub-agent. Output during EVALUATE so the architect can provide it to the sub-agent. |
| Evaluation sub-agent | The EVALUATE state is where the architect spawns a sub-agent using the evaluator prompt; the scaffold tracks round count, verdict, and eval report path |
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
      "file": "protocols/.workspace/implementation_plan/plan.json",
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
| `plans[].file` | string | yes | Target path for plan.json relative to project root |
| `plans[].specs` | string[] | yes | Spec file paths to study; may be empty array |
| `plans[].code_search_roots` | string[] | yes | Directory roots for codebase exploration; may be empty array |
No additional fields are permitted.

#### CLI Arguments

| Command | Flags | Description |
|---------|-------|-------------|
| `init` | `--min-rounds N` (default 1), `--max-rounds N` (required), `--from <path>` (required) | Initialize state file from a validated queue. Prints initial state after creation. |
| `advance` | `--verdict PASS\|FAIL` (EVALUATE only), `--eval-report <path>` (EVALUATE only) | Transition from current state to next. Prints the new state after transitioning. |
| `status` | none | Print full session state: current plan, eval history, queue, completed |

### Outputs

All output is to stdout. The scaffold writes state changes to `impl-scaffold-state.json`.

#### `advance` output

After transitioning, `advance` prints a structured block showing the new state and what to do next. The action text varies by state.

**Entering STUDY_SPECS** (after ORIENT):

```
State:   STUDY_SPECS
ID:      1
Plan:    Protocol Implementation
Domain:  protocols
File:    protocols/.workspace/implementation_plan/plan.json
Specs:   protocols/ws1/specs/ws1-message-contract.md, ...
Roots:   api/, optimizer/, portal/
Action:  Study the specs: protocols/ws1/specs/ws1-message-contract.md, ...
         Review git diffs for spec commits. Advance when done.
```

**Entering STUDY_CODE** (after STUDY_SPECS):

```
State:   STUDY_CODE
ID:      1
Plan:    Protocol Implementation
Domain:  protocols
File:    protocols/.workspace/implementation_plan/plan.json
Roots:   api/, optimizer/, portal/
Action:  Explore the codebase in relation to the specs under study.
         Sub-agents: 3. Search roots: api/, optimizer/, portal/.
         Advance when done.
```

**Entering STUDY_PACKAGES** (after STUDY_CODE):

```
State:   STUDY_PACKAGES
ID:      1
Plan:    Protocol Implementation
Domain:  protocols
File:    protocols/.workspace/implementation_plan/plan.json
Action:  Study the project's technical stack: package manifests, library docs, CLAUDE.md references.
         Advance when done.
```

**Entering REVIEW** (after STUDY_PACKAGES):

```
State:   REVIEW
ID:      1
Plan:    Protocol Implementation
Domain:  protocols
File:    protocols/.workspace/implementation_plan/plan.json
Action:  Review study findings before drafting.
         Plan format: .workspace/implementation_planning/planctl/PLAN_FORMAT.md
         Advance to begin drafting.
```

**Entering DRAFT** (after REVIEW):

```
State:   DRAFT
ID:      1
Plan:    Protocol Implementation
Domain:  protocols
File:    protocols/.workspace/implementation_plan/plan.json
Action:  Draft the implementation plan.
         Output: plan.json + notes/ at protocols/.workspace/implementation_plan/
         Format: .workspace/implementation_planning/planctl/PLAN_FORMAT.md
         Advance when plan and notes are ready.
```

**Entering EVALUATE** (after DRAFT or REFINE, when validation passes):

```
State:   EVALUATE
ID:      1
Plan:    Protocol Implementation
Domain:  protocols
File:    protocols/.workspace/implementation_plan/plan.json
Round:   1/3
Action:  Run evaluation sub-agent against the plan (round 1/3).
         Plan:      protocols/.workspace/implementation_plan/plan.json
         Prompt:    .workspace/implementation_planning/planctl/EVALUATOR_PROMPT.md
         Report to: protocols/.workspace/implementation_plan/evals/round-1.md
         Advance with --verdict PASS|FAIL --eval-report <path>.
```

**Entering VALIDATE** (after DRAFT or REFINE, when validation fails):

```
State:   VALIDATE
ID:      1
Plan:    Protocol Implementation
Domain:  protocols
File:    protocols/.workspace/implementation_plan/plan.json
Action:  Plan validation failed. Fix the plan and advance to re-validate.
         Format: .workspace/implementation_planning/planctl/PLAN_FORMAT.md

FAIL: 3 errors in plan.json

  items[2]: missing required field "depends_on"
    depends_on (string[]): Item IDs that must be complete before this item can begin.

  items[5]: unexpected field "status"
    status is not a valid field. Item status is computed from tests, not stored.

  layers[1].items[3]: references non-existent item "config.typez"
    Layer items must reference valid item IDs from the items array.
```

**Entering REFINE** (after EVALUATE with FAIL verdict):

```
State:   REFINE
ID:      1
Plan:    Protocol Implementation
Domain:  protocols
File:    protocols/.workspace/implementation_plan/plan.json
Round:   1/3
Action:  Evaluation found deficiencies. Spawn a sub-agent to update the plan and notes.
         Eval report: protocols/.workspace/implementation_plan/evals/round-1.md
         Advance when plan is updated.
```

**Entering REFINE** (after EVALUATE with PASS verdict, below min_rounds):

```
State:   REFINE
ID:      1
Plan:    Protocol Implementation
Domain:  protocols
File:    protocols/.workspace/implementation_plan/plan.json
Round:   1/3
Action:  Minimum evaluation rounds not met. Spawn a sub-agent to re-evaluate the plan.
         Eval report: protocols/.workspace/implementation_plan/evals/round-1.md
         Advance to proceed to next evaluation round.
```

**Entering ACCEPT** (after EVALUATE with PASS verdict, at or above min_rounds):

```
State:   ACCEPT
ID:      1
Plan:    Protocol Implementation
Domain:  protocols
File:    protocols/.workspace/implementation_plan/plan.json
Round:   2/3
Action:  Plan accepted. Advance to continue.
```

**Entering ACCEPT** (forced, after EVALUATE with FAIL verdict at max_rounds):

```
State:   ACCEPT
ID:      1
Plan:    Protocol Implementation
Domain:  protocols
File:    protocols/.workspace/implementation_plan/plan.json
Round:   3/3
Action:  Plan accepted (max rounds reached). Advance to continue.
```

**Entering DONE** (after ACCEPT, queue empty):

```
State:   DONE
Action:  All plans complete. Session done.
```

#### `status` output

Prints session config, current plan with eval history, queue, and completed plans with eval trail.

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
| `advance` called with `--verdict` outside of EVALUATE | Error naming the current state. Exit code 1. | Verdict is only valid in EVALUATE |
| `advance` called in EVALUATE without `--verdict` | Error. Exit code 1. | Verdict determines the transition |
| `advance` called in EVALUATE without `--eval-report` | Error. Exit code 1. | Every evaluation must reference its report |
| `advance` called in EVALUATE with `--eval-report` pointing to non-existent file | Error naming the path. Exit code 1. | Report must exist to be recorded |
| `advance` or `status` called before `init` | Error. Exit code 1. | State file must exist |

---

## Behavior

### Initializing a Session

#### Preconditions
- No `impl-scaffold-state.json` exists.
- `--from`, `--max-rounds` are provided.
- `--min-rounds` <= `--max-rounds`.

#### Steps
1. Read and parse the file at `--from`.
2. Validate against the queue schema.
3. If validation fails: print errors and schema, exit code 1.
4. If validation passes: create `impl-scaffold-state.json` with state ORIENT.

#### Postconditions
- State file exists with `min_rounds`, `max_rounds` set, queue populated, completed empty.

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
| STUDY_SPECS | always | STUDY_CODE | — |
| STUDY_CODE | always | STUDY_PACKAGES | — |
| STUDY_PACKAGES | always | REVIEW | — |
| REVIEW | always | DRAFT | — |
| DRAFT | plan.json valid | EVALUATE | Set round to 1. Two transitions in one advance: DRAFT passes through validation gate. |
| DRAFT | plan.json invalid | VALIDATE | Set round to 1. Validation failed; print errors. |
| VALIDATE | plan.json valid | EVALUATE | — |
| VALIDATE | plan.json invalid | _(stays VALIDATE)_ | Print validation errors. Exit code 1. |
| EVALUATE | `--verdict PASS`, round >= `min_rounds` | ACCEPT | Record eval (PASS + eval report path). |
| EVALUATE | `--verdict PASS`, round < `min_rounds` | REFINE | Record eval (PASS + eval report path). Must evaluate again — minimum rounds not met. |
| EVALUATE | `--verdict FAIL`, round < `max_rounds` | REFINE | Record eval (FAIL + eval report path). |
| EVALUATE | `--verdict FAIL`, round >= `max_rounds` | ACCEPT | Record eval (FAIL + eval report path). Maximum rounds reached — forced acceptance. |
| REFINE | plan.json valid | EVALUATE | Increment round. Two transitions in one advance: REFINE passes through validation gate. |
| REFINE | plan.json invalid | VALIDATE | Increment round. Validation failed; print errors. |
| ACCEPT | queue non-empty | ORIENT | Move plan to completed. |
| ACCEPT | queue empty | DONE | Move plan to completed. |
| DONE | — | Error: nothing to advance | Terminal |

**Validation gate.** DRAFT and REFINE do not transition directly to EVALUATE. They first run structural validation of `plan.json`. If validation passes, the scaffold transitions through to EVALUATE in a single `advance` call (two state transitions). If validation fails, the scaffold stops at VALIDATE and prints errors. This is the only case where one `advance` call may perform two transitions — it exists because validation is a mechanical check, not a phase where the architect does work. When validation passes, there is nothing for the architect to do in VALIDATE, so the scaffold skips it.

#### Postconditions
- State file reflects the new state.
- Eval records accumulate on `current_plan.evals` and carry to `completed[].evals`.

#### Error Handling
- Invalid flags for state: specific error per state.
- Invalid verdict value: error.
- Plan validation failure: error with details. State becomes or remains VALIDATE.

---

### Study Phases

The three study phases build context before drafting. Each phase has a specific focus. No flags are required — the architect studies, then advances.

#### STUDY_SPECS

The architect studies the specs listed in `current_plan.specs` and the SPEC_MANIFEST.md. This includes:
- Reading the full spec files
- Reviewing git diffs associated with the specs (commits that introduced or modified them)
- Understanding dependencies, integration points, and cross-references

#### STUDY_CODE

The architect explores the codebase in relation to the specs under study, using sub-agents. The sub-agent count is hardcoded to 3.

The sub-agents search within the directories listed in `current_plan.code_search_roots`. The scaffold does not launch sub-agents itself — it provides the configuration and the architect orchestrates them.

#### STUDY_PACKAGES

The architect studies the project's technical stack:
- Package manifest files in the codebase (go.mod, pyproject.toml, package.json)
- Library documentation via Context7 or other sources as referenced in CLAUDE.md
- Any additional package references in project configuration

---

### REVIEW Phase

The REVIEW phase is a lightweight checkpoint before drafting. The scaffold outputs the path to `planctl/PLAN_FORMAT.md` so the architect can reference the expected plan structure. The architect reviews their study findings (held externally or in memory) and the plan format, then advances to DRAFT.

---

### DRAFT Phase

The DRAFT phase is where the architect generates the implementation plan as a structured JSON file with accompanying notes.

#### Expected Output

The architect produces:

```
<domain>/.workspace/implementation_plan/
├── plan.json          # The implementation plan manifest
└── notes/             # Reference notes per package
    ├── <package>.md
    └── ...
```

The plan format is defined in `planctl/PLAN_FORMAT.md`. The scaffold outputs this path in the DRAFT action description.

When the architect advances from DRAFT, the scaffold automatically validates `plan.json` before proceeding to EVALUATE. See **Validation Gate** below.

---

### Validation Gate

Validation is not a phase where the architect does work — it is a structural check that fires automatically when advancing from DRAFT or REFINE.

#### When It Runs

1. **After DRAFT advance**: The scaffold reads `plan.json` at `current_plan.file` and validates its structure.
2. **After REFINE advance**: Same validation runs again, since the architect may have modified the plan's structure while addressing evaluation deficiencies.

#### Validation Checks

| Check | Description |
|-------|-------------|
| JSON parse | File exists and contains valid JSON |
| Top-level fields | `context`, `refs`, `layers`, `items` are present and correctly typed |
| Context fields | `domain` and `module` are non-empty strings |
| Refs exist | Every path in `refs` resolves to an existing file |
| Item schema | Every item has `id`, `name`, `description`, `depends_on`, `tests` |
| Item ID uniqueness | No duplicate item IDs |
| Layer coverage | Every item appears in exactly one layer; every layer item ID exists |
| Layer ordering | Items only depend on items in equal or earlier layers |
| DAG validity | `depends_on` references are valid item IDs; no dependency cycles |
| Test schema | Every test has `category`, `description`, `passes` with correct types |
| Test categories | Categories are one of: `functional`, `rejection`, `edge_case` |
| Notes files | Every `ref` in items resolves to an existing notes file |

#### On Pass

The scaffold transitions directly to EVALUATE. The VALIDATE state is never visible to the architect. Output shows the EVALUATE action description.

#### On Fail

The scaffold enters the VALIDATE state and prints every validation error with the path to the offending location and a description of what the field is for (derived from `PLAN_FORMAT.md`). The architect fixes the plan and runs `advance` again. This loops until the plan is valid.

```
FAIL: 3 errors in plan.json

  items[2]: missing required field "depends_on"
    depends_on (string[]): Item IDs that must be complete before this item can begin.

  items[5]: unexpected field "status"
    status is not a valid field. Item status is computed from tests, not stored.

  layers[1].items[3]: references non-existent item "config.typez"
    Layer items must reference valid item IDs from the items array.
```

---

### EVALUATE Phase

The EVALUATE phase is where the architect spawns a sub-agent to assess whether the plan covers the specs.

#### Action Description

The scaffold outputs:
1. The path to the plan: `current_plan.file`.
2. The path to the evaluator prompt: `planctl/EVALUATOR_PROMPT.md`.
3. The target path for the evaluation report: `<domain>/.workspace/implementation_plan/evals/round-N.md` (where N is `current_plan.round`).
4. Instructions to advance with `--verdict PASS|FAIL --eval-report <path>`.

The evaluator prompt file contains the full instructions for the sub-agent: what files to read, the 11 evaluation dimensions, the report format, and the verdict rules.

#### Advancing from EVALUATE

The architect provides:
- `--verdict PASS` or `--verdict FAIL` — the verdict from the evaluation.
- `--eval-report <path>` — the path to the evaluation report file. The scaffold verifies the file exists and stores the path on the eval record. The scaffold does not read or parse the report contents.

The transition depends on the verdict and the current round:

| Verdict | Condition | To State | Rationale |
|---------|-----------|----------|-----------|
| PASS | round >= `min_rounds` | ACCEPT | Plan accepted. |
| PASS | round < `min_rounds` | REFINE | Minimum rounds not met. Must evaluate again to prevent false confidence from a single pass. |
| FAIL | round < `max_rounds` | REFINE | More rounds available. Architect addresses deficiencies. |
| FAIL | round >= `max_rounds` | ACCEPT | Maximum rounds exhausted. Forced acceptance — accept what you have and move on. |

---

### REFINE Phase

The REFINE phase tells the architect to spawn a sub-agent to update the plan. The scaffold outputs the eval report path from the previous EVALUATE round. The action description varies based on the previous verdict:

- **After FAIL**: "Evaluation found deficiencies. Spawn a sub-agent to update the plan and notes." The sub-agent reads the eval report to understand what to fix.
- **After PASS below min_rounds**: "Minimum evaluation rounds not met. Spawn a sub-agent to re-evaluate the plan." The sub-agent reviews the plan for any improvements before the next evaluation round.

When the architect advances from REFINE, the scaffold runs the validation gate before proceeding to EVALUATE. See **Validation Gate**.

---

## Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `--min-rounds` | integer | 1 | Minimum evaluation rounds required. PASS before this round forces another REFINE cycle. |
| `--max-rounds` | integer | none (required) | Maximum evaluation rounds allowed. FAIL at this round forces acceptance. |
| `--from` | string | none (required on init) | Path to queue input JSON file |

`min_rounds` and `max_rounds` enforce hard bounds on the evaluate/refine loop:
- **`min_rounds`**: Prevents premature acceptance. A PASS verdict below `min_rounds` sends the plan back to REFINE for another evaluation round. This guards against false confidence from a single evaluation pass.
- **`max_rounds`**: Caps effort. A FAIL verdict at `max_rounds` forces acceptance — the architect has tried enough and should move on with the best plan available.

---

## Reference Files

The scaffold ships with two reference documents that are output during specific phases:

| File | Location | Used In | Purpose |
|------|----------|---------|---------|
| Plan Format | `planctl/PLAN_FORMAT.md` | REVIEW, DRAFT, Validation Gate | JSON schema, item structure, test conventions, notes file organization |
| Evaluator Prompt | `planctl/EVALUATOR_PROMPT.md` | EVALUATE | Full instructions for the evaluation sub-agent: dimensions, report format, verdict rules |

These files are part of the scaffold codebase and are referenced by path in the action descriptions of their respective phases.

---

## State File Schema

```json
{
  "min_rounds": 1,
  "max_rounds": 3,
  "state": "EVALUATE",
  "current_plan": {
    "id": 1,
    "name": "Protocol Implementation",
    "domain": "protocols",
    "topic": "Implementation plan for WS1 and WS2 message contract specs",
    "file": "protocols/.workspace/implementation_plan/plan.json",
    "specs": [
      "protocols/ws1/specs/ws1-message-contract.md",
      "protocols/ws2/specs/ws2-message-contract.md"
    ],
    "code_search_roots": ["api/", "optimizer/", "portal/"],
    "round": 1,
    "evals": []
  },
  "queue": [],
  "completed": [
    {
      "id": 0,
      "name": "Previous Plan",
      "domain": "optimizer",
      "file": "optimizer/.workspace/implementation_plan/plan.json",
      "rounds_taken": 2,
      "evals": [
        { "round": 1, "verdict": "FAIL", "eval_report": "optimizer/.workspace/implementation_plan/evals/round-1.md" },
        { "round": 2, "verdict": "PASS", "eval_report": "optimizer/.workspace/implementation_plan/evals/round-2.md" }
      ]
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `min_rounds` | integer | Minimum eval rounds required. PASS below this forces REFINE. |
| `max_rounds` | integer | Maximum eval rounds allowed. FAIL at this forces ACCEPT. |
| `state` | string | ORIENT, STUDY_SPECS, STUDY_CODE, STUDY_PACKAGES, REVIEW, DRAFT, VALIDATE, EVALUATE, REFINE, ACCEPT, DONE |
| `current_plan.round` | integer | Current evaluation round (0 before first EVALUATE entry) |
| `current_plan.evals` | array | Evaluation history for the current plan |
| `current_plan.evals[].round` | integer | Which round this eval was |
| `current_plan.evals[].verdict` | string | PASS or FAIL |
| `current_plan.evals[].eval_report` | string | Path to the evaluation report file |
| `completed[].evals` | array | Full eval history carried from current_plan |

---

## Invariants

1. **Single active plan.** At most one plan is in `current_plan` at any time.
2. **Round monotonicity.** The round counter only increments.
3. **Min rounds enforced.** PASS below `min_rounds` goes to REFINE, not ACCEPT. The plan must survive at least `min_rounds` evaluations before acceptance.
4. **Max rounds enforced.** FAIL at `max_rounds` goes to ACCEPT (forced). The architect cannot refine indefinitely.
5. **Queue order preserved.** Plans are pulled from the front of the queue.
6. **State file is the only mutable artifact.** The queue input file is read once at init.
7. **No implicit state.** All information for transitions is in the state file.
8. **Eval history is append-only.** Eval records accumulate and are never deleted or modified.
9. **Study phases precede REVIEW.** STUDY_SPECS → STUDY_CODE → STUDY_PACKAGES → REVIEW. No phase is skipped.
10. **Validation precedes evaluation.** The validation gate runs before every EVALUATE entry. The scaffold will not advance to EVALUATE if plan.json has structural errors.
11. **Eval reports must exist.** The scaffold verifies the `--eval-report` file exists before recording it.

---

## Edge Cases

- **Scenario:** `advance --verdict PASS` when round < `min_rounds`.
  - **Expected behavior:** Transition to REFINE, not ACCEPT.
  - **Rationale:** Minimum rounds not met. A single PASS may be false confidence (see launcher session lesson). The plan must be evaluated again.

- **Scenario:** `advance --verdict PASS` when round >= `min_rounds`.
  - **Expected behavior:** Transition to ACCEPT.
  - **Rationale:** Minimum evaluation threshold met. Plan accepted.

- **Scenario:** `advance --verdict FAIL` when round < `max_rounds`.
  - **Expected behavior:** Transition to REFINE.
  - **Rationale:** More evaluation rounds available. Architect addresses deficiencies.

- **Scenario:** `advance --verdict FAIL` when round >= `max_rounds`.
  - **Expected behavior:** Transition to ACCEPT (forced).
  - **Rationale:** Maximum rounds exhausted. Accept the best available plan and move on.

- **Scenario:** Queue has a single plan item.
  - **Expected behavior:** ACCEPT transitions directly to DONE.
  - **Rationale:** No more items in queue.

- **Scenario:** `current_plan.specs` is an empty array.
  - **Expected behavior:** STUDY_SPECS still occurs. The architect studies the SPEC_MANIFEST.md for general context.
  - **Rationale:** Even without targeted specs, the manifest provides orientation.

- **Scenario:** Validation passes on first try after DRAFT.
  - **Expected behavior:** Scaffold transitions directly from DRAFT to EVALUATE in one `advance` call. VALIDATE state is never visible.
  - **Rationale:** Validation is a gate, not a phase. When the plan is valid, there is nothing for the architect to do in VALIDATE.

- **Scenario:** Validation fails after DRAFT.
  - **Expected behavior:** Scaffold enters VALIDATE state. Prints errors. Architect fixes plan and advances. Loops until valid, then transitions to EVALUATE.
  - **Rationale:** VALIDATE becomes visible only when there are errors to fix.

- **Scenario:** Validation fails after REFINE.
  - **Expected behavior:** Same as after DRAFT — enters VALIDATE loop. Architect fixes structural issues, then proceeds to EVALUATE.
  - **Rationale:** REFINE edits may introduce structural errors (broken references, malformed JSON).

- **Scenario:** plan.json does not exist when validation runs.
  - **Expected behavior:** Validation fails with error: file not found at path. Enters VALIDATE state.
  - **Rationale:** The plan must be written before it can be validated.

- **Scenario:** plan.json has a dependency cycle.
  - **Expected behavior:** Validation fails listing the cycle. Enters VALIDATE state.
  - **Rationale:** DAG must be acyclic for layer ordering to be meaningful.

- **Scenario:** `--eval-report` points to a non-existent file.
  - **Expected behavior:** Error. Exit code 1. State unchanged.
  - **Rationale:** The report must exist to be recorded. Prevents recording phantom paths.

---

## Testing Criteria

### Init with configuration
- **Verifies:** Initializing a Session behavior
- **Given:** `--min-rounds 1 --max-rounds 3`
- **When:** `planctl init --from queue.json`
- **Then:** State file has `min_rounds: 1`, `max_rounds: 3`.

### Min exceeds max rejected
- **Verifies:** Rejection table
- **Given:** `--min-rounds 5 --max-rounds 2`
- **When:** `planctl init`
- **Then:** Exit code 1.

### Study phases advance sequentially
- **Verifies:** Invariant 7
- **Given:** State is ORIENT, next plan pulled from queue
- **When:** advance through STUDY_SPECS → STUDY_CODE → STUDY_PACKAGES → REVIEW → DRAFT
- **Then:** Each state transitions to the next in order. REVIEW follows all study phases.

### STUDY_CODE output references specs
- **Verifies:** STUDY_CODE behavior
- **Given:** State is STUDY_SPECS
- **When:** `planctl advance` (transitions to STUDY_CODE)
- **Then:** Output includes "in relation to the specs under study", sub-agent count of 3, and code search roots.

### REVIEW outputs plan format path
- **Verifies:** REVIEW phase behavior
- **Given:** State is STUDY_PACKAGES
- **When:** `planctl advance` (transitions to REVIEW)
- **Then:** Output includes path to `planctl/PLAN_FORMAT.md`. Output does NOT include study notes.

### DRAFT outputs format and target paths
- **Verifies:** DRAFT phase behavior
- **Given:** State is REVIEW
- **When:** `planctl advance` (transitions to DRAFT)
- **Then:** Output includes plan.json path, notes directory, and path to `planctl/PLAN_FORMAT.md`.

### DRAFT advance with valid plan goes directly to EVALUATE
- **Verifies:** Validation gate, edge case
- **Given:** State is DRAFT, plan.json is structurally valid
- **When:** `planctl advance`
- **Then:** State is EVALUATE (not VALIDATE). Round is 1. Output shows EVALUATE action description.

### DRAFT advance with invalid plan enters VALIDATE
- **Verifies:** Validation gate
- **Given:** State is DRAFT, plan.json has structural errors
- **When:** `planctl advance`
- **Then:** State is VALIDATE. Output shows validation errors with field descriptions.

### VALIDATE loops until plan is valid
- **Verifies:** Validation gate loop
- **Given:** State is VALIDATE, plan.json still has errors
- **When:** `planctl advance`
- **Then:** State remains VALIDATE. Errors printed again. Exit code 1.

### VALIDATE passes and transitions to EVALUATE
- **Verifies:** Validation gate
- **Given:** State is VALIDATE, plan.json has been fixed and is now valid
- **When:** `planctl advance`
- **Then:** State is EVALUATE. Output shows EVALUATE action description.

### VALIDATE rejects missing plan file
- **Verifies:** Validation checks
- **Given:** State is VALIDATE, plan.json does not exist at `current_plan.file`
- **When:** `planctl advance`
- **Then:** Exit code 1. Error names the missing path. State remains VALIDATE.

### VALIDATE rejects duplicate item IDs
- **Verifies:** Validation checks
- **Given:** State is VALIDATE, plan.json has two items with id "config.load"
- **When:** `planctl advance`
- **Then:** Exit code 1. Error identifies the duplicate. State remains VALIDATE.

### VALIDATE rejects broken depends_on reference
- **Verifies:** Validation checks
- **Given:** State is VALIDATE, an item depends on "config.typez" which doesn't exist
- **When:** `planctl advance`
- **Then:** Exit code 1. Error identifies the bad reference. State remains VALIDATE.

### VALIDATE rejects item not in any layer
- **Verifies:** Validation checks
- **Given:** State is VALIDATE, an item exists in `items` but not in any layer
- **When:** `planctl advance`
- **Then:** Exit code 1. Error identifies the orphaned item. State remains VALIDATE.

### VALIDATE rejects cross-layer dependency violation
- **Verifies:** Validation checks
- **Given:** State is VALIDATE, an L0 item depends on an L1 item
- **When:** `planctl advance`
- **Then:** Exit code 1. Error identifies the backward dependency. State remains VALIDATE.

### VALIDATE rejects missing notes file
- **Verifies:** Validation checks
- **Given:** State is VALIDATE, an item's `ref` points to a non-existent notes file
- **When:** `planctl advance`
- **Then:** Exit code 1. Error identifies the missing file. State remains VALIDATE.

### VALIDATE rejects missing refs file
- **Verifies:** Validation checks
- **Given:** State is VALIDATE, a path in `refs` does not exist on disk
- **When:** `planctl advance`
- **Then:** Exit code 1. Error identifies the missing ref. State remains VALIDATE.

### EVALUATE outputs evaluator prompt and report paths
- **Verifies:** EVALUATE phase behavior
- **Given:** State is EVALUATE, round 1
- **When:** (entered via DRAFT advance with valid plan)
- **Then:** Output includes plan path, path to `planctl/EVALUATOR_PROMPT.md`, and report target `evals/round-1.md`.

### EVALUATE requires verdict
- **Verifies:** Rejection table
- **Given:** State is EVALUATE
- **When:** `planctl advance` (no --verdict)
- **Then:** Exit code 1.

### EVALUATE requires eval-report
- **Verifies:** Rejection table
- **Given:** State is EVALUATE
- **When:** `planctl advance --verdict PASS` (no --eval-report)
- **Then:** Exit code 1.

### EVALUATE rejects non-existent eval-report
- **Verifies:** Invariant 9
- **Given:** State is EVALUATE
- **When:** `planctl advance --verdict PASS --eval-report evals/missing.md`
- **Then:** Exit code 1. Error names the missing path.

### EVALUATE PASS at min_rounds transitions to ACCEPT
- **Verifies:** Invariant 3, State transitions
- **Given:** State is EVALUATE, `min_rounds: 1`, round 1
- **When:** `planctl advance --verdict PASS --eval-report evals/round-1.md`
- **Then:** State is ACCEPT. Eval record has verdict PASS and eval_report path.

### EVALUATE PASS below min_rounds transitions to REFINE
- **Verifies:** Invariant 3
- **Given:** State is EVALUATE, `min_rounds: 2`, round 1
- **When:** `planctl advance --verdict PASS --eval-report evals/round-1.md`
- **Then:** State is REFINE, not ACCEPT. Eval record has verdict PASS.

### EVALUATE FAIL below max_rounds transitions to REFINE
- **Verifies:** State transitions
- **Given:** State is EVALUATE, `max_rounds: 3`, round 1
- **When:** `planctl advance --verdict FAIL --eval-report evals/round-1.md`
- **Then:** State is REFINE.

### EVALUATE FAIL at max_rounds forces ACCEPT
- **Verifies:** Invariant 4
- **Given:** State is EVALUATE, `max_rounds: 2`, round 2
- **When:** `planctl advance --verdict FAIL --eval-report evals/round-2.md`
- **Then:** State is ACCEPT (forced). Eval record has verdict FAIL.

### REFINE outputs eval report path
- **Verifies:** REFINE phase behavior
- **Given:** State is REFINE, last eval has eval_report path
- **When:** (entered via EVALUATE FAIL)
- **Then:** Output includes the eval report path from the previous round.

### REFINE advance with valid plan goes to EVALUATE
- **Verifies:** Validation gate after REFINE
- **Given:** State is REFINE, plan.json is structurally valid
- **When:** `planctl advance`
- **Then:** State is EVALUATE. Round incremented.

### REFINE advance with invalid plan enters VALIDATE
- **Verifies:** Validation gate after REFINE
- **Given:** State is REFINE, plan.json has structural errors
- **When:** `planctl advance`
- **Then:** State is VALIDATE. Errors printed.

### Eval history carried to completed
- **Verifies:** Invariant 6
- **Given:** A plan with 2 eval rounds (FAIL then PASS)
- **When:** The plan reaches ACCEPT and is moved to completed
- **Then:** `completed[].evals` has both records with verdict and eval_report paths.

### Full lifecycle with single round
- **Verifies:** All states
- **Given:** Init with 1 plan, `--min-rounds 1 --max-rounds 2`
- **When:** ORIENT → STUDY_SPECS → STUDY_CODE → STUDY_PACKAGES → REVIEW → DRAFT → [validation passes] → EVALUATE(PASS, round 1) → ACCEPT → DONE
- **Then:** Completed has 1 entry with eval history. State is DONE.

### Full lifecycle with FAIL and REFINE
- **Verifies:** All states including REFINE
- **Given:** Init with 1 plan, `--min-rounds 1 --max-rounds 3`
- **When:** ORIENT → STUDY_SPECS → STUDY_CODE → STUDY_PACKAGES → REVIEW → DRAFT → [validation passes] → EVALUATE(FAIL, round 1) → REFINE → [validation passes] → EVALUATE(PASS, round 2) → ACCEPT → DONE
- **Then:** Completed has 1 entry. Evals show FAIL then PASS with report paths.

### Full lifecycle with min_rounds enforcement
- **Verifies:** Invariant 3
- **Given:** Init with 1 plan, `--min-rounds 2 --max-rounds 3`
- **When:** ORIENT → ... → EVALUATE(PASS, round 1) → REFINE → [validation passes] → EVALUATE(PASS, round 2) → ACCEPT → DONE
- **Then:** First PASS at round 1 did NOT go to ACCEPT. Second PASS at round 2 did. Evals show two PASS records.

### Full lifecycle with max_rounds forced acceptance
- **Verifies:** Invariant 4
- **Given:** Init with 1 plan, `--min-rounds 1 --max-rounds 2`
- **When:** ORIENT → ... → EVALUATE(FAIL, round 1) → REFINE → [validation passes] → EVALUATE(FAIL, round 2) → ACCEPT → DONE
- **Then:** Second FAIL at max_rounds forced ACCEPT. Evals show two FAIL records. Plan still moved to completed.

### Queue validation rejects missing specs field
- **Verifies:** Input validation
- **Given:** Queue JSON with a plan entry missing `specs` field
- **When:** `planctl init --from queue.json --max-rounds 3`
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
