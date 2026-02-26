#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO="lagz0ne/c3-skill"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)        ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
esac

BIN="$SCRIPT_DIR/c3x-${OS}-${ARCH}"

if [ ! -f "$BIN" ]; then
  # Download from latest GitHub release
  TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
  if [ -z "$TAG" ]; then
    echo "Error: could not determine latest release from ${REPO}" >&2
    exit 1
  fi
  ASSET="c3-skill-${TAG}.zip"
  URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"
  TMP=$(mktemp -d)
  echo "Downloading c3x ${TAG} for ${OS}-${ARCH}..." >&2
  curl -fsSL -o "${TMP}/${ASSET}" "$URL"
  unzip -qo "${TMP}/${ASSET}" "skills/c3/bin/c3x-${OS}-${ARCH}" -d "$TMP"
  mv "${TMP}/skills/c3/bin/c3x-${OS}-${ARCH}" "$BIN"
  chmod +x "$BIN"
  rm -rf "$TMP"
  echo "Installed c3x ${TAG}" >&2
fi

exec "$BIN" "$@"
