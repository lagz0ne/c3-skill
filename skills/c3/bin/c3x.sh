#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../../.." && pwd)"

VERSION_FILE="$SCRIPT_DIR/VERSION"
if [ ! -f "$VERSION_FILE" ]; then
  echo "Error: $VERSION_FILE not found; reinstall the skill" >&2
  exit 1
fi
VERSION=$(tr -d '[:space:]' < "$VERSION_FILE")

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
esac

case "$OS/$ARCH" in
  linux/amd64|linux/arm64|darwin/amd64|darwin/arm64) ;;
  *)
    echo "Error: unsupported platform: $OS/$ARCH" >&2
    echo "hint: supported platforms are linux/darwin on amd64/arm64" >&2
    exit 1
    ;;
esac

asset_name="c3x-${VERSION}-${OS}-${ARCH}"
bin="$SCRIPT_DIR/$asset_name"

if [ -f "$bin" ]; then
  export C3X_VERSION="$VERSION"
  exec "$bin" "$@"
fi

if [ -f "$ROOT_DIR/cli/go.mod" ] && command -v go >/dev/null 2>&1; then
  go build -C "$ROOT_DIR/cli" \
    -tags embedmodel \
    -buildvcs=false \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o "$bin" \
    .
  chmod +x "$bin"
  export C3X_VERSION="$VERSION"
  exec "$bin" "$@"
fi

export C3X_VERSION="$VERSION"
echo "Error: packaged C3 binary not found: $bin" >&2
echo "hint: reinstall the fat C3 skill artifact, or run from source with Go installed" >&2
exit 1
