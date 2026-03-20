# Rules Marketplace: GitHub-Based Distribution for C3 Coding Rules

## Problem

C3 rules are project-local. Teams document the same patterns (error handling, logging conventions, API response formats) independently across projects. There's no way to share, discover, or adopt proven coding rules from other teams or the community.

## Design

### Concept

A **marketplace** is a GitHub repo containing reusable coding rules. Users add repos by URL (shallow clone), browse available rules, and adopt them into their project's `.c3/rules/` through an LLM-guided section-by-section merge that adapts each rule to the project's context.

Distribution model mirrors Claude Code plugins: user-driven, GitHub-based, no central registry.

### Marketplace Repo Format

```
my-go-rules/
├── marketplace.yaml              # manifest (required)
├── rule-error-handling.md        # standard .c3/ rule format
├── rule-config-loading.md
└── rule-graceful-shutdown.md
```

Rule files use the **exact same format** as `.c3/rules/*.md` — same frontmatter, same sections (Goal, Rule, Golden Example, Not This, Scope, Override). Any existing project rule can be extracted into a marketplace repo with zero conversion.

#### marketplace.yaml

```yaml
name: go-patterns
description: Opinionated Go patterns for production services
tags: [go, backend, production]
compatibility:
  languages: [go]
  frameworks: []                  # optional: [gin, echo, fiber]
rules:
  - id: rule-error-handling
    title: Structured Error Handling
    category: reliability
    tags: [errors, observability]
    summary: Wrap errors with context, use sentinel types, structured logging
  - id: rule-config-loading
    title: Config from Environment
    category: operations
    tags: [config, 12-factor]
    summary: Single config struct, env vars, validation at startup
  - id: rule-graceful-shutdown
    title: Graceful Shutdown
    category: reliability
    tags: [lifecycle, signals]
    summary: Signal handling, drain connections, timeout-bounded shutdown
```

**Required fields:** `name`, `description`, `rules` (with `id` and `summary` per entry).
**Optional fields:** `tags`, `compatibility`, per-rule `title`/`category`/`tags`.

### CLI Commands

New `c3x marketplace` subcommand group:

| Command | Purpose |
|---------|---------|
| `c3x marketplace add <github-url>` | Shallow clone repo into `~/.c3/marketplace/<name>/`, register in sources |
| `c3x marketplace list [--tag <tag>] [--source <name>]` | List available rules across all sources, filterable by tag or source |
| `c3x marketplace show <rule-id>` | Print full rule content for preview (searches all sources) |
| `c3x marketplace update [<source-name>]` | Git pull latest from one or all registered sources |
| `c3x marketplace remove <source-name>` | Delete local cache + unregister |

#### Cache Layout

```
~/.c3/marketplace/
├── sources.yaml                  # registered repos
├── go-patterns/                  # shallow clone of github.com/org/go-patterns
│   ├── marketplace.yaml
│   ├── rule-error-handling.md
│   └── rule-config-loading.md
└── react-patterns/
    ├── marketplace.yaml
    └── rule-component-structure.md
```

#### sources.yaml

```yaml
sources:
  - name: go-patterns
    url: https://github.com/org/go-patterns
    fetched: 2026-03-20T10:00:00Z
  - name: react-patterns
    url: https://github.com/org/react-patterns
    fetched: 2026-03-19T15:30:00Z
```

#### Fetch Strategy

Shallow `git clone --depth 1` for initial add, `git pull` for update. Git handles auth (SSH keys, credential helpers) — no custom token management needed. Works with private repos out of the box.

### Adoption Flow (Skill-Driven)

The `references/rule.md` skill reference gains a new **Adopt** mode, triggered by: "adopt rule-X from marketplace", "install rule from", "use marketplace rule".

#### Flow: `Preview → Discover Overlap → Guided Merge → Write → Wire → ADR`

**Step 1: Preview**
```bash
bash <skill-dir>/bin/c3x.sh marketplace show <rule-id>
```
Display the full marketplace rule to the user.

**Step 2: Discover Overlap (2-5 Grep calls)**
LLM searches the project codebase for existing implementations that overlap with the marketplace rule's pattern. Looks for:
- Existing `.c3/rules/` or `.c3/refs/` that cover similar ground
- Code patterns matching the rule's `## Golden Example`
- Anti-patterns matching the rule's `## Not This`

