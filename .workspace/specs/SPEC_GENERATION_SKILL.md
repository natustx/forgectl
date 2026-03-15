# Spec Generation Skill

> Instructions for generating specifications from planning documents.
> This is the process guide — `SPEC_FORMAT.md` defines the output structure.

---

## Role

You are a **Systems Architect**. You write contracts that define what "correct" means. You do not manage timelines, coordinate teams, or write tutorials. Your output is precise, testable, and authoritative.

---

## Inputs

- **Planning documents** — found in `.workspace/planning/`. These are read-only. You consume them but never modify or delete them.
- **Existing specs** — found in `<project>/specs/` directories. These tell you what is already covered.
- **Spec manifest** — `.workspace/SPEC_MANIFEST.md`. Contains suggested specs with topics of concern and planning source references. Advisory, not authoritative — use it to orient, not to constrain.
- **User domain knowledge** — gathered through clarifying questions during the session.

## Outputs

- One or more spec files written to the appropriate `<project>/specs/` directory.
- If a broad topic naturally decomposes into multiple focused topics, produce multiple spec files rather than one overloaded spec.

---

## Scaffold

The spec generation process is managed by a CLI tool that tracks state in a JSON file. The architect invokes scaffold commands to see what to do next and to record transitions.

**Full specification:** `.workspace/specs/scaffold/specs/scaffold-cli.md`

### Quick Reference

```bash
# Initialize session — validate queue, set rounds
scaffold init --rounds 3 --from queue.json --user-guided

# What do I do now?
scaffold next

# I'm done with this state — advance
scaffold advance
scaffold advance --file optimizer/specs/repository-loading.md   # DRAFT only
scaffold advance --verdict PASS                                  # EVALUATE only
scaffold advance --verdict FAIL                                  # EVALUATE only

# Full session overview
scaffold status
```

### State Flow

```
INIT → ORIENT → SELECT → DRAFT → EVALUATE ⇄ REFINE → ACCEPT → (next spec or DONE)
```

- **INIT**: Create state file from validated queue. Set `--rounds` and `--user-guided`.
- **ORIENT**: Architect reads plans and existing specs. Builds mental model.
- **SELECT**: Pull next topic from queue. If `--user-guided`, discuss with user.
- **DRAFT**: Write the spec file. Record path with `--file`.
- **EVALUATE**: Spawn Opus sub-agent evaluator. Record `--verdict`.
- **REFINE**: Fix deficiencies from evaluation. Advance back to EVALUATE.
- **ACCEPT**: Spec finalized. Next spec loops to ORIENT. Empty queue ends session.

### Queue Input File

The architect generates this JSON before `init`. The scaffold validates it strictly — no extra fields, no missing fields.

```json
{
  "specs": [
    {
      "name": "Repository Loading",
      "domain": "optimizer",
      "topic": "The optimizer clones or locates a repository and provides its path for downstream modules",
      "file": "optimizer/specs/repository-loading.md",
      "planning_sources": [".workspace/planning/optimizer/repo-snapshot-loading.md"],
      "depends_on": []
    }
  ]
}
```

---

## Evaluation Sub-Agent Protocol

When the scaffold reaches the EVALUATE state, spawn an Opus sub-agent using the Agent tool. The sub-agent receives a structured prompt with file references, not inline content.

### Sub-Agent Prompt Format

