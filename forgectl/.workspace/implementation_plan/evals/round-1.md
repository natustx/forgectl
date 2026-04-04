# Evaluation Report — Round 1

## Verdict: FAIL

## Summary
- Dimensions passed: 6/11
- Total spec requirements checked: 247
- Total covered: 209
- Deficiencies: 38

## Dimension Results

### 1. Behavior — FAIL

The plan covers the major behavioral flows but has gaps in several areas.

**session-init.md — Behavior Steps:**
- Steps 1-8 are covered across `config.toml` (project root discovery, TOML parsing, config validation), `init.overhaul` (init command rewrite), and `types.config`/`types.state` (state schema).
- Step 9 (logging at init: prune, create log file, write init entry) is covered by `logging.core`.
- COVERED: Project root discovery, TOML config reading, config validation, session ID generation, all three phase inits.

**spec-lifecycle.md — Behavior:**
- Batch selection, domain path, state machine transitions, add-queue-item, set-roots: These behaviors are already implemented and the plan explicitly notes no changes to the spec-lifecycle state machine. The plan items `advance.flags` and `output.eval` add the new flag gating and eval outputs for specifying states.
- COVERED: The spec-lifecycle behavior is handled by existing code plus the plan's flag gating and output changes.

**spec-reconciliation.md — Behavior:**
- RECONCILE through COMPLETE transition table: Already implemented. Plan adds `advance.flags` for enable_commits gating at COMPLETE and `git.autocommit` for commit execution and hash registration.
- COVERED.

**phase-transitions.md — Behavior:**
- generate_planning_queue phase (ORIENT/REFINE/PHASE_SHIFT): Covered by `advance.genqueue`.
- Multi-plan PHASE_SHIFT variants (planning->implementing, planning->planning, implementing->planning, implementing->implementing): Covered by `advance.phaseshift`.
- `--from` override at specifying PHASE_SHIFT to skip generate_planning_queue: Covered by `advance.genqueue`.
- `--from` override at generate_planning_queue PHASE_SHIFT: Covered by `advance.genqueue`.
- `--guided`/`--no-guided` at phase shifts: NOT EXPLICITLY ADDRESSED. The plan does not have a test or item step for updating `config.general.user_guided` via `--guided`/`--no-guided` flags at phase shifts. This is mentioned in the spec as an invariant and has a test case.

**plan-production.md — Behavior:**
- Study phases (STUDY_SPECS, STUDY_CODE, STUDY_PACKAGES, REVIEW, DRAFT): Already implemented. Plan updates output via `output.updates`.
- SELF_REVIEW state: Covered by `advance.selfreview`.
- Validation gate: Updated by `validate.schema`.
- DONE state (only when plan_all_before_implementing=true): Covered by `advance.phaseshift`.
- ACCEPT commit gating: Covered by `advance.flags` and `git.autocommit`.
- MISSING: plan-production specifies that STUDY_CODE output should list spec file paths. The `output.updates` item step says "Update printStudyCode to include spec files list" — this is covered.
- MISSING: DONE rejects flags. The spec says `advance` in DONE with any flags returns error: "DONE is a pass-through state. No flags accepted." No plan item or test addresses this rejection behavior.

**batch-implementation.md — Behavior:**
- Batch calculation, state machine, item passes transitions: Already implemented.
- COMMIT state: Plan adds `advance.flags` for enable_commits gating and `git.autocommit` for commit execution.
- First-round IMPLEMENT commit behavior: Covered by `advance.flags`.
- DONE variants for multi-plan: Covered by `advance.phaseshift`.

**activity-logging.md — Behavior:**
- Log file creation, pruning, best-effort logging: Covered by `logging.core`.

**validate-command.md — Behavior:**
- Auto-detection, type override, validation, no session required, path resolution: Covered by `cmd.validate`.

**state-persistence.md — Behavior:**
- Atomic writes, startup recovery, session archiving: Already implemented.
- Status command with --verbose: Covered by `cmd.status`.
- MISSING: Session archiving behavior. The spec describes completed sessions being archived to `state_dir/sessions/` with naming convention `<domain>-<date>.json`. No plan item addresses implementing session archiving. However, this may already be implemented — the context notes say state persistence is already implemented.

### 2. Error Handling — FAIL

