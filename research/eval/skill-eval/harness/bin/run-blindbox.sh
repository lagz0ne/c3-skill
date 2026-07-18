#!/usr/bin/env bash
set -euo pipefail

runner_started_ms="$(date +%s%3N)"
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
original_args=("$@")

usage() {
  cat <<'USAGE'
Usage: harness/bin/run-blindbox.sh --agent codex|claude|kilo|both|kilo-free|all (--topic <topic> | --prompt-file <file>) [options]

Options:
  --auth env|session   Auth source. env passes whitelisted environment vars.
                       session copies known CLI session auth files into a
                       temporary auth directory mounted into the sandbox.
                       Default: env.
  --model <model>      Pass a model name to the selected agent.
  --effort <level>     Pass low|medium|high|xhigh|max reasoning effort.
  --max-budget-usd N   Claude CLI hard spend ceiling for one run.
  --kilo-agent <name>  Kilo agent to use for kilo runs. Default: code.
  --refresh-models     Refresh Kilo model metadata before resolving free models.
  --run-timeout <sec>  Hard wall-time limit for each child run. Default: 900.
  --max-tool-calls N   Kill the child after N tool calls start. Default: 6.
  --max-tool-result-bytes N
                       Kill after cumulative model-visible tool results exceed N.
  --max-output-bytes N Kill the child at this stdout+stderr size. Default: 524288.
  --label <label>      Stable output label. With --agent both/all, this is a prefix.
  --prompt-file <file> Use a caller-built prompt instead of a bundled topic.
  --seed-repo <path>   Seed the disposable workspace from this Git checkout's HEAD.
  --condition <arm>    with_c3 or without_c3. Required with --prompt-file.
  --run-dir <path>     Raw output directory. Use a temporary private path for paired runs.
  --backend <mode>     direct or tmux. Tmux runs the same command detached and
                       permits observation without mid-run prompt changes.
  --dry-run            Print isolation metadata and redacted command only.
  -h, --help           Show this help.

The runner mounts no repository root. The writable project lives under
harness/runs/<label>.workspace and is mounted as /work/project.
USAGE
}

agent=""
claude_base_system_prompt="You are an isolated repository analysis agent. Follow the mounted task and repository instructions. Work only in /work/project. Use only the allowed tools. Return the requested compact final answer."
codex_model_instructions_source=""
topic=""
prompt_input=""
seed_repo=""
condition="with_c3"
condition_explicit=0
run_dir_override=""
backend="direct"
auth_mode="env"
model=""
effort=""
max_budget_usd=""
kilo_agent="code"
refresh_models=0
run_timeout=900
max_tool_calls=6
max_tool_result_bytes=131072
max_output_bytes=524288
label=""
dry_run=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --agent)
      agent="${2:-}"
      shift 2
      ;;
    --topic)
      topic="${2:-}"
      shift 2
      ;;
    --prompt-file)
      prompt_input="${2:-}"
      shift 2
      ;;
    --seed-repo)
      seed_repo="${2:-}"
      shift 2
      ;;
    --condition)
      condition="${2:-}"
      condition_explicit=1
      shift 2
      ;;
    --run-dir)
      run_dir_override="${2:-}"
      shift 2
      ;;
    --backend)
      backend="${2:-}"
      shift 2
      ;;
    --auth)
      auth_mode="${2:-}"
      shift 2
      ;;
    --model)
      model="${2:-}"
      shift 2
      ;;
    --effort)
      effort="${2:-}"
      shift 2
      ;;
    --max-budget-usd)
      max_budget_usd="${2:-}"
      shift 2
      ;;
    --kilo-agent)
      kilo_agent="${2:-}"
      shift 2
      ;;
    --refresh-models)
      refresh_models=1
      shift
      ;;
    --run-timeout)
      run_timeout="${2:-}"
      shift 2
      ;;
    --max-tool-calls)
      max_tool_calls="${2:-}"
      shift 2
      ;;
    --max-tool-result-bytes)
      max_tool_result_bytes="${2:-}"
      shift 2
      ;;
    --max-output-bytes)
      max_output_bytes="${2:-}"
      shift 2
      ;;
    --label)
      label="${2:-}"
      shift 2
      ;;
    --dry-run)
      dry_run=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ -z "$agent" || ( -z "$topic" && -z "$prompt_input" ) || ( -n "$topic" && -n "$prompt_input" ) ]]; then
  usage >&2
  exit 2
fi

if [[ -n "$prompt_input" ]]; then
  if [[ "$condition_explicit" -ne 1 || -z "$seed_repo" ]]; then
    echo "--prompt-file requires --seed-repo and explicit --condition" >&2
    exit 2
  fi
  if [[ ! -f "$prompt_input" ]]; then
    echo "Missing prompt file: $prompt_input" >&2
    exit 2
  fi
  if ! git -C "$seed_repo" rev-parse --verify HEAD >/dev/null 2>&1; then
    echo "--seed-repo must be a Git checkout with a readable HEAD" >&2
    exit 2
  fi
fi

if [[ "$condition" != "with_c3" && "$condition" != "without_c3" ]]; then
  echo "Unsupported condition: $condition" >&2
  exit 2
fi

if [[ "$auth_mode" != "env" && "$auth_mode" != "session" ]]; then
  echo "Unsupported auth mode: $auth_mode" >&2
  exit 2
fi

if [[ "$backend" != "direct" && "$backend" != "tmux" ]]; then
  echo "Unsupported backend: $backend" >&2
  exit 2
fi

if [[ "$backend" == "tmux" && "${C3_BLINDBOX_TMUX_CHILD:-0}" != "1" ]]; then
  session_name="c3-${label:-blindbox-$$}"
  session_name="${session_name//[^[:alnum:]_-]/-}"
  exec "$script_dir/run-in-tmux.sh" --session "$session_name" -- \
    env C3_BLINDBOX_TMUX_CHILD=1 "$0" "${original_args[@]}"
fi

if [[ -n "$effort" && ! "$effort" =~ ^(low|medium|high|xhigh|max)$ ]]; then
  echo "Unsupported effort: $effort" >&2
  exit 2
fi

if [[ -n "$max_budget_usd" && ! "$max_budget_usd" =~ ^[0-9]+([.][0-9]+)?$ ]]; then
  echo "--max-budget-usd must be a non-negative number" >&2
  exit 2
