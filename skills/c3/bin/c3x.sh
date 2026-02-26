#!/usr/bin/env bash
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
esac
BIN="$SCRIPT_DIR/c3x-${OS}-${ARCH}"
if [ ! -f "$BIN" ]; then
  echo "Error: No c3x binary for ${OS}-${ARCH}" >&2
  exit 1
fi
exec "$BIN" "$@"
