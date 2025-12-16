# C3-Skill Optimization Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Reduce skill file sizes from 800-4900 words to <500 words each by extracting heavy content to references.

**Architecture:** Extract duplicated/heavy content from skills into `references/` files, then cross-reference. Skills become thin orchestration layers; references hold reusable content.

**Tech Stack:** Markdown, YAML frontmatter

---

## Overview

Current state (all skills exceed 500 word limit):

| Skill | Words | Target | Extract To |
|-------|-------|--------|------------|
| c3-audit | 4,884 | ~450 | references/audit-*.md |
| c3-design | 3,434 | ~450 | references/design-*.md |
| c3-adopt | 2,159 | ~400 | (already uses discovery-questions.md) |
| c3-component-design | 1,624 | ~450 | (inline OK, slim prose) |
| c3-container-design | 1,185 | ~450 | (inline OK, slim prose) |
| c3-context-design | 807 | ~400 | (inline OK, slim prose) |
| c3-config | 1,264 | ~400 | references/settings-schema.md |
| c3-migrate | 914 | ~400 | (inline OK) |

Also: Fix CSO descriptions + documentation inconsistencies.

---

## Task 1: Create Layer Rules Reference (for c3-audit)

Lines 70-130 of c3-audit duplicate the defaults.md content. Extract to a single reference.

**Files:**
- Create: `references/layer-audit-rules.md`

**Step 1: Create the reference file**

```markdown
# Layer Audit Rules

Quick reference for auditing C3 documentation layer compliance.

## Context Layer (c3-0) - Bird's-eye view

**MUST INCLUDE:**
- Container responsibilities (WHY each exists)
- Container relationships (how they connect)
- Connecting points (APIs, protocols, events)
- External actors (who/what interacts)
- System boundary (inside vs outside)

**MUST EXCLUDE:**
- Component lists (push to Container)
- How containers work internally (push to Container)
- Implementation details (push to Component)

**LITMUS:** "Is this about WHY containers exist and HOW they relate?"

**DIAGRAMS:** System Context, Container Overview
**AVOID:** Sequence with methods, class diagrams, flowcharts with logic

---

## Container Layer (c3-N) - Inside view

**MUST INCLUDE:**
- Component responsibilities (WHAT each does)
- Component relationships (how they interact)
- Data flows (how data moves)
- Business flows (workflows spanning components)
- Inner patterns (logging, config, errors)

**MUST EXCLUDE:**
- WHY this container exists (push to Context)
- Container-to-container details (push to Context)
- HOW components work (push to Component)

**LITMUS:** "Is this about WHAT components do and HOW they relate?"

**DIAGRAMS:** Component Relationships, Data Flow
**AVOID:** System context, actor diagrams

---

## Component Layer (c3-NNN) - Close-up view

**MUST INCLUDE:**
- Flows (step-by-step processing)
- Dependencies (what it calls)
- Decision logic (branching points)
- Edge cases (non-obvious scenarios)
- Error handling (what can go wrong)

**MUST EXCLUDE:**
- WHAT this component does (already in Container)
- Component relationships (push to Container)
- Container relationships (push to Context)

**LITMUS:** "Is this about HOW this component implements its contract?"

**DIAGRAMS:** Flowcharts, Sequence (to dependencies), State charts
**AVOID:** System context, container overview
```

**Step 2: Verify file created**

Run: `wc -w references/layer-audit-rules.md`
Expected: ~250 words

**Step 3: Commit**

```bash
git add references/layer-audit-rules.md
git commit -m "refactor(references): extract layer audit rules from c3-audit"
```

---

## Task 2: Create Audit Report Templates Reference

Extract the verbose XML templates and report templates from c3-audit.

**Files:**
- Create: `references/audit-report-templates.md`

**Step 1: Create the reference file**

```markdown
# Audit Report Templates

## Structure Audit Output

```xml
<structure_audit>
  <id_pattern_violations>
    <violation doc="path">
      <expected>pattern</expected>
      <actual>found</actual>
      <fix>action</fix>
    </violation>
  </id_pattern_violations>
  <path_violations>...</path_violations>
  <frontmatter_violations>...</frontmatter_violations>
</structure_audit>
```

## Layer Content Audit Output

```xml
<layer_content_audit>
  <context doc=".c3/README.md">
    <required_present>
      <item check="name">PASS|FAIL</item>
    </required_present>
    <violations>
      <violation type="content_too_detailed" severity="high">
        <location>Line N</location>
        <content>description</content>
        <rule>violated rule</rule>
        <fix>action</fix>
      </violation>
    </violations>
  </context>
  <container doc="path" id="c3-N">...</container>
  <component doc="path" id="c3-NNN">...</component>
