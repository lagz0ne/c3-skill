# C3-Skill Plugin

Claude Code plugin for C3 (Context-Container-Component) architecture design methodology.

## Overview

C3-Skill provides a structured approach to documenting system architecture through three levels:

1. **Context**: Bird's-eye view of system landscape, actors, and cross-component interactions
2. **Container**: Individual container characteristics, technology, and component organization
3. **Component**: Implementation-level details, configuration, and technical specifics

## Features

- **Intelligent State Detection**: Reads existing `.c3/` docs to understand current architecture
- **Phased Workflow**: Guides through Understand, Confirm, Scope, Document phases
- **Targeted Socratic Questioning**: Asks only when clarification needed
- **ADR Generation**: Documents decisions with progressive detail across C3 levels
- **Unique IDs**: Every document and heading uniquely identified for precise navigation
- **Auto-Generated TOC**: Build table of contents from frontmatter summaries
- **VitePress Integration**: Beautiful rendered documentation with search and navigation
- **Mermaid Diagrams**: Embedded diagrams at appropriate abstraction levels

## Installation

### From Marketplace

```bash
# Add the marketplace
/plugin marketplace add Lagz0ne/c3-skill

# Install the plugin
/plugin install c3-skill
```

### Manual Installation

1. Clone this repository to `~/.claude/plugins/c3-skill`
2. Restart Claude Code

## Usage

### Initialize C3 Documentation

Create `.c3/` directory in your project:

```bash
mkdir -p .c3/adr
```

### Design Architecture

The main skill activates when you describe architecture work:

```
"I need to design the authentication system"
"Add a caching layer to the backend"
"Document the current microservices architecture"
```

Claude will:
1. Read your existing `.c3/` docs (if any)
2. Understand what level this affects (Context/Container/Component)
3. Confirm understanding via targeted questions
4. Create/update documentation with appropriate diagrams

### Commands

| Command | Description |
|---------|-------------|
| `/c3` | **Start here** - Design, update, or explore architecture |
| `/c3-config` | Configure project settings (diagrams, guardrails, handoff) |
| `/c3-migrate` | Migrate documentation to current skill version |
| `/c3-audit` | Audit C3 docs against codebase reality |

### Skills (8)

| Skill | Purpose |
|-------|---------|
| `c3-design` | Main entry point - design, update, or explore architecture |
| `c3-adopt` | Initialize C3 for any project (existing or fresh) |
| `c3-context-design` | Context-level documentation |
| `c3-container-design` | Container-level documentation |
| `c3-component-design` | Component-level documentation |
| `c3-config` | Create/update `.c3/settings.yaml` preferences |
| `c3-migrate` | Version migration transforms |
| `c3-audit` | Audit C3 docs against codebase reality |

### References

Utility patterns extracted as reference docs (not skills):
- `lookup-patterns.md` - ID-based document retrieval
- `naming-conventions.md` - ID and naming conventions

### Getting Started

**For new users:** Add the following to your project's `CLAUDE.md` to help agents work with C3 documentation:

<details>
<summary><strong>📋 Click to expand CLAUDE.md template (copy this)</strong></summary>

```markdown
# Architecture (C3)

This project uses C3 (Context-Container-Component) architecture documentation.

## Quick Start

\`\`\`bash
# Check if C3 docs exist
ls .c3/ 2>/dev/null && echo "C3 docs exist - read before changing"
\`\`\`

## When to Use C3

| Situation | Action |
|-----------|--------|
| "How does X work?" | Read `.c3/` docs first |
| Adding new feature | Check which container it belongs to |
| Changing how components interact | Use `/c3` or `c3-design` skill |
| Refactoring across modules | Verify changes align with documented flows |
| Onboarding question | Point to `.c3/README.md` |

## Navigation

| Want to understand... | Read |
|-----------------------|------|
| System overview | `.c3/README.md` |
| How containers talk | `.c3/README.md` → Container Interactions |
| What a container does | `.c3/c3-{N}-*/README.md` |
| How a component works | `.c3/c3-{N}-*/c3-{N}{NN}-*.md` |

## Red Flags (Stop and Check C3)

🚩 Adding a new service/module without checking container boundaries
🚩 Changing API contracts without checking documented protocols
🚩 "I'll figure out the architecture as I code"
🚩 Making changes that span multiple containers

## ADR Policy

**DO NOT read ADRs** unless specifically asked about past decisions:
- ADRs are historical records, not active guidance
- Reading them adds context that may cause hallucination
- Focus on current state in Context/Container/Component docs
```

</details>

> 💡 **Tip:** This template is also available at [`templates/CLAUDE.md.c3-section`](templates/CLAUDE.md.c3-section)

This will help agents:
- Check `.c3/` docs before making architectural changes
- Use the right skill for the right situation
- Avoid reading ADRs unnecessarily (reduces hallucination)

## Layer Configuration

Each C3 layer (Context, Container, Component) has configurable rules for what content belongs at that level.