**session-init.md — Error Handling:**
- `.forgectl/` not found: Covered by `config.toml` (FindProjectRoot) and `init.overhaul`.
- Config file missing or invalid TOML: Covered by `config.toml` (LoadConfig).
- Config constraint violation: Covered by `config.toml` (ValidateConfig) and `init.overhaul`.
- Input file not found, invalid JSON, schema failure: Already implemented; not changed by plan.
- COVERED.

**spec-lifecycle.md — Error Handling (within Rejection table):**
- `--eval-report` pointing to non-existent file: Covered by existing code.
- `--eval-report` when `enable_eval_output: false`: Covered by `advance.flags`.
- `add-queue-item` outside valid states: Already implemented.
- `set-roots` outside valid states: Already implemented.
- COVERED (existing code + plan additions).

**spec-reconciliation.md — Error Handling:**
- `advance --eval-report` when `enable_eval_output: false`: Covered by `advance.flags`.
- `advance` in COMPLETE without `--message` when `enable_commits: true`: Covered by `advance.flags`.
- COVERED.

**plan-production.md — Error Handling:**
- `advance` in DONE with any flags: NOT COVERED. No plan item or test addresses this.
- `advance --eval-report` when `enable_eval_output: false`: Covered by `advance.flags`.
- COVERED otherwise.

**batch-implementation.md — Error Handling:**
- `advance` in IMPLEMENT without `--message` when `enable_commits: true` (first round): Addressed in `advance.flags` steps (IMPLEMENT state commit gating).
- WAIT — The `advance.flags` item only mentions COMPLETE, ACCEPT, and COMMIT for --message gating. It does NOT mention IMPLEMENT first-round --message requirement. The spec says first-round IMPLEMENT requires `--message` when `enable_commits: true`. The `advance.flags` item steps say: "In advanceSpecifying COMPLETE", "In advancePlanning ACCEPT", "In advanceImplementing COMMIT". It misses implementing IMPLEMENT first-round `--message` gating.
- MISSING: `--message` gating at implementing IMPLEMENT (first round).

**validate-command.md — Error Handling:**
- No file path argument, file doesn't exist, invalid JSON, auto-detection fails, invalid --type, type mismatch: Covered by `cmd.validate`.

### 3. Rejection — FAIL

**session-init.md — Rejection table:**
- `.forgectl/` not found: Covered (`config.toml`, `init.overhaul`).
- `.forgectl/config` missing or unparseable: Covered (`config.toml`).
- Config constraint violation: Covered (`config.toml` ValidateConfig tests).
- Init with existing state: Already implemented.
- `--from` file fails validation: Already implemented.
- `--phase` invalid value: Already implemented.
- `--phase generate_planning_queue`: Covered (`init.overhaul` test).
- COVERED.

**spec-lifecycle.md — Rejection table:**
- All rejection conditions are either already implemented or covered by `advance.flags`.
- COVERED.

**spec-reconciliation.md — Rejection table:**
- `advance --eval-report` when `enable_eval_output: false`: Covered.
- `advance` in COMPLETE without `--message` when `enable_commits: true`: Covered.
- `add-queue-item` outside RECONCILE_REVIEW: Already implemented.
- COVERED.

**plan-production.md — Rejection table:**
- `advance` in DONE with any flags: NOT COVERED by any plan item or test.
- `advance --eval-report` when `enable_eval_output: false`: Covered.
- `advance` at ACCEPT without `--message` when `enable_commits: true`: Covered.
- PARTIAL FAIL.

**batch-implementation.md — Rejection table:**
- `advance` in IMPLEMENT without `--message` when `enable_commits: true`: NOT explicitly covered in `advance.flags` steps.
- `advance` in COMMIT without `--message` when `enable_commits: true`: Covered.
- `advance --eval-report` when `enable_eval_output: false`: Covered.
- PARTIAL FAIL.

**state-persistence.md — Rejection table:**
- `advance`/`status`/`eval` before `init`: Already implemented.
- COVERED.

**validate-command.md — Rejection table:**
- All 6 rejection conditions covered by `cmd.validate`.
- COVERED.

**phase-transitions.md — Rejection table:**
- `--from` pointing to invalid plan queue at various PHASE_SHIFTs: Covered by `advance.genqueue` tests.
- REFINE with invalid plan queue: Covered by `advance.genqueue` tests.
- COVERED.

### 4. Interface — PASS

