# Plan Production

## Topic of Concern
> The scaffold produces a validated implementation plan through a study-draft-evaluate cycle.

## Context

The planning phase guides the architect through studying specs, codebase, and packages, then drafting an implementation plan as structured JSON (`plan.json`) with accompanying notes. The plan is validated against a structural schema and evaluated through iterative sub-agent review rounds until accepted (or force-accepted at max rounds).

## Depends On
- **phase-transitions** — the specifying→planning phase shift provides the plans queue and triggers ORIENT.
- **session-init** — alternatively, `init --phase planning` starts the planning phase directly.
- **state-persistence** — reads and writes the state file.

## Integration Points

| Spec | Relationship |
|------|-------------|
| SPEC_MANIFEST.md | STUDY_SPECS reads this manifest to locate spec files relevant to the plan |
| Plan format definition (`PLAN_FORMAT.md`) | Defines the JSON schema for `plan.json` and conventions for notes files. Referenced during REVIEW, DRAFT, and validation. |
| Plan evaluator prompt (`evaluators/plan-eval.md`) | Full instructions for the planning evaluation sub-agent: dimensions, report format, verdict rules |
| phase-transitions | ACCEPT transitions to PHASE_SHIFT (planning → implementing) |

---

## Interface

### Inputs

#### `advance` flags — Planning Phase

| State | Flags |
|-------|-------|
| EVALUATE | `--verdict PASS\|FAIL`, `--eval-report <path>` (both required) |
| ACCEPT | `--message <text>` (required when `enable_commits: true`) |

#### `eval` command

| Command | Flags | Description |
|---------|-------|-------------|
| `eval` | none | Output full evaluation context for the sub-agent. Only valid in planning EVALUATE. |

### Outputs

#### `advance` output

**Entering ORIENT** (after PHASE_SHIFT or init with `--phase planning`):

```
State:   ORIENT
Phase:   planning
Plan:    Service Configuration
Domain:  launcher
File:    launcher/.forge_workspace/implementation_plan/plan.json
Action:  After completion of the above, advance to begin studying specs.
```

**Entering STUDY_SPECS** (after ORIENT):

```
State:   STUDY_SPECS
Phase:   planning
Plan:    Service Configuration
Domain:  launcher
File:    launcher/.forge_workspace/implementation_plan/plan.json
Specs:   launcher/specs/service-configuration.md, ...
Roots:   launcher/, api/
Action:  Study the specs: launcher/specs/service-configuration.md, ...
         Review git diffs for spec commits.
         After completion of the above, advance to continue.
```

**Entering STUDY_CODE** (after STUDY_SPECS):

```
State:   STUDY_CODE
Phase:   planning
Plan:    Service Configuration
Domain:  launcher
File:    launcher/.forge_workspace/implementation_plan/plan.json
Roots:   launcher/, api/
Action:  Please spawn 3 haiku sub-agents to explore the codebase.
         Search roots: launcher/, api/.
         After completion of the above, advance to continue.
```

**Entering STUDY_PACKAGES** (after STUDY_CODE):

```
State:   STUDY_PACKAGES
Phase:   planning
Plan:    Service Configuration
Domain:  launcher
File:    launcher/.forge_workspace/implementation_plan/plan.json
Action:  Study the project's technical stack: package manifests, library docs, CLAUDE.md references.
         After completion of the above, advance to continue.
```

**Entering REVIEW** (after STUDY_PACKAGES):

```
State:   REVIEW
Phase:   planning
Plan:    Service Configuration
Domain:  launcher
File:    launcher/.forge_workspace/implementation_plan/plan.json
Action:  Review study findings before drafting.
         Plan format: PLAN_FORMAT.md
         STOP please review and discuss with user before continuing.
         After completion of the above, advance to begin drafting.
```

**Entering DRAFT** (after REVIEW):

