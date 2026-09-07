#!/bin/bash
# Claude Code hook: 端末にベル (BEL) を送って手が空いたことを知らせる
# Stop / Notification で発火させる想定。
# WezTerm 側は audible_bell / visual_bell / bell イベントで拾う
# (~/.config/wezterm/.wezterm.lua)。

# フックの stdout / stderr は Claude Code が JSON として取り込んでしまうので、
# BEL は制御端末へ直接書き込む。
# 制御端末を持たない実行 (claude -p のパイプ実行, CI など) では open に失敗するが、
# 2> の指定を先に置くことでそのエラーも握り潰し、黙って何もしない。
printf '\a' 2> /dev/null > /dev/tty

exit 0
