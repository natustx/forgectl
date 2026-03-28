# Gap Analysis

## What It Is

Gap analysis is the practice of tracing data flow end-to-end across all plans to find seams — places where two plans don't connect, where a handoff is assumed but never defined, or where behavior falls between the cracks of adjacent plans.

Plans are written one at a time, scoped to a single topic of concern. This is intentional — it keeps them small and focused. But it means the spaces *between* plans are where problems hide. Gap analysis is how you find them.

## When to Do It

After a set of plans covers a domain or a full system flow. Don't gap-analyze a single plan in isolation — gaps only appear at boundaries.

Good triggers:
- "We've planned the optimizer, API, and portal. What's missing?"
- "We've defined the WS protocol and the run lifecycle. Do they connect?"
- "We added a new feature (idea memory). What does it touch?"

## How to Do It

### 1. Pick a user-visible flow and trace it end-to-end

Start with something concrete: "The user clicks 'Start Run' and eventually sees ideas to review." Then walk through every plan that flow touches, in order:

```
Portal sends create_run → API receives it → API creates a run →
API sends start_run to optimizer → Optimizer clones repo →
Optimizer builds snapshot → Optimizer runs optimization →
Optimizer generates ideas → Optimizer scores ideas →
Optimizer sends ideas_ready → API stores ideas →
API sends ideas_ready to portal → Portal displays ideas
```

At each `→`, ask: "Is this transition defined in a plan? Does the sending plan's output match the receiving plan's expected input?"

### 2. Look for these specific gap types

| Gap Type | What It Looks Like | Example |
|----------|-------------------|---------|
| **Undefined handoff** | Plan A produces output, Plan B expects input, but no plan defines how A's output becomes B's input | Optimization produces "optimized instructions" but idea generation expects a "compiled module" — what is the artifact? |
| **Missing state transition** | The state machine has a gap — no plan owns the transition between two states | Who transitions the run from `reviewing` to `completed`? |
| **Implicit behavior** | Something must happen but no plan says it does | The optimizer holds the compiled module in memory during review — but no plan says when it releases it |
| **Schema mismatch** | Two plans describe the same data differently | One plan says Score has four dimensions, another treats it as a single number |
| **Cross-domain ownership** | A behavior spans two domains and neither plan claims it | Regeneration touches Portal, API, and Optimizer — which plan owns the end-to-end flow? |

### 3. Enumerate gaps, then prioritize

List every gap you find. Then categorize:
- **Structural** — the system can't work without resolving this (e.g., no plan defines the run lifecycle)
- **Medium** — the system can work but a seam is undefined (e.g., no confirmation event after accept)
- **Low** — forward-compatibility concern (e.g., rejection reason not forwarded for future use)

### 4. Resolve gaps by creating plans or updating existing ones

Each gap becomes either:
- A new plan (if the gap is a missing topic of concern)
- An update to an existing plan (if the gap is an omission within a covered topic)
- A tabled item (if the gap is real but not blocking)

## Example from This Project

After planning the optimizer, API, portal, and WS protocol, a gap analysis found 8 gaps:

1. **Run lifecycle state machine** — no plan owned the full lifecycle → created a new plan
2. **Subprocess bootstrapping** — no plan defined how the optimizer starts → led to the launcher plan
3. **Snapshot ↔ isolation boundary** — resolved by making the optimizer long-lived (snapshot built internally)
4. **Optimized instructions → generation handoff** — resolved by researching how the optimization framework actually works (the compiled module IS the generator)
5. **Ideas → portal display** — resolved by defining push semantics in the WS protocol
6. **Regeneration end-to-end** — spanned 4 plans, needed its own cross-cutting plan
7. **"Proceed" after review** — dead end in existing plans → resolved by workspace output plan
8. **Error recovery** — tabled for later

The most structural gap (#1) was resolved first because every other gap hung off it.

## The Key Principle

**Plans are the nodes. Gaps are the edges.** A plan that is internally perfect but doesn't connect to its neighbors is incomplete. Gap analysis is how you build the graph.
