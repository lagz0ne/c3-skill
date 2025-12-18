# Agent-Gated C3 Adoption

## Problem

C3 adoption suffers from three friction points:

1. **Discovery** - Users don't know C3 exists or when to use it
2. **Activation** - Users know about C3 but forget/don't invoke it
3. **Tunnel vision** - When deep in code, agents ignore existing architectural docs

Current approaches (SessionStart hooks, CLAUDE.md templates, enforcement harnesses) require conscious user action. They don't solve the fundamental problem: **C3 must be invoked before it can help**.

## Solution

**Hook-based gating with AI-inferred context.**

A PreToolUse hook blocks Edit/Write operations until C3 context has been loaded and an ADR exists. The hook provides feedback that guides Claude to run c3-design, which creates the ADR.

### Design Principles

- **No escape** - All edits require C3 consideration
- **Lite view** - Block message shows what C3 can offer (value pitch)
- **AI inference** - Agent figures out which container/components are relevant (no config)
- **ADR-before-code** - Every meaningful change gets documented

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    PreToolUse Hook                      │
│                    (Edit|Write)                         │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Input: tool_input.file_path                            │
│                                                         │
│  ┌─────────────────────────────────────────────────┐   │
│  │ Check 1: Does .c3/ exist?                       │   │
│  │   No → BLOCK: "No C3 docs. Run /c3-adopt"       │   │
│  └─────────────────────────────────────────────────┘   │
│                         │                               │
│                        Yes                              │
│                         ▼                               │
│  ┌─────────────────────────────────────────────────┐   │
│  │ Check 2: Does ADR exist for today's session?    │   │
│  │   No → BLOCK: "Run c3-design first"             │   │
│  │         Include: file path, inferred container  │   │
│  └─────────────────────────────────────────────────┘   │
│                         │                               │
│                        Yes                              │
│                         ▼                               │
│  ┌─────────────────────────────────────────────────┐   │
│  │ ALLOW: Edit proceeds                            │   │
│  └─────────────────────────────────────────────────┘   │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### Block Message Format (Lite View)

When no .c3/ exists:
```
C3 Architecture Gate

No architecture documentation found (.c3/ directory missing).

C3 helps you:
• Document system architecture at Context/Container/Component levels
• Track architectural decisions with ADRs
• Ensure code changes align with system design

To initialize: Use the c3-adopt skill

Blocking: {tool_name} on {file_path}
```

When .c3/ exists but no ADR:
```
C3 Architecture Gate

You're about to edit: {file_path}

This appears to touch: {inferred_container} (inferred from path)

Before editing code, c3-design will help you:
• Load relevant architectural context
• Identify related components
• Create an ADR documenting this change

Run c3-design to proceed.

Blocking: {tool_name} on {file_path}
```

### Flow

1. User asks Claude to make a code change
2. Claude attempts Edit/Write
3. Hook blocks, returns lite view message
4. Claude sees message, understands it needs c3-design
5. Claude invokes c3-design skill
6. c3-design loads context, creates ADR
7. Claude retries Edit/Write
8. Hook sees ADR exists, allows edit
9. Claude makes the edit with architectural context loaded

## Implementation

### Files to Create

```
c3-design/
├── hooks/
│   └── hooks.json              # Hook configuration
├── scripts/
│   └── c3-gate.py              # Gate logic
└── skills/
    └── c3-lite/                # Optional: dedicated lite view skill
        └── SKILL.md
```

### Hook Configuration (hooks/hooks.json)

```json
{
  "description": "C3 architecture gate - requires ADR before code changes",
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write|NotebookEdit",
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PLUGIN_ROOT}/scripts/c3-gate.py",
            "timeout": 10
          }
        ]
      }
    ]
  }
}
```

### Gate Script (scripts/c3-gate.py)

