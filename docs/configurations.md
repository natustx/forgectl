# Forgectl Configuration Reference

All configuration is set at session initialization via CLI flags and persisted in the `config` object within `forgectl-state.json`. Configuration is immutable for the session unless noted otherwise.

## State File Structure

```json
{
  "config": {
    "batch_size": 2,
    "min_rounds": 1,
    "max_rounds": 3,
    "reconcile_min_rounds": 0,
    "reconcile_max_rounds": 3,
    "user_guided": true
  },
  "phase": "specifying",
  "state": "ORIENT",
  "started_at_phase": "specifying",
  ...
}
```

## Configuration Parameters

### batch_size

- **CLI Flag:** `--batch-size N`
- **Type:** integer
- **Default:** none (required)
- **Mutable:** no
- **Constraint:** >= 1

Maximum number of unblocked items delivered per batch during the implementing phase. Smaller batches produce more frequent evaluation checkpoints. Larger batches reduce round-trip overhead.

### min_rounds

- **CLI Flag:** `--min-rounds N`
- **Type:** integer
- **Default:** 1
- **Mutable:** no
- **Constraint:** >= 1, <= max_rounds

Minimum evaluation rounds before a PASS verdict can trigger acceptance. Applies to specifying eval, planning eval, and implementing eval. A PASS verdict below this threshold forces another REFINE cycle. Does not apply to reconciliation (see `reconcile_min_rounds`).

### max_rounds

- **CLI Flag:** `--max-rounds N`
- **Type:** integer
- **Default:** none (required)
- **Mutable:** no
- **Constraint:** >= min_rounds

Maximum evaluation rounds before a FAIL verdict forces acceptance. Applies to specifying eval, planning eval, and implementing eval. Prevents indefinite eval loops. Does not apply to reconciliation (see `reconcile_max_rounds`).

### reconcile_min_rounds

- **CLI Flag:** `--reconcile-min-rounds N`
- **Type:** integer
- **Default:** 0
- **Mutable:** no
- **Constraint:** >= 0, <= reconcile_max_rounds

Minimum evaluation rounds for reconciliation before a PASS verdict can complete the specifying phase. When set to 0 (default), a single PASS immediately transitions to COMPLETE — preserving backward-compatible behavior. When set to a positive value, PASS verdicts below this threshold route to RECONCILE_REVIEW for another cycle.

### reconcile_max_rounds

- **CLI Flag:** `--reconcile-max-rounds N`
- **Type:** integer
- **Default:** 3
- **Mutable:** no
- **Constraint:** >= reconcile_min_rounds

Maximum evaluation rounds for reconciliation. A FAIL verdict at this threshold forces completion to prevent indefinite reconciliation loops. Independent of `max_rounds`.

### user_guided

- **CLI Flag:** `--guided` / `--no-guided`
- **Type:** boolean
- **Default:** true
- **Mutable:** yes (on any `advance` call)

When true, the scaffold inserts pause points at SELECT (specifying), REVIEW (planning), and ORIENT (implementing) with "Stop and review and discuss with user before continuing." Toggling this on any `advance` takes effect immediately.

## Defaults

| Parameter | Default | Notes |
|-----------|---------|-------|
| `batch_size` | — | Required at init. No default. |
| `min_rounds` | 1 | At least one eval round per cycle. |
| `max_rounds` | — | Required at init. No default. |
| `reconcile_min_rounds` | 0 | PASS completes immediately. Set > 0 to require multiple reconciliation rounds. |
| `reconcile_max_rounds` | 3 | Caps reconciliation loops. |
| `user_guided` | true | Guided mode on. Disable with `--no-guided`. |

A minimal init uses only the required flags; everything else falls back to defaults:

```bash
forgectl init --from specs-queue.json --batch-size 2 --max-rounds 3
```

Equivalent expanded form with all defaults explicit:

```bash
forgectl init \
  --from specs-queue.json \
  --batch-size 2 \
  --min-rounds 1 \
  --max-rounds 3 \
  --reconcile-min-rounds 0 \
  --reconcile-max-rounds 3 \
  --guided
```

## Non-Config Session Fields

These fields live at the top level of the state file, outside the `config` object. They are session metadata, not tunable parameters.

| Field | Description |
|-------|-------------|
| `phase` | Active phase: `specifying`, `planning`, `implementing` |
| `state` | Current state within the active phase |
| `started_at_phase` | Which phase the session was initialized at (display only) |
| `phase_shift` | Records from/to during PHASE_SHIFT transitions |
