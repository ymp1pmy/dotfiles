#!/bin/bash
# Build the say-response Go binary on first run (or when source is newer than binary).
SRC="${HOME}/.claude/hooks/say-response"
BIN="${SRC}/say-response"

if [ -d "$SRC" ]; then
  if [ ! -f "$BIN" ] || [ "$SRC/main.go" -nt "$BIN" ]; then
    # ビルド失敗時は既存バイナリがあればそれで続行、なければ何もしない
    go build -o "$BIN" "$SRC/" 2>/dev/null || true
  fi
  [ -x "$BIN" ] && exec "$BIN"
fi
exit 0