fi

if ! [[ "$run_timeout" =~ ^[1-9][0-9]*$ ]]; then
  echo "--run-timeout must be a positive integer number of seconds" >&2
  exit 2
fi

if ! [[ "$max_tool_calls" =~ ^[1-9][0-9]*$ ]]; then
  echo "--max-tool-calls must be a positive integer" >&2
  exit 2
fi

if ! [[ "$max_tool_result_bytes" =~ ^[1-9][0-9]*$ ]]; then
  echo "--max-tool-result-bytes must be a positive integer" >&2
  exit 2
fi

if ! [[ "$max_output_bytes" =~ ^[1-9][0-9]*$ ]]; then
  echo "--max-output-bytes must be a positive integer" >&2
  exit 2
fi

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 127
  fi
}

need_for_run() {
  if [[ "$dry_run" -eq 0 ]]; then
    need "$1"
  fi
}

resolve_cmd() {
  local name="$1"
  if command -v "$name" >/dev/null 2>&1; then
    command -v "$name"
  elif [[ "$dry_run" -eq 1 ]]; then
    printf '/missing/%s\n' "$name"
  else
    echo "Missing required command: $name" >&2
    exit 127
  fi
}

quote_cmd_redacted() {
  local state="normal"
  local name=""

  for arg in "$@"; do
    case "$state" in
      normal)
        printf '%q ' "$arg"
        if [[ "$arg" == "--setenv" ]]; then
          state="setenv_name"
        fi
        ;;
      setenv_name)
        name="$arg"
        printf '%q ' "$arg"
        state="setenv_value"
        ;;
      setenv_value)
        if [[ "$name" =~ (KEY|TOKEN|SECRET|AUTH|PASSWORD|CREDENTIAL) ]]; then
          printf '%q ' "<redacted>"
        else
          printf '%q ' "$arg"
        fi
        state="normal"
        ;;
    esac
  done
  printf '\n'
}

slugify() {
  printf '%s' "$1" | sed -E 's#^kilo/##; s#[^A-Za-z0-9._-]+#-#g; s#[-._]+$##'
}

sha_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  else
    shasum -a 256 "$1" | awk '{print $1}'
  fi
}

sha_text() {
  if command -v sha256sum >/dev/null 2>&1; then
    printf '%s' "$1" | sha256sum | awk '{print $1}'
  else
    printf '%s' "$1" | shasum -a 256 | awk '{print $1}'
  fi
}

sha_tree() {
  python3 - "$1" <<'PY'
import hashlib
import pathlib
import sys

root = pathlib.Path(sys.argv[1])
digest = hashlib.sha256()
for path in sorted(item for item in root.rglob("*") if item.is_file()):
    relative_path = path.relative_to(root)
    if relative_path.parts[0] == "bin":
        continue
    relative = relative_path.as_posix().encode("utf-8")
    content = hashlib.sha256(path.read_bytes()).hexdigest().encode("ascii")
    digest.update(relative)
    digest.update(b"\0")
    digest.update(content)
    digest.update(b"\n")
print(digest.hexdigest())
PY
}

expect_sha() {
  local label="$1"
  local actual="$2"
  local expected="$3"
  if [[ -z "$expected" ]]; then
    return 0
  fi
  if [[ "$actual" != "$expected" ]]; then
    echo "Hash mismatch for $label: expected $expected, got $actual" >&2
    exit 2
  fi
}

resolve_local_c3_binary() {
  local version_file="$repo_root/skills/c3/bin/VERSION"
  local wrapper="$repo_root/skills/c3/bin/c3x.sh"
  if [[ ! -f "$version_file" ]]; then
    echo "Missing local C3 VERSION: $version_file" >&2
    exit 2
  fi
  if [[ ! -f "$wrapper" ]]; then
    echo "Missing local C3 wrapper: $wrapper" >&2
    exit 2
  fi

  local version os arch
  version="$(tr -d '[:space:]' < "$version_file")"
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"
  case "$arch" in
    x86_64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
  esac
  case "$os/$arch" in
    linux/amd64|linux/arm64|darwin/arm64) ;;
    *)
      echo "Unsupported local C3 platform: $os/$arch" >&2
      exit 2
      ;;
  esac

  local binary="$repo_root/skills/c3/bin/c3x-${version}-${os}-${arch}"
  if [[ -n "${C3_FROZEN_BINARY:-}" ]]; then
    binary="$(realpath "$C3_FROZEN_BINARY")"
  fi
  if [[ ! -x "$binary" ]]; then
    echo "Missing executable local C3 binary: $binary" >&2
    exit 2
  fi

  C3_LOCAL_VERSION="$version"
  C3_LOCAL_PLATFORM="$os/$arch"
  C3_LOCAL_BINARY="$binary"
  if [[ -n "${C3_FROZEN_BINARY:-}" ]]; then
    C3_LOCAL_BINARY_SANDBOX="/opt/c3/bin/frozen-c3x"
  else
    C3_LOCAL_BINARY_SANDBOX="/opt/c3/skills/c3/bin/$(basename "$binary")"
  fi
}

