#!/usr/bin/env bash
set -euo pipefail

wrapper="/opt/c3/skills/c3/bin/c3x.sh"
skill="/opt/c3/skills/c3/SKILL.md"
sweep="/opt/c3/skills/c3/references/sweep.md"
if [[ "$#" -eq 0 ]]; then
  echo "usage: c3-impact-bootstrap <behavior or domain>" >&2
  exit 2
fi

route_output="$(mktemp)"
forward_output="$(mktemp)"
reverse_output="$(mktemp)"
evidence_output="$(mktemp)"
trap 'rm -f "$route_output" "$forward_output" "$reverse_output" "$evidence_output"' EXIT

# The treatment instructions already carry the frozen procedure. Verify that
# their local sources are mounted, but do not duplicate their full text into
# the tool result before returning repository evidence.
if [[ ! -s "$skill" || ! -s "$sweep" ]]; then
  echo "C3 impact bootstrap is missing its frozen instruction sources" >&2
  exit 1
fi
printf 'instruction_sources=skills/c3/SKILL.md,skills/c3/references/sweep.md\n'

set +e
C3X_MODE=agent bash "$wrapper" search "$*" --pack --limit 3 > "$route_output"
route_status=$?
set -e

selected_id=""
if [[ "$route_status" -eq 0 ]]; then
  selected_id="$(/usr/bin/awk -F, '/^  / { sub(/^ +/, "", $1); print $1; exit }' "$route_output")"
fi

if [[ -z "$selected_id" ]]; then
  C3X_MODE=agent bash "$wrapper" list --type component > "$route_output"
  selected_id="$(/usr/bin/awk -F, '/^  [^,]+,component,/ { sub(/^ +/, "", $1); print $1; exit }' "$route_output")"
fi

if [[ -z "${selected_id:-}" || ! "$selected_id" =~ ^[A-Za-z0-9._-]+$ ]]; then
  echo "C3 impact bootstrap could not select a returned fact id" >&2
  exit 1
fi

printf 'selected_id=%s\n' "$selected_id" >&2
C3X_MODE=agent bash "$wrapper" graph "$selected_id" --depth 1 --format mermaid > "$forward_output"
C3X_MODE=agent bash "$wrapper" graph "$selected_id" --direction reverse --depth 1 --format mermaid > "$reverse_output"
C3X_MODE=agent bash "$wrapper" read "$selected_id" > "$evidence_output"

printf 'C3 route pack (primary %s):\n' "$selected_id"
/usr/bin/head -c 1800 "$route_output"
printf '\nforward:\n'
/usr/bin/head -c 240 "$forward_output"
printf '\nreverse:\n'
/usr/bin/head -c 240 "$reverse_output"
printf '\nevidence:\n'
/usr/bin/head -c 350 "$evidence_output"
printf '\ntruth contract: pack class/state describes the stored record, not runtime implementation. Prove current behavior from source anchors; keep decision/planned intent separate from current truth.\n'
printf 'coverage rule: if returned rows have blank record_claims, treat that as zero matching record claims and use direct repository search for the requested mechanisms, not adjacent C3 concepts.\n'
printf 'closure: owner/mutation; consumers/state; persistence/event/retry; failure/isolation; tests. Classify A/U/?; source-close every requested lane before design; unsupported impact stays unknown.\n'