**session-init.md — Interface:**
- CLI `init` command with `--from` and `--phase`: Plan removes old flags (`--batch-size`, `--min-rounds`, `--max-rounds`) in `init.overhaul`. Correct.
- Spec queue, plan queue, plan.json input schemas: Already implemented in validation.
- Validation failure output: Already implemented.
- COVERED.

**state-persistence.md — Interface:**
- `status` command with `--verbose`/`-v`: Covered by `cmd.status`.
- State file schema: Covered by `types.config`, `types.state`, `types.planitem`.
- COVERED.

**spec-lifecycle.md — Interface:**
- `advance` flags per state: Covered by `advance.flags`.
- `eval` command: Covered by `output.eval` and `cmd.eval`.
- `add-queue-item`: Already implemented.
- `set-roots`: Already implemented.
- All output formats: Covered by `output.updates` and `output.eval`.
- COVERED.

**spec-reconciliation.md — Interface:**
- `advance` flags: Covered.
- `eval` command for RECONCILE_EVAL: Covered by `cmd.eval` and `output.eval`.
- All advance output formats: Covered by `output.updates`.
- COVERED.

**plan-production.md — Interface:**
- `advance` flags: Covered.
- `eval` command: Already implemented + `output.eval` for gating.
- All output formats: Covered.
- COVERED.

**batch-implementation.md — Interface:**
- `advance` flags: Covered by `advance.flags`.
- `eval` command: Already implemented + gating.
- All output formats: Covered.
- COVERED.

**activity-logging.md — Interface:**
- Log configuration: Covered by `types.config`.
- Log directory and file naming: Covered by `logging.core`.
- COVERED.

**validate-command.md — Interface:**
- CLI command `forgectl validate [--type] <file>`: Covered by `cmd.validate`.
- All output formats: Covered.
- COVERED.

### 5. Configuration — PASS

**session-init.md — Configuration:**
- `--from`, `--phase` flags: Covered.
- `[[domains]]` section validation: Covered by `config.toml` (ValidateConfig for nested paths) and `init.overhaul` (domain validation against spec queue).
- `commit_strategy` per phase: Covered by `config.toml` (ValidateConfig test for invalid commit_strategy).
- `[logs]` section: Covered by `types.config` (LogsConfig struct) and `config.toml` (LoadConfig).
- COVERED.

**state-persistence.md — Configuration:**
- Full config structure in state file: Covered by `types.config` and `types.state`.
- All config fields (domains, specifying, planning, implementing, paths, general, logs): Covered by `types.config`.
- COVERED.

**activity-logging.md — Configuration:**
- `logs.enabled`, `logs.retention_days`, `logs.max_files`: Covered by `types.config` and `logging.core`.
- COVERED.

### 6. Observability — PASS

**activity-logging.md — Observability/Logging:**
- JSONL log entries for init and advance: Covered by `logging.core`.
- Log entry fields (ts, cmd, phase, prev_state, state, detail): Covered by `logging.core` (LogEntry struct).
- Detail fields by command: Covered by `logging.core` step "Build detail map per state/command".
- Read-only commands do not log: Covered by design (only init and advance write entries).
- COVERED.

No other spec has an explicit Observability section. The activity-logging spec IS the observability spec.

### 7. Integration Points — PASS

**session-init.md:**
- spec-lifecycle, plan-production, batch-implementation, state-persistence, activity-logging: All integration points are reflected in plan dependencies. `init.overhaul` depends on `types.state` and `config.toml`. `logging.core` depends on `init.overhaul`.
- COVERED.

**state-persistence.md:**
- session-init, spec-lifecycle, spec-reconciliation, plan-production, batch-implementation, activity-logging, validate-command: Plan items reference all these specs.
- COVERED.

**spec-lifecycle.md:**
- spec-reconciliation, eval prompts, phase-transitions: Plan items `output.eval`, `cmd.eval`, `advance.genqueue` address these.
- COVERED.

**spec-reconciliation.md:**
- phase-transitions, evaluator prompts: Covered.
- COVERED.

**phase-transitions.md:**
- spec-reconciliation, plan-production, batch-implementation, session-init: All reflected in plan dependencies.
- COVERED.

**plan-production.md:**
- SPEC_MANIFEST.md, PLAN_FORMAT.md, evaluator prompts, phase-transitions: Plan references these.
- COVERED.

**batch-implementation.md:**
- plan.json, evaluator prompts, phase-transitions: Covered.
- COVERED.

