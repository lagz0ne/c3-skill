# Architecture
This project uses C3 docs in `.c3/`.
For architecture questions, changes, audits, file context -> `/c3`.
Operations: query, audit, change, ref, sweep.
File lookup: `c3x lookup <file-or-glob>` maps files/directories to components + refs.
CLI: `bash skills/c3/bin/c3x.sh <command>` (must build first: `bash scripts/build.sh`)

---

# C3 skill design

This is a repository containing a Claude code skill called c3. c3 is a trimmed down concept from c4, focusing on rationalizing architectural and architectural change for large codebase

# Required skills, always load them before hand
- load /superpowers:brainstorming, always do
- load /superpowers-developing-for-claude-code:developing-claude-code-plugins, always do
- use AskUserQuestionTool where possible, that'll give better answer

# Workflow
- Starts with brainstorming to understand clearly the intention
- Once it's all understood, use writing-plan and implement in parallel using subagent
- Delegate to /release command once things is done, confirm with user as needed. Assume to patch by default

---

# Skill Development Philosophy

## The Structure vs Outcome Balance

Shift from prescriptive structure to goal-oriented guidance:

| Structural (Avoid) | Outcome-Oriented (Prefer) |
|--------------------|---------------------------|
| "Create separate file per component" | "Each component should be findable and answerable" |
| "Follow this directory structure" | "Structure should enable navigation" |
| "Add these specific sections" | "Enable developer to understand X" |

## Skill Development Loop

```
Goals → Minimal Skill → Meaningful Eval → Run → Trace Failures → Update Skill → Repeat
```

1. **Start with Goals** - What should the skill enable? What questions should users answer?
2. **Write Minimal Skill** - Focus on outcomes, not structure
3. **Build Meaningful Eval** - Question-based, grounded in actual use
4. **Run Eval** - Identify failing questions
5. **Trace Failures** - Map failures back to skill gaps
6. **Update Skill** - Add targeted guidance, remove unhelpful rules
7. **Repeat** - Improving scores should improve real utility

## Skill Writing Principles

1. **Fewer meaningful sections > many shallow sections** - Don't prescribe structure for structure's sake
2. **Outcome-focused instructions** - "Docs should enable X" not "Docs should contain Y"
3. **Anti-goals are important** - Explicitly state what NOT to do (came from eval failures)
4. **Progressive disclosure** - Core requirements first, details only when needed

## When to Use Prescriptive vs Goal-Oriented

| Use Prescriptive | Use Goal-Oriented |
|------------------|-------------------|
| Hard technical constraints (file names, IDs) | Content quality |
| Integration points (CI expects certain paths) | Organization decisions |
| Non-negotiable conventions | Style and depth |

---

# Plugin Structure

## plugin.json

Auto-discovery only. Do NOT add explicit component paths.

```json
{
  "name": "c3-skill",
  "version": "x.x.x"
}
```

## Repository Layout

```
c3-design/
├── .claude-plugin/           # Plugin metadata
│   ├── plugin.json
│   └── marketplace.json
├── cli/                      # Go CLI source
│   ├── main.go
│   ├── cmd/                  # Command implementations
│   ├── internal/             # Core libraries
│   └── templates/            # Embedded doc templates
├── skills/c3/                # Unified skill (auto-discovered)
│   ├── SKILL.md              # Skill definition + intent router
│   ├── bin/                           # CLI binaries (built in CI)
│   │   ├── c3x.sh                    # Platform-detecting wrapper
│   │   ├── VERSION                   # Current version (read by c3x.sh, committed)
│   │   └── c3x-{version}-{os}-{arch} # Cross-compiled binaries (gitignored)
│   └── references/           # Operation-specific guidance
│       ├── onboard.md
│       ├── query.md
│       ├── audit.md
│       ├── change.md
│       ├── ref.md
│       └── sweep.md
└── scripts/
    └── build.sh              # Cross-compile Go CLI
```

## Build System

```bash
bash scripts/build.sh         # Cross-compile Go CLI for 4 targets
cd cli && go test ./...       # Run Go tests
```

## CI/CD

- **Push to dev** -> `distribute.yml` builds, merges to main via PR, creates release
- **Push to main** -> `release.yml` (fallback) checks `skills/c3/bin/VERSION`
- New version -> GitHub Release with plugin zip

## Versioning

All version files must stay in sync:

| File | Purpose |
|------|---------|
| `skills/c3/bin/VERSION` | Source of truth — CI, c3x.sh, and build.sh all read this |
| `.claude-plugin/plugin.json` | Plugin metadata |
| `.claude-plugin/marketplace.json` | Marketplace listing |

Use `/release` command to bump versions consistently.