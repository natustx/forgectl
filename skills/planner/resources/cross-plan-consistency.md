# Cross-Plan Consistency

## What It Is

When you change one plan, other plans that reference, depend on, or assume the same behavior may need to change too. Cross-plan consistency is the discipline of tracing the impact of every decision across the full plan set and updating everything that's affected.

## Why It Matters

Plans form a graph of dependencies. Changing a node without updating its neighbors creates silent contradictions. A plan that says "the API spawns the optimizer per-run" while another says "the optimizer is a long-lived process started by the launcher" is a system that cannot be built.

Unlike code, plans don't have a compiler to catch inconsistencies. The planner is the compiler.

## When to Do It

Every time you:
- Change an architectural decision (e.g., subprocess → long-lived process)
- Add a new plan that affects an existing flow
- Resolve an open question that other plans were waiting on
- Rename a concept, state, or data field
- Change the owner of a responsibility

## How to Do It

### 1. Identify the blast radius

Before writing the change, list every plan that could be affected. Ask:
- Which plans mention the thing I'm changing?
- Which plans depend on the behavior I'm changing?
- Which plans assume the old design?

### 2. Make the primary change

Update the plan where the decision lives.

### 3. Walk the dependency graph

Go to every affected plan and update it. Don't do this from memory — re-read each plan and verify consistency.

### 4. Update the schemas

If the change affects data shapes, update the schema plan. Schemas are the highest-leverage consistency point — if the schema is wrong, everything downstream is wrong.

### 5. Update the manifest

If plans were added, removed, or renamed, update `SPEC_MANIFEST.md`.

## Example from This Project

**Decision:** The optimizer changes from a subprocess spawned per-run to a long-lived process started by the launcher.

**Blast radius:**
1. **Engine and process isolation** — primary plan, complete rewrite. Subprocess model replaced with long-lived service model.
2. **WS message protocol** — WS2 connection model changed from per-run to persistent. Added connection model documentation.
3. **Run state management** — API no longer spawns the optimizer. Removed subprocess lifecycle from API responsibilities.
4. **Optimization pipeline** — worker thread now holds the compiled module across review phase, not just during optimization.
5. **Regeneration flow** — simplified. Optimizer is alive during review, can handle regeneration directly.
6. **Launcher** — new plan created. The launcher now starts the optimizer.
7. **Schemas** — added `RunCompleteCommand` to WS2 so the API can tell the optimizer when to release resources.

**Seven plans touched by one decision.** If only the engine isolation plan had been updated, six plans would have silently contradicted it.

## The Checklist

When making a change, run through:

- [ ] Updated the primary plan
- [ ] Updated all plans that depend on the changed plan
- [ ] Updated all plans that reference the changed concept
- [ ] Updated schemas if data shapes changed
- [ ] Updated the manifest if plans were added/removed
- [ ] Re-read the changed plans to verify they're internally consistent
- [ ] Verified no plan references old terminology or behavior

## The Key Principle

**A plan is only as correct as its relationship to every other plan.** Internal consistency is necessary but not sufficient — the plan must also be consistent with the system.