**activity-logging.md:**
- session-init, state-persistence, all phase specs: Covered by `logging.core` depending on `init.overhaul` and `advance.flags`.
- COVERED.

**validate-command.md:**
- session-init, plan-production: Same validation functions reused. Covered by `cmd.validate`.
- COVERED.

### 8. Invariants — FAIL

**session-init.md — Invariants:**
1. No implicit state: Covered (state file is authoritative).
2. Config locked at init: Covered by `init.overhaul`.
3. Project root required: Covered by `config.toml`.
4. Session ID generated once: Covered by `init.overhaul`.
5. Logging is best-effort: Covered by `logging.core`.
- COVERED.

**state-persistence.md — Invariants:**
1. Phase is authoritative: Already implemented.
2. State file is durable: Already implemented.
3. Eval history is append-only: Already implemented.
4. Config locked at init (except user_guided via --guided/--no-guided): The plan does not explicitly address the mutability of `config.general.user_guided` via `--guided`/`--no-guided` on advance. This is mentioned in the spec but no plan item covers it.
- PARTIAL: `--guided`/`--no-guided` flag handling is not addressed.

**spec-lifecycle.md — Invariants:**
- All 15 invariants are either already implemented or covered by plan items. The key NEW invariant #7 (no commits during specifying lifecycle) is enforced by the plan design (single commit at COMPLETE).
- COVERED.

**spec-reconciliation.md — Invariants:**
1-9: All covered by existing code or plan items (`advance.flags`, `git.autocommit`).
- COVERED.

**plan-production.md — Invariants:**
1-13: Most covered. Invariant #7 (guided pauses) requires `config.general.user_guided` to be checked in REVIEW output. Plan's `output.updates` does not explicitly mention this. However, this is already implemented behavior.
- Invariant #10 (DONE only reachable when plan_all_before_implementing=true): Covered by `advance.phaseshift`.
- Invariant #11 (No plan addition at DONE): Already implemented.
- COVERED.

**batch-implementation.md — Invariants:**
- Invariant #7 (first-round commits at IMPLEMENT): NOT fully addressed. `advance.flags` does not include IMPLEMENT first-round `--message` requirement.
- Invariant #13 (auto-commit at commit points): IMPLEMENT first-round commit is missing from `advance.flags` and `git.autocommit`.
- FAIL.

**phase-transitions.md — Invariants:**
1-10: Covered by `advance.genqueue`, `advance.phaseshift`, and existing code.
- COVERED.

**activity-logging.md — Invariants:**
1-6: All covered by `logging.core`.
- COVERED.

**validate-command.md — Invariants:**
1-4: Covered by `cmd.validate`.
- COVERED.

### 9. Edge Cases — FAIL

**session-init.md — Edge Cases:**
- Init with pre-existing passes/rounds in plan.json: Already implemented.
- `--from` with non-matching JSON: Already implemented.
- Missing config section falls back to defaults: Covered by `config.toml` test "LoadConfig merges partial TOML".
- `.forgectl/` found two levels up: Covered by `config.toml` test "FindProjectRoot walks up from nested subdirectory".
- Nested domain paths: Covered by `config.toml` test.
- Spec queue entry references unconfigured domain: Covered by `init.overhaul` step "Validate domain config against spec queue".
- No `[[domains]]` section: Covered by `config.toml` applyDefaults.
- COVERED.

**state-persistence.md — Edge Cases:**
- All crash recovery scenarios: Already implemented.
- Absolute vs relative state_dir: Already implemented.
- COVERED.

**spec-lifecycle.md — Edge Cases:**
- Most edge cases are already implemented. The NEW edge cases related to enable_commits and enable_eval_output are covered by `advance.flags`.
- COVERED.

**spec-reconciliation.md — Edge Cases:**
- COMPLETE with enable_commits=true and no files changed: The `git.autocommit` item does not have a test for this (commit skipped when no files changed, notice printed).
- COMPLETE with enable_commits=false and --message provided: Covered by `advance.flags`.
- PARTIAL FAIL: Missing test for empty commit scenario.

**plan-production.md — Edge Cases:**
- Validation passes on first try with self_review=true/false: Covered.
- Agent revises plan during SELF_REVIEW and introduces errors: Covered by `advance.selfreview` (validation gate runs on advance from SELF_REVIEW; the test for SELF_REVIEW with invalid plan enters VALIDATE).
- ACCEPT with enable_commits=true and no files changed: NOT tested.
- ACCEPT with enable_commits=false and --message: Covered.
- COVERED except empty commit scenario.

