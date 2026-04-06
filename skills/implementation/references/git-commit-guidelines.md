# Git Commit Guidelines

Do NOT run `git add`, `git commit`, or any git write operations. The user manages all commits.

Your responsibility is to implement code and ensure tests pass. The user decides when and what to commit.

## Commit Message Rules (for `--message` flag)

When `enable_commits` is true in `.forgectl/config`, `forgectl advance` requires a `--message` flag at certain states. This records the message in the state file — it does NOT execute a git commit.

- Write a descriptive message summarizing the changes.
- Do not mention that you are an AI, LLM, or Claude Code in the message.
- Do not append `Co-Authored-By: <model> <noreply@anthropic.com>` to the message.
