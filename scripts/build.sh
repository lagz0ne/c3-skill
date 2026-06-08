#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CLI_DIR="$ROOT/cli"
DEFAULT_OUT_DIR="$ROOT/dist/c3x"

VERSION="dev"
VARIANT="fat"
OUT_DIR="$DEFAULT_OUT_DIR"
TARGET_OS=""
TARGET_ARCH=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version) VERSION="$2"; shift 2 ;;
    --variant) VARIANT="$2"; shift 2 ;;
    --out-dir) OUT_DIR="$2"; shift 2 ;;
    --os) TARGET_OS="$2"; shift 2 ;;
    --arch) TARGET_ARCH="$2"; shift 2 ;;
    *) echo "Unknown arg: $1" >&2; exit 1 ;;
  esac
done

if [ "$VERSION" = "dev" ] && [ -f "$ROOT/skills/c3/bin/VERSION" ]; then
  VERSION=$(tr -d '[:space:]' < "$ROOT/skills/c3/bin/VERSION")
fi

if [ -z "$TARGET_OS" ]; then
  TARGET_OS=$(go env GOOS)
fi
if [ -z "$TARGET_ARCH" ]; then
  TARGET_ARCH=$(go env GOARCH)
fi

build_variant() {
  local variant="$1"
  local tags=""
  local suffix=""
  if [ "$variant" = "fat" ]; then
    tags="-tags embedmodel"
    suffix="-fat"
  elif [ "$variant" != "thin" ]; then
    echo "Unknown variant: $variant" >&2
    exit 1
  fi

  local output_dir="$OUT_DIR/$variant"
  local output="$output_dir/c3x-${VERSION}-${TARGET_OS}-${TARGET_ARCH}${suffix}"
  mkdir -p "$output_dir"
  echo "Building $variant c3x v${VERSION} for ${TARGET_OS}/${TARGET_ARCH}"
  GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" go build \
    -C "$CLI_DIR" \
    $tags \
    -buildvcs=false \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o "$output" \
    .
  chmod +x "$output"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$output" > "$output.sha256"
  else
    shasum -a 256 "$output" > "$output.sha256"
  fi
}

case "$VARIANT" in
  thin) build_variant thin ;;
  fat) build_variant fat ;;
  both)
    build_variant thin
    build_variant fat
    ;;
  *) echo "Unknown variant: $VARIANT" >&2; exit 1 ;;
esac

echo "Built artifacts under $OUT_DIR"
