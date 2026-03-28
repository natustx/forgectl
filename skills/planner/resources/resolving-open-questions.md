# Resolving Open Questions

## What It Is

Every plan may have open questions — decisions that couldn't be made when the plan was written because they required more context, user input, or research. Open questions are not failures of planning; they are markers of intellectual honesty. But they must be actively tracked and systematically resolved, or they accumulate into a system that cannot be built.

## Why It Matters

Open questions are deferred decisions. Every deferred decision is a risk:
- Downstream plans may assume an answer that turns out to be wrong
- Implementation phases can't be defined for plans with unresolved questions
- Open questions breed more open questions — uncertainty compounds

The goal is not to eliminate all open questions immediately, but to have a clear path to resolving each one and to close them before implementation planning begins.

## The Lifecycle of an Open Question

### 1. Created
A plan is written with an open question. The question is specific and answerable — not "how should this work?" but "should Score.overall be auto-computed from the four dimensions, or set independently by the judge?"

### 2. Discussed
The planner raises the question with the user. Present the options, the tradeoffs, and a recommendation. Don't just ask "what do you think?" — frame the decision:

> "Score.overall can be auto-computed (simple average of four dimensions) or judge-set (the LLM reasons about relative importance). Auto-computed is simpler but imposes equal weighting. Judge-set is more flexible but adds another LLM decision point. I recommend judge-set because dimension importance varies by company context."

### 3. Decided
The user makes a call. The planner records the decision.

### 4. Closed
The plan is updated: the open question is removed and replaced with a "Resolved Questions" section or the answer is folded into the proposal. All downstream plans that were waiting on the answer are updated.

## How to Manage Open Questions

### Track them actively
After each planning session, review all plans for open questions. Maintain awareness of what's unresolved.

### Group related questions
Some questions cluster — resolving one resolves others. For example, "how does the optimizer release resources?" and "does the API send a completion signal to the optimizer?" are the same question from two perspectives.

### Research before asking
Some questions can be answered by reading documentation, studying a framework, or tracing the data flow. Don't ask the user to make decisions you can inform through research. In this project, the question "how do optimized instructions flow into generation?" was resolved by reading the optimization framework's documentation — the compiled module IS the generator.

### Present options, not problems
When raising a question to the user, always present:
1. The question, clearly stated
2. The options (at least two)
3. The tradeoffs of each option
4. Your recommendation and why

### Close decisively
Once a question is resolved:
- Update the plan to reflect the decision
- Move the question from "Open Questions" to "Resolved Questions" (or remove the section entirely if no questions remain)
- Update any plans that were blocked by the question
- Check if the resolution creates new questions in other plans

## Example from This Project

**Open question:** "Should rejected ideas be written to workspace output, or only accepted/edited?"

**How it was resolved:**
1. Discussed with the user — presented three reasons to include rejected ideas (documentation of "considered and dismissed," consistency with idea memory corpus, zero cost to include)
2. User agreed
3. Updated workspace output plan: added `rejected/` subdirectory with example file format
4. Removed the open question from the plan
5. Verified no other plans were affected

**Open question:** "Should Score.overall be computed automatically from the four dimensions, or set independently by the judge?"

**How it was resolved:**
1. Discussed with the user — four dimensions aren't equally important across contexts, judge-set allows context-dependent weighting, keeps the door open for configurable weights later
2. User agreed
3. Updated schemas plan: added to "Resolved Questions" section with rationale
4. Removed from "Open Questions"

## Anti-Patterns

- **Vague questions:** "How should errors work?" → Too broad. Split into specific questions.
- **Questions without options:** "What should we do about concurrency?" → Always present at least two options.
- **Zombie questions:** Questions that sit in plans for multiple sessions without being raised → Actively surface them.
- **Answered but not closed:** The user made a decision in conversation but the plan still shows it as open → Always update the plan.

## The Key Principle

**An open question is a promise to make a decision later. Track every promise. Keep every promise.**