```python
#!/usr/bin/env python3
"""
C3 Architecture Gate

Blocks Edit/Write operations until:
1. .c3/ directory exists (C3 initialized)
2. ADR exists for current session/day (architectural review done)

Returns block message with lite view to guide Claude.
"""
import json
import sys
import os
from datetime import date
from pathlib import Path

def get_project_dir():
    return os.environ.get('CLAUDE_PROJECT_DIR', os.getcwd())

def c3_exists(project_dir):
    return (Path(project_dir) / '.c3').is_dir()

def get_todays_adrs(project_dir):
    """Find ADRs created today."""
    adr_dir = Path(project_dir) / '.c3' / 'adr'
    if not adr_dir.exists():
        return []

    today = date.today().strftime('%Y%m%d')
    return list(adr_dir.glob(f'adr-{today}-*.md'))

def infer_container(file_path, project_dir):
    """Best-effort container inference from file path."""
    # This is a simple heuristic - could be enhanced
    rel_path = os.path.relpath(file_path, project_dir)
    parts = rel_path.split(os.sep)

    # Common patterns
    if 'src' in parts:
        idx = parts.index('src')
        if len(parts) > idx + 1:
            return parts[idx + 1]  # e.g., src/api -> api

    if parts:
        return parts[0]

    return "unknown"

def block_no_c3(file_path, tool_name):
    """Return block message when .c3/ doesn't exist."""
    msg = f"""C3 Architecture Gate

No architecture documentation found (.c3/ directory missing).

C3 helps you:
- Document system architecture at Context/Container/Component levels
- Track architectural decisions with ADRs
- Ensure code changes align with system design

To initialize: Use the c3-adopt skill

Blocking: {tool_name} on {file_path}"""

    return msg

def block_no_adr(file_path, tool_name, container):
    """Return block message when no ADR exists."""
    msg = f"""C3 Architecture Gate

You're about to edit: {file_path}

This appears to touch: {container} (inferred from path)

Before editing code, c3-design will help you:
- Load relevant architectural context
- Identify related components
- Create an ADR documenting this change

Run c3-design to proceed.

Blocking: {tool_name} on {file_path}"""

    return msg

def main():
    try:
        input_data = json.load(sys.stdin)
    except json.JSONDecodeError:
        sys.exit(0)  # Allow on parse error

    tool_name = input_data.get('tool_name', '')
    tool_input = input_data.get('tool_input', {})
    file_path = tool_input.get('file_path', '')

    if not file_path:
        sys.exit(0)  # Allow if no file path

    project_dir = get_project_dir()

    # Check 1: Does .c3/ exist?
    if not c3_exists(project_dir):
        print(block_no_c3(file_path, tool_name), file=sys.stderr)
        sys.exit(2)  # Block

    # Check 2: Does today's ADR exist?
    adrs = get_todays_adrs(project_dir)
    if not adrs:
        container = infer_container(file_path, project_dir)
        print(block_no_adr(file_path, tool_name, container), file=sys.stderr)
        sys.exit(2)  # Block

    # All checks passed
    sys.exit(0)  # Allow

if __name__ == '__main__':
    main()
```

### Files to Remove

- `.claude/hooks/` - SessionStart hook (replaced by gate)
- CLAUDE.md template references (no longer needed for adoption)

## ADR Session Tracking

The gate uses "today's date" as a simple session marker:
- ADR filename: `adr-YYYYMMDD-slug.md`
- Gate checks: any ADR with today's date prefix

This means:
- First edit of the day triggers c3-design
- Subsequent edits same day proceed (ADR exists)
- Next day, new ADR required

**Alternative approaches** (future enhancement):
- Session ID tracking (more precise)
- ADR-per-file tracking (more granular)
- ADR scope detection (smarter matching)

## Migration

### What Changes for Users

1. **New behavior**: Edit/Write blocked until ADR exists
2. **Removed**: SessionStart auto-activation hook
3. **Removed**: CLAUDE.md template requirement

### Cleanup Tasks

1. Remove `hooks/session-start/` directory
2. Update plugin.json if needed
3. Update README to document new gating behavior

## Open Questions

1. **Allowlist paths?** Should some files bypass the gate (README, package.json)?
   - Current decision: No escape. Reconsider based on user feedback.

2. **ADR granularity?** One ADR per day vs per-feature vs per-file?
   - Current decision: Per-day. Simple and encourages batching related changes.

3. **Container inference accuracy?** How to improve without config?
   - Current decision: Best-effort heuristic. Can read .c3/TOC.md for hints.

## Success Metrics

- C3 skill invocations increase
- ADR creation rate increases
- Users report less "forgot to check architecture" moments
- Architectural documentation stays current with code

## Next Steps

1. Implement hook and gate script
2. Test on this repository (c3-design)
3. Remove old SessionStart hook
4. Update documentation
5. Release and gather feedback
