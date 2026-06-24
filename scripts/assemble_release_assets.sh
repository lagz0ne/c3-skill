#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

VERSION=""
ARTIFACTS_DIR=""
OUT_DIR=""

usage() {
  cat >&2 <<'EOF'
usage: assemble_release_assets.sh --version VERSION --artifacts-dir DIR --out-dir DIR
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      VERSION="$2"
      shift 2
      ;;
    --artifacts-dir)
      ARTIFACTS_DIR="$2"
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

if [ -z "$VERSION" ] || [ -z "$ARTIFACTS_DIR" ] || [ -z "$OUT_DIR" ]; then
  usage
  exit 1
fi

ARTIFACTS_DIR="$(cd "$ARTIFACTS_DIR" && pwd)"
mkdir -p "$OUT_DIR"
OUT_DIR="$(cd "$OUT_DIR" && pwd)"

for required in .gitattributes .claude-plugin skills; do
  if [ ! -e "$ROOT/$required" ]; then
    echo "Missing release input: $required" >&2
    exit 1
  fi
done

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 is required to assemble skill archives" >&2
  exit 1
fi

tmp_root="$(mktemp -d)"
cleanup() {
  rm -rf "$tmp_root"
}
trap cleanup EXIT

copy_skill_tree() {
  local work="$1"
  rm -rf "$work"
  mkdir -p "$work"
  cp "$ROOT/.gitattributes" "$work/"
  cp -R "$ROOT/.claude-plugin" "$ROOT/skills" "$work/"
  rm -f "$work/skills/c3/bin/c3x-"*
  rm -f "$work/skills/c3/bin/ast-grep-"*
}

zip_skill_tree() {
  local work="$1"
  local archive="$2"
  python3 - "$work" "$archive" <<'PY'
import os
import stat
import sys
import zipfile
from pathlib import Path

work = Path(sys.argv[1])
archive = Path(sys.argv[2])
roots = [work / ".gitattributes", work / ".claude-plugin", work / "skills"]

with zipfile.ZipFile(archive, "w", compression=zipfile.ZIP_DEFLATED) as zf:
    for root in roots:
        paths = [root] if root.is_file() else sorted(path for path in root.rglob("*") if path.is_file())
        for path in paths:
            if path.name == ".DS_Store":
                continue
            rel = path.relative_to(work).as_posix()
            info = zipfile.ZipInfo(rel)
            mode = path.stat().st_mode
            info.compress_type = zipfile.ZIP_DEFLATED
            info.external_attr = (stat.S_IMODE(mode) & 0o777) << 16
            with path.open("rb") as handle:
                zf.writestr(info, handle.read())
PY
}

add_skill_binary_archive() {
  local binary="$1"
  local platform="$2"
  local work="$tmp_root/c3-skill-$platform"
  local ast_platform="${platform%-portable}"
  local ast_grep

  copy_skill_tree "$work"
  cp "$binary" "$work/skills/c3/bin/c3x-${VERSION}-${platform}"
  chmod +x "$work/skills/c3/bin/c3x-${VERSION}-${platform}"
  ast_grep="$(find_ast_grep_binary "$ast_platform")"
  cp "$ast_grep" "$work/skills/c3/bin/$(basename "$ast_grep")"
  chmod +x "$work/skills/c3/bin/$(basename "$ast_grep")"
  zip_skill_tree "$work" "$OUT_DIR/c3-skill-${platform}-v${VERSION}.zip"
}

find_ast_grep_binary() {
  local platform="$1"
  local found
  found="$(find "$ARTIFACTS_DIR" -type f -name "ast-grep-*-${platform}" ! -name '*.sha256' | sort | head -n1 || true)"
  if [ -z "$found" ]; then
    echo "Missing ast-grep binary for $platform" >&2
    exit 1
  fi
  printf '%s\n' "$found"
}

find "$ARTIFACTS_DIR" -type f -path '*/thin/c3x-*' -exec cp {} "$OUT_DIR"/ \;
find "$ARTIFACTS_DIR" -type f -name 'ast-grep-*' -exec cp {} "$OUT_DIR"/ \;
if [ -d "$ARTIFACTS_DIR/semantic-assets" ]; then
  find "$ARTIFACTS_DIR/semantic-assets" -type f -exec cp {} "$OUT_DIR"/ \;
fi

no_binary_work="$tmp_root/c3-skill-no-binary"
copy_skill_tree "$no_binary_work"
zip_skill_tree "$no_binary_work" "$OUT_DIR/c3-skill-v${VERSION}.zip"

linux_fat_platforms="$tmp_root/linux-fat-platforms"
portable_platforms="$tmp_root/portable-platforms"
: > "$linux_fat_platforms"
: > "$portable_platforms"

fat_count=0
while IFS= read -r fat; do
  fat_count=$((fat_count + 1))
  platform="$(basename "$fat" | sed -E "s/^c3x-${VERSION}-//; s/-fat$//")"
  if [[ "$platform" == linux-* ]]; then
    echo "${platform}-portable" >> "$linux_fat_platforms"
  fi
  add_skill_binary_archive "$fat" "$platform"
done < <(find "$ARTIFACTS_DIR" -type f -path '*/fat/c3x-*-fat' ! -name '*.sha256' | sort)

if [ "$fat_count" -eq 0 ]; then
  echo "No fat skill binaries found" >&2
  exit 1
fi

portable_count=0
while IFS= read -r portable; do
  portable_count=$((portable_count + 1))
  platform="$(basename "$portable" | sed -E "s/^c3x-${VERSION}-//")"
  echo "$platform" >> "$portable_platforms"
  add_skill_binary_archive "$portable" "$platform"
done < <(find "$ARTIFACTS_DIR" -type f -path '*/portable/c3x-*-portable' ! -name '*.sha256' | sort)

missing_portable=0
while IFS= read -r expected; do
  if ! grep -Fxq "$expected" "$portable_platforms"; then
    echo "Missing portable skill binary for $expected" >&2
    missing_portable=1
  fi
done < "$linux_fat_platforms"

if [ "$missing_portable" -ne 0 ]; then
  exit 1
fi

rm -f "$OUT_DIR/SHA256SUMS"
(cd "$OUT_DIR" && sha256sum * > SHA256SUMS)
ls -la "$OUT_DIR"
