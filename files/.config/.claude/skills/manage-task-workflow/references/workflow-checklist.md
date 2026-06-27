# Workflow Checklist

## Start of Task

1. Read the user request and identify the expected deliverable.
2. Review `tasks/lessons.md` for related mistakes, constraints, or operating rules.
3. Decide whether the task needs plan-first handling:
   - 3 or more implementation steps
   - architecture or workflow design impact
   - non-trivial validation or rollback risk
4. Open or rewrite `tasks/todo.md` as the active plan for the current session.
5. Define the validation method before editing files.

## During Execution

1. Keep plan items concrete and checkable.
2. Update progress as soon as a step is materially complete.
3. Prefer one focused task per subagent or parallel investigation.
4. Re-plan immediately when:
   - a dependency or assumption fails
   - the task shape changes
   - the current fix feels like a hack
5. Minimize change surface and preserve unrelated work in the repository.

## Before Completion

1. Run the relevant tests, commands, or manual verification steps.
2. Review diffs and confirm the result would pass a staff-level review.
3. Add a short review summary to `tasks/todo.md`:
   - work done
   - verification results
   - remaining issues
4. Mark the task complete only after evidence exists.

## Lessons Loop

Update `tasks/lessons.md` when the user corrects the approach or when a new failure pattern should become a standing rule.
Write each lesson as:

- symptom
- cause
- fix
- recurrence-prevention rule
