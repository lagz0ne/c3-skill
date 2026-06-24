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
SEMANTIC_MODEL_BACKUP=""
AST_GREP_VERSION=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version) VERSION="$2"; shift 2 ;;
    --ast-grep-version) AST_GREP_VERSION="$2"; shift 2 ;;
    --variant) VARIANT="$2"; shift 2 ;;
    --out-dir) OUT_DIR="$2"; shift 2 ;;
    --os) TARGET_OS="$2"; shift 2 ;;
    --arch) TARGET_ARCH="$2"; shift 2 ;;
    *) echo "Unknown arg: $1" >&2; exit 1 ;;
  esac
done

# Resolve OUT_DIR to an absolute path. `go build -C "$CLI_DIR" -o "$output"`
# treats a relative -o as relative to CLI_DIR, so a relative --out-dir (as CI
# passes, e.g. dist/c3x) would land the binary under cli/ and the later chmod
# at repo-root would fail. Absolutize so relative and absolute inputs both work.
mkdir -p "$OUT_DIR"
OUT_DIR="$(cd "$OUT_DIR" && pwd)"

if [ "$VERSION" = "dev" ] && [ -f "$ROOT/skills/c3/bin/VERSION" ]; then
  VERSION=$(tr -d '[:space:]' < "$ROOT/skills/c3/bin/VERSION")
fi

if [ -z "$AST_GREP_VERSION" ] && [ -f "$ROOT/skills/c3/bin/AST_GREP_VERSION" ]; then
  AST_GREP_VERSION=$(tr -d '[:space:]' < "$ROOT/skills/c3/bin/AST_GREP_VERSION")
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
  local build_env=(GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH")
  if [ "$variant" = "fat" ]; then
    tags="-tags embedmodel"
    suffix="-fat"
  elif [ "$variant" = "portable" ]; then
    if [ "$TARGET_OS" != "linux" ]; then
      echo "Portable builds are supported only for linux targets, got: $TARGET_OS/$TARGET_ARCH" >&2
      exit 1
    fi
    tags="-tags netgo,osusergo"
    suffix="-portable"
    build_env+=(CGO_ENABLED=0)
  elif [ "$variant" != "thin" ]; then
    echo "Unknown variant: $variant" >&2
    exit 1
  fi

  local output_dir="$OUT_DIR/$variant"
  local output="$output_dir/c3x-${VERSION}-${TARGET_OS}-${TARGET_ARCH}${suffix}"
  mkdir -p "$output_dir"
  echo "Building $variant c3x v${VERSION} for ${TARGET_OS}/${TARGET_ARCH}"

  if [ "$variant" = "fat" ]; then
    backup_semantic_model_stubs
    if ! prepare_fat_semantic_model; then
      restore_semantic_model_stubs
      exit 1
    fi
  fi

  local build_status=0
  set +e
  env "${build_env[@]}" go build \
    -C "$CLI_DIR" \
    $tags \
    -buildvcs=false \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o "$output" \
    .
  build_status=$?
  set -e

  if [ "$variant" = "fat" ]; then
    restore_semantic_model_stubs
  fi
  if [ "$build_status" -ne 0 ]; then
    exit "$build_status"
  fi

  chmod +x "$output"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$output" > "$output.sha256"
  else
    shasum -a 256 "$output" > "$output.sha256"
  fi
  install_ast_grep "$output_dir"
}

install_ast_grep() {
  local output_dir="$1"
  if [ -z "$AST_GREP_VERSION" ]; then
    echo "AST_GREP_VERSION is empty; set --ast-grep-version or skills/c3/bin/AST_GREP_VERSION" >&2
    exit 1
  fi
  bash "$ROOT/scripts/install_ast_grep.sh" \
    --version "$AST_GREP_VERSION" \
    --os "$TARGET_OS" \
    --arch "$TARGET_ARCH" \
    --out-dir "$output_dir"
}

backup_semantic_model_stubs() {
  SEMANTIC_MODEL_BACKUP="$(mktemp -d)"
  cp -R "$CLI_DIR/internal/store/semantic_model/." "$SEMANTIC_MODEL_BACKUP/"
}

prepare_fat_semantic_model() {
  echo "Preparing embedded semantic model assets"
  C3X_VERSION="$VERSION" go \
    -C "$CLI_DIR" \
    run ./tools/semantic-assets \
    --embed-dir "$CLI_DIR/internal/store/semantic_model" \
    --os "$TARGET_OS" \
    --arch "$TARGET_ARCH"
}

restore_semantic_model_stubs() {
  if [ -n "$SEMANTIC_MODEL_BACKUP" ] && [ -d "$SEMANTIC_MODEL_BACKUP" ]; then
    cp -R "$SEMANTIC_MODEL_BACKUP/." "$CLI_DIR/internal/store/semantic_model/"
    rm -rf "$SEMANTIC_MODEL_BACKUP"
    SEMANTIC_MODEL_BACKUP=""
  fi
  git -C "$ROOT" checkout -- cli/internal/store/semantic_model/
}

case "$VARIANT" in
  thin) build_variant thin ;;
  fat) build_variant fat ;;
  portable) build_variant portable ;;
  both)
    build_variant thin
    build_variant fat
    ;;
  release)
    build_variant thin
    build_variant fat
    if [ "$TARGET_OS" = "linux" ]; then
      build_variant portable
    fi
    ;;
  *) echo "Unknown variant: $VARIANT" >&2; exit 1 ;;
esac

echo "Built artifacts under $OUT_DIR"