write_provenance() {
  local file="$1"
  local prompt_sha version_sha wrapper_sha binary_sha skill_tree_sha seed_head seed_head_sha
  prompt_sha="$(sha_file "$prompt_archive")"

  expect_sha "prompt_archive" "$prompt_sha" "${C3_EXPECT_PROMPT_SHA256:-}"

  {
    printf 'prompt_archive=%s\n' "$prompt_archive"
    printf 'prompt_sha256=%s\n' "$prompt_sha"
    printf 'prompt_source_list=%s\n' "$prompt_source_list"
    printf 'prompt_source_count=%s\n' "$(wc -l < "$prompt_source_list" | tr -d '[:space:]')"
    printf 'codex_model_instructions_sha256=%s\n' "$(sha_file "$codex_model_instructions_source")"
    if [[ "$condition" == "with_c3" ]]; then
      version_sha="$(sha_file "$repo_root/skills/c3/bin/VERSION")"
      wrapper_sha="$(sha_file "$repo_root/skills/c3/bin/c3x.sh")"
      binary_sha="$(sha_file "$C3_LOCAL_BINARY")"
      skill_tree_sha="$(sha_tree "$repo_root/skills/c3")"
      expect_sha "version file" "$version_sha" "${C3_EXPECT_VERSION_SHA256:-}"
      expect_sha "wrapper" "$wrapper_sha" "${C3_EXPECT_WRAPPER_SHA256:-}"
      expect_sha "selected_binary" "$binary_sha" "${C3_EXPECT_SELECTED_BINARY_SHA256:-}"
      expect_sha "skills/c3 tree" "$skill_tree_sha" "${C3_EXPECT_SKILL_TREE_SHA256:-}"
      printf 'local_c3_version=%s\n' "$C3_LOCAL_VERSION"
      printf 'local_c3_platform=%s\n' "$C3_LOCAL_PLATFORM"
      printf 'version_sha256=%s\n' "$version_sha"
      printf 'wrapper_sha256=%s\n' "$wrapper_sha"
      printf 'selected_binary=%s\n' "$C3_LOCAL_BINARY"
      printf 'selected_binary_sandbox=%s\n' "$C3_LOCAL_BINARY_SANDBOX"
      printf 'selected_binary_sha256=%s\n' "$binary_sha"
      printf 'skill_tree_sha256=%s\n' "$skill_tree_sha"
    else
      printf 'local_c3_version=unmounted\n'
      printf 'selected_binary=unmounted\n'
    fi
    printf 'repo_head=%s\n' "$(git -C "$repo_root" rev-parse HEAD 2>/dev/null || printf 'unknown')"
    if [[ -n "$(git -C "$repo_root" status --porcelain 2>/dev/null)" ]]; then
      printf 'repo_dirty=yes\n'
    else
      printf 'repo_dirty=no\n'
    fi
    if [[ "$condition" == "with_c3" ]]; then
      printf 'c3_bootstrap=/opt/c3/bin/c3-impact-bootstrap\n'
    else
      printf 'c3_bootstrap=unmounted\n'
    fi
    printf 'path_has_usr_local_bin=no\n'
    if [[ -n "$seed_repo" ]]; then
      seed_head="$(git -C "$seed_repo" rev-parse HEAD)"
      seed_head_sha="$(sha_text "$seed_head")"
      expect_sha "seed_repo_head" "$seed_head_sha" "${C3_EXPECT_SEED_HEAD_SHA256:-}"
      printf 'seed_repo_head=%s\n' "$seed_head"
      printf 'seed_repo_head_sha256=%s\n' "$seed_head_sha"
    fi
    while IFS=$'\t' read -r source_label source_path; do
      [[ -n "$source_label" && -n "$source_path" ]] || continue
      source_sha="$(sha_file "$source_path")"
      if [[ "$source_label" == "skills/c3/SKILL.md" ]]; then
        expect_sha "$source_label" "$source_sha" "${C3_EXPECT_SKILL_MD_SHA256:-}"
      fi
      printf 'source_sha256[%s]=%s\n' "$source_label" "$source_sha"
    done < "$prompt_source_list"
  } > "$file"
}

kilo_free_models() {
  local refresh_arg=()
  [[ "$refresh_models" -eq 1 ]] && refresh_arg=(--refresh)
  kilo models --verbose "${refresh_arg[@]}" | awk '
    /^kilo\// { model=$0 }
    /"isFree": true/ { print model }
  '
}

timestamp="${HARNESS_RUN_ID:-$(date -u +%Y%m%dT%H%M%SZ)}"
harness_root="$(cd "$script_dir/.." && pwd)"
repo_root="$(cd "$harness_root/../../../.." && pwd)"
codex_model_instructions_source="$harness_root/prompts/minimal-codex-model-instructions.md"
run_dir="$harness_root/runs"
if [[ -n "$run_dir_override" ]]; then
  run_dir="$run_dir_override"
fi
mkdir -p "$run_dir"

if [[ "$agent" == "both" || "$agent" == "all" ]]; then
  if [[ -n "$prompt_input" ]]; then
    echo "generic --prompt-file mode runs one agent at a time" >&2
    exit 2
  fi
  args=(--topic "$topic" --auth "$auth_mode")
  [[ -n "$model" ]] && args+=(--model "$model")
  [[ -n "$effort" ]] && args+=(--effort "$effort")
  [[ "$dry_run" -eq 1 ]] && args+=(--dry-run)
  args+=(--run-timeout "$run_timeout" --max-tool-calls "$max_tool_calls" --max-tool-result-bytes "$max_tool_result_bytes" --max-output-bytes "$max_output_bytes")
  if [[ -n "$label" ]]; then
    "$0" --agent codex "${args[@]}" --label "$label-codex"
    "$0" --agent claude "${args[@]}" --label "$label-claude"
  else
    "$0" --agent codex "${args[@]}"
    "$0" --agent claude "${args[@]}"
  fi
  if [[ "$agent" == "all" ]]; then
    kilo_args=(--topic "$topic" --auth "$auth_mode" --kilo-agent "$kilo_agent")
    [[ "$refresh_models" -eq 1 ]] && kilo_args+=(--refresh-models)
    [[ "$dry_run" -eq 1 ]] && kilo_args+=(--dry-run)
    kilo_args+=(--run-timeout "$run_timeout" --max-tool-calls "$max_tool_calls" --max-tool-result-bytes "$max_tool_result_bytes" --max-output-bytes "$max_output_bytes")
    if [[ -n "$label" ]]; then
      "$0" --agent kilo-free "${kilo_args[@]}" --label "$label-kilo-free"
    else
      "$0" --agent kilo-free "${kilo_args[@]}"
    fi
  fi
  exit 0
fi

