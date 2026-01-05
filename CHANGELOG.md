# Changelog

All notable changes to the C3 Skill plugin will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
