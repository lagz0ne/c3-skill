---
name: c3-adopt
description: Use when bootstrapping C3 documentation for any project - guides through Socratic discovery and delegates to layer skills for document creation
---

# C3 Adopt

## Overview

Bootstrap C3 architecture documentation through Socratic questioning and delegation. Works for existing codebases and fresh projects.

**Announce:** "I'm using the c3-adopt skill to initialize architecture documentation."

---

## 📁 V3 STRUCTURE ENFORCEMENT (MANDATORY)

**This is non-negotiable. C3 v3 uses a specific hierarchical file structure.**

### Required Structure (V3)

```
.c3/
├── README.md                     ← Context (c3-0) IS the README
├── c3-1-{slug}/
│   ├── README.md                 ← Container 1
│   └── c3-101-{slug}.md          ← Component inside container
├── c3-2-{slug}/
│   └── README.md                 ← Container 2
└── adr/
    └── adr-YYYYMMDD-{slug}.md    ← ADRs
```

### Prohibited Patterns (V2 - DO NOT USE)

| Pattern | Why Wrong |
|---------|-----------|
| `context/c3-0.md` | Context IS `.c3/README.md`, not a separate folder |
| `containers/c3-1.md` | Containers are folders: `.c3/c3-1-{slug}/README.md` |
| `components/c3-101.md` | Components live INSIDE their container folder |
| Any `c3-0.md` file | Context is always `README.md` at `.c3/` root |

### File Path Rules

| Level | Path | Example |
|-------|------|---------|
| Context | `.c3/README.md` | `.c3/README.md` (ONLY this) |
| Container | `.c3/c3-{N}-{slug}/README.md` | `.c3/c3-1-backend/README.md` |
| Component | `.c3/c3-{N}-{slug}/c3-{N}{NN}-{slug}.md` | `.c3/c3-1-backend/c3-101-api.md` |
| ADR | `.c3/adr/adr-{YYYYMMDD}-{slug}.md` | `.c3/adr/adr-20251216-auth.md` |

### Red Flags - STOP and Fix If You See

🚩 Creating a `context/` folder
🚩 Creating a `containers/` folder
🚩 Creating a `components/` folder
🚩 Any file named `c3-0.md` (should be `README.md`)
🚩 Component files outside their parent container folder
🚩 Container as a single file instead of a folder with README.md

### Validation Checklist (RUN BEFORE COMPLETING)

- [ ] Context is `.c3/README.md` (not in a subfolder)
- [ ] Each container is a folder: `.c3/c3-{N}-{slug}/`
- [ ] Each container has `README.md` inside its folder
- [ ] Components are inside their container folder
- [ ] No `context/`, `containers/`, or `components/` folders exist

---

## When to Use

| Scenario | Path |
|----------|------|
| Existing codebase needs C3 docs | Explore → Question → Document |
| New project, no code yet | Question → Scaffold → Document |
| `.c3/` exists but needs rebuild | Ask: update, backup+recreate, or abort |

## Process

| Phase | Actions | Delegates To |
|-------|---------|--------------|
| 1. Detect | Check for `.c3/`, existing code | - |
| 2. Discover | Socratic questions per `references/discovery-questions.md` | - |
| 3. Context | Document containers, actors, protocols | c3-context-design |
| 4. Containers | For each container: components, tech, patterns | c3-container-design |
| 5. Components | For key components: flows, dependencies | c3-component-design |
| 6. Settings | Configure if missing | c3-config |

## Socratic Discovery

Use questions from `references/discovery-questions.md`:
- Start broad: "What problem does this system solve?"
- Identify containers: "What are the major deployable parts?"
- Map relationships: "How do containers communicate?"

**For existing codebases:** Explore first, then refine with questions.
**For fresh projects:** Questions only, no code exploration.

## Delegation Pattern

After discovery, delegate to layer skills in order:
1. c3-context-design → Creates c3-0 (Context)
2. c3-container-design → Creates c3-N for each container
3. c3-component-design → Creates c3-NNN for key components
4. c3-config → Creates settings.yaml if missing

---

## ⛔ SKILL DELEGATION ENFORCEMENT (MANDATORY)

**Rule:** When work requires a layer skill, INVOKE it. Never describe what it "would do."

### Anti-Patterns

