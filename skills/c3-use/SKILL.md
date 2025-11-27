---
name: c3-use
description: Entry point for using C3 architecture docs - checks for .c3/ directory, guides reading if exists, offers adoption if not
---

# C3 Use - Smart Architecture Navigation

## Overview

Entry point for working with C3 architecture documentation. Intelligently discovers and presents the right content based on what exists and what the user needs.

**Announce at start:** "I'm using the c3-use skill to help you navigate the architecture documentation."

---

## Phase 1: Discovery via TOC

<extended_thinking>
<goal>Use TOC as the single source of truth for architecture inventory</goal>

<why_toc>
The TOC already contains:
- All document IDs with titles and summaries
- Document counts (Context, Containers, Components, ADRs)
- Hierarchical relationships
- Section headings within each document

No need to traverse the filesystem - TOC has already done that work.
</why_toc>

<discovery_flow>
1. Check if .c3/ exists
2. Read TOC.md (one file, complete inventory)
3. Parse the Quick Reference section for stats
4. Parse document entries for navigation map
5. Decide what to present based on user intent
</discovery_flow>
</extended_thinking>

### Step 1.1: Check for .c3/ Directory

```bash
ls -d .c3 2>/dev/null && echo "EXISTS" || echo "MISSING"
```

**If MISSING:** Jump to [No C3 Found](#no-c3-found)

**If EXISTS:** Continue to Step 1.2

### Step 1.2: Load TOC for Complete Inventory

```bash
# Read the TOC - this is the architecture inventory
cat .c3/TOC.md
```

<extended_thinking>
<goal>Extract architecture inventory from TOC</goal>

<toc_structure>
The TOC has:
1. **Quick Reference** - Stats section with document counts
2. **Document entries** - Each document with:
   - ID and title: `### [c3-1] Backend API Container`
   - Summary: One-line description
   - Sections: Heading IDs for deeper navigation

Parse this to build:
```
Architecture Inventory:
├── Stats: {containers}C / {components}c / {adrs}ADR
├── Context: {title}
├── Containers: [list with summaries]
└── ADRs: [list with titles/status]
```
</toc_structure>

<what_toc_tells_me>
- Total scope (from stats)
- All containers with their purposes (from summaries)
- Component distribution per container
- Decision history (ADR titles show what was decided)

This is enough to present a useful overview OR match user intent to specific docs.
</what_toc_tells_me>
</extended_thinking>

### Step 1.3: Also Load Context Overview (Optional but Recommended)

```bash
# Get Context frontmatter + first section for system purpose
head -80 .c3/README.md 2>/dev/null
```

This provides the system-level summary that TOC may reference but not fully contain.

---

## Phase 2: Intelligent Presentation

<extended_thinking>
<goal>Decide what to show based on TOC inventory and user intent</goal>

<decision_matrix>
| User Intent | What to Load | What to Present |
|-------------|--------------|-----------------|
| No specific question | TOC only | Stats + Container table |
| Mentions area/feature | TOC + relevant container(s) | Focused view |
| Asks "why" / decisions | TOC + ADR list | Decision history |
| Wants to change | Redirect | → c3-design skill |
| New to codebase | TOC + Context | Full orientation |
</decision_matrix>

<presentation_principle>
1. TOC stats give the "shape" of the system
2. Container summaries give the "what"
3. Full docs give the "how" - only load when needed
</presentation_principle>
</extended_thinking>

### Scenario A: General Orientation (No Specific Focus)

When user just wants to understand the architecture, present from TOC:

```markdown
## {System Title} Architecture

**Purpose:** {from Context overview}

### System at a Glance
- **Containers:** {count} deployable units
- **Components:** {count} internal parts
- **Decisions:** {count} ADRs

### Container Overview

| ID | Container | Purpose |
|----|-----------|---------|
| `c3-1` | {title} | {summary from TOC} |
| `c3-2` | {title} | {summary from TOC} |
| ... | ... | ... |

### How to Explore

- **System context:** Ask about `c3-0` for boundaries and actors
- **Specific container:** Ask about any `c3-N` for its architecture
- **Decisions:** Ask about ADRs to understand the "why"

What would you like to explore?
```

### Scenario B: Focused Exploration (User Mentions Area)

<extended_thinking>
<goal>When user has a focus, match it to TOC entries and load relevant docs</goal>

<matching_strategy>
User says: "how does authentication work?"

1. Search TOC for matching terms:
   - Container titles/summaries containing "auth"
   - Component entries with "auth"
   - ADR titles mentioning "auth"

2. Identify relevant documents:
   - c3-1 Backend (likely has auth)
   - c3-101 Auth Service (if exists)
   - adr-20251124-auth (if exists)

3. Load those specific docs:
   - Container README for architecture
   - Component doc for implementation
   - ADR for decision rationale

4. Present combined view
</matching_strategy>
</extended_thinking>

When user mentions a specific area:

```bash
# Load Context for system boundary
cat .c3/README.md

# Load relevant container(s) - matched from TOC
cat .c3/c3-{N}-*/README.md

# Load relevant component(s) if specific enough
cat .c3/c3-{N}-*/c3-{NNN}-*.md
```

Present:

```markdown
## {Focus Area} in {System Title}

### System Context
{Relevant excerpt from Context}

### {Container Title} (`c3-{N}`)
{Container README overview}

### Related Components
| ID | Component | What It Does |
|----|-----------|--------------|
| `c3-{N}01` | {title} | {summary} |

### Related Decisions
| ADR | Decision |
|-----|----------|
| `adr-YYYYMMDD-slug` | {title} |

Would you like details on any of these?
```

### Scenario C: Decision History

When user asks about decisions:

```bash
# ADRs are listed in TOC, or enumerate directly
ls .c3/adr/adr-*.md 2>/dev/null
```

From TOC, extract ADR entries and present:

```markdown
## Architecture Decision Records

| Date | ID | Decision | Status |
|------|----|---------| -------|
| {date} | `adr-YYYYMMDD-slug` | {title} | {status} |

These capture the "why" behind architectural choices.

Which decision would you like to explore?
```

---

## Phase 3: Navigation Support

<extended_thinking>
<goal>After initial presentation, help user navigate efficiently</goal>

<navigation_patterns>
User typically wants to:
1. **Drill down** into a specific container → Load that container's README
2. **See components** of a container → Load component docs from that container
3. **Understand a decision** → Load specific ADR
4. **Make changes** → Redirect to c3-design
5. **Check code reality** → Redirect to c3-audit

The key is: don't load everything upfront. Load on demand based on user's next question.
</navigation_patterns>
</extended_thinking>

| User Says | Action |
|-----------|--------|
| "Tell me more about c3-1" | `cat .c3/c3-1-*/README.md` |
| "What's c3-101?" | `cat .c3/c3-1-*/c3-101-*.md` |
| "Why did we choose X?" | Search ADRs, load matching one |
| "I need to change Y" | → Use `c3-design` skill |
| "Does code match docs?" | → Use `c3-audit` skill |

### Quick ID Lookup (from c3-locate patterns)

```bash
# Context
cat .c3/README.md

# Container (e.g., c3-1)
cat .c3/c3-1-*/README.md

# Component (e.g., c3-101)
cat .c3/c3-1-*/c3-101-*.md

# ADR (e.g., adr-20251124-caching)
cat .c3/adr/adr-20251124-caching.md
```

---

## No C3 Found

When `.c3/` directory doesn't exist:

```markdown
This project doesn't have C3 architecture documentation yet.

**C3** (Context-Container-Component) is a layered approach to architecture docs:

| Layer | What It Captures |
|-------|------------------|
| **Context** | System boundary, actors, external integrations |
| **Container** | Deployable units, tech stack, component organization |
| **Component** | Implementation details, configuration, code patterns |

Would you like to set up C3 documentation for this project?
```

**If yes:** Use `c3-adopt` skill
**If no:** Acknowledge and offer other help

---

## Extended Thinking: User Intent Matching

<extended_thinking>
<goal>Parse user's request to determine the best presentation</goal>

<intent_signals>
**General orientation** (→ Scenario A):
- "What's the architecture?"
- "How is this structured?"
- "Give me an overview"
- "I'm new to this codebase"
- No specific technical terms

**Focused exploration** (→ Scenario B):
- "How does X work?"
- "Where is Y handled?"
- "Show me the Z part"
- Mentions specific technology or feature name
- References a container/component by name or ID

**Decision inquiry** (→ Scenario C):
- "Why did we..."
- "What was the reasoning..."
- "When was X decided?"
- "What alternatives were considered?"

**Change intent** (→ redirect to c3-design):
- "I need to add..."
- "We should change..."
- "How do I modify..."
- "I want to refactor..."
</intent_signals>

<response_flow>
1. Parse intent from user's words
2. If general → use TOC stats + container summaries
3. If focused → search TOC for matches, load those docs
4. If decisions → present ADR list from TOC
5. If changes → redirect to c3-design
</response_flow>
</extended_thinking>

---

## C3 Structure Quick Reference

For users unfamiliar with C3:

### Layer Hierarchy

```
Context (c3-0)          ← System boundary, actors
├── Container (c3-1)    ← Deployable unit
│   ├── Component (c3-101)  ← Internal part
│   └── Component (c3-102)
├── Container (c3-2)
│   └── ...
└── ADRs                ← Decision history (cross-cutting)
```

### ID System

| Pattern | Layer | Example |
|---------|-------|---------|
| `c3-0` | Context | System overview |
| `c3-{N}` | Container | `c3-1`, `c3-2` |
| `c3-{N}{NN}` | Component | `c3-101`, `c3-215` |
| `adr-{YYYYMMDD}-{slug}` | Decision | `adr-20251124-caching` |

---

## Related Skills

| Skill | When to Redirect |
|-------|------------------|
| `c3-design` | User wants to make architecture changes |
| `c3-locate` | Need precise ID-based content retrieval |
| `c3-adopt` | Project needs C3 initialization |
| `c3-config` | User wants to change settings |
| `c3-audit` | User wants to verify docs match code |
| `c3-toc` | Need to rebuild or verify TOC accuracy |
