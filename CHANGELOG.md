# Changelog

All notable changes to the C3 Skill plugin will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [3.3.5] - 2026-01-20

### Fixed
- **plugin.json**: Added explicit `commands`, `skills`, `agents` declarations - auto-discovery wasn't working reliably, explicit declarations ensure Claude Code finds all components

### Documentation
- **CLAUDE.md**: Added comprehensive plugin structure checklist with pre-release requirements, plugin.json template, and common issues troubleshooting table

## [3.3.4] - 2026-01-20

### Fixed
- **plugin.json**: Added missing `hooks` field referencing `hooks/hooks.json` - hooks were defined but not loaded because plugin manifest didn't reference the hooks file

## [3.3.3] - 2026-01-20

### Fixed
- **plugin.json**: Removed explicit component paths - paths were relative to `.claude-plugin/` not plugin root, causing auto-discovery to fail. Now uses Claude Code's default auto-discovery.
- **onboard skill**: Aligned frontmatter `name: onboard` with directory name (was `onboarding-c3`)
- **onboard command**: Updated skill reference to match corrected skill name
- **c3-orchestrator agent**: Changed invalid color `orange` to `yellow`

### Documentation
- **CLAUDE.md**: Added plugin troubleshooting section with validation workflow

## [3.3.2] - 2026-01-20

### Fixed
- **plugin.json**: Added missing `skills` and `hooks` declarations - skills and hooks were not loading due to missing manifest entries

## [3.3.1] - 2026-01-19

### Fixed
- **TOC hook timing**: Moved TOC rebuild from Stop hook to PreToolUse on Bash (git commit) - ensures TOC changes are included in commits instead of being left behind
- **build-toc.sh**: Removed timestamp from TOC header, added content comparison to skip updates when nothing changed

## [3.3.0] - 2026-01-19

### Added
- **c3-orchestrator agent**: Multi-agent system for orchestrating architectural changes with Socratic dialogue before ADR generation
  - `c3-analyzer`: Current state extraction sub-agent (sonnet)
  - `c3-impact`: Dependency tracing and risk assessment sub-agent (sonnet)
  - `c3-patterns`: Convention checking against refs sub-agent (sonnet)
  - `c3-synthesizer`: Critical thinking synthesis sub-agent (opus)
- **ADR-gated code changes**: New `approved-files` field in ADR frontmatter
  - c3-gate now blocks Edit/Write unless file is in accepted ADR's approved-files list
  - Enforces analysis-before-change workflow in C3-adopted projects

### Changed
- **SessionStart hook**: Now routes to agents (c3-navigator for questions, c3-orchestrator for changes) instead of skills
- **c3 skill routing**: Updated mode selection to use agents for navigation and changes
- **Session harness**: Simplified with inline mermaid diagram showing routing flow

### Fixed
- Mermaid diagram syntax in session hooks (removed quotes in labels)

## [3.2.0] - 2026-01-19

### Added
- **c3-navigator agent**: Dedicated agent that triggers on any question in projects with `.c3/` directory. Provides architecture answers with visual diagrams via diashort service. Runs in separate context window for token efficiency.
- **c3-summarizer sub-agent**: Haiku-powered extraction agent that efficiently summarizes C3 documentation (~500 tokens output). Called by c3-navigator via Task tool for optimal token usage.

### Architecture
- Two-agent orchestration pattern: navigator identifies scope, summarizer extracts relevant facts
- 70-90% token savings compared to reading all `.c3/` docs in main context
- Auto-generated Mermaid diagrams rendered via diashort service
- Adaptive output format based on query type (structural, behavioral, flow)

## [3.1.0] - 2026-01-14

### Added
- **Goal-driven templates with staged onboarding workflow**: Templates now guide through staged adoption with clear goals at each step

### Fixed
- **Removed auxiliary references**: Completed migration from auxiliary category to refs system throughout all documentation and skills

## [3.0.0] - 2026-01-09

### Breaking Changes
- **Refs system replaces auxiliary category**: Components no longer have auxiliary category. Code references are now managed through explicit `refs/` documents that link architecture to code.
- **Templates restructured**: Merged `component-foundation.md` and `component-feature.md` into unified `component.md` with Goal-first structure.

### Added
- **Refs system**: New `templates/ref.md` for creating explicit code-to-architecture references
- **Ref lookup in c3-query**: Query skill now resolves refs to find relevant code locations
- **Ref maintenance in c3-alter**: Alter skill updates refs when code moves or changes
- **Goal-first template structure**: All templates now start with Goal section for clearer intent

### Changed
- **Context template**: Restructured with Goal-first approach, clearer actor/system sections
- **Container templates**: Service, database, queue templates all use Goal-first structure
- **Component template**: Unified template replaces separate foundation/feature templates
- **Audit checks**: Updated to validate refs instead of auxiliary category

### Documentation
- Layer navigation updated for refs resolution
- V3 structure documentation reflects refs system
- Removed all auxiliary category references

### Rationale
The auxiliary category was ambiguous and led to inconsistent categorization. Refs provide explicit, traceable links between architecture documentation and actual code, improving maintainability and audit accuracy.

## [2.5.0] - 2026-01-06

