# Changelog

All notable changes to the C3 Skill plugin will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
