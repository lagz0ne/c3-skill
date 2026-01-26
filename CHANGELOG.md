# Changelog

All notable changes to the C3 Skill plugin will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [3.7.0] - 2026-01-26

### Added
- **c3-dev agent**: New TDD execution agent for implementing ADR-approved changes
  - Creates tasks per work item linked to ADR
  - Implements RED-GREEN TDD cycle with Socratic dialogue
  - Validates code against patterns via c3-patterns dispatch
  - Creates summary task with integrity check before ADR transition
  - Granular task states for parallel visibility: pending → in_progress → blocked → testing → implementing → completed
- **README.md**: Comprehensive plugin documentation with d2 diagrams via diashort
  - Agent ecosystem diagram
  - TDD workflow diagram
  - C3 structure diagram
  - Simple 4-step example walkthrough

### Changed
- **c3-orchestrator Phase 6**: Now dispatches c3-dev when user chooses "Execute now"
- **c3-orchestrator Phase 7**: Clarified to only apply for manual implementation path
- **c3-adr-transition**: Added integrity check for summary task (Step 1b)
  - Verifies summary task exists with correct metadata before transition
  - Added TaskList and TaskGet tools for integrity verification
  - Added example blocks for better triggering

### Fixed
- **Command descriptions**: Added trigger phrases to query.md and alter.md
- **c3.md command**: Added missing argument-hint field
- **c3 skill description**: Added "verify docs", "check documentation" triggers
- **c3-synthesizer**: Added token limit constraint for consistency

## [3.6.0] - 2026-01-21

### Added
- **Content Separation Verification (Phase 9)**: New audit phase that validates proper separation between components (domain logic) and refs (usage patterns)
  - The Separation Test: "Would this content change if we swapped the underlying technology?"
  - Detects missing refs for technologies, integration patterns in components, duplicated patterns
- **c3-content-classifier agent**: LLM-based content classification for Phase 9 audit
  - Identifies misplaced content between components and refs
  - Includes worked example with reasoning
  - Accepts optional technology context for better classification
- **content-separation.md reference**: Canonical definition for component vs ref separation
- **references/README.md**: Index of all reference files for discoverability
- **Proactive ref extraction in onboarding**: Section 1.2.3 guides extracting refs during component documentation

### Changed
- **ref.md template**: Added When/Why/Conventions structure for clearer ref documentation
- **Skill/agent relationship clarification**: c3-query and c3-alter skills now document their relationship to c3-navigator and c3-orchestrator agents
- **c3/SKILL.md Adopt mode simplified**: Now routes directly to /onboard skill without duplicating workflow
- **CLAUDE.md plugin guidance**: Updated to emphasize auto-discovery (explicit component paths break loading)

### Fixed
- **Skill name typo**: Fixed `onboarding-c3` → `/onboard` in c3/SKILL.md
- **Template typo**: Fixed `Applicable atterns` → `Applicable Patterns` in component.md

## [3.5.1] - 2026-01-21

### Fixed
- **Plugin loading broken by explicit paths**: Removed `commands`, `skills`, `agents`, `hooks` declarations from plugin.json - Claude Code uses auto-discovery
- **Release command validation**: Added Step 4 to validate plugin.json and remove explicit paths before every release

## [3.5.0] - 2026-01-21

### Added
- **ADR Lifecycle Tracking**: Complete tracking from acceptance to implementation
  - `base-commit` field captured when ADR transitions to accepted
  - `c3-verifier` script compares approved-files vs actual git changes
  - `c3-adr-transition` agent handles status transitions with verification
  - Verification results documented: matched, unplanned, untouched files
- **Phase 7 in c3-orchestrator**: Implementation completion with verification
- **Simplified c3-gate**: Now requires any accepted ADR (not specific file approval)
- **11-check validation hub in c3-synthesizer**: Consolidates all validation before ADR generation
  - 6 boundary checks from c3-impact: ownership, redundancy, sibling overlap, composition, leaky abstraction, correct layer
  - 4 ref checks from c3-patterns: follows ref, ref usage, missing ref, stale ref
  - 1 context alignment check in synthesizer

### Changed
- **Validation happens BEFORE ADR**: Loop closes at synthesizer, ADR is always valid when generated
- **c3-impact expanded**: Now performs boundary analysis (6 checks)
- **c3-patterns expanded**: Now performs ref health checks (4 checks)
- **c3-alter workflow simplified**: Removed Stage 4b (audit), validation now implicit in ADR creation

### Removed
- **c3-adr-auditor agent**: Consolidated into c3-synthesizer validation hub
- **c3-audit-adr skill**: No longer needed
- **Phase 5a in c3-orchestrator**: Audit phase removed, validation moved earlier

## [3.4.3] - 2026-01-21

### Added
- **c3-adr-auditor agent**: New agent that validates ADRs against C3 architectural principles before approval
  - Checks abstraction boundaries (component doing sibling's job)
  - Checks composition rules (orchestration vs hand-off)
  - Checks context alignment (contradicting Key Decisions)
  - Checks ref compliance (touching pattern domain without citation)
- **c3-audit-adr skill**: Direct invocation wrapper for ADR auditing (`/c3 audit-adr`)
- **Phase 5a in c3-orchestrator**: ADR audit gate between generation and acceptance - ADR cannot be accepted until auditor returns PASS
- **Stage 4b in c3-alter**: Audit step integrated into alter workflow checklist

### Changed
- **c3-orchestrator workflow**: Now includes mandatory audit step before user can accept ADR
- **c3-alter workflow**: Progress checklist updated with Stage 4b (audit) and Stage 4c (accept)

## [3.4.2] - 2026-01-21

### Fixed
- **SessionStart hook**: Use JSON `hookSpecificOutput.additionalContext` format for context injection - plain text output was showing as "Success" but not being injected into conversation context
- **hooks.json**: Add `matcher: "startup|resume"` to SessionStart hook for proper event matching

## [3.4.1] - 2026-01-20

### Fixed
- **plugin.json**: Removed explicit component paths - auto-discovery looks at parent of `.claude-plugin/` for skills/agents/commands/hooks directories, explicit paths were breaking loading when plugin.json is nested

## [3.4.0] - 2026-01-20

### Added
- **c3-ref skill**: New skill for managing cross-cutting patterns as first-class architecture artifacts
  - `add` mode: Create refs from discovered patterns with automatic citing component updates
  - `update` mode: Modify refs with impact analysis across all citing components
  - `list` mode: Show all refs with citation counts
  - `usage` mode: Show which components cite a specific ref

- **Constraint Chain Query**: New query type in c3-query skill
  - Ask "what constraints apply to c3-XXX" to see full inheritance chain
  - Shows Context → Container → Refs constraints with MAY/MUST NOT boundaries
  - Generates visual diagram of constraint inheritance

- **Phase 4b: Pattern Violation Gate** in c3-orchestrator
  - Refs are now blocking constraints - violations cannot be silently bypassed
  - Changes that break patterns require explicit override with justification in ADR
  - Options: update the pattern, override with justification, or rethink approach

- **Phase 8: Abstraction Boundaries** in audit-checks
  - Detects when components take on container/context responsibilities
  - Checks for: container bleeding, context bleeding, component orchestrating peers, ref bypass
  - FAIL on clear violations, WARN on potential issues needing review

- **Layer Constraints sections** in component.md and container.md templates
  - Explicit MUST/MUST NOT boundaries for each layer
  - Prevents abstraction violations through clear documentation

### Changed
- **ADR template**: Added "Pattern Overrides" section for documenting justified pattern violations

### Documentation
- **CLAUDE.md**: Updated plugin testing instructions with correct structure explanation

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