</layer_content_audit>
```

## Drift Findings Output

```xml
<drift_findings>
  <finding type="phantom|technology_mismatch|undocumented" severity="critical|high|medium">
    <doc_id>c3-N</doc_id>
    <documented>what doc says</documented>
    <actual>what code shows</actual>
    <action>fix action</action>
  </finding>
</drift_findings>
```

## ADR Audit Output

```xml
<adr_audit_result adr="adr-YYYYMMDD-slug">
  <doc_changes_complete>yes|no</doc_changes_complete>
  <code_verification_pass>yes|no</code_verification_pass>
  <recommendation>READY TO MARK IMPLEMENTED | GAPS REMAIN</recommendation>
  <gaps_if_any>
    <gap type="doc|code">description</gap>
  </gaps_if_any>
</adr_audit_result>
```

## Methodology Audit Report

```markdown
# C3 Methodology Audit Report

**Date:** YYYY-MM-DD
**Target:** [path]

## Summary

| Category | Pass | Warn | Fail |
|----------|------|------|------|
| Structure | N | N | N |
| Layer Content | N | N | N |
| Diagrams | N | N | N |
| Contract Chain | N | N | N |

**Overall:** COMPLIANT / NEEDS_FIXES / NON_COMPLIANT

## Violations

[Tables of violations by type]

## Priority Actions

1. [action]
```

## Full Audit Report

```markdown
# C3 Audit Report

**Date:** YYYY-MM-DD
**Mode:** [Full | ADR | Container | Quick]

## Statistics

| Metric | Documented | In Code | Match |
|--------|------------|---------|-------|
| Containers | N | N | check |
| Components | N | N | check |

## Findings by Severity

### Critical/High/Medium

| ID | Type | Issue | Action |
|----|------|-------|--------|

## Recommended Actions

1. [action]
```
```

**Step 2: Verify file created**

Run: `wc -w references/audit-report-templates.md`
Expected: ~300 words

**Step 3: Commit**

```bash
git add references/audit-report-templates.md
git commit -m "refactor(references): extract audit report templates"
```

---

## Task 3: Slim c3-audit Skill

Replace verbose inline content with references.

**Files:**
- Modify: `skills/c3-audit/SKILL.md`

**Step 1: Read current file structure**

Run: `head -70 skills/c3-audit/SKILL.md`
Identify sections to keep vs reference.

**Step 2: Rewrite SKILL.md to slim version**

The new structure should be approximately:

```markdown
---
name: c3-audit
description: Use when verifying C3 documentation quality - checks methodology compliance (layer rules, structure, diagrams) and implementation conformance (docs vs code drift)
---

# C3 Audit

## Overview

Verify C3 documentation from two angles:

1. **Methodology Compliance** - Do docs follow C3 layer rules?
2. **Implementation Conformance** - Do docs match code reality?

**Announce:** "I'm using the c3-audit skill to audit documentation."

## Audit Modes

| Mode | Purpose | When |
|------|---------|------|
| **Methodology** | Verify layer rules | After doc changes |
| **Full** | All docs vs code | Periodic health check |
| **ADR** | Verify ADR implemented | After ADR work |
| **ADR-Plan** | ADR ↔ Plan coherence | Before handoff |
| **Container** | Single container audit | After container changes |
| **Quick** | Inventory check only | Fast sanity check |

## Mode 0: Methodology Audit

### Phase 1: Load Layer Rules

Load rules from `references/layer-audit-rules.md` (or layer skill defaults).

### Phase 2: Structure Check

Verify IDs, paths, frontmatter per `references/v3-structure.md`.

### Phase 3: Layer Content Audit

For each doc, verify content matches layer rules:
- Context: Has container inventory, protocols. NO component lists.
- Container: Has component inventory, tech stack. NO step-by-step algorithms.
- Component: Has flow, dependencies. NO sibling relationships.

### Phase 4: Diagram Audit

Verify diagram types appropriate per layer (see `references/layer-audit-rules.md`).

### Phase 5: Contract Chain

Check for orphans (undocumented) and phantoms (documented but missing).

### Phase 6: Load Layer Skills

Load c3-context-design, c3-container-design, c3-component-design for suggestion generation.

### Phase 7: Generate Report

Use template from `references/audit-report-templates.md`.

## Mode 1: Full Audit

1. Load all C3 docs (inventory)
2. Explore codebase (Task tool, Explore agent)
3. Compare: inventory drift, technology drift, structure drift, protocol drift
4. Generate drift report

## Mode 2: ADR Audit

1. Load ADR, extract Changes Across Layers + Verification Checklist
2. Verify each doc change was made
3. Verify code matches (use verification items)
4. If all pass: transition status to `implemented`, rebuild TOC

## Mode 3: ADR-Plan Coherence

1. Extract ADR sections + Implementation Plan
2. Verify: Layer Changes → Code Changes (all mapped)
3. Verify: Verifications → Acceptance Criteria (all mapped)
4. Check for orphans, vague locations, untestable criteria

## Mode 4: Container Audit

Focus on single container: technology match, component coverage, pattern compliance.

## Mode 5: Quick Audit

```bash
# Inventory counts only
grep -c "| c3-[0-9]" .c3/README.md
ls -d .c3/c3-[0-9]-*/ 2>/dev/null | wc -l
```

## Checklist

See `references/audit-report-templates.md` for full checklists per mode.

## Related

- `references/layer-audit-rules.md` - Layer compliance rules
- `references/audit-report-templates.md` - Report formats
- `references/v3-structure.md` - ID/path patterns
```

