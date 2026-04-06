# plan.json Schema

> Structured implementation plan manifest created during the **planning phase**
> and consumed during the **implementing phase**.
> Lives at `<domain>/.forge_workspace/implementation_plan/plan.json`.

---

## Root

Only these 4 top-level fields are allowed.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `context` | Context | **yes** | Domain and project metadata. |
| `refs` | Ref[] | no | Reference files (specs, notes). Omitted if empty. |
| `layers` | Layer[] | **yes** | Ordered dependency tiers grouping items. |
| `items` | Item[] | **yes** | Implementation work items. |

---

## Context

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `domain` | string | **yes** | Domain name (matches top-level directory). Non-empty. |
| `module` | string | **yes** | Go module name or package identifier. Non-empty. |

---

## Ref

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | **yes** | Short identifier (e.g., `"spec-logging"`, `"notes-config"`). |
| `path` | string | **yes** | Path relative to the plan.json directory. Must exist on disk. **No `#anchor` fragments** — forgectl runs `os.Stat()` on the raw string. |

**Important:** Refs must be objects `{"id": "...", "path": "..."}`, not strings.

### Path Resolution

Paths in `refs[].path` and `items[].refs` are resolved relative to the **plan.json directory** (`<domain>/<workspace_dir>/implementation_plan/`).

Paths in `items[].files` and `items[].specs` are resolved relative to the **project root** (the directory containing `.forgectl/`).

Example: if plan.json is at `api/.forge_workspace/implementation_plan/plan.json`:
- `refs[].path: "notes/config.md"` resolves to `api/.forge_workspace/implementation_plan/notes/config.md`
- `items[].specs: ["api/specs/foo.md"]` resolves to `<project_root>/api/specs/foo.md`

---

## Layer

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | **yes** | Layer identifier (e.g., `"L0"`, `"L1"`). |
| `name` | string | **yes** | Human-readable name (e.g., `"Foundation"`, `"Core"`). |
| `items` | string[] | **yes** | Item IDs in this layer, in suggested implementation order. |

### Layer Rules

- Every item must appear in exactly one layer.
- Items in layer N may only depend on items in layers 0..N.
- Layer ordering is significant — L0 before L1 before L2.

---

## Item

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | **yes** | Unique identifier. Convention: `<package>.<concern>` (e.g., `"config.load"`). |
| `name` | string | **yes** | Short action-oriented name. |
| `description` | string | **yes** | One or two sentences explaining what and why. |
| `depends_on` | string[] | **yes** | Item IDs that must complete first. Use `[]` for no deps. **Never null.** |
| `specs` | string[] | **yes** | Spec file paths relative to project root. Use `[]` for no specs. **Never null.** No `#anchors`. |
| `refs` | string[] | **yes** | Notes file paths relative to plan.json directory. Use `[]` for no refs. **Never null.** No `#anchors`. |
| `files` | string[] | **yes** | File paths to create/modify, relative to project root. Use `[]` for no files. **Never null.** |
| `steps` | string[] | no | Ordered implementation instructions. Omitted if empty. |
| `tests` | Test[] | **yes** | Acceptance criteria. Use `[]` for items with no tests. **Never null.** |

### Fields Added by Phase Transition

When forgectl transitions from planning to implementing, it adds these to every item:

| Field | Type | Initial | Description |
|-------|------|---------|-------------|
| `passes` | string | `"pending"` | Status: `"pending"` → `"done"` → `"passed"` or `"failed"` |
| `rounds` | int | `0` | Evaluation round counter. Incremented after each eval cycle. |

**Do not include `passes` or `rounds` when drafting.** Forgectl adds them automatically.

---

## Test

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `category` | string | **yes** | One of: `"functional"`, `"rejection"`, `"edge_case"` |
| `description` | string | **yes** | What this test verifies — specific enough to evaluate unambiguously. |

### Test Categories

| Category | Maps to | Purpose |
|----------|---------|---------|
| `functional` | Spec Behavior sections | Core happy-path behavior |
| `rejection` | Spec Rejection table | Invalid inputs rejected correctly |
| `edge_case` | Spec Edge Cases | Boundary conditions handled |

---

## Validation Summary