### Added
- **Complexity-first documentation approach**: Assess container complexity BEFORE documenting aspects
- **Harness complexity rules**: COMPLEXITY-BEFORE-ASPECTS, DISCOVERY-OVER-CHECKLIST, DEPTH-MATCHES-COMPLEXITY
- **Type-specific container templates**: `container-service.md`, `container-database.md`, `container-queue.md` with discovery prompts
- **External system templates**: `external.md` and `external-aspect.md` for documenting external dependencies
- **Complexity levels**: trivial → simple → moderate → complex → critical with clear documentation depth rules

### Changed
- **Harness as single source**: All complexity rules defined in `skill-harness.md`, templates reference it
- **Templates use discovery prompts**: No more pre-populated aspect checklists that bias AI
- **Container templates simplified**: Type-specific signals to scan, not assumptions to check off
- **c3 skill Adopt flow**: Now explicitly requires complexity-first and discovery-over-checklist

### Rationale
This release prevents AI bias from pre-populated checklists. By requiring complexity assessment first and discovery through code analysis, documentation reflects what actually exists rather than what templates assume.

## [2.4.0] - 2026-01-06

### Added
- **Component types reference**: New `references/component-types.md` consolidates Foundation/Feature/Auxiliary guidance with decision flowchart and dependency rules
- **Progressive complexity diagram**: Skill harness now shows simple → complex skill selection visually
- **Violation examples in harness**: Concrete wrong vs right examples for common mistakes

### Changed
- **Templates: diagram-first approach**: Linkage tables replaced with mermaid flowcharts, testing tables replaced with strategy prose
- **Skill descriptions**: All three skills now use third-person format with specific trigger phrases per plugin-dev guidelines
- **Template comments**: Multi-line AI hints consolidated to single-line format for token efficiency
- **Component templates**: Added type selector hints at top (e.g., `<!-- USE: Core primitives -->`)

### Documentation
- Added `docs/plans/2026-01-06-diagram-first-templates-design.md` capturing the design rationale

## [2.3.0] - 2026-01-05

### Added
- **Component references section**: All component templates now include `## References` section for explicit code-to-architecture links
- **Reference validation in audit**: New Phase 4 validates reference validity (symbols/paths/globs exist) and coverage (major code areas referenced)
- **Pre-execution checklist**: Plan template includes checklist item to update component references before implementation

### Changed
- Onboarding workflow now includes step to populate `## References` after drafting each component
- ADR workflow tracks "References Affected" and updates references during execution when code moves
- Audit procedure renumbered: References Validation (Phase 4), Diagram Accuracy (Phase 5), ADR Lifecycle (Phase 6)

### Documentation
- Added `docs/plans/2026-01-05-component-references-design.md` capturing the design rationale

## [2.2.0] - 2026-01-05

### Added
- **Staged onboarding with recursive learning loop**: Adoption now progresses through 5 stages (Context → Containers → Auxiliary → Foundation → Feature) with analysis and validation at each level
- **Socratic questioning**: All skills now use `AskUserQuestion` tool to clarify understanding before proceeding, continuing until confident (no open questions)
- **Bidirectional navigation**: When conflicts are discovered at deeper levels, the system ascends to fix parent documentation, then re-descends
- **Tiered assumptions**: High-impact changes (new External Systems, container boundaries) require user confirmation; low-impact changes (linkage reasoning, naming) auto-proceed
- **Progress tracking in ADR-000**: Adoption ADR now includes a Progress section showing documented vs remaining items per level
- **Intent clarification for queries**: c3-query now clarifies ambiguous queries before searching

### Changed
- `commands/onboard.md`: Complete rewrite implementing 5-stage recursive learning loop
- `skills/c3-alter/SKILL.md`: Rewritten with 7-stage workflow (Intent → Understand → Scope → ADR → Plan → Execute → Verify) using same loop pattern
- `skills/c3-query/SKILL.md`: Added Step 0 for Socratic intent clarification before navigation
- `templates/adr-000.md`: Added Adoption Progress table and Open Questions section

### Documentation
- Added `docs/plans/2026-01-05-staged-onboarding-design.md` capturing the full design rationale

## [2.1.0] - 2026-01-05

### Added
- Component file creation during onboarding (subagent creates docs for each discovered component)
- Category-specific component templates: `component-foundation.md`, `component-auxiliary.md`, `component-feature.md`
- Testing documentation at each layer: E2E (Context), Integration/Mocking/Fixtures (Container), Unit (Component)

### Changed
- Component categories simplified to Foundation/Auxiliary/Feature (removed Presentation)
- Audit checks streamlined to 4 essential checks (removed template-enforced redundancies)
- Container template now uses Foundation/Auxiliary/Feature sections with descriptions

### Removed
- Generic `component.md` template (replaced by category-specific templates)

## [2.0.1] - 2026-01-05

### Fixed
- `c3-init.sh` now falls back to `sed` when `envsubst` is not available

## [2.0.0] - 2026-01-02

### Added
- OpenCode support with dual-platform distribution
- `/release` command for version bumping and changelog generation
- Activation harnesses in all skills for consistent behavior
- Shared layer navigation pattern extracted to `references/layer-navigation.md`

### Changed
- **BREAKING**: Restructured skills from 5 to 3 (`c3`, `c3-query`, `c3-alter`)
- Skills now invoke other skills directly (not agents)
- Commands delegate to skills (`/c3`, `/query`, `/alter`, `/onboard`)

### Removed
- Dead code: orphan templates, duplicate references
- Unused agent definitions

### Documentation
- Added comprehensive CLAUDE.md with migration awareness
- Session history tracking in CLAUDE.md
