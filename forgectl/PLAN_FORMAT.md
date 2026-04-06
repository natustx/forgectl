# PLAN_FORMAT

This document defines the `plan.json` format consumed by forgectl during the
planning and implementing phases.

This is the human-readable companion to
`skills/implementation_planning/references/plan-format.json`.

Source of truth:
- `forgectl/state/types.go`
- `forgectl/state/validate.go`
- `forgectl/state/advance.go`

## Purpose

`plan.json` is a structured implementation manifest. It answers:

> What concrete implementation work is required to fully implement these specs?

Forgectl validates this file during planning and mutates it during the planning
to implementing phase shift by adding runtime tracking fields.

Location convention:

```text
<domain>/.forge_workspace/implementation_plan/plan.json
```

## Top-level structure

Only these top-level fields are allowed:

```json
{
  "context": { ... },
  "refs": [ ... ],
  "layers": [ ... ],
  "items": [ ... ]
}
```

- `context`: required
- `refs`: optional
- `layers`: required
- `items`: required

## Context

```json
"context": {
  "domain": "launcher",
  "module": "spectacular/launcher"
}
```

- `domain`: required, non-empty string
- `module`: required, non-empty string

No other `context` fields are recognized by forgectl.

## Refs

`refs` is an optional array of reference objects.

```json
"refs": [
  { "id": "spec-config", "path": "../../specs/service-configuration.md" },
  { "id": "notes-config", "path": "notes/config.md" }
]
```

Rules:
- Refs must be objects, not strings.
- `id` is a short identifier for the reference.
- `path` is resolved relative to the directory containing `plan.json`.
- `path` must exist on disk.
- `path` must not contain `#anchor` fragments.

## Layers

Layers define dependency tiers.

```json
"layers": [
  { "id": "L0", "name": "Foundation", "items": ["config.types", "config.load"] },
  { "id": "L1", "name": "Core", "items": ["daemon.spawn", "daemon.health"] }
]
```

Rules:
- Every layer needs `id`, `name`, and `items`.
- Every item must appear in exactly one layer.
- Items may depend only on items in the same or earlier layer.
- Layer order is significant.

## Items

Each item is a unit of implementation work.

```json
{
  "id": "config.load",
  "name": "Load YAML, apply defaults, validate",
  "description": "Parse config file, apply default values, reject invalid config.",
  "depends_on": ["config.types"],
  "steps": ["Implement LoadConfig()", "Add validation"],
  "files": ["internal/config/load.go", "internal/config/load_test.go"],
  "spec": "service-configuration.md",
  "ref": "notes/config.md",
  "tests": [
    { "category": "functional", "description": "Default ports applied when services are empty" },
    { "category": "rejection", "description": "Missing services section rejected" }
  ]
}
```

Required fields:
- `id`
- `name`
- `description`
- `depends_on`
- `tests`

Optional fields:
- `steps`
- `files`
- `spec`
- `ref`

Rules:
- `id` must be unique across all items.
- `depends_on` must be an array, never `null`.
- `tests` must be an array, never `null`.
- `spec` is a single string, not an array.
- `ref` is resolved relative to `plan.json` and must exist on disk.
- `ref` must not contain a `#anchor`.

## Tests

Each item’s `tests` array defines acceptance criteria.

Allowed categories:
- `functional`
- `rejection`
- `edge_case`

Each test entry requires:
- `category`
- `description`

Example:

```json
{ "category": "edge_case", "description": "Times out after configured duration" }
```

## Runtime fields added by forgectl

When forgectl transitions from planning to implementing, it adds these fields
to every item:

```json
{
  "passes": "pending",
  "rounds": 0
}
```

Do not include `passes` or `rounds` while drafting the plan.

## Validation rules

Forgectl enforces these checks:

1. The JSON parses successfully.
2. Only `context`, `refs`, `layers`, and `items` are present at the top level.
3. `context.domain` and `context.module` are non-empty.
4. Every `refs[].path` exists.
5. Every item has the required fields.
6. Item IDs are unique.
7. Every item appears in exactly one layer.
8. No layer references a non-existent item.
9. No item depends on a later-layer item.
10. No item depends on a non-existent item.
11. The dependency graph is acyclic.
12. Every `items[].ref` path exists.
13. Every test category is `functional`, `rejection`, or `edge_case`.

## Minimal complete example

```json
{
  "context": {
    "domain": "launcher",
    "module": "spectacular/launcher"
  },
  "refs": [
    { "id": "spec-config", "path": "../../specs/service-configuration.md" },
    { "id": "notes-config", "path": "notes/config.md" }
  ],
  "layers": [
    { "id": "L0", "name": "Foundation", "items": ["config.types"] },
    { "id": "L1", "name": "Core", "items": ["config.load"] }
  ],
  "items": [
    {
      "id": "config.types",
      "name": "Service config type definitions",
      "description": "Define validated service endpoint configuration types.",
      "depends_on": [],
      "files": ["internal/config/types.go"],
      "tests": [
        { "category": "functional", "description": "Type definitions match the spec contract" }
      ]
    },
    {
      "id": "config.load",
      "name": "Load and validate configuration",
      "description": "Parse config input, apply defaults, and reject invalid values.",
      "depends_on": ["config.types"],
      "files": ["internal/config/load.go", "internal/config/load_test.go"],
      "spec": "service-configuration.md",
      "ref": "notes/config.md",
      "tests": [
        { "category": "functional", "description": "Valid config loads successfully" },
        { "category": "rejection", "description": "Invalid config is rejected with an error" }
      ]
    }
  ]
}
```