| Pattern | Why It's Wrong | Correct Action |
|---------|----------------|----------------|
| "c3-context-design would create..." | Hallucinating skill behavior | Use Skill tool → c3-context-design |
| "Following c3-container-design template..." | You don't have its template loaded | Invoke the skill first |
| "This is simple, I'll create docs directly" | Layer skills have guardrails you're bypassing | Always invoke for layer work |
| Summarizing a skill's output without invoking | Fabrication | Invoke, get real output |

### Red Flags

🚩 Using layer skill name as a noun ("the c3-context-design approach")
🚩 Describing layer skill output without a Skill tool call in the conversation
🚩 "I'll apply c3-X-design principles" without invoking it
🚩 Creating `.c3/` files without invoking the appropriate layer skill

### Self-Check

- [ ] Did I use the Skill tool for each layer I'm creating docs for?
- [ ] Am I quoting actual skill output, not imagined output?
- [ ] Is there a Skill tool invocation for context, container, and component work?

### Escape Hatch

None. Layer work = layer skill invocation. No exceptions.

---

## ⛔ OUTPUT VERIFICATION ENFORCEMENT (MANDATORY)

**Rule:** Claiming completion requires verification evidence in the conversation.

### Verification Requirements

| Claim | Required Evidence |
|-------|-------------------|
| "Created .c3/ structure" | `mkdir` commands + `ls .c3/` showing structure |
| "Context doc created" | Write command to `.c3/README.md` visible |
| "Container docs created" | Write commands to `.c3/c3-{N}-*/README.md` visible |
| "Delegated to skill X" | Skill tool invocation visible |
| "V3 structure validated" | Validation checklist executed with results |

### Anti-Patterns

| Pattern | Why It's Wrong | Correct Action |
|---------|----------------|----------------|
| "I've scaffolded the .c3 structure" (no file ops visible) | No evidence of creation | Show the mkdir/write commands |
| "Adoption complete" (no validation) | Structure errors are common | Run validation checklist |
| "Delegation complete" (no skill invocation) | Hallucination | Show Skill tool usage |

### Red Flags

🚩 Completion claim without corresponding tool usage
🚩 "Done" without running V3 structure validation checklist
🚩 Describing artifacts that weren't created in this conversation

### Self-Check

- [ ] For each file I claim exists, is there evidence of its creation?
- [ ] Did I run the V3 STRUCTURE ENFORCEMENT validation checklist?
- [ ] Can a reviewer see proof in this conversation?

### Escape Hatch

None. Unverified completion = not complete.

## Scaffolding

**Create V3 structure before delegation:**

```bash
# Create base structure
mkdir -p .c3/adr

# Create container folders (NOT a containers/ folder!)
mkdir -p .c3/c3-1-{slug}
mkdir -p .c3/c3-2-{slug}
# ... for each container

# Context goes in README.md at root (NOT context/c3-0.md!)
touch .c3/README.md
```

**Remember:**
- Context → `.c3/README.md` (the root README IS the context)
- Container → `.c3/c3-{N}-{slug}/README.md` (folder with README inside)
- Component → `.c3/c3-{N}-{slug}/c3-{N}{NN}-{slug}.md` (inside container folder)

Generate TOC after documentation is complete.

## Checklist

- [ ] Scenario detected (existing vs fresh)
- [ ] Discovery questions completed
- [ ] **📁 V3 structure scaffolded** (no context/, containers/, components/ folders)
- [ ] Context created as `.c3/README.md`
- [ ] Containers created as `.c3/c3-{N}-{slug}/README.md`
- [ ] Components created inside their container folders
- [ ] Settings configured
- [ ] TOC generated
- [ ] **📁 Structure validation passed** (see V3 STRUCTURE ENFORCEMENT)

## Platform Docs (Recommended)

Platform concerns (deployment, networking, secrets, CI/CD) reduce onboarding friction. Ask user if they want to document:

```
.c3/platform/
├── deployment.md      # c3-0-deployment
├── networking.md      # c3-0-networking
├── secrets.md         # c3-0-secrets
└── ci-cd.md           # c3-0-cicd
```

Use `c3-0-*` IDs for Context-level platform docs.

## Related

- `references/discovery-questions.md` - Socratic question templates
- `references/v3-structure.md` - File structure conventions
- `references/archetype-hints.md` - Container type patterns
- `references/derivation-guardrails.md` - Hierarchy rules
