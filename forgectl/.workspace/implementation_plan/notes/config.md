# Config Notes

## .forgectl/config TOML Structure

The config file lives at `<project_root>/.forgectl/config`. It is TOML format.

### Project Root Discovery

Walk up from `os.Getwd()` (or `--dir` flag if provided) until `.forgectl/` is found. If not found: error "No .forgectl directory found." Exit code 1.

All relative paths in the state file resolve from the project root.

### TOML Structure

```toml
[general]
enable_commits = false
user_guided    = false

[[domains]]
name = "optimizer"
path = "optimizer"

[[domains]]
name = "portal"
path = "portal"

[specifying]
batch           = 3
commit_strategy = "all-specs"

[specifying.eval]
min_rounds        = 1
max_rounds        = 3
model             = "opus"
type              = "eval"
count             = 1
enable_eval_output = false

[specifying.cross_reference]
min_rounds  = 1
max_rounds  = 2
model       = "haiku"
type        = "explore"
count       = 3
user_review = false

[specifying.cross_reference.eval]
model = "opus"
type  = "eval"
count = 1

[specifying.reconciliation]
min_rounds  = 0
max_rounds  = 3
model       = "opus"
type        = "eval"
count       = 1
user_review = false

[planning]
batch                      = 1
commit_strategy            = "strict"
self_review                = false
plan_all_before_implementing = false

[planning.study_code]
model = "haiku"
type  = "explore"
count = 3

[planning.eval]
min_rounds        = 1
max_rounds        = 3
model             = "opus"
type              = "eval"
count             = 1
enable_eval_output = false

[planning.refine]
model = "opus"
type  = "refine"
count = 1

[implementing]
batch           = 2
commit_strategy = "scoped"

[implementing.eval]
min_rounds        = 1
max_rounds        = 3
model             = "opus"
type              = "eval"
count             = 1
enable_eval_output = false

[paths]
state_dir     = ".forgectl/state"
workspace_dir = ".forge_workspace"

[logs]
enabled        = true
retention_days = 90
max_files      = 50
```

### Defaults

All fields have defaults; .forgectl/config is optional (scaffold uses defaults if missing or fields are absent).

| Field | Default |
|-------|---------|
| general.enable_commits | false |
| general.user_guided | false |
| specifying.batch | 3 |
| specifying.commit_strategy | "all-specs" |
| specifying.eval.min_rounds | 1 |
| specifying.eval.max_rounds | 3 |
| specifying.eval.model | "opus" |
| specifying.eval.type | "eval" |
| specifying.eval.count | 1 |
| specifying.eval.enable_eval_output | false |
| specifying.cross_reference.min_rounds | 1 |
| specifying.cross_reference.max_rounds | 2 |
| specifying.cross_reference.model | "haiku" |
| specifying.cross_reference.type | "explore" |
| specifying.cross_reference.count | 3 |
| specifying.cross_reference.user_review | false |
| specifying.cross_reference.eval.model | "opus" |
| specifying.cross_reference.eval.type | "eval" |
| specifying.cross_reference.eval.count | 1 |
| specifying.reconciliation.min_rounds | 0 |
| specifying.reconciliation.max_rounds | 3 |
| specifying.reconciliation.model | "opus" |
| specifying.reconciliation.type | "eval" |
| specifying.reconciliation.count | 1 |
| specifying.reconciliation.user_review | false |
| planning.batch | 1 |
| planning.commit_strategy | "strict" |
| planning.self_review | false |
| planning.plan_all_before_implementing | false |
| planning.study_code.model | "haiku" |
| planning.study_code.type | "explore" |
| planning.study_code.count | 3 |
| planning.eval.min_rounds | 1 |
| planning.eval.max_rounds | 3 |
| planning.eval.model | "opus" |
| planning.eval.type | "eval" |
| planning.eval.count | 1 |
| planning.eval.enable_eval_output | false |
| planning.refine.model | "opus" |
| planning.refine.type | "refine" |
| planning.refine.count | 1 |
| implementing.batch | 2 |
| implementing.commit_strategy | "scoped" |
| implementing.eval.min_rounds | 1 |
| implementing.eval.max_rounds | 3 |
| implementing.eval.model | "opus" |
| implementing.eval.type | "eval" |
| implementing.eval.count | 1 |
| implementing.eval.enable_eval_output | false |
| paths.state_dir | ".forgectl/state" |
| paths.workspace_dir | ".forge_workspace" |
| logs.enabled | true |
| logs.retention_days | 90 |
| logs.max_files | 50 |

### Validation at Init

- `specifying.commit_strategy` must be one of: `strict`, `all-specs`, `scoped`, `tracked`, `all`
- `planning.commit_strategy` must be one of: same set
- `implementing.commit_strategy` must be one of: same set
- `specifying.eval.min_rounds <= specifying.eval.max_rounds`
- `planning.eval.min_rounds <= planning.eval.max_rounds`
- `implementing.eval.min_rounds <= implementing.eval.max_rounds`
- No domain path is a prefix of another domain path (nested paths rejected)
- When `--phase specifying` and domains are configured: spec queue entry domains must match configured domain names; spec file paths must start with domain.path + "/specs/"
- `--phase generate_planning_queue` is rejected: error "generate_planning_queue requires a completed specifying phase. Use --phase specifying instead."

### TOML Library

Add dependency: `github.com/BurntSushi/toml` (or similar well-adopted TOML library).

### Session ID (UUID v4)

Generated at init using `crypto/rand`. Format: `xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx`.

No external dependency needed — implement with `crypto/rand.Read()` and manual bit manipulation.

### `--dir` flag removal / project root discovery

The current `--dir` flag on root command points to the directory containing the state file. After implementing project root discovery, the state file location is derived from the config (`paths.state_dir`), so `--dir` should be replaced by root-discovery logic.

Keep `--dir` as a fallback override for the root directory search starting point if needed, but make it optional (default to current working directory for root search).
