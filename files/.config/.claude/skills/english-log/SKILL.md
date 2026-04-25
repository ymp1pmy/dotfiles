---
name: english-log
description: >-
  Review the current session for grammar corrections from EnglishCoach and append them to ~/english_log.md.
  Use after a working session to record what you learned.
license: MIT
---

# English Log

Scan the current conversation for grammar corrections made by the EnglishCoach output style, then append them to `~/english_log.md`.

## Steps

1. Look through the conversation for lines matching this pattern:
   > **[Grammar]:** "{original}" -> "{corrected}" -- {explanation}

2. Read `~/english_log.md` to check if the header already exists. If the file doesn't exist or is empty, create it with this header:

```
| Date | Original | Corrected | Note |
|------|----------|-----------|------|
```

3. For each correction found, append one row:

```
| YYYY-MM-DD | original text | corrected text | explanation |
```

   Use today's date in YYYY-MM-DD format.

4. Tell the user how many corrections were logged (e.g. "Logged 3 corrections to ~/english_log.md.").

If no corrections are found in the session, say so and exit.
