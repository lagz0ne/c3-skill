#!/usr/bin/env bash
set -euo pipefail

# Cross-compile Go CLI for 4 targets and assemble into skills/c3/bin/
#
# Usage:
#   ./scripts/build.sh           # Build all targets
#   ./scripts/build.sh --version 6.0.0  # Set version string

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CLI_DIR="$ROOT/cli"
BIN_DIR="$ROOT/skills/c3/bin"

# Parse args
VERSION="dev"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --version) VERSION="$2"; shift 2 ;;
    *) echo "Unknown arg: $1" >&2; exit 1 ;;
  esac
done

# Read version from skills VERSION file if not explicitly set
if [ "$VERSION" = "dev" ] && [ -f "$ROOT/skills/c3/bin/VERSION" ]; then
  VERSION=$(cat "$ROOT/skills/c3/bin/VERSION" | tr -d '[:space:]')
fi

echo "Building c3x CLI v${VERSION}"
echo "================================"

TARGETS=(
  "linux:amd64"
  "linux:arm64"
  "darwin:amd64"
  "darwin:arm64"
)

mkdir -p "$BIN_DIR"

for target in "${TARGETS[@]}"; do
  OS="${target%%:*}"
  ARCH="${target##*:}"
  OUTPUT="$BIN_DIR/c3x-${VERSION}-${OS}-${ARCH}"

  echo "  Building ${OS}/${ARCH} -> c3x-${VERSION}-${OS}-${ARCH}"
  CGO_ENABLED=0 GOOS="$OS" GOARCH="$ARCH" go build \
    -C "$CLI_DIR" \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o "$OUTPUT" \
    .
done

# Write VERSION file so c3x.sh knows which binary to use
echo "$VERSION" > "$BIN_DIR/VERSION"

chmod +x "$BIN_DIR"/c3x-*

echo ""
echo "Built binaries:"
ls -lh "$BIN_DIR"/c3x-* 2>/dev/null
echo ""
echo "Done."