**Step 3: Verify word count**

Run: `wc -w skills/c3-audit/SKILL.md`
Expected: ~400-500 words (down from 4884)

**Step 4: Commit**

```bash
git add skills/c3-audit/SKILL.md
git commit -m "refactor(c3-audit): slim skill from 4884 to ~450 words"
```

---

## Task 4: Slim c3-design Skill

Extract exploration mode details and phase details.

**Files:**
- Create: `references/design-exploration-mode.md`
- Create: `references/design-phases.md`
- Modify: `skills/c3-design/SKILL.md`

**Step 1: Create exploration mode reference**

```markdown
# C3 Design - Exploration Mode

Read-only navigation of existing C3 documentation.

## When to Use

- "What's the architecture?"
- "How does X work?"
- "Why did we choose X?"

## Process

1. **Load TOC:** `cat .c3/TOC.md`

2. **Parse for overview:**
   - Document counts
   - Container summaries
   - ADR list

3. **Present based on intent:**

   **General orientation:**
   ```
   ## {System} Architecture
   **Purpose:** {from Context}
   ### At a Glance
   - Containers: N
   - Components: N
   - Decisions: N ADRs
   ### Container Overview
   | ID | Container | Purpose |
   ```

   **Focused:** Load relevant docs for user's area.

   **Decisions:** Present ADR list, load specific ones.

4. **Navigate on demand:**
   - "Tell me about c3-1" → Load container
   - "Why X?" → Search ADRs
   - "I need to change Y" → Switch to Design Mode
```

**Step 2: Create design phases reference**

```markdown
# C3 Design - Phases

## Phase 1: Surface Understanding

Read TOC, form hypothesis about what layers are affected.

**Output:** Initial layer impact hypothesis
**Gate:** Hypothesis formed

## Phase 2: Iterative Scoping

Loop: HYPOTHESIZE → EXPLORE → DISCOVER

1. **Hypothesize:** "This change affects [layers]"
2. **Explore:** Use layer skills to investigate
3. **Discover:** Find actual impacts, adjust hypothesis

**Output:** Stable scope (layers, docs, changes)
**Gate:** No new discoveries in last iteration

## Phase 3: ADR Creation

Create ADR in `.c3/adr/adr-YYYYMMDD-slug.md`:
- Status: proposed
- Changes Across Layers: what docs change
- Verification Checklist: how to verify implementation

**Output:** ADR file exists
**Gate:** File created with all sections

## Phase 4: Handoff

Execute `.c3/settings.yaml` handoff steps (or defaults).

**Output:** Handoff complete
**Gate:** All steps executed
```

**Step 3: Rewrite c3-design SKILL.md**

Target ~450 words:

