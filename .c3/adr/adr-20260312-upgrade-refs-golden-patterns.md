---
id: adr-20260312-upgrade-refs-golden-patterns
title: Upgrade Refs — Golden Patterns, Enforcement, Measurement
type: adr
status: implemented
date: 2026-03-12
goal: Make refs actionable by adding golden patterns with prescriptive How sections
affects: [c3-113, c3-116, c3-210]
---

# Upgrade Refs — Golden Patterns, Enforcement, Measurement

## Goal

Make refs actionable by adding golden patterns (prescriptive `## How` sections), enforcement (ref compliance gates in change and audit flows), and measurement (ref governance metric in coverage output).

## Work Breakdown

**Skill Instructions (Track A):**
- `ref.md`: Discovery-first Add flow with quality gate (compliance questions derivable from `## How`)
- `change.md`: Phase 3b — Ref Compliance Gate with adversarial framing, structured verdict output
- `audit.md`: Phase 7b — Ref Compliance spot-check against golden patterns

**Template + Schema (Track B):**
- `cli/templates/ref.md`: 7-criteria quality rubric, format-flexible `## How`, dual-purpose `## Not This`
- `schema.go`: How purpose updated to "Golden pattern — prescriptive examples and implementation guidance"

**Go — Ref Governance Metric (Track C):**
- `index.go`: `RefGovernanceResult` struct + `RefGovernance()` function
- `coverage.go`: Integrated ref governance into coverage output (JSON + human)

**Go — Scope Cross-Check (Track D):**
- `check_enhanced.go`: WARN when ref scopes container but child component doesn't cite it

**Pre-flight:**
- `code-map.yaml`: Mapped `cli/internal/schema/**` and `cli/internal/index/**` to c3-113
- Updated c3-113 and c3-116 component docs

## Risks

- Ref governance starts at 0% for existing components — creates adoption backlog
- Scope cross-check may be noisy for broadly-scoped refs — mitigated by WARN severity
