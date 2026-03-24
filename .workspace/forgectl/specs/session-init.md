# Session Initialization

## Topic of Concern
> The scaffold initializes a session from a validated input file at a specified phase.

## Context

The `init` command creates a new `forgectl-state.json` from a user-provided input file. The input schema varies by phase: a spec queue for specifying, a plans queue for planning, or a plan.json for implementing. The scaffold validates the input, rejects malformed data with actionable errors, and sets the starting state for the chosen phase.

Sessions can begin at any of the three phases, allowing users to skip earlier phases when inputs already exist.

## Depends On
- **state-persistence** — provides the write mechanism and file layout for the state file.

## Integration Points

| Spec | Relationship |
|------|-------------|
| spec-lifecycle | Consumes the spec queue populated during specifying init |
| plan-production | Consumes the plan queue populated during planning init |
| batch-implementation | Consumes the plan.json validated during implementing init |
| state-persistence | State file schema defines the structure created here |

---

## Interface

### Inputs

#### CLI Command

| Command | Flags | Description |
|---------|-------|-------------|
| `init` | `--from <path>` (required), `--batch-size N` (required), `--min-rounds N` (default 1), `--max-rounds N` (required), `--phase specifying\|planning\|implementing` (default specifying), `--guided` / `--no-guided` (default guided) | Initialize state file from validated input |

#### Spec Queue Input File (`--phase specifying`)

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
| `specs[].name` | string | yes | Display name for the spec |
| `specs[].domain` | string | yes | Domain grouping |
| `specs[].topic` | string | yes | One-sentence topic of concern |
| `specs[].file` | string | yes | Target file path relative to project root |
| `specs[].planning_sources` | string[] | yes | Planning document paths the spec is derived from; may be empty array |
| `specs[].depends_on` | string[] | yes | Names of specs this one depends on; may be empty array |

No additional fields are permitted.

#### Plan Queue Input File (`--phase planning`)

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
| `plans` | array | yes | Ordered list of plans to generate and implement |
| `plans[].name` | string | yes | Display name for the plan |
| `plans[].domain` | string | yes | Domain grouping |
| `plans[].topic` | string | yes | One-sentence topic of concern |
| `plans[].file` | string | yes | Target path for plan.json relative to project root |
| `plans[].specs` | string[] | yes | Spec file paths to study; may be empty array |
| `plans[].code_search_roots` | string[] | yes | Directory roots for codebase exploration; may be empty array |

No additional fields are permitted.

#### Plan.json Input File (`--phase implementing`)

A `plan.json` file conforming to the schema defined in `PLAN_FORMAT.md`. The scaffold validates the full plan structure during init and adds `passes` and `rounds` fields to each item.

### Outputs

#### Validation Failure Output

When the input file fails validation, the scaffold prints:
1. Each validation error (missing field, extra field, wrong type) with the path to the offending location.
2. The complete valid schema as a reference.

The scaffold exits with a non-zero code on validation failure.

### Rejection

| Condition | Signal | Rationale |
|-----------|--------|-----------|
| `init` called when `forgectl-state.json` already exists | Error: "State file already exists. Delete it to reinitialize." Exit code 1. | Prevents accidental loss of in-progress state |
| `--from` file fails schema validation | Error listing violations. Prints full valid schema. Exit code 1. | User needs to see what's wrong |
| `--batch-size` < 1 | Error: "--batch-size must be at least 1." Exit code 1. | Invalid configuration |
| `--min-rounds` < 1 | Error: "--min-rounds must be at least 1." Exit code 1. | At least one eval round required |
| `--min-rounds` exceeds `--max-rounds` | Error: "--min-rounds cannot exceed --max-rounds." Exit code 1. | Invalid configuration |
| `--phase` not one of the three values | Error: "--phase must be specifying, planning, or implementing." Exit code 1. | Invalid phase |

---

## Behavior

### Initializing a Session

#### Preconditions
- No `forgectl-state.json` exists.
- `--from`, `--batch-size`, `--max-rounds` are provided.
- `--min-rounds` <= `--max-rounds`.
- `--batch-size` >= 1, `--min-rounds` >= 1.
- `--phase` is one of `specifying`, `planning`, `implementing` (default: `specifying`).

