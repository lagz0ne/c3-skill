#!/bin/bash
# Logs all tool access for audit trail
# Receives hook input via stdin as JSON with tool_name and tool_input

AUDIT_LOG="${C3_AUDIT_LOG:-/tmp/c3-audit.log}"

# Read hook input from stdin
INPUT=$(cat)

# Timestamp
TS=$(date '+%Y-%m-%d %H:%M:%S')

# Extract tool_name and tool_input from hook JSON
TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // "unknown"')
TOOL_INPUT=$(echo "$INPUT" | jq -c '.tool_input // {}')

case "$TOOL_NAME" in
    Read)
        FILE_PATH=$(echo "$TOOL_INPUT" | jq -r '.file_path // "unknown"')
        echo "[$TS] READ: $FILE_PATH" >> "$AUDIT_LOG"
        ;;
    Write)
        FILE_PATH=$(echo "$TOOL_INPUT" | jq -r '.file_path // "unknown"')
        echo "[$TS] WRITE: $FILE_PATH" >> "$AUDIT_LOG"
        ;;
    Edit)
        FILE_PATH=$(echo "$TOOL_INPUT" | jq -r '.file_path // "unknown"')
        echo "[$TS] EDIT: $FILE_PATH" >> "$AUDIT_LOG"
        ;;
    Glob)
        PATTERN=$(echo "$TOOL_INPUT" | jq -r '.pattern // "unknown"')
        PATH_ARG=$(echo "$TOOL_INPUT" | jq -r '.path // "."')
        echo "[$TS] GLOB: $PATTERN in $PATH_ARG" >> "$AUDIT_LOG"
        ;;
    Grep)
        PATTERN=$(echo "$TOOL_INPUT" | jq -r '.pattern // "unknown"')
        PATH_ARG=$(echo "$TOOL_INPUT" | jq -r '.path // "."')
        echo "[$TS] GREP: '$PATTERN' in $PATH_ARG" >> "$AUDIT_LOG"
        ;;
    Bash)
        COMMAND=$(echo "$TOOL_INPUT" | jq -r '.command // "unknown"' | head -c 200)
        echo "[$TS] BASH: $COMMAND" >> "$AUDIT_LOG"
        ;;
    Task)
        DESC=$(echo "$TOOL_INPUT" | jq -r '.description // "unknown"')
        echo "[$TS] TASK: $DESC" >> "$AUDIT_LOG"
        ;;
    Skill)
        SKILL=$(echo "$TOOL_INPUT" | jq -r '.skill // "unknown"')
        echo "[$TS] SKILL: $SKILL" >> "$AUDIT_LOG"
        ;;
    *)
        echo "[$TS] $TOOL_NAME: $TOOL_INPUT" >> "$AUDIT_LOG"
        ;;
esac

# Always allow - this is just logging
exit 0