if [[ "$agent" == "kilo-free" ]]; then
  if [[ -n "$model" ]]; then
    echo "--agent kilo-free discovers free Kilo models automatically; use --agent kilo --model <model>" >&2
    exit 2
  fi
  if [[ "$dry_run" -eq 1 ]]; then
    read -r -a dry_models <<< "${KILO_FREE_MODELS:-kilo/kilo-auto/free}"
    for kilo_model in "${dry_models[@]}"; do
      child_label="${label:-$timestamp-kilo-free-$topic}-$(slugify "$kilo_model")"
      "$0" --agent kilo --topic "$topic" --auth "$auth_mode" --model "$kilo_model" --kilo-agent "$kilo_agent" --label "$child_label" --dry-run
    done
    exit 0
  fi
  need kilo
  mapfile -t models < <(kilo_free_models)
  if [[ "${#models[@]}" -eq 0 ]]; then
    echo "No free Kilo models found from: kilo models --verbose" >&2
    exit 2
  fi
  fanout_status=0
  for kilo_model in "${models[@]}"; do
    child_label="${label:-$timestamp-kilo-free-$topic}-$(slugify "$kilo_model")"
    child_args=(--agent kilo --topic "$topic" --auth "$auth_mode" --model "$kilo_model" --kilo-agent "$kilo_agent" --label "$child_label")
    [[ "$dry_run" -eq 1 ]] && child_args+=(--dry-run)
    child_args+=(--run-timeout "$run_timeout" --max-tool-calls "$max_tool_calls" --max-tool-result-bytes "$max_tool_result_bytes" --max-output-bytes "$max_output_bytes")
    if ! "$0" "${child_args[@]}"; then
      echo "Kilo model failed: $kilo_model" >&2
      fanout_status=1
    fi
  done
  exit "$fanout_status"
fi

if [[ "$agent" != "codex" && "$agent" != "claude" && "$agent" != "kilo" ]]; then
  echo "Unsupported agent: $agent" >&2
  exit 2
fi

need_for_run bwrap
need_for_run python3
need_for_run jq

if [[ -z "$label" ]]; then
  label="$timestamp-$agent-${topic:-paired}"
fi

final_file="$run_dir/$label.md"
stdout_file="$run_dir/$label.stdout.txt"
stderr_file="$run_dir/$label.stderr.txt"
meta_file="$run_dir/$label.meta.txt"
workspace_dir="$run_dir/$label.workspace"
agent_output_dir="$run_dir/.agent-output-$label"
agent_final_file="$agent_output_dir/$label.md"
c3_usage_dir="$workspace_dir/.c3-eval"
c3_usage_file="$c3_usage_dir/$label.c3-usage.tsv"
c3_sandbox_cache_dir="$c3_usage_dir/cache"
prompt_file="$(mktemp)"
prompt_source_list="$run_dir/$label.prompt.sources.tsv"
prompt_archive="$run_dir/$label.prompt.md"
provenance_file="$run_dir/$label.provenance.txt"
guard_status_file="$run_dir/$label.guard.json"
auth_root=""
trap 'rm -f "$prompt_file"; [[ -z "$auth_root" ]] || rm -rf "$auth_root"' EXIT

rm -rf "$agent_output_dir"
mkdir -p "$agent_output_dir"
if [[ "$condition" == "with_c3" ]]; then
  mkdir -p "$c3_sandbox_cache_dir"
fi

rm -f "$prompt_source_list"
if [[ -n "$prompt_input" ]]; then
  cp "$prompt_input" "$prompt_file"
  printf 'case-prompt\t%s\n' "$prompt_input" > "$prompt_source_list"
else
  C3_PROMPT_SOURCE_LIST="$prompt_source_list" "$harness_root/bin/build-agent-prompt.sh" "$topic" > "$prompt_file"
fi
cp "$prompt_file" "$prompt_archive"
if [[ "$condition" == "with_c3" ]]; then
  resolve_local_c3_binary
fi
write_provenance "$provenance_file"

seed_workspace() {
  rm -rf "$workspace_dir"
  mkdir -p "$workspace_dir"
  if [[ -n "$seed_repo" ]]; then
    git -C "$seed_repo" archive --format=tar HEAD | tar -xf - -C "$workspace_dir"
    find "$workspace_dir" -type f \( -name AGENTS.md -o -name CLAUDE.md \) -delete
    find "$workspace_dir" -type d \( -name .agents -o -name .claude -o -name .codex \) -prune -exec rm -rf {} +
    if [[ "$condition" == "without_c3" ]]; then
      find "$workspace_dir" -type d -name .c3 -prune -exec rm -rf {} +
    else
      find "$workspace_dir" -type f -path '*/.c3/c3.db' -delete
    fi
    return
  fi
  cat > "$workspace_dir/README.md" <<'README'
# Blindbox C3 Growth Project

This is an isolated eval project. Use the mounted local C3 wrapper only:

```bash
C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh <command>
```
README
}

seed_workspace

if [[ "$condition" == "with_c3" ]]; then
  mkdir -p "$c3_usage_dir"
  rm -f "$c3_usage_file"
fi

ambient_instruction_file_count="$(find "$workspace_dir" -type f \( -name AGENTS.md -o -name CLAUDE.md \) -print | wc -l | tr -d '[:space:]')"
ambient_agent_dir_count="$(find "$workspace_dir" -type d \( -name .agents -o -name .claude -o -name .codex \) -print | wc -l | tr -d '[:space:]')"
if [[ "$ambient_instruction_file_count" -ne 0 || "$ambient_agent_dir_count" -ne 0 ]]; then
  echo "Ambient agent instructions remain after workspace sanitization" >&2
  exit 2
fi

baseline_source="$harness_root/prompts/neutral-repository-baseline.md"
cp "$baseline_source" "$workspace_dir/AGENTS.md"
cp "$baseline_source" "$workspace_dir/CLAUDE.md"
baseline_agents_sha256="$(sha_file "$workspace_dir/AGENTS.md")"
baseline_claude_sha256="$(sha_file "$workspace_dir/CLAUDE.md")"
baseline_instruction_hash_match=0
if [[ "$baseline_agents_sha256" == "$baseline_claude_sha256" ]]; then
  baseline_instruction_hash_match=1
fi
baseline_instruction_file_count="$(find "$workspace_dir" -type f \( -name AGENTS.md -o -name CLAUDE.md \) -print | wc -l | tr -d '[:space:]')"
unexpected_instruction_file_count="$(find "$workspace_dir" -type f \( -name AGENTS.md -o -name CLAUDE.md \) ! -path "$workspace_dir/AGENTS.md" ! -path "$workspace_dir/CLAUDE.md" -print | wc -l | tr -d '[:space:]')"
baseline_c3_reference_count="$(grep -Eic '(^|[^[:alnum:]_])c3([^[:alnum:]_]|$)' "$baseline_source" || true)"
if [[ "$baseline_instruction_hash_match" -ne 1 || "$baseline_instruction_file_count" -ne 2 || "$unexpected_instruction_file_count" -ne 0 || "$baseline_c3_reference_count" -ne 0 ]]; then
  echo "Neutral baseline parity check failed" >&2
  exit 2