#### Steps
1. Read and parse the file at `--from`.
2. Validate against the schema for the specified `--phase`.
3. If validation fails: print errors and schema, exit code 1.
4. If validation passes:
   - For `--phase specifying`: create state file with phase `specifying`, state ORIENT, spec queue populated.
   - For `--phase planning`: create state file with phase `planning`, state ORIENT, plan queue populated.
   - For `--phase implementing`: validate plan.json, add `passes: "pending"` and `rounds: 0` to items, create state file with phase `implementing`, state ORIENT.

#### Postconditions
- State file exists with `batch_size`, `min_rounds`, `max_rounds`, `user_guided` set.
- Phase and state reflect the starting point.
- For `--phase implementing`: plan.json items have `passes` and `rounds` fields.

#### Error Handling
- File not found: error with path. Exit code 1.
- Invalid JSON: error with parse details. Exit code 1.
- Schema failure: error listing violations, print valid schema. Exit code 1.

---

## Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `--from` | string | none (required) | Path to input file (schema varies by `--phase`) |
| `--batch-size` | integer | none (required) | Max items per batch in implementing phase |
| `--min-rounds` | integer | 1 | Minimum evaluation rounds per cycle. Used in all phases. |
| `--max-rounds` | integer | none (required) | Maximum evaluation rounds per cycle. Used in all phases. |
| `--phase` | string | specifying | Starting phase: `specifying`, `planning`, `implementing` |
| `--guided` | boolean | true | Enable user-guided mode. Can be changed on any `advance`. |
| `--no-guided` | boolean | — | Disable user-guided mode. |

---

## Invariants

1. **No implicit state.** All information for transitions is in the state file (and plan.json during implementing).

---

## Edge Cases

- **Scenario:** `--phase implementing` with plan.json that has items already containing `passes` or `rounds` fields.
  - **Expected:** Fields are overwritten with `passes: "pending"` and `rounds: 0`.
  - **Rationale:** Init always starts fresh. Pre-existing tracking fields from a previous session are reset.

- **Scenario:** `--from` points to a valid JSON file that does not match any queue schema.
  - **Expected:** Schema validation fails with specific field-level errors and the valid schema printed as reference.
  - **Rationale:** Users need to see both what's wrong and what's expected to fix the file.

---

## Testing Criteria

### Init defaults to specifying phase
- **Verifies:** Default phase selection.
- **When:** `forgectl init --from specs-queue.json --batch-size 2 --max-rounds 3`
- **Then:** `phase: "specifying"`, `state: "ORIENT"`, `started_at_phase: "specifying"`.

### Init at planning phase
- **Verifies:** Phase selection with `--phase planning`.
- **When:** `forgectl init --phase planning --from plans-queue.json --batch-size 2 --max-rounds 3`
- **Then:** `phase: "planning"`, `state: "ORIENT"`. Specifying section is null.

### Init at implementing phase
- **Verifies:** Phase selection with `--phase implementing` and plan.json mutation.
- **When:** `forgectl init --phase implementing --from plan.json --batch-size 2 --max-rounds 3`
- **Then:** `phase: "implementing"`, `state: "ORIENT"`. plan.json items have `passes` and `rounds`.

### Init rejects existing state
- **Verifies:** Rejection when state file already exists.
- **Given:** State file exists.
- **When:** `forgectl init --from specs-queue.json --batch-size 2 --max-rounds 3`
- **Then:** Exit code 1.

### Init rejects invalid queue
- **Verifies:** Schema validation catches missing required fields.
- **Given:** Spec queue missing `file` field.
- **When:** `forgectl init --from bad-queue.json --batch-size 2 --max-rounds 3`
- **Then:** Exit code 1.

### Init rejects min exceeding max
- **Verifies:** Configuration constraint enforcement.
- **Given:** `--min-rounds 5 --max-rounds 2`
- **When:** `forgectl init --from specs-queue.json --batch-size 2 --min-rounds 5 --max-rounds 2`
- **Then:** Exit code 1.

### Init rejects batch-size less than 1
- **Verifies:** Configuration constraint enforcement.
- **Given:** `--batch-size 0`
- **When:** `forgectl init --from specs-queue.json --batch-size 0 --max-rounds 3`
- **Then:** Exit code 1.

---

## Implements
- Phase-selectable init (`--phase specifying|planning|implementing`)
- Input validation for spec queue, plan queue, and plan.json schemas
- Configuration parameters: batch_size, min_rounds, max_rounds, guided
