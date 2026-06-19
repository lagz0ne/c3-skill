#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: harness/bin/run-blindbox.sh --agent codex|claude|kilo|both|kilo-free|all --topic <topic> [options]

Options:
  --auth env|session   Auth source. env passes whitelisted environment vars.
                       session copies known CLI session auth files into a
                       temporary auth directory mounted into the sandbox.
                       Default: env.
  --model <model>      Pass a model name to the selected agent.
  --kilo-agent <name>  Kilo agent to use for kilo runs. Default: code.
  --refresh-models     Refresh Kilo model metadata before resolving free models.
  --run-timeout <sec>  Timeout for each child run. Default: no timeout.
  --label <label>      Stable output label. With --agent both/all, this is a prefix.
  --dry-run            Print isolation metadata and redacted command only.
  -h, --help           Show this help.

The runner mounts no repository root. The writable project lives under
harness/runs/<label>.workspace and is mounted as /work/project.
USAGE
}

agent=""
topic=""
auth_mode="env"
model=""
kilo_agent="code"
refresh_models=0
run_timeout=0
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
    --auth)
      auth_mode="${2:-}"
      shift 2
      ;;
    --model)
      model="${2:-}"
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

if [[ -z "$agent" || -z "$topic" ]]; then
  usage >&2
  exit 2
fi

if [[ "$auth_mode" != "env" && "$auth_mode" != "session" ]]; then
  echo "Unsupported auth mode: $auth_mode" >&2
  exit 2
fi

if ! [[ "$run_timeout" =~ ^[0-9]+$ ]]; then
  echo "--run-timeout must be a non-negative integer number of seconds" >&2
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

kilo_free_models() {
  local refresh_arg=()
  [[ "$refresh_models" -eq 1 ]] && refresh_arg=(--refresh)
  kilo models --verbose "${refresh_arg[@]}" | awk '
    /^kilo\// { model=$0 }
    /"isFree": true/ { print model }
  '
}

timestamp="${HARNESS_RUN_ID:-$(date -u +%Y%m%dT%H%M%SZ)}"
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
harness_root="$(cd "$script_dir/.." && pwd)"
repo_root="$(cd "$harness_root/../../../.." && pwd)"
run_dir="$harness_root/runs"
mkdir -p "$run_dir"

if [[ "$agent" == "both" || "$agent" == "all" ]]; then
  args=(--topic "$topic" --auth "$auth_mode")
  [[ -n "$model" ]] && args+=(--model "$model")
  [[ "$dry_run" -eq 1 ]] && args+=(--dry-run)
  [[ "$run_timeout" -gt 0 ]] && args+=(--run-timeout "$run_timeout")
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
    [[ "$run_timeout" -gt 0 ]] && kilo_args+=(--run-timeout "$run_timeout")
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
    [[ "$run_timeout" -gt 0 ]] && child_args+=(--run-timeout "$run_timeout")
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
[[ "$run_timeout" -gt 0 ]] && need_for_run timeout
run_prefix=()
if [[ "$run_timeout" -gt 0 ]]; then
  run_prefix=(timeout "${run_timeout}s")
fi

prompt_file="$(mktemp)"
auth_root=""
trap 'rm -f "$prompt_file"; [[ -z "$auth_root" ]] || rm -rf "$auth_root"' EXIT
"$harness_root/bin/build-agent-prompt.sh" "$topic" > "$prompt_file"

if [[ -z "$label" ]]; then
  label="$timestamp-$agent-$topic"
fi

final_file="$run_dir/$label.md"
stdout_file="$run_dir/$label.stdout.txt"
stderr_file="$run_dir/$label.stderr.txt"
meta_file="$run_dir/$label.meta.txt"
workspace_dir="$run_dir/$label.workspace"

seed_workspace() {
  rm -rf "$workspace_dir"
  mkdir -p "$workspace_dir"
  cat > "$workspace_dir/README.md" <<'README'
# Blindbox C3 Growth Project

This is an isolated eval project. Use the mounted local C3 wrapper only:

```bash
C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh <command>
```
README
}

