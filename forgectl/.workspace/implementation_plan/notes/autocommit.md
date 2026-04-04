# Auto-Commit Notes

## When Commits Fire

Commits are gated on `config.general.enable_commits`. When false: --message is not required; if provided, print warning and ignore.

| Event | Commit trigger | Commit strategy config |
|-------|---------------|----------------------|
| specifying COMPLETE | Single commit for all spec work | `specifying.commit_strategy` |
| planning ACCEPT | One commit per plan | `planning.commit_strategy` |
| implementing COMMIT | One commit per batch | `implementing.commit_strategy` |

## Commit Strategies

Each strategy determines which files are staged (`git add`) before `git commit`:

| Strategy | git add behavior |
|----------|-----------------|
| `strict` | `git add` only the specific files listed in the plan item `files` fields (implementing) or spec `file` paths (specifying). |
| `all-specs` | `git add` all spec files in the completed domain(s). E.g., `git add optimizer/specs/` |
| `scoped` | `git add` all files within the domain directory. E.g., `git add optimizer/` |
| `tracked` | `git add -u` — only stage changes to already-tracked files. |
| `all` | `git add -A` — stage everything in the working tree. |

## Execution Steps

```
1. Determine files to stage (per commit_strategy).
2. Run: git -C <project_root> add <files or flags>
3. Run: git -C <project_root> commit -m <message>
4. Capture stdout to extract commit hash.
5. Return hash on success; return error on failure.
```

Parsing commit hash from `git commit` output:
```
[branch <hash>] message
```
Use regex or string split to extract the short hash from the first line.

Or run `git rev-parse HEAD` after commit to get the full hash.

## Hash Registration (specifying COMPLETE)

After successful commit at COMPLETE:
1. For each `s.Specifying.Completed` spec:
   - Append the commit hash to `spec.CommitHashes`.
   - Set `spec.CommitHash` (single hash field for backward compat) to the hash.
2. Save state.

This replaces the add-commit and reconcile-commit commands entirely.

## Auto-Commit for Specifying

The specifying phase produces ONE commit at COMPLETE covering all spec work. This is triggered in advanceSpecifying when advancing from COMPLETE → PHASE_SHIFT:

```
if enable_commits && message != "" {
    hash, err := GitAutoCommit(projectRoot, specifyingCommitStrategy, message)
    if err: return error
    registerHashOnAllCompletedSpecs(s, hash)
}
```

Staging for specifying with `all-specs` strategy: stage all completed spec files.
```
git add <spec.file for each completed spec>
```

## Auto-Commit for Planning

At planning ACCEPT, if enable_commits=true:
- Stage the plan.json and all notes files.
- Commit with --message.

For `strict` strategy: stage `currentPlan.file` + all notes referenced in the plan.

## Auto-Commit for Implementing

At implementing COMMIT, if enable_commits=true:
- Stage files per commit_strategy (scoped by default = entire domain dir).
- Commit with --message.

## Error Handling

If git commit fails (non-zero exit):
- Print error: "git commit failed: <stderr>".
- State does NOT advance. Exit code 1.
- The --message flag stays required; user must fix and re-run advance.

## Function Signatures

```go
// AutoCommit stages files and commits. Returns commit hash.
func AutoCommit(projectRoot string, strategy string, stageTargets []string, message string) (string, error)

// stageTargets is a list of paths or flags depending on strategy:
// - strict: list of file paths
// - all-specs: list of spec directory paths
// - scoped: list of domain directories
// - tracked: nil (uses -u flag)
// - all: nil (uses -A flag)
```

## Removed Functions

The following git.go functions are no longer needed after removing add-commit/reconcile-commit:
- `AddCommitToSpec(s, specID, hash)` — replaced by hash registration in COMPLETE handler
- `ReconcileCommit(s, workDir, hash)` — removed entirely
- `GitShowFiles(workDir, hash)` — removed (only used by ReconcileCommit)
- `GitHashExists(workDir, hash)` — can be kept or removed (no current callers after cleanup)
