#!/bin/bash
# Analyze audit log for test compliance
# Categorizes file access and flags violations

set -e

usage() {
    echo "Usage: $0 <audit-log> [--strict]"
    echo ""
    echo "Options:"
    echo "  --strict   Fail if any access outside allowed paths"
    exit 1
}

AUDIT_LOG="$1"
STRICT=0

[[ -z "$AUDIT_LOG" ]] && usage
[[ ! -f "$AUDIT_LOG" ]] && echo "Audit log not found: $AUDIT_LOG" && exit 1

shift
while [[ $# -gt 0 ]]; do
    case $1 in
        --strict) STRICT=1; shift ;;
        *) usage ;;
    esac
done

# Extract workspace and plugin root from log header
WORKSPACE=$(grep "^# Workspace:" "$AUDIT_LOG" | cut -d' ' -f3-)
PLUGIN_ROOT=$(dirname "$(dirname "$(dirname "$AUDIT_LOG")")")

echo "================================"
echo "Audit Analysis"
echo "================================"
echo "Log: $AUDIT_LOG"
echo "Workspace: $WORKSPACE"
echo "Plugin: $PLUGIN_ROOT"
echo ""

# Categorize accesses
echo "## Access Categories"
echo ""

echo "### Allowed: Workspace files"
grep -E '(READ|WRITE|EDIT):' "$AUDIT_LOG" | grep "$WORKSPACE" | grep -v ".claude" || echo "(none)"
echo ""

echo "### Allowed: Plugin skill files"
grep -E 'READ:' "$AUDIT_LOG" | grep "$PLUGIN_ROOT" | grep -E '(skills|agents|references)' || echo "(none)"
echo ""

echo "### Allowed: Plugin hooks"
grep -E 'READ:' "$AUDIT_LOG" | grep "$PLUGIN_ROOT" | grep "hooks" || echo "(none)"
echo ""

echo "### Search operations (Glob/Grep)"
grep -E '(GLOB|GREP):' "$AUDIT_LOG" || echo "(none)"
echo ""

# Flag suspicious accesses
echo "## Potential Violations"
echo ""

VIOLATIONS=0

# Files outside workspace and plugin
echo "### Accessed outside allowed paths:"
OUTSIDE=$(grep -E 'READ:' "$AUDIT_LOG" | grep -v "$WORKSPACE" | grep -v "$PLUGIN_ROOT" | grep -v "^#" || true)
if [[ -n "$OUTSIDE" ]]; then
    echo "$OUTSIDE"
    VIOLATIONS=$((VIOLATIONS + $(echo "$OUTSIDE" | wc -l)))
else
    echo "(none)"
fi
echo ""

# User home directory access (except plugin)
echo "### Home directory access:"
HOME_ACCESS=$(grep -E 'READ:.*(/home/|/Users/)' "$AUDIT_LOG" | grep -v "$PLUGIN_ROOT" || true)
if [[ -n "$HOME_ACCESS" ]]; then
    echo "$HOME_ACCESS"
    VIOLATIONS=$((VIOLATIONS + $(echo "$HOME_ACCESS" | wc -l)))
else
    echo "(none)"
fi
echo ""

# System file access
echo "### System file access:"
SYS_ACCESS=$(grep -E 'READ:.*(^/etc/|^/usr/|^/var/)' "$AUDIT_LOG" || true)
if [[ -n "$SYS_ACCESS" ]]; then
    echo "$SYS_ACCESS"
    VIOLATIONS=$((VIOLATIONS + $(echo "$SYS_ACCESS" | wc -l)))
else
    echo "(none)"
fi
echo ""

# Summary
echo "## Summary"
echo ""
echo "Total READ operations: $(grep -c 'READ:' "$AUDIT_LOG" || echo 0)"
echo "Total WRITE operations: $(grep -c 'WRITE:' "$AUDIT_LOG" || echo 0)"
echo "Total EDIT operations: $(grep -c 'EDIT:' "$AUDIT_LOG" || echo 0)"
echo "Total BASH operations: $(grep -c 'BASH:' "$AUDIT_LOG" || echo 0)"
echo ""
echo "Potential violations: $VIOLATIONS"
echo ""

if [[ "$STRICT" == "1" && "$VIOLATIONS" -gt 0 ]]; then
    echo "STRICT MODE: Test FAILED due to $VIOLATIONS violation(s)"
    exit 1
fi

if [[ "$VIOLATIONS" -gt 0 ]]; then
    echo "WARNING: Review violations above"
    exit 0
else
    echo "OK: All accesses within allowed paths"
    exit 0
fi
