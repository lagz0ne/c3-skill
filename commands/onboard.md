---
description: Initialize C3 architecture documentation for this project
argument-hint: [--force]
allowed-tools: Bash(rm:*), Bash(test:*), Bash(PROJECT=*), Read, Glob, Grep, Write, Edit, AskUserQuestion, Skill
---

# C3 Onboarding

## Pre-flight Check

- .c3 directory exists: !`test -d "$CLAUDE_PROJECT_DIR/.c3" && echo "yes" || echo "no"`
- Arguments: $ARGUMENTS

**If .c3 exists AND --force NOT passed:**
```
C3 architecture documentation already exists.
To start fresh: /onboard --force
```

**If .c3 exists AND --force passed:**
```bash
rm -rf "$CLAUDE_PROJECT_DIR/.c3"
```

## Instructions

Invoke the `onboard` skill to handle the full onboarding workflow.

Use the Skill tool with skill: "c3-skill:onboard"

The skill implements the staged recursive learning loop:
1. Context level (system boundary)
2. Container level (deployable units)
3. Component level (Foundation then Feature)
4. Refs discovery (shared patterns)
