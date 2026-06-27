---
name: manage-task-workflow
description: Plan, execute, and close repository tasks with a disciplined workflow. Use when Claude needs to turn a multi-step request into a tracked task in `.agents/tasks/todo.md`, review `.agents/tasks/lessons.md`, keep progress visible, re-plan when blocked, and validate outcomes before marking work complete.
---

# Manage Task Workflow

Use this skill to run a repeatable execution loop for repository work.
Keep `.agents/tasks/todo.md` as the active session plan and use `.agents/tasks/lessons.md` to avoid repeating mistakes.

Read these references before substantial implementation:

- `references/workflow-checklist.md`

## Core Workflow

1. Review the current request and read `.agents/tasks/lessons.md` for relevant past corrections.
2. Make the active plan explicit in `.agents/tasks/todo.md`.
3. Keep the plan small, checkable, and tied to validation.
4. Execute one meaningful step at a time and update progress as work advances.
5. Stop and re-plan when the current path becomes unclear, risky, or blocked.
6. Verify the result with tests, logs, diffs, or other concrete evidence before closing the task.

## Task Management Rules

- Treat `.agents/tasks/todo.md` as the current task only.
- Archive older task plans instead of mixing multiple sessions into one file.
- Add a short review section to `.agents/tasks/todo.md` before finishing.
- Record new user corrections in `.agents/tasks/lessons.md` so the next session starts smarter.

## Execution Guardrails

- Default to Plan mode thinking for work with 3 or more steps, architecture impact, or non-trivial validation.
- Use the Agent tool with `Explore` subagent for read-only research, or `general-purpose` for complex parallel analysis.
- Prefer root-cause fixes over temporary patches.
- Pause once before major changes and ask whether there is a simpler or more elegant solution.
- Treat bug reports as end-to-end repair tasks: inspect logs, failures, and tests without asking the user to drive the debugging process.

## Output Contract

- Leave `.agents/tasks/todo.md` updated with completed checkboxes and a short review.
- Mention what was verified and what remains unverified.
- If the task exposed a new correction pattern, update `.agents/tasks/lessons.md`.
