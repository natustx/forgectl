# Strategic Reassessment

## What It Is

Strategic reassessment is the practice of periodically stepping back from the details to ask: "Given everything we've planned, what is the most impactful thing to do next?" It is the antidote to tunnel vision — the tendency to keep refining what's in front of you instead of asking whether you're working on the right thing.

## Why It Matters

Planning is seductive. You can always find another detail to specify, another edge case to consider, another diagram to draw. But not all planning work is equally valuable. Strategic reassessment forces you to evaluate the plan set as a whole and identify where the highest-leverage work is.

This is especially important after:
- Completing a major planning milestone (e.g., all core plans written)
- Resolving a batch of open questions
- Making an architectural change that ripples across plans
- The user shifting priorities or context

## How to Do It

### 1. Summarize the current state

In a few sentences, describe what exists: how many plans, which domains are covered, what's tabled, what's resolved vs. open. This is for your own clarity — you need to see the forest before evaluating the trees.

### 2. Ask the strategic question

Frame it relative to the project's goals:
- "What's the single most impactful addition to this system?"
- "What gap, if left unresolved, will cause the most pain during implementation?"
- "What's the riskiest assumption in the current plans?"
- "Where are we over-planning vs. under-planning?"

### 3. Answer honestly, even if it's uncomfortable

Sometimes the most impactful thing is to stop planning and start building. Sometimes it's to throw away a plan that isn't working. Sometimes it's to table an entire domain because it's not needed for v1.

### 4. Present the assessment to the user

Don't just say "I think we should do X." Explain:
- What you evaluated
- What you considered
- Why X is the highest-leverage option
- What the alternative would be

### 5. Let the user redirect

Strategic reassessment is a moment for the user to steer. They may agree, or they may have priorities you don't know about. The planner proposes; the user decides.

## Example from This Project

**After completing the core plans (optimizer, API, portal, WS protocol):**

Strategic question: "What's the single smartest and most radically innovative addition to the project?"

Assessment considered:
- Auto-implementation of accepted ideas (rejected — Spectacular is an idea generator, not an implementor)
- Idea memory / compounding runs (accepted — makes each run more valuable than the last)

The idea memory feature was added because it was *accretive* — it makes the system more valuable with use, which is the most defensible kind of feature.

**After adding idea memory and resolving gaps:**

Strategic question: "What's the most impactful addition at this point?"

Assessment considered:
- Idea synthesis / funneling pipeline (tabled — compelling but complex, not needed for v1)

The concept was tabled with a comprehensive document so it can be revisited without re-deriving the analysis.

## When NOT to Reassess

- In the middle of writing a plan (finish the plan first)
- When only one obvious next step exists (just do it)
- When the user has given clear direction (follow it)

Strategic reassessment is a tool for decision points, not a procrastination mechanism.

## The Key Principle

**The most important planning decision is what to plan next.** Everything else is execution.