Forgectl validates plan.json at two points: during DRAFT advance and during implementing PHASE_SHIFT.

1. Only `context`, `refs`, `layers`, `items` allowed at top level.
2. `context.domain` and `context.module` must be non-empty strings.
3. All `refs[].path` and `items[].refs` paths are resolved relative to the plan.json directory and must exist on disk. No `#anchors`.
4. All `items[].files` and `items[].specs` paths are resolved relative to the project root.
5. Item IDs must be unique.
6. Every item must appear in exactly one layer.
7. Items can only depend on same-layer or earlier-layer items.
8. No circular dependencies (DAG enforced via DFS).
9. All `depends_on` IDs must reference existing items.
10. `depends_on`, `specs`, `refs`, `files`, and `tests` must be arrays, never null.
11. Test categories must be `"functional"`, `"rejection"`, or `"edge_case"`.
12. Every path in `items[].refs` must resolve to an existing notes file (relative to plan.json directory).

---

## Example

```json
{
  "context": {
    "domain": "launcher",
    "module": "spectacular/launcher"
  },
  "refs": [
    {"id": "spec-config", "path": "notes/spec-refs/service-configuration.md"},
    {"id": "notes-config", "path": "notes/config.md"}
  ],
  "layers": [
    {"id": "L0", "name": "Foundation", "items": ["config.types", "config.load"]},
    {"id": "L1", "name": "Core", "items": ["daemon.spawn", "daemon.health"]}
  ],
  "items": [
    {
      "id": "config.types",
      "name": "Service config type definitions",
      "description": "Go structs for validated service endpoint configuration.",
      "depends_on": [],
      "specs": ["launcher/specs/service-configuration.md"],
      "refs": ["notes/config.md"],
      "files": ["launcher/internal/config/types.go"],
      "steps": ["Define ServiceEndpoint struct", "Define ServicesConfig struct"],
      "tests": [
        {"category": "functional", "description": "Three named fields, not a map"}
      ]
    },
    {
      "id": "config.load",
      "name": "Load YAML, apply defaults, validate",
      "description": "Parse config file, apply default values, reject invalid config.",
      "depends_on": ["config.types"],
      "specs": ["launcher/specs/service-configuration.md"],
      "refs": ["notes/config.md"],
      "files": ["launcher/internal/config/load.go", "launcher/internal/config/load_test.go"],
      "steps": ["Implement LoadConfig()", "Add default port logic", "Add validation"],
      "tests": [
        {"category": "functional", "description": "Default ports applied when services are empty"},
        {"category": "rejection", "description": "Missing services section rejected"},
        {"category": "edge_case", "description": "Duplicate ports allowed"}
      ]
    },
    {
      "id": "daemon.spawn",
      "name": "Spawn detached process",
      "description": "Start a system process in a new process group.",
      "depends_on": ["config.load"],
      "specs": ["launcher/specs/launching-system-processes.md"],
      "refs": ["notes/daemon.md"],
      "files": ["launcher/internal/daemon/spawn.go"],
      "tests": [
        {"category": "functional", "description": "Process starts in new process group"}
      ]
    },
    {
      "id": "daemon.health",
      "name": "Health check spawned process",
      "description": "Poll process health endpoint until ready or timeout.",
      "depends_on": ["daemon.spawn"],
      "specs": ["launcher/specs/launching-system-processes.md"],
      "refs": ["notes/daemon.md"],
      "files": ["launcher/internal/daemon/health.go"],
      "tests": [
        {"category": "functional", "description": "Returns healthy after endpoint responds"},
        {"category": "edge_case", "description": "Times out after configured duration"}
      ]
    }
  ]
}
```

---

## Source

- Type definitions: `forgectl/state/types.go` (`PlanJSON`, `PlanItem`, `PlanLayerDef`, `PlanRef`, `PlanTest`)
- Validation: `forgectl/state/validate.go` (`ValidatePlanJSON`)
- Phase mutation: `forgectl/state/advance.go` (`advancePhaseShift`)
- Human-readable format: `forgectl/PLAN_FORMAT.md`
- Companion schema-shaped reference: `skills/implementation_planning/references/plan-format.json`
