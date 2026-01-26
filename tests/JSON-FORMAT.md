# Claude CLI JSON Output Format Reference

## Running Claude with JSON Output

```bash
claude -p "prompt" \
    --plugin-dir /path/to/plugin \
    --dangerously-skip-permissions \
    --max-turns N \
    --output-format stream-json \
    --verbose
```

**Required flags:**
- `--output-format stream-json` - JSON output mode
- `--verbose` - Required with stream-json (will error without it)

## Working Directory Detection

The `cd` approach correctly sets the working directory:

```bash
(cd /path/to/project && claude -p "..." ...)
```

The init message confirms: `"cwd":"/path/to/project"`

## Message Types

### Init Message
```json
{
  "type": "system",
  "subtype": "init",
  "cwd": "/path/to/project",
  "session_id": "...",
  "plugins": [...]
}
```

### Hook Messages
```json
{
  "type": "system",
  "subtype": "hook_started",
  "hook_name": "SessionStart:startup",
  "hook_event": "SessionStart"
}
```

### Assistant Messages with Tool Use
```json
{
  "type": "assistant",
  "message": {
    "content": [
      {"type": "text", "text": "..."},
      {"type": "tool_use", "name": "ToolName", "input": {...}}
    ]
  }
}
```

## Detecting Skill Invocation

**Direct Skill tool:**
```json
{
  "type": "assistant",
  "message": {
    "content": [{
      "type": "tool_use",
      "name": "Skill",
      "input": {
        "skill": "c3-skill:c3"
      }
    }]
  }
}
```

**Extract:** `content[].input.skill` where `name == "Skill"`

## Detecting Agent Dispatch

**Task tool with subagent_type:**
```json
{
  "type": "assistant",
  "message": {
    "content": [{
      "type": "tool_use",
      "name": "Task",
      "input": {
        "subagent_type": "c3-skill:c3-navigator",
        "prompt": "...",
        "description": "..."
      }
    }]
  }
}
```

**Extract:** `content[].input.subagent_type` where `name == "Task"`

## Result Message
```json
{
  "type": "result",
  "subtype": "success",
  "result": "The response text",
  "duration_ms": 7463,
  "num_turns": 2,
  "total_cost_usd": 0.111
}
```

## Test Assertions

### Check for Skill Invocation
```bash
# Using jq to extract skill name
cat output.json | jq -r '
  select(.type == "assistant") |
  .message.content[]? |
  select(.type == "tool_use" and .name == "Skill") |
  .input.skill
'
```

### Check for Agent Dispatch
```bash
# Using jq to extract agent type
cat output.json | jq -r '
  select(.type == "assistant") |
  .message.content[]? |
  select(.type == "tool_use" and .name == "Task") |
  .input.subagent_type
'
```

### Check Success
```bash
# Verify test completed successfully
cat output.json | jq -r 'select(.type == "result") | .subtype'
# Expected: "success"
```

## Grep-Based Assertions (No jq)

For simpler bash-only assertions:

```bash
# Check if skill was invoked
grep -q '"name":"Skill"' output.json && \
grep -q '"skill":"c3-skill:c3"' output.json

# Check if agent was dispatched
grep -q '"name":"Task"' output.json && \
grep -q '"subagent_type":"c3-skill:c3-navigator"' output.json

# Check for success
grep -q '"subtype":"success"' output.json
```
