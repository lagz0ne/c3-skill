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
mkdir -p .c3/{containers,components,adr,scripts}
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

- `/c3` - Design or update architecture (main command)
- `/c3-init` - Initialize `.c3/` structure from scratch
- `/c3-migrate` - Migrate documentation to current skill version

## Versioning

C3-Skill uses sequential version numbers to track documentation format changes.

### Version File

Each project using C3 has a `.c3/VERSION` file containing the version number:
```
1
```

### Migration

When the skill evolves, run `/c3-migrate` to upgrade your documentation:
```
/c3-migrate
```

This will:
1. Detect your current version
2. Show what changes are needed
3. Apply transforms with your confirmation

## Documentation Structure

```
.c3/
├── README.md                     # Primary context (id: context)
├── CTX-actors.md                 # Auxiliary context (optional)
├── index.md                      # Navigation index
├── TOC.md                        # Auto-generated TOC
├── VERSION                       # Version number (e.g., "2")
├── containers/
│   └── C3-1-backend.md           # Container documents
├── components/
│   ├── C3-101-db-pool.md         # Component documents (flat!)
│   └── C3-102-auth-service.md    # No nested subfolders
├── adr/
│   └── ADR-001-rest-api.md       # Architecture decisions
└── scripts/
    └── build-toc.sh              # TOC generator
```

> **V2 Change:** Components are now flat in `components/` directory. The container number is encoded in the component ID (e.g., `C3-101` = container 1, component 01).

## Document Conventions

### Unique IDs

Every document has a unique ID:

- **context**: Primary context (`.c3/README.md`)
- **CTX-slug**: Auxiliary context (e.g., `CTX-actors`)
- **C3-<C>-slug**: Container level (e.g., `C3-1-backend`; single digit container number)
- **C3-<C><NN>-slug**: Component level (e.g., `C3-101-db-pool`; container digit + 2-digit component number)
- **ADR-NNN-slug**: Architecture decisions (e.g., `ADR-001-rest-api`)

### Simplified Frontmatter

**For README.md (primary context):**
```yaml
---
id: context
title: System Overview
summary: >
  Bird's-eye view of the system, actors, and key interactions.
  Read this first to understand the overall architecture.
---
```

**For components:**
```yaml
---
id: C3-101-db-pool
title: Database Connection Pool Component
summary: >
  Explains PostgreSQL connection pooling strategy, configuration, and
  retry behavior. Read this to understand how the backend manages database
  connections efficiently and handles connection failures.
---
```

### Heading IDs

Every heading has a unique anchor:

```markdown
## Configuration {#C3-101-configuration}
<!--
Explains environment variables and configuration loading. Read to understand
how to configure the pool for different environments.
-->
```

## Scripts

### Generate TOC

```bash
.c3/scripts/build-toc.sh
```

Regenerate after:
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