```
State:   DRAFT
Phase:   planning
Plan:    Service Configuration
Domain:  launcher
File:    launcher/.forge_workspace/implementation_plan/plan.json
Action:  Draft the implementation plan.
         Output: plan.json + notes/ at launcher/.forge_workspace/implementation_plan/
         Format: PLAN_FORMAT.md
         After completion of the above, advance to validate.
```

**Entering VALIDATE** (after DRAFT or REFINE, when validation fails):

```
State:   VALIDATE
Phase:   planning
Plan:    Service Configuration
Domain:  launcher
File:    launcher/.forge_workspace/implementation_plan/plan.json
Action:  Plan validation failed. Fix the plan.
         After completion of the above, advance to re-validate.
         Format: PLAN_FORMAT.md

FAIL: 3 errors in plan.json

  items[2]: missing required field "depends_on"
    depends_on (string[]): Item IDs that must be complete before this item can begin.

  items[5]: unexpected field "status"
    status is not a valid field. Item status is computed from tests, not stored.

  layers[1].items[3]: references non-existent item "config.typez"
    Layer items must reference valid item IDs from the items array.
```

**Entering EVALUATE** (after DRAFT or REFINE, when validation passes):

```
State:   EVALUATE
Phase:   planning
Plan:    Service Configuration
Domain:  launcher
File:    launcher/.forge_workspace/implementation_plan/plan.json
Round:   1/3
Action:  Please spawn 1 opus sub-agent to evaluate the plan.
         Sub-agent runs: forgectl eval
         After completion of the above, advance with --verdict PASS|FAIL --eval-report <path>
```

**Entering REFINE** (after EVALUATE with FAIL verdict):

```
State:   REFINE
Phase:   planning
Plan:    Service Configuration
Domain:  launcher
File:    launcher/.forge_workspace/implementation_plan/plan.json
Round:   1/3
Action:  Evaluation found deficiencies. Please spawn 1 opus sub-agent to update the plan and notes.
         Eval report: launcher/.forge_workspace/implementation_plan/evals/round-1.md
         After completion of the above, advance to continue.
```

**Entering REFINE** (after EVALUATE with PASS verdict, below min_rounds):

```
State:   REFINE
Phase:   planning
Plan:    Service Configuration
Domain:  launcher
File:    launcher/.forge_workspace/implementation_plan/plan.json
Round:   1/3
Action:  Minimum evaluation rounds not met. Please spawn 1 opus sub-agent to re-evaluate the plan.
         Eval report: launcher/.forge_workspace/implementation_plan/evals/round-1.md
         After completion of the above, advance to proceed to next evaluation round.
```

**Entering ACCEPT** (after EVALUATE with PASS verdict, at or above min_rounds):

```
State:   ACCEPT
Phase:   planning
Plan:    Service Configuration
Domain:  launcher
File:    launcher/.forge_workspace/implementation_plan/plan.json
Round:   2/3
Action:  Plan accepted.
         After completion of the above, advance to continue.
```

**Entering ACCEPT** (forced, after EVALUATE with FAIL verdict at max_rounds):

```
State:   ACCEPT
Phase:   planning
Plan:    Service Configuration
Domain:  launcher
File:    launcher/.forge_workspace/implementation_plan/plan.json
Round:   3/3
Action:  Plan accepted (max rounds reached).
         After completion of the above, advance to continue.
```

#### `eval` output

```
=== PLAN EVALUATION ROUND 1/3 ===
Plan:   Service Configuration
Domain: launcher
File:   launcher/.forge_workspace/implementation_plan/plan.json

--- EVALUATOR INSTRUCTIONS ---

<contents of evaluators/plan-eval.md>

--- PLAN REFERENCES ---

Plan:    launcher/.forge_workspace/implementation_plan/plan.json
Format:  PLAN_FORMAT.md
Specs:
  - launcher/specs/service-configuration.md

--- REPORT OUTPUT ---

Write your evaluation report to:
  launcher/.forge_workspace/implementation_plan/evals/round-1.md
```

