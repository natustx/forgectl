# Scaffold Changes

> Notes on how the planning scaffold (planctl) should change based on the launcher planning session and subsequent design review.

## What changed

The DRAFT phase now produces a `plan.json` + `notes/` directory instead of an `IMPLEMENTATION_PLAN.md` markdown file.

### Old output
```
<domain>/.workspace/implementation_plan/
└── IMPLEMENTATION_PLAN.md
```

### New output
```
<domain>/.workspace/implementation_plan/
├── plan.json
├── notes/
│   ├── <package>.md
│   └── ...
└── evals/
    ├── round-1.md
    └── ...
```

## Format definition

See `.workspace/implementation_planning/planctl/PLAN_FORMAT.md` for the full JSON schema, item structure, test structure, notes file conventions, and examples.

## Reference files

| File | Purpose |
|------|---------|
| `.workspace/implementation_planning/planctl/PLAN_FORMAT.md` | JSON schema and conventions for plan.json |
| `.workspace/implementation_planning/planctl/EVALUATOR_PROMPT.md` | Full instructions for the evaluation sub-agent |
| `launcher/.workspace/implementation_plan/plan.json` | First plan generated in this format |
| `launcher/.workspace/implementation_plan/notes/` | Reference notes files (config, daemon, launch, stop, cli) |

## Summary of scaffold changes

### State machine

- **SELECT → REVIEW**: Renamed. Lightweight checkpoint showing plan format path.
- **New: Validation gate**: Automatic structural validation of plan.json fires after DRAFT and after REFINE. If valid, transitions directly to EVALUATE (VALIDATE state is never visible). If invalid, enters VALIDATE loop until fixed.
- **Full flow**: ORIENT → STUDY_SPECS → STUDY_CODE → STUDY_PACKAGES → REVIEW → DRAFT → [validate] → EVALUATE ⇄ REFINE → [validate] → EVALUATE ... → ACCEPT → DONE

### Flags simplified

| Removed | Reason |
|---------|--------|
| `--file` | Queue provides plan.json path. No override needed. |
| `--notes` | Study phase findings are ephemeral. Not stored in state. |
| `--message` | Scaffold doesn't do git operations. |
| `--deficiencies` | Deficiencies live in the eval report file, not CLI flags. |
| `--fixed` | The plan diff is the fix. Eval report documents what was wrong. |
| `--sub-agents` | Hardcoded to 3. |
| `--user-guided` | Every phase is already manual. |

| Added | Purpose |
|-------|---------|
| `--eval-report <path>` | EVALUATE only. Stores path to eval report file. Scaffold verifies file exists. |

Only EVALUATE takes flags (`--verdict`, `--eval-report`). All other phases advance with bare `planctl advance`.

### State file simplified

| Removed | Reason |
|---------|--------|
| `sub_agents` | Hardcoded |
| `user_guided` | Removed |
| `study` (on ActivePlan/CompletedPlan) | Study notes no longer stored in state |
| `commit_hash` (on CompletedPlan) | Scaffold doesn't do git |
| `evals[].deficiencies` | In eval report file |
| `evals[].fixed` | In plan diff |

| Added | Reason |
|-------|--------|
| `evals[].eval_report` | Path to evaluation report file |

### Action description changes

Each phase now outputs relevant reference file paths:
- **STUDY_CODE**: "Explore the codebase in relation to the specs under study"
- **REVIEW**: Shows `planctl/PLAN_FORMAT.md` path only. No study notes.
- **DRAFT**: Shows plan.json path, notes dir, `planctl/PLAN_FORMAT.md` path.
- **EVALUATE**: Shows plan path, `planctl/EVALUATOR_PROMPT.md` path, report target path.
- **REFINE**: Shows eval report path from previous round.
- **VALIDATE** (on failure): Shows validation errors with field descriptions from PLAN_FORMAT.md.

---

## Evaluator sub-agent

Moved to standalone file: `.workspace/implementation_planning/planctl/EVALUATOR_PROMPT.md`

### Lessons from the launcher session

Round 1 used a narrow prompt (Testing Criteria, Invariants, Edge Cases only) and returned PASS. Round 2 used the full prompt above and returned FAIL with 5 deficiencies:

1. Missing error handling test (daemon file write failure)
2. Missing configuration tests (timeout/interval defaults)
3. Observability requirements completely unaddressed (all 3 specs)
4. Missing edge case test (PID reuse)
5. Two Testing Criteria entries not covered

The narrow evaluation gave false confidence. The full evaluation must be the default.
