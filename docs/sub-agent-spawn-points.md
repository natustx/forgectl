# Sub-Agent Spawn Points

The scaffold does not spawn sub-agents. It outputs instructions telling the architect what to spawn. The architect (or the skill driving the session) is responsible for spawning them.

## Consistent Output Wording

Every spawn instruction follows the pattern:

```
Please spawn {count} {model} sub-agent(s) to {purpose}.
```

## Field Descriptions

Each spawn point is configured with three fields:

- `model` — which Claude model to use (e.g., `"opus"`, `"haiku"`). The scaffold includes this value verbatim in its spawn instructions. Can be a model name or a descriptive phrase (e.g., `"opus explorer"`, `"spec-eval-expert"`).
- `type` — the role of the sub-agent at this spawn point: `"eval"` (evaluate output and render a verdict), `"explore"` (read and search code or specs), or `"refine"` (apply corrections based on eval findings).
- `count` — how many sub-agents to spawn in parallel at this point.

## Effect of `enable_eval_output`

`enable_eval_output` affects what the spawn instructions include for eval spawn points in the planning and implementing phases:

- When `enable_eval_output: true`: spawn instructions tell the sub-agent to write a report to a specific path, and `advance` requires `--eval-report <path>`.
- When `enable_eval_output: false`: spawn instructions omit the report path. The sub-agent communicates its verdict verbally to the architect; no file is written.

The specifying phase always requires `--eval-report` and is not affected by `enable_eval_output`.

## Spawn Points

| # | Phase | State | Purpose | model | type | count | Config Key |
|---|-------|-------|---------|-------|------|-------|------------|
| 1 | Specifying | EVALUATE | Evaluate spec batch | opus | eval | 1 | `specifying.eval` |
| 2 | Specifying | CROSS_REFERENCE | Cross-reference domain specs | haiku | explore | 3 | `specifying.cross_reference` |
| 3 | Specifying | CROSS_REFERENCE_EVAL | Evaluate cross-reference work | opus | eval | 1 | `specifying.cross_reference.eval` |
| 4 | Specifying | RECONCILE_EVAL | Evaluate cross-domain reconciliation | opus | eval | 1 | `specifying.reconciliation` |
| 5 | Planning | STUDY_CODE | Explore codebase | haiku | explore | 3 | `planning.study_code` |
| 6 | Planning | EVALUATE | Evaluate plan | opus | eval | 1 | `planning.eval` |
| 7 | Implementing | EVALUATE | Evaluate implementation batch | opus | eval | 1 | `implementing.eval` |

## Configuration

`model` is a string that can be a model name (e.g., `"opus"`, `"haiku"`) or a descriptive phrase (e.g., `"opus explorer"`, `"spec-eval-expert"`). The scaffold includes this value verbatim in its spawn instructions.

`type` identifies the role of the sub-agent at this spawn point: `"eval"`, `"explore"`, or `"refine"`.

Each spawn point has `type`, `model`, and `count` in `.forgectl/config`:

```toml
[specifying.eval]
type = "eval"
model = "opus"
count = 1

[specifying.cross_reference]
type = "explore"
model = "haiku"
count = 3

[specifying.cross_reference.eval]
type = "eval"
model = "opus"
count = 1

[specifying.reconciliation]
type = "eval"
model = "opus"
count = 1

[planning.study_code]
type = "explore"
model = "haiku"
count = 3

[planning.eval]
type = "eval"
model = "opus"
count = 1

[implementing.eval]
type = "eval"
model = "opus"
count = 1
```