Subsequent rounds include previous evaluations:

```
=== PLAN EVALUATION ROUND 2/3 ===
...

--- PREVIOUS EVALUATIONS ---

Round 1: FAIL — launcher/.forge_workspace/implementation_plan/evals/round-1.md

--- REPORT OUTPUT ---

Write your evaluation report to:
  launcher/.forge_workspace/implementation_plan/evals/round-2.md
```

#### Eval Report Locations

Planning eval reports:
```
<domain>/.forge_workspace/implementation_plan/evals/round-N.md
```

#### `status` output — Planning (compact)

The compact `status` output for planning shows the current plan, round, action, and progress:

```
Plan:    Service Configuration (launcher)
File:    launcher/.forge_workspace/implementation_plan/plan.json
Round:   1/3

Action:  Run evaluation sub-agent against the plan (round 1/3).
         Sub-agent: forgectl eval
         Advance with --verdict PASS|FAIL --eval-report <path>.

Progress: round 1 of 3
```

#### `status --verbose` output — Planning section

With `--verbose`, the eval history is appended:

```
--- Planning ---

  Evals: (none yet)
```

Or when complete:

```
--- Planning ---

  Accepted (2 rounds)
    Round 1: FAIL — launcher/.forge_workspace/implementation_plan/evals/round-1.md
    Round 2: PASS — launcher/.forge_workspace/implementation_plan/evals/round-2.md
```

### Rejection

| Condition | Signal | Rationale |
|-----------|--------|-----------|
| `advance` in planning EVALUATE without `--verdict` | Error. Exit code 1. | Verdict determines the transition |
| `advance` in planning EVALUATE without `--eval-report` | Error. Exit code 1. | Every evaluation must reference its report |
| `advance --eval-report` pointing to non-existent file | Error naming the path. Exit code 1. | Report must exist to be recorded |
| `advance` in planning ACCEPT without `--message` when `enable_commits: true` | Error. Exit code 1. | Accepted plans need a commit message when commits are enabled |
| `eval` outside of planning EVALUATE | Error naming current state and phase. Exit code 1. | Eval context only available in EVALUATE |

---

## Behavior

### State Machine

```
ORIENT → STUDY_SPECS → STUDY_CODE → STUDY_PACKAGES → REVIEW → DRAFT
                                                                  │
                                                        ┌─────────┴─────────┐
                                                   plan valid          plan invalid
                                                        │                   │
                                                        ▼                   ▼
                                                   EVALUATE            VALIDATE
                                                        │                   │
                                              ┌─────────┼─────────┐   fix + advance
                                              │         │         │        │
                                        PASS ≥ min  PASS < min  FAIL < max │
                                              │         │         │        │
                                              ▼         ▼         ▼        │
                                           ACCEPT    REFINE    REFINE ◄────┘
                                              │         │         │
                                              ▼         └────┬────┘
                                        PHASE_SHIFT          │
                                                        plan valid → EVALUATE
                                                        plan invalid → VALIDATE

                                        FAIL ≥ max → ACCEPT (forced)
```

### Transition Table

| From State | Condition | To State | Side Effects |
|------------|-----------|----------|-------------|
| ORIENT | always | STUDY_SPECS | — |
| STUDY_SPECS | always | STUDY_CODE | — |
| STUDY_CODE | always | STUDY_PACKAGES | — |
| STUDY_PACKAGES | always | REVIEW | — |
| REVIEW | always | DRAFT | — |
| DRAFT | plan.json valid | EVALUATE | Set round to 1. Two transitions in one advance. |
| DRAFT | plan.json invalid | VALIDATE | Set round to 1. Print errors. |
| VALIDATE | plan.json valid | EVALUATE | — |
| VALIDATE | plan.json invalid | _(stays VALIDATE)_ | Print errors. Exit code 1. |
| EVALUATE | `--verdict PASS`, round >= `planning.eval.min_rounds` | ACCEPT | Record eval. |
| EVALUATE | `--verdict PASS`, round < `planning.eval.min_rounds` | REFINE | Record eval. Min rounds not met. |
| EVALUATE | `--verdict FAIL`, round < `planning.eval.max_rounds` | REFINE | Record eval. |
| EVALUATE | `--verdict FAIL`, round >= `planning.eval.max_rounds` | ACCEPT | Record eval. Forced acceptance. |
| REFINE | plan.json valid | EVALUATE | Increment round. Two transitions in one advance. |
| REFINE | plan.json invalid | VALIDATE | Increment round. Print errors. |
| ACCEPT | always | PHASE_SHIFT | Set phase shift from planning → implementing. |

