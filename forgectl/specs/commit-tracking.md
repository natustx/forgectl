# Commit Tracking

## Topic of Concern
> The scaffold registers git commit hashes against completed specs.

## Context

After specs are accepted during the specifying phase, the architect registers git commits that contain the spec work. Two commands support this: `add-commit` for explicit registration by spec ID, and `reconcile-commit` for automatic matching based on which files a commit touched. All hashes are validated against the git repository before registration.

These commands are manual operations — they are always available regardless of the `enable_commits` configuration setting. The `enable_commits` flag controls whether the scaffold automatically creates commits and requires `--message` during state transitions; `add-commit` and `reconcile-commit` are for retroactively associating existing commits with specs.

## Depends On
- **spec-lifecycle** — provides the completed specs list with file paths for matching.
- **state-persistence** — reads and writes the state file.

## Integration Points

| Spec | Relationship |
|------|-------------|
| Git repository | Hashes validated via `git cat-file -t`; file lists retrieved via `git show --name-only` |
| spec-lifecycle | Completed specs receive commit hashes (via `commit_hashes` field); ACCEPT auto-commits with `--message` (TODO: not yet implemented) |

---

## Interface

### Inputs

#### CLI Commands

| Command | Flags | Description |
|---------|-------|-------------|
| `add-commit` | `--id N` (required), `--hash <hash>` (required) | Register a commit hash to a completed spec. Hash validated against git. Duplicates rejected. |
| `reconcile-commit` | `--hash <hash>` (required) | Auto-register a commit to all completed specs whose files were touched. Runs `git show --name-only` to match files. |

### Outputs

#### `add-commit` output
Confirms registration: prints the spec name and the registered hash.

#### `reconcile-commit` output
Reports which specs were updated with the commit hash, listing each matched spec by name and file path.

### Rejection

| Condition | Signal | Rationale |
|-----------|--------|-----------|
| `add-commit` / `reconcile-commit` with hash not in git | Error: "commit does not exist in the repository." Exit code 1. | Prevents invalid hashes |
| `add-commit` with hash already registered | Error: "commit already registered." Exit code 1. | Prevents duplicates |
| `add-commit` targeting an active (not completed) spec | Error: "spec is still active." Exit code 1. | Commits for completed specs only |

---

## Behavior

### Commit Hash Validation

All commands that accept a commit hash (`add-commit`, `reconcile-commit`, and the auto-commit in specifying `advance --verdict PASS` when `enable_commits: true`) validate that the hash exists in git using `git cat-file -t`. The object type must be `commit`. Non-existent hashes, tags, blobs, and tree objects are rejected.

### add-commit

Appends a commit hash to a specific completed spec by ID. The hash is validated against git and checked for duplicates before appending.

### reconcile-commit

Runs `git show --name-only <hash>` to determine which files were changed, then matches file paths against `completed[].file`. The hash is appended to every matching spec that doesn't already have it. Reports which specs were updated.

---

## Invariants

1. **Hash validity.** Every stored commit hash references a real git commit object.
2. **No duplicates.** A hash cannot appear twice on the same spec.
3. **Completed specs only.** Commits can only be registered to specs that have reached ACCEPT.

---

## Edge Cases

- **Scenario:** `reconcile-commit` with a hash that touches files not matching any completed spec.
  - **Expected:** No specs updated. Command succeeds with a message indicating zero matches.
  - **Rationale:** Not every commit is spec-related; the command reports rather than fails.

- **Scenario:** `add-commit` with a git tag hash instead of a commit hash.
  - **Expected:** Rejected. `git cat-file -t` returns `tag`, not `commit`.
  - **Rationale:** Only commit objects are valid; tags, blobs, and trees are not accepted to prevent ambiguity.

- **Scenario:** `reconcile-commit` where the commit touches multiple specs, some already have the hash.
  - **Expected:** Hash appended only to specs that don't already have it. Already-registered specs skipped silently.
  - **Rationale:** Idempotent behavior for multi-spec commits; duplicates are prevented per-spec without failing the whole operation.

---

## Testing Criteria

### add-commit registers hash to completed spec
- **Verifies:** Normal registration flow.
- **Given:** Spec 1 is completed.
- **When:** `forgectl add-commit --id 1 --hash abc123`
- **Then:** Hash appears in spec 1's commit list.

### add-commit rejects invalid hash
- **Verifies:** Git validation rejects non-existent hashes.
- **Given:** Hash does not exist in git.
- **When:** `forgectl add-commit --id 1 --hash invalid`
- **Then:** Exit code 1.

### add-commit rejects duplicate hash
- **Verifies:** Duplicate prevention.
- **Given:** Hash already registered to spec 1.
- **When:** `forgectl add-commit --id 1 --hash abc123`
- **Then:** Exit code 1.

### add-commit rejects active spec
- **Verifies:** Completed-only constraint.
- **Given:** Spec 2 is still in progress.
- **When:** `forgectl add-commit --id 2 --hash abc123`
- **Then:** Exit code 1.

### reconcile-commit matches files to specs
- **Verifies:** Automatic file-path matching.
- **Given:** Commit touches `optimizer/specs/repository-loading.md`.
- **When:** `forgectl reconcile-commit --hash abc123`
- **Then:** Hash registered to the spec with that file path.

### reconcile-commit skips already-registered
- **Verifies:** Idempotent behavior across multiple specs.
- **Given:** Hash already on spec 1, commit also touches spec 1's file.
- **When:** `forgectl reconcile-commit --hash abc123`
- **Then:** Spec 1 not duplicated. Other matching specs updated.

---

## Implements
- Commit hash tracking for completed specs (add-commit, reconcile-commit)
- Git-validated hash registration with duplicate prevention
