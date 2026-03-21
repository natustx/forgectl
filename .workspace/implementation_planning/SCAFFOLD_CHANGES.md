# Scaffold Changes

> Notes on how the planning scaffold (planctl) should change based on the launcher planning session.

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
└── notes/
    ├── <package>.md
    └── ...
```

## Format definition

See `.workspace/implementation_planning/PLAN_FORMAT.md` for the full JSON schema, item structure, test structure, notes file conventions, and examples.

## Reference files

| File | Purpose |
|------|---------|
| `.workspace/implementation_planning/PLAN_FORMAT.md` | JSON schema and conventions for plan.json |
| `launcher/.workspace/implementation_plan/plan.json` | First plan generated in this format |
| `launcher/.workspace/implementation_plan/notes/` | Reference notes files (config, daemon, launch, stop, cli) |

## What the scaffold needs to change

1. **DRAFT state output:** The `--file` flag and `File` field in the queue currently point to a single markdown file. Should point to a directory or a JSON file instead.

2. **EVALUATE state:** The evaluator needs to understand the JSON format. See "Evaluator sub-agent" section below for the full prompt structure.

3. **Study phase notes should feed into plan generation.** The scaffold records study notes (specs, code, packages) — these should inform the plan's `context` block and `refs` list.

4. **Queue schema may need a `format` field** to distinguish markdown plans from JSON plans, if both formats need to coexist.

5. **Notes generation is a sub-step of DRAFT.** The scaffold doesn't currently track whether notes files have been created. This could be a validation check before advancing from DRAFT to EVALUATE.

6. **Evaluation reports written to disk.** The evaluator should write its report to `<domain>/.workspace/implementation_plan/evals/round-N.md`. This provides an audit trail that travels with the plan and allows REFINE agents to read exact deficiencies.

---

## Evaluator sub-agent

### Prompt structure

The evaluator is a sub-agent that reads the plan and specs, writes its report to `evals/round-N.md`, and returns a verdict. It does NOT modify the plan — only the eval report file.

The evaluator must check **every section** of every spec — not just Testing Criteria. Round 1 of the launcher evaluation only checked Testing Criteria, Invariants, and Edge Cases. Round 2 added Behavior, Error Handling, Rejection, Interface, Configuration, Observability, and Integration Points — and found 5 deficiencies that round 1 missed.

### Files the evaluator reads

1. The plan: `<domain>/.workspace/implementation_plan/plan.json`
2. All specs listed in the plan's `refs` array
3. The format definition: `.workspace/implementation_planning/PLAN_FORMAT.md`

### Evaluation dimensions

The evaluator checks ALL of the following. For each, it reads the spec section line by line.

| # | Dimension | Spec section | What to check |
|---|-----------|-------------|---------------|
| 1 | Behavior | §Behavior (Preconditions, Steps, Postconditions) | Every step has a corresponding plan item or test |
| 2 | Error Handling | §Error Handling tables within Behavior sections | Every failure row has a corresponding test |
| 3 | Rejection | §Rejection table | Every row maps to a test |
| 4 | Interface | §Interface (Inputs, Outputs) | Plan items produce types/functions matching the spec's interface |
| 5 | Configuration | §Configuration table | Every parameter (type, default, description) is addressed |
| 6 | Observability | §Observability/Logging tables | Every logging requirement (level + what) is addressed |
| 7 | Integration Points | §Integration Points table | Plan dependencies reflect cross-spec relationships |
| 8 | Invariants | §Invariants | Each invariant has an enforcement mechanism in the plan |
| 9 | Edge Cases | §Edge Cases | Each edge case has a corresponding test |
| 10 | Testing Criteria | §Testing Criteria | Every entry has a corresponding test in the plan |
| 11 | Dependencies & Format | Plan structure | IDs match, DAG is valid, layers respected, format compliant |

### Report format

The evaluator writes its report to `<domain>/.workspace/implementation_plan/evals/round-N.md` with:

- Verdict: PASS or FAIL
- Per-dimension: PASS or FAIL with specifics
- For FAIL: list of every deficiency with the spec section and missing coverage
- Summary: counts of what passed vs total (e.g., "47/49 Testing Criteria covered")

### Verdict rules

- **PASS** only if ALL dimensions pass
- **FAIL** if ANY dimension has deficiencies

### Lessons from the launcher session

Round 1 used a narrow prompt (Testing Criteria, Invariants, Edge Cases only) and returned PASS. Round 2 used the full prompt above and returned FAIL with 5 deficiencies:

1. Missing error handling test (daemon file write failure)
2. Missing configuration tests (timeout/interval defaults)
3. Observability requirements completely unaddressed (all 3 specs)
4. Missing edge case test (PID reuse)
5. Two Testing Criteria entries not covered

The narrow evaluation gave false confidence. The full evaluation must be the default.
