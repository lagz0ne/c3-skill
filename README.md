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
| `/c3-use` | **Start here** - Get oriented with existing docs or adopt C3 |
| `/c3` | Design or update architecture |
| `/c3-init` | Initialize `.c3/` structure from scratch |
| `/c3-config` | Configure project settings (diagrams, guardrails, handoff) |
| `/c3-migrate` | Migrate documentation to current skill version |
| `/c3-audit` | Audit C3 docs against codebase reality |

### Skills

| Skill | Purpose |
|-------|---------|
| `c3-use` | Entry point - guides reading or offers adoption |
| `c3-adopt` | Initialize C3 for existing project via Socratic discovery |
| `c3-design` | Main architecture design workflow |
| `c3-config` | Create/update `.c3/settings.yaml` preferences |
| `c3-locate` | Find documents by ID or heading |
| `c3-context-design` | Context-level documentation |
| `c3-container-design` | Container-level documentation |
| `c3-component-design` | Component-level documentation |
| `c3-naming` | ID and naming conventions |
| `c3-toc` | Table of contents management |
| `c3-migrate` | Version migration transforms |

### Getting Started

**For new users:** Add to your project's `CLAUDE.md`:

```markdown
# Architecture

Use `/c3-skill:c3-use` to get started with architecture documentation.
```

This will check for `.c3/` and guide you through reading existing docs or setting up C3.

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

Use the c3-toc skill to rebuild the table of contents after:
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

See `examples/` directory for:
- Sample `.c3/` structure
- Example context/container/component documents
- Sample ADRs

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