### Study Phases

Three study phases build context before drafting. No flags required — the architect studies, then advances.

#### STUDY_SPECS
Study the specs listed in `current_plan.specs` and the SPEC_MANIFEST.md: full spec files, git diffs, dependencies, cross-references.

#### STUDY_CODE
Explore the codebase using sub-agents (count hardcoded to 3) within `current_plan.code_search_roots`.

#### STUDY_PACKAGES
Study the project's technical stack: package manifests, library documentation, CLAUDE.md references.

### REVIEW Phase

Lightweight checkpoint before drafting. Outputs the path to `PLAN_FORMAT.md`. The architect reviews study findings and the plan format, then advances to DRAFT.

### DRAFT Phase

The architect generates the implementation plan as structured JSON with accompanying notes:

```
<domain>/.forge_workspace/implementation_plan/
├── plan.json
└── notes/
    ├── <package>.md
    └── ...
```

### Validation Gate

Fires automatically when advancing from DRAFT or REFINE. Not a phase where the architect does work.

#### Validation Checks

| Check | Description |
|-------|-------------|
| JSON parse | File exists and contains valid JSON |
| Top-level fields | `context`, `refs`, `layers`, `items` present and correctly typed |
| Context fields | `domain` and `module` are non-empty strings |
| Refs exist | Every path in `refs` resolves to an existing file |
| Item schema | Every item has `id`, `name`, `description`, `depends_on`, `tests` |
| Item ID uniqueness | No duplicate item IDs |
| Layer coverage | Every item in exactly one layer; every layer item ID exists |
| Layer ordering | Items only depend on items in equal or earlier layers |
| DAG validity | `depends_on` references are valid; no cycles |
| Test schema | Every test has `category`, `description`, `passes` with correct types |
| Test categories | One of: `functional`, `rejection`, `edge_case` |
| Notes files | Every `ref` in items resolves to an existing notes file |

**On pass:** transitions directly to EVALUATE. VALIDATE is never visible.

**On fail:** enters VALIDATE, prints errors with field descriptions. Loops until valid.

### EVALUATE Phase

Uses `eval` command to output full evaluation context for the sub-agent. The evaluator prompt (`evaluators/plan-eval.md`) defines 11 assessment dimensions.

### REFINE Phase

Outputs the eval report path. Action varies:
- After FAIL: "Evaluation found deficiencies. Spawn a sub-agent to update the plan and notes."
- After PASS below min_rounds: "Minimum evaluation rounds not met."

Advancing from REFINE runs the validation gate.

---

## Invariants

1. **Study phases precede REVIEW.** STUDY_SPECS → STUDY_CODE → STUDY_PACKAGES → REVIEW. No phase is skipped.
2. **Validation precedes evaluation.** The validation gate runs before every EVALUATE entry.
3. **Round monotonicity.** The planning round counter only increments.
4. **Min rounds enforced.** PASS below `planning.eval.min_rounds` forces another cycle.
5. **Max rounds enforced.** FAIL at `planning.eval.max_rounds` forces acceptance.
6. **Guided pauses.** When `config.general.user_guided` is true, REVIEW output includes "STOP please review and discuss with user before continuing."
7. **Commit gating.** `--message` is only required when `enable_commits` is `true`.
8. **Batch limitation.** `planning.batch` > 1 is not yet supported. TODO: reserved for future use.

