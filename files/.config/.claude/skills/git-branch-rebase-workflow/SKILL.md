---
name: git-branch-rebase-workflow
description: Enforce a consistent Git workflow for repository work. Use when Claude needs to start implementation, fix bugs, or sync with origin in a Git repository and should (1) create a new working branch from the repository's default branch and (2) bring origin changes into local branches with rebase instead of merge.
---

# Git Branch Rebase Workflow

## Overview

Follow this skill to keep local Git work aligned with the repository default branch and to avoid unnecessary merge commits during routine synchronization.

## Core Rules

- Start every new task from the repository's default branch.
- Create a new working branch before editing files.
- When incorporating origin updates into a local branch, use rebase instead of merge.
- Do not commit directly on the default branch unless the user explicitly asks for it.

## Identify the Default Branch

Prefer discovering the remote default branch instead of guessing.

```bash
git symbolic-ref refs/remotes/origin/HEAD
```

If that is unavailable, inspect the remote:

```bash
git remote show origin
```

Treat the branch pointed to by `origin/HEAD` as the base branch for new work.

## Start New Work

1. Fetch the latest remote state.
2. Check out the default branch.
3. Rebase the local default branch onto its remote tracking branch.
4. Create a new branch from the updated default branch.

Example:

```bash
git fetch origin
git checkout main
git pull --rebase origin main
git checkout -b feature/my-task
```

Replace `main` with the actual default branch name when it differs.

## Sync an Existing Working Branch

When the current working branch falls behind `origin/HEAD`, stay on that working branch, prune remote-tracking refs, and rebase the branch onto the latest default branch tip.

Typical flow:

```bash
git fetch --prune
git rebase main
```

Replace `main` with the actual default branch name when it differs.

Use this flow specifically to bring the current working branch up to date with the branch referenced by `origin/HEAD` without creating a merge commit.

## Conflict Handling

- Stop at conflicts and resolve them deliberately.
- After resolving files, continue with `git rebase --continue`.
- If the rebase path is wrong or unsafe, use `git rebase --abort` and reassess before proceeding.

## Output Expectations

When using this skill in a task:

- State which branch was treated as the default branch.
- State the new working branch name when one is created.
- Mention that synchronization used rebase.
- Call out any conflicts or cases that were left unverified.
