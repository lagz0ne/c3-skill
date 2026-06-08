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
    echo "hint: use a fat C3 build for offline or unsupported environments" >&2
    exit 1
    ;;
esac

VARIANT="${C3X_VARIANT:-thin}"
if [ "${C3X_USE_FAT:-}" != "" ]; then
  VARIANT="fat"
fi

asset_name="c3x-${VERSION}-${OS}-${ARCH}"
fat_name="${asset_name}-fat"

if [ "$VARIANT" = "fat" ]; then
  fat_bin="$SCRIPT_DIR/$fat_name"
  if [ ! -f "$fat_bin" ]; then
    echo "Error: fat binary not found: $fat_bin" >&2
    echo "hint: install the fat C3 skill artifact or unset C3X_VARIANT/C3X_USE_FAT for thin mode" >&2
    exit 1
  fi
  export C3X_VERSION="$VERSION"
  exec "$fat_bin" "$@"
fi

if [ "${C3X_USE_LOCAL_BIN:-}" != "" ] && [ -f "$SCRIPT_DIR/$asset_name" ]; then
  export C3X_VERSION="$VERSION"
  exec "$SCRIPT_DIR/$asset_name" "$@"
fi

if [ "${C3X_CACHE_DIR:-}" != "" ]; then
  cache_root="$C3X_CACHE_DIR"
elif [ "${XDG_CACHE_HOME:-}" != "" ]; then
  cache_root="$XDG_CACHE_HOME/c3x"
else
  cache_root="$HOME/.cache/c3x"
fi
cache_dir="$cache_root/$VERSION"
bin="$cache_dir/$asset_name"

mkdir -p "$cache_dir"
if [ -d "$cache_root" ]; then
  find "$cache_root" -mindepth 1 -maxdepth 1 -type d ! -name "$VERSION" -exec rm -rf {} +
fi

if [ ! -f "$bin" ] && [ -f "$ROOT_DIR/cli/go.mod" ] && command -v go >/dev/null 2>&1; then
  go build -C "$ROOT_DIR/cli" -buildvcs=false -ldflags="-s -w -X main.version=${VERSION}" -o "$bin" .
  chmod +x "$bin"
fi

if [ ! -f "$bin" ]; then
  release_base="${C3X_RELEASE_BASE_URL:-https://github.com/lagz0ne/c3-design/releases/download/v${VERSION}}"
  tmp="$bin.tmp"
  checksum_tmp="$bin.sha256.tmp"

  sha256_file() {
    if command -v sha256sum >/dev/null 2>&1; then
      sha256sum "$1" | awk '{print $1}'
    else
      shasum -a 256 "$1" | awk '{print $1}'
    fi
  }

  fetch() {
    url="$1"
    out="$2"
    if command -v curl >/dev/null 2>&1; then
      curl -fsSL "$url" -o "$out"
    elif command -v wget >/dev/null 2>&1; then
      wget -q "$url" -O "$out"
    else
      echo "Error: need curl or wget to download C3 thin binary" >&2
      echo "hint: install curl/wget, prefill $cache_dir, or use the fat C3 skill artifact" >&2
      exit 1
    fi
  }

  if ! fetch "$release_base/$asset_name.sha256" "$checksum_tmp"; then
    echo "Error: failed to download checksum for $asset_name" >&2
    echo "hint: connect once to GitHub Releases or use the fat C3 skill artifact for offline installs" >&2
    exit 1
  fi
  if ! fetch "$release_base/$asset_name" "$tmp"; then
    rm -f "$tmp" "$checksum_tmp"
    echo "Error: failed to download $asset_name" >&2
    echo "hint: connect once to GitHub Releases or use the fat C3 skill artifact for offline installs" >&2
    exit 1
  fi

  expected="$(awk '{print $1; exit}' "$checksum_tmp")"
  actual="$(sha256_file "$tmp")"
  rm -f "$checksum_tmp"
  if [ "$expected" != "$actual" ]; then
    rm -f "$tmp"
    echo "Error: checksum mismatch for $asset_name" >&2
    echo "hint: clear $cache_dir and retry, or use the fat C3 skill artifact" >&2
    exit 1
  fi
  chmod +x "$tmp"
  mv "$tmp" "$bin"
fi

export C3X_VERSION="$VERSION"
export C3_SEMANTIC_CACHE_DIR="${C3_SEMANTIC_CACHE_DIR:-$cache_dir/semantic}"
exec "$bin" "$@"
