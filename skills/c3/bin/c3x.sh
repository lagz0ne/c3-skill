#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

VERSION_FILE="$SCRIPT_DIR/VERSION"
if [ ! -f "$VERSION_FILE" ]; then
  echo "Error: $VERSION_FILE not found — reinstall the plugin" >&2
  exit 1
fi
VERSION=$(tr -d '[:space:]' < "$VERSION_FILE")

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)        ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
esac

BIN="$SCRIPT_DIR/c3x-${VERSION}-${OS}-${ARCH}"

if [ ! -f "$BIN" ]; then
  echo "Error: binary not found: $BIN" >&2
  echo "hint: reinstall the plugin or run: bash scripts/build.sh" >&2
  exit 1
fi

# Remove stale binaries from previous versions.
for old in "$SCRIPT_DIR"/c3x-*; do
  [ "$old" = "$BIN" ] && continue
  rm -f "$old"
done

exec "$BIN" "$@"
