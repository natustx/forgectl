# Commands Notes

## validate Command

### Synopsis

```
forgectl validate <file>
```

Standalone schema validation. Does not require a state file.

### Auto-Detection Logic

Inspect top-level keys of the JSON file:
- `{specs: [...]}` → spec-queue
- `{plans: [...]}` → plan-queue
- `{context: ..., items: [...], layers: [...]}` → plan

### Flags

```
--type <type>   Override auto-detection. Values: spec-queue, plan-queue, plan
```

### Path Resolution

For plan.json:
- `refs[].path`: relative to the directory containing plan.json (filepath.Dir)
- `items[].refs`: relative to the directory containing plan.json
- `items[].files`: relative to the project root (resolved via root discovery)
- `items[].specs`: display references only, NOT validated on disk

### Output on Success

```
✓ <filename>: valid <type>
```
Exit code 0.

### Output on Failure

```
FAIL: N errors in <filename>
  <error 1>
  <error 2>
  ...
```
Exit code 1.

### Implementation

- New file: `cmd/validate.go`
- Register with root command in `cmd/root.go`
- Call `state.ValidateSpecQueue`, `state.ValidatePlanQueue`, or `state.ValidatePlanJSON` from `state/validate.go`
- ValidatePlanJSON already takes `baseDir` — pass `filepath.Dir(file)` as the plan dir
- For project root path resolution (files/specs), discover root and pass to validate function (or adjust path before checking)

## status --verbose Flag

### Addition to cmd/status.go

```go
var verbose bool
statusCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show full session overview")
```

### Verbose Output Content (from state-persistence.md)

Full session overview:
1. Session header (same as non-verbose)
2. Current state and action (same as non-verbose)
3. Queue contents (for specifying: remaining specs; for planning: remaining plans; for implementing: remaining items by layer)
4. Completed items with eval history (for specifying: completed specs with rounds, eval records, commit hashes; for planning: completed plans)
5. Prior phase summaries (if session passed through earlier phases)
6. Per-item detail in implementing: each item's passes/rounds

Verbose output is appended after the standard status output.

### Mid-specifying verbose example

```
--- Queue ---
  [4] cache-invalidation.md (optimizer)
  [5] portal-rendering.md (portal)

--- Completed ---
  [1] repository-loading.md (optimizer)
      Rounds: 2, Evals: [r1: FAIL, r2: PASS]
      Commits: abc1234
  [2] snapshot-diffing.md (optimizer)
      Rounds: 1, Evals: [r1: PASS]
```

## eval Command Extension (RECONCILE_EVAL + CROSS_REFERENCE_EVAL)

### Current State

eval command is valid in: planning EVALUATE, implementing EVALUATE.

### New Valid States

Add: specifying RECONCILE_EVAL, specifying CROSS_REFERENCE_EVAL.

### RECONCILE_EVAL Output

See output.md for full format. Loads `evaluators/reconcile-eval.md`, embeds contents.

### CROSS_REFERENCE_EVAL Output

Loads `evaluators/cross-reference-eval.md`, embeds contents.

```
=== CROSS-REFERENCE EVALUATION ROUND N/N ===

--- EVALUATOR INSTRUCTIONS ---

<contents of evaluators/cross-reference-eval.md>

--- DOMAIN ---

<domain>: N specs

--- SPECS ---

  [1] <spec1>
  [2] <spec2>

--- REPORT OUTPUT ---  (only when enable_eval_output: true)

Write your evaluation report to:
  <domain>/specs/.eval/cross-reference-rN.md
```

### Implementation

In `cmd/eval.go`:
- Remove hardcoded state check (or extend it to cover all eval-valid states).
- Dispatch to appropriate `PrintEvalOutput` variant based on phase + state.

In `state/output.go`:
- Add `PrintReconcileEvalOutput(w, s)` function.
- Add `PrintCrossRefEvalOutput(w, s)` function.
- Both load evaluator files from embedded path relative to binary.

### Embedded Evaluator Files

The spec says evaluator prompts are "embedded in binary". Use Go's `//go:embed` directive to embed files at build time:

```go
//go:embed evaluators/spec-eval.md
var specEvalPrompt string

//go:embed evaluators/plan-eval.md
var planEvalPrompt string

//go:embed evaluators/impl-eval.md
var implEvalPrompt string

//go:embed evaluators/reconcile-eval.md
var reconcileEvalPrompt string

//go:embed evaluators/cross-reference-eval.md
var crossRefEvalPrompt string
```

This requires moving embed declarations to the package that contains the evaluator files, or using a sub-package. The most natural place is a new `evaluators` package or embedding from `state` package using relative paths.

Alternatively, embed in `main.go` or `cmd/root.go` and pass to output functions.

## Activity Logging

### Session ID

Generated at init using `crypto/rand`:

```go
func GenerateSessionID() string {
    var uuid [16]byte
    rand.Read(uuid[:])
    uuid[6] = (uuid[6] & 0x0f) | 0x40 // version 4
    uuid[8] = (uuid[8] & 0x3f) | 0x80 // variant RFC 4122
    return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
        uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}
```

### Log File Naming

```
~/.forgectl/logs/<phase>-<session_id_prefix>.jsonl
```
Where `session_id_prefix` is the first 8 hex chars (before first `-`).

Log file is created at init. Single file per session (even across phase shifts).

### Log Entry Format

```json
{"ts":"2026-03-29T14:32:01Z","cmd":"init","phase":"specifying","state":"ORIENT","detail":{"batch":3,"rounds":"1-3","guided":true}}
{"ts":"2026-03-29T14:33:12Z","cmd":"advance","phase":"specifying","prev_state":"ORIENT","state":"SELECT","detail":{"domain":"optimizer","batch":1}}
{"ts":"2026-03-29T15:10:33Z","cmd":"advance","phase":"specifying","prev_state":"DRAFT","state":"EVALUATE","detail":{"round":1}}
{"ts":"2026-03-29T15:12:01Z","cmd":"advance","phase":"specifying","prev_state":"EVALUATE","state":"REFINE","detail":{"round":1,"verdict":"FAIL","eval_report":"optimizer/specs/.eval/batch-1-r1.md"}}
```

### Detail Fields by State

**init:**
```json
{"batch": N, "rounds": "min-max", "guided": bool}
```

**advance — specifying EVALUATE:**
```json
{"round": N, "verdict": "PASS|FAIL", "eval_report": "<path>"}
```
(eval_report omitted if enable_eval_output=false)

**advance — implementing ORIENT:**
```json
{"layer": "<id>", "unblocked": N, "remaining": N}
```

**advance — implementing COMMIT:**
```json
{"layer": "<id>", "batch": N, "items": ["id1", "id2"]}
```

### Pruning at Init

After creating the log file, prune old files in `~/.forgectl/logs/`:

1. Delete files older than `logs.retention_days` days.
2. If more than `logs.max_files` files remain, delete oldest until at most `max_files` remain.
3. Prune is best-effort: if it fails, proceed silently.

### Best-Effort Logging

Never block primary workflow. If:
- Log directory can't be created: proceed silently.
- Log file can't be opened: proceed silently.
- Write fails: proceed silently.

Log errors are non-fatal.

### Implementation Files

- `state/logging.go` (new): `LogEntry` struct, `OpenSessionLog`, `WriteLog`, `PruneOldLogs`
- `cmd/init.go`: call `OpenSessionLog` after state file created, write init entry
- `cmd/advance.go`: call `WriteLog` after state transition completes
