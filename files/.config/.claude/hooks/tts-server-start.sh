#!/bin/bash
# Start Kokoro TTS server when CLAUDE_VOICE=1 is set.
# Install: pip install kokoro-fastapi uvicorn
[ "$CLAUDE_VOICE" != "1" ] && exit 0

# Skip if already running
curl -sf --max-time 1 http://localhost:8880/health > /dev/null 2>&1 && exit 0

# /tmp は全ユーザー書き込み可能でシンボリックリンク攻撃の余地があるため、
# 自ユーザー専用の XDG_RUNTIME_DIR (なければ ~/.cache) を使う
RUN_DIR="${XDG_RUNTIME_DIR:-$HOME/.cache}"
mkdir -p "$RUN_DIR"
PID_FILE="$RUN_DIR/kokoro-tts.pid"
LOG_FILE="$RUN_DIR/kokoro-tts.log"

if python3 -c "import kokoro_fastapi" 2>/dev/null; then
  nohup python3 -m uvicorn kokoro_fastapi.main:app \
    --host 127.0.0.1 --port 8880 \
    > "$LOG_FILE" 2>&1 &
  echo $! > "$PID_FILE"
  # Wait up to 15s for the server to become ready
  for _ in $(seq 1 15); do
    sleep 1
    curl -sf --max-time 1 http://localhost:8880/health > /dev/null 2>&1 && break
  done
fi