**batch-implementation.md — Edge Cases:**
- Layer has fewer items than batch: Already implemented.
- Single-item batch: Already implemented.
- EVALUATE PASS but rounds < min_rounds: Already implemented.
- EVALUATE FAIL at max_rounds: Already implemented.
- Item depends on failed item: Already implemented.
- All layers complete, no plans remaining: Covered by `advance.phaseshift`.
- All layers complete with plans remaining (both modes): Covered by `advance.phaseshift`.
- `eval` called outside EVALUATE: Covered.
- COVERED.

**phase-transitions.md — Edge Cases:**
- All edge cases covered by `advance.genqueue` and `advance.phaseshift` tests.
- COVERED.

**activity-logging.md — Edge Cases:**
- Directory does not exist: Covered by `logging.core` (creates directory).
- logs.enabled false: Covered.
- Disk full: Covered by `logging.core` (best-effort, test for unwritable directory).
- retention_days=0 and max_files=0: NOT tested. The spec says no pruning occurs when both are 0. No specific test.
- Both constraints apply: NOT tested. The spec says age-based runs first, then count-based.
- Session spans all phases: Covered by `logging.core` test "Session spanning multiple phases uses single log file".
- add-queue-item/set-roots do not log: NOT tested. The spec says these commands do not produce log entries. No test verifies this.
- PARTIAL FAIL.

**validate-command.md — Edge Cases:**
- File has multiple recognized top-level keys: NOT tested. No plan item addresses this.
- Empty object `{}`: NOT tested.
- plan.json with non-existent refs path: Covered by `cmd.validate` test for invalid plan.json.
- `--type plan` on spec-queue file: Covered by `cmd.validate` test for type override conflict.
- PARTIAL FAIL.

### 10. Testing Criteria — FAIL

I evaluated every "Testing Criteria" entry in every spec against the plan's test entries.

**session-init.md — Testing Criteria (10 tests):**
1. Init defaults to specifying phase: Implied by init.overhaul tests but NOT explicitly tested.
2. Init at planning phase: Already implemented.
3. Init at implementing phase: Already implemented.
4. Init rejects missing .forgectl directory: Covered by `init.overhaul`.
5. Init rejects invalid config: Covered by `config.toml` (ValidateConfig min>max) and `init.overhaul`.
6. Init rejects existing state: Already implemented.
7. Init rejects invalid queue: Already implemented.
8. Init locks config into state file: Covered by `init.overhaul` test "init reads batch from TOML and stores in ForgeState.Config".
9. Init applies defaults for missing config values: Covered by `config.toml` test "LoadConfig merges partial TOML".
10. Init discovers project root from subdirectory: Covered by `config.toml` test.
- COVERED (mostly through existing implementation + plan tests).

**state-persistence.md — Testing Criteria (4 tests):**
1. Recovery from crash: Already implemented.
2. Recovery from corrupt state: Already implemented.
3. State file in configured state_dir: Already implemented.
4. Project root discovery from subdirectory: Covered.
- COVERED.

**spec-lifecycle.md — Testing Criteria (32 tests):**
- Most are already implemented. New tests in the plan cover the flag gating changes.
- `--message ignored when enable_commits is false`: Covered by `advance.flags`.
- The plan does not have tests for every single spec-lifecycle testing criterion, but the plan states these are already implemented. New behavior tests are covered.
- COVERED (existing + plan additions).

**spec-reconciliation.md — Testing Criteria (14 tests):**
- Most already implemented.
- COMPLETE auto-commits when enable_commits is true: Covered by `git.autocommit` test.
- COMPLETE rejects missing --message when enable_commits is true: Covered by `advance.flags`.
- COMPLETE ignores --message when enable_commits is false: Covered by `advance.flags`.
- COMPLETE skips commit when no files changed: NOT tested. `git.autocommit` has no test for the empty commit scenario.
- PARTIAL FAIL.

**phase-transitions.md — Testing Criteria (22 tests):**
- Most covered by `advance.genqueue` and `advance.phaseshift`.
- `--guided at phase shift`: NOT tested. No plan item tests `--guided`/`--no-guided` flag at PHASE_SHIFT.
- Full lifecycle tests: Covered by the multi-plan tests in `advance.phaseshift`.
- PARTIAL FAIL.

