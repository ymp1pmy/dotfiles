#!/bin/bash
# Stop the Kokoro TTS server started by tts-server-start.sh.
[ "$CLAUDE_VOICE" != "1" ] && exit 0

PID_FILE="/tmp/kokoro-tts.pid"
if [ -f "$PID_FILE" ]; then
  kill "$(cat "$PID_FILE")" 2>/dev/null && rm -f "$PID_FILE"
fi
