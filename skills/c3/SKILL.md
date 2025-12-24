---
name: c3
description: Use when exploring, documenting, or changing system architecture using C3 methodology
---

# C3 Architecture

Help explore and document architecture through guided discovery.

Start by loading the current state (.c3/README.md + TOC.md), then ask questions one at a time. Once you understand the scope, either answer or create changes with an ADR.

## The Process

**Loading the base:**
- Read .c3/README.md and .c3/TOC.md first
- If neither exists, guide through initial setup

**Understanding the request:**
- Ask questions one at a time
- Load relevant docs as gaps emerge
- Keep track of what you've learned (traces)

**Reaching "enough":**
- Confirm scope with user before acting
- If user corrects, go back and ask more

**Acting:**
- Question → answer from traces
- Change → create ADR first
- Verify → audit (docs vs code) or conform (docs vs spec)
- If new gaps emerge, return to discovery

## Layer Model

**c3-0** - System overview
- Actors (who uses this)
- Container relationships (intentions, not details)
- High-level only

**c3-N** - Orchestration
- Parts and how they relate
- Tech choices that shape organization
- Each part listed becomes a c3-NNN

**c3-NNN** - Implementation
- One doc per part
- One aspect per doc

## The Cascade

Upper layers define WHAT. Lower layers define HOW.

## Cross-Cutting Concerns

- Define patterns at c3-N
- Reference in c3-NNN, don't redefine

## Writing Documents

When creating/updating docs, load quality criteria:
- c3-0 → read references/context-quality.md
- c3-N → read references/container-quality.md
- c3-NNN → read references/component-quality.md

## Key Principles

- One question at a time
- Load docs on demand
- Confirm before acting
- ADR for any change