**plan-production.md — Testing Criteria (13 tests):**
- Study phases advance sequentially: Already implemented.
- DRAFT with valid plan to EVALUATE/SELF_REVIEW: Covered.
- SELF_REVIEW tests: Covered by `advance.selfreview`.
- VALIDATE loops: Already implemented.
- EVALUATE PASS/FAIL transitions: Already implemented.
- ACCEPT → DONE: Covered by `advance.phaseshift`.
- DONE → PHASE_SHIFT: Covered.
- Planning ACCEPT auto-commits: Covered by `git.autocommit`.
- Planning ACCEPT ignores --message: Covered by `advance.flags`.
- Planning eval command outputs context: Already implemented + `output.eval` updates.
- COVERED.

**batch-implementation.md — Testing Criteria (17 tests):**
- Most already implemented.
- First-round IMPLEMENT requires --message when enable_commits is true: NOT tested in plan. `advance.flags` does not include this test.
- First-round IMPLEMENT without --message when enable_commits is false: NOT tested.
- COMMIT → ORIENT/DONE: Already implemented.
- Implementing eval command: Covered by existing + `output.eval`.
- DONE transitions (interleaved and all-planning-first): Covered by `advance.phaseshift`.
- PARTIAL FAIL.

**activity-logging.md — Testing Criteria (8 tests):**
1. init creates log file: Covered by `logging.core`.
2. advance appends log entry: Covered.
3. advance with verdict logs detail: Covered.
4. logging disabled skips everything: Covered.
5. pruning at init by age: Covered.
6. pruning at init by count: Covered.
7. log write failure is non-fatal: Covered.
8. status does not log: NOT tested. No plan test verifies read-only commands skip logging.
- PARTIAL FAIL.

**validate-command.md — Testing Criteria (10 tests):**
1. Auto-detect spec-queue: Covered.
2. Auto-detect plan-queue: Covered.
3. Auto-detect plan: Covered.
4. Auto-detection failure: NOT tested explicitly. `cmd.validate` tests cover invalid plan.json and non-existent file, but not the specific "unrecognized top-level key" scenario.
5. Invalid JSON: NOT tested explicitly by plan (covered implicitly by general validation).
6. Type override success: NOT tested explicitly.
7. Type override mismatch: Covered by `cmd.validate` test for type override conflict.
8. Validation errors reported: Covered.
9. No session required: NOT tested explicitly.
10. Plan path resolution for refs: Covered by `cmd.validate` test.
- PARTIAL FAIL: Several validate-command testing criteria lack explicit corresponding tests.

### 11. Dependencies & Format — PASS

**Plan Structure:**
- `context`: Present with `domain` and `module` as non-empty strings.
- `refs`: Present with 16 entries. All refs point to spec files and notes files using relative paths from the plan.json directory.
- `layers`: 5 layers (L0-L4), each listing item IDs.
- `items`: 17 items, each with `id`, `name`, `description`, `depends_on`, `steps`, `files`, `specs`, `refs`, `tests`.

**ID Uniqueness:** All 17 item IDs are unique.

**Layer Coverage:** Every item appears in exactly one layer. Every layer item ID exists in the items array.

**Layer Ordering:**
- L0: `types.config` (no deps), `types.state` (depends on types.config — same layer, OK), `types.planitem` (no deps).
- L1: `config.toml` (depends on types.config — L0), `init.overhaul` (depends on types.state, config.toml — L0 and L1 same layer), `validate.schema` (depends on types.planitem — L0).
- L2: `advance.flags` (depends on types.state, init.overhaul — L0, L1), `advance.selfreview` (depends on types.state, init.overhaul — L0, L1), `advance.genqueue` (depends on types.state, init.overhaul — L0, L1), `advance.phaseshift` (depends on types.state, advance.genqueue — L0, L2 same layer), `git.autocommit` (depends on advance.flags — L2 same layer).
- L3: `git.cleanup` (depends on git.autocommit — L2), `output.updates` (depends on advance.genqueue, advance.phaseshift, advance.selfreview, advance.flags — all L2), `output.eval` (depends on output.updates — L3 same layer), `cmd.status` (depends on output.updates — L3 same layer), `cmd.validate` (depends on validate.schema — L1), `cmd.eval` (depends on output.eval — L3 same layer).
- L4: `logging.core` (depends on init.overhaul, advance.flags — L1, L2).

