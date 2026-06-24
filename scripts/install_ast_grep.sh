#!/usr/bin/env bash
set -euo pipefail

VERSION=""
TARGET_OS=""
TARGET_ARCH=""
OUT_DIR=""

usage() {
  cat >&2 <<'EOF'
usage: install_ast_grep.sh --version VERSION --os OS --arch ARCH --out-dir DIR
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      VERSION="$2"
      shift 2
      ;;
    --os)
      TARGET_OS="$2"
      shift 2
      ;;
    --arch)
      TARGET_ARCH="$2"
      shift 2
      ;;
    --out-dir)
      OUT_DIR="$2"
      shift 2
      ;;
    *)
      echo "Unknown arg: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [ -z "$VERSION" ] || [ -z "$TARGET_OS" ] || [ -z "$TARGET_ARCH" ] || [ -z "$OUT_DIR" ]; then
  usage
  exit 1
fi

case "$TARGET_OS/$TARGET_ARCH" in
  linux/amd64) package="@ast-grep/cli-linux-x64-gnu" ;;
  linux/arm64) package="@ast-grep/cli-linux-arm64-gnu" ;;
  darwin/arm64) package="@ast-grep/cli-darwin-arm64" ;;
  *)
    echo "unsupported ast-grep target: $TARGET_OS/$TARGET_ARCH" >&2
    exit 1
    ;;
esac

if ! command -v npm >/dev/null 2>&1; then
  echo "npm is required to fetch pinned ast-grep binary package" >&2
  exit 1
fi

mkdir -p "$OUT_DIR"
OUT_DIR="$(cd "$OUT_DIR" && pwd)"
output="$OUT_DIR/ast-grep-${VERSION}-${TARGET_OS}-${TARGET_ARCH}"

tmp="$(mktemp -d)"
cleanup() {
  rm -rf "$tmp"
}
trap cleanup EXIT

echo "Fetching ast-grep v${VERSION} for ${TARGET_OS}/${TARGET_ARCH}"
npm pack "${package}@${VERSION}" --pack-destination "$tmp" >/dev/null
tar -xzf "$tmp"/*.tgz -C "$tmp"
cp "$tmp/package/ast-grep" "$output"
chmod +x "$output"

if command -v sha256sum >/dev/null 2>&1; then
  sha256sum "$output" > "$output.sha256"
else
  shasum -a 256 "$output" > "$output.sha256"
fi