### Defaults

Each layer skill includes a `defaults.md` with sensible defaults:
- `skills/c3-context-design/defaults.md`
- `skills/c3-container-design/defaults.md`
- `skills/c3-component-design/defaults.md`

### Customization

Customize layer rules in `.c3/settings.yaml`:

```yaml
context:
  useDefaults: true    # Load defaults, then merge customizations
  guidance: |
    your team's context-level guidance
  include: |
    compliance requirements
    audit trails
  exclude: |
    internal service details
  litmus: |
    "Does this affect external stakeholders?"

container:
  useDefaults: true
  guidance: |
    your team's container-level guidance

component:
  useDefaults: false   # Only use what's specified here
  guidance: |
    your team's component-level guidance
  include: |
    your custom include list
```

### Configuration Keys

| Key | Type | Description |
|-----|------|-------------|
| `useDefaults` | boolean | `true`: merge with defaults. `false`: only use settings.yaml |
| `guidance` | prose | General documentation guidance |
| `include` | list | Items that belong at this layer |
| `exclude` | list | Items to push to other layers |
| `litmus` | prose | Decision question for content placement |
| `diagrams` | list | Diagram types for this layer |

## Versioning

C3-Skill uses `c3-version` frontmatter to track documentation format changes.

### Migration

When the skill evolves, run `/c3-migrate` to upgrade your documentation:
```
/c3-migrate
```

This will:
1. Detect your current version from frontmatter
2. Show what changes are needed
3. Apply transforms with your confirmation

## Documentation Structure

```
.c3/
├── README.md                      # Context (id: c3-0, c3-version: 3)
├── actors.md                      # Aux context (applies downward)
├── TOC.md                         # Auto-generated
├── adr/
│   └── adr-20250115-rest-api.md
├── c3-1-backend/                  # Container folder
│   ├── README.md                  # Container doc (id: c3-1)
│   ├── c3-101-db-pool.md          # Component (id: c3-101)
│   └── c3-102-auth-service.md
└── c3-2-frontend/
    ├── README.md                  # Container doc (id: c3-2)
    └── c3-201-api-client.md
```

> **V3+ Changes:**
> - Containers are now folders (`c3-{N}-name/`) containing their components
> - All IDs use numeric format: `c3-0` (context), `c3-1` (container), `c3-101` (component)
> - All filenames are lowercase
> - Version tracked via `c3-version` frontmatter (format: `YYYYMMDD-slug`)
> - ADRs use date-based naming: `adr-{YYYYMMDD}-{slug}` (e.g., `adr-20250115-rest-api.md`)

## Document Conventions

### Unique IDs

Every document has a unique ID:

- `c3-0`: Context level (`.c3/README.md`)
- `c3-{N}`: Container level (e.g., `c3-1` in `c3-1-backend/README.md`)
- `c3-{N}{NN}`: Component level (e.g., `c3-101` in `c3-1-backend/c3-101-db-pool.md`)
- `adr-{YYYYMMDD}-{slug}`: Architecture decisions (date-based, e.g., `adr-20250115-rest-api`)

### Frontmatter

**For README.md (context):**
```yaml
---
id: c3-0
c3-version: 3
title: System Overview
summary: >
  Bird's-eye view of the system, actors, and key interactions.
  Read this first to understand the overall architecture.
---
```

**For containers:**
```yaml
---
id: c3-1
title: Backend Services
summary: >
  Core backend services handling API, authentication, and data persistence.
---
```

**For components:**
```yaml
---
id: c3-101
title: Database Connection Pool
summary: >
  Explains PostgreSQL connection pooling strategy, configuration, and
  retry behavior.
---
```

### Heading IDs

Every heading has a unique anchor:

```markdown
## Configuration {#c3-101-configuration}
<!--
Explains environment variables and configuration loading. Read to understand
how to configure the pool for different environments.
-->
```

## Scripts

### Generate TOC

Ask Claude to "rebuild the TOC" (uses the plugin's `build-toc.sh` script) after:
- Creating new documents
- Updating summaries
- Adding new sections

## VitePress Setup

To render documentation:

1. Install VitePress in `.c3/`:
   ```bash
   cd .c3
   npm init -y
   npm install -D vitepress
   ```

2. Add to `package.json`:
   ```json
   {
     "scripts": {
       "docs:dev": "vitepress dev",
       "docs:build": "vitepress build",
       "docs:preview": "vitepress preview"
     }
   }
   ```

3. Start dev server:
   ```bash
   npm run docs:dev
   ```

## Examples

See the plugin's own `.c3/` directory for a real-world example of C3 documentation structure.

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

MIT License - see LICENSE file for details

## Links

- [Repository](https://github.com/Lagz0ne/c3-skill)
- [Issues](https://github.com/Lagz0ne/c3-skill/issues)
- [C4 Model](https://c4model.com/)
- [ADR](https://adr.github.io/)