---

## Edge Cases

- **Scenario:** Validation passes on first try after DRAFT.
  - **Expected:** Transitions directly to EVALUATE in one `advance` call.
  - **Rationale:** VALIDATE is a gate, not a user-facing state. When the plan is valid, the architect never sees VALIDATE.

- **Scenario:** Validation fails after REFINE.
  - **Expected:** Enters VALIDATE loop.
  - **Rationale:** Refinement may introduce structural errors; the validation gate catches them before re-evaluation.

- **Scenario:** plan.json does not exist when validation runs.
  - **Expected:** Validation fails with file-not-found.
  - **Rationale:** The plan must physically exist to be validated. Missing file is the first check.

- **Scenario:** plan.json has a dependency cycle.
  - **Expected:** Validation fails listing the cycle.
  - **Rationale:** Cycles make layer ordering impossible. The DAG check detects and reports them.

- **Scenario:** `--eval-report` points to non-existent file.
  - **Expected:** Error. State unchanged.
  - **Rationale:** Eval reports are audit artifacts. Accepting a reference to a non-existent file would leave a broken trail.

---

## Testing Criteria

### Study phases advance sequentially
- **Verifies:** Sequential state progression through study phases.
- **Given:** ORIENT (planning).
- **When:** advance through STUDY_SPECS → STUDY_CODE → STUDY_PACKAGES → REVIEW → DRAFT.
- **Then:** Each transitions in order.

### DRAFT with valid plan goes to EVALUATE
- **Verifies:** Validation gate passes transparently.
- **Given:** DRAFT, plan.json valid.
- **When:** `advance`
- **Then:** State is EVALUATE. Round is 1.

### DRAFT with invalid plan enters VALIDATE
- **Verifies:** Validation gate catches errors.
- **Given:** DRAFT, plan.json invalid.
- **When:** `advance`
- **Then:** State is VALIDATE.

### VALIDATE loops until valid
- **Verifies:** VALIDATE rejects until plan is fixed.
- **Given:** VALIDATE, plan.json still invalid.
- **When:** `advance`
- **Then:** Stays VALIDATE. Exit code 1.

### Planning EVALUATE PASS at min_rounds → ACCEPT
- **Verifies:** PASS with sufficient rounds triggers acceptance.
- **Given:** EVALUATE, min_rounds: 1, round 1.
- **When:** `advance --verdict PASS --eval-report evals/round-1.md`
- **Then:** State is ACCEPT.

### Planning EVALUATE FAIL at max_rounds → ACCEPT (forced)
- **Verifies:** FAIL at max rounds forces acceptance.
- **Given:** EVALUATE, max_rounds: 2, round 2.
- **When:** `advance --verdict FAIL --eval-report evals/round-2.md`
- **Then:** State is ACCEPT.

### Planning ACCEPT → PHASE_SHIFT
- **Verifies:** Acceptance triggers phase transition.
- **Given:** ACCEPT (planning).
- **When:** `advance`
- **Then:** State is PHASE_SHIFT.

### Planning eval command outputs context
- **Verifies:** Eval command assembles full evaluation context.
- **Given:** EVALUATE (planning), round 1.
- **When:** `forgectl eval`
- **Then:** Output includes plan-eval.md contents, plan references, report target.

---

## Implements
- Planning phase: structured study → draft → validate → evaluate → accept
- Plan validation gate with 12 structural checks
- Eval round enforcement (`planning.eval.min_rounds`/`max_rounds`) with forced acceptance
- Commit gating via `enable_commits` configuration
- Domain artifacts in `.forge_workspace/`
- Dual evaluator prompts: plan-eval.md for planning sub-agent
- `planning.batch` > 1: TODO, not yet supported