fi

treatment_instruction_policy="none"
treatment_instruction_sha256="unmounted"
treatment_runtime_layer="none"
treatment_instruction_text=""
treatment_instruction_toml=""
if [[ "$condition" == "with_c3" ]]; then
  treatment_source="$harness_root/prompts/c3-impact-treatment.md"
  treatment_instruction_sha256="$(sha_file "$treatment_source")"
  treatment_instruction_text="$(<"$treatment_source")"
  treatment_instruction_toml="$(jq -Rs . "$treatment_source")"
  if [[ "$agent" == "kilo" ]]; then
    for instruction_file in "$workspace_dir/AGENTS.md" "$workspace_dir/CLAUDE.md"; do
      printf '\n' >> "$instruction_file"
      cat "$treatment_source" >> "$instruction_file"
    done
  fi
  treatment_agents_sha256="$(sha_file "$workspace_dir/AGENTS.md")"
  treatment_claude_sha256="$(sha_file "$workspace_dir/CLAUDE.md")"
  if [[ "$treatment_agents_sha256" != "$treatment_claude_sha256" ]]; then
    echo "C3 treatment instruction parity check failed" >&2
    exit 2
  fi
  treatment_instruction_policy="forced_c3_impact"
  case "$agent" in
    codex) treatment_runtime_layer="codex_developer_instructions" ;;
    claude) treatment_runtime_layer="claude_append_system_prompt" ;;
    *) treatment_runtime_layer="repository_instructions" ;;
  esac
fi

sandbox_path="/opt/node/bin:/opt/claude:/opt/kilo:/usr/bin:/bin"

bwrap_args=(
  bwrap
  --die-with-parent
  --unshare-pid
  --unshare-ipc
  --new-session
  --clearenv
  --proc /proc
  --dev /dev
  --tmpfs /tmp
  --dir /run
  --dir /run/systemd
  --dir /work
  --dir /work/home
  --dir /work/home/.cache
  --dir /work/home/.config
  --dir /work/home/.local
  --dir /work/home/.local/share
  --dir /opt
  --bind "$workspace_dir" /work/project
  --bind "$agent_output_dir" /runs
  --ro-bind "$prompt_file" /prompt.md
  --ro-bind "$codex_model_instructions_source" /opt/codex-model-instructions.md
  --chdir /work/project
  --setenv HOME /work/home
  --setenv USER blindbox
  --setenv LOGNAME blindbox
  --setenv XDG_CACHE_HOME /work/project/.c3-eval/cache
  --setenv XDG_CONFIG_HOME /work/home/.config
  --setenv XDG_DATA_HOME /work/home/.local/share
  --setenv CODEX_HOME /work/home/.codex
  --setenv CLAUDE_CONFIG_DIR /work/home/.claude
  --setenv CLAUDE_CODE_SAFE_MODE 1
  --setenv PATH "$sandbox_path"
  --setenv NO_COLOR 1
)

if [[ "$condition" == "with_c3" ]]; then
  bwrap_args+=(
    --dir /opt/c3
    --dir /opt/c3/bin
    --dir /opt/c3/skills
    --dir /opt/c3/source
    --ro-bind "$harness_root/bin/c3-impact-bootstrap.sh" /opt/c3/bin/c3-impact-bootstrap
    --ro-bind "$repo_root/skills/c3" /opt/c3/skills/c3
    --ro-bind "$repo_root/skills/c3" /opt/c3/source/c3
    --ro-bind "$harness_root/bin/c3-usage-proxy.sh" /opt/c3/skills/c3/bin/c3x.sh
    --dir /work/project/.agents
    --dir /work/project/.agents/skills
    --ro-bind "$repo_root/skills/c3" /work/project/.agents/skills/c3
    --ro-bind "$harness_root/bin/c3-usage-proxy.sh" /work/project/.agents/skills/c3/bin/c3x.sh
    --dir /work/project/.claude
    --dir /work/project/.claude/skills
    --ro-bind "$repo_root/skills/c3" /work/project/.claude/skills/c3
    --ro-bind "$harness_root/bin/c3-usage-proxy.sh" /work/project/.claude/skills/c3/bin/c3x.sh
    --setenv C3_REAL_WRAPPER /opt/c3/source/c3/bin/c3x.sh
    --setenv C3_USAGE_LOG "/work/project/.c3-eval/$label.c3-usage.tsv"
  )
  if [[ -n "${C3_FROZEN_BINARY:-}" ]]; then
    bwrap_args+=(
      --ro-bind "$C3_LOCAL_BINARY" /opt/c3/bin/frozen-c3x
      --setenv C3X_LOCAL_BINARY /opt/c3/bin/frozen-c3x
      --setenv C3X_LOCAL_VERSION "$C3_LOCAL_VERSION"
    )
  fi
fi

auth_source="whitelisted_environment"

if [[ -d /run/systemd/resolve ]]; then
  bwrap_args+=(--ro-bind /run/systemd/resolve /run/systemd/resolve)
fi

