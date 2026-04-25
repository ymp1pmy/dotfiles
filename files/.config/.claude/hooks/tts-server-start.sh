#!/bin/bash
# Start Kokoro TTS server when CLAUDE_VOICE=1 is set.
# Install: pip install kokoro-fastapi uvicorn
[ "$CLAUDE_VOICE" != "1" ] && exit 0

# Skip if already running
curl -sf --max-time 1 http://localhost:8880/health > /dev/null 2>&1 && exit 0

PID_FILE="/tmp/kokoro-tts.pid"

if python3 -c "import kokoro_fastapi" 2>/dev/null; then
  nohup python3 -m uvicorn kokoro_fastapi.main:app \
    --host 127.0.0.1 --port 8880 \
    > /tmp/kokoro-tts.log 2>&1 &
  echo $! > "$PID_FILE"
  # Wait up to 15s for the server to become ready
  for _ in $(seq 1 15); do
    sleep 1
    curl -sf --max-time 1 http://localhost:8880/health > /dev/null 2>&1 && break
  done
fi