**Step 3: Section-by-Section Guided Merge**
For each rule section (Goal, Rule, Golden Example, Not This, Scope):
- Show marketplace version alongside what the project currently does
- User picks via `AskUserQuestion`:
  - **Adopt as-is** — take the marketplace version verbatim
  - **Adapt** — LLM rewrites the section to fit the project's conventions, tech stack, naming
  - **Skip** — omit this section (only for optional sections like Scope, Override)

**Step 4: Write**
```bash
bash <skill-dir>/bin/c3x.sh add rule <slug>
# Then c3x set for each section with adapted content
```

**Step 5: Wire**
```bash
bash <skill-dir>/bin/c3x.sh wire <component> rule-<slug>
```
For each component the LLM identified as using the pattern.

**Step 6: Adoption ADR**
```yaml
---
id: adr-YYYYMMDD-adopt-rule-{slug}
title: Adopt {Rule Title} from {marketplace-source}
status: implemented
affects: [rule-{slug}]
---

Adopted from marketplace source `{source-name}` ({github-url}).
Sections adapted: {list of adapted sections}.
```

### What Changes in Existing Code

| Area | File(s) | Change |
|------|---------|--------|
| Go CLI | `cli/cmd/marketplace.go` (new) | `marketplace` command group: add, list, show, update, remove |
| Go CLI | `cli/cmd/marketplace_add.go` (new) | Shallow clone, parse manifest, register source |
| Go CLI | `cli/cmd/marketplace_list.go` (new) | Aggregate rules from all sources, filter by tag/source |
| Go CLI | `cli/cmd/marketplace_show.go` (new) | Find rule across sources, print content |
| Go CLI | `cli/internal/marketplace/` (new) | Manifest parsing, source registry, cache management |
| Go CLI | `cli/cmd/root.go` | Register marketplace subcommand |
| Go CLI | `cli/main.go` | No change (auto-discovers from root) |
| Skill | `skills/c3/references/rule.md` | Add Adopt mode (section after Migrate) |
| Skill | `skills/c3/SKILL.md` | Add marketplace intent keywords |

### marketplace.yaml Validation

`c3x marketplace add` validates the manifest on fetch:
- `name` must be non-empty and valid as a directory name
- Each `rules[].id` must match a `rule-*.md` file in the repo
- Duplicate source names are rejected (use `remove` + `add` to update URL)
- Missing `marketplace.yaml` → error with hint to create one

### Edge Cases

| Case | Behavior |
|------|----------|
| Rule ID conflicts across sources | `marketplace show` returns first match; `marketplace show --source <name>` for disambiguation |
| Rule already exists in project | Discover Overlap (Step 2) surfaces it; user decides to merge or skip |
| Source repo becomes unavailable | `marketplace update` reports error; cached version remains usable |
| No `.c3/` in project | Adoption flow runs `c3x init` first (same as other operations) |
| Private repo | Git credentials handle auth transparently (SSH, credential helper) |

## Non-Goals

- **Central registry** — users bring their own repo URLs. Discovery is external (GitHub search, READMEs, word of mouth).
- **Version pinning** — shallow clone gets latest. If users need versioning, they use git tags + branches in their marketplace repos.
- **Automatic rule updates** — adoption is copy + adapt. No ongoing sync. If the upstream rule changes, users re-adopt manually.
- **Marketplace publishing from c3x** — creating a marketplace repo is just creating a git repo with rule files + `marketplace.yaml`. No special tooling needed.

## Testing

- `c3x marketplace add <url>` → clones into `~/.c3/marketplace/<name>/`, creates `sources.yaml`
- `c3x marketplace list` → aggregates rules from all sources with correct metadata
- `c3x marketplace list --tag go` → filters correctly
- `c3x marketplace show rule-error-handling` → prints full rule content
- `c3x marketplace update` → pulls latest for all sources
- `c3x marketplace remove <name>` → deletes cache + unregisters
- Invalid `marketplace.yaml` → clear error message
- Missing rule file (listed in manifest but not present) → validation error on add
- Duplicate source name → rejected with helpful message
- Adopt flow: end-to-end from `marketplace show` → `.c3/rules/rule-*.md` with adapted content
