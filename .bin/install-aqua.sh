#!/bin/bash
set -ue

# OS detection for sha256sum command
if [[ "$OSTYPE" == "darwin"* ]]; then
    SHA256_CMD="shasum -a 256"
else
    SHA256_CMD="sha256sum"
fi

# SEE: https://aquaproj.github.io/docs/products/aqua-installer#shell-script
if [[ ! -x "$HOME/.local/share/aquaproj-aqua/bin/aqua" ]]; then
  # カレントディレクトリ（実行場所依存）ではなく専用の一時ディレクトリに落とす
  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT
  curl -sSfL -o "$tmpdir/aqua-installer" https://raw.githubusercontent.com/aquaproj/aqua-installer/v3.1.1/aqua-installer
  echo "e9d4c99577c6b2ce0b62edf61f089e9b9891af1708e88c6592907d2de66e3714  $tmpdir/aqua-installer" | $SHA256_CMD -c -
  chmod +x "$tmpdir/aqua-installer"
  "$tmpdir/aqua-installer" -v v2.48.1
fi

# AQUA_CONFIG points to the aqua.yaml file (aqua's own env var for proxy lookup).
# May be pre-set by install script when ~/.config/aquaproj-aqua couldn't be symlinked.
AQUA_CONFIG=${AQUA_CONFIG:-$HOME/.config/aquaproj-aqua/aqua.yaml}
export AQUA_CONFIG
AQUA_POLICY_CONFIG="$(dirname "$AQUA_CONFIG")/policy.yaml"
export AQUA_POLICY_CONFIG
# Pass policy file explicitly — env-var-only form silently ignores it in some envs
$HOME/.local/share/aquaproj-aqua/bin/aqua policy allow "$AQUA_POLICY_CONFIG"
$HOME/.local/share/aquaproj-aqua/bin/aqua -c "$AQUA_CONFIG" i
