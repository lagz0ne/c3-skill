#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" == "--child" ]]; then
  status_file="${2:?missing status file}"
  channel="${3:?missing wait channel}"
  shift 3
  [[ "${1:-}" == "--" ]] || exit 2
  shift

  child_status=125
  finish_child() {
    printf '%s\n' "$child_status" > "$status_file"
    tmux wait-for -S "$channel"
  }
  trap finish_child EXIT

  set +e
  "$@"
  child_status=$?
  set -e
  exit "$child_status"
fi

usage() {
  printf '%s\n' 'Usage: run-in-tmux.sh --session <name> -- <command> [args...]' >&2
}

session=""
if [[ "${1:-}" == "--session" ]]; then
  session="${2:-}"
  shift 2
fi
if [[ -z "$session" || "${1:-}" != "--" || $# -lt 2 ]]; then
  usage
  exit 2
fi
shift

if tmux has-session -t "$session" 2>/dev/null; then
  printf 'tmux session already exists: %s\n' "$session" >&2
  exit 2
fi

status_file="$(mktemp)"
channel="c3-tmux-${session}-$$"
cleanup() {
  tmux kill-session -t "$session" 2>/dev/null || true
  rm -f "$status_file"
}
trap cleanup EXIT

printf -v launch '%q ' "$0" --child "$status_file" "$channel" -- "$@"
tmux new-session -d -s "$session" -c "$PWD"
tmux set-option -t "$session" remain-on-exit on >/dev/null
pane_id="$(tmux display-message -p -t "$session" '#{pane_id}')"
tmux respawn-pane -k -t "$pane_id" "$launch"
tmux wait-for "$channel"

tmux capture-pane -p -t "$pane_id" -S -
child_status="$(<"$status_file")"
if [[ ! "$child_status" =~ ^[0-9]+$ ]]; then
  printf 'tmux child did not write a valid exit status\n' >&2
  exit 125
fi
exit "$child_status"
