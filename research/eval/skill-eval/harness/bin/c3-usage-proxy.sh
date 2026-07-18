#!/usr/bin/env bash
set -u

: "${C3_REAL_WRAPPER:?C3_REAL_WRAPPER is required}"
: "${C3_USAGE_LOG:?C3_USAGE_LOG is required}"

case "${1:-}" in
  search|lookup|list) category="route" ;;
  graph) category="impact" ;;
  read|eval) category="evidence" ;;
  *) category="other" ;;
esac

set +e
PATH=/usr/bin:/bin C3X_MODE=agent /bin/bash "$C3_REAL_WRAPPER" "$@"
status=$?
set -e
printf '%s\t%s\n' "$category" "$status" >> "$C3_USAGE_LOG"
exit "$status"
