---
description: Initialize C3 architecture documentation for this project
argument-hint: [--force]
allowed-tools: Bash(rm:*), Bash(test:*), Bash(PROJECT=*), Task, AskUserQuestion
---

# C3 Onboarding

## Current State

- .c3 directory exists: !`test -d "$CLAUDE_PROJECT_DIR/.c3" && echo "yes" || echo "no"`
- Arguments: $ARGUMENTS

## Instructions

You are initializing C3 architecture documentation for this project.

### Step 1: Check Existing State

**If .c3 already exists AND --force was NOT passed:**

Stop and inform the user:
```
C3 architecture documentation already exists in this project.

To start fresh, run: /onboard --force

This will remove the existing .c3 directory and create new documentation.
```

**If .c3 already exists AND --force WAS passed:**

Remove the existing .c3 directory:
```bash
rm -rf "$CLAUDE_PROJECT_DIR/.c3"
```

### Step 2: Gather Project Info

Use AskUserQuestion to collect:

1. **PROJECT**: Project name (e.g., "MyApp", "Acme Platform")
2. **Containers**: The main architectural containers (e.g., "backend", "frontend", "api", "worker")
   - Ask: "What are the main containers/modules in your project?"
   - Examples: backend/frontend, api/web/worker, core/plugins

### Step 3: Initialize Structure

Run the c3-init.sh script with the gathered info:

```bash
PROJECT="<project_name>" C1="<container1>" C2="<container2>" ... "${CLAUDE_PLUGIN_ROOT}/scripts/c3-init.sh"
```

This creates:
```
.c3/
├── README.md                       (Context - c3-0)
├── c3-1-<container1>/README.md     (Container - c3-1)
├── c3-2-<container2>/README.md     (Container - c3-2)
└── adr/adr-00000000-c3-adoption.md (Adoption ADR)
```

### Step 4: Fill Templates with AI

Dispatch a subagent (Task tool, subagent_type: "Explore") to analyze codebase and fill templates:

**Subagent prompt:**
```
You are filling C3 templates for <PROJECT>.

Templates are already in place at .c3/. Your job:
1. Analyze codebase structure
2. Fill inventory tables with actual components
3. Create mermaid diagrams with IDs
4. Add linkages with reasoning

## Context (c3-0 - .c3/README.md):
1. Identify actors (users, schedulers, webhooks, external triggers)
2. Confirm containers match code structure
3. Identify external systems (databases, APIs, caches, queues)
4. Draw mermaid diagram with IDs (A1 for actors, c3-N for containers, E1 for externals)
5. Fill linkages with REASONING (why they connect, not just that they do)

## Each Container (c3-N - .c3/c3-N-*/README.md):
1. Analyze container scope for components
2. Categorize by concern:
   - Foundation: entry points, identity, integration
   - Auxiliary: library wrappers, framework usage, utilities
   - Business: domain services, business logic
   - Presentation: styling, composition, state (or N/A if not applicable)
3. Draw internal mermaid diagram with component IDs
4. Fill fulfillment section (which components handle Context links)
5. Fill linkages with REASONING

## ADR-000 (.c3/adr/adr-00000000-c3-adoption.md):
1. Document why C3 was adopted
2. List all containers created
3. Mark verification checklist items

Rules:
- Diagram goes FIRST in each file, uses IDs from inventory tables
- Every linkage needs REASONING (why, not just that)
- Foundation/Auxiliary/Business/Presentation categories for components
- Fulfillment section maps Context links to specific components
- Keep template structure, fill in actual content
- Empty sections are OK if category doesn't apply
```

### Step 5: Confirm Completion

After subagent completes, summarize what was created and suggest next steps:
- Review the generated documentation
- Run `/c3 audit` to verify consistency
- Use `/c3` for ongoing architecture work