```markdown
---
name: c3-design
description: Use when designing, updating, or exploring system architecture with C3 methodology - iterative scoping through hypothesis, exploration, and discovery across Context/Container/Component layers
---

# C3 Architecture Design

## Overview

Transform requirements into C3 documentation through iterative scoping. Also supports exploration-only mode.

**Core principle:** Hypothesis → Explore → Discover → Iterate until stable.

**Announce:** "I'm using the c3-design skill to guide architecture design."

## Mode Detection

| Intent | Mode |
|--------|------|
| "What's the architecture?" | Exploration |
| "How does X work?" | Exploration |
| "I need to add/change..." | Design |
| "Why did we choose X?" | Exploration |

### Exploration Mode

See `references/design-exploration-mode.md` for full workflow.

Quick: Load TOC → Present overview → Navigate on demand.

### Design Mode

Full workflow with mandatory phases.

## Design Phases

**IMMEDIATELY create TodoWrite items:**
1. Phase 1: Surface Understanding
2. Phase 2: Iterative Scoping
3. Phase 3: ADR Creation (MANDATORY)
4. Phase 4: Handoff

See `references/design-phases.md` for detailed phase requirements.

**Rules:**
- Mark `in_progress` when starting
- Mark `completed` only when gate met
- Phase 3 gate: ADR file MUST exist
- Phase 4 gate: Handoff steps executed

## Quick Reference

| Phase | Output | Gate |
|-------|--------|------|
| 1. Surface | Layer hypothesis | Hypothesis formed |
| 2. Scope | Stable scope | No new discoveries |
| 3. ADR | ADR file | File exists |
| 4. Handoff | Complete | Steps executed |

## Layer Skill Delegation

| Impact | Delegate To |
|--------|-------------|
| Container inventory changes | c3-context-design |
| Component organization | c3-container-design |
| Implementation details | c3-component-design |

## Checklist

- [ ] Mode detected (exploration vs design)
- [ ] If design: TodoWrite phases created
- [ ] Each phase completed with gate
- [ ] ADR created (design mode)
- [ ] Handoff executed

## Related

- `references/design-exploration-mode.md`
- `references/design-phases.md`
- `references/adr-template.md`
```

**Step 4: Verify word count**

Run: `wc -w skills/c3-design/SKILL.md`
Expected: ~400-450 words

**Step 5: Commit**

```bash
git add references/design-exploration-mode.md references/design-phases.md skills/c3-design/SKILL.md
git commit -m "refactor(c3-design): slim skill from 3434 to ~450 words"
```

---

## Task 5: Slim c3-adopt Skill

The skill already references `references/discovery-questions.md`. Focus on trimming verbose prose.

**Files:**
- Modify: `skills/c3-adopt/SKILL.md`

**Step 1: Read current file**

Run: `wc -l skills/c3-adopt/SKILL.md`
Identify sections with verbose prose to trim.

**Step 2: Trim to essentials**

Keep:
- Overview (1 paragraph)
- When to Use table
- Process table (delegate to layer skills)
- Fresh vs Existing decision
- Checklist

Remove/trim:
- Extended examples
- Verbose explanations
- Redundant instructions

**Step 3: Verify word count**

Run: `wc -w skills/c3-adopt/SKILL.md`
Target: ~400 words

**Step 4: Commit**

```bash
git add skills/c3-adopt/SKILL.md
git commit -m "refactor(c3-adopt): slim skill from 2159 to ~400 words"
```

---

## Task 6: Slim Layer Skills (Context, Container, Component)

These skills (807-1624 words) need moderate trimming.

**Files:**
- Modify: `skills/c3-context-design/SKILL.md`
- Modify: `skills/c3-container-design/SKILL.md`
- Modify: `skills/c3-component-design/SKILL.md`

**Step 1: Review each skill**

For each, identify:
- Core orchestration (keep)
- Verbose explanations (trim)
- Duplicated content (reference defaults.md instead)

**Step 2: Trim each skill**

Target word counts:
- c3-context-design: 807 → ~400
- c3-container-design: 1185 → ~450
- c3-component-design: 1624 → ~450

**Step 3: Commit**

```bash
git add skills/c3-context-design/SKILL.md skills/c3-container-design/SKILL.md skills/c3-component-design/SKILL.md
git commit -m "refactor(layer-skills): slim context/container/component skills"
```

---

## Task 7: Slim c3-config Skill

**Files:**
- Modify: `skills/c3-config/SKILL.md`

**Step 1: Trim verbose prose**

Keep:
- Overview
- Settings structure reference
- Socratic refinement process
- Checklist

**Step 2: Verify word count**

Target: ~400 words

**Step 3: Commit**

```bash
git add skills/c3-config/SKILL.md
git commit -m "refactor(c3-config): slim skill from 1264 to ~400 words"
```

