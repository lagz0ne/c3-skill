# C3-Skill Plugin Design

**Date**: 2025-01-21
**Status**: Approved for Implementation

## Executive Summary

The c3-skill plugin provides a structured methodology for documenting system architecture using the Context-Container-Component (C3) approach. It guides developers through top-down architecture design with intelligent state detection, Socratic questioning, and auto-generated documentation with VitePress.

## Overview

### What is C3?

C3 is a three-level architecture documentation methodology:

1. **Context**: Bird's-eye view of system landscape - actors, boundaries, cross-component interactions
2. **Container**: Individual container characteristics - technology, middleware, component organization
3. **Component**: Implementation details - configuration, dependencies, technical specifics

Each level maintains appropriate abstraction with embedded mermaid diagrams, cross-references, and unique identifiers.

### Plugin Goals

- **Intelligent**: Detect current state from existing docs, infer intention from requests
- **Guided**: Phased workflow with targeted Socratic questioning
- **Structured**: Unique IDs, consistent conventions, heading-level frontmatter
- **Navigable**: VitePress-powered docs with auto-generated TOC
- **Maintainable**: Scripts for lookup, validation, and TOC generation

## Architecture Approach

**Hybrid Progressive Elaboration**: Main orchestrator skill detects intention and progressively invokes sub-skills as needed.

### Plugin Structure

```
c3-skill/
├── .claude-plugin/
│   ├── plugin.json           # Plugin metadata
│   └── marketplace.json      # Marketplace listing
├── skills/
│   ├── c3-design/
│   │   ├── SKILL.md          # Main orchestrator skill
│   │   ├── intentions/       # Intention-specific workflows
│   │   │   ├── from-scratch.md
│   │   │   ├── update-existing.md
│   │   │   ├── migrate-system.md
│   │   │   └── audit-current.md
│   │   └── templates/        # Reusable templates
│   ├── c3-context-design/
│   │   └── SKILL.md          # Context level skill
│   ├── c3-container-design/
│   │   └── SKILL.md          # Container level skill
│   └── c3-component-design/
│       └── SKILL.md          # Component level skill
├── README.md
├── LICENSE
└── .gitignore
```

### Skill Workflow

**Main Skill (c3-design)**: 4-Phase Orchestrator

**Phase 1: Understand Current Situation**
- Check for `.c3/` directory (required location)
- Read existing documents, parse frontmatter
- Analyze user request to determine affected C3 levels
- Use Socratic questioning ONLY if ambiguous

**Phase 2: Analyze & Confirm Intention** (Socratic Required)
- Map requirement against current structure
- Determine impacted levels (Context/Container/Component)
- Present understanding concisely
- Confirm alignment with targeted questions

**Phase 3: Scoping the Change**
- Identify cross-cutting concerns
- Define boundaries (what will NOT change)
- Clarify scope if uncertain

**Phase 4: Suggest Changes via ADR**
- Create Architecture Decision Record in `.c3/adr/ADR-NNN-title.md`
- Start high-level, progressively add detail
- ADR becomes planning artifact

**Sub-Skills**: Invoked via SlashCommand as needed:
- `/c3-context-design`: Design Context level
- `/c3-container-design`: Design Container level
- `/c3-component-design`: Design Component level

## Documentation Conventions

### Unique ID System

**Format**: `[CATEGORY-NUMBER-SLUG]`

| Category | Prefix | Location | Example |
|----------|--------|----------|---------|
| Context | `CTX-` | `.c3/` | `CTX-001-system-overview.md` |
| Container | `CON-` | `.c3/containers/` | `CON-001-backend.md` |
| Component | `COM-` | `.c3/components/{container}/` | `COM-001-db-pool.md` |
| ADR | `ADR-` | `.c3/adr/` | `ADR-001-rest-api.md` |

### Simplified Frontmatter

**Document-Level** (focuses on summary):
```yaml
---
id: COM-001-db-pool
title: Database Connection Pool Component
summary: >
  Explains PostgreSQL connection pooling strategy, configuration, and
  retry behavior. Read this to understand how the backend manages database
  connections efficiently and handles connection failures.
---
```

**Heading-Level** (optional, for important sections):
```markdown
## Configuration {#com-001-configuration}
<!--
Explains environment variables, configuration loading strategy, and
differences between development and production. Read this section to
understand how to configure the connection pool for different environments.
-->
```

### Directory Structure

```
.c3/
├── index.md                      # Conventions & navigation
├── TOC.md                        # Auto-generated (never edit)
├── CTX-XXX-*.md                  # Context documents
├── containers/
│   └── CON-XXX-*.md              # Container documents
├── components/
│   ├── {container-name}/
│   │   └── COM-XXX-*.md          # Component documents
├── adr/
│   └── ADR-XXX-*.md              # Architecture decisions
├── scripts/
│   ├── build-toc.sh              # Generate TOC
│   ├── lookup.sh                 # Content lookup
│   └── validate.sh               # Structure validation
└── .vitepress/
    └── config.ts                 # VitePress config
```

## Sub-Skill Design

### Context Level (c3-context-design)

**Purpose**: Define system landscape - containers, users, systems, and cross-component interactions.

**Context Defines**:
- Components underneath (containers, users, external systems)
- Cross-component interactions and protocols
- High-level concerns (gateway, deployment, auth, protocols)

