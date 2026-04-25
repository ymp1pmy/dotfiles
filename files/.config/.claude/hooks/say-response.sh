#!/bin/bash
# Build the say-response Go binary on first run (or when source is newer than binary).
SRC="${HOME}/.claude/hooks/say-response"
BIN="${SRC}/say-response"

if [ -d "$SRC" ]; then
  if [ ! -f "$BIN" ] || [ "$SRC/main.go" -nt "$BIN" ]; then
    go build -o "$BIN" "$SRC/" 2>/dev/null || exit 0
  fi
  exec "$BIN"
fi
