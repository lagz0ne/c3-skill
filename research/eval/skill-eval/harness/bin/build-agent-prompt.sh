#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: harness/bin/build-agent-prompt.sh <topic>

Build the blindbox prompt packet for one C3 skill-eval topic.
USAGE
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -ne 1 ]]; then
  usage >&2
  exit 2
fi

topic="$1"
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
harness_root="$(cd "$script_dir/.." && pwd)"
skill_eval_root="$(cd "$harness_root/.." && pwd)"
repo_root="$(cd "$skill_eval_root/../../.." && pwd)"
topic_prompt="$harness_root/topics/$topic/prompt.md"

if [[ ! -f "$topic_prompt" ]]; then
  echo "Unknown topic: $topic" >&2
  echo "Available topics:" >&2
  find "$harness_root/topics" -mindepth 1 -maxdepth 1 -type d -printf '  %f\n' | sort >&2
  exit 2
fi

emit_file() {
  local label="$1"
  local path="$2"

  if [[ -n "${C3_PROMPT_SOURCE_LIST:-}" ]]; then
    printf '%s\t%s\n' "$label" "$path" >> "$C3_PROMPT_SOURCE_LIST"
  fi

  printf '\n===== BEGIN FILE: %s =====\n' "$label"
  sed -n '1,$p' "$path"
  printf '\n===== END FILE: %s =====\n' "$label"
}

cat <<'HEADER'
# C3 Blindbox Agent Packet

Use only the context in this packet and the mounted isolated project. Do not rely
on ambient repository files, global skills, plugins, memories, AGENTS.md,
CLAUDE.md, or any other host instruction files.

The local C3 wrapper is mounted in the sandbox at:

```bash
C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh <command>
```

HEADER

emit_file "skills/c3/SKILL.md" "$repo_root/skills/c3/SKILL.md"

for name in audit canvas change onboard query; do
  ref="$repo_root/skills/c3/references/$name.md"
  emit_file "skills/c3/references/$name.md" "$ref"
done

emit_file "harness/prompts/agent-run.md" "$harness_root/prompts/agent-run.md"
emit_file "harness/topics/$topic/prompt.md" "$topic_prompt"