**Output**:
- `.c3/CTX-XXX-{slug}.md` with:
  - Architecture/flowchart diagram (required)
  - Sequence diagrams for protocols
  - Flow diagrams for special concepts
  - Links to all containers

### Container Level (c3-container-design)

**Purpose**: Define individual containers - technology, middleware, component organization.

**Container Defines**:
- Container identity and responsibility
- Technology stack
- Middleware/proxy layer (auth, cookies, rate limiting)
- Component organization
- Inter-container communication

**Abstraction Level**: WHAT and WHY, not HOW. Focus on characteristics and properties, not code.

**Output**:
- `.c3/containers/CON-XXX-{slug}.md` with:
  - C4 Container diagram or component structure
  - Middleware pipeline diagram
  - Data flow diagrams
  - Links up to context, down to components, across to related containers

### Component Level (c3-component-design)

**Purpose**: Implementation-level detail - configuration, dependencies, technical specifics.

**Component Defines**:
- Technical implementation (libraries, patterns)
- Configuration (env vars, config files)
- Dependencies and interfaces
- Error handling and performance characteristics

**Abstraction Level**: Implementation details. Code examples and configuration snippets are appropriate.

**Output**:
- `.c3/components/{container-name}/COM-XXX-{slug}.md` with:
  - Sequence/class/state/flow diagrams
  - Configuration examples
  - Code snippets where helpful
  - Links up to container, across to related components

### ADR Structure

**Purpose**: Document architectural decisions with progressive detail across C3 levels.

**Output**:
- `.c3/adr/ADR-XXX-{slug}.md` with:
  - Context: Current situation and why change needed
  - Decision: High-level approach (Context level first)
  - Container-level details
  - Component-level impact
  - Alternatives considered
  - Consequences and mitigation

## Auto-Generated TOC

The TOC is built from document and heading frontmatter summaries:

**Script**: `.c3/scripts/build-toc.sh`

**Process**:
1. Loop through all files (Context → Container → Component → ADR)
2. Extract frontmatter summaries
3. Extract heading summaries (from HTML comments)
4. Build tree structure with cross-references
5. Generate `.c3/TOC.md`

**When to Regenerate**:
- After creating new documents
- After updating summaries
- When documents seem out of date

**TOC Provides**:
- Document summaries (why read this doc)
- Section summaries (what each heading explains)
- Cross-references between levels
- Quick navigation links

## Supporting Scripts

### build-toc.sh
Generates table of contents from frontmatter summaries.

### lookup.sh
Provides content lookup by ID, category, or search term:
- `lookup.sh doc <ID>` - Find document path
- `lookup.sh list <CATEGORY>` - List all in category
- `lookup.sh search <TERM>` - Search content
- `lookup.sh heading <DOC-ID> <HEADING-ID>` - Extract heading

### validate.sh
Validates document structure:
- Check ID uniqueness
- Validate frontmatter
- Check broken links
- Verify heading IDs follow convention

## VitePress Integration

**Benefits**:
- Beautiful rendered documentation
- Sidebar navigation with TOC
- Search functionality
- Link validation
- Heading anchors for deep linking

**Configuration**: `.c3/.vitepress/config.ts`
- Sidebar automatically generated from directory structure
- Mermaid diagrams rendered inline
- Outline panel shows document structure

## Plugin Metadata

**plugin.json**:
```json
{
  "name": "c3-skill",
  "description": "C3 (Context-Container-Component) architecture design methodology with structured top-down documentation, mermaid diagrams, and intention-based workflows",
  "version": "1.0.0",
  "author": {
    "name": "Lagz0ne"
  },
  "homepage": "https://github.com/Lagz0ne/c3-skill",
  "repository": "https://github.com/Lagz0ne/c3-skill",
  "license": "MIT",
  "keywords": ["architecture", "design", "c3", "documentation", "mermaid", "system-design"]
}
```

## Implementation Phases

### Phase 1: Core Skills
- Main c3-design orchestrator skill
- Three sub-skills (context, container, component)
- Basic intention detection
- Phased workflow implementation

### Phase 2: Documentation Structure
- Document templates with frontmatter
- ID conventions and heading structure
- Basic VitePress setup

### Phase 3: Automation Scripts
- build-toc.sh for TOC generation
- lookup.sh for content queries
- validate.sh for structure checking

### Phase 4: Polish & Package
- README and documentation
- Example .c3 structure
- Plugin metadata for marketplace
- Testing with real projects

## Success Criteria

- ✓ Skills activate appropriately based on user intent
- ✓ Phased workflow guides without overwhelming
- ✓ TOC regenerates correctly from frontmatter
- ✓ Lookup scripts find content reliably
- ✓ VitePress renders navigation correctly
- ✓ Documents follow conventions consistently
- ✓ Cross-references link properly

## Future Enhancements

- AI-powered diagram generation from descriptions
- Interactive diagram editing
- Automated ADR generation from git diffs
- Template library for common patterns
- Validation webhooks for CI/CD

## References

- C4 Model: https://c4model.com/
- ADR: https://adr.github.io/
- VitePress: https://vitepress.dev/
- Claude Code Skills: https://docs.claude.com/en/docs/claude-code/skills
