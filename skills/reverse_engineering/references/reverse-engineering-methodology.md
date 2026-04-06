# Reverse Engineering Methodology

This document defines the investigation protocol for reverse-engineering code into specifications. It is loaded during the DRAFT state of the forgectl specifying phase.

---

## Core Principle

You document the implementation as it exists. You are a forensic recorder.

- If the code has a bug, describe the buggy behavior as the defined behavior.
- If the code handles an edge case illogically, capture that handling as stated behavior.
- If the code silently swallows errors, state that errors are silently swallowed.
- If you *think* something should happen but the code doesn't do it — leave it out entirely.

When source comments explain *why* a behavior must be preserved (regulatory, compatibility, known-intentional), capture that rationale alongside the behavior. Strip implementation references. If a comment contradicts the code, document the code's behavior and ignore the comment.

---

## Two-Phase Process

### Phase 1 — Investigation (Implementation-Aware)

You are deep in the code. Read function signatures, trace conditional branches, follow call chains, inspect variable types, map data flow. You see everything. Take notes. Trace paths. Build your understanding.

### Phase 2 — Output (Implementation-Free)

Strip away every implementation artifact. The output describes observable behavior only — inputs, outputs, transformations, side effects, and rules. No function names, class names, file paths, framework references. A reader must be able to reimplement the behavior on any stack without seeing the source code.

**Example of the phase separation:**

During investigation you read:
```
func CalculateDiscount(price float64, tier string) float64 {
    if tier == "gold" { return price * 0.8 }
    return price * 0.95
}
```

The specification output:
> When a discount is calculated, gold-tier customers receive a 20% reduction from the original price. All other customers receive a 5% reduction. No validation is performed on the price value or tier designation.

---

## Exhaustive Exploration Protocol

For each topic, work through every item below during Phase 1 before writing the spec.

### 1. Entry Points
- Identify every entry point into this topic (public functions, API endpoints, event handlers, message consumers, scheduled triggers).
- Note the exact signature, parameters, and defaults for each.

### 2. Code Path Tracing
- For every entry point, trace **every** conditional branch (if, else, switch, match, ternary, guard clauses).
- Follow each branch to its terminal point (return, throw, side effect, void completion).
- Include branches that exist but are unreachable from current callers.

### 3. Data Flow
- What data comes in? In what shape and types?
- How is it transformed at each step? Note each transformation exactly.
- What data goes out? What mutations occurred?
- What state is read? Written? What external systems are called?

### 4. Boundary Behavior
- What happens at null/undefined/empty inputs — only if the code actually encounters them on a reachable path?
- What error handling exists? What errors are caught, propagated, or silently ignored?
- What happens when external dependencies fail — only if the code has handling for it?

### 5. Side Effects

Two tests:
- **Did it talk to something outside itself?** Database, API, file, message queue — reading or writing. If it left the building, it's a side effect.
- **Did it change something that outlasts this operation?** Updated a record, incremented a counter, wrote a log, flipped a flag. If the thing is different afterward, it's a side effect.

Document every external interaction, every event emitted, every mutation of shared state, in the exact order they occur.

### 6. Configuration-Driven Behavior
If behavior changes based on configuration (feature flags, environment variables, config files), document **every path** — not just the currently active one.

### 7. Implicit Behavior
- Default values applied silently
- Type coercions that happen implicitly
- Ordering or sorting that is assumed but not enforced (or is enforced — note which)

### 8. Concurrency
Document the **observable outcome** of concurrent use — not the mechanism. The caller sees correct results or incorrect results.

> Concurrent operations on the same user's balance may produce incorrect results. If two operations run simultaneously, one operation's changes may be lost.

Don't describe locks, transactions, or retry internals.

### 9. Async and Events
If the topic emits an event or triggers async work, the emission is a side effect of your topic. Document that it happens, what data it carries, and when it fires.

What happens on the consumer side is a different topic — unless the code waits for a response and uses it to continue processing.

---

## Scope Boundaries

As you trace behavior, you will follow paths outside your topic. The topic statement is your stopping rule.

**If the behavior still answers the topic statement, it's in scope. The moment it stops, you've hit a boundary.**

At boundaries, document:
- What your topic sends to the external concern
- What your topic receives back
- What assumptions your topic makes about the response

Do not spec the other side.

**The sharpening test:** "Could this behavior change without changing what my topic does?" If yes, it's across a boundary.

---

## Shared Behavior

When your topic's code calls shared utilities that other topics also use, **inline the resulting behavior** in your spec. The reader must understand the full behavior without leaving the page.

When the same shared behavior appears across multiple specs, mark it so maintainers know which specs to update if the shared behavior changes. In the spec, use a callout:

> **Shared with [Decimal Rounding]:** The amount is rounded to two decimal places. When the value is exactly halfway, it rounds to the nearest even digit.

The shared behavior also gets its own canonical spec.

---

## Spec Format Adaptations

Use the standard spec format from [spec-format.md](../../specs/references/spec-format.md). Add these callouts inline within behavior sections where needed:

| Concept | Spec format |
|---|---|
| Surprising or inconsistent behavior | `> **Notable:** [description of the surprising behavior, stated as fact]` |
| Code that exists but is unreachable | `> **Unreachable ([reason]):** [description of the behavior]` |
| Behavior shared across topics | `> **Shared with [Topic Name]:** [inlined behavior description]` |
| Rationale from source comments | `> **Rationale (source comment):** [why this behavior must be preserved, implementation references stripped]` |

These are documentation markers — they carry semantic meaning for maintenance without changing the spec's structure.

---

## Completeness Checklist

Before finalizing a spec, verify every item:

### Scope
- [ ] Topic statement passes the one-sentence test (no "and" joining unrelated capabilities)
- [ ] Boundaries declared with what goes out and what comes back
- [ ] Everything in the spec answers the topic statement

### Tracing
- [ ] Every entry point documented
- [ ] Every conditional branch traced to its terminal point
- [ ] Unreachable paths documented and marked
- [ ] Every data transformation between input and output described

### Side Effects
- [ ] Every external interaction (reads and writes) documented
- [ ] Every state mutation documented
- [ ] Order of side effects matches the implementation

### Error Behavior
- [ ] Every caught error documented
- [ ] Every propagated error documented
- [ ] Every silently ignored error documented
- [ ] Cases where no error handling exists stated explicitly

### Other
- [ ] All configuration-driven paths documented, not just the active one
- [ ] Concurrency outcomes documented (correct or incorrect results)
- [ ] Shared behavior inlined and marked
- [ ] Notable/surprising behavior marked inline
- [ ] Source comment rationale captured (implementation references stripped)

### Final Check
- [ ] Zero function names, class names, variable names, file paths, library names, or framework references in the output
- [ ] A developer on a different stack could reimplement this behavior from the spec alone