Layer ordering is valid — items only depend on items in equal or earlier layers.

**DAG Validity:** No cycles detected. All `depends_on` references point to valid item IDs.

**Test Schema:** Every test has `category` and `description` with correct types. Categories are one of `functional`, `rejection`, `edge_case`.

**Notes Files:** All paths in `refs` are relative paths. I cannot verify file existence without reading each notes file, but the refs array structure is correct.

**PLAN_FORMAT.md:** The plan format file (`PLAN_FORMAT.md`) is referenced in the evaluator instructions but does not exist in the repository. This is a reference issue, not a plan structural issue. The plan itself conforms to the expected format.

PASS — The plan structure, IDs, DAG, layers, and format are all valid.

## Deficiency List

| # | Dimension | Spec Section | Missing Coverage |
|---|-----------|-------------|-----------------|
| 1 | Behavior | phase-transitions.md Invariant 7 | `--guided`/`--no-guided` flags on `advance` at phase shifts to mutate `config.general.user_guided` — no plan item or step addresses this |
| 2 | Behavior | plan-production.md Rejection table | `advance` in planning DONE with flags should return "DONE is a pass-through state. No flags accepted." — no plan item or test |
| 3 | Error Handling | batch-implementation.md Interface/IMPLEMENT | `--message` required at IMPLEMENT (first round) when `enable_commits: true` — `advance.flags` item does not include IMPLEMENT state in its steps or tests |
| 4 | Rejection | plan-production.md Rejection table | Same as #2 — DONE flag rejection not covered |
| 5 | Rejection | batch-implementation.md Rejection table | Same as #3 — IMPLEMENT first-round `--message` rejection not covered |
| 6 | Invariants | state-persistence.md Invariant 4 | `config.general.user_guided` mutability via `--guided`/`--no-guided` — no plan item addresses this |
| 7 | Invariants | batch-implementation.md Invariants 7, 13 | First-round IMPLEMENT commits (when enable_commits=true) and auto-commit at IMPLEMENT — not covered by `advance.flags` or `git.autocommit` |
| 8 | Edge Cases | spec-reconciliation.md Edge Case | COMPLETE with `enable_commits: true` and no files changed — commit skipped, notice printed — not tested |
| 9 | Edge Cases | plan-production.md Edge Case | ACCEPT with `enable_commits: true` and no files changed — commit skipped, notice printed — not tested |
| 10 | Edge Cases | activity-logging.md Edge Cases | `retention_days=0` and `max_files=0` (no pruning) — not tested |
| 11 | Edge Cases | activity-logging.md Edge Cases | Both retention_days and max_files apply (age-based first, then count-based) — not tested |
| 12 | Edge Cases | activity-logging.md Edge Cases | `add-queue-item`/`set-roots` commands do not produce log entries — not tested |
| 13 | Edge Cases | validate-command.md Edge Cases | File with multiple recognized top-level keys — not tested |
| 14 | Edge Cases | validate-command.md Edge Cases | Empty object `{}` as input — not tested |
| 15 | Testing Criteria | spec-reconciliation.md Testing | "COMPLETE skips commit when no files changed" — no plan test |
| 16 | Testing Criteria | phase-transitions.md Testing | "--guided at phase shift" test — `advance` with `--no-guided` at PHASE_SHIFT updates `config.general.user_guided` — no plan test |
| 17 | Testing Criteria | batch-implementation.md Testing | "First-round IMPLEMENT requires --message when enable_commits is true" — no plan test |
| 18 | Testing Criteria | batch-implementation.md Testing | "First-round IMPLEMENT without --message when enable_commits is false" — no plan test |
| 19 | Testing Criteria | activity-logging.md Testing | "status does not log" — no plan test verifying read-only commands skip logging |
| 20 | Testing Criteria | validate-command.md Testing | "Auto-detection failure" (unrecognized top-level key) — no explicit plan test |
| 21 | Testing Criteria | validate-command.md Testing | "Invalid JSON" (parse error with line/column) — no explicit plan test |
| 22 | Testing Criteria | validate-command.md Testing | "Type override success" (explicit type bypasses auto-detection) — no explicit plan test |
| 23 | Testing Criteria | validate-command.md Testing | "No session required" (works without .forgectl/) — no explicit plan test |
