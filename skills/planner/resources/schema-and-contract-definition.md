# Schema and Contract Definition

## What It Is

Schemas and contracts are the connective tissue between domains. They define the exact shape of data as it crosses boundaries — between services, between processes, between layers. Unlike behavioral plans that describe *what happens*, schemas describe *what the data looks like* at every boundary crossing.

Schema definition is a distinct planning activity. It is not just "another plan" — it is the activity that forces precision on every other plan. A plan can say "the optimizer sends ideas to the API." A schema says exactly what an idea contains, field by field, with types and constraints.

## Why It Matters

Vague plans produce vague systems. Two teams implementing the same plan will produce incompatible interfaces if the data contract isn't defined. Schemas eliminate ambiguity:

- "Service A sends results to Service B" → ambiguous
- "Service A sends a list of `Result` objects, each containing an `id` (string), `payload` (Payload object with `title`, `body`), `score` (number, 0-10), and `status` (enum: pending, approved, rejected)" → unambiguous

## When to Do It

After the major plans are written but before implementation planning. Schemas sit between planning and implementation — they are technology-agnostic (no language-specific types) but concrete (exact field names, types, and constraints).

## Types of Contracts to Define

### Data Models
The core objects the system works with. These are the nouns of the system — the entities that multiple components must agree on.

### Message Schemas
Every message that crosses a process boundary. Group them by connection (e.g., frontend ↔ API, API ↔ backend worker). Each message needs: a type discriminator, the exact payload fields, and which direction it flows.

### State Definitions
Enums and state machines that multiple components must agree on. Status enums, lifecycle states, and any shared constants.

### Storage vs Transport Distinction

A critical insight: what you store internally is not always what you send over the wire. Define both:

- **Transport models** — the shape of data as it crosses boundaries. Optimized for readability and structure (nested objects, clean hierarchy).
- **Storage models** — the shape of data as it's held internally. Optimized for queryability (flat fields, timestamps, lineage tracking, audit fields).

The service layer converts between them. Define the conversion explicitly so both sides of every boundary agree.

## How to Define Schemas

### 1. Start from the plans

Read every plan and extract every data object mentioned. List them all, then group by:
- Objects that exist only inside one domain (internal — define but don't export)
- Objects that cross domain boundaries (contracts — these must be precisely defined)

### 2. Define the cross-boundary objects first

These are the contracts. Use a notation that is language-agnostic but precise:

```
class Order:
    id: string
    items: list[OrderItem]
    total: number
    status: OrderStatus = pending
```

Include types, defaults, and constraints. Don't use language-specific syntax (no `Optional[str]` or `*string`). Use plain types: `string`, `number`, `boolean`, `list[X]`, `X | null`.

### 3. Define message schemas with direction and discriminator

```
Command: cancel_order (Frontend → API)
    order_id: string
    reason: string
```

### 4. Create discriminated unions

Group all messages by connection and direction so the system can parse any incoming message by looking at the `type` field.

### 5. Validate by tracing a flow

Pick a user action and trace the data through every schema boundary:
- Frontend sends command → fields match the schema?
- API translates to backend command → fields match?
- Backend sends result event → fields match?
- API translates to frontend event → fields match?
- Frontend updates its store → fields match?

If any step has a mismatch, the schema is wrong.

## State Machine Design

State machines deserve special attention. One effective approach is a positional numeric scheme:

| Place | Meaning |
|-------|---------|
| Thousands | Phase — major stage |
| Hundreds | Step — discrete step within a phase |
| Tens | Sub-step — granular progress (future use) |
| Ones | Detail — finest-grain status (future use) |

This system has two important properties:
1. **Codes are comparable** — `state_code >= 5000` means "at or past review"
2. **Gaps are intentional** — reserved ranges (2000-2999) allow future phases without renumbering

When designing a state machine:
- Define every state with a code and a human-readable name
- Identify which process owns each state transition
- Mark cancel checkpoints explicitly
- Reserve gaps for future states

## Dependency Audits

After defining schemas, audit the full dependency graph of every model. For each class, list what other classes it references. Then check for cycles.

**Why this matters:** JSON serialization will infinite-loop on circular references. A cycle like `A → B → C → A` means serializing `A` requires serializing `B`, which requires `C`, which requires `A` again — stack overflow.

**How to audit:**
1. List every model as `ModelName → [dependencies]`
2. Walk the graph looking for cycles
3. Identify the longest chain (this affects nesting depth in JSON)
4. Record the audit date and result in the schema document

All dependencies should form a directed acyclic graph (DAG). If a cycle is found, restructure — typically by extracting the shared dependency into its own model or by using an ID reference instead of embedding.

## Embedding Pattern for Shared Models

When one model is a superset of another, **embed** the smaller model rather than duplicating its fields. This prevents drift — if the embedded model gains a field, the parent gets it automatically.

**When to embed:**
- A storage model extends a transport model (e.g., `StoredOrder` embeds `Order` + adds internal tracking fields)
- An event carries a full model plus additional context (e.g., `UpdatedOrderEvent` embeds `Order` + adds `audit: AuditContext`)
- A client-facing event extends a backend event (e.g., `PortalStatusEvent` embeds `StatusEvent` + adds `allowed_actions`)

**When NOT to embed:**
- The two models share some fields but have different semantics
- Embedding would create a circular dependency
- The models are in different domains with no shared contract

In planning documents, notate embedding by listing the embedded model name without a field name:

```
class StoredOrder:
    Order                       # embedded — all fields promoted
    internal: OrderInternal     # named field
    audit: AuditContext         # named field
```

## Naming Conventions for Message Types

Establish a consistent naming pattern for all messages in shared contracts. Inconsistent naming creates cognitive overhead when tracing messages across system boundaries.

**The `verb_noun` pattern:**
- Events (things that happened): `created_run`, `updated_idea`, `regenerated_idea`, `status_update`
- Commands (things to do): `create_run`, `accept_idea`, `regenerate_idea`, `start_run`

Both follow `verb_noun` — events use past tense (`created`, `updated`), commands use imperative (`create`, `accept`).

**Why consistency matters:** When tracing a message end-to-end (e.g., backend sends `completed_task` → API translates → frontend receives `completed_task`), consistent naming makes the flow obvious. If one connection used `completed_task` but another used `task_completed`, developers would need to remember the flip at every boundary.

**When two connections share the same event name** but with different payloads, disambiguate with a prefix. For example, `StatusUpdateEvent` (backend → API) and `PortalStatusUpdateEvent` (API → frontend, with additional fields). The wire format (`"status_update"`) can stay the same since the connections are separate — the class name disambiguates in code.

## The Key Principle

**If two plans mention the same data, that data needs a schema.** The schema is the contract that prevents the plans from silently disagreeing about what the data looks like.
