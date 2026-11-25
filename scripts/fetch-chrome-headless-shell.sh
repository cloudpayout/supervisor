#!/usr/bin/env bash
# Provision chrome-headless-shell for spec 017 (pin 157.0.8103.78, linux64).
# The binary must stay next to its ICU/data files from the Chrome for Testing zip.
set -euo pipefail
VERSION="${CHROME_HEADLESS_SHELL_VERSION:-157.0.8103.78}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
STAGE="$ROOT/bin/chrome-headless-shell-linux64"
LINK="$ROOT/bin/chrome-headless-shell"
URL="https://storage.googleapis.com/chrome-for-testing-public/${VERSION}/linux64/chrome-headless-shell-linux64.zip"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
echo "fetching chrome-headless-shell ${VERSION}"
curl -fsSL "$URL" -o "$TMP/shell.zip"
unzip -q "$TMP/shell.zip" -d "$TMP"
SRC="$(find "$TMP" -type f -name chrome-headless-shell | head -1)"
if [[ -z "$SRC" ]]; then
  echo "chrome-headless-shell binary not found in archive" >&2
  exit 1
fi
rm -rf "$STAGE"
mkdir -p "$ROOT/bin"
cp -a "$(dirname "$SRC")" "$STAGE"
chmod 755 "$STAGE/chrome-headless-shell"
ln -sfn chrome-headless-shell-linux64/chrome-headless-shell "$LINK"
echo "installed $LINK -> $STAGE/chrome-headless-shell"
"$STAGE/chrome-headless-shell" --version || true