---

## Task 8: Fix CSO Descriptions

Standardize all descriptions to "Use when..." format.

**Files:**
- Modify: All 8 `skills/*/SKILL.md` frontmatter

**Step 1: Update descriptions**

| Skill | Current Start | New Description |
|-------|---------------|-----------------|
| c3-audit | "Audit C3..." | "Use when verifying C3 documentation quality - checks methodology compliance and implementation conformance" |
| c3-design | "Use when..." | (already good) |
| c3-adopt | "Initialize C3..." | "Use when bootstrapping C3 documentation for any project - guides through Socratic discovery and delegates to layer skills" |
| c3-context-design | "Explore Context..." | "Use when exploring Context level impact during scoping - system boundaries, actors, cross-container interactions" |
| c3-container-design | "Explore Container..." | "Use when exploring Container level impact during scoping - technology choices, component organization, patterns" |
| c3-component-design | "Explore Component..." | "Use when exploring Component level impact during scoping - implementation details, configuration, dependencies" |
| c3-config | "Create and refine..." | "Use when configuring project preferences in .c3/settings.yaml - diagram tools, layer guidance, guardrails" |
| c3-migrate | "Migrate .c3/..." | "Use when upgrading .c3/ documentation to current skill version - reads VERSION, applies transforms" |

**Step 2: Verify each frontmatter**

Run: `head -5 skills/*/SKILL.md`
All should start with "Use when..."

**Step 3: Commit**

```bash
git add skills/*/SKILL.md
git commit -m "refactor(skills): standardize CSO descriptions to 'Use when...' format"
```

---

## Task 9: Fix Documentation Inconsistencies

**Files:**
- Modify: `README.md`
- Modify: `CLAUDE.md`

**Step 1: Remove examples/ reference from README**

Lines 294-299 reference nonexistent `examples/` directory. Remove or replace.

Find:
```markdown
## Examples

See `examples/` directory for:
- Sample `.c3/` structure
- Example context/container/component documents
- Sample ADRs
```

Replace with:
```markdown
## Examples

See the plugin's own `.c3/` directory for a real-world example of C3 documentation structure.
```

**Step 2: Fix CLAUDE.md utility skills table**

The "Utility Skills" section lists skills that don't exist (c3-use, c3-locate, c3-toc, c3-naming).

Find and remove from the table:
- c3-use
- c3-locate
- c3-toc
- c3-naming

Or note they are patterns in references (not skills):
```markdown
### Utility Patterns (References, not Skills)

| Pattern | Reference File | Purpose |
|---------|---------------|---------|
| Locate | `references/lookup-patterns.md` | ID-based document retrieval |
| Naming | `references/naming-conventions.md` | ID and naming conventions |
```

**Step 3: Commit**

```bash
git add README.md CLAUDE.md
git commit -m "docs: fix references to nonexistent examples/ and utility skills"
```

---

## Task 10: Final Verification

**Step 1: Verify all word counts**

```bash
for f in skills/*/SKILL.md; do echo "$f: $(wc -w < "$f") words"; done
```

All should be <500 words.

**Step 2: Verify references created**

```bash
ls -la references/
```

Should include new files:
- layer-audit-rules.md
- audit-report-templates.md
- design-exploration-mode.md
- design-phases.md

**Step 3: Verify CSO descriptions**

```bash
for f in skills/*/SKILL.md; do echo "=== $f ===" && head -5 "$f"; done
```

All descriptions should start with "Use when..."

**Step 4: Final commit and version bump**

```bash
# Update VERSION
echo "$(date +%Y%m%d)-skill-optimization" > VERSION

git add VERSION
git commit -m "chore: bump version to $(cat VERSION)"
```

---

## Summary

| Task | Files | Action |
|------|-------|--------|
| 1 | references/layer-audit-rules.md | Create |
| 2 | references/audit-report-templates.md | Create |
| 3 | skills/c3-audit/SKILL.md | Slim 4884→450 |
| 4 | skills/c3-design/SKILL.md + refs | Slim 3434→450 |
| 5 | skills/c3-adopt/SKILL.md | Slim 2159→400 |
| 6 | Layer skills (3) | Slim each |
| 7 | skills/c3-config/SKILL.md | Slim 1264→400 |
| 8 | All skill frontmatters | Fix CSO |
| 9 | README.md, CLAUDE.md | Fix docs |
| 10 | Verification | Verify all |

**Total estimated commits:** 10-12
