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
AST_GREP_VERSION_FILE="$SCRIPT_DIR/AST_GREP_VERSION"
AST_GREP_VERSION=""
if [ -f "$AST_GREP_VERSION_FILE" ]; then
  AST_GREP_VERSION=$(tr -d '[:space:]' < "$AST_GREP_VERSION_FILE")
fi

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
ast_grep_bin=""
if [ -n "$AST_GREP_VERSION" ]; then
  candidate="$SCRIPT_DIR/ast-grep-${AST_GREP_VERSION}-${OS}-${ARCH}"
  if [ -f "$candidate" ]; then
    ast_grep_bin="$candidate"
  fi
fi

if [ -z "${C3_AST_GREP:-}" ] && [ -n "$ast_grep_bin" ]; then
  export C3_AST_GREP="$ast_grep_bin"
fi

maybe_install_ast_grep() {
  if [ "${1-}" != "eval" ]; then
    return 0
  fi
  if [ -n "${C3_AST_GREP:-}" ] || [ -z "$AST_GREP_VERSION" ]; then
    return 0
  fi
  if [ ! -x "$ROOT_DIR/scripts/install_ast_grep.sh" ] || ! command -v npm >/dev/null 2>&1; then
    return 0
  fi
  local candidate="$SCRIPT_DIR/ast-grep-${AST_GREP_VERSION}-${OS}-${ARCH}"
  if [ -f "$candidate" ]; then
    export C3_AST_GREP="$candidate"
    return 0
  fi
  if bash "$ROOT_DIR/scripts/install_ast_grep.sh" \
    --version "$AST_GREP_VERSION" \
    --os "$OS" \
    --arch "$ARCH" \
    --out-dir "$SCRIPT_DIR" >/dev/null 2>&1 && [ -f "$candidate" ]; then
    export C3_AST_GREP="$candidate"
  fi
}

# Development and migration tooling can build the current checkout into a private
# binary, then still route every operation through this repository-local wrapper.
# The override is explicit and session-scoped; installed/global C3 remains unused.
if [ -n "${C3X_LOCAL_BINARY:-}" ]; then
  if [ ! -x "$C3X_LOCAL_BINARY" ]; then
    echo "Error: C3X_LOCAL_BINARY is not executable: $C3X_LOCAL_BINARY" >&2
    exit 1
  fi
  export C3X_VERSION="${C3X_LOCAL_VERSION:-$VERSION}"
  maybe_install_ast_grep "${1-}"
  exec "$C3X_LOCAL_BINARY" "$@"
fi

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
  maybe_install_ast_grep "${1-}"
  exec "$bin" "$@"
fi

if [ "$OS" = "linux" ] && [ -f "$portable_bin" ]; then
  export C3X_VERSION="$VERSION"
  maybe_install_ast_grep "${1-}"
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
  maybe_install_ast_grep "${1-}"
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
