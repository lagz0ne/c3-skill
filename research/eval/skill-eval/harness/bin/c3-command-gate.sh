#!/usr/bin/env bash
set -u

: "${C3_USAGE_LOG:?C3_USAGE_LOG is required}"

tool_name="${0##*/}"
real_tool="${C3_GATE_REAL_TOOL:-/usr/bin/$tool_name}"

# Direct skill triggers may read the mounted C3 instructions before routing.
# Canonicalize every path so traversal cannot escape the mounted skill roots.
case "$tool_name" in
  cat|sed|head|tail)
    skill_read=0
    outside_path=0
    skill_roots="${C3_SKILL_READ_ROOTS:-/opt/c3/skills/c3:/work/project/.agents/skills/c3:/work/project/.claude/skills/c3}"
    for arg in "$@"; do
      [[ "$arg" == -* || "$arg" != */* ]] && continue
      canonical="$(/usr/bin/realpath -m -- "$arg")"
      matched=0
      IFS=: read -ra roots <<< "$skill_roots"
      for root in "${roots[@]}"; do
        canonical_root="$(/usr/bin/realpath -m -- "$root")"
        if [[ "$canonical" == "$canonical_root" || "$canonical" == "$canonical_root/"* ]]; then
          matched=1
          skill_read=1
          break
        fi
      done
      [[ "$matched" -eq 1 ]] || outside_path=1
    done
    if [[ "$skill_read" -eq 1 && "$outside_path" -eq 0 ]]; then
      exec "$real_tool" "$@"
    fi
    ;;
esac

route_ok=0
impact_ok=0
evidence_ok=0
if [[ -f "$C3_USAGE_LOG" ]]; then
  while IFS=$'\t' read -r category status _; do
    [[ "$status" == "0" ]] || continue
    case "$category" in
      route) route_ok=1 ;;
      impact) impact_ok=1 ;;
      evidence) evidence_ok=1 ;;
    esac
  done < "$C3_USAGE_LOG"
fi

if [[ "$route_ok" -ne 1 ]]; then
  printf '%s\n' \
    'C3 treatment gate: repository inspection is locked; required next: route' \
    'Run search "<behavior or domain>"; if it returns no match, run list --type component and use only a returned id.' >&2
  exit 78
fi

if [[ "$impact_ok" -ne 1 ]]; then
  printf '%s\n' \
    'C3 treatment gate: repository inspection is locked; required next: impact' \
    'Run: C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh graph <matched-id> --depth 1' >&2
  exit 78
fi

if [[ "$evidence_ok" -ne 1 ]]; then
  printf '%s\n' \
    'C3 treatment gate: repository inspection is locked; required next: evidence' \
    'Run: C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh read <relevant-id> --section <section>' >&2
  exit 78
fi

exec "$real_tool" "$@"
