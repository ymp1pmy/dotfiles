---
name: apply-output-formatter
description: Format AI-generated artifacts before finishing work. Use when Claude created or edited files directly or indirectly and needs to apply the repository's formatter rule, defaulting to the VS Code formatter configured for the file type when no other formatter is specified.
---

# Apply Output Formatter

Use this skill after creating or editing deliverables in the repository.
Treat formatting as part of completion, not an optional cleanup step.

Read these references before formatting:

- `references/formatter-checklist.md`

## Core Workflow

1. Identify every file changed by the task, including generated artifacts when they are formatter-compatible text files.
2. Determine whether the user or repository explicitly requested a formatter.
3. If no formatter is specified, use the VS Code default formatter for the file type.
4. Run formatting before final verification so tests, diffs, and reviews reflect the final file contents.
5. Skip formatting only when the file type is not formatter-compatible or formatting would corrupt the artifact.

## Output Contract

- State which files were formatted.
- Mention which formatter path was used when it matters.
- Call out any files intentionally left unformatted and why.
