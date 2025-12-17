# Migration: 20251217-adr-status-enforcement

## Changes

### ADR Status Enforcement Added (c3-audit, c3-design, c3-migrate)

Added mandatory status workflow enforcement:
- New ADRs start as `status: proposed`
- Status workflow: `proposed` → `accepted` → `implemented`
- Only `implemented` ADRs appear in TOC

Changes by skill:

**c3-audit** - Added "📋 ADR STATUS ENFORCEMENT (MANDATORY)" section:
- Status workflow diagram
- Transition rules table (who can do what)
- Verification checklist for `implemented` status
- Red flags for common violations
- Audit checklist for ADR status validation

**c3-design** - Added reminder in Design Phases:
- New ADRs start as `status: proposed`
- Won't appear in TOC until `implemented`
- Reference to workflow in adr-template.md

**c3-migrate** - Added post-migration ADR status check:
- Commands to verify all ADRs have status field
- Commands to list ADRs by status
- Reminder about TOC filtering

## Transforms

**No automatic transforms required.**

This migration adds skill-internal enforcement. User `.c3/` directories are not affected - ADR status field was already documented in adr-template.md.

## Verification

```bash
# Verify ADR status enforcement added
grep -l "ADR STATUS ENFORCEMENT" skills/c3-audit/SKILL.md
grep -l "ADR Status" skills/c3-design/SKILL.md
grep -l "ADR Status Check" skills/c3-migrate/SKILL.md
```
