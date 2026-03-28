# Discussion Before Writing

## What It Is

Every plan is a decision. Every decision deserves a conversation before it becomes a document. Discussion-before-writing is the discipline of reaching alignment with the user on the *direction* of a plan before committing it to a file.

This is not about asking permission. It is about ensuring the planner and the user share the same understanding of the problem, the options, and the chosen path.

## Why It Matters

A plan written without discussion has two failure modes:
1. **Wrong direction** — the planner makes assumptions the user disagrees with. The plan is written, reviewed, rejected, and rewritten. Wasted effort.
2. **Missed context** — the user has information the planner doesn't. A conversation surfaces it. A document doesn't.

Discussion is cheap. Rewriting plans is expensive. The ratio of discussion time to writing time should be at least 2:1 for any non-trivial plan.

## The Discussion Pattern

### 1. Frame the problem

State what needs to be decided, not what you think the answer is:

> "The optimizer needs to release the compiled module from memory at some point. When should that happen?"

Not:

> "I'm going to add a RunCompleteCommand to WS2."

### 2. Present options with tradeoffs

Always present at least two options. For each, describe:
- What it does
- What it costs (complexity, performance, coupling)
- What it enables or prevents

Use a simple structure:

> **Option A: Persistent connection** — the API connects once at startup and holds the connection open. Simpler, but requires reconnection logic if the connection drops.
>
> **Option B: Per-run connection** — the API opens a new connection for each run. Cleaner lifecycle, but connection overhead per run.

### 3. State your recommendation and why

Don't hide behind false neutrality. The planner should have an opinion:

> "I recommend Option A. The optimizer is long-lived, so a persistent connection matches the process model. Reconnection logic is straightforward."

### 4. Let the user decide

The user may:
- Agree → proceed to writing
- Disagree → adjust the direction
- Ask questions → deepen the discussion
- Redirect → the problem isn't what you thought it was

All of these are valuable outcomes. None of them are wasted time.

### 5. Write only after alignment

Once the user says "yes, do it" or "let's go with that" — then write the plan. Not before.

## Depth of Discussion

Not every plan needs the same depth of discussion. Calibrate based on:

**Light discussion (state the plan, confirm, write):**
- Straightforward additions with no design choices
- Plans that follow directly from a prior decision
- "I'll add this to the manifest" → "sure"

**Medium discussion (present options, recommend, align):**
- Plans with 2-3 reasonable approaches
- Changes that affect one other plan
- "Should the WS2 connection be persistent or per-run?"

**Deep discussion (explore the problem space, iterate on direction):**
- Architectural decisions that affect multiple domains
- New features that change the system's value proposition
- "What's the most impactful addition to the project?"

## Recognizing When to Stop Discussing and Start Writing

Discussion has diminishing returns. Signs it's time to write:
- The user has said "yes" or "let's do it" or "sounds good"
- The options have been explored and a clear winner emerged
- The user is asking for details that belong in the plan, not the conversation
- You're going in circles

## Example from This Project

**Deep discussion: Regeneration mechanism**

The planner identified 5 possible regeneration flavors, each with different complexity and quality tradeoffs. Rather than picking one and writing the plan:

1. All 5 flavors were presented with tradeoffs
2. The user said "table the fancy ones, pick the simplest"
3. The simplest (naive retry with higher temperature) was planned
4. The other 4 were tabled with comprehensive documentation

Result: a focused plan for v1 *and* a rich tabled document for future work. Both produced in one conversation because the discussion surfaced all the options before writing committed to one.

**Light discussion: Adding rejected ideas to workspace output**

The planner presented three reasons to include them, the user agreed immediately, the plan was updated. Total discussion: 30 seconds. Appropriate depth for a straightforward decision.

## Anti-Patterns

- **Writing first, discussing after:** "Here's the plan I wrote, what do you think?" → The user is now reviewing, not collaborating. The plan biases the conversation.
- **Asking without options:** "What should we do about X?" → Unhelpful. Present options.
- **Over-discussing simple decisions:** Spending 10 minutes on whether a field should be called `reason` or `feedback` → Make a call, move on.
- **Discussion without closure:** A great conversation that doesn't end with "so we're going with X?" → Always close with a decision.

## The Key Principle

**A plan is the artifact. The discussion is the process.** The artifact is only as good as the process that produced it. Discuss first. Write second. Always.
