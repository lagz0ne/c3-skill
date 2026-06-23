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
  linux/amd64|linux/arm64|darwin/arm64) ;;
  *)
    echo "Error: unsupported platform: $OS/$ARCH" >&2
    echo "hint: supported platforms are linux amd64/arm64 and darwin arm64" >&2
    exit 1
    ;;
esac

asset_name="c3x-${VERSION}-${OS}-${ARCH}"
bin="$SCRIPT_DIR/$asset_name"
portable_bin="$SCRIPT_DIR/${asset_name}-portable"

print_wrapper_help() {
  cat <<EOF
Usage: c3x <command> [options]

Commands:
  versions           List available and installed C3 runtime versions
  install            Install a C3 runtime into the shared cache
  uninstall          Remove an installed C3 runtime from the shared cache
  cache              Inspect or prune the shared C3 cache
  check              Check the current C3 project documents
  eval               Evaluate the current C3 project documents

This no-binary wrapper runs bundled binaries when present. Without a bundled
binary, real commands delegate to @c3x/cli@${VERSION}, which resolves the
project runtime version or latest release before downloading runtime assets.
EOF
}

if [ -f "$bin" ]; then
  export C3X_VERSION="$VERSION"
  exec "$bin" "$@"
fi

if [ "$OS" = "linux" ] && [ -f "$portable_bin" ]; then
  export C3X_VERSION="$VERSION"
  exec "$portable_bin" "$@"
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

case "${1-}" in
  ""|-h|--help|help)
    print_wrapper_help
    exit 0
    ;;
  -V|--version|version)
    printf 'c3x %s\n' "$VERSION"
    exit 0
    ;;
esac

if command -v npm >/dev/null 2>&1; then
  exec npm exec --yes --package "@c3x/cli@${VERSION}" -- c3x "$@"
fi

echo "Error: packaged C3 binary not found: $bin" >&2
echo "hint: install npm so the no-binary skill can use @c3x/cli@${VERSION}, reinstall a fat/portable C3 skill artifact, or run from source with Go installed" >&2
exit 1