seed_workspace

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
  --dir /opt/c3
  --dir /opt/c3/skills
  --bind "$workspace_dir" /work/project
  --bind "$run_dir" /runs
  --ro-bind "$prompt_file" /prompt.md
  --ro-bind "$repo_root/skills/c3" /opt/c3/skills/c3
  --chdir /work/project
  --setenv HOME /work/home
  --setenv USER blindbox
  --setenv LOGNAME blindbox
  --setenv XDG_CACHE_HOME /work/home/.cache
  --setenv XDG_CONFIG_HOME /work/home/.config
  --setenv XDG_DATA_HOME /work/home/.local/share
  --setenv CODEX_HOME /work/home/.codex
  --setenv CLAUDE_CONFIG_DIR /work/home/.claude
  --setenv CLAUDE_CODE_SAFE_MODE 1
  --setenv PATH /opt/node/bin:/opt/claude:/opt/kilo:/usr/local/bin:/usr/bin:/bin
  --setenv NO_COLOR 1
)

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
      agent_cmd=("$codex_cmd" exec --ask-for-approval never --sandbox workspace-write --skip-git-repo-check --ignore-user-config --cd /work/project --output-last-message "/runs/$label.md")
    else
      codex_js="$(readlink -f "$codex_cmd")"
      node_bin="$(readlink -f "$node_cmd")"
      node_root="$(cd "$(dirname "$node_bin")/.." && pwd)"
      codex_js_sandbox="/opt/node/${codex_js#"$node_root"/}"
      bwrap_args+=(--ro-bind "$node_root" /opt/node)
      agent_cmd=(/opt/node/bin/node "$codex_js_sandbox"
        --ask-for-approval never
        exec
        --skip-git-repo-check
        --ephemeral
        --ignore-user-config
        --sandbox workspace-write
        --cd /work/project
        --output-last-message "/runs/$label.md"
      )
    fi
    [[ -n "$model" ]] && agent_cmd+=(--model "$model")
    agent_cmd+=(-)
    ;;
  claude)
    need_for_run claude
    claude_cmd="$(resolve_cmd claude)"
    if [[ "$claude_cmd" == /missing/* ]]; then
      agent_cmd=("$claude_cmd" --print)
    else
      claude_bin="$(readlink -f "$claude_cmd")"
      bwrap_args+=(--dir /opt/claude --ro-bind "$claude_bin" /opt/claude/claude)
      # NOTE: do NOT pass --bare. --bare only accepts an ANTHROPIC_API_KEY and
      # rejects the OAuth/subscription token (CLAUDE_CODE_OAUTH_TOKEN) we stage,
      # surfacing as "Not logged in". Plain --print honors the OAuth token.
      agent_cmd=(/opt/claude/claude
        --print
        --safe-mode
        --no-session-persistence
        --permission-mode bypassPermissions
        --allowedTools "Bash,Read,Write,Edit,MultiEdit,Glob,Grep,LS"
        --output-format text
      )
    fi
    [[ -n "$model" ]] && agent_cmd+=(--model "$model")
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
  printf 'topic=%s\n' "$topic"
  printf 'auth_mode=%s\n' "$auth_mode"
  printf 'model=%s\n' "${model:-default}"
  printf 'label=%s\n' "$label"
  printf 'home=/work/home\n'
  printf 'cwd=/work/project\n'
  printf 'prompt=/prompt.md\n'
  printf 'local_c3=/opt/c3/skills/c3/bin/c3x.sh\n'
  printf 'workspace=%s\n' "$workspace_dir"
  printf 'final=%s\n' "$final_file"
  printf 'stdout=%s\n' "$stdout_file"
  printf 'stderr=%s\n' "$stderr_file"
  printf 'mounted_repo_root=no\n'
  printf 'mounted_global_home=no\n'
  printf 'auth_source=%s\n' "$auth_source"
  printf 'run_timeout=%s\n' "$run_timeout"
  if [[ "$agent" == "kilo" ]]; then
    printf 'kilo_agent=%s\n' "$kilo_agent"
  fi
} > "$meta_file"

if [[ "$dry_run" -eq 1 ]]; then
  echo "Blindbox metadata:"
  sed -n '1,$p' "$meta_file"
  echo
  echo "Command:"
  quote_cmd_redacted "${run_prefix[@]}" "${bwrap_args[@]}" "${agent_cmd[@]}"
  exit 0
fi

case "$agent" in
  codex)
    "${run_prefix[@]}" "${bwrap_args[@]}" "${agent_cmd[@]}" < "$prompt_file" > "$stdout_file" 2> "$stderr_file"
    if [[ ! -s "$final_file" ]]; then
      cp "$stdout_file" "$final_file"
    fi
    ;;
  claude)
    "${run_prefix[@]}" "${bwrap_args[@]}" "${agent_cmd[@]}" < "$prompt_file" > "$final_file" 2> "$stderr_file"
    cp "$final_file" "$stdout_file"
    ;;
  kilo)
    "${run_prefix[@]}" "${bwrap_args[@]}" "${agent_cmd[@]}" > "$stdout_file" 2> "$stderr_file"
    jq -r 'select(.type == "text") | .part.text' "$stdout_file" > "$final_file"
    ;;
esac

echo "Wrote $final_file"