stage_session_auth() {
  auth_root="$(mktemp -d)"
  chmod 700 "$auth_root"

  case "$agent" in
    codex)
      local codex_src="${CODEX_HOST_AUTH:-$HOME/.codex/auth.json}"
      if [[ ! -f "$codex_src" ]]; then
        echo "Missing Codex session auth file: $codex_src" >&2
        exit 2
      fi
      mkdir -p "$auth_root/codex"
      install -m 600 "$codex_src" "$auth_root/codex/auth.json"
      bwrap_args+=(--bind "$auth_root/codex" /work/home/.codex)
      bwrap_args+=(--dir /work/home/.claude)
      auth_source="temporary_session_copy:codex_auth_json"
      ;;
    claude)
      local claude_src="${CLAUDE_HOST_CREDENTIALS:-$HOME/.claude/.credentials.json}"
      if [[ ! -f "$claude_src" ]]; then
        echo "Missing Claude session auth file: $claude_src" >&2
        exit 2
      fi
      mkdir -p "$auth_root/claude"
      install -m 600 "$claude_src" "$auth_root/claude/.credentials.json"
      bwrap_args+=(--dir /work/home/.codex)
      bwrap_args+=(--bind "$auth_root/claude" /work/home/.claude)
      # claude-code is "logged in" only when it also finds the account config
      # (~/.claude.json with oauthAccount) — the .credentials.json tokens alone
      # read as "Not logged in". Stage it minus the project history (the bulk +
      # the only sensitive part) so the sandbox sees the account, not the history.
      local claude_cfg="${CLAUDE_HOST_CONFIG:-$HOME/.claude.json}"
      if [[ -f "$claude_cfg" ]]; then
        # account config, minus project history; stage in both the legacy HOME
        # location and inside CLAUDE_CONFIG_DIR so either resolution path finds it.
        jq 'del(.projects)' "$claude_cfg" > "$auth_root/claude.json" 2>/dev/null \
          || install -m 600 "$claude_cfg" "$auth_root/claude.json"
        chmod 600 "$auth_root/claude.json"
        install -m 600 "$auth_root/claude.json" "$auth_root/claude/.claude.json"
        bwrap_args+=(--bind "$auth_root/claude.json" /work/home/.claude.json)
      fi
      # The `claude --print` headless path authenticates from CLAUDE_CODE_OAUTH_TOKEN;
      # feed it the live OAuth access token so a fresh sandbox is logged in without
      # the interactive credentials handshake.
      local claude_token
      claude_token="$(jq -r '.claudeAiOauth.accessToken // empty' "$claude_src" 2>/dev/null)"
      if [[ -n "$claude_token" ]]; then
        bwrap_args+=(--setenv CLAUDE_CODE_OAUTH_TOKEN "$claude_token")
      fi
      auth_source="temporary_session_copy:claude_credentials_json+account+oauth_token"
      ;;
    kilo)
      local kilo_auth_src="${KILO_HOST_AUTH:-$HOME/.local/share/kilo/auth.json}"
      local kilo_config_src="${KILO_HOST_CONFIG:-$HOME/.config/kilo/kilo.jsonc}"
      if [[ ! -f "$kilo_auth_src" ]]; then
        echo "Missing Kilo session auth file: $kilo_auth_src" >&2
        exit 2
      fi
      mkdir -p "$auth_root/kilo-data/kilo" "$auth_root/kilo-config/kilo"
      install -m 600 "$kilo_auth_src" "$auth_root/kilo-data/kilo/auth.json"
      if [[ -f "$kilo_config_src" ]]; then
        install -m 600 "$kilo_config_src" "$auth_root/kilo-config/kilo/kilo.jsonc"
      fi
      bwrap_args+=(--dir /work/home/.codex)
      bwrap_args+=(--dir /work/home/.claude)
      bwrap_args+=(--bind "$auth_root/kilo-data/kilo" /work/home/.local/share/kilo)
      bwrap_args+=(--bind "$auth_root/kilo-config/kilo" /work/home/.config/kilo)
      auth_source="temporary_session_copy:kilo_auth_json"
      ;;
  esac
}

if [[ "$auth_mode" == "session" ]]; then
  stage_session_auth
else
  bwrap_args+=(--dir /work/home/.codex)
  bwrap_args+=(--dir /work/home/.claude)
fi

for path in /usr /bin /lib /lib64 /etc; do
  if [[ -e "$path" ]]; then
    bwrap_args+=(--ro-bind "$path" "$path")
  fi
done

for name in \
  OPENAI_API_KEY OPENAI_BASE_URL OPENAI_ORG_ID OPENAI_PROJECT \
  ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN ANTHROPIC_BASE_URL \
  KILO_API_KEY KILO_AUTH_TOKEN KILO_SERVER_USERNAME KILO_SERVER_PASSWORD \
  AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY AWS_SESSION_TOKEN AWS_REGION AWS_PROFILE \
  HTTPS_PROXY HTTP_PROXY NO_PROXY SSL_CERT_FILE REQUESTS_CA_BUNDLE; do
  if [[ -n "${!name:-}" ]]; then
    bwrap_args+=(--setenv "$name" "${!name}")
  fi
done