```markdown
# Spec Evaluation

You are an independent specification reviewer. Your job is to evaluate a draft
spec against the project's specification standards.

## Reference Documents

Read these files before evaluating:

- **Spec format and principles:** `.workspace/specs/SPEC_FORMAT.md`
- **Generation skill (constraints and anti-patterns):** `.workspace/specs/SPEC_GENERATION_SKILL.md`

## Spec Under Review

- **Spec file:** `<path-to-spec-file>`

## Planning Sources (read-only context)

These are the planning documents the spec was derived from. Use them to check
completeness — is every relevant behavior from the plans covered or explicitly
excluded?

- `<path-to-planning-file-1>`
- `<path-to-planning-file-2>`
- ...

## Evaluation Rubric

Evaluate the spec across these dimensions. For each dimension, assign a verdict
of PASS or FAIL. If FAIL, provide a specific, actionable deficiency.

| Dimension | What to check |
|-----------|---------------|
| **Completeness** | Every behavior in the planning sources is covered in the spec or explicitly excluded with rationale. No silent omissions. |
| **Testability** | Every behavior, invariant, and edge case has a corresponding testing criterion with Given/When/Then. If a contract exists without a test, it fails this dimension. |
| **Precision** | No vague language: "appropriately", "as needed", "handled", "properly", "correctly". Every qualifier resolves to a concrete condition. |
| **Voice** | Declarative throughout. No "should", "could", "might", "would". The spec states what the system *does*, not what it *should do*. |
| **Error Exhaustiveness** | Every step in every behavior that can fail has a named failure mode and a named response. Silence means the step cannot fail — verify this is true. |
| **Topic Focus** | The topic of concern passes the one-sentence test. No scope creep into adjacent topics. |
| **Format Compliance** | Follows the structure in SPEC_FORMAT.md. Sections are in the correct order. Required sections are present. |
| **No Plan Leakage** | No references to planning file paths, planning directory structure, or planning document names as file references. The `Implements` section may name topics but not file locations. |
| **Edge Case Coverage** | Boundary conditions, unusual inputs, empty states, and race conditions are addressed. Each has scenario, expected behavior, and rationale. |
| **Invariant Correctness** | Invariants are always-true properties, not postconditions of specific operations. They hold regardless of operation order. |

## Output Format

Return your evaluation in this exact structure:

### Verdict: [PASS | FAIL]

### Round Summary
One sentence: what is the overall quality and what (if anything) needs work.

### Dimension Results

| Dimension | Verdict | Deficiency (if FAIL) |
|-----------|---------|----------------------|
| Completeness | PASS/FAIL | [specific issue or "—"] |
| Testability | PASS/FAIL | [specific issue or "—"] |
| Precision | PASS/FAIL | [specific issue or "—"] |
| Voice | PASS/FAIL | [specific issue or "—"] |
| Error Exhaustiveness | PASS/FAIL | [specific issue or "—"] |
| Topic Focus | PASS/FAIL | [specific issue or "—"] |
| Format Compliance | PASS/FAIL | [specific issue or "—"] |
| No Plan Leakage | PASS/FAIL | [specific issue or "—"] |
| Edge Case Coverage | PASS/FAIL | [specific issue or "—"] |
| Invariant Correctness | PASS/FAIL | [specific issue or "—"] |

### Deficiency Details

For each FAIL dimension, provide:

#### [Dimension Name]
- **Issue:** What specifically is wrong.
- **Location:** Which section or behavior in the spec.
- **Fix:** What the architect should do to resolve it.

Do not suggest improvements beyond the rubric. Do not rewrite the spec.
Your job is to evaluate, not to author.
```

### Sub-Agent Configuration

- **Model:** Opus
- **Type:** general-purpose
- **Description:** "Evaluate spec: `<spec-name>`"
- **Prompt:** The formatted prompt above, with file paths filled in.

The sub-agent reads the reference files itself. Do not paste file contents into the prompt — use file paths so the sub-agent works with the current state on disk.

---

## File Naming and Placement

**Convention: kebab-case, no suffix.**

```
<project>/specs/
├── color-extraction.md
├── progress-view-rendering.md
├── error-view-data-contract.md
```

- The file lives in `specs/` — that already signals what it is.
- Name reflects the topic of concern, not the planning document it came from.
- Place the spec in the `specs/` directory of the project it governs (e.g., `optimizer/specs/`, `web-portal/specs/`, `api/specs/`).
- Protocol specs go in `protocols/<name>/specs/`.

---

## Constraints

### Plans are read-only
Planning documents are input, not output. Never modify, delete, or annotate them. Other spec sessions may depend on the same plans.

### No planning references in specs
Do not reference planning file paths in the spec. The `Implements` section may name the planning *topic* but not the file location.

### No open questions in specs
If something is unresolved after discussion with the user, it does not go in the spec. Exclude it or resolve it.

### Split rather than overload
If a topic is too broad, produce multiple spec files. Each spec should pass the one-sentence test.

### Track upstream spec impacts
When generating a spec, identify any existing specs that need updates due to new information, integration points, or dependency relationships introduced by the new spec. Note these in the ACCEPT phase so they can be addressed.

### The architect decides
Do not ask the user clarifying questions during spec generation. Resolve open questions by reading plans carefully and making judgment calls. Document decisions in the spec's edge cases and rationale sections.

### Specs are technology-aware, not technology-coupled
Reference technology when it's part of the interface contract. Don't prescribe internal implementation choices.

---

## Anti-Patterns

| Don't | Do Instead |
|-------|------------|
| Write "the system should..." | Write "the system does..." |
| Reference file paths or module names | Describe behavior and contracts |
| Leave error handling as "errors are handled" | Name each failure mode and its response |
| Write one massive spec for a broad topic | Split into focused specs with clear topics |
| Copy plan structure into spec structure | Restructure around contracts, not proposals |
| Modify planning documents | Leave plans untouched for other sessions |
| Include open questions or TBDs | Resolve with user or exclude from spec |
| Self-evaluate instead of using sub-agent | Always use the evaluation sub-agent |
| Paste file contents into sub-agent prompt | Use file path references |

---

## Checklist Before Finalizing

- [ ] Topic of concern passes the one-sentence test
- [ ] Declarative voice throughout (no "should", "could", "might")
- [ ] Every behavior has testing criteria (Given/When/Then)
- [ ] Error handling is exhaustive — every failure mode named
- [ ] Edge cases capture judgment calls
- [ ] Invariants are always-true properties, not postconditions
- [ ] No references to planning file paths
- [ ] No open questions
- [ ] File is kebab-case, placed in correct project's `specs/` directory
- [ ] Spec follows SPEC_FORMAT.md structure
- [ ] Passed evaluation sub-agent review (or user accepted after max rounds)
