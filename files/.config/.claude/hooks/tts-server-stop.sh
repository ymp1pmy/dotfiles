#!/bin/bash
# Stop the Kokoro TTS server started by tts-server-start.sh.
[ "$CLAUDE_VOICE" != "1" ] && exit 0

PID_FILE="${XDG_RUNTIME_DIR:-$HOME/.cache}/kokoro-tts.pid"
[ -f "$PID_FILE" ] || exit 0

PID=$(cat "$PID_FILE")
# PID ファイルが古い/書き換えられた場合に無関係なプロセスを殺さないよう、
# コマンド名が uvicorn (kokoro) であることを確認してから kill する
if [ -n "$PID" ] && ps -p "$PID" -o command= 2>/dev/null | grep -q 'uvicorn.*kokoro'; then
  kill "$PID" 2>/dev/null
fi
rm -f "$PID_FILE"