case "$agent" in
  codex)
    need_for_run codex
    need_for_run node
    codex_cmd="$(resolve_cmd codex)"
    node_cmd="$(resolve_cmd node)"
    if [[ "$codex_cmd" == /missing/* || "$node_cmd" == /missing/* ]]; then
      agent_cmd=("$codex_cmd" exec --json --ask-for-approval never --sandbox workspace-write --skip-git-repo-check --ignore-user-config --cd /work/project --output-last-message "/runs/$label.md")
    else
      codex_js="$(readlink -f "$codex_cmd")"
      node_bin="$(readlink -f "$node_cmd")"
      node_root="$(cd "$(dirname "$node_bin")/.." && pwd)"
      codex_js_sandbox="/opt/node/${codex_js#"$node_root"/}"
      bwrap_args+=(--ro-bind "$node_root" /opt/node)
      agent_cmd=(/opt/node/bin/node "$codex_js_sandbox"
        --ask-for-approval never
        exec
        --json
        --skip-git-repo-check
        --ephemeral
        --ignore-user-config
        --sandbox workspace-write
        --cd /work/project
        --output-last-message "/runs/$label.md"
      )
    fi
    [[ -n "$model" ]] && agent_cmd+=(--model "$model")
    [[ -n "$effort" ]] && agent_cmd+=(-c "model_reasoning_effort=\"$effort\"")
    agent_cmd+=(-c 'model_instructions_file="/opt/codex-model-instructions.md"')
    agent_cmd+=(-c 'tool_output_token_limit=400')
    if [[ "$condition" == "with_c3" ]]; then
      agent_cmd+=(-c "developer_instructions=$treatment_instruction_toml")
    fi
    agent_cmd+=(-)
    ;;
  claude)
    need_for_run claude
    need_for_run jq
    claude_cmd="$(resolve_cmd claude)"
    if [[ "$claude_cmd" == /missing/* ]]; then
      agent_cmd=("$claude_cmd" --print)
    else
      claude_bin="$(readlink -f "$claude_cmd")"
      bwrap_args+=(--dir /opt/claude --ro-bind "$claude_bin" /opt/claude/claude)
      # NOTE: do NOT pass --bare. --bare only accepts an ANTHROPIC_API_KEY and
      # rejects the OAuth/subscription token (CLAUDE_CODE_OAUTH_TOKEN) we stage,
      # surfacing as "Not logged in". Plain --print honors the OAuth token.
      claude_allowed_tools="Bash,Read,Write,Edit,MultiEdit,Glob,Grep,LS"
      agent_cmd=(/opt/claude/claude
        --print
        --safe-mode
        --no-session-persistence
        --system-prompt "$claude_base_system_prompt"
        --permission-mode bypassPermissions
        --allowedTools "$claude_allowed_tools"
        --verbose
        --output-format stream-json
      )
    fi
    [[ -n "$model" ]] && agent_cmd+=(--model "$model")
    [[ -n "$effort" ]] && agent_cmd+=(--effort "$effort")
    [[ -n "$max_budget_usd" ]] && agent_cmd+=(--max-budget-usd "$max_budget_usd")
    if [[ "$condition" == "with_c3" ]]; then
      agent_cmd+=(--append-system-prompt "$treatment_instruction_text")
    fi
    ;;
  kilo)
    need_for_run kilo
    need_for_run node
    need_for_run jq
    kilo_cmd="$(resolve_cmd kilo)"
    node_cmd="$(resolve_cmd node)"
    if [[ "$kilo_cmd" == /missing/* || "$node_cmd" == /missing/* ]]; then
      agent_cmd=("$kilo_cmd" run --format json --file /prompt.md --dir /work/project --agent "$kilo_agent")
    else
      kilo_js="$(readlink -f "$kilo_cmd")"
      node_bin="$(readlink -f "$node_cmd")"
      node_root="$(cd "$(dirname "$node_bin")/.." && pwd)"
      kilo_root="$(cd "$(dirname "$kilo_js")/../../../.." && pwd)"
      kilo_js_sandbox="/opt/kilo-global/${kilo_js#"$kilo_root"/}"
      bwrap_args+=(--ro-bind "$node_root" /opt/node)
      bwrap_args+=(--ro-bind "$kilo_root" /opt/kilo-global)
      bwrap_args+=(--dir /opt/kilo --symlink /opt/kilo-global/node_modules/@kilocode/cli/bin/kilo /opt/kilo/kilo)
      agent_cmd=(/opt/node/bin/node "$kilo_js_sandbox"
        run
        --format json
        --file /prompt.md
        --dir /work/project
        --agent "$kilo_agent"
        --auto
        --title "$label"
      )
    fi
    [[ -n "$model" ]] && agent_cmd+=(--model "$model")
    agent_cmd+=("Use /prompt.md as the full task. Work only in /work/project and return the requested C3 growth artifacts.")
    ;;
esac

{
  printf 'agent=%s\n' "$agent"
  printf 'topic=%s\n' "${topic:-external-prompt}"
  printf 'condition=%s\n' "$condition"
  printf 'execution_backend=%s\n' "$backend"
  printf 'auth_mode=%s\n' "$auth_mode"
  printf 'model=%s\n' "${model:-default}"
  printf 'reasoning_effort=%s\n' "${effort:-default}"
  if [[ "$agent" == "claude" ]]; then
    printf 'claude_base_system_prompt_sha256=%s\n' "$(printf '%s' "$claude_base_system_prompt" | sha256sum | awk '{print $1}')"
  fi
  printf 'label=%s\n' "$label"
  printf 'home=/work/home\n'
  printf 'cwd=/work/project\n'
  printf 'prompt=/prompt.md\n'
  if [[ "$condition" == "with_c3" ]]; then
    printf 'local_c3=/opt/c3/skills/c3/bin/c3x.sh\n'
    printf 'c3_uptake_gate=supervisor_transcript_exact_first_command\n'
  else
    printf 'local_c3=unmounted\n'
    printf 'c3_uptake_gate=unmounted\n'
  fi
  printf 'workspace=%s\n' "$workspace_dir"
  printf 'final=%s\n' "$final_file"
  printf 'stdout=%s\n' "$stdout_file"
  printf 'stderr=%s\n' "$stderr_file"
  printf 'prompt_archive=%s\n' "$prompt_archive"
  printf 'provenance=%s\n' "$provenance_file"
  printf 'mounted_repo_root=no\n'
  printf 'mounted_global_home=no\n'
  printf 'instruction_policy=sanitized_then_neutral_baseline\n'
  printf 'baseline_instruction_file_count=%s\n' "$baseline_instruction_file_count"
  printf 'unexpected_instruction_file_count=%s\n' "$unexpected_instruction_file_count"
  printf 'baseline_instruction_hash_match=%s\n' "$baseline_instruction_hash_match"
  printf 'baseline_instruction_sha256=%s\n' "$baseline_agents_sha256"
  printf 'baseline_c3_reference_count=%s\n' "$baseline_c3_reference_count"
  printf 'treatment_instruction_policy=%s\n' "$treatment_instruction_policy"
  printf 'treatment_instruction_sha256=%s\n' "$treatment_instruction_sha256"
  printf 'treatment_runtime_layer=%s\n' "$treatment_runtime_layer"
  printf 'ambient_instruction_policy=stripped_before_neutral_baseline\n'
  printf 'ambient_instruction_file_count=%s\n' "$ambient_instruction_file_count"
  printf 'ambient_agent_dir_count=%s\n' "$ambient_agent_dir_count"
  printf 'auth_source=%s\n' "$auth_source"
  printf 'run_timeout=%s\n' "$run_timeout"
  printf 'runtime_guard=process_supervisor\n'
  printf 'max_tool_calls=%s\n' "$max_tool_calls"
  printf 'max_tool_result_bytes=%s\n' "$max_tool_result_bytes"
  printf 'max_output_bytes=%s\n' "$max_output_bytes"
  if [[ "$agent" == "kilo" ]]; then
    printf 'kilo_agent=%s\n' "$kilo_agent"
  fi
  sed -n '1,$p' "$provenance_file"
} > "$meta_file"

setup_finished_ms="$(date +%s%3N)"
printf 'setup_elapsed_ms=%s\n' "$((setup_finished_ms - runner_started_ms))" >> "$meta_file"

supervisor_cmd=(
  python3 "$harness_root/bin/supervise-agent.py"
  --stdin "$prompt_file"
  --stdout "$stdout_file"
  --stderr "$stderr_file"
  --status "$guard_status_file"
  --max-seconds "$run_timeout"
  --max-tool-calls "$max_tool_calls"
  --max-tool-result-bytes "$max_tool_result_bytes"
  --max-output-bytes "$max_output_bytes"
  --
)

if [[ "$dry_run" -eq 1 ]]; then
  echo "Blindbox metadata:"
  sed -n '1,$p' "$meta_file"
  echo
  echo "Command:"
  quote_cmd_redacted "${supervisor_cmd[@]}" "${bwrap_args[@]}" "${agent_cmd[@]}"
  exit 0
fi

if [[ "$condition" == "with_c3" ]]; then
  c3_cache_prepare_file="$agent_output_dir/$label.c3-cache-prepare.txt"
  set +e
  (
    cd "$workspace_dir"
    C3X_MODE=agent C3X_LOCAL_BINARY="$C3_LOCAL_BINARY" \
      C3X_LOCAL_VERSION="$C3_LOCAL_VERSION" \
      bash "$repo_root/skills/c3/bin/c3x.sh" check
  ) > "$c3_cache_prepare_file" 2>&1
  c3_cache_prepare_status=$?
  set -e
  if [[ "$c3_cache_prepare_status" -ne 0 ]]; then
    printf 'c3_cache_prepare_status=failed\n' >> "$meta_file"
    cat "$c3_cache_prepare_file" >&2
    exit "$c3_cache_prepare_status"
  fi
  printf 'c3_cache_prepare_status=ready\n' >> "$meta_file"
fi

agent_started_ms="$(date +%s%3N)"
run_status=0
set +e
"${supervisor_cmd[@]}" "${bwrap_args[@]}" "${agent_cmd[@]}"
run_status=$?
case "$agent" in
  codex)
    if [[ "$run_status" -eq 0 && -s "$agent_final_file" ]]; then
      cp "$agent_final_file" "$final_file"
    elif [[ "$run_status" -eq 0 ]]; then
      cp "$stdout_file" "$final_file"
    else
      : > "$final_file"
    fi
    ;;
  claude)
    if [[ "$run_status" -eq 0 ]]; then
      jq -ser 'map(select(.type == "result" and .result != null)) | last | .result' "$stdout_file" > "$final_file"
      extract_status=$?
      if [[ "$extract_status" -ne 0 ]]; then
        run_status="$extract_status"
      fi
    else
      : > "$final_file"
    fi
    ;;
  kilo)
    jq -r 'select(.type == "text") | .part.text' "$stdout_file" > "$final_file"
    ;;
esac
set -e
agent_finished_ms="$(date +%s%3N)"
printf 'agent_elapsed_ms=%s\n' "$((agent_finished_ms - agent_started_ms))" >> "$meta_file"
printf 'runner_elapsed_ms=%s\n' "$((agent_finished_ms - runner_started_ms))" >> "$meta_file"
if [[ -s "$guard_status_file" ]]; then
  printf 'runtime_guard_status=%s\n' "$(jq -r '.status' "$guard_status_file")" >> "$meta_file"
  printf 'runtime_guard_reason=%s\n' "$(jq -r '.reason // "none"' "$guard_status_file")" >> "$meta_file"
  printf 'runtime_guard_tool_calls=%s\n' "$(jq -r '.tool_calls_observed' "$guard_status_file")" >> "$meta_file"
  printf 'runtime_guard_tool_result_bytes=%s\n' "$(jq -r '.tool_result_bytes_observed' "$guard_status_file")" >> "$meta_file"
  printf 'runtime_guard_output_bytes=%s\n' "$(jq -r '.output_bytes_observed' "$guard_status_file")" >> "$meta_file"
fi
c3_invocation_count=0
c3_route_command_count=0
c3_impact_command_count=0
c3_evidence_command_count=0
c3_success_count=0
c3_route_success_count=0
c3_impact_success_count=0
c3_evidence_success_count=0
if [[ "$condition" == "with_c3" && -s "$c3_usage_file" ]]; then
  c3_invocation_count="$(wc -l < "$c3_usage_file" | tr -d '[:space:]')"
  c3_route_command_count="$(awk -F '\t' '$1 == "route" { count++ } END { print count + 0 }' "$c3_usage_file")"
  c3_impact_command_count="$(awk -F '\t' '$1 == "impact" { count++ } END { print count + 0 }' "$c3_usage_file")"
  c3_evidence_command_count="$(awk -F '\t' '$1 == "evidence" { count++ } END { print count + 0 }' "$c3_usage_file")"
  c3_success_count="$(awk -F '\t' '$2 == "0" { count++ } END { print count + 0 }' "$c3_usage_file")"
  c3_route_success_count="$(awk -F '\t' '$1 == "route" && $2 == "0" { count++ } END { print count + 0 }' "$c3_usage_file")"
  c3_impact_success_count="$(awk -F '\t' '$1 == "impact" && $2 == "0" { count++ } END { print count + 0 }' "$c3_usage_file")"
  c3_evidence_success_count="$(awk -F '\t' '$1 == "evidence" && $2 == "0" { count++ } END { print count + 0 }' "$c3_usage_file")"
fi
printf 'c3_invocation_count=%s\n' "$c3_invocation_count" >> "$meta_file"
printf 'c3_route_command_count=%s\n' "$c3_route_command_count" >> "$meta_file"
printf 'c3_impact_command_count=%s\n' "$c3_impact_command_count" >> "$meta_file"
printf 'c3_evidence_command_count=%s\n' "$c3_evidence_command_count" >> "$meta_file"
printf 'c3_success_count=%s\n' "$c3_success_count" >> "$meta_file"
printf 'c3_route_success_count=%s\n' "$c3_route_success_count" >> "$meta_file"
printf 'c3_impact_success_count=%s\n' "$c3_impact_success_count" >> "$meta_file"
printf 'c3_evidence_success_count=%s\n' "$c3_evidence_success_count" >> "$meta_file"
rm -f "$c3_usage_file"
if [[ "$condition" == "with_c3" ]]; then
  rmdir "$c3_usage_dir" 2>/dev/null || true
fi

if [[ "$run_status" -ne 0 ]]; then
  exit "$run_status"
fi

echo "Wrote $final_file"
